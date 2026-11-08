# Architecture

## Overview

Quantum Orchestrator is a Kubernetes Operator that provides a unified API for
submitting, scheduling, and tracking quantum computing workloads across
multiple cloud providers, on-prem GPUs, and direct QPU vendors.

## Provider Matrix

| Provider | Cloud / On-Prem | Auth | Workload Type |
|----------|----------------|------|---------------|
| **IBM Quantum** | IBM Cloud | IAM API key → bearer token (1h) | Nighthawk, Heron, Starling, Qiskit v2.5, RelayBP QEC |
| **AWS Braket** | AWS Cloud | IRSA / IAM | IonQ, Rigetti, IQM, QuEra, Hybrid Jobs |
| **Azure Quantum** | Azure Cloud | Workload Identity / SP | IonQ, Quantinuum, Rigetti, QIR, Sessions |
| **NVIDIA CUDA-Q** | On-prem GPU / NGC | `nvidia.com/gpu` resource | CUDA-Q sim, cuQuantum, CUDA-QX QEC |
| **NVIDIA NVQLink** | On-prem GPU + QPU | RoCE / RDMA | Real-time QEC (QPU + GPU) |
| **Direct QPU** | Cloud / On-prem | API key | IonQ, Rigetti, Quantinuum, IQM |
| **qBraid** | Any | API key | Universal (10+ providers) |
| **Local** | In-cluster | — | Simulators (Aer, qsim, Qibo) |

## Components

### Operator (Go)

The core controller that:
1. Watches 6 CRDs: `QuantumJob`, `QuantumSession`, `QPUBackend`, `QuantumProviderBinding`, `QuantumSimulator`, `NVQLinkLogicalQPU`
2. Resolves the target provider from the job spec
3. Submits the job via the appropriate `Provider` adapter
4. Polls status and updates the CRD `status` subresource
5. Handles finalizers for clean cancellation

### Provider Adapters (Go)

| Adapter | Provider | Auth | Key Detail |
|---------|----------|------|------------|
| `IBMAdapter` | IBM Quantum | IBM Cloud IAM (API key → bearer token, 1h expiry) | Instance CRN, region, Nighthawk/Heron/Starling |
| `BraketAdapter` | AWS Braket | IRSA / IAM | QuantumTask, Hybrid Jobs, S3 results |
| `AzureQuantumAdapter` | Azure Quantum | Workload Identity / SP | Sessions, QIR, Integrated Hybrid |
| `NvidiaAdapter` | NVIDIA CUDA-Q / cuQuantum / NVQLink | `nvidia.com/gpu` | NGC container, multi-node, CUDA-Q → Braket |
| `DirectAdapter` | IonQ, Rigetti, Quantinuum, IQM | API key | REST clients |
| `QbraidAdapter` | Any (via qBraid) | API key | Universal runtime sidecar |

### Transpiler Sidecar (Python)

A FastAPI service that converts circuits between formats:
- OpenQASM 3 ↔ QIR ↔ Quil ↔ Q# ↔ Qiskit ↔ Cirq ↔ PennyLane ↔ CUDA-Q ↔ PBC ↔ Clifford+T
- Uses qBraid SDK, qBraid-QIR, and Qiskit v2.5 (C API, PBC pipeline)

### CLI (quantumctl)

A cobra-based CLI that talks to the Kubernetes API:
- `quantumctl submit -f job.yaml`
- `quantumctl get jobs`
- `quantumctl logs <job>`
- `quantumctl results <job> --plot`
- `quantumctl devices`
- `quantumctl cost`

## IBM Cloud Integration

### Authentication Flow

```
┌──────────────┐         ┌──────────────────────────────┐
│  K8s Secret  │         │  IBM Cloud IAM               │
│  (API key)   │────────▶│  POST /identity/token        │
│              │         │  → bearer token (≤ 1h)       │
└──────────────┘         └──────────────┬───────────────┘
                                       │
                                       ▼
┌──────────────────────────────────────────────────────────┐
│  IBM Quantum Compute Service REST API                     │
│  POST /api/v1/jobs                                        │
│  GET  /api/v1/jobs/{id}                                   │
│  GET  /api/v1/jobs/{id}/results                           │
│  GET  /api/v1/backends                                    │
│                                                          │
│  Headers:                                                 │
│    Authorization: Bearer <token>                          │
│    x-ibm-cloud-api-instance: <CRN>                        │
└──────────────────────────────────────────────────────────┘
```

### Key Details

- **API key**: Long-lived, 44 chars, created in IBM Cloud dashboard
- **Bearer token**: Short-lived (≤ 1 hour), refreshed automatically by the adapter
- **Instance CRN**: Identifies the Quantum Compute Service instance
- **Region**: `us-east` (default, supports Open Plan) or `eu-south`
- **Endpoints**: `quantum-cloud.ibm.com` (cloud), local (on-prem Quantum System)
- **Firewall**: Must allow outbound HTTPS to IBM Quantum API endpoints

### IBM Quantum Compute Service vs On-Prem

| | IBM Cloud | On-Prem (Quantum System) |
|---|-----------|------------------------|
| Endpoint | `quantum-cloud.ibm.com` | Local (client infrastructure) |
| Auth | IAM API key | IAM API key (from onboarding) |
| Processors | Nighthawk, Heron | Nighthawk r2, Heron r3 |
| Use Case | Cloud research, production | HPC co-location (ORNL pattern) |

## NVIDIA NGC Integration

### Container Distribution

NVIDIA distributes CUDA-Q and cuQuantum via the **NGC (NVIDIA GPU Cloud)**
container registry:

| Image | Tag | Size | CUDA | Notes |
|-------|-----|------|------|-------|
| `nvcr.io/nvidia/quantum/cuda-quantum` | `cu13-0.16.0` | ~4.3 GB | CUDA 13 | Latest stable |
| `nvcr.io/nvidia/quantum/cuda-quantum` | `cu12-0.14.2` | ~4.3 GB | CUDA 12 | CUDA 12 compatible |
| `nvcr.io/nvidia/nightly/cuda-quantum` | `cu13-latest` | ~4.3 GB | CUDA 13 | Nightly / pre-release |

- **Multi-node**: Yes (multi-GPU, multi-node clusters)
- **Multi-arch**: amd64
- **Signed**: Yes (NVIDIA signature)
- **Scanned**: Yes (malware scan)
- **Source**: [github.com/NVIDIA/cuda-quantum](https://github.com/NVIDIA/cuda-quantum) (Apache 2.0)

### Contents

The NGC container includes:
- **CUDA-Q**: Quantum SDK (C++ + Python), NVQ++ compiler (LLVM-based)
- **cuQuantum**: cuStateVec, cuDensityMat, cuTensorNet, cuPauliProp, cuStabilizer
- **CUDA-QX**: QEC decoders (BP+OSD, tensor-network, nv-qldpc)
- **Frameworks**: Qiskit, Cirq, PennyLane, Qibo, QuTiP, TKET, Torch Quantum integrations

### Pulling the Image

```bash
# Stable (CUDA 13)
docker pull nvcr.io/nvidia/quantum/cuda-quantum:cu13-0.16.0

# Run with GPU
docker run --gpus all -it nvcr.io/nvidia/quantum/cuda-quantum:cu13-0.16.0

# In Kubernetes (via the operator)
# resources.limits: nvidia.com/gpu: 1
```

### NGC Access

- Free for most users: register at [developer.nvidia.com/ngc](https://developer.nvidia.com/ngc)
- Pull with: `docker login nvcr.io` (use NGC credentials)
- In Kubernetes: create a `docker-registry` secret with NGC credentials

```bash
kubectl create secret docker-registry ngc-pull-secret \
  --docker-server=nvcr.io \
  --docker-username=$NGC_USER \
  --docker-password=$NGC_KEY \
  -n quantum-system
```

## OpenShift AI Integration

The Quantum Orchestrator integrates natively with the Red Hat OpenShift AI
platform, enabling quantum workloads to be scheduled alongside AI/ML workloads
on the same GPU infrastructure.

### OpenShift AI Components Used

| Component | Role in Quantum Workloads |
|-----------|--------------------------|
| **Workbenches** | Jupyter notebooks preloaded with Qiskit, CUDA-Q, cuQuantum, PennyLane. Self-service quantum development. |
| **Kubeflow Pipelines 2.0** | AI pipelines with reusable components. Quantum jobs as pipeline steps (e.g. VQE → property prediction → ranking). |
| **Kubeflow Trainer v2** | Distributed training with TrainJob API. Classical side of hybrid quantum-classical workloads. |
| **Model Serving** | vLLM, custom runtimes, multi-node GPU. Serve AI models trained on quantum-enhanced data. |
| **Hardware Profiles** | Target specific GPU types (NVIDIA A100/H100, AMD MI355X) for quantum workloads. |
| **NVIDIA GPU Operator** | `nvidia.com/gpu` resources, GPU monitoring, MIG partitioning. |
| **AMD GPU Operator** | `amd.com/gpu` resources for AMD MI355X (IBM co-design pattern). |

### OpenShift Platform Components

| Component | Role |
|-----------|------|
| **OLM / OperatorHub** | Install, upgrade, and manage the Quantum Orchestrator operator. |
| **Console Plugin** | Custom pages in the OpenShift Web Console: quantum jobs, QPU backends, QEC status. |
| **NVIDIA GPU Operator** | GPU device plugin, driver management, MIG, NCCL. |
| **AMD GPU Operator** | AMD GPU device plugin, ROCm, MI355X support. |
| **RDMA / Multus** | Low-latency interconnect for NVQLink and IBM QEC co-design. |
| **Dynamic Accelerator Slicer (DAS)** | GPU partitioning for multi-tenant quantum workloads. |
| **Node Feature Discovery** | Auto-detect GPU models, RDMA NICs, and accelerator hardware. |
| **GitOps (Argo CD / Flux)** | Declarative deployment of the operator and configs. |
| **External Secrets Operator** | Sync IBM Cloud, AWS, Azure credentials from vaults. |

### Integration Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│  OpenShift AI Pipeline (Kubeflow Pipelines 2.0)                        │
│                                                                         │
│  ┌─────────────┐   ┌──────────────────┐   ┌────────────────────────┐   │
│  │ Step 1:     │   │ Step 2:          │   │ Step 3:                │   │
│  │ Data Prep   │──▶│ Quantum Job      │──▶│ Model Training         │   │
│  │ (GPU)       │   │ (QPU + GPU)      │   │ (Kubeflow Trainer v2)  │   │
│  │             │   │                  │   │                        │   │
│  │ Load        │   │ VQE on           │   │ Train AI model         │   │
│  │ molecule    │   │ Nighthawk r2     │   │ on quantum-enhanced    │   │
│  │ data        │   │ + AMD MI355X     │   │ data (energy, props)   │   │
│  └─────────────┘   └──────────────────┘   └────────────────────────┘   │
│                                                                         │
│  ┌─────────────┐   ┌──────────────────┐   ┌────────────────────────┐   │
│  │ Step 4:     │   │ Step 5:          │   │ Step 6:                │   │
│  │ Evaluation  │──▶│ Model Serving    │──▶│ Monitoring             │   │
│  │ (GPU)       │   │ (vLLM, multi-    │   │ (Prometheus, OTel,     │   │
│  │             │   │  node GPU)       │   │  OpenShift Console)    │   │
│  └─────────────┘   └──────────────────┘   └────────────────────────┘   │
│                                                                         │
│  Kueue co-schedules: GPU quotas (AI) + QPU quotas (quantum)            │
│  Per-tenant isolation via QuantumProviderBinding + Hardware Profiles   │
└─────────────────────────────────────────────────────────────────────────┘
```

### Console Plugin

The Quantum Orchestrator provides an OpenShift Console Plugin that adds:

- **Perspective**: "Quantum Computing" (alongside Admin, Developer, AI)
- **Pages**: Quantum Jobs, QPU Backends, NVQLink Logical QPUs, QEC Status
- **Tabs**: Provider, Status, Results, Cost, Logs (on QuantumJob pages)
- **Actions**: Submit, Cancel, Re-run, Download Results

```yaml
# ConsolePlugin CR (created by the operator)
apiVersion: console.openshift.io/v1
kind: ConsolePlugin
metadata:
  name: quantum-orchestrator-plugin
  namespace: openshift-console
spec:
  displayName: Quantum Orchestrator
  backend:
    type: Service
    service:
      name: quantum-orchestrator-console
      port: 9001
      routePath: /
```

### GPU / Accelerator Scheduling

```yaml
# OpenShift Hardware Profile for quantum workloads
apiVersion: console.openshift.io/v1
kind: HardwareProfile
metadata:
  name: quantum-gpu
spec:
  nodeSelector:
    nvidia.com/gpu.present: "true"
  tolerations:
    - key: nvidia.com/gpu
      operator: Exists
  gpus:
    - type: nvidia
      count: 1
      minMemoryGiB: 40
```

### RDMA for NVQLink / QEC

```yaml
# Multus RDMA network for NVQLink interconnect
apiVersion: multus-cni.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: nvqlink-rdma
  namespace: quantum-system
spec:
  config: '{
    "cniVersion": "0.3.1",
    "type": "macvlan",
    "master": "net1",
    "mode": "bridge"
  }'
# Pod annotation:
#   k8s.v1.cni.cncf.io/networks: nvqlink-rdma
```

## Data Flow

```
User → quantumctl / kubectl → QuantumJob CRD
  → Operator Reconcile Loop
    → Provider Adapter
      ├── IBM:      IBM Cloud IAM → Quantum Compute Service REST
      ├── Braket:   IRSA → AWS Braket API → S3
      ├── Azure:    Workload Identity → Azure Quantum REST → Blob
      ├── NVIDIA:   NGC container → GPU (in-cluster) → PVC/S3
      ├── Direct:   API key → Provider REST
      └── qBraid:   Sidecar → qBraid SDK → Any provider
    → Poll Status
    → Update CRD Status
    → Fetch Results → S3 / Blob / MinIO / PVC
```

## Multi-Tenancy

- `QuantumProviderBinding` maps a namespace/tenant to a cloud account
- The operator checks the binding before submitting
- Credentials are injected via K8s Secrets (ESO / Vault)
- Kueue enforces per-tenant QPU quotas

## QEC / HPC Integration Patterns

| Pattern | QPU | Classical | Interconnect | Decoder |
|---------|-----|-----------|-------------|---------|
| IBM RelayBP | Nighthawk r2 | AMD MI355X + EPYC (FPGA) | Co-located | RelayBP |
| NVIDIA NVQLink | Quantinuum Helios | NVIDIA GH200 | RoCE (3.84μs) | nv-qldpc, BP+OSD |
| ORNL co-design | Nighthawk r2 | AMD MI355X + EPYC (Umbra) | Co-located | Algorithm-aware |
| CUDA-QX | Any (via Braket) | NVIDIA GPU | In-cluster | Tensor-network, BP+OSD |

## References

### IBM
- [IBM Quantum Compute Service REST API](https://quantum.cloud.ibm.com/docs/en/api/qiskit-runtime-rest)
- [IBM Cloud IAM](https://cloud.ibm.com/apidocs/iam-identity-token-api)
- [Qiskit v2.5 Release](https://www.ibm.com/quantum/blog/qiskit-2-5-release-summary)
- [ORNL + IBM + AMD QEC Platform](https://newsroom.ibm.com/2026-11-05-oak-ridge-national-lab,-ibm,-and-amd-launch-new-co-design-platform-for-quantum-error-correction)

### NVIDIA
- [CUDA-Q NGC Container](https://catalog.ngc.nvidia.com/orgs/nvidia/teams/quantum/containers/cuda-quantum)
- [CUDA-Q Documentation](https://nvidia.github.io/cuda-quantum/latest/)
- [cuQuantum SDK](https://docs.nvidia.com/cuda/cuquantum/latest/)
- [NVQLink Architecture](https://developer.nvidia.com/blog/nvidia-nvqlink-architecture-integrates-accelerated-computing-with-quantum-processors)
- [CUDA-Q GitHub](https://github.com/NVIDIA/cuda-quantum)

### CNCF / Kubernetes
- [Kueue](https://github.com/kubernetes-sigs/kueue)
- [Argo Workflows](https://github.com/argoproj/argo-workflows)
- [External Secrets Operator](https://github.com/external-secrets/external-secrets)

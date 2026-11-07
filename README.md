# Quantum Orchestrator

> Cloud-native orchestration of quantum workloads on Kubernetes — provider-agnostic, multi-cloud, hybrid-aware.

Quantum Orchestrator is a Kubernetes Operator that schedules quantum computing jobs (simulations, QPU runs, hybrid quantum-classical pipelines) as first-class workloads alongside CPU/GPU workloads. It supports **AWS Braket**, **Azure Quantum**, **direct QPU providers** (IonQ, Rigetti, Quantinuum, IQM), and **local simulators** through a single, stable CRD API.

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                     Kubernetes / OpenShift Cluster                   │
│                                                                      │
│  ┌──────────┐    ┌───────────────────────────────────────────────┐   │
│  │ quantumctl│───▶│          Quantum Operator (Controller)        │   │
│  │  / YAML  │    │   ┌───────────┐  ┌─────────────────────────┐ │   │
│  └──────────┘    │   │ CRD Loop  │  │ Scheduling (Kueue)      │ │   │
│                  │   └─────┬─────┘  └─────────────────────────┘ │   │
│  ┌──────────┐    │         │                                     │   │
│  │  Web UI  │◀───│  ┌──────▼──────────────────────────────────┐ │   │
│  └──────────┘    │  │  Provider Adapter Layer                 │ │   │
│                  │  │  Braket │ Azure Quantum │ Direct │ qBraid│ │   │
│  ┌──────────┐    │  │  ┌──────▼──────────────────────────────┐ │ │   │
│  │Transpiler│◀───│  │  │ IR / Transpilation (QASM3, QIR)     │ │ │   │
│  │ (sidecar)│    │  │  └─────────────────────────────────────┘ │ │   │
│  └──────────┘    │  └──────────────────────────────────────────┘ │   │
│                  └───────────────────────────────────────────────┘   │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │ Kueue │ Argo Workflows │ Prometheus │ OpenTelemetry           │  │
│  └────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────┘
         │                    │                      │
         ▼                    ▼                      ▼
   ┌───────────┐      ┌──────────────┐       ┌────────────────┐
   │  AWS      │      │  Azure       │       │  Direct QPUs   │
   │  Braket   │      │  Quantum     │       │  IonQ/Rigetti/ │
   │  + S3     │      │  + Blob      │       │  Quantinuum/…  │
   └───────────┘      └──────────────┘       └────────────────┘
```

## Quick Start

### Prerequisites

- Kubernetes 1.28+ (or OpenShift 4.14+)
- `kubectl` v1.28+
- A quantum provider account (AWS Braket, Azure Quantum, or direct)

### Install with Helm

```bash
helm repo add quantum-orchestrator https://charts.quantum-orchestrator.io
helm install quantum-orchestrator quantum-orchestrator/quantum-orchestrator \
  --namespace quantum-system --create-namespace
```

### Submit a Quantum Job

```yaml
# quantum-job.yaml
apiVersion: quantum.orchestrator.io/v1alpha1
kind: QuantumJob
metadata:
  name: bell-state
spec:
  circuit: |
    OPENQASM 3.0;
    include "stdgates.inc";
    qubit[2] q;
    bit[2] c;
    h q[0];
    cx q[0], q[1];
    measure q[0] -> c[0];
    measure q[1] -> c[1];
  target:
    provider: aws-braket
    device: sv1
  shots: 1024
  credentials:
    secretRef:
      name: braket-creds
```

```bash
kubectl apply -f quantum-job.yaml
kubectl get quantumjobs
kubectl logs -f quantumjob/bell-state
```

## CRDs

| CRD | Purpose |
|-----|---------|
| `QuantumJob` | Primary workload: circuit, target, shots, hybrid config, credentials |
| `QuantumSession` | Long-lived QPU session (Azure Sessions / Braket reserved time) |
| `QPUBackend` | Device inventory: provider, qubits, gate-set, error rates, cost |
| `QuantumProviderBinding` | Tenant → cloud account / workspace mapping |
| `QuantumSimulator` | Local/managed simulator as a Deployment |

## Provider Support

| Provider | Circuit Formats | Hybrid Workflows | Auth |
|----------|----------------|-----------------|------|
| AWS Braket | OpenQASM 2/3, Braket IR | Braket Hybrid Jobs | IRSA / IAM |
| Azure Quantum | QIR, QASM, Quil, Q# | Sessions, Integrated Hybrid (QIR) | Workload Identity |
| IonQ (direct) | OpenQASM 2/3 | — | API key |
| Rigetti (direct) | Quil, QIR, OpenQASM | — | API key |
| Quantinuum (direct) | QIR, OpenQASM | — | API key |
| IBM Quantum | OpenQASM 3, Qiskit v2.5, PBC, Clifford+T | RelayBP QEC, HPC error mitigation, dynamic circuits | IBM Cloud API |
| NVIDIA CUDA-Q | OpenQASM 3, CUDA-Q kernels | CUDA-Q → Braket, GPU sim | nvidia.com/gpu |
| NVIDIA cuQuantum | OpenQASM 3, Qiskit, Cirq, PennyLane, Qibo | Multi-GPU / Multi-node | nvidia.com/gpu |
| NVIDIA NVQLink | CUDA-Q, CUDA-QX QEC | Real-time QEC (QPU + GPU) | RoCE / RDMA |
| Local (simulator) | OpenQASM 3 | — | — |

## Development

```bash
# Build the operator
make build

# Run tests
make test

# Deploy to a dev cluster
make deploy

# Build the transpiler sidecar
cd transpiler && docker build -t quantum-transpiler:dev .

# Run the CLI
make quantumctl
./bin/quantumctl submit -f examples/bell-state.yaml
```

## Project Structure

```
quantum-orchestrator/
├── api/v1alpha1/          # CRD type definitions
├── cmd/
│   ├── operator/          # Operator entrypoint
│   └── quantumctl/        # CLI tool
├── internal/
│   ├── adapter/           # Provider adapters (Braket, Azure, Direct, qBraid)
│   ├── controller/        # Reconcile loops
│   ├── scheduler/         # Kueue integration
│   └── telemetry/         # Prometheus + OTel metrics
├── transpiler/            # Circuit IR conversion sidecar (Python)
├── notebooks/             # Pre-built quantum dev environment images
├── charts/                # Helm chart
├── config/                # Kustomize overlays (CRD, RBAC, manager)
├── docs/                  # Architecture, guides, provider docs
├── test/                  # E2E + benchmark tests
└── .github/               # CI/CD workflows
```

## Ecosystem

This project integrates with:

- **Scheduling**: [Kueue](https://github.com/kubernetes-sigs/kueue), [Argo Workflows](https://github.com/argoproj/argo-workflows)
- **Quantum SDKs**: [Qiskit](https://github.com/Qiskit/qiskit), [Cirq](https://github.com/quantumlib/Cirq), [PennyLane](https://github.com/PennyAI/PennyLane), [QDK](https://github.com/microsoft/quantum), [PyQuil](https://github.com/rigetti/pyquil)
- **Cross-platform**: [qBraid](https://github.com/qBraid/qBraid), [OpenQASM](https://github.com/openqasm/openqasm), [QIR](https://github.com/qir-alliance/qir)
- **NVIDIA**: [CUDA-Q](https://nvidia.github.io/cuda-quantum/), [cuQuantum](https://github.com/NVIDIA/cuQuantum), [CUDA-QX](https://nvidia.github.io/cudaqx/), [NVQLink](https://www.nvidia.com/en-us/solutions/quantum-computing/nvqlink/)
- **IBM**: [Qiskit](https://github.com/Qiskit/qiskit) (v2.5, C API, PBC), [IBM Quantum Compute Service](https://quantum.cloud.ibm.com), [RelayBP](https://www.ibm.com/quantum/blog/qdc-2025), [Qiskit Algorithms](https://github.com/qiskit-community/qiskit-algorithms)
- **Observability**: [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-specification), [Prometheus](https://github.com/prometheus/prometheus), [Grafana](https://github.com/grafana/grafana)
- **Secrets**: [External Secrets Operator](https://github.com/external-secrets/external-secrets)

## IBM Integration

| Component | Description |
|-----------|-------------|
| Qiskit v2.5 | Quantum SDK (Python + C API), PBC, Clifford+T, dynamic circuits |
| Nighthawk r2 | 120Q + 458 physical, square lattice, high-speed reset |
| Heron r3 | 156Q, heavy-hexagonal, 330K CLOPS |
| Starling | Fault-tolerant QC (2029 target), 100+ logical qubits |
| RelayBP | Real-time QEC decoder (AMD FPGA) |
| HPC Error Mitigation | 100x cost reduction, GPU/CPU accelerated |
| ORNL+IBM+AMD | QEC co-design platform (Nighthawk + MI355X + EPYC) |

## NVIDIA Integration

| Component | Description |
|-----------|-------------|
| CUDA-Q | Quantum SDK (C++/Python) with GPU-accelerated simulators, Braket integration |
| cuQuantum | GPU libraries: cuStateVec, cuDensityMat, cuTensorNet, cuPauliProp, cuStabilizer |
| CUDA-QX | Quantum error correction: BP+OSD, tensor-network, nv-qldpc decoders |
| NVQLink | Real-time QPU-GPU coupling (RoCE/RDMA), cudaq::device_call, Logical QPU model |

## Prior Art

- [Qubernetes](https://github.com/q8s/q8s) — quantum jobs on K8s
- [IBM Quantum Operator](https://github.com/redhat-et/ibm-quantum-operator) — Qiskit on IBM Q
- [quantum-dev](https://github.com/acuto/quantum-dev) — quantum dev environments

## License

[Apache 2.0](LICENSE)

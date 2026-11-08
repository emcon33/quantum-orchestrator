# NVIDIA CUDA-Q / cuQuantum / NVQLink Provider Guide

## Overview

NVIDIA's quantum stack provides GPU-accelerated quantum simulation,
hybrid quantum-classical computing, and real-time quantum error
correction. The operator supports four NVIDIA workload types:

| Workload | SDK | GPU | Use Case |
|----------|-----|-----|----------|
| CUDA-Q simulation | `cudaq` | A100 / H100 / B200 | Circuit simulation, VQE, QAOA |
| cuQuantum distributed | `cuquantum` | Multi-GPU / Multi-node | Large-scale statevector, stabilizer, tensor network |
| CUDA-Q → Braket | `cudaq` + Braket | GPU + Cloud QPU | Hybrid: GPU classical + IonQ/Rigetti/IQM QPU |
| NVQLink real-time QEC | `cudaqx` | GH200 + QPU | Real-time surface code / LDPC decoding |

## Authentication

### In-Cluster (GPU workloads)

No cloud credentials needed. The operator creates K8s Jobs with
`nvidia.com/gpu` resource requests. Requires:

- NVIDIA GPU Operator or Device Plugin installed
- GPU nodes labeled `nvidia.com/gpu.present=true`
- `nvidia.com/gpu` resource available in the cluster

```yaml
nvidia:
  gpu:
    model: h100
    count: 1
    minMemoryGiB: 80
```

### Cloud QPU (via CUDA-Q → Braket)

Use standard AWS Braket credentials (IRSA or IAM):

```yaml
credential:
  secretRef:
    name: braket-creds
nvidia:
  cudaqTarget: "braket"
  cudaqKernel:
    inline: |
      import cudaq
      device = "arn:aws:braket:us-east-1::device/qpu/ionq/Aria-1-25Q"
      cudaq.set_target("braket", machine=device)
      # ... kernel code ...
```

### NVQLink (QPU + GPU)

Requires both the QPU provider credentials and GPU access:

```yaml
credential:
  secretRef:
    name: quantinuum-helios-creds
nvidia:
  nvqlinkRef:
    name: surface-code-d7  # references an NVQLinkLogicalQPU CRD
```

## Supported CUDA-Q Targets

| Target | Description |
|--------|-------------|
| `nvidia` | Single-GPU CUDA-Q simulator |
| `nvidia/multi` | Multi-GPU CUDA-Q simulator |
| `braket` | AWS Braket QPU (IonQ, Rigetti, IQM, QuEra) |
| `quantinuum` | Quantinuum QPU (direct or via NVQLink) |
| `ionq` | IonQ QPU (direct) |

## cuQuantum Engines

| Engine | Library | Max Qubits (H100) | Notes |
|--------|---------|-------------------|-------|
| `statevec` | cuStateVec | ~28 | Statevector simulation |
| `densitymat` | cuDensityMat | ~22 | Density matrix (mixed states) |
| `tensornet` | cuTensorNet | ~35 | Tensor network contraction |
| `pauliprop` | cuPauliProp | ~40+ | Pauli propagation (v25.11) |
| `stabilizer` | cuStabilizer | ~50+ | Stabilizer / Clifford circuits |

### Multi-GPU / Multi-Node

```yaml
nvidia:
  cuQuantum:
    engine: statevec
    gpus: 4        # 4x H100
    nodes: 2       # 2 nodes → 8 GPUs total
    maxQubits: 35
```

## CUDA-QX QEC Decoders

| Decoder | Description |
|---------|-------------|
| `nv-qldpc-decoder` | GPU-accelerated LDPC decoder (best for large codes) |
| `bp-osd` | Belief Propagation + OSD (adaptive convergence) |
| `tensor-network` | Exact ML decoding via tensor network (small codes) |
| `union-find` | Classical union-find (baseline) |

### Real-Time QEC via NVQLink

```yaml
nvidia:
  sdk: cudaqx
  qec:
    decoder: nv-qldpc-decoder
    code: surface
    distance: 7
    realTime: true       # streaming decode via NVQLink
    historyDepth: 20
  nvqlinkRef:
    name: surface-code-d7
```

The NVQLink Logical QPU CRD defines the QPU + GPU + interconnect:

```yaml
kind: NVQLinkLogicalQPU
spec:
  qpu:
    provider: quantinuum
    deviceId: quantinuum.helios
    physicalQubits: 98
  realTimeHost:
    gpu: gh200
    gpus: 1
  interconnect:
    technology: roce
    targetLatencyUs: 5
    bandwidthGbps: 200
    nic: connectx-7
```

## GPU Resource Scheduling

The operator requests GPUs via the standard Kubernetes resource
`nvidia.com/gpu`. For multi-GPU workloads:

```yaml
nvidia:
  gpu:
    model: h100
    count: 4
```

This maps to:

```yaml
resources:
  limits:
    nvidia.com/gpu: 4
```

### Kueue Integration

Define a Kueue `ResourceFlavor` for each GPU type:

```yaml
apiVersion: kueue.x-k8s.io/v1beta1
kind: ResourceFlavor
metadata:
  name: nvidia-h100
spec:
  resource:
    nvidia.com/gpu: "1"
  nodeLabel:
    nvidia.com/gpu.product: "NVIDIA-H100"
```

Then assign per-tenant QPU+GPU quotas in the `ClusterQueue`.

## Framework Integration

cuQuantum integrates with (zero code changes):

- Qiskit
- Cirq
- PennyLane
- Qibo
- QuTiP
- QuEST
- TKET
- Torch Quantum
- XACC
- Gruppy
- ExaTN
- AWS Braket

## NVIDIA NGC Container Library

NVIDIA distributes CUDA-Q and cuQuantum via the **NGC (NVIDIA GPU Cloud)**
container registry. All containers are signed, scanned, and support multi-node.

### Available Images

| Image | Tag | CUDA | Size | Notes |
|-------|-----|------|------|-------|
| `nvcr.io/nvidia/quantum/cuda-quantum` | `cu13-0.16.0` | CUDA 13 | ~4.3 GB | **Latest stable** |
| `nvcr.io/nvidia/quantum/cuda-quantum` | `cu12-0.14.2` | CUDA 12 | ~4.3 GB | CUDA 12 compatible |
| `nvcr.io/nvidia/nightly/cuda-quantum` | `cu13-latest` | CUDA 13 | ~4.3 GB | Nightly / pre-release |
| `ghcr.io/quantum-orchestrator/notebook-nvidia` | `latest` | — | — | Jupyter + CUDA-Q + cuQuantum |
| `ghcr.io/quantum-orchestrator/notebook-nvidia-qec` | `latest` | — | — | Jupyter + CUDA-QX + stim |

### Container Contents

The NGC `cuda-quantum` image includes:

- **CUDA-Q**: Quantum SDK (C++ + Python), NVQ++ compiler (LLVM-based)
- **cuQuantum**: cuStateVec, cuDensityMat, cuTensorNet, cuPauliProp, cuStabilizer
- **CUDA-QX**: QEC decoders (BP+OSD, tensor-network, nv-qldpc)
- **Framework integrations**: Qiskit, Cirq, PennyLane, Qibo, QuTiP, TKET, Torch Quantum, XACC
- **Multi-node**: Yes (NCCL-based multi-GPU / multi-node)
- **Multi-arch**: amd64

### Pulling the Image

```bash
# 1. Register at developer.nvidia.com/ngc (free)
# 2. Log in to NGC
docker login nvcr.io
#    Username: $oauthtoken (or your NGC user)
#    Password: <NGC API key>

# 3. Pull the stable image (CUDA 13)
docker pull nvcr.io/nvidia/quantum/cuda-quantum:cu13-0.16.0

# 4. Run with GPU
docker run --gpus all -it nvcr.io/nvidia/quantum/cuda-quantum:cu13-0.16.0
```

### In Kubernetes

```bash
# Create a pull secret for NGC
kubectl create secret docker-registry ngc-pull-secret \
  --docker-server=nvcr.io \
  --docker-username=$NGC_USER \
  --docker-password=$NGC_KEY \
  -n quantum-system
```

```yaml
# The operator uses this in the Pod spec:
spec:
  imagePullSecrets:
    - name: ngc-pull-secret
  containers:
    - name: cuda-q
      image: nvcr.io/nvidia/quantum/cuda-quantum:cu13-0.16.0
      resources:
        limits:
          nvidia.com/gpu: 1
```

### Multi-Node GPU (cuQuantum Distributed)

```yaml
nvidia:
  cuQuantum:
    engine: statevec
    gpus: 4        # 4 GPUs per node
    nodes: 2       # 2 nodes → 8 GPUs total
    maxQubits: 35
  gpu:
    model: h100
    count: 4
```

The NGC image supports multi-node out of the box. Use Kueue Gang Scheduling
or Volcano for co-scheduling across nodes. NCCL handles inter-node GPU communication.

### Source Code

- **GitHub**: [github.com/NVIDIA/cuda-quantum](https://github.com/NVIDIA/cuda-quantum) (Apache 2.0)
- **NGC Catalog**: [catalog.ngc.nvidia.com](https://catalog.ngc.nvidia.com/orgs/nvidia/teams/quantum/containers/cuda-quantum)
- **Docs**: [nvidia.github.io/cuda-quantum](https://nvidia.github.io/cuda-quantum/latest/)

## References

- [CUDA-Q Documentation](https://nvidia.github.io/cuda-quantum/latest/)
- [cuQuantum SDK](https://docs.nvidia.com/cuda/cuquantum/latest/)
- [CUDA-QX QEC](https://nvidia.github.io/cudaqx/components/qec/introduction.html)
- [NVQLink Architecture](https://developer.nvidia.com/blog/nvidia-nvqlink-architecture-integrates-accelerated-computing-with-quantum-processors)
- [CUDA-Q + Braket Integration](https://developer.nvidia.com/blog/accelerated-quantum-supercomputing-with-the-nvidia-cuda-q-and-amazon-braket-integration)
- [cuQuantum v25.11](https://developer.nvidia.com/blog/advanced-large-scale-quantum-simulation-techniques-in-cuquantum-sdk-v25-11)
- [NVQLink + Quantinuum Helios](https://www.quantinuum.com/blog/introducing-helios-the-most-accurate-quantum-computer-in-the-world)
- [NVQLink Paper](https://arxiv.org/abs/2510.25213)

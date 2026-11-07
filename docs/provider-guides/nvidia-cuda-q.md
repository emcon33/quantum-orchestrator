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

## Container Images

| Image | Contents |
|-------|----------|
| `nvcr.io/nvidia/cuda-quantum:latest` | CUDA-Q + cuQuantum + CUDA-QX |
| `ghcr.io/quantum-orchestrator/notebook-nvidia:latest` | CUDA-Q + cuQuantum + Jupyter |
| `ghcr.io/quantum-orchestrator/notebook-nvidia-qec:latest` | CUDA-QX + stim + pymatching + Jupyter |

## References

- [CUDA-Q Documentation](https://nvidia.github.io/cuda-quantum/latest/)
- [cuQuantum SDK](https://docs.nvidia.com/cuda/cuquantum/latest/)
- [CUDA-QX QEC](https://nvidia.github.io/cudaqx/components/qec/introduction.html)
- [NVQLink Architecture](https://developer.nvidia.com/blog/nvidia-nvqlink-architecture-integrates-accelerated-computing-with-quantum-processors)
- [CUDA-Q + Braket Integration](https://developer.nvidia.com/blog/accelerated-quantum-supercomputing-with-the-nvidia-cuda-q-and-amazon-braket-integration)
- [cuQuantum v25.11](https://developer.nvidia.com/blog/advanced-large-scale-quantum-simulation-techniques-in-cuquantum-sdk-v25-11)
- [NVQLink + Quantinuum Helios](https://www.quantinuum.com/blog/introducing-helios-the-most-accurate-quantum-computer-in-the-world)
- [NVQLink Paper](https://arxiv.org/abs/2510.25213)

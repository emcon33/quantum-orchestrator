# AI Augmented by Quantum Computing

Quantum computing and AI are converging. This document describes the use cases
where quantum processing units (QPUs) augment classical AI workloads, and how
the Quantum Orchestrator supports each one.

## Why Quantum + AI?

Classical AI hits walls at specific problem classes:

| Problem | Why Classical Struggles | How Quantum Helps |
|---------|------------------------|-------------------|
| Molecular simulation | Hilbert space grows as 2^N qubits | Quantum circuits natively represent quantum states |
| Combinatorial optimization | NP-hard; exponential search space | QAOA / quantum annealing explore superposition |
| Linear algebra (matrix inversion) | O(n³) classically | HHL algorithm: O(log n) with fault-tolerant QC |
| Kernel methods | Kernel matrix is O(n²) | Quantum kernels: exponential feature space in polynomial time |
| Sampling from complex distributions | MCMC is slow for multimodal targets | Quantum generative models sample directly |
| Error correction decoding | Real-time decoding is latency-critical | GPU/FPGA-accelerated quantum decoders |

The Quantum Orchestrator schedules these hybrid workloads as first-class
Kubernetes objects, routing the quantum portion to a QPU (IBM, AWS, Azure,
NVIDIA, direct) and the classical AI portion to GPU/CPU nodes.

## Use Cases

### 1. Quantum Machine Learning (QML)

**Description:** Train quantum neural networks (variational quantum circuits)
for classification, regression, and generative tasks. The classical optimizer
runs on GPU; the quantum feature map and variational layers run on a QPU.

**Why it works:** Quantum circuits can represent exponentially large feature
spaces with polynomially many parameters. Quantum kernels access feature spaces
inaccessible to classical kernel methods.

**Providers:**
- IBM Nighthawk / Heron (120-156 qubits for deep ansätze)
- NVIDIA CUDA-Q (GPU-accelerated simulation for training)
- AWS Braket (IonQ, Rigetti for QPU inference)

**Operator configuration:**

```yaml
apiVersion: quantum.orchestrator.io/v1alpha1
kind: QuantumJob
metadata:
  name: qml-classifier
  labels:
    use-case: qml
    model: variational-classifier
spec:
  target:
    provider: ibm          # QPU for quantum layers
    device: ibm_nighthawk_r2
  shots: 2048
  credential:
    secretRef:
      name: ibm-quantum-creds
  hybrid:
    image: ghcr.io/my-org/qml-trainer:latest   # Classical optimizer (GPU)
    entrypoint:
      - python
      - -m
      - qml_trainer
      - --epochs
      - "100"
      - --optimizer
      - adam
      - --batch-size
      - "32"
    resources:
      limits:
        nvidia.com/gpu: "1"
        memory: 16Gi
      requests:
        nvidia.com/gpu: "1"
        memory: 8Gi
    env:
      - name: QPU_BACKEND
        value: "ibm_nighthawk_r2"
      - name: CLASSICAL_BACKEND
        value: "gpu"
  ibm:
    qiskit:
      version: "2.5"
      compilationMode: standard
      dynamicCircuits: true
    errorMitigation:
      strategy: zero-noise
      hpcPowered: true
  result:
    backend: s3
    bucket: qml-results
    prefix: classifier/
```

**Workflow:**
```
┌─────────────────────────────────────────────────────────────────────┐
│  QML Training Loop (per epoch)                                      │
│                                                                     │
│  ┌──────────────────┐    ┌──────────────────────────────────┐      │
│  │  Classical GPU    │    │  QPU (IBM Nighthawk / IonQ)      │      │
│  │                  │    │                                  │      │
│  │  1. Forward pass │───▶│  2. Quantum feature map          │      │
│  │     (data →      │    │     (Hadamard test / amplitude)  │      │
│  │      amplitudes) │◀───│  3. Variational layer            │      │
│  │                  │    │     (parameterized gates)        │      │
│  │  4. Backward     │    │                                  │      │
│  │     pass (GPU)   │    │                                  │      │
│  │                  │    │                                  │      │
│  │  5. Update       │    │                                  │      │
│  │     parameters   │    │                                  │      │
│  └──────────────────┘    └──────────────────────────────────┘      │
│                                                                     │
│  Kueue schedules: GPU quota for classical, QPU quota for quantum    │
└─────────────────────────────────────────────────────────────────────┘
```

---

### 2. Quantum-Enhanced Optimization for AI

**Description:** Use QAOA (Quantum Approximate Optimization Algorithm) to
solve combinatorial sub-problems in AI pipelines: neural architecture search
(NAS), hyperparameter optimization, feature selection, and scheduling.

**Why it works:** QAOA encodes combinatorial problems into quantum circuits.
The quantum processor explores the solution space in superposition, and the
classical optimizer (GPU) adjusts the QAOA parameters. This is a natural
hybrid quantum-classical loop.

**Providers:**
- NVIDIA CUDA-Q + cuQuantum (GPU-accelerated simulation for large problem instances)
- IBM Nighthawk (120 qubits for large combinatorial problems)
- AWS Braket (Rigetti, IQM for QPU execution)

**Operator configuration:**

```yaml
apiVersion: quantum.orchestrator.io/v1alpha1
kind: QuantumJob
metadata:
  name: qaoa-nas
  labels:
    use-case: optimization
    algorithm: qaoa
    application: neural-architecture-search
spec:
  target:
    provider: nvidia        # GPU-accelerated QAOA
    device: nvidia
    simulator: true
  shots: 4096
  credential:
    secretRef:
      name: nvidia-gpu-creds
  nvidia:
    sdk: cudaq
    cudaqTarget: "nvidia"
    cudaqKernel:
      language: python
      inline: |
        import cudaq
        # QAOA for neural architecture search
        # Problem: select optimal layer types, widths, connections
        # Encoding: binary decision per architecture choice
        problem_size = 20  # binary decisions
        p = 4              # QAOA depth

        @cudaq.kernel
        def qaoa_nas(theta: list[float]):
            q = cudaq.qvector(problem_size)
            # Initialize superposition
            for i in range(problem_size):
                h(q[i])
            # QAOA: cost unitary + mixer unitary
            for layer in range(p):
                # Cost Hamiltonian (architecture fitness)
                # Mixer Hamiltonian (X rotations)
                ...
            mz(q)

        # Classical optimizer adjusts theta
        best_theta = optimize(qaoa_nas, shots=4096)
        best_architecture = decode(best_theta)
        return best_architecture
    gpu:
      model: h100
      count: 1
      minMemoryGiB: 80
  result:
    backend: minio
    bucket: qml-results
    prefix: qaoa-nas/
```

**AI applications for QAOA:**

| AI Sub-problem | QAOA Encoding | Qubits |
|---------------|---------------|--------|
| Neural Architecture Search | Binary: include/exclude each layer choice | 10-50 |
| Hyperparameter Optimization | Discrete: grid selection per param | 20-80 |
| Feature Selection | Binary: include/exclude each feature | 10-100 |
| Training Data Scheduling | Assignment: which batch on which GPU | 20-60 |
| Model Compression | Binary: prune/keep each connection | 50-200 |
| Cluster Scheduling | Assignment: job → node | 20-100 |

---

### 3. Quantum Chemistry for AI-Driven Drug Discovery

**Description:** Use VQE (Variational Quantum Eigensolver) to compute molecular
ground-state energies. The quantum processor simulates the molecular Hamiltonian;
the classical AI (GPU) optimizes the ansatz parameters and feeds results into
molecular property prediction models.

**Why it works:** Molecular electronic structure is a quantum problem. Classical
methods (DFT, CCSD) scale polynomially but with large constants. Quantum
simulators scale naturally with molecule size. VQE bridges the two: quantum
processor for the Hamiltonian, classical GPU for the optimizer.

**Providers:**
- IBM Nighthawk / Heron (120-156 qubits for medium molecules)
- NVIDIA cuQuantum (GPU-accelerated for large Hamiltonians)
- AWS Braket (IonQ, Rigetti for QPU)
- Azure Quantum (Quantinuum for high-fidelity)

**Operator configuration:**

```yaml
apiVersion: quantum.orchestrator.io/v1alpha1
kind: QuantumJob
metadata:
  name: vqe-h2o-drug
  labels:
    use-case: quantum-chemistry
    application: drug-discovery
    molecule: h2o
spec:
  circuit:
    inline: |
      OPENQASM 3.0;
      include "stdgates.inc";
      // H2O molecule VQE ansatz (18 qubits, 2-electron)
      qubit[18] q;
      bit[18] c;
      // Fermion→qubit mapping (JW or Bravyi-Kitaev)
      // UCCSD ansatz: excitations + parameterized gates
      x q[0]; x q[1];  // occupied orbitals
      rx(theta[0]) q[2]; rx(theta[1]) q[3];
      cx q[2], q[3]; rz(theta[2]) q[3]; cx q[2], q[3];
      // ... 18 qubit UCCSD ansatz ...
      measure q[0] -> c[0];
      measure q[17] -> c[17];
    format: openqasm3
  target:
    provider: ibm
    device: ibm_nighthawk_r2
  shots: 8192
  parameters:
    # UCCSD amplitudes (classical optimizer updates these)
    theta[0]: "0.0"
    theta[1]: "0.0"
  credential:
    secretRef:
      name: ibm-quantum-creds
  hybrid:
    image: ghcr.io/my-org/vqe-optimizer:latest
    entrypoint:
      - python
      - -m
      - vqe
      - --molecule
      - h2o
      - --ansatz
      - uccsd
      - --iterations
      - "500"
      - --optimizer
      - spsa
      - --downstream
      - property-prediction  # Feed into AI model
    resources:
      limits:
        nvidia.com/gpu: "2"
        memory: 32Gi
      requests:
        nvidia.com/gpu: "1"
        memory: 16Gi
  ibm:
    errorMitigation:
      strategy: zero-noise
      hpcPowered: true
      extrapolationOrders: [1, 2, 3]
  result:
    backend: s3
    bucket: drug-discovery
    prefix: h2o-vqe/
```

**Pipeline:**
```
Molecule (SMILES) → Hamiltonian → VQE (QPU + GPU) → Energy
  → Property Prediction (AI model) → Drug Score → Ranking
```

---

### 4. AI-Assisted Quantum Error Correction

**Description:** Use classical ML/AI models to decode quantum error syndromes
in real time. The AI model (trained on GPU) predicts error locations from
syndrome measurements, enabling faster and more accurate decoding than
traditional methods.

**Why it works:** QEC decoding is a pattern recognition problem. Neural networks
and other ML models excel at this. The quantum processor generates syndromes;
the AI decoder runs on GPU/FPGA with microsecond latency.

**Providers:**
- IBM Nighthawk r2 + AMD FPGA (RelayBP decoder)
- NVIDIA GH200 + NVQLink (nv-qldpc, BP+OSD decoders)
- NVIDIA CUDA-QX (tensor-network, BP+OSD)

**Operator configuration:**

```yaml
# NVQLink: Real-time QEC with AI decoder
apiVersion: quantum.orchestrator.io/v1alpha1
kind: QuantumJob
metadata:
  name: qec-ai-decoder
  labels:
    use-case: qec
    decoder: relaybp
    ai-augmented: "true"
spec:
  target:
    provider: ibm
    device: ibm_nighthawk_r2
  shots: 100000
  credential:
    secretRef:
      name: ibm-quantum-creds
  ibm:
    qec:
      decoder: relaybp
      decoderHardware: fpga       # AMD FPGA for real-time
      code: surface
      distance: 7
      logicalQubits: 70
      hardOperations: 438
      realTime: true
      algorithmAware: true        # AI-aware decoding
      hardwareAwareCompilation: true
    hpCoDesign:
      enabled: true
      gpu: mi355x                # AMD GPU for AI model inference
      cpu: epyc-5th-gen
      gpuNodes: 2
      interconnect: roce
      targetLatencyUs: 5
      profile: qec-decoding
  result:
    backend: s3
    bucket: qec-results
    prefix: relaybp-ai/
```

**AI decoder architectures:**

| Decoder | Hardware | Latency | Accuracy |
|---------|----------|---------|----------|
| RelayBP (IBM) | AMD FPGA | ~5 μs | Near-optimal for surface codes |
| nv-qldpc (NVIDIA) | NVIDIA GPU | ~10 μs | Best for LDPC codes |
| BP+OSD (CUDA-QX) | NVIDIA GPU | ~20 μs | Good general-purpose |
| Neural net decoder | NVIDIA GPU | ~15 μs | Learned from data |
| Union-find (classical) | CPU | ~100 μs | Baseline |

---

### 5. Quantum-Enhanced Generative AI

**Description:** Use quantum circuits as generative models. The quantum state
encodes a probability distribution over the output space. Measuring the circuit
samples from this distribution. Training adjusts the circuit parameters to match
the target distribution.

**Why it works:** A quantum state with N qubits represents a distribution over
2^N outcomes with only 2^N - 1 complex amplitudes. This is an exponentially
compact representation of high-dimensional distributions.

**Applications:**
- Quantum Boltzmann machines (unsupervised learning)
- Quantum generative adversarial networks (QGANs)
- Quantum variational autoencoders
- Quantum-inspired sampling for diffusion models

**Operator configuration:**

```yaml
apiVersion: quantum.orchestrator.io/v1alpha1
kind: QuantumJob
metadata:
  name: quantum-gan
  labels:
    use-case: generative-ai
    model: qgan
spec:
  target:
    provider: nvidia        # GPU simulation for training
    device: nvidia
    simulator: true
  shots: 16384
  credential:
    secretRef:
      name: nvidia-gpu-creds
  nvidia:
    sdk: cudaq
    cudaqTarget: "nvidia"
    cudaqKernel:
      language: python
      inline: |
        import cudaq
        # Quantum GAN: generator + discriminator
        # Generator: parameterized quantum circuit
        # Discriminator: classical neural network (GPU)

        @cudaq.kernel
        def generator(z: list[float], phi: list[float]):
            q = cudaq.qvector(8)  # 8 qubits → 256-dim output
            # Encode classical latent variable z
            for i in range(4):
                rz(z[i] * 3.14, q[i])
            # Variational layer
            for layer in range(3):
                h(q)
                for i in range(7):
                    cx(q[i], q[i+1])
                for i in range(8):
                    rz(phi[layer*8+i], q[i])
            mz(q)

        # Classical discriminator: PyTorch on GPU
        # Training loop: alternating G and D updates
        ...
    gpu:
      model: h100
      count: 1
      minMemoryGiB: 80
  result:
    backend: minio
    bucket: genai-results
    prefix: qgan/
```

---

### 6. AI for Quantum Circuit Optimization

**Description:** Use reinforcement learning and evolutionary algorithms to
optimize quantum circuit transpilation, gate scheduling, and error-aware
compilation. The AI model learns hardware-specific heuristics that outperform
rule-based transpilers.

**Why it works:** Quantum circuit compilation is NP-hard (gate mapping, routing).
RL agents can learn effective heuristics without exhaustive search. The quantum
simulator provides the reward signal.

**Providers:**
- NVIDIA CUDA-Q + cuQuantum (fast simulation for RL training)
- IBM Qiskit v2.5 C API (fast transpilation for evaluation)
- Any provider for validation on real QPUs

**Operator configuration:**

```yaml
apiVersion: quantum.orchestrator.io/v1alpha1
kind: QuantumJob
metadata:
  name: circuit-opt-rl
  labels:
    use-case: circuit-optimization
    method: reinforcement-learning
spec:
  target:
    provider: nvidia        # GPU for RL training + simulation
    device: nvidia
    simulator: true
  shots: 10000
  credential:
    secretRef:
      name: nvidia-gpu-creds
  nvidia:
    sdk: cudaq
    cudaqTarget: "nvidia"
    gpu:
      model: h100
      count: 2             # 1 for RL, 1 for simulation
      minMemoryGiB: 160
  ibm:
    # Validate optimized circuits on real QPU
    qiskit:
      version: "2.5"
      compilationMode: pbc
      useCAPI: true
  result:
    backend: minio
    bucket: circuit-opt
    prefix: rl-agent/
```

---

### 7. Quantum-Accelerated Linear Algebra for AI

**Description:** Use quantum algorithms (HHL, quantum PCA, quantum SVM) to
accelerate linear algebra operations that are bottlenecks in AI: matrix
inversion, eigendecomposition, kernel computations.

**Why it works:** HHL solves Ax=b in O(log n) vs O(n³) classically. Quantum PCA
finds principal components in exponential feature spaces. These are most useful
for large, sparse, well-conditioned matrices.

**Note:** Requires fault-tolerant QC for practical speedup. In the NISQ era,
these are useful for small problems or as subroutines.

**Providers:**
- IBM Starling (fault-tolerant, 2029 target) — primary target
- IBM Nighthawk (NISQ, limited to small matrices)
- NVIDIA cuQuantum (GPU-accelerated classical simulation)

**Operator configuration:**

```yaml
apiVersion: quantum.orchestrator.io/v1alpha1
kind: QuantumJob
metadata:
  name: hhl-matrix-inversion
  labels:
    use-case: linear-algebra
    algorithm: hhl
    application: ai-training
spec:
  target:
    provider: ibm
    device: ibm_nighthawk_r2   # NISQ: small matrices only
  shots: 4096
  credential:
    secretRef:
      name: ibm-quantum-creds
  ibm:
    qiskit:
      version: "2.5"
      compilationMode: standard
      useCAPI: true
    errorMitigation:
      strategy: zero-noise
      hpcPowered: true
  result:
    backend: s3
    bucket: linear-algebra
    prefix: hhl/
```

---

## Use Case Summary

| # | Use Case | Quantum Role | AI Role | Primary Provider |
|---|----------|-------------|---------|-----------------|
| 1 | QML (classification, regression) | Feature map, variational layers | Optimizer, data pipeline | IBM / NVIDIA / AWS |
| 2 | QAOA for AI optimization | Solution space exploration | Parameter adjustment | NVIDIA / IBM |
| 3 | Quantum chemistry → drug discovery | Molecular Hamiltonian | Property prediction, screening | IBM / NVIDIA / Azure |
| 4 | AI-assisted QEC | Syndrome generation | Real-time decoding | IBM / NVIDIA |
| 5 | Quantum generative AI | Distribution representation | Training loop | NVIDIA |
| 6 | AI for circuit optimization | Simulation (reward signal) | RL / evolutionary search | NVIDIA / IBM |
| 7 | Quantum linear algebra | HHL, QPCA, QSVM | Model training | IBM Starling (FT) |

## Scheduling These Workloads

All use cases benefit from the operator's Kueue integration:

```yaml
# Kueue ResourceFlavors for AI + Quantum
apiVersion: kueue.x-k8s.io/v1beta1
kind: ResourceFlavor
metadata:
  name: gpu-a100
spec:
  resource:
    nvidia.com/gpu: "1"
---
apiVersion: kueue.x-k8s.io/v1beta1
kind: ResourceFlavor
metadata:
  name: qpu-nighthawk
spec:
  resource:
    quantum.orchestrator.io/qpu: "ibm_nighthawk_r2"
---
apiVersion: kueue.x-k8s.io/v1beta1
kind: ClusterQueue
metadata:
  name: ai-quantum-research
spec:
  resourceGroups:
    - flavors:
        - name: gpu-a100
          resources:
            - name: nvidia.com/gpu
              nominalQuota: 8
        - name: qpu-nighthawk
          resources:
            - name: quantum.orchestrator.io/qpu
              nominalQuota: 4
  fleet: ai-quantum
```

This ensures:
- GPU and QPU resources are co-scheduled
- Per-team quotas prevent one use case from starving others
- Preemption protects high-priority workloads (e.g., real-time QEC)

## References

- [IBM QDC 2025: Scaling for quantum advantage](https://www.ibm.com/quantum/blog/qdc-2025)
- [NVIDIA CUDA-Q + Braket: Accelerated quantum supercomputing](https://developer.nvidia.com/blog/accelerated-quantum-supercomputing-with-the-nvidia-cuda-q-and-amazon-braket-integration)
- [ORNL + IBM + AMD QEC co-design platform](https://newsroom.ibm.com/2026-11-05-oark-ridge-national-lab,-ibm,-and-amd-launch-new-co-design-platform-for-quantum-error-correction)
- [NVIDIA NVQLink architecture](https://developer.nvidia.com/blog/nvidia-nvqlink-architecture-integrates-accelerated-computing-with-quantum-processors)
- [Qiskit v2.5: PBC, C API, Clifford+T](https://www.ibm.com/quantum/blog/qiskit-2-5-release-summary)
- [cuQuantum SDK v25.11](https://developer.nvidia.com/blog/advanced-large-scale-quantum-simulation-techniques-in-cuquantum-sdk-v25-11)
- [NVIDIA GB200 NVL72: Journey to quantum computing](https://blogs.nvidia.com/blog/journey-to-quantum-computing)

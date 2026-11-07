# IBM Quantum Provider Guide

## IBM Quantum Stack (2025–2026)

### Hardware

| Processor | Qubits | Topology | Highlights | Status |
|-----------|--------|----------|------------|--------|
| **Nighthawk r2** | 120 + 458 physical | Square lattice | High-speed reset, 30% more circuit depth, 218+ couplers | Live (Sep 2026) |
| **Nighthawk r1** | 120 | Square lattice | First square-lattice chip, 218 tunable couplers | Live (Dec 2025) |
| **Heron r3** | 156 | Heavy-hexagonal | Lowest median 2-qubit gate errors, **330,000 CLOPS** | Live (Jul 2025) |
| **Starling** | TBD (logical) | — | First fault-tolerant QC, 100+ logical qubits | 2029 target |

### Software

| Component | Version | Key Features |
|-----------|---------|-------------|
| **Qiskit** | v2.5 (Jul 2026) | C API, Pauli-based computation (PBC), Clifford+T compilation, expanded transpiler |
| **IBM Quantum Compute Service** | — | (formerly Qiskit Runtime) managed cloud service |
| **RelayBP** | — | Real-time QEC decoder, AMD FPGA, flexible + accurate + fast + compact |
| **HPC Error Mitigation** | — | 100x cost reduction, GPU/CPU accelerated |
| **Dynamic Circuits** | — | 24% accuracy increase at 100+ qubits |

### QEC / HPC Integration

- **ORNL + IBM + AMD co-design platform** (Nov 2026): Nighthawk r2 + AMD MI355X GPUs + 5th Gen EPYC
- **70 logical qubits, 438 hard operations** demonstrated
- **Target**: 100 logical qubits, 100K hard ops by 2028 (DOE Quantum Genesis Mission)
- Algorithm-aware decoders, hardware-aware compilation, software orchestration

## Authentication

### IBM Cloud API Key

```bash
kubectl create secret generic ibm-quantum-creds \
  --from-literal=token=YOUR_IBM_CLOUD_API_KEY \
  -n quantum-system
```

### IBM Quantum Compute Service

```yaml
credential:
  secretRef:
    name: ibm-quantum-creds
    key: token
```

## Supported Processors

| CRD `target.device` | Processor | Qubits | Notes |
|---------------------|-----------|--------|-------|
| `ibm_nighthawk_r2` | Nighthawk r2 | 120 + 458 | High-speed reset, best for deep circuits |
| `ibm_nighthawk_r1` | Nighthawk r1 | 120 | Square lattice, 218 couplers |
| `ibm_heron_r3` | Heron r3 | 156 | Best for wide circuits, 330K CLOPS |
| `ibm_starling` | Starling | TBD | Fault-tolerant (2029) |

## Qiskit v2.5 Features

### Pauli-based Computation (PBC)

```yaml
ibm:
  qiskit:
    version: "2.5"
    compilationMode: pbc
    useCAPI: true
    presetPassManager: pbc
```

PBC compilation pipeline:
1. Decompose circuit into Pauli operations
2. Compile using PBC-specific passes
3. Map to hardware with hardware-aware optimization

### Clifford+T Compilation

```yaml
ibm:
  qiskit:
    compilationMode: clifford-t
    presetPassManager: clifford-t
```

### C API Transpilation

```yaml
ibm:
  qiskit:
    useCAPI: true  # Rust-based transpiler (Qiskit v2.5+)
```

## Error Mitigation

| Strategy | Description | HPC-Powered |
|----------|-------------|-------------|
| `zero-noise` | Zero-noise extrapolation | ✅ 100x cost reduction |
| `probabilistic-error-mitigation` | PEM with noise model learning | ✅ |
| `readout-mitigation` | Readout error correction | — |
| `dynamic-circuits` | Mid-circuit measurement + reset | — |

```yaml
ibm:
  errorMitigation:
    strategy: zero-noise
    hpcPowered: true
    readoutMitigation: true
    extrapolationOrders: [1, 2, 3, 4]
```

## Quantum Error Correction (RelayBP)

RelayBP is IBM's real-time QEC decoder, implemented on AMD FPGA:
- Flexible: supports multiple code families
- Accurate: near-optimal decoding
- Fast: microsecond latency
- Compact: FPGA footprint

```yaml
ibm:
  qec:
    decoder: relaybp
    decoderHardware: fpga
    code: surface
    distance: 5
    logicalQubits: 70
    hardOperations: 438
    realTime: true
    algorithmAware: true
    hardwareAwareCompilation: true
```

### HPC Co-Design Platform

The ORNL + IBM + AMD pattern: QPU + GPU + CPU co-located for low-latency QEC.

```yaml
ibm:
  hpCoDesign:
    enabled: true
    gpu: mi355x          # AMD Instinct MI355X
    cpu: epyc-5th-gen   # 5th Gen AMD EPYC
    gpuNodes: 2
    interconnect: roce
    targetLatencyUs: 5
    profile: qec-decoding
```

## Container Images

| Image | Contents |
|-------|----------|
| `ibmquantum/qiskit:latest` | Qiskit v2.5 + IBM Quantum Compute Service |
| `ghcr.io/quantum-orchestrator/notebook-ibm:latest` | Qiskit + PennyLane + Cirq + Qibo + stim + pymatching + Jupyter |

## Roadmap

| Milestone | Target |
|-----------|--------|
| Quantum advantage (useful computation) | End of 2026 |
| 7,500 gates per circuit | End of 2026 |
| 10,000 gates per circuit | 2027 |
| 100 logical qubits, 100K hard ops | 2028 (DOE Q Competition) |
| IBM Quantum Starling (fault-tolerant) | 2029 |

## References

- [IBM Quantum Nighthawk Announcement](https://newsroom.ibm.com/2025-11-12-ibm-delivers-new-quantum-processors,-software,-and-algorithm-breakthroughs-on-path-to-advantage-and-fault-tolerance)
- [Qiskit v2.5 Release](https://www.ibm.com/quantum/blog/qiskit-2-5-release-summary)
- [ORNL + IBM + AMD QEC Platform](https://newsroom.ibm.com/2026-11-05-oak-ridge-national-lab,-ibm,-and-amd-launch-new-co-design-platform-for-quantum-error-correction)
- [IBM QEC Continuum](https://www.ibm.com/quantum/blog/qec-continuum)
- [IBM Quantum System Two (Chicago)](https://newsroom.ibm.com/2026-11-06-ibm-unveils-its-most-advanced-quantum-computer-in-chicago)
- [IBM Quantum Processor Types](https://quantum.cloud.ibm.com/docs/en/guides/processor-types)
- [IBM QDC 2025](https://www.ibm.com/quantum/blog/qdc-2025)
- [Qiskit GitHub](https://github.com/Qiskit/qiskit)
- [IBM Quantum Documentation](https://quantum.cloud.ibm.com/docs)

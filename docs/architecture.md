# Architecture

## Overview

Quantum Orchestrator is a Kubernetes Operator that provides a unified API for
submitting, scheduling, and tracking quantum computing workloads across
multiple cloud providers and direct QPU vendors.

## Components

### Operator (Go)

The core controller that:
1. Watches `QuantumJob`, `QuantumSession`, `QPUBackend`, `QuantumProviderBinding`, and `QuantumSimulator` CRDs
2. Resolves the target provider from the job spec
3. Submits the job via the appropriate `Provider` adapter
4. Polls status and updates the CRD `status` subresource
5. Handles finalizers for clean cancellation

### Provider Adapters (Go)

| Adapter | Provider | Auth |
|---------|----------|------|
| `BraketAdapter` | AWS Braket | IRSA / IAM |
| `AzureQuantumAdapter` | Azure Quantum | Workload Identity |
| `DirectAdapter` | IonQ, Rigetti, Quantinuum, IQM | API key |
| `QbraidAdapter` | Any (via qBraid) | API key |

All adapters implement the `adapter.Provider` interface.

### Transpiler Sidecar (Python)

A FastAPI service that converts circuits between formats:
- OpenQASM 3 ↔ QIR ↔ Quil ↔ Q# ↔ Qiskit ↔ Cirq ↔ PennyLane
- Uses qBraid SDK and qBraid-QIR under the hood
- Called by the operator via HTTP when a format conversion is needed

### CLI (quantumctl)

A cobra-based CLI that talks to the Kubernetes API:
- `quantumctl submit -f job.yaml`
- `quantumctl get jobs`
- `quantumctl logs <job>`
- `quantumctl results <job> --plot`
- `quantumctl devices`
- `quantumctl cost`

## Data Flow

```
User → quantumctl / kubectl → QuantumJob CRD
  → Operator Reconcile Loop
    → Provider Adapter (Braket / Azure / Direct)
      → Cloud API (CreateTask / SubmitJob)
    → Poll Status (GetTask / GetJob)
    → Update CRD Status
    → Fetch Results → S3 / Blob / MinIO
```

## Multi-Tenancy

- `QuantumProviderBinding` maps a namespace/tenant to a cloud account
- The operator checks the binding before submitting
- Credentials are injected via K8s Secrets (ESO / Vault)
- Kueue enforces per-tenant QPU quotas

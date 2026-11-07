# Azure Quantum Provider Guide

## Authentication

### Option 1: Workload Identity (Recommended on AKS)

1. Create an Entra ID app registration
2. Assign `Quantum Job Submission` role to the app
3. Configure Azure Workload Identity on AKS
4. The operator pod automatically gets a token via Projected Token

### Option 2: Service Principal

```bash
kubectl create secret generic azure-quantum-creds \
  --from-literal=workspace=my-workspace \
  --from-literal=subscription=SUB-ID \
  --from-literal=tenantId=TENANT-ID \
  --from-literal=clientId=CLIENT-ID \
  --from-literal=clientSecret=SECRET \
  -n quantum-system
```

## Supported Providers (via Azure Quantum)

| Provider | Devices | Notes |
|----------|---------|-------|
| IonQ | `ionq.simulator`, `ionq.qpu.aria-1-25q` | Trapped ion |
| Quantinuum | `quantinuum.sim.h2-1e`, `quantinuum.h2-1` | High-fidelity |
| Rigetti | `rigetti.sim.qvm`, `rigetti.qpu.cepheus-1-108q` | Superconducting |

## Integrated Hybrid (QIR Adaptive RI)

For adaptive circuits on Quantinuum:

1. Write the circuit in Q# or Qiskit
2. The transpiler sidecar compiles to QIR `.bc`
3. The adapter submits as a QIR Adaptive RI job
4. Azure Quantum handles the classical-quantum loop

```yaml
target:
  provider: azure-quantum
  device: quantinuum.h2-1
  simulator: false
```

## Sessions

```yaml
# Create a session
apiVersion: quantum.orchestrator.io/v1alpha1
kind: QuantumSession
metadata:
  name: my-session
spec:
  target:
    provider: azure-quantum
    device: ionq.qpu.aria-1-25q
  duration: "1h"
  qubits: 25
  credential:
    secretRef:
      name: azure-quantum-creds

# Submit jobs within the session
# (QuantumJobs reference the session via label)
```

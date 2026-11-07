# CRD Reference

## QuantumJob

The primary workload object.

### Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `circuit.inline` | string | no | OpenQASM 3 source |
| `circuit.ref` | ObjectRef | no | ConfigMap/Secret reference |
| `circuit.format` | string | no | Default: `openqasm3` |
| `target.provider` | string | **yes** | `aws-braket`, `azure-quantum`, `ionq`, `rigetti`, `quantinuum`, `local` |
| `target.device` | string | **yes** | Device ID (e.g. `sv1`, `ionq.simulator`) |
| `target.simulator` | bool | no | Use local simulator |
| `shots` | int32 | no | Default: 1024 |
| `hybrid.image` | string | no | Classical component image |
| `hybrid.entrypoint` | []string | no | Command args |
| `hybrid.resources` | ResourceReq | no | CPU/GPU for classical |
| `credential.secretRef` | SecretRef | **yes** | Provider credentials |
| `scheduling.queue` | string | no | Kueue queue |
| `scheduling.priorityClass` | string | no | Priority |
| `scheduling.retries` | int32 | no | Default: 3 |
| `result.backend` | string | no | `s3`, `azure-blob`, `minio` |
| `result.bucket` | string | no | Storage bucket |
| `parameters` | map | no | Circuit parameters |

### Status

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | `Pending` → `Submitted` → `Queued` → `Running` → `Completed`/`Failed` |
| `providerJobId` | string | Provider's job ID |
| `resultLocation` | string | URI to results |
| `startTime` | Time | |
| `completionTime` | Time | |
| `costEstimate` | float64 | USD |

## QuantumSession

Long-lived QPU session (Azure Sessions / Braket reserved time).

## QPUBackend

Device inventory with qubits, gate set, error rates, cost.

## QuantumProviderBinding

Tenant → cloud account / workspace mapping.

## QuantumSimulator

Local/managed simulator as a Deployment.

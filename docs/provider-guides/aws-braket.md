# AWS Braket Provider Guide

## Authentication

### Option 1: IRSA (Recommended on EKS)

1. Create an IAM role with `AmazonBraketFullAccess`
2. Trust policy: `sts.amazonaws.com` + `pods` (IRSA)
3. Annotate the ServiceAccount:

```yaml
annotations:
  eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/quantum-orchestrator
```

### Option 2: Static Credentials

```bash
kubectl create secret generic braket-creds \
  --from-file=credentials=creds.json \
  -n quantum-system
```

## Supported Devices

| Device | Type | Qubits | Notes |
|--------|------|--------|-------|
| `sv1` | Simulator | 34 | Statevector, free |
| `dm1` | Simulator | 14 | Density matrix |
| `tn1` | Simulator | 34 | Tensor network |
| `ionq.harmony` | QPU | 11 | Trapped ion |
| `rigetti.ankaa` | QPU | 84 | Superconducting |
| `iqm.garnet` | QPU | 10 | Superconducting |

## Hybrid Jobs

```yaml
hybrid:
  image: ghcr.io/my-org/vqe-optimizer:latest
  entrypoint: ["python", "-m", "optimizer"]
  resources:
    limits:
      cpu: "2"
      memory: 4Gi
```

The operator packages this as a Braket Hybrid Job:
1. Pushes the image to ECR
2. Calls `CreateHybridJob` with the image URI + QPU target
3. Watches via `GetHybridJob`

## Results

Results land in S3. The operator reads the S3 URI from the task metadata
and writes standardised JSON to the configured result bucket.

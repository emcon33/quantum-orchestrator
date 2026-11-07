# Getting Started

## Prerequisites

- Kubernetes 1.28+ or OpenShift 4.14+
- `kubectl` or `oc`
- A quantum provider account

## Install

```bash
# Option 1: Helm
helm repo add quantum-orchestrator https://charts.quantum-orchestrator.io
helm install quantum-orchestrator quantum-orchestrator/quantum-orchestrator \
  -n quantum-system --create-namespace

# Option 2: Kustomize
kubectl apply -k config/default
```

## Create Credentials

### AWS Braket

```bash
# Using IRSA (recommended)
# Add the annotation to the ServiceAccount:
#   eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT:role/quantum-orchestrator

# Or create a secret:
kubectl create secret generic braket-creds \
  --from-literal=credentials=base64-encoded-iam-creds \
  -n quantum-system
```

### Azure Quantum

```bash
# Using Workload Identity (recommended on AKS)
# Or create a secret:
kubectl create secret generic azure-quantum-creds \
  --from-literal=workspace=my-workspace \
  --from-literal=subscription=SUB-ID \
  -n quantum-system
```

## Submit a Job

```bash
kubectl apply -f config/samples/quantumjob_bell_state.yaml
kubectl get quantumjobs -n quantum-system
```

## Verify

```bash
kubectl get quantumjobs -w
kubectl describe quantumjob bell-state
kubectl get quantumjobs bell-state -o jsonpath='{.status.resultLocation}'
```

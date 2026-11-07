#!/usr/bin/env bash
# Regenerate deepcopy functions and CRD manifests.
set -euo pipefail

# Install controller-gen if not present
if ! command -v controller-gen &>/dev/null; then
  go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
fi

# Generate deepcopy
controller-gen object header:"" paths="./api/..."

# Generate CRDs
controller-gen crd paths="./api/..." output:crd:artifacts:config=config/crd/bases

echo "Code generation complete."

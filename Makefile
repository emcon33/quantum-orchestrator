# Quantum Orchestrator Makefile

IMG ?= ghcr.io/quantum-orchestrator/operator:$(shell git describe --tags --always 2>/dev/null || echo "dev")
TRANSPIler_IMG ?= ghcr.io/quantum-orchestrator/transpiler:$(shell git describe --tags --always 2>/dev/null || echo "dev")

# Get the currently git commit hash
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X main.commit=$(COMMIT) -X main.date=$(shell date -u +%Y-%m-%d)

# Build tags
BUILD_TAGS ?=

.PHONY: all
all: build

## ── Build ──────────────────────────────────────────────

.PHONY: build
build: ## Build the operator binary
	CGO_ENABLED=0 GOOS=linux go build -ldflags="$(LDFLAGS)" -tags "$(BUILD_TAGS)" -o bin/operator ./cmd/operator

.PHONY: quantumctl
quantumctl: ## Build the quantumctl CLI
	CGO_ENABLED=0 GOOS=linux go build -ldflags="$(LDFLAGS)" -tags "$(BUILD_TAGS)" -o bin/quantumctl ./cmd/quantumctl

.PHONY: docker-build
docker-build: ## Build the operator Docker image
	docker build -t $(IMG) .

.PHONY: docker-build-transpiler
docker-build-transpiler: ## Build the transpiler Docker image
	docker build -t $(TRANSPIler_IMG) -f transpiler/Dockerfile transpiler/

.PHONY: docker-push
docker-push: ## Push Docker images
	docker push $(IMG)
	docker push $(TRANSPIler_IMG)

## ── Test ───────────────────────────────────────────────

.PHONY: test
test: ## Run unit tests
	go test ./... -v -coverprofile=coverage.out -covermode=atomic

.PHONY: test-e2e
test-e2e: ## Run E2E tests (requires a cluster)
	go test ./test/e2e/... -v -count=1 -timeout 30m

## ── Lint ───────────────────────────────────────────────

.PHONY: lint
lint: ## Run linters
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	go fmt ./...
	goimports -local github.com/quantum-orchestrator -w .

## ── Code Generation ────────────────────────────────────

.PHONY: generate
generate: ## Generate deepcopy functions
	controller-gen object header:"" paths="./api/..."

.PHONY: manifests
manifests: ## Generate CRD manifests
	controller-gen crd paths="./api/..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate-all
generate-all: generate manifests ## Generate all code

## ── Deploy ─────────────────────────────────────────────

.PHONY: deploy
deploy: ## Deploy to the cluster (kustomize)
	kubectl apply -k config/default

.PHONY: undeploy
undeploy: ## Undeploy from the cluster
	kubectl delete -k config/default

.PHONY: helm-install
helm-install: ## Install via Helm
	helm install quantum-orchestrator charts/quantum-orchestrator \
		--namespace quantum-system --create-namespace

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall via Helm
	helm uninstall quantum-orchestrator --namespace quantum-system

## ── Clean ──────────────────────────────────────────────

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/ coverage.out

## ── Help ───────────────────────────────────────────────

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

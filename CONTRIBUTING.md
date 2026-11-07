# Contributing

Thanks for your interest in contributing!

## Getting Started

1. Fork the repository
2. Install Go 1.22+, Python 3.10+
3. Run `make build test`
4. Make your changes
5. Submit a PR with a clear description

## Development

```bash
# Clone
git clone https://github.com/quantum-orchestrator/quantum-orchestrator.git
cd quantum-orchestrator

# Build
make build

# Test
make test

# Lint
make lint

# Deploy to dev cluster
make deploy
```

## Pull Request Guidelines

- One logical change per PR
- Include tests for new functionality
- Update docs if the API or behavior changes
- Follow [Conventional Commits](https://www.conventionalcommits.org/)
- Link related issues

## Code of Conduct

We follow the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/policies-guidance/cncf-code-of-conduct.md).
Be kind, be constructive, be inclusive.

## Security

Report vulnerabilities via [SECURITY.md](SECURITY.md). Do NOT open public issues.

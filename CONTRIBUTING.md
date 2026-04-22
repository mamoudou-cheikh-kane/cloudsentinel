# Contributing to CloudSentinel

Thank you for your interest in CloudSentinel! This document describes how to contribute to the project.

## 🎯 Philosophy

CloudSentinel aims to be a reference implementation of distributed system testing. We value:

- **Clarity over cleverness** — code should be easy to understand
- **Tests as documentation** — behavior is documented by tests, not prose
- **Fail-safe by default** — chaos experiments always have rollbacks
- **Multi-cloud portability** — avoid provider-specific logic outside of providers

## 🛠️ Development Setup

### Prerequisites

- Python 3.11+, Go 1.21+, Node 20+
- Docker with at least 4GB RAM
- kubectl, Helm, Terraform, kind
- pre-commit

### Clone and setup

```bash
git clone https://github.com/mamoudou-cheikh-kane/cloudsentinel.git
cd cloudsentinel
pre-commit install
```

### Running tests locally

```bash
# Python services (example for orchestrator)
cd orchestrator && poetry install && poetry run pytest

# Go agent
cd agent && go test ./...

# Terraform validation
cd infrastructure && terraform fmt -check -recursive
```

## 📝 Commit Conventions

We use [Conventional Commits](https://www.conventionalcommits.org/). Examples:
feat(orchestrator): add GKE provider support
fix(agent): handle nil pointer on metric collection
docs(readme): update quickstart instructions
chore(ci): bump golangci-lint to v1.61
test(chaos): add scenario for DNS failures
refactor(api): extract database layer
Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`, `build`.

## 🔀 Pull Request Process

1. Create a branch from `main`: `git checkout -b feat/my-feature`
2. Make changes with meaningful commits
3. Ensure all tests pass and pre-commit hooks succeed
4. Push and open a PR with a clear description
5. Request review
6. Merge after approval (squash merge preferred)

## 🐛 Reporting Bugs

Open an issue with:
- Clear title describing the problem
- Steps to reproduce
- Expected vs actual behavior
- Environment (OS, K8s version, cloud provider)
- Relevant logs

## 💡 Proposing Features

Open a discussion first before implementing significant features. Small improvements can go directly to PR.

## 🔒 Reporting Security Issues

Do **not** open public issues for security vulnerabilities. Contact the maintainer directly.

## 📄 License

By contributing, you agree that your contributions will be licensed under Apache License 2.0.

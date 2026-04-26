# Contributing to CloudSentinel

First off, thank you for considering contributing to CloudSentinel! 🎉
Every contribution — from a typo fix to a new cloud provider — helps make
this project better for everyone.

This document is the **single source of truth** for contributing. If
something here is unclear, that is a bug; please open an issue or a PR
to fix it.

---

## Table of contents

1. [Code of conduct](#code-of-conduct)
2. [Project philosophy](#project-philosophy)
3. [Where to get help](#where-to-get-help)
4. [Reporting bugs](#reporting-bugs)
5. [Suggesting features](#suggesting-features)
6. [Good first issues](#good-first-issues)
7. [Development setup](#development-setup)
8. [Per-service guide](#per-service-guide)
9. [Coding conventions](#coding-conventions)
10. [Testing strategy](#testing-strategy)
11. [Pull request process](#pull-request-process)
12. [Documentation](#documentation)
13. [Security issues](#security-issues)
14. [License](#license)

---

## Code of conduct

This project adheres to a [Code of Conduct](./CODE_OF_CONDUCT.md). By
participating, you are expected to uphold it. Please report unacceptable
behavior to the maintainer.

## Project philosophy

CloudSentinel aims to be a reference implementation of distributed
systems testing. The core values are:

- **Clarity over cleverness.** Code should be easy to read for someone
  joining the project today.
- **Tests as documentation.** Behavior is documented by tests, not by
  long prose. If a function is not tested, its behavior is not defined.
- **Fail-safe by default.** Every chaos experiment must have an
  automatic rollback. The system should never leave a cluster in a
  broken state on purpose.
- **Typed contracts everywhere.** Pydantic on the API, Protobuf on the
  agent, TypeScript on the dashboard. If a contract changes, the type
  checker should complain before the user does.
- **One image per service.** Every service ships with a multi-stage
  Dockerfile, runs as a non-root user, and exposes a healthcheck.

## Where to get help

| I want to... | Use |
|----|----|
| Ask a question about how to use the project | [GitHub Discussions](https://github.com/mamoudou-cheikh-kane/cloudsentinel/discussions) |
| Report a bug | [GitHub Issues](https://github.com/mamoudou-cheikh-kane/cloudsentinel/issues) → "Bug report" template |
| Propose a feature | [GitHub Issues](https://github.com/mamoudou-cheikh-kane/cloudsentinel/issues) → "Feature request" template |
| Find a starter task | Issues labeled [`good first issue`](https://github.com/mamoudou-cheikh-kane/cloudsentinel/labels/good%20first%20issue) |

## Reporting bugs

A good bug report contains:

- A clear, descriptive title (e.g. `agent panics on nodes without /proc/cpuinfo`)
- The exact steps to reproduce
- What you expected to happen vs what actually happened
- Your environment: OS, Kubernetes version, Go/Python/Node version
- Any relevant logs (use code blocks, not screenshots)

Please search existing issues first — your bug may already be tracked.

## Suggesting features

Before opening a PR for a non-trivial feature, please open an issue
first to discuss the approach. This avoids the situation where a great
PR has to be rejected because the design does not fit the project
direction. For typo fixes, dependency bumps, and small improvements,
go straight to a PR.

## Good first issues

If you are new to the project, look for issues labeled
[`good first issue`](https://github.com/mamoudou-cheikh-kane/cloudsentinel/labels/good%20first%20issue).
Examples of good first contributions:

- Add a new fault type to the agent (e.g. `disk_fill`, `network_loss`)
- Implement a stub Terraform module for AKS or GKE
- Add a new MkDocs page (Tutorial, FAQ, etc.)
- Improve test coverage of a specific module
- Add an alternative ML model and compare its F1 score

---

## Development setup

### Prerequisites

- **Docker 24+** with at least 4 GB RAM allocated
- **kubectl 1.28+**
- **kind 0.20+** (for local Kubernetes clusters)
- **Terraform 1.5+**
- **Go 1.24+**
- **Python 3.11+** with Poetry 2.x
- **Node 20+** with npm
- **pre-commit** (`pip install --user pre-commit`)

### Clone and bootstrap

```bash
git clone https://github.com/mamoudou-cheikh-kane/cloudsentinel.git
cd cloudsentinel
pre-commit install
```

The pre-commit hook chain runs `ruff`, `ruff-format`, `gofmt`, `go vet`,
`terraform fmt`, secret detection, and end-of-file fixers on every
commit. If a hook fails, fix the issue and commit again.

### Verify your setup

```bash
pre-commit run --all-files
```

You should see every hook passing. If something fails, the error
message tells you exactly which file to fix.

---

## Per-service guide

CloudSentinel is a multi-service repo. Each service has its own
language, its own dependencies, and its own test suite.

### `orchestrator/` — Python CLI (Poetry)

```bash
cd orchestrator
poetry install
poetry run pytest                          # 24 tests
poetry run cloudsentinel --help            # explore the CLI
```

Tech: Python 3.11, Typer, Rich, Pydantic v2, gRPC client.

### `agent/` — Go DaemonSet

```bash
cd agent
go mod download
go test ./...                              # 6 tests
go build -o bin/agent ./cmd/agent          # produce the binary
docker build -t cloudsentinel-agent:dev .  # multi-stage distroless
```

Tech: Go 1.24, gRPC, Prometheus client, gopsutil.

### `api/` — FastAPI backend

```bash
cd api
poetry install
poetry run pytest                          # 11 tests
poetry run uvicorn app.main:app --reload   # run locally on :8000
```

OpenAPI docs at `http://localhost:8000/docs`.

Tech: FastAPI, SQLModel, SQLite, Pydantic v2.

### `dashboard/` — Next.js frontend

```bash
cd dashboard
npm install
npm run lint                               # ESLint
npm run dev                                # http://localhost:3000
```

Make sure the API is running on port 8000 first, then create
`dashboard/.env.local` with:
NEXT_PUBLIC_API_BASE_URL=http://localhost:8000

Tech: Next.js 16, TypeScript, Tailwind CSS v4, shadcn/ui.

### `ml-pipeline/` — Anomaly detection

```bash
cd ml-pipeline
poetry install
poetry run pytest                          # 13 tests
poetry run python -m cloudsentinel_ml.cli train
```

Tech: scikit-learn, pandas, numpy, joblib, Typer.

### `infrastructure/` — Terraform

```bash
cd infrastructure
terraform fmt -check -recursive
terraform -chdir=modules/kind init && terraform -chdir=modules/kind validate
```

Tech: Terraform 1.5, Kind provider (modules for AKS/EKS/GKE are
placeholder stubs at the moment).

### `docs/` — MkDocs site

```bash
pip install --user mkdocs mkdocs-material
mkdocs serve --dev-addr 127.0.0.1:8001
```

The docs auto-deploy to GitHub Pages on every push to `main` that
touches `docs/**` or `mkdocs.yml`.

---

## Coding conventions

### Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org/).
Examples:
feat(orchestrator): add GKE provider support
fix(agent): handle nil pointer on metric collection
docs(readme): update quickstart instructions
chore(ci): bump golangci-lint to v1.61
test(chaos): add scenario for DNS failures
refactor(api): extract database layer

Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`,
`test`, `chore`, `ci`, `build`.

The pre-commit chain enforces this format — if your commit message
does not match, the commit is rejected.

### Python

- Format: `ruff-format` (PEP 8 compliant, 88 char line length)
- Lint: `ruff` with the project rules in `pyproject.toml`
- Type hints required on all public functions
- Pydantic v2 for data models
- One module per concern; no god-files

### Go

- Format: `gofmt -s` (with simplification)
- Lint: `go vet`
- Errors are values; use `fmt.Errorf("context: %w", err)` to wrap
- Public functions get a doc comment starting with the function name
- Table-driven tests for all logic with branches

### TypeScript

- Lint: `eslint`
- React functional components only (no class components)
- Tailwind utility classes preferred over custom CSS
- shadcn/ui components for primitives (Button, Dialog, Input, etc.)

### Terraform

- Format: `terraform fmt`
- Validate: `terraform validate`
- Variables typed and documented
- Outputs documented

---

## Testing strategy

CloudSentinel follows the testing pyramid: lots of fast unit tests, a
smaller integration layer, and a thin top of end-to-end tests.

| Suite | Type | Count | Runtime |
|-------|------|-------|---------|
| `orchestrator/tests/` | Unit | 24 | ~1 s |
| `agent/internal/...` | Unit (Go) | 6 | ~1 s |
| `api/tests/` | Integration (TestClient + SQLite) | 11 | <1 s |
| `ml-pipeline/tests/` | Unit | 13 | ~2 s |

**Rules**:
- Every PR must add tests for new behavior
- Existing tests must keep passing on main
- Tests must run in **under 10 seconds total** locally
- Use `pytest` markers (`@pytest.mark.slow`) for any test slower than 1 s

---

## Pull request process

1. **Fork** the repo and create a branch from `main`:
   `git checkout -b feat/awesome-feature`
2. **Make changes** with meaningful commits (Conventional Commits format)
3. **Run tests + pre-commit** locally — every check must pass:
```bash
   pre-commit run --all-files
   cd orchestrator && poetry run pytest
   cd ../agent && go test ./...
   cd ../api && poetry run pytest
   cd ../ml-pipeline && poetry run pytest
```
4. **Push** to your fork and open a PR against `main`
5. **PR description** should explain *what* changed and *why*. Link any
   related issue with `Closes #123`.
6. **CI must be green** — all four GitHub Actions pipelines pass
7. **Wait for review** — I aim to respond within 48 hours
8. **Address feedback** by pushing more commits to the same branch
9. **Squash merge** is preferred to keep main history linear

---

## Documentation

- **README.md** — high-level overview for first-time visitors
- **`docs/`** — MkDocs site with Getting Started, Architecture,
  ML Pipeline. Live at
  [`mamoudou-cheikh-kane.github.io/cloudsentinel`](https://mamoudou-cheikh-kane.github.io/cloudsentinel/)
- **Per-service `README.md`** — quick reference for the service you
  are working on
- **`docs/adr/`** — Architecture Decision Records for non-obvious
  choices

When your PR touches behavior, update the docs in the same PR.

---

## Security issues

**Do not open public issues for security vulnerabilities.**

If you discover a security issue, please contact the maintainer
directly at [mamoudoucheikhk@gmail.com](mailto:mamoudoucheikhk@gmail.com)
with:

- A description of the vulnerability
- Steps to reproduce
- The impact

We will acknowledge receipt within 72 hours and aim to publish a fix
within 14 days for critical issues.

---

## License

CloudSentinel is licensed under [Apache License 2.0](./LICENSE).

By contributing, you agree that your contributions will be licensed
under the same Apache 2.0 license, and you certify that you have the
right to submit them under that license.

Thank you again for your contribution! 🛡️

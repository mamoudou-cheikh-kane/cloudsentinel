# 🛡️ CloudSentinel

[![Python CI](https://github.com/mamoudou-cheikh-kane/cloudsentinel/actions/workflows/python-ci.yml/badge.svg)](https://github.com/mamoudou-cheikh-kane/cloudsentinel/actions/workflows/python-ci.yml)
[![Go CI](https://github.com/mamoudou-cheikh-kane/cloudsentinel/actions/workflows/go-ci.yml/badge.svg)](https://github.com/mamoudou-cheikh-kane/cloudsentinel/actions/workflows/go-ci.yml)
[![Terraform CI](https://github.com/mamoudou-cheikh-kane/cloudsentinel/actions/workflows/terraform-ci.yml/badge.svg)](https://github.com/mamoudou-cheikh-kane/cloudsentinel/actions/workflows/terraform-ci.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

> **Multi-cloud chaos engineering platform for Kubernetes.**
> Provision clusters, inject faults, collect metrics, detect anomalies — all in one typed, tested toolkit.

[![Python](https://img.shields.io/badge/python-3.11-blue)](https://www.python.org)
[![Go](https://img.shields.io/badge/go-1.24-00ADD8)](https://go.dev)
[![TypeScript](https://img.shields.io/badge/typescript-5-3178C6)](https://www.typescriptlang.org)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.30-326CE5)](https://kubernetes.io)
[![FastAPI](https://img.shields.io/badge/FastAPI-0.136-009688)](https://fastapi.tiangolo.com)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000)](https://nextjs.org)

---

## 🎯 What is CloudSentinel?

CloudSentinel is an end-to-end platform that lets you:

- 🏗️ **Provision** Kubernetes clusters (Kind, EKS, AKS, GKE) through a single Python CLI
- 💥 **Inject** controlled faults (CPU stress, memory pressure, network latency, disk fill) into any node
- 📊 **Collect** node-level metrics via a lightweight Go agent running as a DaemonSet
- 🎛️ **Drive** scenarios from a REST API, a web dashboard, or the command line
- 🤖 **Detect** anomalies in collected metrics with an Isolation Forest model

## 🏗️ Architecture

```text
       ┌──────────────────────────────┐
       │   USER (browser / CLI)       │
       └──────────────┬───────────────┘
                      │
         ┌────────────┴────────────┐
         ▼                         ▼
  ┌────────────────┐     ┌────────────────┐
  │   DASHBOARD    │     │  ORCHESTRATOR  │
  │ (Next.js+SHCN) │     │   (Python CLI) │
  └────────┬───────┘     └────────┬───────┘
           │ REST                 │ gRPC
           ▼                      ▼
  ┌────────────────┐     ┌────────────────┐
  │      API       │     │   AGENT (Go)   │
  │   (FastAPI)    │     │  DaemonSet/K8s │
  └────────────────┘     └────────┬───────┘
                                  │ metrics
                                  ▼
                         ┌────────────────┐
                         │   ML PIPELINE  │
                         │ (scikit-learn) │
                         └────────────────┘
```

Full details in [`docs/architecture.md`](./docs/architecture.md).

## 📦 Components

| Component | Language | Role |
|-----------|----------|------|
| `orchestrator/` | Python + Poetry | CLI that provisions clusters (Terraform) and runs chaos scenarios over gRPC |
| `agent/` | Go 1.24 | DaemonSet pod exposing Prometheus metrics + gRPC fault injection |
| `api/` | Python (FastAPI + SQLModel) | REST API persisting scenarios in SQLite, CORS-ready for the dashboard |
| `dashboard/` | TypeScript (Next.js + Tailwind + shadcn/ui) | Web UI to list, create, and run scenarios with 5s polling |
| `ml-pipeline/` | Python (scikit-learn) | Isolation Forest anomaly detection on node metrics |
| `infrastructure/` | Terraform | Modules for Kind (done), AKS/EKS/GKE (placeholders) |
| `docs/` | Markdown (MkDocs Material) | Architecture, Getting Started, ML Pipeline guides |

## 🚀 Quick Start

**Prerequisites**: Docker 24+, kubectl 1.28+, kind 0.20+, Terraform 1.5+, Go 1.24+, Python 3.11+ with Poetry 2.x, Node 20+

```bash
# 1. Provision a local Kind cluster
cd orchestrator && poetry install
poetry run cloudsentinel deploy --name dev --workers 2

# 2. Build and deploy the Go agent
cd ../agent && docker build -t cloudsentinel-agent:0.1.0 .
kind load docker-image cloudsentinel-agent:0.1.0 --name dev
kubectl apply -f deploy/daemonset.yaml

# 3. Start the API
cd ../api && poetry install
poetry run uvicorn app.main:app --port 8000

# 4. Start the dashboard (in another terminal)
cd ../dashboard && npm install && npm run dev

# 5. Browse to http://localhost:3000 and run your first scenario 🎉
```

Full walkthrough in [`docs/getting-started.md`](./docs/getting-started.md).

## ✨ What's inside

### Orchestrator (Python CLI)
- Typer-based CLI with Rich output (`deploy`, `destroy`, `list`, `chaos-run`)
- Strategy pattern over providers: `KindProvider` implemented, stubs for AKS/EKS/GKE
- Pydantic models for cluster/scenario configs with DNS-compatible name validation
- gRPC client auto-generated from `agent/proto/agent.proto`
- **24 pytest tests**, 100% coverage on config and chaos models

### Agent (Go)
- Dual-server design: HTTP `:9100` (Prometheus + healthz) + gRPC `:50051`
- CPU / memory / disk / network metrics via `gopsutil`
- `InjectFault` / `Rollback` / `Health` / `GetStatus` RPCs
- Distroless multi-stage Dockerfile → **24 MB image**, non-root `65532`
- Kubernetes DaemonSet manifest with Prometheus scrape annotations
- **6 Go tests** covering the gRPC server and the in-memory fault registry

### API (FastAPI + SQLite)
- OpenAPI auto-documented at `/docs`
- `GET/POST /scenarios`, `GET /scenarios/{id}`, `POST /scenarios/{id}/run`
- SQLModel ORM, scenarios persist across restarts
- CORS enabled for the dashboard, healthz endpoint for k8s probes
- Multi-stage Dockerfile (Poetry 2.3.4 → python-slim runtime, non-root user, HEALTHCHECK)
- **11 pytest tests** covering every endpoint + validation rules

### Dashboard (Next.js + shadcn/ui)
- Live scenarios table with 5s polling
- Create scenarios via a validated shadcn Dialog form
- Per-row **Run** button triggering the API
- Typed fetch client in `src/lib/api/` mirroring the Pydantic schemas

### ML Pipeline (scikit-learn)
- Reproducible 5,250-row synthetic dataset (5k normal + 250 anomalies)
- `StandardScaler` + `IsolationForest` wrapper with save/load to joblib
- Typer CLI: `generate-dataset`, `train`, `predict`
- **Precision 0.96 / Recall 1.00 / F1 0.98** on the synthetic test split
- **13 pytest tests**

## 🧪 Testing

| Suite | Tests | Runtime |
|-------|-------|---------|
| Python — orchestrator | 24 | ~1s |
| Go — agent | 6 | ~1s |
| Python — API (FastAPI + SQLite) | 11 | <1s |
| Python — ML pipeline | 13 | ~2s |
| **Total** | **54** | **~5s** |

All suites run on every push via three GitHub Actions workflows
(`python-ci.yml`, `go-ci.yml`, `terraform-ci.yml`).

## 🧰 Tech Stack

**Languages**: Python 3.11 · Go 1.24 · TypeScript 5 · Protobuf · Terraform 1.5
**Backend**: FastAPI · SQLModel · Pydantic v2 · Poetry 2.x · gRPC
**Frontend**: Next.js 16 · Tailwind CSS v4 · shadcn/ui · Radix · React hooks
**Infrastructure**: Kubernetes · Docker (multi-stage) · Kind · Prometheus
**ML**: scikit-learn (Isolation Forest) · pandas · numpy · joblib
**CI/CD**: GitHub Actions (3 pipelines) · pre-commit · Ruff · gofmt
**Docs**: MkDocs Material

## 📖 Documentation

Full documentation built with MkDocs Material:

```bash
pip install --user mkdocs mkdocs-material
mkdocs serve --dev-addr 127.0.0.1:8001
```

Pages: [Home](./docs/index.md) · [Getting Started](./docs/getting-started.md) · [Architecture](./docs/architecture.md) · [ML Pipeline](./docs/ml-pipeline.md)

## 🗺️ Roadmap

- [x] Project scaffolding + pre-commit hooks + CI/CD (3 pipelines)
- [x] Terraform module for Kind
- [x] Orchestrator Python CLI + provider abstraction
- [x] Go agent (Prometheus + gRPC) + K8s DaemonSet
- [x] Chaos engine + scenario YAML format
- [x] FastAPI backend with SQLite persistence
- [x] Next.js dashboard with live polling
- [x] ML anomaly detection pipeline
- [x] MkDocs documentation site
- [ ] Terraform modules for AKS / EKS / GKE
- [ ] Real fault injection in the agent (currently returns a skeleton response)
- [ ] Deploy MkDocs to GitHub Pages
- [ ] Record an end-to-end demo video

## 📄 License

Apache License 2.0 — see [LICENSE](./LICENSE)

## 🙋 About

Built by **Kane** as a deep-dive into distributed systems testing, Kubernetes
tooling, and end-to-end product engineering.

Contact: [mamoudoucheikhk@gmail.com](mailto:mamoudoucheikhk@gmail.com)

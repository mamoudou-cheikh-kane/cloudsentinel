# 🛡️ CloudSentinel

> Multi-cloud chaos engineering platform for validating distributed system resilience across Kubernetes clusters on AWS, GCP, Azure, and bare-metal environments.

[![Status](https://img.shields.io/badge/status-in%20development-yellow)](https://github.com/mamoudou-cheikh-kane/cloudsentinel)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![Python](https://img.shields.io/badge/python-3.11-blue)](https://www.python.org)
[![Go](https://img.shields.io/badge/go-1.21-blue)](https://go.dev)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.30-326CE5)](https://kubernetes.io)

## 🎯 What is CloudSentinel?

CloudSentinel automates the validation of cloud orchestration tools under adverse conditions. It deploys Kubernetes clusters across multiple environments, injects controlled failures (network latency, node loss, resource exhaustion), and uses machine learning to classify test results.

## ✨ Key Features

- **Multi-cloud orchestration** — deploy test clusters on AWS EKS, GCP GKE, Azure AKS, or local kind
- **Chaos engineering engine** — library of reproducible failure scenarios with automatic rollback
- **Real-time telemetry** — Go agents collect metrics and stream via Redis to TimescaleDB
- **ML-powered analysis** — automatically classify failures as flaky, regression, or infrastructure issues
- **Production-grade CI/CD** — GitHub Actions for fast feedback, Jenkins for long-running chaos tests
- **Interactive dashboard** — React frontend for exploring test runs and comparing cluster health

## 🏗️ Architecture
┌─────────────┐       ┌──────────────┐       ┌─────────────┐
│ Orchestrator│──────▶│  Terraform   │──────▶│  Cloud APIs │
│   (Python)  │       │   Modules    │       │ (AWS/GCP/..)│
└──────┬──────┘       └──────────────┘       └─────────────┘
│                                             │
▼                                             ▼
┌─────────────┐       ┌──────────────┐       ┌─────────────┐
│Chaos Engine │──────▶│ K8s Clusters │◀──────│ Agents (Go) │
└─────────────┘       └──────┬───────┘       └──────┬──────┘
│                      │
▼                      ▼
┌──────────────┐       ┌─────────────┐
│   FastAPI    │◀──────│Redis Streams│
└──────┬───────┘       └─────────────┘
│
┌──────────┼──────────┐
▼          ▼          ▼
┌─────────┐ ┌──────────┐ ┌────────┐
│Dashboard│ │TimescaleDB│ │ML Model│
│(React)  │ │          │ │        │
└─────────┘ └──────────┘ └────────┘
## 📦 Repository Structure

| Directory | Purpose |
|-----------|---------|
| `orchestrator/` | Python CLI to deploy clusters across providers |
| `agent/` | Go agent running as DaemonSet on cluster nodes |
| `chaos-engine/` | Fault injection library (network, CPU, pods, etc.) |
| `api/` | FastAPI backend for test results ingestion & query |
| `dashboard/` | React dashboard for visualizing runs |
| `ml-pipeline/` | Classification & anomaly detection models |
| `infrastructure/` | Terraform modules for each cloud provider |
| `ci/` | GitHub Actions workflows + Jenkinsfile |
| `docs/` | Architecture decisions and tutorials |

## 🚀 Quick Start

*Project in active development — not yet ready for external use.*

```bash
# Deploy a local test cluster (coming soon)
cloudsentinel deploy --config examples/configs/local.yaml

# Run a chaos scenario (coming soon)
cloudsentinel chaos run --scenario network-partition --cluster local-test
```

## 🗺️ Roadmap

- [x] Project scaffolding and tooling setup
- [ ] Terraform modules (kind, GKE, AKS, EKS)
- [ ] Python orchestrator CLI
- [ ] Go telemetry agent
- [ ] Chaos engine with 10+ scenarios
- [ ] CI/CD pipelines (GitHub Actions + Jenkins)
- [ ] FastAPI + TimescaleDB backend
- [ ] React dashboard
- [ ] ML classification pipeline
- [ ] Public demo deployment

## 🧪 Tech Stack

**Languages**: Python 3.11, Go 1.21, TypeScript
**Infrastructure**: Terraform, Kubernetes, Docker
**Backend**: FastAPI, PostgreSQL/TimescaleDB, Redis
**Frontend**: React, Vite, Tailwind, Recharts
**ML**: scikit-learn, XGBoost, MLflow
**CI/CD**: GitHub Actions, Jenkins
**Cloud**: AWS, GCP, Azure, kind

## 📖 Documentation

See [`docs/`](./docs) for architecture decisions (ADRs), tutorials, and API references.

## 📄 License

Apache License 2.0 — see [LICENSE](./LICENSE)

## 🙋 About

Built by Kane as a deep-dive into distributed systems testing.

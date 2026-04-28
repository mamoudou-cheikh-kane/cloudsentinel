# CloudSentinel

**Multi-cloud chaos engineering platform for Kubernetes.**

[![Release](https://img.shields.io/github/v/release/mamoudou-cheikh-kane/cloudsentinel?label=release)](https://github.com/mamoudou-cheikh-kane/cloudsentinel/releases)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/mamoudou-cheikh-kane/cloudsentinel/blob/main/LICENSE)

CloudSentinel is a **production-grade chaos engineering platform** that lets you:

- :material-server-network: **Provision** Kubernetes clusters (Kind, EKS, AKS, GKE) through a single Python CLI
- :material-fire: **Inject real faults** into any node (CPU stress, memory pressure, network latency, disk fill)
- :material-chart-line: **Collect** node-level metrics through a lightweight Go agent running as a DaemonSet
- :material-monitor-dashboard: **Drive** scenarios from a REST API, a web dashboard, or the command line
- :material-robot: **Detect** anomalies in collected metrics with an Isolation Forest model

!!! success "v0.2.0 highlight"
    All four fault types are now fully implemented and validated end-to-end on a real Kind cluster. See the [v0.2.0 release notes](https://github.com/mamoudou-cheikh-kane/cloudsentinel/releases/tag/v0.2.0) for the full changelog.

---

## Try it in three commands

```bash
git clone https://github.com/mamoudou-cheikh-kane/cloudsentinel
cd cloudsentinel
make demo
```

The `make demo` target spins up a local Kind cluster, builds and loads the agent image, deploys the DaemonSet, and exposes the agent's gRPC + Prometheus endpoints on `localhost`. About three minutes from clone to a working chaos engineering setup.

For the full hands-on guide with copy-paste commands for every fault type, head to the [Playground](playground.md).

---

## Components

| Component | Language | Role |
|-----------|----------|------|
| **Orchestrator** | Python + Poetry | CLI that provisions clusters via Terraform and runs chaos scenarios |
| **Agent** | Go 1.24 | DaemonSet pod exposing Prometheus metrics + gRPC fault injection (4 real fault types) |
| **API** | Python (FastAPI + SQLModel) | REST API persisting scenarios in SQLite |
| **Dashboard** | TypeScript (Next.js + Tailwind) | Web UI to list, create, and run scenarios |
| **ML Pipeline** | Python (scikit-learn) | Isolation Forest anomaly detection, F1 = 0.98 |

---

## What's verified

Every fault implementation is tested with measurements that **prove the behavior**, not just check return codes:

| Fault | Verified by | Result |
|-------|-------------|--------|
| **CPU Stress** | `syscall.Getrusage` | 1.5 CPU-seconds per 1 wall-second with 2 workers at 80% intensity |
| **Memory Pressure** | `runtime.MemStats` | 100 MB requested → 100 MB allocated → 0 MB after Stop+GC |
| **Disk Fill** | `os.Stat` | 10 MiB requested → 10485760 bytes exact, fsync verified |
| **Network Latency** | Mock TCExecutor | Idempotent Stop, no leaked qdiscs on Start failure |

35 unit tests pass in approximately 2 seconds. End-to-end validated on Kind v0.23.0.

---

## Quick links

- :material-rocket-launch: [Getting Started](getting-started.md) — install prerequisites and run the agent
- :material-gamepad-variant: [Playground](playground.md) — copy-paste commands for every fault type
- :material-blueprint: [Architecture](architecture.md) — Strategy Pattern, Registry, Mock TCExecutor
- :material-brain: [ML Pipeline](ml-pipeline.md) — Isolation Forest training and inference

---

## Roadmap

| Niveau | Status | Highlights |
|--------|--------|------------|
| **Niveau 1** — Real fault injection | :white_check_mark: **shipped (v0.2.0)** | 4 fault types, Strategy Pattern, robust rollback, make demo |
| **Niveau 2** — Scenarios as Code | :hourglass: in design | Declarative YAML with Prometheus assertions |
| **Niveau 3** — LLM-assisted | :hourglass: planned | Auto-generate hypotheses from cluster topology |
| **Niveau 4** — Helm + CRD | :hourglass: planned | GitOps-friendly deployment |
| **Niveau 5** — CNCF Sandbox | :hourglass: planned | Multi-region, KubeCon talk, comparison page |

---

## Why CloudSentinel?

Cloud-native systems fail in surprising ways. Chaos engineering is the practice of deliberately injecting those failures in controlled conditions so your team can observe and fix them **before** they happen in production.

Tools like Chaos Mesh and LitmusChaos exist, but I wanted to **understand the internals**: how do you actually consume CPU on a node? How do you allocate memory that the kernel does not swap? How do you guarantee rollback when the agent crashes?

CloudSentinel answers those questions with code you can read, tests that prove the behavior, and a single `make demo` to try it on your laptop.

# CloudSentinel

**Multi-cloud chaos engineering platform for Kubernetes.**

CloudSentinel is an end-to-end platform that lets you:

- Provision Kubernetes clusters (Kind, EKS, AKS, GKE) through a single CLI.
- Inject controlled faults (CPU stress, memory pressure, network latency,
  disk fill) into any node of a cluster.
- Collect node-level metrics through a lightweight Go agent running as a
  DaemonSet.
- Drive scenarios from a REST API, a web dashboard, or the command line.
- Detect anomalies in the collected metrics with an Isolation Forest model
  trained on historical data.

---

## Components

| Component | Language | Role |
|-----------|----------|------|
| **Orchestrator** | Python | CLI that provisions clusters and runs chaos scenarios |
| **Agent** | Go | DaemonSet pod exposing Prometheus metrics + gRPC fault injection |
| **API** | Python (FastAPI) | REST API that persists scenarios in SQLite |
| **Dashboard** | TypeScript (Next.js) | Web UI to list, create, and run scenarios |
| **ML Pipeline** | Python (scikit-learn) | Anomaly detection on node metrics |

---

## Quick links

- [Getting Started](getting-started.md) — install, run a local cluster,
  inject your first fault.
- [Architecture](architecture.md) — how the components fit together.
- [ML Pipeline](ml-pipeline.md) — train the anomaly detector and classify
  metrics.

---

## Why CloudSentinel?

Cloud-native systems fail in surprising ways. Chaos engineering is the
practice of deliberately injecting those failures in controlled conditions
so your team can observe and fix them **before** they happen in production.

CloudSentinel provides a reusable, typed, and testable toolkit to run
chaos experiments at the node level — and ships with a machine-learning
model that learns what "normal" looks like on your cluster.

# Architecture

CloudSentinel is split into five loosely-coupled services that communicate
over well-defined contracts (REST, gRPC, Prometheus).

## High-level diagram

```text
┌─────────────────────────────────────────────┐
│  USER (browser)                             │
└──────────────────┬──────────────────────────┘
                   │ HTTP
                   ▼
┌─────────────────────────────────────────────┐
│  DASHBOARD (Next.js + Tailwind + shadcn/ui) │
│  - Lists scenarios with 5s polling          │
│  - Create form in a Dialog                  │
│  - Per-row Run button                       │
└──────────────────┬──────────────────────────┘
                   │ REST (JSON)
                   ▼
┌─────────────────────────────────────────────┐
│  API (FastAPI + SQLModel + SQLite)          │
│  - GET/POST /scenarios                      │
│  - POST /scenarios/{id}/run                 │
│  - CORS for localhost:3000                  │
└──────────────────┬──────────────────────────┘
                   │ (orchestration)
                   ▼
┌─────────────────────────────────────────────┐
│  ORCHESTRATOR (Python CLI + chaos engine)   │
│  - Provisions clusters via Terraform        │
│  - Discovers agents through the K8s API     │
│  - Speaks gRPC to each agent                │
└──────────────────┬──────────────────────────┘
                   │ gRPC (protobuf)
                   ▼
┌─────────────────────────────────────────────┐
│  AGENT Go (DaemonSet on every node)         │
│  - HTTP :9100  /metrics (Prometheus)        │
│  - gRPC :50051 InjectFault/Rollback/Health  │
│  - Distroless image (~24 MB)                │
└─────────────────────────────────────────────┘
                   ▲
                   │ scraped by Prometheus / fed to ML
                   │
┌─────────────────────────────────────────────┐
│  ML PIPELINE (scikit-learn Isolation Forest)│
│  - Trains on historical metrics             │
│  - Classifies each observation normal/anom. │
└─────────────────────────────────────────────┘
```

## Design principles

- **Typed contracts everywhere.** The Go agent exposes a `.proto`
  definition; Python and TypeScript clients are generated or mirrored
  from it. Pydantic schemas on the API mirror the Go messages.
- **Separation of concerns.** The orchestrator never writes to the DB.
  The API never talks to Kubernetes. The agent does one thing: collect
  metrics and apply local faults.
- **One image per service.** Each service ships a multi-stage Dockerfile.
  The agent uses `distroless/static-debian12` for a 24 MB attack surface.
- **Reproducible builds.** Terraform modules, `poetry.lock`, `go.sum`, and
  `package-lock.json` lock every dependency. CI re-runs the full test
  matrix on each push.

## Data flow — running a scenario

1. User clicks **Run** in the dashboard.
2. The dashboard POSTs to `/scenarios/{id}/run` on the API.
3. The API reads the scenario from SQLite and records a `running` status.
4. (Planned) The API hands off to the orchestrator, which discovers
   matching agent pods and calls `InjectFault` on each.
5. Each agent applies the fault locally and reports back on the gRPC
   stream.
6. When the fault duration expires, the orchestrator calls `Rollback` on
   every agent it touched and the API transitions the scenario to
   `completed`.

The dashboard polls `/scenarios` every 5 seconds, so the status change
surfaces in the table without a manual refresh.

# Getting Started

A local setup: a Kind cluster, the agent DaemonSet, and the API + dashboard
running on your machine.

## Prerequisites

- Docker 24+
- kubectl 1.28+
- kind 0.20+
- Terraform 1.5+
- Go 1.24+
- Python 3.11+ and Poetry 2.x
- Node.js 20+

## 1. Provision a local cluster

```bash
cd orchestrator
poetry install
poetry run cloudsentinel deploy --name dev --workers 2
```

## 2. Build and deploy the agent

```bash
cd agent
docker build -t cloudsentinel-agent:0.1.0 .
kind load docker-image cloudsentinel-agent:0.1.0 --name dev
kubectl apply -f deploy/daemonset.yaml
```

## 3. Start the API

```bash
cd api
poetry install
poetry run uvicorn app.main:app --port 8000
```

OpenAPI docs at http://localhost:8000/docs.

## 4. Start the dashboard

```bash
cd dashboard
npm install
npm run dev
```

Dashboard at http://localhost:3000.

## 5. Run your first scenario

Via the dashboard or via the API:

```bash
curl -X POST http://localhost:8000/scenarios \
  -H 'Content-Type: application/json' \
  -d '{"name":"demo","faults":[{"type":"cpu_stress","duration_seconds":30}]}'

curl -X POST http://localhost:8000/scenarios/1/run
```

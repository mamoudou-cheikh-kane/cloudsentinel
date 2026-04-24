# CloudSentinel API

HTTP API that exposes CloudSentinel chaos engineering scenarios over REST.

## Run locally

```bash
poetry install
poetry run uvicorn app.main:app --reload --port 8000
```

Then browse to:

- http://localhost:8000/ — service info
- http://localhost:8000/healthz — liveness probe
- http://localhost:8000/docs — interactive Swagger UI

See the main [CloudSentinel README](../README.md) for the full project.

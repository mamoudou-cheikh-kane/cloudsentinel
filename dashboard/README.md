# CloudSentinel Dashboard

Next.js + Tailwind + shadcn/ui dashboard for the CloudSentinel API.

## Run locally

```bash
npm install
npm run dev
```

The dashboard expects the API to be reachable at
`NEXT_PUBLIC_API_BASE_URL` (default: `http://localhost:8000`).

Start the API in a separate terminal:

```bash
cd ../api
poetry run uvicorn app.main:app --port 8000
```

Then browse to http://localhost:3000.

## Features

- Live list of chaos scenarios with 5s polling
- Create a new scenario via a Dialog form
- Trigger a scenario run with per-row action buttons
- Typed API client in `src/lib/api/`

See the main [CloudSentinel README](../README.md) for the full project.

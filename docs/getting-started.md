# Getting Started

This guide walks you through running CloudSentinel locally on a Kind cluster, from prerequisites to your first injected fault.

!!! tip "Want the fastest path?"
    Skip straight to `make demo` (see below). The full manual procedure is documented for users who want to understand each step.

## Prerequisites

| Tool | Version | Why |
|------|---------|-----|
| **Docker** | 24+ | Build and run the agent image |
| **kubectl** | 1.28+ | Talk to the Kind cluster |
| **kind** | 0.20+ | Local Kubernetes cluster |
| **Go** | 1.24+ | Build the agent (only if you want to run tests locally) |
| **make** | any | Run the demo target |

You can verify everything in one command:

```bash
make check-prereqs
```

If something is missing, the script tells you exactly which tool needs to be installed and links to the install instructions.

## Option 1 — Run the demo (recommended)

The `make demo` target performs every step automatically:

```bash
git clone https://github.com/mamoudou-cheikh-kane/cloudsentinel
cd cloudsentinel
make demo
```

Six steps run in sequence:

1. Verifies that `docker`, `kind`, and `kubectl` are installed
2. Builds the agent Docker image (`cloudsentinel-agent:0.2.0`)
3. Creates a local Kind cluster named `cloudsentinel-demo`
4. Loads the freshly built image into the Kind node
5. Applies the DaemonSet manifest and waits for the rollout
6. Port-forwards `:50051` (gRPC) and `:9100` (Prometheus metrics) to localhost

After about three minutes the script prints `Demo running. The agent pod is healthy and exposing metrics.` and holds the port-forward open.

## Option 2 — Manual procedure

If you prefer to drive each step yourself or to integrate CloudSentinel into your own toolchain, here is the manual flow.

### 1. Build the agent image

```bash
cd agent
docker build -t cloudsentinel-agent:0.2.0 .
```

The Dockerfile is a multi-stage build (Alpine builder, Debian `tc-stage`, distroless runtime). The final image is around 56 MB and includes the `tc` binary plus its 10 runtime libraries.

### 2. Create a Kind cluster

```bash
kind create cluster --name cloudsentinel-demo --wait 60s
```

### 3. Load the image into Kind

```bash
kind load docker-image cloudsentinel-agent:0.2.0 --name cloudsentinel-demo
```

### 4. Apply the DaemonSet

```bash
kubectl apply -f agent/deploy/daemonset.yaml
kubectl -n cloudsentinel rollout status daemonset/cloudsentinel-agent --timeout=120s
```

### 5. Port-forward the agent

```bash
POD=$(kubectl -n cloudsentinel get pods -l app.kubernetes.io/name=cloudsentinel-agent \
  -o jsonpath='{.items[0].metadata.name}')

kubectl -n cloudsentinel port-forward "$POD" 50051:50051 9100:9100
```

## Verify it works

In another terminal, hit the health endpoint:

```bash
curl -s http://localhost:9100/healthz
```

Expected output:
ok

Inspect the metrics:

```bash
curl -s http://localhost:9100/metrics | head -20
```

You should see entries like `cloudsentinel_cpu_usage_percent`, `cloudsentinel_memory_*`, `cloudsentinel_disk_*`, and `cloudsentinel_agent_goroutines`.

## Inject your first fault

With the port-forward still running, send a real CPU stress fault via gRPC:

```bash
grpcurl -plaintext -d '{
  "type": "FAULT_TYPE_CPU_STRESS",
  "duration_seconds": 30,
  "parameters": {"intensity": "80", "workers": "4"}
}' localhost:50051 cloudsentinel.agent.v1.AgentService/InjectFault
```

The agent assigns a `fault_id` and starts the workers immediately. Watch the structured logs in another terminal:

```bash
kubectl -n cloudsentinel logs -l app.kubernetes.io/name=cloudsentinel-agent -f
```

You should see something like:

```json
{"time":"...","level":"INFO","msg":"fault registered","fault_id":"...","type":"cpu_stress","duration":30000000000}
{"time":"...","level":"INFO","msg":"fault injected","fault_id":"...","type":"cpu_stress","duration_seconds":30}
{"time":"...","level":"INFO","msg":"fault stopped","fault_id":"...","type":"cpu_stress"}
```

The agent automatically stops the fault after the configured duration and cleans up. For copy-paste commands covering memory pressure, disk fill, and network latency, see the [Playground](playground.md).

## Troubleshooting

### The pod stays in `Init:0/1` for a long time

The init container pulls `registry.k8s.io/build-image/debian-iptables:bookworm-v1.0.0` the first time, which weighs around 50 MB. On a slow connection this can take a minute or two. Subsequent runs reuse the cached image.

### `make demo` fails at step 4 with `error: timed out waiting for the condition`

Same root cause as above: the init container is still pulling. Wait a minute and re-run `make demo`. The cluster, image, and manifests are already in place, so the second run picks up where the first stopped.

### `tc: command not found` inside the pod

You probably built an older agent image. Make sure you are running `cloudsentinel-agent:0.2.0` or later, which copies `tc` and its runtime libraries from a Debian `tc-stage` into the distroless runtime.

### `permission denied` when invoking tc

The pod needs `CAP_NET_ADMIN`. The DaemonSet manifest in `agent/deploy/daemonset.yaml` adds it back after dropping all capabilities. If you customized the manifest, double-check the `securityContext.capabilities.add` block contains `NET_ADMIN`.

## Cleanup

When you are done, tear down the cluster:

```bash
make demo-clean
```

This deletes the Kind cluster and frees up Docker resources. Your machine returns to its previous state.

## Next steps

- :material-gamepad-variant: [Playground](playground.md) — copy-paste commands for every fault type
- :material-blueprint: [Architecture](architecture.md) — how the agent works internally
- :material-brain: [ML Pipeline](ml-pipeline.md) — train the anomaly detector on collected metrics

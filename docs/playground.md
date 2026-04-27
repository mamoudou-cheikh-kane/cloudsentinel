# 🎮 Playground

Try CloudSentinel in **three commands** on your local machine. Everything below is real, not a simulation — these commands create a working Kubernetes cluster, deploy the agent, and inject genuine faults that consume CPU, memory, disk, and network bandwidth.

!!! note "Prerequisites"
    You need **Docker**, **kind** ≥ 0.20, **kubectl** ≥ 1.28, and **Go** 1.24+ installed locally. Run `make check-prereqs` to verify your setup.

## ⚡ Quick start

Clone the repository, change into it, and run the demo target:

git clone https://github.com/mamoudou-cheikh-kane/cloudsentinel
cd cloudsentinel
make demo

The `make demo` target performs six steps automatically:

1. Verifies that `docker`, `kind`, and `kubectl` are installed
2. Builds the agent Docker image (`cloudsentinel-agent:0.2.0`)
3. Creates a local Kind cluster named `cloudsentinel-demo`
4. Loads the freshly built image into the Kind node
5. Applies the DaemonSet manifest and waits for rollout
6. Port-forwards `:50051` (gRPC) and `:9100` (Prometheus metrics)

After about three minutes, you have a working chaos engineering setup on your laptop.

## 💥 Inject your first fault

Once `make demo` is running and the agent pod is `Ready`, open **another terminal** and inject a real CPU stress fault:

=== "CPU Stress"

grpcurl -plaintext -d '{
  "type": "FAULT_TYPE_CPU_STRESS",
  "duration_seconds": 30,
  "parameters": {"intensity": "80", "workers": "4"}
}' localhost:50051 cloudsentinel.agent.v1.AgentService/InjectFault

    The agent will spawn 4 worker goroutines that consume 80% of CPU for 30 seconds, then auto-cleanup. Verified by `syscall.Getrusage`: 1.5 CPU-seconds per 1 wall-second.

=== "Memory Pressure"

grpcurl -plaintext -d '{
  "type": "FAULT_TYPE_MEMORY_PRESSURE",
  "duration_seconds": 30,
  "parameters": {"size_mb": "200"}
}' localhost:50051 cloudsentinel.agent.v1.AgentService/InjectFault

    Allocates 200 MB and touches every 4 KiB page every 200 ms to keep it resident. Verified by `runtime.MemStats`: 200 MB requested → 200 MB allocated → 0 MB after Stop+GC.

=== "Disk Fill"

grpcurl -plaintext -d '{
  "type": "FAULT_TYPE_DISK_FILL",
  "duration_seconds": 60,
  "parameters": {"size_mb": "100"}
}' localhost:50051 cloudsentinel.agent.v1.AgentService/InjectFault

    Writes 100 MiB in 1 MiB chunks with `fsync` to make sure the bytes hit durable storage. Verified by `os.Stat`: 100 MiB requested → 104857600 bytes exact.

=== "Network Latency"

grpcurl -plaintext -d '{
  "type": "FAULT_TYPE_NETWORK_LATENCY",
  "duration_seconds": 30,
  "parameters": {"delay_ms": "200", "jitter_ms": "20", "interface": "eth0"}
}' localhost:50051 cloudsentinel.agent.v1.AgentService/InjectFault

    Adds 200 ms ± 20 ms of latency to the node's `eth0` interface using `tc` and `netem`. The fault is removed cleanly on Stop, and the init container cleans orphan rules at pod startup.

## 📺 Watch the demo

A short video walkthrough showing `make demo` from clone to fault injection is coming soon. In the meantime, you can follow the steps above on your own machine.

## 🔍 Verify it really works

While a fault is running, watch the agent's logs in real time:

kubectl -n cloudsentinel logs -l app.kubernetes.io/name=cloudsentinel-agent -f

You should see structured JSON logs like:

{"time":"2026-04-27T21:11:40Z","level":"INFO","msg":"fault registered","fault_id":"daef1638-9e31-44fc-ab3a-c1e8c6b232d3","type":"cpu_stress","duration":60000000000}
{"time":"2026-04-27T21:11:40Z","level":"INFO","msg":"fault injected","fault_id":"daef1638-9e31-44fc-ab3a-c1e8c6b232d3","type":"cpu_stress","duration_seconds":60}
{"time":"2026-04-27T21:12:39Z","level":"INFO","msg":"fault stopped","fault_id":"daef1638-9e31-44fc-ab3a-c1e8c6b232d3","type":"cpu_stress"}

The auto-cleanup happens at exactly the configured duration, regardless of whether the gRPC client is still connected.

## 📊 What's verified

Every fault implementation is tested with measurements that **prove the behavior**, not just check return codes:

| Fault | Verified by | Result |
|-------|-------------|--------|
| **CPU Stress** | `syscall.Getrusage` measures consumed CPU-seconds | 1.5 CPU-seconds per 1 wall-second with 2 workers at 80% |
| **Memory Pressure** | `runtime.MemStats` measures heap allocation | 100 MB requested → 100 MB allocated → 0 MB after Stop+GC |
| **Disk Fill** | `os.Stat` measures file size on disk | 10 MiB requested → 10485760 bytes exact, fsync verified |
| **Network Latency** | Mock TCExecutor verifies tc invocation | Idempotent Stop, no leaked qdiscs on Start failure |

35 unit tests pass in approximately 2 seconds. End-to-end validated on Kind v0.23.0 with WSL2 Ubuntu 22.04.

## 🏗️ What's happening under the hood

When you call `InjectFault` on the gRPC API, here is what happens inside the agent:

1. The gRPC server validates the request (type, duration, parameters)
2. A `Fault` object is constructed by the right factory (`NewCPUStress`, `NewMemoryPressure`, etc.)
3. The Registry assigns a UUID, stores the fault, and starts a `time.AfterFunc` timer
4. `Fault.Start(ctx)` is called: worker goroutines spawn, memory is allocated, file is created, or `tc` rules are added
5. The gRPC response returns immediately with the assigned `fault_id`
6. After the configured duration, the timer fires and `Fault.Stop(ctx)` runs the cleanup
7. Optionally, the client can call `Rollback(fault_id)` early to stop the fault before its timer fires

Each `Fault` implementation is a separate Go file in `internal/faults/`. They share the same interface (`ID`, `Type`, `Start`, `Stop`, `StartedAt`, `Duration`) so adding a new fault type only requires writing the new file plus one line in the `buildFault` dispatcher.

## 🛠️ Clean up

When you are done, tear down the demo cluster:

make demo-clean

This deletes the Kind cluster and frees up the local Docker resources. Your machine is back to its previous state.

## 🚀 Next steps

* Read the [Architecture](architecture.md) page for the full design overview
* Browse the source code on [GitHub](https://github.com/mamoudou-cheikh-kane/cloudsentinel)
* Check the [Roadmap](https://github.com/mamoudou-cheikh-kane/cloudsentinel#-roadmap) for what is coming in Niveau 2

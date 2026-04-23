# CloudSentinel Agent

Lightweight Go agent that runs as a Kubernetes DaemonSet on each node.
Exposes Prometheus metrics on `:9100` and a gRPC fault injection API on `:50051`.

See main [CloudSentinel README](../README.md) for project overview.

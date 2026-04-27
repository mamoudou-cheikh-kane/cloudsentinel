#!/usr/bin/env bash
# CloudSentinel end-to-end demo on a local Kind cluster.
#
# Steps:
#   1. (Re)create a Kind cluster named $1
#   2. Load the agent image $2 into the cluster
#   3. Apply the agent DaemonSet
#   4. Wait for the agent pod to become Ready
#   5. Forward the gRPC and metrics ports locally
#   6. Show the agent's Prometheus metrics
#   7. Hold the port-forward until the user hits Ctrl+C

set -euo pipefail

CLUSTER_NAME="${1:-cloudsentinel-demo}"
AGENT_IMAGE="${2:-cloudsentinel-agent:0.2.0}"
NAMESPACE="cloudsentinel"

color_blue()   { printf '\033[0;34m%s\033[0m\n' "$*"; }
color_green()  { printf '\033[0;32m%s\033[0m\n' "$*"; }
color_yellow() { printf '\033[0;33m%s\033[0m\n' "$*"; }

step() {
  echo ""
  color_blue ">>> $*"
}

# ---- 1. Create the Kind cluster ----
step "1/6 - Creating Kind cluster '$CLUSTER_NAME'"
if kind get clusters 2>/dev/null | grep -q "^$CLUSTER_NAME$"; then
  color_yellow "    cluster already exists, skipping creation"
else
  kind create cluster --name "$CLUSTER_NAME" --wait 60s
fi
kubectl cluster-info --context "kind-$CLUSTER_NAME" >/dev/null
color_green "    cluster ready"

# ---- 2. Load the agent image ----
step "2/6 - Loading agent image '$AGENT_IMAGE' into Kind"
kind load docker-image "$AGENT_IMAGE" --name "$CLUSTER_NAME"
color_green "    image loaded"

# ---- 3. Apply the DaemonSet ----
step "3/6 - Applying the agent DaemonSet"
kubectl apply -f agent/deploy/daemonset.yaml
color_green "    manifests applied"

# ---- 4. Wait for the agent to be ready ----
step "4/6 - Waiting for the agent pod to become Ready"
kubectl -n "$NAMESPACE" rollout status daemonset/cloudsentinel-agent --timeout=120s
POD=$(kubectl -n "$NAMESPACE" get pods -l app.kubernetes.io/name=cloudsentinel-agent \
  -o jsonpath='{.items[0].metadata.name}')
color_green "    agent pod: $POD"

# ---- 5. Port-forward gRPC and metrics ----
step "5/6 - Port-forwarding agent ports (50051 grpc, 9100 metrics)"
kubectl -n "$NAMESPACE" port-forward "$POD" 50051:50051 9100:9100 \
  >/tmp/cloudsentinel-pf.log 2>&1 &
PF_PID=$!
trap 'kill $PF_PID 2>/dev/null || true' EXIT INT TERM
sleep 2

# ---- 6. Show metrics + healthz ----
step "6/6 - Querying agent metrics and health"

echo ""
color_yellow "    GET /healthz:"
curl -s http://localhost:9100/healthz
echo ""

echo ""
color_yellow "    GET /metrics (first 20 lines):"
curl -s http://localhost:9100/metrics | head -n 20

echo ""
color_green ">>> Demo running. The agent pod is healthy and exposing metrics."
echo ""
echo "    Pod:        $POD"
echo "    Namespace:  $NAMESPACE"
echo "    Metrics:    http://localhost:9100/metrics"
echo "    Health:     http://localhost:9100/healthz"
echo "    gRPC:       localhost:50051"
echo ""
echo "    To inject a fault from another terminal:"
echo "      grpcurl -plaintext -d '{\"type\":\"FAULT_TYPE_CPU_STRESS\",\"duration_seconds\":15}' \\"
echo "              localhost:50051 cloudsentinel.agent.v1.AgentService/InjectFault"
echo ""
echo "    To clean up:"
echo "      make demo-clean"
echo ""
color_yellow "    Press Ctrl+C to stop the port-forward and exit."
wait $PF_PID

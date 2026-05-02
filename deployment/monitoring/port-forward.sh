#!/usr/bin/env bash
set -euo pipefail

echo "Prometheus: http://localhost:9090"
echo "Grafana:    http://localhost:3000"
echo
echo "Hint (from your laptop):"
echo "  ssh -L 3000:localhost:3000 -L 9090:localhost:9090 root@<VPS_IP>"
echo

kubectl -n monitoring port-forward svc/prometheus-server 9090:80 &
PROM_PID=$!

kubectl -n monitoring port-forward svc/grafana 3000:80 &
GRAF_PID=$!

cleanup() {
  kill "$PROM_PID" "$GRAF_PID" 2>/dev/null || true
}
trap cleanup EXIT

wait


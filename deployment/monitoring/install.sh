#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$ROOT_DIR"

kubectl get ns monitoring >/dev/null 2>&1 || kubectl create namespace monitoring

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts >/dev/null
helm repo add grafana https://grafana.github.io/helm-charts >/dev/null
helm repo update >/dev/null

echo "Installing/Upgrading Prometheus..."
helm upgrade --install prometheus prometheus-community/prometheus \
  -n monitoring \
  -f deployment/monitoring/values-prometheus.yaml

echo "Installing/Upgrading Grafana..."
helm upgrade --install grafana grafana/grafana \
  -n monitoring \
  -f deployment/monitoring/values-grafana.yaml

echo
echo "Waiting for rollouts..."
kubectl -n monitoring rollout status deploy/prometheus-server --timeout=180s
kubectl -n monitoring rollout status deploy/grafana --timeout=180s

echo
echo "Done."
echo "Grafana password:"
kubectl -n monitoring get secret grafana -o jsonpath='{.data.admin-password}' | base64 -d; echo


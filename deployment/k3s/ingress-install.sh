#!/usr/bin/env bash
set -euo pipefail

# Installs ingress-nginx into k3s.

kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.1/deploy/static/provider/cloud/deploy.yaml

echo "Waiting for ingress-nginx controller to be ready..."
kubectl -n ingress-nginx rollout status deploy/ingress-nginx-controller --timeout=180s

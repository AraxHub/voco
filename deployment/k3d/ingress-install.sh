#!/usr/bin/env bash
set -euo pipefail

kubectl create namespace ingress-nginx >/dev/null 2>&1 || true

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx >/dev/null 2>&1 || true
helm repo update >/dev/null

helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --set controller.service.type=LoadBalancer \
  --wait

echo "OK: ingress-nginx installed"


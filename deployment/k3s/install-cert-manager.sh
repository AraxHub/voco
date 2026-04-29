#!/usr/bin/env bash
# Устанавливает cert-manager (нужен для Let's Encrypt в кластере).
# Официальный манифест: https://cert-manager.io/docs/installation/kubernetes/
set -euo pipefail

kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.17.1/cert-manager.yaml

echo "Ждём готовность cert-manager…"
kubectl -n cert-manager rollout status deploy/cert-manager --timeout=120s
kubectl -n cert-manager rollout status deploy/cert-manager-webhook --timeout=120s
echo "Готово. Дальше: kubectl apply -f deployment/k8s/overlays/k3s/letsencrypt-clusterissuer.yaml"

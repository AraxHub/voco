#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-voco}"

echo "==> Recreate k3d cluster: ${CLUSTER_NAME}"
./deployment/k3d/cluster-delete.sh || true
./deployment/k3d/cluster-create.sh

echo "==> Install ingress-nginx"
./deployment/k3d/ingress-install.sh

echo "==> Build images"
docker build -t voco-backend:local -f deployment/voco-local/Dockerfile .
docker build -t voco-frontend:local -f deployment/k8s/images/frontend/Dockerfile .

echo "==> Import images into k3d"
k3d image import voco-backend:local -c "${CLUSTER_NAME}"
k3d image import voco-frontend:local -c "${CLUSTER_NAME}"

echo "==> Apply manifests"
./deployment/k3d/apply.sh

echo "OK: cluster recreated and app deployed"


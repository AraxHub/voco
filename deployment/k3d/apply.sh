#!/usr/bin/env bash
set -euo pipefail

kubectl apply -k deployment/k8s/overlays/k3d
echo "OK: applied k3d overlay"


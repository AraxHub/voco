#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-voco}"
k3d cluster delete "${CLUSTER_NAME}"
echo "OK: k3d cluster ${CLUSTER_NAME} deleted"


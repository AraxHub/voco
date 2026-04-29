#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-voco}"

# Stops k3d node containers (server/lb) so they don't consume CPU/RAM.
k3d cluster stop "${CLUSTER_NAME}"
echo "OK: k3d cluster ${CLUSTER_NAME} stopped"
echo "To start again: k3d cluster start ${CLUSTER_NAME}"


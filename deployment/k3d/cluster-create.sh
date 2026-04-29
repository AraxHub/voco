#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-voco}"

# Ports exposed from k3d loadbalancer to host
# - 80: ingress
# - 7880: LiveKit signaling (ws)
# - 62000-62020/udp: LiveKit WebRTC media
k3d cluster create "${CLUSTER_NAME}" \
  --servers 1 \
  --agents 0 \
  --k3s-arg "--disable=traefik@server:0" \
  --port "80:80@loadbalancer" \
  --port "7880:7880@loadbalancer" \
  --port "62000-62020:62000-62020/udp@loadbalancer"

kubectl config use-context "k3d-${CLUSTER_NAME}"

echo "OK: k3d cluster ${CLUSTER_NAME} created"


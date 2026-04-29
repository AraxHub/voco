#!/usr/bin/env bash
set -euo pipefail

# Installs k3s on a single-node VPS.
# - Disables Traefik and ServiceLB (we use ingress-nginx)
# - Expands NodePort range so we can use 7880/7881 and 62000-62020 as nodePorts for LiveKit

if [[ ${EUID:-0} -ne 0 ]]; then
  echo "Run as root (sudo)." >&2
  exit 1
fi

export INSTALL_K3S_EXEC=${INSTALL_K3S_EXEC:-"server --disable traefik --disable servicelb --service-node-port-range=1-65535"}

curl -sfL https://get.k3s.io | sh -

echo
echo "k3s installed."
echo "kubeconfig: /etc/rancher/k3s/k3s.yaml"
echo
echo "To use kubectl as your user:"
echo "  sudo cat /etc/rancher/k3s/k3s.yaml > ~/.kube/config"
echo "  sudo chown \$(id -u):\$(id -g) ~/.kube/config"

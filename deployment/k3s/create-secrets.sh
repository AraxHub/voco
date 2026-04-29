#!/usr/bin/env bash
set -euo pipefail

# Creates/updates Kubernetes Secret `voco-secrets` from an env file.
# Usage:
#   cp deployment/k3s/secrets.env.example deployment/k3s/secrets.env
#   edit deployment/k3s/secrets.env
#   ./deployment/k3s/create-secrets.sh

ENV_FILE=${ENV_FILE:-deployment/k3s/secrets.env}

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Env file not found: $ENV_FILE" >&2
  echo "Create it from deployment/k3s/secrets.env.example" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

kubectl create namespace voco --dry-run=client -o yaml | kubectl apply -f -

kubectl -n voco create secret generic voco-secrets \
  --from-literal=VOCO_LIVEKIT_API_KEY="${VOCO_LIVEKIT_API_KEY:?missing VOCO_LIVEKIT_API_KEY}" \
  --from-literal=VOCO_LIVEKIT_API_SECRET="${VOCO_LIVEKIT_API_SECRET:?missing VOCO_LIVEKIT_API_SECRET}" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Secret voco-secrets applied."

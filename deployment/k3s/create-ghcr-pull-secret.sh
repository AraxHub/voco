#!/usr/bin/env bash
set -euo pipefail

# Creates/updates docker-registry Secret for pulling private GHCR images.
# Reads only GHCR_* from secrets.env (does not `source` the whole file —
# VOCO_PG_URL may contain & which breaks shell sourcing).
#
#   GHCR_USERNAME=AraxHub
#   GHCR_TOKEN=ghp_...   # classic PAT: read:packages
#
#   ./deployment/k3s/create-ghcr-pull-secret.sh
#   ./deployment/k3s/apply.sh

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$ROOT_DIR"

ENV_FILE=${ENV_FILE:-deployment/k3s/secrets.env}

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Env file not found: $ENV_FILE" >&2
  echo "Create it from deployment/k3s/secrets.env.example" >&2
  exit 1
fi

read_env() {
  local key="$1"
  local line
  line=$(grep -E "^${key}=" "$ENV_FILE" | tail -n1 || true)
  if [[ -z "$line" ]]; then
    echo ""
    return
  fi
  printf '%s\n' "${line#*=}"
}

GHCR_USERNAME=$(read_env GHCR_USERNAME)
GHCR_TOKEN=$(read_env GHCR_TOKEN)
GHCR_EMAIL=$(read_env GHCR_EMAIL)
GHCR_EMAIL="${GHCR_EMAIL:-noreply@users.noreply.github.com}"

if [[ -z "$GHCR_USERNAME" || "$GHCR_USERNAME" == "CHANGE_ME" ]]; then
  echo "Missing GHCR_USERNAME in $ENV_FILE" >&2
  exit 1
fi
if [[ -z "$GHCR_TOKEN" || "$GHCR_TOKEN" == "CHANGE_ME" ]]; then
  echo "Missing GHCR_TOKEN in $ENV_FILE" >&2
  exit 1
fi

kubectl create namespace voco --dry-run=client -o yaml | kubectl apply -f -

kubectl -n voco create secret docker-registry ghcr-pull \
  --docker-server=ghcr.io \
  --docker-username="${GHCR_USERNAME}" \
  --docker-password="${GHCR_TOKEN}" \
  --docker-email="${GHCR_EMAIL}" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Secret applied: voco/ghcr-pull"
echo "Next: ./deployment/k3s/apply.sh"

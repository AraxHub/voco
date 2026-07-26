#!/usr/bin/env bash
set -euo pipefail

# Build local images on the VPS for the first deployment stage.
# Later you can switch to a registry and stop building on the server.
#
# Optional: source deployment/k3s/frontend-build.env to override VITE_* defaults.

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$ROOT_DIR"

FRONTEND_ENV="${FRONTEND_ENV:-deployment/k3s/frontend-build.env}"
if [[ -f "$FRONTEND_ENV" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$FRONTEND_ENV"
  set +a
fi

VITE_API_BASE_URL="${VITE_API_BASE_URL:-}"
VITE_KEYCLOAK_URL="${VITE_KEYCLOAK_URL:-https://auth.voco-online.ru}"
VITE_KEYCLOAK_REALM="${VITE_KEYCLOAK_REALM:-voco}"
VITE_KEYCLOAK_CLIENT_ID="${VITE_KEYCLOAK_CLIENT_ID:-voco-frontend}"

echo "Building backend image..."
docker build -t voco-backend:local -f deployment/k8s/images/backend/Dockerfile .

echo "Building frontend image..."
echo "  VITE_KEYCLOAK_URL=${VITE_KEYCLOAK_URL}"
echo "  VITE_KEYCLOAK_REALM=${VITE_KEYCLOAK_REALM}"
echo "  VITE_KEYCLOAK_CLIENT_ID=${VITE_KEYCLOAK_CLIENT_ID}"
docker build -t voco-frontend:local -f deployment/k8s/images/frontend/Dockerfile \
  --build-arg "VITE_API_BASE_URL=${VITE_API_BASE_URL}" \
  --build-arg "VITE_KEYCLOAK_URL=${VITE_KEYCLOAK_URL}" \
  --build-arg "VITE_KEYCLOAK_REALM=${VITE_KEYCLOAK_REALM}" \
  --build-arg "VITE_KEYCLOAK_CLIENT_ID=${VITE_KEYCLOAK_CLIENT_ID}" \
  .

echo "Building Keycloak image..."
docker build -t voco-keycloak:local -f deployment/k8s/images/keycloak/Dockerfile .

if command -v k3s >/dev/null 2>&1; then
  echo "Importing images into k3s containerd..."
  docker save voco-backend:local voco-frontend:local voco-keycloak:local | sudo k3s ctr images import -
else
  echo "WARN: k3s not found — skipped containerd import." >&2
fi

echo "Done."
echo "If deployments already exist with :local tag:"
echo "  kubectl -n voco delete pod -l app=voco-backend,app=voco-frontend"
echo "  kubectl -n voco rollout restart deploy/keycloak"

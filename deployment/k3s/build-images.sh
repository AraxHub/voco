#!/usr/bin/env bash
set -euo pipefail

# Build local images on the VPS for the first deployment stage.
# Later you can switch to a registry and stop building on the server.

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)

cd "$ROOT_DIR"

echo "Building backend image..."
docker build -t voco-backend:local -f deployment/k8s/images/backend/Dockerfile .

echo "Building frontend image..."
docker build -t voco-frontend:local -f deployment/k8s/images/frontend/Dockerfile .

echo "Done."

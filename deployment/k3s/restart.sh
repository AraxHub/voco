#!/usr/bin/env bash
set -euo pipefail

# Force backend + frontend pods to pick up current image tags.
kubectl -n voco rollout restart deploy/voco-backend deploy/voco-frontend
kubectl -n voco rollout status deploy/voco-backend --timeout=180s
kubectl -n voco rollout status deploy/voco-frontend --timeout=180s

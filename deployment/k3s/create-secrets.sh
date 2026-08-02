#!/usr/bin/env bash
set -euo pipefail

# Creates/updates Kubernetes Secrets from deployment/k3s/secrets.env
# Usage:
#   cp deployment/k3s/secrets.env.example deployment/k3s/secrets.env
#   edit deployment/k3s/secrets.env  # fill CHANGE_ME values
#   ./deployment/k3s/create-secrets.sh
#
# Values may contain & ? # — file is parsed as KEY=VALUE lines (not `source`),
# so VOCO_PG_URL does not need shell quoting (quoting still OK).

ENV_FILE=${ENV_FILE:-deployment/k3s/secrets.env}

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Env file not found: $ENV_FILE" >&2
  echo "Create it from deployment/k3s/secrets.env.example" >&2
  exit 1
fi

# Safe KEY=VALUE loader (handles & in URLs; strips optional surrounding quotes).
while IFS= read -r line || [[ -n "$line" ]]; do
  [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
  if [[ "$line" =~ ^([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
    key="${BASH_REMATCH[1]}"
    val="${BASH_REMATCH[2]}"
    if [[ "$val" =~ ^\"(.*)\"$ ]]; then
      val="${BASH_REMATCH[1]}"
    elif [[ "$val" =~ ^\'(.*)\'$ ]]; then
      val="${BASH_REMATCH[1]}"
    fi
    printf -v "$key" '%s' "$val"
    export "$key"
  fi
done < "$ENV_FILE"

require() {
  local name="$1"
  local val="${!name-}"
  if [[ -z "$val" || "$val" == "CHANGE_ME" ]]; then
    echo "Missing or placeholder secret: $name (set a real value in $ENV_FILE)" >&2
    exit 1
  fi
}

require VOCO_LIVEKIT_API_KEY
require VOCO_LIVEKIT_API_SECRET
require VOCO_PG_URL
require KC_BOOTSTRAP_ADMIN_PASSWORD
require KC_DB_PASSWORD
require KC_SMTP_PASSWORD
require VOCO_KEYCLOAK_ADMIN_CLIENT_SECRET
require VOCO_WEBPUSH_VAPID_PUBLIC
require VOCO_WEBPUSH_VAPID_PRIVATE

KC_BOOTSTRAP_ADMIN_USERNAME="${KC_BOOTSTRAP_ADMIN_USERNAME:-admin}"
KC_SMTP_USER="${KC_SMTP_USER:-resend}"

kubectl create namespace voco --dry-run=client -o yaml | kubectl apply -f -

kubectl -n voco create secret generic voco-secrets \
  --from-literal=VOCO_LIVEKIT_API_KEY="${VOCO_LIVEKIT_API_KEY}" \
  --from-literal=VOCO_LIVEKIT_API_SECRET="${VOCO_LIVEKIT_API_SECRET}" \
  --from-literal=VOCO_PG_URL="${VOCO_PG_URL}" \
  --from-literal=VOCO_KEYCLOAK_ADMIN_CLIENT_SECRET="${VOCO_KEYCLOAK_ADMIN_CLIENT_SECRET}" \
  --from-literal=VOCO_WEBPUSH_VAPID_PUBLIC="${VOCO_WEBPUSH_VAPID_PUBLIC}" \
  --from-literal=VOCO_WEBPUSH_VAPID_PRIVATE="${VOCO_WEBPUSH_VAPID_PRIVATE}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n voco create secret generic keycloak-secrets \
  --from-literal=KC_BOOTSTRAP_ADMIN_USERNAME="${KC_BOOTSTRAP_ADMIN_USERNAME}" \
  --from-literal=KC_BOOTSTRAP_ADMIN_PASSWORD="${KC_BOOTSTRAP_ADMIN_PASSWORD}" \
  --from-literal=KC_DB_PASSWORD="${KC_DB_PASSWORD}" \
  --from-literal=KC_SMTP_USER="${KC_SMTP_USER}" \
  --from-literal=KC_SMTP_PASSWORD="${KC_SMTP_PASSWORD}" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Secrets applied: voco-secrets, keycloak-secrets"
echo "Remember: LiveKit keys in secrets must match livekit-configmap.yaml keys:"
echo "  ${VOCO_LIVEKIT_API_KEY}: <same-secret>"

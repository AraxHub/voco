#!/bin/sh
set -eu

KC_URL="${KC_URL:-http://keycloak:8080}"
KC_ADMIN="${KC_ADMIN:-admin}"
KC_ADMIN_PASSWORD="${KC_ADMIN_PASSWORD:-admin}"
KC_REALM="${KC_REALM:-voco}"

KC_SMTP_HOST="${KC_SMTP_HOST:-mailpit}"
KC_SMTP_PORT="${KC_SMTP_PORT:-1025}"
KC_SMTP_FROM="${KC_SMTP_FROM:-noreply@voco.local}"
KC_SMTP_FROM_DISPLAY_NAME="${KC_SMTP_FROM_DISPLAY_NAME:-VOCO}"
KC_SMTP_USER="${KC_SMTP_USER:-}"
KC_SMTP_PASSWORD="${KC_SMTP_PASSWORD:-}"
KC_SMTP_SSL="${KC_SMTP_SSL:-false}"
KC_SMTP_STARTTLS="${KC_SMTP_STARTTLS:-false}"
KC_SMTP_AUTH="${KC_SMTP_AUTH:-false}"

echo "keycloak setup: waiting for admin API at ${KC_URL}..."

for i in $(seq 1 60); do
  if /opt/keycloak/bin/kcadm.sh config credentials \
    --server "${KC_URL}" \
    --realm master \
    --user "${KC_ADMIN}" \
    --password "${KC_ADMIN_PASSWORD}" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

/opt/keycloak/bin/kcadm.sh config credentials \
  --server "${KC_URL}" \
  --realm master \
  --user "${KC_ADMIN}" \
  --password "${KC_ADMIN_PASSWORD}"

json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

# SMTP must be sent as one JSON object (dotted smtpServer.* attrs do not always persist via kcadm).
SMTP_JSON=$(cat <<EOF
{
  "loginTheme": "voco",
  "accountTheme": "voco",
  "internationalizationEnabled": true,
  "defaultLocale": "ru",
  "supportedLocales": ["en", "ru"],
  "webAuthnPolicyPasswordlessPasskeysEnabled": false,
  "verifyEmail": true,
  "resetPasswordAllowed": true,
  "smtpServer": {
    "host": "$(json_escape "${KC_SMTP_HOST}")",
    "port": "$(json_escape "${KC_SMTP_PORT}")",
    "from": "$(json_escape "${KC_SMTP_FROM}")",
    "fromDisplayName": "$(json_escape "${KC_SMTP_FROM_DISPLAY_NAME}")",
    "ssl": "$(json_escape "${KC_SMTP_SSL}")",
    "starttls": "$(json_escape "${KC_SMTP_STARTTLS}")",
    "auth": "$(json_escape "${KC_SMTP_AUTH}")",
    "user": "$(json_escape "${KC_SMTP_USER}")",
    "password": "$(json_escape "${KC_SMTP_PASSWORD}")"
  }
}
EOF
)
echo "${SMTP_JSON}" >/tmp/voco-realm-smtp.json
/opt/keycloak/bin/kcadm.sh update "realms/${KC_REALM}" -f /tmp/voco-realm-smtp.json
rm -f /tmp/voco-realm-smtp.json

# Disable passwordless WebAuthn registration if present on an existing realm.
/opt/keycloak/bin/kcadm.sh update "authentication/required-actions/webauthn-register-passwordless" \
  -r "${KC_REALM}" \
  -s enabled=false >/dev/null 2>&1 \
  && echo "keycloak setup: disabled webauthn-register-passwordless" \
  || echo "keycloak setup: webauthn-register-passwordless not found (ok)"

# Partial realm imports often miss built-in required actions; register essentials.
register_required_action() {
  provider_id="$1"
  name="$2"
  if /opt/keycloak/bin/kcadm.sh get "authentication/required-actions/${provider_id}" \
    -r "${KC_REALM}" >/dev/null 2>&1; then
    return 0
  fi
  /opt/keycloak/bin/kcadm.sh create authentication/register-required-action \
    -r "${KC_REALM}" \
    -s "providerId=${provider_id}" \
    -s "name=${name}" >/dev/null 2>&1 \
    && echo "keycloak setup: registered required action ${provider_id}" \
    || echo "keycloak setup: could not register ${provider_id} (ok if already present)"
}

register_required_action "VERIFY_EMAIL" "Verify Email"
register_required_action "UPDATE_PASSWORD" "Update Password"
register_required_action "UPDATE_PROFILE" "Update Profile"
register_required_action "CONFIGURE_TOTP" "Configure OTP"
register_required_action "webauthn-register" "Webauthn Register"
register_required_action "UPDATE_EMAIL" "Update Email"
register_required_action "VERIFY_PROFILE" "Verify Profile"
register_required_action "delete_credential" "Delete Credential"
register_required_action "update_user_locale" "Update User Locale"

# VERIFY_EMAIL: enabled + default for new registrations.
/opt/keycloak/bin/kcadm.sh update "authentication/required-actions/VERIFY_EMAIL" \
  -r "${KC_REALM}" \
  -s enabled=true \
  -s defaultAction=true >/dev/null 2>&1 \
  && echo "keycloak setup: VERIFY_EMAIL enabled as default action" \
  || echo "keycloak setup: VERIFY_EMAIL update skipped"

# Password reset / AIA: enable UPDATE_PASSWORD and raise max auth age
# so change-password does not force username+password every time.
/opt/keycloak/bin/kcadm.sh update "authentication/required-actions/UPDATE_PASSWORD" \
  -r "${KC_REALM}" \
  -s enabled=true \
  -s defaultAction=false \
  -s 'config.max_auth_age=["2592000"]' >/dev/null 2>&1 \
  && echo "keycloak setup: UPDATE_PASSWORD enabled (max_auth_age=30d)" \
  || echo "keycloak setup: UPDATE_PASSWORD update skipped"

echo "keycloak setup: realm ${KC_REALM} themes + SMTP (${KC_SMTP_HOST}:${KC_SMTP_PORT}) applied"

# --- Email OTP 2FA (mesutpiskin email-authenticator SPI) ---
EMAIL_OTP_FLOW="${EMAIL_OTP_FLOW:-browser-email-otp}"
EMAIL_OTP_PROVIDER="${EMAIL_OTP_PROVIDER:-email-authenticator}"

# CSV helpers (Keycloak image has no awk).
csv_field() {
  # csv_field <n> — print nth 1-based comma field from stdin (simple, no quoted commas).
  n="$1"
  while IFS= read -r line; do
    i=1
    rest="$line"
    while [ "$i" -lt "$n" ]; do
      rest="${rest#*,}"
      i=$((i + 1))
    done
    printf '%s\n' "${rest%%,*}"
  done
}

flow_alias_exists() {
  /opt/keycloak/bin/kcadm.sh get authentication/flows -r "${KC_REALM}" \
    --fields alias --format csv --noquotes 2>/dev/null \
    | grep -qx "$1"
}

# Wait until the SPI is registered (custom image may still be warming up).
EMAIL_OTP_READY=0
for i in $(seq 1 30); do
  if /opt/keycloak/bin/kcadm.sh get authentication/authenticator-providers -r "${KC_REALM}" 2>/dev/null \
    | grep -q "\"id\" : \"${EMAIL_OTP_PROVIDER}\""; then
    EMAIL_OTP_READY=1
    break
  fi
  sleep 2
done

if [ "${EMAIL_OTP_READY}" != "1" ]; then
  echo "keycloak setup: ERROR ${EMAIL_OTP_PROVIDER} provider not found — rebuild Keycloak image with the SPI JAR"
  exit 1
fi

if flow_alias_exists "${EMAIL_OTP_FLOW}"; then
  echo "keycloak setup: flow ${EMAIL_OTP_FLOW} already exists"
else
  /opt/keycloak/bin/kcadm.sh create "authentication/flows/browser/copy" -r "${KC_REALM}" \
    -s "newName=${EMAIL_OTP_FLOW}"
  echo "keycloak setup: copied browser → ${EMAIL_OTP_FLOW}"
fi

# Forms sub-flow: displayName ends with " forms", authenticationFlow=true, field3=flowId.
FORMS_FLOW_ID="$(/opt/keycloak/bin/kcadm.sh get "authentication/flows/${EMAIL_OTP_FLOW}/executions" -r "${KC_REALM}" \
  --format csv --fields displayName,authenticationFlow,flowId --noquotes 2>/dev/null \
  | grep ',true,' | grep ' forms,' | head -n1 | csv_field 3)"

FORMS_ALIAS=""
if [ -n "${FORMS_FLOW_ID}" ]; then
  FORMS_ALIAS="$(/opt/keycloak/bin/kcadm.sh get authentication/flows -r "${KC_REALM}" \
    --format csv --fields id,alias --noquotes 2>/dev/null \
    | grep -F "${FORMS_FLOW_ID}," | head -n1 | csv_field 2)"
fi

if [ -z "${FORMS_ALIAS}" ]; then
  FORMS_ALIAS="${EMAIL_OTP_FLOW} forms"
fi

echo "keycloak setup: forms sub-flow alias=${FORMS_ALIAS}"

# kcadm path segments must be URL-encoded (alias can contain spaces).
FORMS_ALIAS_ENC="$(printf '%s' "${FORMS_ALIAS}" | sed 's/ /%20/g')"

HAS_EMAIL_OTP="$(/opt/keycloak/bin/kcadm.sh get "authentication/flows/${EMAIL_OTP_FLOW}/executions" -r "${KC_REALM}" \
  --format csv --fields providerId --noquotes 2>/dev/null | grep -c "^${EMAIL_OTP_PROVIDER}$" || true)"

if [ "${HAS_EMAIL_OTP}" = "0" ]; then
  /opt/keycloak/bin/kcadm.sh create "authentication/flows/${FORMS_ALIAS_ENC}/executions/execution" \
    -r "${KC_REALM}" \
    -b "{\"provider\":\"${EMAIL_OTP_PROVIDER}\"}"
  echo "keycloak setup: added ${EMAIL_OTP_PROVIDER} to ${FORMS_ALIAS}"
else
  echo "keycloak setup: ${EMAIL_OTP_PROVIDER} already in ${EMAIL_OTP_FLOW}"
fi

EXEC_ID="$(/opt/keycloak/bin/kcadm.sh get "authentication/flows/${EMAIL_OTP_FLOW}/executions" -r "${KC_REALM}" \
  --format csv --fields id,providerId --noquotes 2>/dev/null \
  | grep ",${EMAIL_OTP_PROVIDER}$" | head -n1 | csv_field 1)"

if [ -z "${EXEC_ID}" ]; then
  echo "keycloak setup: ERROR could not find ${EMAIL_OTP_PROVIDER} execution id"
  exit 1
fi

/opt/keycloak/bin/kcadm.sh update "authentication/flows/${EMAIL_OTP_FLOW}/executions" -r "${KC_REALM}" \
  -b "{\"id\":\"${EXEC_ID}\",\"requirement\":\"REQUIRED\"}"
echo "keycloak setup: ${EMAIL_OTP_PROVIDER} requirement=REQUIRED (id=${EXEC_ID})"

# New executions land at priority 0 (before Username Password). Move Email OTP below password form.
password_index() {
  /opt/keycloak/bin/kcadm.sh get "authentication/flows/${EMAIL_OTP_FLOW}/executions" -r "${KC_REALM}" \
    --format csv --fields providerId,index --noquotes 2>/dev/null \
    | grep "^auth-username-password-form," | head -n1 | csv_field 2
}

email_otp_index() {
  /opt/keycloak/bin/kcadm.sh get "authentication/flows/${EMAIL_OTP_FLOW}/executions" -r "${KC_REALM}" \
    --format csv --fields providerId,index --noquotes 2>/dev/null \
    | grep "^${EMAIL_OTP_PROVIDER}," | head -n1 | csv_field 2
}

i=0
while [ "$i" -lt 20 ]; do
  P_IDX="$(password_index)"
  E_IDX="$(email_otp_index)"
  if [ -n "${P_IDX}" ] && [ -n "${E_IDX}" ] && [ "${E_IDX}" -gt "${P_IDX}" ]; then
    echo "keycloak setup: Email OTP index=${E_IDX} after password index=${P_IDX}"
    break
  fi
  /opt/keycloak/bin/kcadm.sh create "authentication/executions/${EXEC_ID}/lower-priority" -r "${KC_REALM}" >/dev/null
  i=$((i + 1))
done

EXISTING_CFG="$(/opt/keycloak/bin/kcadm.sh get "authentication/flows/${EMAIL_OTP_FLOW}/executions" -r "${KC_REALM}" \
  --format csv --fields id,providerId,authenticationConfig --noquotes 2>/dev/null \
  | grep ",${EMAIL_OTP_PROVIDER}," | head -n1 | csv_field 3)"

CFG_JSON='{"alias":"voco-email-otp","config":{"skipSetup":"true","emailProviderType":"KEYCLOAK","length":"6","ttl":"300","simulationMode":"false","resendCooldown":"30","maxAttempts":"5","showMaskedEmailOnOtpForm":"true","enableFallback":"true"}}'

if [ -z "${EXISTING_CFG}" ] || [ "${EXISTING_CFG}" = "null" ]; then
  /opt/keycloak/bin/kcadm.sh create "authentication/executions/${EXEC_ID}/config" -r "${KC_REALM}" \
    -b "${CFG_JSON}"
  echo "keycloak setup: created email OTP config (skipSetup=true)"
else
  /opt/keycloak/bin/kcadm.sh update "authentication/config/${EXISTING_CFG}" -r "${KC_REALM}" \
    -b "${CFG_JSON}"
  echo "keycloak setup: updated email OTP config (skipSetup=true)"
fi

/opt/keycloak/bin/kcadm.sh update "realms/${KC_REALM}" -s "browserFlow=${EMAIL_OTP_FLOW}"
echo "keycloak setup: bound browserFlow=${EMAIL_OTP_FLOW}"

CLIENT_ID="$(/opt/keycloak/bin/kcadm.sh get clients -r "${KC_REALM}" -q clientId=voco-frontend --fields id --format csv --noquotes 2>/dev/null | tail -n1)"
SCOPE_ID="$(/opt/keycloak/bin/kcadm.sh get client-scopes -r "${KC_REALM}" -q name=account --fields id --format csv --noquotes 2>/dev/null | tail -n1)"

if [ -n "${CLIENT_ID}" ]; then
  /opt/keycloak/bin/kcadm.sh update "clients/${CLIENT_ID}" -r "${KC_REALM}" \
    -s 'rootUrl=http://localhost:5173' \
    -s 'baseUrl=/' >/dev/null 2>&1 \
    && echo "keycloak setup: voco-frontend rootUrl=http://localhost:5173" \
    || echo "keycloak setup: voco-frontend rootUrl update skipped"
fi

if [ -n "${CLIENT_ID}" ] && [ -n "${SCOPE_ID}" ]; then
  /opt/keycloak/bin/kcadm.sh create "clients/${CLIENT_ID}/default-client-scopes/${SCOPE_ID}" -r "${KC_REALM}" >/dev/null 2>&1 \
    && echo "keycloak setup: attached account scope to voco-frontend" \
    || echo "keycloak setup: account scope already attached (ok)"
fi

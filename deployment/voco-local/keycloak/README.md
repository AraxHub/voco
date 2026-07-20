# Keycloak theme `voco`

Dark UI aligned with the VOCO frontend (gradient background, glass card, purple/blue accents).

## Apply locally

```bash
docker compose -f deployment/voco-local/docker-compose.yml -p voco-local build keycloak
docker compose -f deployment/voco-local/docker-compose.yml -p voco-local up -d --force-recreate keycloak keycloak-realm-setup mailpit
```

Custom image (`deployment/voco-local/keycloak/Dockerfile`) bundles the Email OTP SPI JAR on top of Keycloak 26.6.3.

Service `keycloak-realm-setup` on every `docker compose up` sets:

- Login / Account themes (`voco`)
- SMTP + `verifyEmail`
- Browser flow **`browser-email-otp`** (Email OTP REQUIRED after password, `skipSetup=true`)

Passkeys are disabled. Realm import alone does not update an existing realm.

If theme still looks default, hard-refresh the login page (Ctrl+Shift+R) or open a private window.

## Email (SMTP)

Local stack includes **Mailpit**:

| Port | Role |
|------|------|
| `1025` | SMTP (Keycloak → Mailpit) |
| `8025` | Web UI — open [http://localhost:8025](http://localhost:8025) |

Defaults (via `keycloak-realm-setup` env):

- host `mailpit`, port `1025`, no auth
- from `noreply@voco.local`
- `verifyEmail=true` + required action VERIFY_EMAIL

### Test locally (transactional)

1. Register a user on [http://localhost:5173](http://localhost:5173) (Keycloak registration).
2. Open Mailpit UI → verification email.
3. On login page: forgot password → reset link appears in Mailpit.

### Test Email OTP 2FA

1. Log in with email + password.
2. Keycloak shows the OTP form and sends a 6-digit code (Mailpit).
3. Enter the code → redirect back to VOCO.

SPI: [mesutpiskin/keycloak-2fa-email-authenticator](https://github.com/mesutpiskin/keycloak-2fa-email-authenticator) `v26.4.3` / `KC26.6.3`.

### Resend (real inbox)

In `deployment/voco-local/.env` (see `.env.example`):

```bash
KC_SMTP_HOST=smtp.resend.com
KC_SMTP_PORT=587
KC_SMTP_FROM=noreply@your-verified-domain.com
KC_SMTP_FROM_DISPLAY_NAME=VOCO
KC_SMTP_AUTH=true
KC_SMTP_SSL=false
KC_SMTP_STARTTLS=true
KC_SMTP_USER=resend
KC_SMTP_PASSWORD=re_xxxxxxxx
```

Then re-run setup:

```bash
docker compose -f deployment/voco-local/docker-compose.yml -p voco-local run --rm keycloak-realm-setup
```

`from` must be a Resend-verified domain/sender. OTP and verify/reset emails use the same SMTP.

## Edit theme

Files under `themes/voco/`:

- `login/resources/css/voco.css` — login / register / reset password / OTP form
- `account/resources/css/voco.css` — account console
- `login/messages/messages_ru.properties` — Russian strings

Dev mode disables theme caching in `docker-compose.yml`; refresh the browser after CSS changes.

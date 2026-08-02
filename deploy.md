# Voco — деплой и CI/CD (кратко)

Полное описание фич, Keycloak session, swagger и журнал работ: локально в **`docs/`** (gitignored) → [`docs/README.md`](docs/README.md), [`docs/cicd-deploy.md`](docs/cicd-deploy.md).

## Pipeline

`.github/workflows/pipeline.yml`: PR → тесты; `main` → GHCR → SSH deploy backend/frontend.

Образы: `ghcr.io/araxhub/voco-{backend,frontend}:sha-<7>` / `:latest`.

## GitHub Secrets

`VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY` (+ опционально `VPS_SSH_PORT`).

## VPS secrets (`deployment/k3s/secrets.env`)

Обязательные: LiveKit keys, `VOCO_PG_URL`, Keycloak SMTP/DB/admin, GHCR pull,  
**также:** `VOCO_KEYCLOAK_ADMIN_CLIENT_SECRET`, `VOCO_WEBPUSH_VAPID_PUBLIC`, `VOCO_WEBPUSH_VAPID_PRIVATE`.

```bash
./deployment/k3s/create-secrets.sh
./deployment/k3s/create-ghcr-pull-secret.sh
./deployment/k3s/apply.sh
```

После деплоя: Keycloak SSO Session Idle = 30 days; admin client `view-users`; VAPID keys.

## Fallback

```bash
./deployment/k3s/build-images.sh && ./deployment/k3s/apply.sh
```

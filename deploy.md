# Voco — деплой и CI/CD

## Pipeline (GitHub Actions)

Один workflow **Pipeline** (`.github/workflows/pipeline.yml`):

```
Test API ──┐
Test Frontend ──┴─→ Build & push images ──→ Deploy to VPS
```

- **PR** — только тесты.
- **push в `main`** — тесты → GHCR → деплой backend + frontend.

Образы:

- `ghcr.io/araxhub/voco-backend:sha-<7>` / `:latest`
- `ghcr.io/araxhub/voco-frontend:sha-<7>` / `:latest`

**Keycloak** и **LiveKit** в CI не входят — как раньше, образы/манифесты на VPS вручную при необходимости.

### Secrets в GitHub (репо voco)

Settings → Secrets and variables → Actions — те же, что у roadmap (можно тот же ключ):

| Secret | Значение |
|--------|----------|
| `VPS_HOST` | `155.212.190.163` |
| `VPS_USER` | `root` |
| `VPS_SSH_KEY` | приватный ключ (`roadmap_deploy` или отдельный) |
| `VPS_SSH_PORT` | опционально |

### Один раз на VPS

В `~/voco/deployment/k3s/secrets.env` (или как у тебя лежит репо):

```bash
GHCR_USERNAME=AraxHub
GHCR_TOKEN=ghp_...   # тот же PAT read:packages, что для roadmap
```

```bash
cd ~/voco   # путь к клону на сервере
git pull
./deployment/k3s/create-ghcr-pull-secret.sh
./deployment/k3s/apply.sh
```

Дальше push в `main` сам обновляет backend/frontend.

---

## Ручной fallback

```bash
./deployment/k3s/build-images.sh
./deployment/k3s/apply.sh
# или только рестарт текущих тегов:
./deployment/k3s/restart.sh
```

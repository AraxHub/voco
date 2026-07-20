# VOCO

Приложение “видеокомнаты по ссылке”:
- **frontend**: React + Vite (UI)
- **backend**: Go (Gin) — создаёт комнаты и выдаёт JWT для LiveKit
- **livekit**: LiveKit Server — WebRTC (сигналинг по TCP, медиа по UDP)

Ниже — практическая инструкция, как **развернуть на VPS** через **k3s**. Я специально подписываю, **зачем** каждая команда и **что она делает**, чтобы по дороге можно было учиться.

---

## Архитектура сети (коротко)

- **HTTP (TCP/80)**: сайт и API приходят через Ingress (`/` и `/api`).
- **LiveKit signaling (TCP/7880)**: WebSocket `ws://<ip>:7880` — соединение браузера с LiveKit.
- **LiveKit media (UDP/62000-62020)**: WebRTC трафик (аудио/видео).

Если **UDP диапазон закрыт**, то чаще всего выглядит так: “в комнату заходит, но нет звука/видео или всё рвётся”.

---

## Быстрый деплой на VPS (k3s, IP-only, без TLS)

### 0) Что нужно заранее

- VPS на Ubuntu/Debian с публичным IP.
- Открытые порты на фаерволле VPS/провайдера:
  - **TCP 80** (Ingress)
  - **TCP 7880-7881** (LiveKit signaling/rtc-tcp)
  - **UDP 62000-62020** (LiveKit media)

### 1) Зайти на VPS и обновить пакеты

```bash
sudo apt update
```

- **зачем**: обновляет список пакетов (иначе ставить софт будем по старым индексам).

```bash
sudo apt upgrade -y
```

- **зачем**: обновляет систему; часто закрывает security-дыры перед установкой k3s/docker.

### 2) Поставить базовые утилиты и Docker

```bash
sudo apt install -y ca-certificates curl git docker.io
```

- **ca-certificates**: чтобы `curl` и контейнеры нормально ходили по HTTPS.
- **curl**: скачивать установщик k3s.
- **git**: клонировать репозиторий.
- **docker.io**: собирать контейнеры на VPS (на первом этапе мы так и делаем).

```bash
sudo usermod -aG docker $USER
```

- **зачем**: добавляет твоего пользователя в группу `docker`, чтобы можно было запускать `docker ...` без `sudo`.
- **важно**: после этого надо **перелогиниться в ssh**, иначе группа не применится.

### 3) Клонировать репозиторий на VPS

```bash
git clone <URL_ТВОЕГО_РЕПО>
cd Voco
```

- **зачем**: берём манифесты/скрипты из репо, чтобы деплой был воспроизводимым.

### 4) Установить k3s

```bash
sudo ./deployment/k3s/bootstrap-k3s.sh
```

Скрипт делает 3 ключевых вещи:
- ставит **k3s** (лёгкий Kubernetes),
- отключает **Traefik** и **ServiceLB** (мы ставим ingress-nginx),
- расширяет `service-node-port-range=1-65535`, чтобы NodePort мог быть **7880/7881 и 62000-62020** (иначе Kubernetes не разрешит такие nodePort).

### 5) Подключить kubectl к кластеру

```bash
mkdir -p ~/.kube
sudo cat /etc/rancher/k3s/k3s.yaml > ~/.kube/config
sudo chown "$(id -u)":"$(id -g)" ~/.kube/config
kubectl get nodes
```

- **зачем**: `kubectl` — главный клиент для управления Kubernetes.
- k3s хранит конфиг в `/etc/rancher/k3s/k3s.yaml`, мы копируем его в `~/.kube/config`.

### 6) Поставить ingress-nginx

```bash
./deployment/k3s/ingress-install.sh
kubectl -n ingress-nginx get pods
```

- **зачем**: Ingress сам по себе не работает без контроллера. ingress-nginx будет принимать HTTP на 80 и проксировать на frontend/backend сервисы.\n+  Важно: у нас в k3s отключён `servicelb`, поэтому скрипт дополнительно переводит сервис ingress-nginx в `NodePort` и закрепляет `80/443` на ноде.

### 7) Создать секреты (LiveKit API key/secret для backend)

Мы **не коммитим** реальные секреты.

```bash
cp deployment/k3s/secrets.env.example deployment/k3s/secrets.env
nano deployment/k3s/secrets.env
./deployment/k3s/create-secrets.sh
```

- **зачем**: backend подписывает JWT ключом/секретом; LiveKit проверяет JWT по тем же ключам.
- **create-secrets.sh** читает `secrets.env` и создаёт/обновляет Kubernetes Secret `voco-secrets`.

Важно: значения должны **совпадать** с конфигом LiveKit `keys:` (см. `deployment/k8s/overlays/k3s/livekit-configmap.yaml`).

Самый простой старт: оставь значения по умолчанию **`devkey/devsecret`** (они уже прописаны в livekit-configmap).

Если хочешь сгенерить свои:

```bash
KEY="voco_$(openssl rand -hex 6)"
SECRET="$(openssl rand -hex 32)"
echo "VOCO_LIVEKIT_API_KEY=$KEY"
echo "VOCO_LIVEKIT_API_SECRET=$SECRET"
```

После генерации нужно:\n- записать их в `deployment/k3s/secrets.env`\n- обновить `keys:` в `deployment/k8s/overlays/k3s/livekit-configmap.yaml`\n- применить изменения (`./deployment/k3s/create-secrets.sh` и `./deployment/k3s/apply.sh`).

### 8) Собрать образы прямо на VPS

```bash
./deployment/k3s/build-images.sh
```

- **зачем**: на “первом этапе” мы не используем registry, поэтому строим `voco-backend:local` и `voco-frontend:local` локально на сервере.

### 9) Применить Kubernetes манифесты (kustomize overlay для k3s)

```bash
./deployment/k3s/apply.sh
kubectl -n voco get pods,svc,ing
```

- **зачем**: `kubectl apply -k` применяет overlay `deployment/k8s/overlays/k3s`.

### 10) Проверки

```bash
curl -I http://155.212.190.163/
curl -i http://155.212.190.163/api/v1/rooms
kubectl -n voco logs -f deploy/voco-backend
kubectl -n voco logs -f deploy/livekit
```

- **зачем**: проверить, что Ingress отдаёт фронт, API отвечает, а LiveKit не падает.

---

## Что лежит в репозитории для k3s

- `deployment/k8s/overlays/k3s/`: kustomize overlay для VPS/k3s (Ingress + backend/frontend + LiveKit NodePort).
- `deployment/k3s/`: скрипты установки/деплоя на VPS.

---

## Этап 2 (нормально): registry + CI

Идея: собирать контейнеры в CI и пушить в registry (GHCR/DockerHub), а k3s будет тянуть их по тегу.
Это делает деплой повторяемым и убирает сборку на сервере.

Смотри подсказки в `deployment/k3s/README.md` (про `images:` в kustomize и `imagePullSecret` для приватного registry).


---
title: "ADR-0001: Архитектура взаимодействия Frontend ↔ Backend ↔ LiveKit"
status: accepted
date: 2026-04-27
---

## Контекст
Нужно поддержать MVP видеокомнат «по ссылке»:
- фронт создаёт/открывает комнату по URL `/room/:roomId`;
- бэк хранит метаданные комнаты (TTL/валидность) и **выпускает LiveKit JWT**;
- медиа/чат идут через LiveKit, бэк **не проксирует** WebRTC трафик.

## Решение
Разделяем ответственность:
- **Frontend (React/Vite)**: UI, запросы к бэку (`/api/v1/*`), подключение к LiveKit по `serverUrl` + `token`.
- **Backend (Go/Gin)**: REST API для создания комнаты и выдачи токена; хранение room state (in-memory store) и генерация JWT (LiveKit API key/secret).
- **LiveKit Server**: signaling WebSocket, SFU (WebRTC), чат/данные по DataChannel.

### API (Backend)
- `POST /api/v1/rooms` → `{ roomId, joinUrl }`
- `POST /api/v1/rooms/:roomId/token` `{ name }` → `{ token, livekitUrl }`

Токен выпускается с `VideoGrant`:
- `RoomJoin: true`
- `Room: <roomId>`

## Диаграмма взаимодействия (sequence)

```mermaid
sequenceDiagram
  autonumber
  actor User
  participant FE as Frontend(React)
  participant BE as Backend(Go/Gin)
  participant LK as LiveKitServer

  User->>FE: Открывает главную /
  User->>FE: Нажимает "Создать комнату"
  FE->>BE: POST /api/v1/rooms
  BE-->>FE: {roomId, joinUrl}
  FE-->>User: Навигация на /room/:roomId

  User->>FE: Вводит имя, включает mic/cam
  FE->>BE: POST /api/v1/rooms/:roomId/token {name}
  BE->>BE: Проверка комнаты + продление TTL
  BE->>BE: Генерация LiveKit JWT (APIKey/Secret)
  BE-->>FE: {token, livekitUrl}

  FE->>LK: Connect(serverUrl=livekitUrl, token)
  LK-->>FE: Подтверждение join + состояние комнаты

  FE<->>LK: WebRTC media (audio/video/screenShare)
  FE<->>LK: Chat/данные (DataChannel)
```

## Диаграмма компонентов (C4-ish, упрощённо)

```mermaid
flowchart LR
  User[UserBrowser]
  FE[Frontend(React/Vite)]
  BE[Backend(Go/Gin)]
  Store[RoomStore(in-memory_TTL)]
  LK[LiveKitServer(SFU)]

  User --> FE
  FE -->|"REST: /api/v1/rooms"| BE
  FE -->|"REST: /api/v1/rooms/:id/token"| BE
  BE --> Store
  BE -->|"JWT_sign(APIKey/Secret)" LK
  FE -->|"WS+WebRTC(signaling/media)" LK
```

## Последствия
- **Плюсы**:
  - простой бэк: нет обработки медиа-трафика;
  - масштабирование: LiveKit масштабируется отдельно;
  - безопасность: фронт никогда не видит API secret, только JWT.
- **Минусы/ограничения**:
  - доступность `livekitUrl`: значение, возвращаемое бэком, должно быть достижимо браузером (в docker-compose это часто `ws://localhost:7880` с хоста).
  - state комнаты (TTL) — в памяти процесса бэка (при рестарте пропадает).

## Связанные реализации в репозитории
- Backend routes: `internal/api/http/controllers/rooms/controller.go`
- Token issuance: `internal/usecase/rooms/usecase.go`
- Frontend API calls: `frontend/src/lib/api.ts`


FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/voco ./cmd/voco


FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/voco /app/voco
COPY deployment/voco-local/.env /app/.env

EXPOSE 8080

ENV VOCO_DEV_MODE=true

CMD ["/app/voco"]
package app

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	httpapi "voco/internal/api/http"
	roomscontroller "voco/internal/api/http/controllers/rooms"
	"voco/internal/pkg/logger"
	"voco/internal/repository/cache/inmemory"
	roomsuc "voco/internal/usecase/rooms"
)

type App struct {
	cfg Config
}

func NewApp(cfg Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Run() error {
	log := logger.New(a.cfg.Log)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store := inmemory.NewCache(a.cfg.Cache)
	store.CleanUp(ctx, time.Hour)

	roomsUseCase := roomsuc.New(store, a.cfg.LiveKit)
	roomsCtrl := roomscontroller.New(roomsUseCase, a.cfg.Server.BaseUrl)

	server := httpapi.NewServer(a.cfg.Server, log)
	server.AddController(roomsCtrl)

	if err := server.Start(ctx); err != nil {
		log.Error("http server stopped with error", "error", err)
		return err
	}

	return nil
}

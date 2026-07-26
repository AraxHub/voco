package app

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"voco/internal/adapters/memory"
	pgadapter "voco/internal/adapters/postgres"
	httpapi "voco/internal/api/http"
	roomscontroller "voco/internal/api/http/controllers/rooms"
	"voco/internal/api/http/middlewares"
	"voco/internal/pkg/auth"
	"voco/internal/pkg/logger"
	"voco/internal/pkg/migrate"
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

	// До коннекта приложения: DSN с search_path=voco ещё невалиден, пока нет schema.
	if err := migrate.UpEmbedded(a.cfg.Pg.MigrateDSN()); err != nil {
		log.Error("migrate failed", "error", err)
		return err
	}
	log.Info("migrate ok")

	db, err := pgadapter.New(ctx, a.cfg.Pg)
	if err != nil {
		log.Error("postgres connect failed", "error", err)
		return err
	}
	defer db.Close()

	cache := memory.NewCache(a.cfg.Cache)
	cache.CleanUp(ctx, time.Hour)

	authSvc, err := auth.New(ctx, a.cfg.Keycloak)
	if err != nil {
		log.Error("auth init failed", "error", err)
		return err
	}
	if authSvc.Enabled() {
		log.Info("auth enabled", "keycloak", true)
	}

	roomsUseCase := roomsuc.New(cache, a.cfg.LiveKit)
	roomsCtrl := roomscontroller.New(roomsUseCase, a.cfg.Server.BaseUrl)

	server := httpapi.NewServer(a.cfg.Server, log, middlewares.NewMW(authSvc))
	server.AddController(roomsCtrl)

	if err := server.Start(ctx); err != nil {
		log.Error("http server stopped with error", "error", err)
		return err
	}

	return nil
}

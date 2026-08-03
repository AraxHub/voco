package app

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	blobpg "voco/internal/adapters/blob/postgres"
	kcdir "voco/internal/adapters/keycloak"
	pgadapter "voco/internal/adapters/postgres"
	webpushadapter "voco/internal/adapters/webpush"
	httpapi "voco/internal/api/http"
	blobsctrl "voco/internal/api/http/controllers/blobs"
	calendarctrl "voco/internal/api/http/controllers/calendar"
	chatctrl "voco/internal/api/http/controllers/chat"
	pushctrl "voco/internal/api/http/controllers/push"
	roomscontroller "voco/internal/api/http/controllers/rooms"
	usersctrl "voco/internal/api/http/controllers/users"
	"voco/internal/api/http/middlewares"
	"voco/internal/api/ws"
	"voco/internal/domain"
	"voco/internal/pkg/auth"
	"voco/internal/pkg/logger"
	"voco/internal/pkg/migrate"
	pgrepo "voco/internal/repository/postgres"
	calendaruc "voco/internal/usecase/calendar"
	chatuc "voco/internal/usecase/chat"
	"voco/internal/usecase/notify"
	roomsuc "voco/internal/usecase/rooms"
	usersuc "voco/internal/usecase/users"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "voco/internal/swaggerdocs"
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

	authSvc, err := auth.New(ctx, a.cfg.Keycloak)
	if err != nil {
		log.Error("auth init failed", "error", err)
		return err
	}
	if authSvc.Enabled() {
		log.Info("auth enabled", "keycloak", true)
	}

	userRepo := pgrepo.NewUserRepo(db)
	dir := kcdir.NewDirectory(a.cfg.KeycloakAdmin)
	usersUC := usersuc.New(userRepo, dir)
	usersUC.StartSyncLoop(ctx, a.cfg.Features.UserSyncInterval)

	blobStore := blobpg.New(db)
	roomRepo := pgrepo.NewRoomRepo(db)
	roomsUC := roomsuc.New(roomRepo, a.cfg.LiveKit)

	hub := ws.NewHub(authSvc, func(sub string) (domain.UserID, error) {
		u, ok, err := userRepo.GetByKeycloakSub(context.Background(), sub)
		if err != nil {
			return domain.UserID{}, err
		}
		if !ok {
			return domain.UserID{}, domain.ErrUserNotFound
		}
		return u.ID, nil
	})

	chatRepo := pgrepo.NewChatRepo(db)
	chatUC := chatuc.New(chatRepo, blobStore, roomsUC, userRepo, hub, chatuc.Config{
		MaxImageBytes: a.cfg.Features.MaxImageBytes,
		MaxFileBytes:  a.cfg.Features.MaxFileBytes,
	})

	calRepo := pgrepo.NewCalendarRepo(db)
	calUC := calendaruc.New(calRepo, roomsUC, hub)

	pushSender := webpushadapter.NewSender(db, a.cfg.WebPush)
	notify.NewWorker(calRepo, hub, pushSender, log).Start(ctx, 30*time.Second)

	roomsCtrl := roomscontroller.New(roomsUC, a.cfg.Server.BaseUrl, usersUC)
	maxImg := a.cfg.Features.MaxImageBytes
	usersCtrl := usersctrl.New(usersUC, blobStore, a.cfg.Server.BaseUrl, maxImg)
	blobsCtrl := blobsctrl.New(blobStore, usersUC, a.cfg.Server.BaseUrl, maxImg)
	chatCtrl := chatctrl.New(chatUC, usersUC, a.cfg.Server.BaseUrl)
	calCtrl := calendarctrl.New(calUC, usersUC)
	pushCtrl := pushctrl.New(pgrepo.NewPushRepo(db), usersUC, a.cfg.WebPush.VAPIDPublic)

	serverCfg := a.cfg.Server
	serverCfg.WriteTimeout = 0 // WebSocket

	server := httpapi.NewServer(serverCfg, log, middlewares.NewMW(authSvc))
	server.AddController(roomsCtrl, usersCtrl, blobsCtrl, chatCtrl, calCtrl, pushCtrl)
	server.AddRoutes(func(r *gin.Engine) {
		r.GET("/api/v1/ws", hub.Handle)
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	})

	if err := server.Start(ctx); err != nil {
		log.Error("http server stopped with error", "error", err)
		return err
	}
	return nil
}

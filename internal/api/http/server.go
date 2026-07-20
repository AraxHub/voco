package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"voco/internal/api/http/middlewares"
)

type Controller interface {
	RegisterRoutes(r *gin.Engine, mw middlewares.MW)
}

type Server struct {
	cfg         ServerConfig
	mw          middlewares.MW
	server      *http.Server
	log         *slog.Logger
	controllers []Controller
	extraRoutes []func(*gin.Engine)
}

type ServerConfig struct {
	Host           string        `envconfig:"HOST" default:"0.0.0.0"`
	Port           string        `envconfig:"PORT" default:"8080"`
	ReadTimeout    time.Duration `envconfig:"READ_TIMEOUT" default:"10s"`
	WriteTimeout   time.Duration `envconfig:"WRITE_TIMEOUT" default:"10s"`
	BaseUrl        string        `envconfig:"BASE_URL" default:"http://localhost:8080"`
	AllowedOrigins []string      `envconfig:"CORS_ALLOWED_ORIGINS" default:"http://localhost:5173,http://127.0.0.1:5173,http://localhost:3000,http://127.0.0.1:3000"`
}

func NewServer(cfg ServerConfig, log *slog.Logger, mw middlewares.MW) *Server {
	return &Server{cfg: cfg, log: log, mw: mw}
}

func (s *Server) AddController(c ...Controller) {
	s.controllers = append(s.controllers, c...)
}

func (s *Server) AddRoutes(fn func(*gin.Engine)) {
	if fn == nil {
		return
	}
	s.extraRoutes = append(s.extraRoutes, fn)
}

// Start поднимает роутер, запускает сервер и блокируется до отмены ctx, затем делает graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.AccessLog(s.log))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     s.cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300 * time.Second,
	}))

	for _, c := range s.controllers {
		c.RegisterRoutes(r, s.mw)
	}
	for _, fn := range s.extraRoutes {
		fn(r)
	}

	s.server = &http.Server{
		Addr:         s.cfg.Host + ":" + s.cfg.Port,
		Handler:      r,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info("starting http server", "host", s.cfg.Host, "port", s.cfg.Port)
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}

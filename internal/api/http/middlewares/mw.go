package middlewares

import (
	"voco/internal/pkg/auth"

	"github.com/gin-gonic/gin"
)

// MW — набор HTTP-middleware для подключения к отдельным эндпоинтам.
type MW struct {
	Auth         gin.HandlerFunc
	AuthOptional gin.HandlerFunc
}

func NewMW(authSvc *auth.Service) MW {
	return MW{
		Auth:         Auth(authSvc),
		AuthOptional: AuthOptional(authSvc),
	}
}

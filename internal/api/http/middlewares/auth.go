package middlewares

import (
	"net/http"
	"strings"

	"voco/internal/pkg/auth"

	"github.com/gin-gonic/gin"
)

type contextKey string

const UserContextKey contextKey = "authUser"

func Auth(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil || !svc.Enabled() {
			c.Next()
			return
		}

		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}

		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		user, err := svc.Verify(c.Request.Context(), raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":  "invalid token",
				"detail": err.Error(),
			})
			return
		}

		c.Set(string(UserContextKey), user)
		c.Next()
	}
}

func AuthOptional(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil || !svc.Enabled() {
			c.Next()
			return
		}
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			c.Next()
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		user, err := svc.Verify(c.Request.Context(), raw)
		if err != nil {
			c.Next()
			return
		}
		c.Set(string(UserContextKey), user)
		c.Next()
	}
}

func UserFromContext(c *gin.Context) (auth.User, bool) {
	v, ok := c.Get(string(UserContextKey))
	if !ok {
		return auth.User{}, false
	}
	user, ok := v.(auth.User)
	return user, ok
}

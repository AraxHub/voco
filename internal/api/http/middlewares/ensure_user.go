package middlewares

import (
	"net/http"

	"voco/internal/domain"
	"voco/internal/pkg/auth"
	"voco/internal/usecase/users"

	"github.com/gin-gonic/gin"
)

const DomainUserKey = "domainUser"

type EnsureUserFunc func(c *gin.Context, authUser auth.User) (domain.User, error)

func EnsureUser(uc *users.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		au, ok := UserFromContext(c)
		if !ok {
			// auth disabled — synthetic local user for dev
			c.Next()
			return
		}
		u, err := uc.EnsureFromAuth(c.Request.Context(), au.Sub, au.Email, au.Name)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Set(DomainUserKey, u)
		c.Next()
	}
}

func DomainUser(c *gin.Context) (domain.User, bool) {
	v, ok := c.Get(DomainUserKey)
	if !ok {
		return domain.User{}, false
	}
	u, ok := v.(domain.User)
	return u, ok
}

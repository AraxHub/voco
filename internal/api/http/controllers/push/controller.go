package pushctrl

import (
	"net/http"

	"voco/internal/api/http/middlewares"
	pgrepo "voco/internal/repository/postgres"
	"voco/internal/usecase/users"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	repo  *pgrepo.PushRepo
	users *users.Usecase
	vapid string
}

func New(repo *pgrepo.PushRepo, usersUC *users.Usecase, vapidPublic string) *Controller {
	return &Controller{repo: repo, users: usersUC, vapid: vapidPublic}
}

func (c *Controller) RegisterRoutes(r *gin.Engine, mw middlewares.MW) {
	api := r.Group("/api/v1", mw.Auth, middlewares.EnsureUser(c.users))
	api.GET("/push/vapidPublicKey", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"publicKey": c.vapid})
	})
	api.GET("/push/settings", c.getSettings)
	api.PUT("/push/settings", c.putSettings)
	api.POST("/push/subscribe", c.subscribe)
	api.DELETE("/push/subscribe", c.unsubscribe)
}

func (c *Controller) getSettings(ctx *gin.Context) {
	u, ok := middlewares.DomainUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	en, _ := c.repo.GetPushEnabled(ctx.Request.Context(), u.ID)
	ctx.JSON(http.StatusOK, gin.H{"pushEnabled": en})
}

func (c *Controller) putSettings(ctx *gin.Context) {
	u, ok := middlewares.DomainUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		PushEnabled bool `json:"pushEnabled"`
	}
	_ = ctx.ShouldBindJSON(&req)
	_ = c.repo.SetPushEnabled(ctx.Request.Context(), u.ID, req.PushEnabled)
	ctx.JSON(http.StatusOK, gin.H{"pushEnabled": req.PushEnabled})
}

func (c *Controller) subscribe(ctx *gin.Context) {
	u, ok := middlewares.DomainUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		Endpoint string `json:"endpoint"`
		Keys     struct {
			P256dh string `json:"p256dh"`
			Auth   string `json:"auth"`
		} `json:"keys"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = c.repo.UpsertSubscription(ctx.Request.Context(), u.ID, req.Endpoint, req.Keys.P256dh, req.Keys.Auth)
	_ = c.repo.SetPushEnabled(ctx.Request.Context(), u.ID, true)
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) unsubscribe(ctx *gin.Context) {
	u, ok := middlewares.DomainUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		Endpoint string `json:"endpoint"`
	}
	_ = ctx.ShouldBindJSON(&req)
	_ = c.repo.DeleteSubscription(ctx.Request.Context(), u.ID, req.Endpoint)
	ctx.Status(http.StatusNoContent)
}

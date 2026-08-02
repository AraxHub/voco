package usersctrl

import (
	"net/http"

	"voco/internal/api/http/middlewares"
	"voco/internal/usecase/users"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	uc *users.Usecase
}

func New(uc *users.Usecase) *Controller {
	return &Controller{uc: uc}
}

func (c *Controller) RegisterRoutes(r *gin.Engine, mw middlewares.MW) {
	api := r.Group("/api/v1", mw.Auth, middlewares.EnsureUser(c.uc))
	api.GET("/users/me", c.me)
	api.PATCH("/users/me", c.updateMe)
	api.GET("/users/search", c.search)
}

type userDTO struct {
	ID          string `json:"id"`
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	LastSeenAt  string `json:"lastSeenAt"`
}

func (c *Controller) me(ctx *gin.Context) {
	u, ok := middlewares.DomainUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx.JSON(http.StatusOK, userDTO{
		ID: u.ID.String(), Nickname: u.Nickname, Email: u.Email, DisplayName: u.DisplayName,
		LastSeenAt: u.LastSeenAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (c *Controller) updateMe(ctx *gin.Context) {
	u, ok := middlewares.DomainUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		Nickname    string `json:"nickname"`
		DisplayName string `json:"displayName"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := c.uc.UpdateMe(ctx.Request.Context(), u.ID, req.Nickname, req.DisplayName)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, userDTO{
		ID: updated.ID.String(), Nickname: updated.Nickname, Email: updated.Email, DisplayName: updated.DisplayName,
		LastSeenAt: updated.LastSeenAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (c *Controller) search(ctx *gin.Context) {
	list, err := c.uc.Search(ctx.Request.Context(), ctx.Query("q"), 20)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]userDTO, 0, len(list))
	for _, u := range list {
		out = append(out, userDTO{
			ID: u.ID.String(), Nickname: u.Nickname, Email: u.Email, DisplayName: u.DisplayName,
			LastSeenAt: u.LastSeenAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	ctx.JSON(http.StatusOK, out)
}

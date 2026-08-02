package rooms

import (
	"context"
	"net/http"
	"strings"

	"voco/internal/api/http/middlewares"
	"voco/internal/domain"
	"voco/internal/usecase/users"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomUC interface {
	CreateRoom(ctx context.Context, title string, owner *domain.UserID) (domain.Room, error)
	IssueToken(ctx context.Context, roomID domain.RoomID, participantName string, identity string) (string, string, error)
}

type Controller struct {
	uc      RoomUC
	users   *users.Usecase
	BaseUrl string
}

func New(uc RoomUC, url string, usersUC *users.Usecase) *Controller {
	return &Controller{uc: uc, BaseUrl: url, users: usersUC}
}

func (c *Controller) RegisterRoutes(r *gin.Engine, mw middlewares.MW) {
	api := r.Group("/api/v1")

	api.POST("/rooms/:roomId/token", mw.AuthOptional, c.maybeEnsureUser, c.issueToken)

	create := []gin.HandlerFunc{mw.Auth}
	if c.users != nil {
		create = append(create, middlewares.EnsureUser(c.users))
	}
	api.POST("/rooms", append(create, c.createRoom)...)
}

func (c *Controller) maybeEnsureUser(ctx *gin.Context) {
	if c.users == nil {
		ctx.Next()
		return
	}
	if _, ok := middlewares.UserFromContext(ctx); !ok {
		ctx.Next()
		return
	}
	middlewares.EnsureUser(c.users)(ctx)
}

func (c *Controller) createRoom(ctx *gin.Context) {
	var req CreateRoomRequest

	if err := ctx.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var owner *domain.UserID
	if u, ok := middlewares.DomainUser(ctx); ok {
		owner = &u.ID
	}

	room, err := c.uc.CreateRoom(ctx.Request.Context(), req.Title, owner)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	base := strings.TrimRight(c.BaseUrl, "/")
	joinURL := base + "/room/" + room.ID.String()

	ctx.JSON(http.StatusOK, CreateRoomResponse{
		RoomID:  room.ID.String(),
		JoinURL: joinURL,
	})
}

func (c *Controller) issueToken(ctx *gin.Context) {
	roomIDParam := ctx.Param("roomId")
	parsed, err := uuid.Parse(roomIDParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomId"})
		return
	}
	roomID := domain.RoomID(parsed)

	var req IssueTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	identity := ""
	if u, ok := middlewares.DomainUser(ctx); ok {
		identity = u.ID.String()
	}

	token, liveKitUrl, err := c.uc.IssueToken(ctx.Request.Context(), roomID, req.Name, identity)
	if err != nil {
		if err == domain.ErrRoomNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, IssueTokenResponse{
		Token:      token,
		LiveKitURL: liveKitUrl,
	})
}

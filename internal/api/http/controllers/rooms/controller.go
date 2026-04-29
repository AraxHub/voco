package rooms

import (
	"context"
	"net/http"
	"strings"

	"voco/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomUC interface {
	CreateRoom(ctx context.Context, title string) (domain.Room, error)
	IssueToken(ctx context.Context, roomID domain.RoomID, participantName string) (string, string, error)
}

type Controller struct {
	uc      RoomUC
	BaseUrl string
}

func New(uc RoomUC, url string) *Controller {
	return &Controller{
		uc:      uc,
		BaseUrl: url,
	}
}

func (c *Controller) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	api.POST("/rooms", c.createRoom)
	api.POST("/rooms/:roomId/token", c.issueToken)
}

func (c *Controller) createRoom(ctx *gin.Context) {
	var req CreateRoomRequest

	if err := ctx.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := c.uc.CreateRoom(ctx.Request.Context(), req.Title)
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

	token, liveKitUrl, err := c.uc.IssueToken(ctx.Request.Context(), roomID, req.Name)
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

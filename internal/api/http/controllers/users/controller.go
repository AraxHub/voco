package usersctrl

import (
	"io"
	"net/http"
	"strings"

	"voco/internal/api/http/middlewares"
	"voco/internal/domain"
	"voco/internal/usecase/ports"
	"voco/internal/usecase/users"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	uc      *users.Usecase
	blobs   ports.BlobStore
	baseURL string
	maxBytes int64
}

func New(uc *users.Usecase, blobs ports.BlobStore, baseURL string, maxBytes int64) *Controller {
	if maxBytes <= 0 {
		maxBytes = 10 << 20
	}
	return &Controller{uc: uc, blobs: blobs, baseURL: strings.TrimRight(baseURL, "/"), maxBytes: maxBytes}
}

func (c *Controller) RegisterRoutes(r *gin.Engine, mw middlewares.MW) {
	api := r.Group("/api/v1", mw.Auth, middlewares.EnsureUser(c.uc))
	api.GET("/users/me", c.me)
	api.PATCH("/users/me", c.updateMe)
	api.PUT("/users/me/avatar", c.putAvatar)
	api.GET("/users/search", c.search)
}

type userDTO struct {
	ID          string  `json:"id"`
	Nickname    string  `json:"nickname"`
	Email       string  `json:"email"`
	DisplayName string  `json:"displayName"`
	LastSeenAt  string  `json:"lastSeenAt"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

func (c *Controller) toDTO(u domain.User) userDTO {
	dto := userDTO{
		ID: u.ID.String(), Nickname: u.Nickname, Email: u.Email, DisplayName: u.DisplayName,
		LastSeenAt: u.LastSeenAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if u.AvatarBlobID != nil {
		url := c.blobURL(u.AvatarBlobID.String())
		dto.AvatarURL = &url
	}
	return dto
}

func (c *Controller) blobURL(id string) string {
	if c.baseURL == "" {
		return "/api/v1/blobs/" + id
	}
	return c.baseURL + "/api/v1/blobs/" + id
}

func (c *Controller) me(ctx *gin.Context) {
	u, ok := middlewares.DomainUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fresh, err := c.uc.Me(ctx.Request.Context(), u.ID)
	if err != nil {
		ctx.JSON(http.StatusOK, c.toDTO(u))
		return
	}
	ctx.JSON(http.StatusOK, c.toDTO(fresh))
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
	ctx.JSON(http.StatusOK, c.toDTO(updated))
}

func (c *Controller) putAvatar(ctx *gin.Context) {
	u, ok := middlewares.DomainUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if c.blobs == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "blobs unavailable"})
		return
	}
	fh, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	f, err := fh.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, c.maxBytes+1))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if int64(len(data)) > c.maxBytes {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}
	ct := fh.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image required"})
		return
	}
	owner := u.ID
	blob, err := c.blobs.Put(ctx.Request.Context(), domain.Blob{
		OwnerUserID: &owner, ContentType: ct, Data: data,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id := blob.ID
	updated, err := c.uc.UpdateAvatar(ctx.Request.Context(), u.ID, &id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, c.toDTO(updated))
}

func (c *Controller) search(ctx *gin.Context) {
	list, err := c.uc.Search(ctx.Request.Context(), ctx.Query("q"), 20)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]userDTO, 0, len(list))
	for _, u := range list {
		out = append(out, c.toDTO(u))
	}
	ctx.JSON(http.StatusOK, out)
}

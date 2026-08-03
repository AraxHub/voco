package blobsctrl

import (
	"io"
	"net/http"
	"strings"

	"voco/internal/api/http/middlewares"
	"voco/internal/domain"
	"voco/internal/usecase/ports"
	"voco/internal/usecase/users"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	blobs   ports.BlobStore
	users   *users.Usecase
	baseURL string
	maxBytes int64
}

func New(blobs ports.BlobStore, usersUC *users.Usecase, baseURL string, maxBytes int64) *Controller {
	if maxBytes <= 0 {
		maxBytes = 10 << 20
	}
	return &Controller{blobs: blobs, users: usersUC, baseURL: strings.TrimRight(baseURL, "/"), maxBytes: maxBytes}
}

func (c *Controller) RegisterRoutes(r *gin.Engine, mw middlewares.MW) {
	api := r.Group("/api/v1", mw.Auth, middlewares.EnsureUser(c.users))
	api.POST("/blobs", c.upload)
	api.GET("/blobs/:id", c.get)
}

func BlobURL(baseURL, id string) string {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		return "/api/v1/blobs/" + id
	}
	return base + "/api/v1/blobs/" + id
}

func (c *Controller) upload(ctx *gin.Context) {
	u, ok := middlewares.DomainUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fh, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	if fh.Size > c.maxBytes {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
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
	if ct == "" {
		ct = "application/octet-stream"
	}
	owner := u.ID
	blob, err := c.blobs.Put(ctx.Request.Context(), domain.Blob{
		OwnerUserID: &owner,
		ContentType: ct,
		Data:        data,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"id":          blob.ID.String(),
		"contentType": blob.ContentType,
		"byteSize":    blob.ByteSize,
		"url":         BlobURL(c.baseURL, blob.ID.String()),
	})
}

func (c *Controller) get(ctx *gin.Context) {
	if _, ok := middlewares.DomainUser(ctx); !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	blob, err := c.blobs.Get(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ctx.Header("Cache-Control", "private, max-age=86400")
	ctx.Data(http.StatusOK, blob.ContentType, blob.Data)
}

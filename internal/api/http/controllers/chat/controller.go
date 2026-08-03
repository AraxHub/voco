package chatctrl

import (
	"io"
	"net/http"
	"strings"

	"voco/internal/api/http/middlewares"
	"voco/internal/domain"
	"voco/internal/usecase/chat"
	"voco/internal/usecase/users"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	uc     *chat.Usecase
	users  *users.Usecase
	baseURL string
}

func New(uc *chat.Usecase, usersUC *users.Usecase, baseURL string) *Controller {
	return &Controller{uc: uc, users: usersUC, baseURL: baseURL}
}

func (c *Controller) RegisterRoutes(r *gin.Engine, mw middlewares.MW) {
	api := r.Group("/api/v1", mw.Auth, middlewares.EnsureUser(c.users))
	api.GET("/conversations", c.list)
	api.POST("/conversations/direct", c.direct)
	api.POST("/conversations/groups", c.group)
	api.GET("/conversations/:id/members", c.members)
	api.POST("/conversations/:id/members", c.addMember)
	api.PATCH("/conversations/:id", c.patchConversation)
	api.POST("/conversations/:id/leave", c.leave)
	api.POST("/conversations/:id/admins", c.promote)
	api.GET("/conversations/:id/messages", c.messages)
	api.POST("/conversations/:id/messages", c.send)
	api.PATCH("/conversations/:id/messages/:mid", c.edit)
	api.DELETE("/conversations/:id/messages/:mid", c.del)
	api.POST("/conversations/:id/messages/:mid/reactions", c.react)
	api.POST("/conversations/:id/read", c.read)
	api.POST("/conversations/:id/typing", c.typing)
	api.GET("/conversations/:id/request", c.getRequest)
	api.POST("/conversations/:id/request/accept", c.accept)
	api.POST("/conversations/:id/request/block", c.blockReq)
	api.POST("/conversations/:id/call", c.call)
	api.POST("/calls/:roomId/accept", c.acceptCall)
	api.POST("/calls/:roomId/decline", c.declineCall)
	api.POST("/calls/:roomId/cancel", c.cancelCall)
	api.POST("/users/:id/block", c.blockUser)
	api.DELETE("/users/:id/block", c.unblockUser)
}

func (c *Controller) me(ctx *gin.Context) (domain.User, bool) {
	u, ok := middlewares.DomainUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
	return u, ok
}

func (c *Controller) list(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	list, err := c.uc.ListConversations(ctx.Request.Context(), u.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

func (c *Controller) direct(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	var req struct {
		UserID string `json:"userId"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	peer, err := uuid.Parse(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	conv, reqst, err := c.uc.GetOrCreateDirect(ctx.Request.Context(), u.ID, peer)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"conversation": conv, "request": reqst})
}

func (c *Controller) group(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	var req struct {
		Title   string   `json:"title"`
		Members []string `json:"memberIds"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ids := make([]domain.UserID, 0, len(req.Members))
	for _, s := range req.Members {
		id, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	conv, err := c.uc.CreateGroup(ctx.Request.Context(), u.ID, req.Title, ids, nil)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, conv)
}

func (c *Controller) members(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	list, err := c.uc.ListMembers(ctx.Request.Context(), u.ID, cid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, m := range list {
		name := strings.TrimSpace(m.DisplayName)
		if name == "" {
			name = strings.TrimSpace(m.Nickname)
		}
		if name == "" {
			name = m.UserID.String()
		}
		out = append(out, gin.H{
			"userId":      m.UserID,
			"role":        m.Role,
			"nickname":    m.Nickname,
			"displayName": m.DisplayName,
			"name":        name,
		})
	}
	ctx.JSON(http.StatusOK, out)
}

func (c *Controller) addMember(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		UserID string `json:"userId"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	target, err := uuid.Parse(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	if err := c.uc.AddGroupMember(ctx.Request.Context(), u.ID, target, cid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) patchConversation(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Title string `json:"title"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.uc.RenameGroup(ctx.Request.Context(), u.ID, cid, req.Title); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) leave(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := c.uc.Leave(ctx.Request.Context(), u.ID, cid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) promote(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, _ := uuid.Parse(ctx.Param("id"))
	var req struct {
		UserID string `json:"userId"`
	}
	_ = ctx.ShouldBindJSON(&req)
	target, err := uuid.Parse(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	if err := c.uc.PromoteAdmin(ctx.Request.Context(), u.ID, target, cid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) messages(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, _ := uuid.Parse(ctx.Param("id"))
	list, err := c.uc.ListMessages(ctx.Request.Context(), u.ID, cid, 50)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	members, _ := c.uc.ListMembers(ctx.Request.Context(), u.ID, cid)
	names := map[uuid.UUID]string{}
	for _, m := range members {
		name := strings.TrimSpace(m.DisplayName)
		if name == "" {
			name = strings.TrimSpace(m.Nickname)
		}
		if name == "" {
			name = "Участник"
		}
		names[m.UserID] = name
	}
	out := make([]gin.H, 0, len(list))
	for _, m := range list {
		out = append(out, gin.H{
			"ID":              m.ID,
			"id":              m.ID,
			"ConversationID":  m.ConversationID,
			"SenderID":        m.SenderID,
			"senderId":        m.SenderID,
			"senderName":      names[m.SenderID],
			"Body":            m.Body,
			"CreatedAt":       m.CreatedAt,
			"EditedAt":        m.EditedAt,
			"DeletedForAllAt": m.DeletedForAllAt,
			"Reactions":       m.Reactions,
		})
	}
	ctx.JSON(http.StatusOK, out)
}

func (c *Controller) send(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, _ := uuid.Parse(ctx.Param("id"))
	ct := ctx.ContentType()
	var body string
	var atts []chat.AttachmentInput
	if strings.HasPrefix(ct, "multipart/") {
		body = ctx.PostForm("body")
		form, err := ctx.MultipartForm()
		if err == nil && form != nil {
			for _, fh := range form.File["files"] {
				f, err := fh.Open()
				if err != nil {
					continue
				}
				data, _ := io.ReadAll(f)
				_ = f.Close()
				kind := domain.AttachmentFile
				if strings.HasPrefix(fh.Header.Get("Content-Type"), "image/") {
					kind = domain.AttachmentImage
				}
				atts = append(atts, chat.AttachmentInput{
					Filename: fh.Filename, ContentType: fh.Header.Get("Content-Type"), Data: data, Kind: kind,
				})
			}
		}
	} else {
		var req struct {
			Body string `json:"body"`
		}
		_ = ctx.ShouldBindJSON(&req)
		body = req.Body
	}
	msg, err := c.uc.SendMessage(ctx.Request.Context(), u.ID, cid, body, atts)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	senderName := "Участник"
	if meUser, err := c.users.Me(ctx.Request.Context(), u.ID); err == nil {
		if s := strings.TrimSpace(meUser.DisplayName); s != "" {
			senderName = s
		} else if s := strings.TrimSpace(meUser.Nickname); s != "" {
			senderName = s
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"ID": msg.ID, "id": msg.ID,
		"ConversationID": msg.ConversationID,
		"SenderID": msg.SenderID, "senderId": msg.SenderID,
		"senderName": senderName,
		"Body": msg.Body, "CreatedAt": msg.CreatedAt,
		"EditedAt": msg.EditedAt, "DeletedForAllAt": msg.DeletedForAllAt,
		"Reactions": msg.Reactions,
	})
}

func (c *Controller) edit(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	mid, _ := uuid.Parse(ctx.Param("mid"))
	var req struct {
		Body string `json:"body"`
	}
	_ = ctx.ShouldBindJSON(&req)
	msg, err := c.uc.EditMessage(ctx.Request.Context(), u.ID, mid, req.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, msg)
}

func (c *Controller) del(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	mid, _ := uuid.Parse(ctx.Param("mid"))
	mode := domain.DeleteMode(ctx.DefaultQuery("mode", "for_all"))
	if err := c.uc.DeleteMessage(ctx.Request.Context(), u.ID, mid, mode); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) react(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	mid, _ := uuid.Parse(ctx.Param("mid"))
	var req struct {
		Emoji string `json:"emoji"`
		Add   bool   `json:"add"`
	}
	_ = ctx.ShouldBindJSON(&req)
	if err := c.uc.React(ctx.Request.Context(), u.ID, mid, req.Emoji, req.Add); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) read(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, _ := uuid.Parse(ctx.Param("id"))
	var req struct {
		MessageID string `json:"messageId"`
	}
	_ = ctx.ShouldBindJSON(&req)
	mid, _ := uuid.Parse(req.MessageID)
	if err := c.uc.MarkRead(ctx.Request.Context(), u.ID, cid, mid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) typing(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, _ := uuid.Parse(ctx.Param("id"))
	_ = c.uc.Typing(ctx.Request.Context(), u.ID, cid)
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) getRequest(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	req, found, err := c.uc.GetIncomingPendingRequest(ctx.Request.Context(), u.ID, cid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !found {
		ctx.JSON(http.StatusOK, gin.H{"incomingPending": false})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"incomingPending": true,
		"request": gin.H{
			"conversationId": req.ConversationID,
			"fromUserId":     req.FromUserID,
			"toUserId":       req.ToUserID,
			"status":         req.Status,
			"createdAt":      req.CreatedAt,
		},
	})
}

func (c *Controller) accept(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, _ := uuid.Parse(ctx.Param("id"))
	if err := c.uc.AcceptRequest(ctx.Request.Context(), u.ID, cid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) blockReq(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, _ := uuid.Parse(ctx.Param("id"))
	if err := c.uc.BlockRequest(ctx.Request.Context(), u.ID, cid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) call(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	cid, _ := uuid.Parse(ctx.Param("id"))
	room, expires, err := c.uc.CallFromChat(ctx.Request.Context(), u.ID, cid)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	base := strings.TrimRight(c.baseURL, "/")
	ctx.JSON(http.StatusOK, gin.H{
		"roomId":    room.ID.String(),
		"joinUrl":   base + "/room/" + room.ID.String(),
		"expiresAt": expires,
	})
}

func (c *Controller) acceptCall(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	rid, err := domain.ParseRoomID(ctx.Param("roomId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomId"})
		return
	}
	if err := c.uc.AcceptCall(ctx.Request.Context(), u.ID, rid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) declineCall(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	rid, err := domain.ParseRoomID(ctx.Param("roomId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomId"})
		return
	}
	if err := c.uc.DeclineCall(ctx.Request.Context(), u.ID, rid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) cancelCall(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	rid, err := domain.ParseRoomID(ctx.Param("roomId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomId"})
		return
	}
	if err := c.uc.CancelCall(ctx.Request.Context(), u.ID, rid); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) blockUser(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	target, _ := uuid.Parse(ctx.Param("id"))
	if err := c.uc.BlockUser(ctx.Request.Context(), u.ID, target); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) unblockUser(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	target, _ := uuid.Parse(ctx.Param("id"))
	_ = c.uc.UnblockUser(ctx.Request.Context(), u.ID, target)
	ctx.Status(http.StatusNoContent)
}

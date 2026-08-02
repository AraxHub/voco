package calendarctrl

import (
	"net/http"
	"time"

	"voco/internal/api/http/middlewares"
	"voco/internal/domain"
	"voco/internal/usecase/calendar"
	"voco/internal/usecase/users"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	uc    *calendar.Usecase
	users *users.Usecase
}

func New(uc *calendar.Usecase, usersUC *users.Usecase) *Controller {
	return &Controller{uc: uc, users: usersUC}
}

func (c *Controller) RegisterRoutes(r *gin.Engine, mw middlewares.MW) {
	api := r.Group("/api/v1", mw.Auth, middlewares.EnsureUser(c.users))
	api.GET("/calendar/events", c.list)
	api.POST("/calendar/events", c.create)
	api.PATCH("/calendar/events/:id", c.patch)
	api.POST("/calendar/events/:id/cancel", c.cancel)
	api.POST("/calendar/events/:id/rsvp", c.rsvp)
	api.GET("/calendar/freebusy", c.freebusy)
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
	from, _ := time.Parse(time.RFC3339, ctx.Query("from"))
	to, _ := time.Parse(time.RFC3339, ctx.Query("to"))
	if from.IsZero() {
		from = time.Now().UTC().AddDate(0, -1, 0)
	}
	if to.IsZero() {
		to = time.Now().UTC().AddDate(0, 2, 0)
	}
	list, err := c.uc.List(ctx.Request.Context(), u.ID, from, to)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

func (c *Controller) create(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Timezone    string   `json:"timezone"`
		RRule       string   `json:"rrule"`
		StartsAt    string   `json:"startsAt"`
		EndsAt      string   `json:"endsAt"`
		AttendeeIDs []string `json:"attendeeIds"`
		Reminders   []int    `json:"reminders"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	starts, err1 := time.Parse(time.RFC3339, req.StartsAt)
	ends, err2 := time.Parse(time.RFC3339, req.EndsAt)
	if err1 != nil || err2 != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid time"})
		return
	}
	ids := make([]domain.UserID, 0, len(req.AttendeeIDs))
	for _, s := range req.AttendeeIDs {
		id, err := uuid.Parse(s)
		if err == nil {
			ids = append(ids, id)
		}
	}
	ev, err := c.uc.Create(ctx.Request.Context(), u.ID, req.Title, req.Description, req.Timezone, req.RRule, starts, ends, ids, req.Reminders)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, ev)
}

func (c *Controller) patch(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	id, _ := uuid.Parse(ctx.Param("id"))
	var req struct {
		StartsAt string `json:"startsAt"`
		EndsAt   string `json:"endsAt"`
	}
	_ = ctx.ShouldBindJSON(&req)
	starts, err1 := time.Parse(time.RFC3339, req.StartsAt)
	ends, err2 := time.Parse(time.RFC3339, req.EndsAt)
	if err1 != nil || err2 != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid time"})
		return
	}
	ev, err := c.uc.Reschedule(ctx.Request.Context(), u.ID, id, starts, ends)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, ev)
}

func (c *Controller) cancel(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	id, _ := uuid.Parse(ctx.Param("id"))
	ev, err := c.uc.Cancel(ctx.Request.Context(), u.ID, id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, ev)
}

func (c *Controller) rsvp(ctx *gin.Context) {
	u, ok := c.me(ctx)
	if !ok {
		return
	}
	id, _ := uuid.Parse(ctx.Param("id"))
	var req struct {
		Status string `json:"status"`
	}
	_ = ctx.ShouldBindJSON(&req)
	if err := c.uc.RSVP(ctx.Request.Context(), u.ID, id, domain.RSVPStatus(req.Status)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) freebusy(ctx *gin.Context) {
	_, ok := c.me(ctx)
	if !ok {
		return
	}
	from, _ := time.Parse(time.RFC3339, ctx.Query("from"))
	to, _ := time.Parse(time.RFC3339, ctx.Query("to"))
	raw := ctx.QueryArray("userIds")
	if len(raw) == 0 {
		if q := ctx.Query("userIds"); q != "" {
			raw = []string{q}
		}
	}
	ids := make([]domain.UserID, 0, len(raw))
	for _, s := range raw {
		id, err := uuid.Parse(s)
		if err == nil {
			ids = append(ids, id)
		}
	}
	busy, err := c.uc.FreeBusy(ctx.Request.Context(), ids, from, to)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, busy)
}

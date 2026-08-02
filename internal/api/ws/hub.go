package ws

import (
	"encoding/json"
	"net/http"
	"sync"

	"voco/internal/domain"
	"voco/internal/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*websocket.Conn]struct{}
	auth    *auth.Service
	resolve func(sub string) (domain.UserID, error)
}

func NewHub(authSvc *auth.Service, resolve func(sub string) (domain.UserID, error)) *Hub {
	return &Hub{
		clients: map[uuid.UUID]map[*websocket.Conn]struct{}{},
		auth:    authSvc,
		resolve: resolve,
	}
}

func (h *Hub) PublishToUsers(userIDs []domain.UserID, event string, payload any) {
	msg, err := json.Marshal(map[string]any{"event": event, "payload": payload})
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, id := range userIDs {
		for c := range h.clients[id] {
			_ = c.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func (h *Hub) Handle(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}
	if h.auth == nil || !h.auth.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth disabled"})
		return
	}
	u, err := h.auth.Verify(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	uid, err := h.resolve(u.Sub)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.mu.Lock()
	if h.clients[uid] == nil {
		h.clients[uid] = map[*websocket.Conn]struct{}{}
	}
	h.clients[uid][conn] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients[uid], conn)
		if len(h.clients[uid]) == 0 {
			delete(h.clients, uid)
		}
		h.mu.Unlock()
		_ = conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

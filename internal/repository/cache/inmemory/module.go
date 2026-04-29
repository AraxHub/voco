package inmemory

import (
	"context"
	"sync"
	"time"

	"voco/internal/domain"
)

type Config struct {
	Ttl time.Duration
}

type item struct {
	room      domain.Room
	expiresAt time.Time
}

type Cache struct {
	storage map[domain.RoomID]item
	mu      sync.RWMutex
}

func NewCache(cfg Config) *Cache {
	_ = cfg
	return &Cache{
		storage: make(map[domain.RoomID]item),
	}
}

// CleanUp запускает фоновую очистку истёкших записей.
func (c *Cache) CleanUp(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}

	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				c.deleteExpired(time.Now())
			}
		}
	}()
}

func (c *Cache) deleteExpired(now time.Time) {
	c.mu.Lock()
	for id, it := range c.storage {
		if it.expiresAt.IsZero() {
			continue
		}
		if now.After(it.expiresAt) {
			delete(c.storage, id)
		}
	}
	c.mu.Unlock()
}

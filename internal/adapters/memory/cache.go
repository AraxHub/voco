package memory

import (
	"context"
	"sync"
	"time"

	"voco/internal/domain"
)

type Config struct {
	Ttl time.Duration `envconfig:"TTL" default:"24h"`
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

func (c *Cache) Get(ctx context.Context, id domain.RoomID) (domain.Room, bool, error) {
	_ = ctx

	c.mu.RLock()
	it, ok := c.storage[id]
	c.mu.RUnlock()
	if !ok {
		return domain.Room{}, false, nil
	}

	if !it.expiresAt.IsZero() && time.Now().After(it.expiresAt) {
		c.mu.Lock()
		it2, ok2 := c.storage[id]
		if ok2 && it2.expiresAt == it.expiresAt {
			delete(c.storage, id)
		}
		c.mu.Unlock()
		return domain.Room{}, false, nil
	}

	return it.room, true, nil
}

func (c *Cache) Upsert(ctx context.Context, room domain.Room, ttl time.Duration) error {
	_ = ctx

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	c.mu.Lock()
	c.storage[room.ID] = item{room: room, expiresAt: exp}
	c.mu.Unlock()
	return nil
}

func (c *Cache) Delete(ctx context.Context, id domain.RoomID) error {
	_ = ctx

	c.mu.Lock()
	delete(c.storage, id)
	c.mu.Unlock()
	return nil
}

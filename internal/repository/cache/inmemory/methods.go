package inmemory

import (
	"context"
	"time"

	"voco/internal/domain"
)

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
		// double-check under write lock
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

package users

import (
	"context"
	"strings"
	"sync"
	"time"

	"voco/internal/domain"

	"github.com/google/uuid"
)

type MemoryStore struct {
	mu    sync.RWMutex
	byID  map[uuid.UUID]domain.User
	bySub map[string]uuid.UUID
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		byID:  map[uuid.UUID]domain.User{},
		bySub: map[string]uuid.UUID{},
	}
}

func (s *MemoryStore) UpsertByKeycloakSub(_ context.Context, sub, email, displayName string) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.bySub[sub]; ok {
		u := s.byID[id]
		u.Email = email
		if displayName != "" {
			u.DisplayName = displayName
		}
		u.UpdatedAt = time.Now().UTC()
		s.byID[id] = u
		return u, nil
	}
	now := time.Now().UTC()
	u := domain.User{
		ID:          uuid.New(),
		KeycloakSub: sub,
		Email:       email,
		DisplayName: displayName,
		Nickname:    "",
		CreatedAt:   now,
		UpdatedAt:   now,
		LastSeenAt:  now,
	}
	s.byID[u.ID] = u
	s.bySub[sub] = u.ID
	return u, nil
}

func (s *MemoryStore) GetByID(_ context.Context, id domain.UserID) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byID[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, nil
}

func (s *MemoryStore) GetByKeycloakSub(_ context.Context, sub string) (domain.User, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.bySub[sub]
	if !ok {
		return domain.User{}, false, nil
	}
	return s.byID[id], true, nil
}

func (s *MemoryStore) UpdateProfile(_ context.Context, id domain.UserID, nickname, displayName string) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.byID {
		if u.ID != id && strings.EqualFold(u.Nickname, nickname) && nickname != "" {
			return domain.User{}, domain.ErrNicknameTaken
		}
	}
	u, ok := s.byID[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	u.Nickname = nickname
	if displayName != "" {
		u.DisplayName = displayName
	}
	u.UpdatedAt = time.Now().UTC()
	s.byID[id] = u
	return u, nil
}

func (s *MemoryStore) SearchByNickname(_ context.Context, query string, limit int) ([]domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q := strings.ToLower(query)
	out := make([]domain.User, 0)
	for _, u := range s.byID {
		if u.Nickname != "" && strings.Contains(strings.ToLower(u.Nickname), q) {
			out = append(out, u)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (s *MemoryStore) TouchLastSeen(_ context.Context, id domain.UserID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.byID[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.LastSeenAt = time.Now().UTC()
	s.byID[id] = u
	return nil
}

func (s *MemoryStore) ListAll(_ context.Context) ([]domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.User, 0, len(s.byID))
	for _, u := range s.byID {
		out = append(out, u)
	}
	return out, nil
}

func (s *MemoryStore) UpsertSynced(_ context.Context, sub, email, displayName, nickname string) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if id, ok := s.bySub[sub]; ok {
		u := s.byID[id]
		u.Email = email
		u.DisplayName = displayName
		if u.Nickname == "" && nickname != "" {
			// avoid clash
			clash := false
			for _, o := range s.byID {
				if o.ID != id && strings.EqualFold(o.Nickname, nickname) {
					clash = true
					break
				}
			}
			if !clash {
				u.Nickname = nickname
			}
		}
		u.UpdatedAt = now
		s.byID[id] = u
		return u, nil
	}
	u := domain.User{
		ID: uuid.New(), KeycloakSub: sub, Email: email, DisplayName: displayName,
		Nickname: nickname, CreatedAt: now, UpdatedAt: now, LastSeenAt: now,
	}
	for _, o := range s.byID {
		if strings.EqualFold(o.Nickname, nickname) && nickname != "" {
			u.Nickname = ""
			break
		}
	}
	s.byID[u.ID] = u
	s.bySub[sub] = u.ID
	return u, nil
}

package chat

import (
	"context"
	"sync"
	"time"

	"voco/internal/domain"

	"github.com/google/uuid"
)

type MemoryStore struct {
	mu            sync.RWMutex
	conversations map[uuid.UUID]domain.Conversation
	members       map[uuid.UUID]map[uuid.UUID]domain.ConversationMember // cid -> uid
	direct        map[[2]uuid.UUID]uuid.UUID
	requests      map[uuid.UUID]domain.MessageRequest
	blocks        map[[2]uuid.UUID]struct{}
	messages      map[uuid.UUID]domain.Message
	hidden        map[[2]uuid.UUID]struct{} // user, message
	reads         map[[2]uuid.UUID]uuid.UUID
	reactions     map[uuid.UUID][]domain.MessageReaction
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		conversations: map[uuid.UUID]domain.Conversation{},
		members:       map[uuid.UUID]map[uuid.UUID]domain.ConversationMember{},
		direct:        map[[2]uuid.UUID]uuid.UUID{},
		requests:      map[uuid.UUID]domain.MessageRequest{},
		blocks:        map[[2]uuid.UUID]struct{}{},
		messages:      map[uuid.UUID]domain.Message{},
		hidden:        map[[2]uuid.UUID]struct{}{},
		reads:         map[[2]uuid.UUID]uuid.UUID{},
		reactions:     map[uuid.UUID][]domain.MessageReaction{},
	}
}

func (s *MemoryStore) CreateConversation(_ context.Context, c domain.Conversation, members []domain.ConversationMember) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conversations[c.ID] = c
	s.members[c.ID] = map[uuid.UUID]domain.ConversationMember{}
	for _, m := range members {
		s.members[c.ID][m.UserID] = m
	}
	return nil
}

func (s *MemoryStore) GetConversation(_ context.Context, id domain.ConversationID) (domain.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.conversations[id]
	if !ok {
		return domain.Conversation{}, domain.ErrConversationNotFound
	}
	return c, nil
}

func (s *MemoryStore) ListConversationsForUser(_ context.Context, userID domain.UserID) ([]domain.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Conversation
	for cid, mems := range s.members {
		if m, ok := mems[userID]; ok && m.LeftAt == nil {
			out = append(out, s.conversations[cid])
		}
	}
	return out, nil
}

func (s *MemoryStore) AddMember(_ context.Context, m domain.ConversationMember) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.members[m.ConversationID] == nil {
		s.members[m.ConversationID] = map[uuid.UUID]domain.ConversationMember{}
	}
	s.members[m.ConversationID][m.UserID] = m
	return nil
}

func (s *MemoryStore) GetMember(_ context.Context, cid domain.ConversationID, uid domain.UserID) (domain.ConversationMember, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.members[cid][uid]
	return m, ok, nil
}

func (s *MemoryStore) ListMembers(_ context.Context, cid domain.ConversationID) ([]domain.ConversationMember, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.ConversationMember
	for _, m := range s.members[cid] {
		out = append(out, m)
	}
	return out, nil
}

func (s *MemoryStore) SetMemberRole(_ context.Context, cid domain.ConversationID, uid domain.UserID, role domain.MemberRole) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.members[cid][uid]
	if !ok {
		return domain.ErrNotFound
	}
	m.Role = role
	s.members[cid][uid] = m
	return nil
}

func (s *MemoryStore) MarkLeft(_ context.Context, cid domain.ConversationID, uid domain.UserID, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.members[cid][uid]
	if !ok {
		return domain.ErrNotFound
	}
	m.LeftAt = &at
	s.members[cid][uid] = m
	return nil
}

func (s *MemoryStore) GetDirectPair(_ context.Context, a, b domain.UserID) (domain.ConversationID, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.direct[[2]uuid.UUID{a, b}]
	return id, ok, nil
}

func (s *MemoryStore) PutDirectPair(_ context.Context, a, b domain.UserID, cid domain.ConversationID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.direct[[2]uuid.UUID{a, b}] = cid
	return nil
}

func (s *MemoryStore) GetMessageRequest(_ context.Context, cid domain.ConversationID) (domain.MessageRequest, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.requests[cid]
	return r, ok, nil
}

func (s *MemoryStore) UpsertMessageRequest(_ context.Context, req domain.MessageRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests[req.ConversationID] = req
	return nil
}

func (s *MemoryStore) IsBlocked(_ context.Context, blocker, blocked domain.UserID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.blocks[[2]uuid.UUID{blocker, blocked}]
	return ok, nil
}

func (s *MemoryStore) Block(_ context.Context, blocker, blocked domain.UserID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blocks[[2]uuid.UUID{blocker, blocked}] = struct{}{}
	return nil
}

func (s *MemoryStore) Unblock(_ context.Context, blocker, blocked domain.UserID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.blocks, [2]uuid.UUID{blocker, blocked})
	return nil
}

func (s *MemoryStore) InsertMessage(_ context.Context, m domain.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[m.ID] = m
	return nil
}

func (s *MemoryStore) GetMessage(_ context.Context, id domain.MessageID) (domain.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.messages[id]
	if !ok {
		return domain.Message{}, domain.ErrNotFound
	}
	m.Reactions = append([]domain.MessageReaction{}, s.reactions[id]...)
	return m, nil
}

func (s *MemoryStore) UpdateMessageBody(_ context.Context, id domain.MessageID, body string, editedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.messages[id]
	if !ok {
		return domain.ErrNotFound
	}
	m.Body = body
	m.EditedAt = &editedAt
	s.messages[id] = m
	return nil
}

func (s *MemoryStore) SoftDeleteMessage(_ context.Context, id domain.MessageID, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.messages[id]
	if !ok {
		return domain.ErrNotFound
	}
	m.DeletedForAllAt = &at
	m.Body = ""
	s.messages[id] = m
	return nil
}

func (s *MemoryStore) HideMessage(_ context.Context, userID domain.UserID, messageID domain.MessageID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hidden[[2]uuid.UUID{userID, messageID}] = struct{}{}
	return nil
}

func (s *MemoryStore) ListMessages(_ context.Context, cid domain.ConversationID, viewer domain.UserID, limit int, _ *time.Time) ([]domain.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var all []domain.Message
	for _, m := range s.messages {
		if m.ConversationID != cid {
			continue
		}
		if _, hid := s.hidden[[2]uuid.UUID{viewer, m.ID}]; hid {
			continue
		}
		m.Reactions = append([]domain.MessageReaction{}, s.reactions[m.ID]...)
		all = append(all, m)
	}
	// newest first
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].CreatedAt.After(all[i].CreatedAt) {
				all[i], all[j] = all[j], all[i]
			}
		}
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (s *MemoryStore) AddReaction(_ context.Context, r domain.MessageReaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.reactions[r.MessageID]
	for _, x := range list {
		if x.UserID == r.UserID && x.Emoji == r.Emoji {
			return nil
		}
	}
	s.reactions[r.MessageID] = append(list, r)
	return nil
}

func (s *MemoryStore) RemoveReaction(_ context.Context, messageID domain.MessageID, userID domain.UserID, emoji string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.reactions[messageID]
	out := list[:0]
	for _, x := range list {
		if x.UserID == userID && x.Emoji == emoji {
			continue
		}
		out = append(out, x)
	}
	s.reactions[messageID] = out
	return nil
}

func (s *MemoryStore) SetRead(_ context.Context, cid domain.ConversationID, userID domain.UserID, lastMessageID domain.MessageID, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reads[[2]uuid.UUID{cid, userID}] = lastMessageID
	return nil
}

func (s *MemoryStore) GetRead(_ context.Context, cid domain.ConversationID, userID domain.UserID) (domain.MessageID, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.reads[[2]uuid.UUID{cid, userID}]
	return id, ok, nil
}

func (s *MemoryStore) UpdateConversationAvatar(_ context.Context, cid domain.ConversationID, blobID *domain.BlobID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.conversations[cid]
	if !ok {
		return domain.ErrConversationNotFound
	}
	c.AvatarBlobID = blobID
	s.conversations[cid] = c
	return nil
}

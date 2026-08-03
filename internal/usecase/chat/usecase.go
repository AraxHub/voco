package chat

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"voco/internal/domain"
	"voco/internal/usecase/ports"

	"github.com/google/uuid"
)

const (
	MaxGroupMembers = 100
	MaxMessageLen   = 4000
)

type Realtime interface {
	PublishToUsers(userIDs []domain.UserID, event string, payload any)
}

type RoomCreator interface {
	CreateRoom(ctx context.Context, title string, owner *domain.UserID) (domain.Room, error)
	CloseRoom(ctx context.Context, id domain.RoomID) error
}

type UserReader interface {
	GetByID(ctx context.Context, id domain.UserID) (domain.User, error)
}

type Store interface {
	CreateConversation(ctx context.Context, c domain.Conversation, members []domain.ConversationMember) error
	GetConversation(ctx context.Context, id domain.ConversationID) (domain.Conversation, error)
	ListConversationsForUser(ctx context.Context, userID domain.UserID) ([]domain.Conversation, error)
	AddMember(ctx context.Context, m domain.ConversationMember) error
	GetMember(ctx context.Context, cid domain.ConversationID, uid domain.UserID) (domain.ConversationMember, bool, error)
	ListMembers(ctx context.Context, cid domain.ConversationID) ([]domain.ConversationMember, error)
	SetMemberRole(ctx context.Context, cid domain.ConversationID, uid domain.UserID, role domain.MemberRole) error
	MarkLeft(ctx context.Context, cid domain.ConversationID, uid domain.UserID, at time.Time) error
	GetDirectPair(ctx context.Context, a, b domain.UserID) (domain.ConversationID, bool, error)
	PutDirectPair(ctx context.Context, a, b domain.UserID, cid domain.ConversationID) error
	GetMessageRequest(ctx context.Context, cid domain.ConversationID) (domain.MessageRequest, bool, error)
	UpsertMessageRequest(ctx context.Context, req domain.MessageRequest) error
	IsBlocked(ctx context.Context, blocker, blocked domain.UserID) (bool, error)
	Block(ctx context.Context, blocker, blocked domain.UserID) error
	Unblock(ctx context.Context, blocker, blocked domain.UserID) error
	InsertMessage(ctx context.Context, m domain.Message) error
	GetMessage(ctx context.Context, id domain.MessageID) (domain.Message, error)
	UpdateMessageBody(ctx context.Context, id domain.MessageID, body string, editedAt time.Time) error
	SoftDeleteMessage(ctx context.Context, id domain.MessageID, at time.Time) error
	HideMessage(ctx context.Context, userID domain.UserID, messageID domain.MessageID) error
	ListMessages(ctx context.Context, cid domain.ConversationID, viewer domain.UserID, limit int, before *time.Time) ([]domain.Message, error)
	AddReaction(ctx context.Context, r domain.MessageReaction) error
	RemoveReaction(ctx context.Context, messageID domain.MessageID, userID domain.UserID, emoji string) error
	SetRead(ctx context.Context, cid domain.ConversationID, userID domain.UserID, lastMessageID domain.MessageID, at time.Time) error
	GetRead(ctx context.Context, cid domain.ConversationID, userID domain.UserID) (domain.MessageID, bool, error)
	CountUnread(ctx context.Context, cid domain.ConversationID, userID domain.UserID) (int, error)
	UpdateConversationAvatar(ctx context.Context, cid domain.ConversationID, blobID *domain.BlobID) error
	UpdateConversationTitle(ctx context.Context, cid domain.ConversationID, title string) error
}

type Config struct {
	MaxImageBytes int64
	MaxFileBytes  int64
}

type Usecase struct {
	store Store
	blobs ports.BlobStore
	rooms RoomCreator
	users UserReader
	rt    Realtime
	cfg   Config

	callsMu sync.Mutex
	calls   map[uuid.UUID]*pendingCall // roomID -> call
}

type pendingCall struct {
	RoomID         domain.RoomID
	ConversationID domain.ConversationID
	CallerID       domain.UserID
	CalleeID       domain.UserID
	ExpiresAt      time.Time
	done           bool
}

func New(store Store, blobs ports.BlobStore, rooms RoomCreator, users UserReader, rt Realtime, cfg Config) *Usecase {
	if cfg.MaxImageBytes <= 0 {
		cfg.MaxImageBytes = 10 << 20
	}
	if cfg.MaxFileBytes <= 0 {
		cfg.MaxFileBytes = 25 << 20
	}
	return &Usecase{
		store: store, blobs: blobs, rooms: rooms, users: users, rt: rt, cfg: cfg,
		calls: map[uuid.UUID]*pendingCall{},
	}
}

func orderedPair(a, b domain.UserID) (domain.UserID, domain.UserID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

func (uc *Usecase) GetOrCreateDirect(ctx context.Context, me, peer domain.UserID) (domain.Conversation, domain.MessageRequest, error) {
	if me == peer {
		return domain.Conversation{}, domain.MessageRequest{}, domain.ErrValidation
	}
	blocked, err := uc.store.IsBlocked(ctx, peer, me)
	if err != nil {
		return domain.Conversation{}, domain.MessageRequest{}, err
	}
	if blocked {
		return domain.Conversation{}, domain.MessageRequest{}, domain.ErrBlocked
	}
	low, high := orderedPair(me, peer)
	if cid, ok, err := uc.store.GetDirectPair(ctx, low, high); err != nil {
		return domain.Conversation{}, domain.MessageRequest{}, err
	} else if ok {
		c, err := uc.store.GetConversation(ctx, cid)
		if err != nil {
			return domain.Conversation{}, domain.MessageRequest{}, err
		}
		req, _, _ := uc.store.GetMessageRequest(ctx, cid)
		uc.enrichConversationTitle(ctx, me, &c)
		return c, req, nil
	}

	now := time.Now().UTC()
	c := domain.Conversation{
		ID: uuid.New(), Type: domain.ConversationDirect, CreatedBy: me, CreatedAt: now,
	}
	members := []domain.ConversationMember{
		{ConversationID: c.ID, UserID: me, Role: domain.RoleAdmin, JoinedAt: now},
		{ConversationID: c.ID, UserID: peer, Role: domain.RoleMember, JoinedAt: now},
	}
	if err := uc.store.CreateConversation(ctx, c, members); err != nil {
		return domain.Conversation{}, domain.MessageRequest{}, err
	}
	if err := uc.store.PutDirectPair(ctx, low, high, c.ID); err != nil {
		return domain.Conversation{}, domain.MessageRequest{}, err
	}
	req := domain.MessageRequest{
		ConversationID: c.ID, FromUserID: me, ToUserID: peer,
		Status: domain.MessageRequestPending, CreatedAt: now,
	}
	if err := uc.store.UpsertMessageRequest(ctx, req); err != nil {
		return domain.Conversation{}, domain.MessageRequest{}, err
	}
	uc.publish(members, "conversation.created", c)
	uc.publish(members, "conversation.updated", c)
	uc.enrichConversationTitle(ctx, me, &c)
	return c, req, nil
}

func (uc *Usecase) CreateGroup(ctx context.Context, me domain.UserID, title string, memberIDs []domain.UserID, avatar *domain.BlobID) (domain.Conversation, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return domain.Conversation{}, fmt.Errorf("%w: укажите название группы", domain.ErrValidation)
	}
	uniq := map[domain.UserID]struct{}{me: {}}
	for _, id := range memberIDs {
		uniq[id] = struct{}{}
	}
	if len(uniq) > MaxGroupMembers {
		return domain.Conversation{}, domain.ErrValidation
	}
	now := time.Now().UTC()
	c := domain.Conversation{
		ID: uuid.New(), Type: domain.ConversationGroup, Title: title,
		AvatarBlobID: avatar, CreatedBy: me, CreatedAt: now,
	}
	members := make([]domain.ConversationMember, 0, len(uniq))
	for id := range uniq {
		role := domain.RoleMember
		if id == me {
			role = domain.RoleAdmin
		}
		members = append(members, domain.ConversationMember{
			ConversationID: c.ID, UserID: id, Role: role, JoinedAt: now,
		})
	}
	if err := uc.store.CreateConversation(ctx, c, members); err != nil {
		return domain.Conversation{}, err
	}
	uc.publish(members, "conversation.created", c)
	uc.publish(members, "conversation.updated", c)
	return c, nil
}

func (uc *Usecase) PromoteAdmin(ctx context.Context, me, target domain.UserID, cid domain.ConversationID) error {
	m, ok, err := uc.store.GetMember(ctx, cid, me)
	if err != nil {
		return err
	}
	if !ok || m.Role != domain.RoleAdmin || m.LeftAt != nil {
		return domain.ErrForbidden
	}
	t, ok, err := uc.store.GetMember(ctx, cid, target)
	if err != nil {
		return err
	}
	if !ok || t.LeftAt != nil {
		return domain.ErrNotFound
	}
	return uc.store.SetMemberRole(ctx, cid, target, domain.RoleAdmin)
}

func (uc *Usecase) Leave(ctx context.Context, me domain.UserID, cid domain.ConversationID) error {
	m, ok, err := uc.store.GetMember(ctx, cid, me)
	if err != nil {
		return err
	}
	if !ok || m.LeftAt != nil {
		return domain.ErrNotFound
	}
	return uc.store.MarkLeft(ctx, cid, me, time.Now().UTC())
}

type MemberInfo struct {
	UserID      domain.UserID
	Role        domain.MemberRole
	JoinedAt    time.Time
	LeftAt      *time.Time
	Nickname    string
	DisplayName string
}

func (uc *Usecase) ListMembers(ctx context.Context, me domain.UserID, cid domain.ConversationID) ([]MemberInfo, error) {
	if _, ok, err := uc.store.GetMember(ctx, cid, me); err != nil || !ok {
		return nil, domain.ErrForbidden
	}
	members, err := uc.store.ListMembers(ctx, cid)
	if err != nil {
		return nil, err
	}
	out := make([]MemberInfo, 0, len(members))
	for _, m := range members {
		if m.LeftAt != nil {
			continue
		}
		info := MemberInfo{UserID: m.UserID, Role: m.Role, JoinedAt: m.JoinedAt, LeftAt: m.LeftAt}
		if uc.users != nil {
			if u, err := uc.users.GetByID(ctx, m.UserID); err == nil {
				info.Nickname = u.Nickname
				info.DisplayName = u.DisplayName
			}
		}
		out = append(out, info)
	}
	return out, nil
}

func (uc *Usecase) AddGroupMember(ctx context.Context, me, target domain.UserID, cid domain.ConversationID) error {
	if me == target {
		return domain.ErrValidation
	}
	c, err := uc.store.GetConversation(ctx, cid)
	if err != nil {
		return err
	}
	if c.Type != domain.ConversationGroup {
		return domain.ErrValidation
	}
	admin, ok, err := uc.store.GetMember(ctx, cid, me)
	if err != nil || !ok || admin.LeftAt != nil || admin.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	existing, ok, err := uc.store.GetMember(ctx, cid, target)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if ok && existing.LeftAt == nil {
		return nil
	}
	members, err := uc.store.ListMembers(ctx, cid)
	if err != nil {
		return err
	}
	active := 0
	for _, m := range members {
		if m.LeftAt == nil {
			active++
		}
	}
	if active >= MaxGroupMembers {
		return domain.ErrValidation
	}
	if err := uc.store.AddMember(ctx, domain.ConversationMember{
		ConversationID: cid, UserID: target, Role: domain.RoleMember, JoinedAt: now,
	}); err != nil {
		return err
	}
	members, _ = uc.store.ListMembers(ctx, cid)
	uc.publish(members, "conversation.updated", map[string]any{"id": cid, "added": target})
	return nil
}

func (uc *Usecase) RenameGroup(ctx context.Context, me domain.UserID, cid domain.ConversationID, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("%w: укажите название группы", domain.ErrValidation)
	}
	c, err := uc.store.GetConversation(ctx, cid)
	if err != nil {
		return err
	}
	if c.Type != domain.ConversationGroup {
		return domain.ErrValidation
	}
	admin, ok, err := uc.store.GetMember(ctx, cid, me)
	if err != nil || !ok || admin.LeftAt != nil || admin.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if err := uc.store.UpdateConversationTitle(ctx, cid, title); err != nil {
		return err
	}
	members, _ := uc.store.ListMembers(ctx, cid)
	uc.publish(members, "conversation.updated", map[string]any{"id": cid, "title": title})
	return nil
}

// GetIncomingPendingRequest returns a pending message request addressed to me, if any.
func (uc *Usecase) GetIncomingPendingRequest(ctx context.Context, me domain.UserID, cid domain.ConversationID) (domain.MessageRequest, bool, error) {
	m, ok, err := uc.store.GetMember(ctx, cid, me)
	if err != nil {
		return domain.MessageRequest{}, false, err
	}
	if !ok || m.LeftAt != nil {
		return domain.MessageRequest{}, false, domain.ErrForbidden
	}
	req, ok, err := uc.store.GetMessageRequest(ctx, cid)
	if err != nil || !ok {
		return domain.MessageRequest{}, false, err
	}
	if req.ToUserID != me || req.Status != domain.MessageRequestPending {
		return domain.MessageRequest{}, false, nil
	}
	return req, true, nil
}

func (uc *Usecase) AcceptRequest(ctx context.Context, me domain.UserID, cid domain.ConversationID) error {
	req, ok, err := uc.store.GetMessageRequest(ctx, cid)
	if err != nil {
		return err
	}
	if !ok || req.ToUserID != me || req.Status != domain.MessageRequestPending {
		return domain.ErrForbidden
	}
	now := time.Now().UTC()
	req.Status = domain.MessageRequestAccepted
	req.ResolvedAt = &now
	if err := uc.store.UpsertMessageRequest(ctx, req); err != nil {
		return err
	}
	// Mark all messages as read for accepter (Telegram-like).
	msgs, err := uc.store.ListMessages(ctx, cid, me, 1, nil)
	if err == nil && len(msgs) > 0 {
		_ = uc.store.SetRead(ctx, cid, me, msgs[0].ID, now)
	}
	members, _ := uc.store.ListMembers(ctx, cid)
	uc.publish(members, "conversation.updated", map[string]any{"id": cid, "request": req})
	return nil
}

func (uc *Usecase) BlockRequest(ctx context.Context, me domain.UserID, cid domain.ConversationID) error {
	req, ok, err := uc.store.GetMessageRequest(ctx, cid)
	if err != nil {
		return err
	}
	if !ok || req.ToUserID != me || req.Status != domain.MessageRequestPending {
		return domain.ErrForbidden
	}
	now := time.Now().UTC()
	req.Status = domain.MessageRequestBlocked
	req.ResolvedAt = &now
	if err := uc.store.UpsertMessageRequest(ctx, req); err != nil {
		return err
	}
	return uc.store.Block(ctx, me, req.FromUserID)
}

func (uc *Usecase) BlockUser(ctx context.Context, me, target domain.UserID) error {
	if me == target {
		return domain.ErrValidation
	}
	return uc.store.Block(ctx, me, target)
}

func (uc *Usecase) UnblockUser(ctx context.Context, me, target domain.UserID) error {
	return uc.store.Unblock(ctx, me, target)
}

type AttachmentInput struct {
	Filename    string
	ContentType string
	Data        []byte
	Kind        domain.AttachmentKind
}

func (uc *Usecase) SendMessage(ctx context.Context, me domain.UserID, cid domain.ConversationID, body string, attachments []AttachmentInput) (domain.Message, error) {
	body = strings.TrimSpace(body)
	if body == "" && len(attachments) == 0 {
		return domain.Message{}, domain.ErrValidation
	}
	if utf8.RuneCountInString(body) > MaxMessageLen {
		return domain.Message{}, domain.ErrValidation
	}
	if err := uc.ensureCanWrite(ctx, me, cid); err != nil {
		return domain.Message{}, err
	}

	now := time.Now().UTC()
	msg := domain.Message{
		ID: uuid.New(), ConversationID: cid, SenderID: me, Body: body, CreatedAt: now,
	}
	for _, a := range attachments {
		limit := uc.cfg.MaxFileBytes
		if a.Kind == domain.AttachmentImage {
			limit = uc.cfg.MaxImageBytes
		}
		if int64(len(a.Data)) > limit {
			return domain.Message{}, domain.ErrValidation
		}
		owner := me
		blob, err := uc.blobs.Put(ctx, domain.Blob{
			OwnerUserID: &owner, ContentType: a.ContentType, Data: a.Data,
		})
		if err != nil {
			return domain.Message{}, err
		}
		msg.Attachments = append(msg.Attachments, domain.MessageAttachment{
			ID: uuid.New(), BlobID: blob.ID, Kind: a.Kind, Filename: a.Filename,
		})
	}
	if err := uc.store.InsertMessage(ctx, msg); err != nil {
		return domain.Message{}, err
	}
	members, _ := uc.store.ListMembers(ctx, cid)
	uc.publish(members, "message.created", msg)
	return msg, nil
}

func (uc *Usecase) ensureCanWrite(ctx context.Context, me domain.UserID, cid domain.ConversationID) error {
	m, ok, err := uc.store.GetMember(ctx, cid, me)
	if err != nil {
		return err
	}
	if !ok || m.LeftAt != nil {
		return domain.ErrForbidden
	}
	c, err := uc.store.GetConversation(ctx, cid)
	if err != nil {
		return err
	}
	if c.Type == domain.ConversationDirect {
		req, ok, err := uc.store.GetMessageRequest(ctx, cid)
		if err != nil {
			return err
		}
		if ok && req.Status == domain.MessageRequestBlocked {
			return domain.ErrBlocked
		}
		// Peer may have blocked me.
		members, err := uc.store.ListMembers(ctx, cid)
		if err != nil {
			return err
		}
		for _, mem := range members {
			if mem.UserID == me || mem.LeftAt != nil {
				continue
			}
			blocked, err := uc.store.IsBlocked(ctx, mem.UserID, me)
			if err != nil {
				return err
			}
			if blocked {
				return domain.ErrBlocked
			}
			// If pending and I'm the recipient who hasn't accepted — still can? TG: recipient can reply after accept only.
			// Sender can keep sending while pending.
			if ok && req.Status == domain.MessageRequestPending && me == req.ToUserID {
				return domain.ErrMessageRequestPending
			}
		}
	}
	return nil
}

func (uc *Usecase) EditMessage(ctx context.Context, me domain.UserID, mid domain.MessageID, body string) (domain.Message, error) {
	body = strings.TrimSpace(body)
	if body == "" || utf8.RuneCountInString(body) > MaxMessageLen {
		return domain.Message{}, domain.ErrValidation
	}
	msg, err := uc.store.GetMessage(ctx, mid)
	if err != nil {
		return domain.Message{}, err
	}
	if msg.SenderID != me || msg.DeletedForAllAt != nil {
		return domain.Message{}, domain.ErrForbidden
	}
	now := time.Now().UTC()
	if err := uc.store.UpdateMessageBody(ctx, mid, body, now); err != nil {
		return domain.Message{}, err
	}
	msg.Body = body
	msg.EditedAt = &now
	members, _ := uc.store.ListMembers(ctx, msg.ConversationID)
	uc.publish(members, "message.updated", msg)
	return msg, nil
}

func (uc *Usecase) DeleteMessage(ctx context.Context, me domain.UserID, mid domain.MessageID, mode domain.DeleteMode) error {
	msg, err := uc.store.GetMessage(ctx, mid)
	if err != nil {
		return err
	}
	if mode == domain.DeleteForMe {
		return uc.store.HideMessage(ctx, me, mid)
	}
	if msg.SenderID != me {
		return domain.ErrForbidden
	}
	now := time.Now().UTC()
	if err := uc.store.SoftDeleteMessage(ctx, mid, now); err != nil {
		return err
	}
	members, _ := uc.store.ListMembers(ctx, msg.ConversationID)
	msg.DeletedForAllAt = &now
	uc.publish(members, "message.deleted", msg)
	return nil
}

func (uc *Usecase) React(ctx context.Context, me domain.UserID, mid domain.MessageID, emoji string, add bool) error {
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		return domain.ErrValidation
	}
	msg, err := uc.store.GetMessage(ctx, mid)
	if err != nil {
		return err
	}
	if _, ok, err := uc.store.GetMember(ctx, msg.ConversationID, me); err != nil || !ok {
		return domain.ErrForbidden
	}
	if add {
		err = uc.store.AddReaction(ctx, domain.MessageReaction{
			MessageID: mid, UserID: me, Emoji: emoji, CreatedAt: time.Now().UTC(),
		})
	} else {
		err = uc.store.RemoveReaction(ctx, mid, me, emoji)
	}
	if err != nil {
		return err
	}
	members, _ := uc.store.ListMembers(ctx, msg.ConversationID)
	uc.publish(members, "message.updated", map[string]any{"id": mid, "emoji": emoji, "add": add, "userId": me})
	return nil
}

func (uc *Usecase) MarkRead(ctx context.Context, me domain.UserID, cid domain.ConversationID, last domain.MessageID) error {
	if _, ok, err := uc.store.GetMember(ctx, cid, me); err != nil || !ok {
		return domain.ErrForbidden
	}
	now := time.Now().UTC()
	if err := uc.store.SetRead(ctx, cid, me, last, now); err != nil {
		return err
	}
	members, _ := uc.store.ListMembers(ctx, cid)
	uc.publish(members, "read", map[string]any{"conversationId": cid, "userId": me, "messageId": last})
	return nil
}

func (uc *Usecase) Typing(ctx context.Context, me domain.UserID, cid domain.ConversationID) error {
	members, err := uc.store.ListMembers(ctx, cid)
	if err != nil {
		return err
	}
	ok := false
	for _, m := range members {
		if m.UserID == me && m.LeftAt == nil {
			ok = true
			break
		}
	}
	if !ok {
		return domain.ErrForbidden
	}
	uc.publish(members, "typing", map[string]any{"conversationId": cid, "userId": me})
	return nil
}

func (uc *Usecase) CallFromChat(ctx context.Context, me domain.UserID, cid domain.ConversationID) (domain.Room, time.Time, error) {
	c, err := uc.store.GetConversation(ctx, cid)
	if err != nil {
		return domain.Room{}, time.Time{}, err
	}
	if c.Type != domain.ConversationDirect {
		return domain.Room{}, time.Time{}, domain.ErrValidation
	}
	if err := uc.ensureCanWrite(ctx, me, cid); err != nil && err != domain.ErrMessageRequestPending {
		req, ok, _ := uc.store.GetMessageRequest(ctx, cid)
		if ok && req.Status == domain.MessageRequestBlocked {
			return domain.Room{}, time.Time{}, domain.ErrBlocked
		}
		if err == domain.ErrBlocked {
			return domain.Room{}, time.Time{}, err
		}
	}
	members, err := uc.store.ListMembers(ctx, cid)
	if err != nil {
		return domain.Room{}, time.Time{}, err
	}
	var peer domain.UserID
	for _, m := range members {
		if m.LeftAt == nil && m.UserID != me {
			peer = m.UserID
			break
		}
	}
	if peer == uuid.Nil {
		return domain.Room{}, time.Time{}, domain.ErrNotFound
	}

	callerName := "Абонент"
	if uc.users != nil {
		if u, err := uc.users.GetByID(ctx, me); err == nil {
			callerName = userLabel(u)
		}
	}

	title := "Звонок · " + callerName
	owner := me
	room, err := uc.rooms.CreateRoom(ctx, title, &owner)
	if err != nil {
		return domain.Room{}, time.Time{}, err
	}

	expires := time.Now().UTC().Add(time.Minute)
	pc := &pendingCall{
		RoomID: room.ID, ConversationID: cid,
		CallerID: me, CalleeID: peer, ExpiresAt: expires,
	}
	uc.callsMu.Lock()
	uc.calls[room.ID.UUID()] = pc
	uc.callsMu.Unlock()

	payload := map[string]any{
		"roomId":         room.ID.String(),
		"conversationId": cid.String(),
		"callerId":       me.String(),
		"callerName":     callerName,
		"expiresAt":      expires,
	}
	// Both sides get an in-chat call card; callee also gets the fullscreen ring.
	uc.publishUsers([]domain.UserID{me, peer}, "call.started", payload)
	uc.publishUsers([]domain.UserID{peer}, "call.incoming", payload)

	go uc.watchCallTimeout(room.ID.UUID())

	return room, expires, nil
}

func (uc *Usecase) publishUsers(ids []domain.UserID, event string, payload any) {
	if uc.rt == nil {
		return
	}
	uc.rt.PublishToUsers(ids, event, payload)
}

func (uc *Usecase) watchCallTimeout(roomUUID uuid.UUID) {
	uc.callsMu.Lock()
	pc := uc.calls[roomUUID]
	uc.callsMu.Unlock()
	if pc == nil {
		return
	}
	wait := time.Until(pc.ExpiresAt)
	if wait < 0 {
		wait = 0
	}
	time.Sleep(wait)

	uc.callsMu.Lock()
	pc = uc.calls[roomUUID]
	if pc == nil || pc.done {
		uc.callsMu.Unlock()
		return
	}
	pc.done = true
	delete(uc.calls, roomUUID)
	caller, callee := pc.CallerID, pc.CalleeID
	roomID := pc.RoomID
	uc.callsMu.Unlock()

	_ = uc.rooms.CloseRoom(context.Background(), roomID)
	payload := map[string]any{"roomId": roomID.String(), "reason": "timeout"}
	uc.publishUsers([]domain.UserID{caller, callee}, "call.missed", payload)
}

func (uc *Usecase) AcceptCall(ctx context.Context, me domain.UserID, roomID domain.RoomID) error {
	uc.callsMu.Lock()
	pc := uc.calls[roomID.UUID()]
	if pc == nil || pc.done {
		uc.callsMu.Unlock()
		return domain.ErrNotFound
	}
	if pc.CalleeID != me {
		uc.callsMu.Unlock()
		return domain.ErrForbidden
	}
	pc.done = true
	delete(uc.calls, roomID.UUID())
	caller := pc.CallerID
	uc.callsMu.Unlock()

	uc.publishUsers([]domain.UserID{caller, me}, "call.accepted", map[string]any{
		"roomId": roomID.String(),
	})
	return nil
}

func (uc *Usecase) DeclineCall(ctx context.Context, me domain.UserID, roomID domain.RoomID) error {
	uc.callsMu.Lock()
	pc := uc.calls[roomID.UUID()]
	if pc == nil || pc.done {
		uc.callsMu.Unlock()
		return domain.ErrNotFound
	}
	if pc.CalleeID != me {
		uc.callsMu.Unlock()
		return domain.ErrForbidden
	}
	pc.done = true
	delete(uc.calls, roomID.UUID())
	caller := pc.CallerID
	uc.callsMu.Unlock()

	_ = uc.rooms.CloseRoom(ctx, roomID)
	uc.publishUsers([]domain.UserID{caller, me}, "call.declined", map[string]any{
		"roomId": roomID.String(),
	})
	return nil
}

func (uc *Usecase) CancelCall(ctx context.Context, me domain.UserID, roomID domain.RoomID) error {
	uc.callsMu.Lock()
	pc := uc.calls[roomID.UUID()]
	if pc == nil || pc.done {
		uc.callsMu.Unlock()
		return domain.ErrNotFound
	}
	if pc.CallerID != me {
		uc.callsMu.Unlock()
		return domain.ErrForbidden
	}
	pc.done = true
	delete(uc.calls, roomID.UUID())
	callee := pc.CalleeID
	uc.callsMu.Unlock()

	_ = uc.rooms.CloseRoom(ctx, roomID)
	uc.publishUsers([]domain.UserID{me, callee}, "call.cancelled", map[string]any{
		"roomId": roomID.String(),
	})
	return nil
}

func userLabel(u domain.User) string {
	if s := strings.TrimSpace(u.DisplayName); s != "" {
		return s
	}
	if s := strings.TrimSpace(u.Nickname); s != "" {
		return s
	}
	return "Чат"
}

func (uc *Usecase) enrichConversationTitle(ctx context.Context, me domain.UserID, c *domain.Conversation) {
	if c == nil || c.Type != domain.ConversationDirect || uc.users == nil {
		return
	}
	members, err := uc.store.ListMembers(ctx, c.ID)
	if err != nil {
		return
	}
	for _, m := range members {
		if m.LeftAt != nil || m.UserID == me {
			continue
		}
		u, err := uc.users.GetByID(ctx, m.UserID)
		if err != nil {
			return
		}
		c.Title = userLabel(u)
		return
	}
}

func (uc *Usecase) ListConversations(ctx context.Context, me domain.UserID) ([]domain.Conversation, error) {
	list, err := uc.store.ListConversationsForUser(ctx, me)
	if err != nil {
		return nil, err
	}
	for i := range list {
		uc.enrichConversationTitle(ctx, me, &list[i])
	}
	return list, nil
}

// ConversationPreview is a chat-list row with last message and unread.
type ConversationPreview struct {
	Conversation domain.Conversation
	AvatarURL    string
	PeerUserID   *domain.UserID
	LastMessage  *domain.Message
	UnreadCount  int
}

func (uc *Usecase) conversationAvatarURL(ctx context.Context, me domain.UserID, c domain.Conversation, baseURL string) (avatarURL string, peerID *domain.UserID) {
	blobURL := func(id domain.BlobID) string {
		base := strings.TrimRight(baseURL, "/")
		path := "/api/v1/blobs/" + id.String()
		if base == "" {
			return path
		}
		return base + path
	}
	if c.Type == domain.ConversationGroup && c.AvatarBlobID != nil {
		return blobURL(*c.AvatarBlobID), nil
	}
	if c.Type != domain.ConversationDirect || uc.users == nil {
		if c.AvatarBlobID != nil {
			return blobURL(*c.AvatarBlobID), nil
		}
		return "", nil
	}
	members, err := uc.store.ListMembers(ctx, c.ID)
	if err != nil {
		return "", nil
	}
	for _, m := range members {
		if m.LeftAt != nil || m.UserID == me {
			continue
		}
		pid := m.UserID
		peerID = &pid
		u, err := uc.users.GetByID(ctx, m.UserID)
		if err != nil {
			return "", peerID
		}
		if u.AvatarBlobID != nil {
			return blobURL(*u.AvatarBlobID), peerID
		}
		return "", peerID
	}
	return "", nil
}

func (uc *Usecase) ListConversationPreviews(ctx context.Context, me domain.UserID, baseURL string) ([]ConversationPreview, error) {
	list, err := uc.ListConversations(ctx, me)
	if err != nil {
		return nil, err
	}
	out := make([]ConversationPreview, 0, len(list))
	for _, c := range list {
		p := ConversationPreview{Conversation: c}
		p.AvatarURL, p.PeerUserID = uc.conversationAvatarURL(ctx, me, c, baseURL)
		msgs, err := uc.store.ListMessages(ctx, c.ID, me, 1, nil)
		if err == nil && len(msgs) > 0 {
			msg := msgs[0]
			p.LastMessage = &msg
		}
		if n, err := uc.store.CountUnread(ctx, c.ID, me); err == nil {
			p.UnreadCount = n
		}
		out = append(out, p)
	}
	// Sort by last activity (newest first); empty chats by CreatedAt.
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			ti := out[i].Conversation.CreatedAt
			if out[i].LastMessage != nil {
				ti = out[i].LastMessage.CreatedAt
			}
			tj := out[j].Conversation.CreatedAt
			if out[j].LastMessage != nil {
				tj = out[j].LastMessage.CreatedAt
			}
			if tj.After(ti) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}

func (uc *Usecase) ListMessages(ctx context.Context, me domain.UserID, cid domain.ConversationID, limit int) ([]domain.Message, error) {
	if _, ok, err := uc.store.GetMember(ctx, cid, me); err != nil || !ok {
		return nil, domain.ErrForbidden
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return uc.store.ListMessages(ctx, cid, me, limit, nil)
}

func (uc *Usecase) SetAvatar(ctx context.Context, me domain.UserID, cid domain.ConversationID, blobID domain.BlobID) error {
	m, ok, err := uc.store.GetMember(ctx, cid, me)
	if err != nil || !ok || m.LeftAt != nil {
		return domain.ErrForbidden
	}
	c, err := uc.store.GetConversation(ctx, cid)
	if err != nil {
		return err
	}
	if c.Type != domain.ConversationGroup {
		return domain.ErrValidation
	}
	return uc.store.UpdateConversationAvatar(ctx, cid, &blobID)
}

func (uc *Usecase) publish(members []domain.ConversationMember, event string, payload any) {
	if uc.rt == nil {
		return
	}
	ids := make([]domain.UserID, 0, len(members))
	for _, m := range members {
		if m.LeftAt == nil {
			ids = append(ids, m.UserID)
		}
	}
	uc.rt.PublishToUsers(ids, event, payload)
}

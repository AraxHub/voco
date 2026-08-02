package postgres

import (
	"context"
	"errors"
	"time"

	pgadapter "voco/internal/adapters/postgres"
	"voco/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ChatRepo struct {
	db *pgadapter.Client
}

func NewChatRepo(db *pgadapter.Client) *ChatRepo {
	return &ChatRepo{db: db}
}

func (r *ChatRepo) CreateConversation(ctx context.Context, c domain.Conversation, members []domain.ConversationMember) error {
	tx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		INSERT INTO conversations (id, type, title, avatar_blob_id, created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		c.ID, string(c.Type), c.Title, c.AvatarBlobID, c.CreatedBy, c.CreatedAt)
	if err != nil {
		return err
	}
	for _, m := range members {
		_, err = tx.Exec(ctx, `
			INSERT INTO conversation_members (conversation_id, user_id, role, joined_at)
			VALUES ($1,$2,$3,$4)`, m.ConversationID, m.UserID, string(m.Role), m.JoinedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *ChatRepo) GetConversation(ctx context.Context, id domain.ConversationID) (domain.Conversation, error) {
	var c domain.Conversation
	var typ string
	err := r.db.QueryRow(ctx, `
		SELECT id, type, title, avatar_blob_id, created_by, created_at FROM conversations WHERE id=$1`, id).
		Scan(&c.ID, &typ, &c.Title, &c.AvatarBlobID, &c.CreatedBy, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Conversation{}, domain.ErrConversationNotFound
	}
	c.Type = domain.ConversationType(typ)
	return c, err
}

func (r *ChatRepo) ListConversationsForUser(ctx context.Context, userID domain.UserID) ([]domain.Conversation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.type, c.title, c.avatar_blob_id, c.created_by, c.created_at
		FROM conversations c
		JOIN conversation_members m ON m.conversation_id = c.id
		WHERE m.user_id = $1 AND m.left_at IS NULL
		ORDER BY c.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Conversation
	for rows.Next() {
		var c domain.Conversation
		var typ string
		if err := rows.Scan(&c.ID, &typ, &c.Title, &c.AvatarBlobID, &c.CreatedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Type = domain.ConversationType(typ)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ChatRepo) AddMember(ctx context.Context, m domain.ConversationMember) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO conversation_members (conversation_id, user_id, role, joined_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (conversation_id, user_id) DO UPDATE SET left_at = NULL, role = EXCLUDED.role`,
		m.ConversationID, m.UserID, string(m.Role), m.JoinedAt)
	return err
}

func (r *ChatRepo) GetMember(ctx context.Context, cid domain.ConversationID, uid domain.UserID) (domain.ConversationMember, bool, error) {
	var m domain.ConversationMember
	var role string
	err := r.db.QueryRow(ctx, `
		SELECT conversation_id, user_id, role, joined_at, left_at FROM conversation_members
		WHERE conversation_id=$1 AND user_id=$2`, cid, uid).
		Scan(&m.ConversationID, &m.UserID, &role, &m.JoinedAt, &m.LeftAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ConversationMember{}, false, nil
	}
	m.Role = domain.MemberRole(role)
	return m, true, err
}

func (r *ChatRepo) ListMembers(ctx context.Context, cid domain.ConversationID) ([]domain.ConversationMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT conversation_id, user_id, role, joined_at, left_at FROM conversation_members WHERE conversation_id=$1`, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ConversationMember
	for rows.Next() {
		var m domain.ConversationMember
		var role string
		if err := rows.Scan(&m.ConversationID, &m.UserID, &role, &m.JoinedAt, &m.LeftAt); err != nil {
			return nil, err
		}
		m.Role = domain.MemberRole(role)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *ChatRepo) SetMemberRole(ctx context.Context, cid domain.ConversationID, uid domain.UserID, role domain.MemberRole) error {
	tag, err := r.db.Exec(ctx, `UPDATE conversation_members SET role=$3 WHERE conversation_id=$1 AND user_id=$2`, cid, uid, string(role))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ChatRepo) MarkLeft(ctx context.Context, cid domain.ConversationID, uid domain.UserID, at time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE conversation_members SET left_at=$3 WHERE conversation_id=$1 AND user_id=$2`, cid, uid, at)
	return err
}

func (r *ChatRepo) GetDirectPair(ctx context.Context, a, b domain.UserID) (domain.ConversationID, bool, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT conversation_id FROM direct_pairs WHERE user_low=$1 AND user_high=$2`, a, b).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	return id, true, err
}

func (r *ChatRepo) PutDirectPair(ctx context.Context, a, b domain.UserID, cid domain.ConversationID) error {
	_, err := r.db.Exec(ctx, `INSERT INTO direct_pairs (user_low, user_high, conversation_id) VALUES ($1,$2,$3)`, a, b, cid)
	return err
}

func (r *ChatRepo) GetMessageRequest(ctx context.Context, cid domain.ConversationID) (domain.MessageRequest, bool, error) {
	var req domain.MessageRequest
	var status string
	err := r.db.QueryRow(ctx, `
		SELECT conversation_id, from_user_id, to_user_id, status, created_at, resolved_at
		FROM message_requests WHERE conversation_id=$1`, cid).
		Scan(&req.ConversationID, &req.FromUserID, &req.ToUserID, &status, &req.CreatedAt, &req.ResolvedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MessageRequest{}, false, nil
	}
	req.Status = domain.MessageRequestStatus(status)
	return req, true, err
}

func (r *ChatRepo) UpsertMessageRequest(ctx context.Context, req domain.MessageRequest) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO message_requests (conversation_id, from_user_id, to_user_id, status, created_at, resolved_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (conversation_id) DO UPDATE SET status=EXCLUDED.status, resolved_at=EXCLUDED.resolved_at`,
		req.ConversationID, req.FromUserID, req.ToUserID, string(req.Status), req.CreatedAt, req.ResolvedAt)
	return err
}

func (r *ChatRepo) IsBlocked(ctx context.Context, blocker, blocked domain.UserID) (bool, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT 1 FROM user_blocks WHERE blocker_id=$1 AND blocked_id=$2`, blocker, blocked).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *ChatRepo) Block(ctx context.Context, blocker, blocked domain.UserID) error {
	_, err := r.db.Exec(ctx, `INSERT INTO user_blocks (blocker_id, blocked_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, blocker, blocked)
	return err
}

func (r *ChatRepo) Unblock(ctx context.Context, blocker, blocked domain.UserID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_blocks WHERE blocker_id=$1 AND blocked_id=$2`, blocker, blocked)
	return err
}

func (r *ChatRepo) InsertMessage(ctx context.Context, m domain.Message) error {
	tx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		INSERT INTO messages (id, conversation_id, sender_id, body, edited_at, deleted_for_all_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		m.ID, m.ConversationID, m.SenderID, m.Body, m.EditedAt, m.DeletedForAllAt, m.CreatedAt)
	if err != nil {
		return err
	}
	for _, a := range m.Attachments {
		_, err = tx.Exec(ctx, `
			INSERT INTO message_attachments (id, message_id, blob_id, kind, filename)
			VALUES ($1,$2,$3,$4,$5)`, a.ID, m.ID, a.BlobID, string(a.Kind), a.Filename)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *ChatRepo) GetMessage(ctx context.Context, id domain.MessageID) (domain.Message, error) {
	var m domain.Message
	err := r.db.QueryRow(ctx, `
		SELECT id, conversation_id, sender_id, body, edited_at, deleted_for_all_at, created_at
		FROM messages WHERE id=$1`, id).
		Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Body, &m.EditedAt, &m.DeletedForAllAt, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Message{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Message{}, err
	}
	m.Attachments, _ = r.loadAttachments(ctx, id)
	m.Reactions, _ = r.loadReactions(ctx, id)
	return m, nil
}

func (r *ChatRepo) loadAttachments(ctx context.Context, mid domain.MessageID) ([]domain.MessageAttachment, error) {
	rows, err := r.db.Query(ctx, `SELECT id, blob_id, kind, filename FROM message_attachments WHERE message_id=$1`, mid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.MessageAttachment
	for rows.Next() {
		var a domain.MessageAttachment
		var kind string
		if err := rows.Scan(&a.ID, &a.BlobID, &kind, &a.Filename); err != nil {
			return nil, err
		}
		a.Kind = domain.AttachmentKind(kind)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *ChatRepo) loadReactions(ctx context.Context, mid domain.MessageID) ([]domain.MessageReaction, error) {
	rows, err := r.db.Query(ctx, `SELECT message_id, user_id, emoji, created_at FROM message_reactions WHERE message_id=$1`, mid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.MessageReaction
	for rows.Next() {
		var x domain.MessageReaction
		if err := rows.Scan(&x.MessageID, &x.UserID, &x.Emoji, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *ChatRepo) UpdateMessageBody(ctx context.Context, id domain.MessageID, body string, editedAt time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE messages SET body=$2, edited_at=$3 WHERE id=$1`, id, body, editedAt)
	return err
}

func (r *ChatRepo) SoftDeleteMessage(ctx context.Context, id domain.MessageID, at time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE messages SET body='', deleted_for_all_at=$2 WHERE id=$1`, id, at)
	return err
}

func (r *ChatRepo) HideMessage(ctx context.Context, userID domain.UserID, messageID domain.MessageID) error {
	_, err := r.db.Exec(ctx, `INSERT INTO message_hidden (user_id, message_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, userID, messageID)
	return err
}

func (r *ChatRepo) ListMessages(ctx context.Context, cid domain.ConversationID, viewer domain.UserID, limit int, _ *time.Time) ([]domain.Message, error) {
	rows, err := r.db.Query(ctx, `
		SELECT m.id, m.conversation_id, m.sender_id, m.body, m.edited_at, m.deleted_for_all_at, m.created_at
		FROM messages m
		LEFT JOIN message_hidden h ON h.message_id = m.id AND h.user_id = $2
		WHERE m.conversation_id = $1 AND h.message_id IS NULL
		ORDER BY m.created_at DESC
		LIMIT $3`, cid, viewer, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Body, &m.EditedAt, &m.DeletedForAllAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Attachments, _ = r.loadAttachments(ctx, m.ID)
		m.Reactions, _ = r.loadReactions(ctx, m.ID)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *ChatRepo) AddReaction(ctx context.Context, react domain.MessageReaction) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO message_reactions (message_id, user_id, emoji, created_at) VALUES ($1,$2,$3,$4)
		ON CONFLICT DO NOTHING`, react.MessageID, react.UserID, react.Emoji, react.CreatedAt)
	return err
}

func (r *ChatRepo) RemoveReaction(ctx context.Context, messageID domain.MessageID, userID domain.UserID, emoji string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM message_reactions WHERE message_id=$1 AND user_id=$2 AND emoji=$3`, messageID, userID, emoji)
	return err
}

func (r *ChatRepo) SetRead(ctx context.Context, cid domain.ConversationID, userID domain.UserID, lastMessageID domain.MessageID, at time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO message_reads (conversation_id, user_id, last_read_message_id, read_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (conversation_id, user_id) DO UPDATE SET last_read_message_id=EXCLUDED.last_read_message_id, read_at=EXCLUDED.read_at`,
		cid, userID, lastMessageID, at)
	return err
}

func (r *ChatRepo) GetRead(ctx context.Context, cid domain.ConversationID, userID domain.UserID) (domain.MessageID, bool, error) {
	var id *uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT last_read_message_id FROM message_reads WHERE conversation_id=$1 AND user_id=$2`, cid, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) || id == nil {
		return uuid.Nil, false, nil
	}
	return *id, true, err
}

func (r *ChatRepo) UpdateConversationAvatar(ctx context.Context, cid domain.ConversationID, blobID *domain.BlobID) error {
	_, err := r.db.Exec(ctx, `UPDATE conversations SET avatar_blob_id=$2 WHERE id=$1`, cid, blobID)
	return err
}

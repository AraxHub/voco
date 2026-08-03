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

const (
	ConversationTable = "conversations"

	ConversationColID           = "id"
	ConversationColType         = "type"
	ConversationColTitle        = "title"
	ConversationColAvatarBlobID = "avatar_blob_id"
	ConversationColCreatedBy    = "created_by"
	ConversationColCreatedAt    = "created_at"
)

func ConversationColumns() []string {
	return []string{
		ConversationColID,
		ConversationColType,
		ConversationColTitle,
		ConversationColAvatarBlobID,
		ConversationColCreatedBy,
		ConversationColCreatedAt,
	}
}

func ConversationSelect(alias string) string {
	return selectList(alias, ConversationColumns())
}

const (
	ConversationMemberTable = "conversation_members"

	ConversationMemberColConversationID = "conversation_id"
	ConversationMemberColUserID         = "user_id"
	ConversationMemberColRole           = "role"
	ConversationMemberColJoinedAt       = "joined_at"
	ConversationMemberColLeftAt         = "left_at"
)

func ConversationMemberColumns() []string {
	return []string{
		ConversationMemberColConversationID,
		ConversationMemberColUserID,
		ConversationMemberColRole,
		ConversationMemberColJoinedAt,
		ConversationMemberColLeftAt,
	}
}

func ConversationMemberSelect(alias string) string {
	return selectList(alias, ConversationMemberColumns())
}

func ConversationMemberInsertColumns() []string {
	return []string{
		ConversationMemberColConversationID,
		ConversationMemberColUserID,
		ConversationMemberColRole,
		ConversationMemberColJoinedAt,
	}
}

const (
	DirectPairTable = "direct_pairs"

	DirectPairColUserLow        = "user_low"
	DirectPairColUserHigh       = "user_high"
	DirectPairColConversationID = "conversation_id"
)

func DirectPairColumns() []string {
	return []string{
		DirectPairColUserLow,
		DirectPairColUserHigh,
		DirectPairColConversationID,
	}
}

func DirectPairSelect(alias string) string {
	return selectList(alias, DirectPairColumns())
}

const (
	MessageRequestTable = "message_requests"

	MessageRequestColConversationID = "conversation_id"
	MessageRequestColFromUserID     = "from_user_id"
	MessageRequestColToUserID       = "to_user_id"
	MessageRequestColStatus         = "status"
	MessageRequestColCreatedAt      = "created_at"
	MessageRequestColResolvedAt     = "resolved_at"
)

func MessageRequestColumns() []string {
	return []string{
		MessageRequestColConversationID,
		MessageRequestColFromUserID,
		MessageRequestColToUserID,
		MessageRequestColStatus,
		MessageRequestColCreatedAt,
		MessageRequestColResolvedAt,
	}
}

func MessageRequestSelect(alias string) string {
	return selectList(alias, MessageRequestColumns())
}

const (
	UserBlockTable = "user_blocks"

	UserBlockColBlockerID = "blocker_id"
	UserBlockColBlockedID = "blocked_id"
)

func UserBlockColumns() []string {
	return []string{
		UserBlockColBlockerID,
		UserBlockColBlockedID,
	}
}

func UserBlockSelect(alias string) string {
	return selectList(alias, UserBlockColumns())
}

const (
	MessageTable = "messages"

	MessageColID              = "id"
	MessageColConversationID  = "conversation_id"
	MessageColSenderID        = "sender_id"
	MessageColBody            = "body"
	MessageColEditedAt        = "edited_at"
	MessageColDeletedForAllAt = "deleted_for_all_at"
	MessageColCreatedAt       = "created_at"
)

func MessageColumns() []string {
	return []string{
		MessageColID,
		MessageColConversationID,
		MessageColSenderID,
		MessageColBody,
		MessageColEditedAt,
		MessageColDeletedForAllAt,
		MessageColCreatedAt,
	}
}

func MessageSelect(alias string) string {
	return selectList(alias, MessageColumns())
}

const (
	MessageAttachmentTable = "message_attachments"

	MessageAttachmentColID        = "id"
	MessageAttachmentColMessageID = "message_id"
	MessageAttachmentColBlobID    = "blob_id"
	MessageAttachmentColKind      = "kind"
	MessageAttachmentColFilename  = "filename"
)

func MessageAttachmentColumns() []string {
	return []string{
		MessageAttachmentColID,
		MessageAttachmentColMessageID,
		MessageAttachmentColBlobID,
		MessageAttachmentColKind,
		MessageAttachmentColFilename,
	}
}

func MessageAttachmentSelect(alias string) string {
	return selectList(alias, MessageAttachmentColumns())
}

func MessageAttachmentLoadColumns() []string {
	return []string{
		MessageAttachmentColID,
		MessageAttachmentColBlobID,
		MessageAttachmentColKind,
		MessageAttachmentColFilename,
	}
}

const (
	MessageReactionTable = "message_reactions"

	MessageReactionColMessageID = "message_id"
	MessageReactionColUserID    = "user_id"
	MessageReactionColEmoji     = "emoji"
	MessageReactionColCreatedAt = "created_at"
)

func MessageReactionColumns() []string {
	return []string{
		MessageReactionColMessageID,
		MessageReactionColUserID,
		MessageReactionColEmoji,
		MessageReactionColCreatedAt,
	}
}

func MessageReactionSelect(alias string) string {
	return selectList(alias, MessageReactionColumns())
}

const (
	MessageHiddenTable = "message_hidden"

	MessageHiddenColUserID    = "user_id"
	MessageHiddenColMessageID = "message_id"
)

func MessageHiddenColumns() []string {
	return []string{
		MessageHiddenColUserID,
		MessageHiddenColMessageID,
	}
}

func MessageHiddenSelect(alias string) string {
	return selectList(alias, MessageHiddenColumns())
}

const (
	MessageReadTable = "message_reads"

	MessageReadColConversationID    = "conversation_id"
	MessageReadColUserID            = "user_id"
	MessageReadColLastReadMessageID = "last_read_message_id"
	MessageReadColReadAt            = "read_at"
)

func MessageReadColumns() []string {
	return []string{
		MessageReadColConversationID,
		MessageReadColUserID,
		MessageReadColLastReadMessageID,
		MessageReadColReadAt,
	}
}

func MessageReadSelect(alias string) string {
	return selectList(alias, MessageReadColumns())
}

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
	cols := ConversationColumns()
	_, err = tx.Exec(ctx,
		"INSERT INTO "+ConversationTable+" ("+ConversationSelect("")+") VALUES ("+placeholders(len(cols))+")",
		c.ID, string(c.Type), c.Title, c.AvatarBlobID, c.CreatedBy, c.CreatedAt)
	if err != nil {
		return err
	}
	memberCols := ConversationMemberInsertColumns()
	for _, m := range members {
		_, err = tx.Exec(ctx,
			"INSERT INTO "+ConversationMemberTable+" ("+selectList("", memberCols)+") VALUES ("+placeholders(len(memberCols))+")",
			m.ConversationID, m.UserID, string(m.Role), m.JoinedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *ChatRepo) GetConversation(ctx context.Context, id domain.ConversationID) (domain.Conversation, error) {
	var c domain.Conversation
	var typ string
	err := r.db.QueryRow(ctx,
		"SELECT "+ConversationSelect("")+" FROM "+ConversationTable+" WHERE "+ConversationColID+" = $1", id).
		Scan(&c.ID, &typ, &c.Title, &c.AvatarBlobID, &c.CreatedBy, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Conversation{}, domain.ErrConversationNotFound
	}
	c.Type = domain.ConversationType(typ)
	return c, err
}

func (r *ChatRepo) ListConversationsForUser(ctx context.Context, userID domain.UserID) ([]domain.Conversation, error) {
	rows, err := r.db.Query(ctx,
		"SELECT "+ConversationSelect("c")+" FROM "+ConversationTable+" c"+
			" JOIN "+ConversationMemberTable+" m ON m."+ConversationMemberColConversationID+" = c."+ConversationColID+
			" WHERE m."+ConversationMemberColUserID+" = $1 AND m."+ConversationMemberColLeftAt+" IS NULL"+
			" ORDER BY c."+ConversationColCreatedAt+" DESC",
		userID)
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
	cols := ConversationMemberInsertColumns()
	_, err := r.db.Exec(ctx,
		"INSERT INTO "+ConversationMemberTable+" ("+selectList("", cols)+") VALUES ("+placeholders(len(cols))+")"+
			" ON CONFLICT ("+ConversationMemberColConversationID+", "+ConversationMemberColUserID+") DO UPDATE SET "+
			ConversationMemberColLeftAt+" = NULL, "+
			ConversationMemberColRole+" = EXCLUDED."+ConversationMemberColRole,
		m.ConversationID, m.UserID, string(m.Role), m.JoinedAt)
	return err
}

func (r *ChatRepo) GetMember(ctx context.Context, cid domain.ConversationID, uid domain.UserID) (domain.ConversationMember, bool, error) {
	var m domain.ConversationMember
	var role string
	err := r.db.QueryRow(ctx,
		"SELECT "+ConversationMemberSelect("")+" FROM "+ConversationMemberTable+
			" WHERE "+ConversationMemberColConversationID+" = $1 AND "+ConversationMemberColUserID+" = $2",
		cid, uid).
		Scan(&m.ConversationID, &m.UserID, &role, &m.JoinedAt, &m.LeftAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ConversationMember{}, false, nil
	}
	m.Role = domain.MemberRole(role)
	return m, true, err
}

func (r *ChatRepo) ListMembers(ctx context.Context, cid domain.ConversationID) ([]domain.ConversationMember, error) {
	rows, err := r.db.Query(ctx,
		"SELECT "+ConversationMemberSelect("")+" FROM "+ConversationMemberTable+
			" WHERE "+ConversationMemberColConversationID+" = $1", cid)
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
	tag, err := r.db.Exec(ctx,
		"UPDATE "+ConversationMemberTable+" SET "+ConversationMemberColRole+" = $3"+
			" WHERE "+ConversationMemberColConversationID+" = $1 AND "+ConversationMemberColUserID+" = $2",
		cid, uid, string(role))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ChatRepo) MarkLeft(ctx context.Context, cid domain.ConversationID, uid domain.UserID, at time.Time) error {
	_, err := r.db.Exec(ctx,
		"UPDATE "+ConversationMemberTable+" SET "+ConversationMemberColLeftAt+" = $3"+
			" WHERE "+ConversationMemberColConversationID+" = $1 AND "+ConversationMemberColUserID+" = $2",
		cid, uid, at)
	return err
}

func (r *ChatRepo) GetDirectPair(ctx context.Context, a, b domain.UserID) (domain.ConversationID, bool, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		"SELECT "+DirectPairColConversationID+" FROM "+DirectPairTable+
			" WHERE "+DirectPairColUserLow+" = $1 AND "+DirectPairColUserHigh+" = $2",
		a, b).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	return id, true, err
}

func (r *ChatRepo) PutDirectPair(ctx context.Context, a, b domain.UserID, cid domain.ConversationID) error {
	cols := DirectPairColumns()
	_, err := r.db.Exec(ctx,
		"INSERT INTO "+DirectPairTable+" ("+DirectPairSelect("")+") VALUES ("+placeholders(len(cols))+")",
		a, b, cid)
	return err
}

func (r *ChatRepo) GetMessageRequest(ctx context.Context, cid domain.ConversationID) (domain.MessageRequest, bool, error) {
	var req domain.MessageRequest
	var status string
	err := r.db.QueryRow(ctx,
		"SELECT "+MessageRequestSelect("")+" FROM "+MessageRequestTable+
			" WHERE "+MessageRequestColConversationID+" = $1", cid).
		Scan(&req.ConversationID, &req.FromUserID, &req.ToUserID, &status, &req.CreatedAt, &req.ResolvedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MessageRequest{}, false, nil
	}
	req.Status = domain.MessageRequestStatus(status)
	return req, true, err
}

func (r *ChatRepo) UpsertMessageRequest(ctx context.Context, req domain.MessageRequest) error {
	cols := MessageRequestColumns()
	_, err := r.db.Exec(ctx,
		"INSERT INTO "+MessageRequestTable+" ("+MessageRequestSelect("")+") VALUES ("+placeholders(len(cols))+")"+
			" ON CONFLICT ("+MessageRequestColConversationID+") DO UPDATE SET "+
			MessageRequestColStatus+" = EXCLUDED."+MessageRequestColStatus+", "+
			MessageRequestColResolvedAt+" = EXCLUDED."+MessageRequestColResolvedAt,
		req.ConversationID, req.FromUserID, req.ToUserID, string(req.Status), req.CreatedAt, req.ResolvedAt)
	return err
}

func (r *ChatRepo) IsBlocked(ctx context.Context, blocker, blocked domain.UserID) (bool, error) {
	var n int
	err := r.db.QueryRow(ctx,
		"SELECT 1 FROM "+UserBlockTable+
			" WHERE "+UserBlockColBlockerID+" = $1 AND "+UserBlockColBlockedID+" = $2",
		blocker, blocked).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *ChatRepo) Block(ctx context.Context, blocker, blocked domain.UserID) error {
	cols := UserBlockColumns()
	_, err := r.db.Exec(ctx,
		"INSERT INTO "+UserBlockTable+" ("+UserBlockSelect("")+") VALUES ("+placeholders(len(cols))+")"+
			" ON CONFLICT DO NOTHING",
		blocker, blocked)
	return err
}

func (r *ChatRepo) Unblock(ctx context.Context, blocker, blocked domain.UserID) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM "+UserBlockTable+
			" WHERE "+UserBlockColBlockerID+" = $1 AND "+UserBlockColBlockedID+" = $2",
		blocker, blocked)
	return err
}

func (r *ChatRepo) InsertMessage(ctx context.Context, m domain.Message) error {
	tx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	cols := MessageColumns()
	_, err = tx.Exec(ctx,
		"INSERT INTO "+MessageTable+" ("+MessageSelect("")+") VALUES ("+placeholders(len(cols))+")",
		m.ID, m.ConversationID, m.SenderID, m.Body, m.EditedAt, m.DeletedForAllAt, m.CreatedAt)
	if err != nil {
		return err
	}
	attCols := MessageAttachmentColumns()
	for _, a := range m.Attachments {
		_, err = tx.Exec(ctx,
			"INSERT INTO "+MessageAttachmentTable+" ("+MessageAttachmentSelect("")+") VALUES ("+placeholders(len(attCols))+")",
			a.ID, m.ID, a.BlobID, string(a.Kind), a.Filename)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *ChatRepo) GetMessage(ctx context.Context, id domain.MessageID) (domain.Message, error) {
	var m domain.Message
	err := r.db.QueryRow(ctx,
		"SELECT "+MessageSelect("")+" FROM "+MessageTable+" WHERE "+MessageColID+" = $1", id).
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
	cols := MessageAttachmentLoadColumns()
	rows, err := r.db.Query(ctx,
		"SELECT "+selectList("", cols)+" FROM "+MessageAttachmentTable+
			" WHERE "+MessageAttachmentColMessageID+" = $1", mid)
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
	rows, err := r.db.Query(ctx,
		"SELECT "+MessageReactionSelect("")+" FROM "+MessageReactionTable+
			" WHERE "+MessageReactionColMessageID+" = $1", mid)
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
	_, err := r.db.Exec(ctx,
		"UPDATE "+MessageTable+" SET "+MessageColBody+" = $2, "+MessageColEditedAt+" = $3 WHERE "+MessageColID+" = $1",
		id, body, editedAt)
	return err
}

func (r *ChatRepo) SoftDeleteMessage(ctx context.Context, id domain.MessageID, at time.Time) error {
	_, err := r.db.Exec(ctx,
		"UPDATE "+MessageTable+" SET "+MessageColBody+" = '', "+MessageColDeletedForAllAt+" = $2 WHERE "+MessageColID+" = $1",
		id, at)
	return err
}

func (r *ChatRepo) HideMessage(ctx context.Context, userID domain.UserID, messageID domain.MessageID) error {
	cols := MessageHiddenColumns()
	_, err := r.db.Exec(ctx,
		"INSERT INTO "+MessageHiddenTable+" ("+MessageHiddenSelect("")+") VALUES ("+placeholders(len(cols))+")"+
			" ON CONFLICT DO NOTHING",
		userID, messageID)
	return err
}

func (r *ChatRepo) ListMessages(ctx context.Context, cid domain.ConversationID, viewer domain.UserID, limit int, _ *time.Time) ([]domain.Message, error) {
	rows, err := r.db.Query(ctx,
		"SELECT "+MessageSelect("m")+" FROM "+MessageTable+" m"+
			" LEFT JOIN "+MessageHiddenTable+" h ON h."+MessageHiddenColMessageID+" = m."+MessageColID+
			" AND h."+MessageHiddenColUserID+" = $2"+
			" WHERE m."+MessageColConversationID+" = $1 AND h."+MessageHiddenColMessageID+" IS NULL"+
			" ORDER BY m."+MessageColCreatedAt+" DESC"+
			" LIMIT $3",
		cid, viewer, limit)
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
	cols := MessageReactionColumns()
	_, err := r.db.Exec(ctx,
		"INSERT INTO "+MessageReactionTable+" ("+MessageReactionSelect("")+") VALUES ("+placeholders(len(cols))+")"+
			" ON CONFLICT DO NOTHING",
		react.MessageID, react.UserID, react.Emoji, react.CreatedAt)
	return err
}

func (r *ChatRepo) RemoveReaction(ctx context.Context, messageID domain.MessageID, userID domain.UserID, emoji string) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM "+MessageReactionTable+
			" WHERE "+MessageReactionColMessageID+" = $1 AND "+MessageReactionColUserID+" = $2 AND "+MessageReactionColEmoji+" = $3",
		messageID, userID, emoji)
	return err
}

func (r *ChatRepo) SetRead(ctx context.Context, cid domain.ConversationID, userID domain.UserID, lastMessageID domain.MessageID, at time.Time) error {
	cols := MessageReadColumns()
	_, err := r.db.Exec(ctx,
		"INSERT INTO "+MessageReadTable+" ("+MessageReadSelect("")+") VALUES ("+placeholders(len(cols))+")"+
			" ON CONFLICT ("+MessageReadColConversationID+", "+MessageReadColUserID+") DO UPDATE SET "+
			MessageReadColLastReadMessageID+" = EXCLUDED."+MessageReadColLastReadMessageID+", "+
			MessageReadColReadAt+" = EXCLUDED."+MessageReadColReadAt,
		cid, userID, lastMessageID, at)
	return err
}

func (r *ChatRepo) GetRead(ctx context.Context, cid domain.ConversationID, userID domain.UserID) (domain.MessageID, bool, error) {
	var id *uuid.UUID
	err := r.db.QueryRow(ctx,
		"SELECT "+MessageReadColLastReadMessageID+" FROM "+MessageReadTable+
			" WHERE "+MessageReadColConversationID+" = $1 AND "+MessageReadColUserID+" = $2",
		cid, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) || id == nil {
		return uuid.Nil, false, nil
	}
	return *id, true, err
}

func (r *ChatRepo) CountUnread(ctx context.Context, cid domain.ConversationID, userID domain.UserID) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `
SELECT COUNT(*) FROM `+MessageTable+` m
LEFT JOIN `+MessageHiddenTable+` h ON h.`+MessageHiddenColMessageID+` = m.`+MessageColID+` AND h.`+MessageHiddenColUserID+` = $2
WHERE m.`+MessageColConversationID+` = $1
  AND h.`+MessageHiddenColMessageID+` IS NULL
  AND m.`+MessageColSenderID+` <> $2
  AND m.`+MessageColDeletedForAllAt+` IS NULL
  AND (
    NOT EXISTS (
      SELECT 1 FROM `+MessageReadTable+` r
      WHERE r.`+MessageReadColConversationID+` = $1 AND r.`+MessageReadColUserID+` = $2
        AND r.`+MessageReadColLastReadMessageID+` IS NOT NULL
    )
    OR m.`+MessageColCreatedAt+` > (
      SELECT lm.`+MessageColCreatedAt+` FROM `+MessageTable+` lm
      JOIN `+MessageReadTable+` r ON r.`+MessageReadColLastReadMessageID+` = lm.`+MessageColID+`
      WHERE r.`+MessageReadColConversationID+` = $1 AND r.`+MessageReadColUserID+` = $2
    )
  )`, cid, userID).Scan(&n)
	return n, err
}

func (r *ChatRepo) UpdateConversationAvatar(ctx context.Context, cid domain.ConversationID, blobID *domain.BlobID) error {
	_, err := r.db.Exec(ctx,
		"UPDATE "+ConversationTable+" SET "+ConversationColAvatarBlobID+" = $2 WHERE "+ConversationColID+" = $1",
		cid, blobID)
	return err
}

func (r *ChatRepo) UpdateConversationTitle(ctx context.Context, cid domain.ConversationID, title string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE "+ConversationTable+" SET "+ConversationColTitle+" = $2 WHERE "+ConversationColID+" = $1",
		cid, title)
	return err
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

type ConversationID = uuid.UUID
type MessageID = uuid.UUID

type ConversationType string

const (
	ConversationDirect ConversationType = "direct"
	ConversationGroup  ConversationType = "group"
)

type MemberRole string

const (
	RoleMember MemberRole = "member"
	RoleAdmin  MemberRole = "admin"
)

type MessageRequestStatus string

const (
	MessageRequestPending  MessageRequestStatus = "pending"
	MessageRequestAccepted MessageRequestStatus = "accepted"
	MessageRequestBlocked  MessageRequestStatus = "blocked"
)

type AttachmentKind string

const (
	AttachmentImage AttachmentKind = "image"
	AttachmentFile  AttachmentKind = "file"
)

type Conversation struct {
	ID           ConversationID
	Type         ConversationType
	Title        string
	AvatarBlobID *BlobID
	CreatedBy    UserID
	CreatedAt    time.Time
}

type ConversationMember struct {
	ConversationID ConversationID
	UserID         UserID
	Role           MemberRole
	JoinedAt       time.Time
	LeftAt         *time.Time
}

type MessageRequest struct {
	ConversationID ConversationID
	FromUserID     UserID
	ToUserID       UserID
	Status         MessageRequestStatus
	CreatedAt      time.Time
	ResolvedAt     *time.Time
}

type Message struct {
	ID               MessageID
	ConversationID   ConversationID
	SenderID         UserID
	Body             string
	EditedAt         *time.Time
	DeletedForAllAt  *time.Time
	CreatedAt        time.Time
	Attachments      []MessageAttachment
	Reactions        []MessageReaction
}

type MessageAttachment struct {
	ID       uuid.UUID
	BlobID   BlobID
	Kind     AttachmentKind
	Filename string
}

type MessageReaction struct {
	MessageID MessageID
	UserID    UserID
	Emoji     string
	CreatedAt time.Time
}

type DeleteMode string

const (
	DeleteForAll DeleteMode = "for_all"
	DeleteForMe  DeleteMode = "for_me"
)

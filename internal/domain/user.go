package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserID = uuid.UUID
type BlobID = uuid.UUID

type User struct {
	ID           UserID
	KeycloakSub  string
	Nickname     string
	Email        string
	DisplayName  string
	AvatarBlobID *BlobID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastSeenAt   time.Time
}

type Blob struct {
	ID          BlobID
	OwnerUserID *UserID
	ContentType string
	ByteSize    int64
	SHA256      string
	Data        []byte
	CreatedAt   time.Time
}

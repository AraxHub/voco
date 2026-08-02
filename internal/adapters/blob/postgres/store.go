package blobpg

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	pgadapter "voco/internal/adapters/postgres"
	"voco/internal/domain"

	"github.com/google/uuid"
)

type Store struct {
	db *pgadapter.Client
}

func New(db *pgadapter.Client) *Store {
	return &Store{db: db}
}

func (s *Store) Put(ctx context.Context, blob domain.Blob) (domain.Blob, error) {
	if blob.ID == uuid.Nil {
		blob.ID = uuid.New()
	}
	if blob.CreatedAt.IsZero() {
		blob.CreatedAt = time.Now().UTC()
	}
	sum := sha256.Sum256(blob.Data)
	blob.SHA256 = hex.EncodeToString(sum[:])
	blob.ByteSize = int64(len(blob.Data))

	_, err := s.db.Exec(ctx, `
		INSERT INTO blobs (id, owner_user_id, content_type, byte_size, sha256, data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, blob.ID, blob.OwnerUserID, blob.ContentType, blob.ByteSize, blob.SHA256, blob.Data, blob.CreatedAt)
	if err != nil {
		return domain.Blob{}, fmt.Errorf("blob put: %w", err)
	}
	return blob, nil
}

func (s *Store) Get(ctx context.Context, id domain.BlobID) (domain.Blob, error) {
	var b domain.Blob
	err := s.db.QueryRow(ctx, `
		SELECT id, owner_user_id, content_type, byte_size, sha256, data, created_at
		FROM blobs WHERE id = $1
	`, id).Scan(&b.ID, &b.OwnerUserID, &b.ContentType, &b.ByteSize, &b.SHA256, &b.Data, &b.CreatedAt)
	if err != nil {
		return domain.Blob{}, domain.ErrNotFound
	}
	return b, nil
}

func (s *Store) Delete(ctx context.Context, id domain.BlobID) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM blobs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

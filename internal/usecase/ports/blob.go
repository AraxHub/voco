package ports

import (
	"context"

	"voco/internal/domain"
)

// BlobStore abstracts attachment bytes. PG BYTEA now; S3 later without usecase changes.
type BlobStore interface {
	Put(ctx context.Context, blob domain.Blob) (domain.Blob, error)
	Get(ctx context.Context, id domain.BlobID) (domain.Blob, error)
	Delete(ctx context.Context, id domain.BlobID) error
}

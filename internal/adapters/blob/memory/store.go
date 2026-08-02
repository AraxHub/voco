package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"voco/internal/domain"

	"github.com/google/uuid"
)

type BlobStore struct {
	mu   sync.RWMutex
	data map[uuid.UUID]domain.Blob
}

func NewBlobStore() *BlobStore {
	return &BlobStore{data: map[uuid.UUID]domain.Blob{}}
}

func (s *BlobStore) Put(_ context.Context, blob domain.Blob) (domain.Blob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if blob.ID == uuid.Nil {
		blob.ID = uuid.New()
	}
	if blob.CreatedAt.IsZero() {
		blob.CreatedAt = time.Now().UTC()
	}
	sum := sha256.Sum256(blob.Data)
	blob.SHA256 = hex.EncodeToString(sum[:])
	blob.ByteSize = int64(len(blob.Data))
	s.data[blob.ID] = blob
	return blob, nil
}

func (s *BlobStore) Get(_ context.Context, id domain.BlobID) (domain.Blob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.data[id]
	if !ok {
		return domain.Blob{}, domain.ErrNotFound
	}
	return b, nil
}

func (s *BlobStore) Delete(_ context.Context, id domain.BlobID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(s.data, id)
	return nil
}

package users

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"voco/internal/domain"

	"github.com/google/uuid"
)

type Store interface {
	UpsertByKeycloakSub(ctx context.Context, sub, email, displayName, nickname string) (domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (domain.User, error)
	GetByKeycloakSub(ctx context.Context, sub string) (domain.User, bool, error)
	UpdateProfile(ctx context.Context, id domain.UserID, nickname, displayName string) (domain.User, error)
	SearchByNickname(ctx context.Context, query string, limit int) ([]domain.User, error)
	TouchLastSeen(ctx context.Context, id domain.UserID) error
	ListAll(ctx context.Context) ([]domain.User, error)
	UpsertSynced(ctx context.Context, sub, email, displayName, nickname string) (domain.User, error)
}

type DirectoryUser struct {
	Sub         string
	Email       string
	DisplayName string
	Username    string
}

type Directory interface {
	ListUsers(ctx context.Context) ([]DirectoryUser, error)
}

type Usecase struct {
	store Store
	dir   Directory
}

func New(store Store, dir Directory) *Usecase {
	return &Usecase{store: store, dir: dir}
}

func (uc *Usecase) EnsureFromAuth(ctx context.Context, keycloakSub, email, name, username string) (domain.User, error) {
	if strings.TrimSpace(keycloakSub) == "" {
		return domain.User{}, domain.ErrValidation
	}
	nick := strings.TrimSpace(username)
	if nick == "" {
		nick = strings.TrimSpace(email)
	}
	u, err := uc.store.UpsertByKeycloakSub(ctx, keycloakSub, email, name, nick)
	if err != nil {
		return domain.User{}, err
	}
	_ = uc.store.TouchLastSeen(ctx, u.ID)
	return u, nil
}

func (uc *Usecase) Me(ctx context.Context, id domain.UserID) (domain.User, error) {
	return uc.store.GetByID(ctx, id)
}

func (uc *Usecase) UpdateMe(ctx context.Context, id domain.UserID, nickname, displayName string) (domain.User, error) {
	nickname = strings.TrimSpace(nickname)
	displayName = strings.TrimSpace(displayName)
	// Nickname == login: if client omits it, keep existing and only refresh display name.
	if nickname == "" {
		u, err := uc.store.GetByID(ctx, id)
		if err != nil {
			return domain.User{}, err
		}
		nickname = u.Nickname
	}
	if nickname == "" {
		return domain.User{}, domain.ErrValidation
	}
	if utf8.RuneCountInString(nickname) > 64 {
		return domain.User{}, domain.ErrValidation
	}
	return uc.store.UpdateProfile(ctx, id, nickname, displayName)
}

func (uc *Usecase) Search(ctx context.Context, q string, limit int) ([]domain.User, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return uc.store.SearchByNickname(ctx, q, limit)
}

func (uc *Usecase) SyncFromKeycloak(ctx context.Context) (int, error) {
	if uc.dir == nil {
		return 0, nil
	}
	list, err := uc.dir.ListUsers(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, d := range list {
		nick := strings.TrimSpace(d.Username)
		if nick == "" {
			nick = strings.TrimSpace(d.Email)
		}
		if _, err := uc.store.UpsertSynced(ctx, d.Sub, d.Email, d.DisplayName, nick); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (uc *Usecase) StartSyncLoop(ctx context.Context, every time.Duration) {
	if every <= 0 {
		every = 15 * time.Minute
	}
	go func() {
		t := time.NewTicker(every)
		defer t.Stop()
		_, _ = uc.SyncFromKeycloak(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				_, _ = uc.SyncFromKeycloak(ctx)
			}
		}
	}()
}

func NewUserID() domain.UserID { return uuid.New() }

package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	pgadapter "voco/internal/adapters/postgres"
	"voco/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepo struct {
	db *pgadapter.Client
}

func NewUserRepo(db *pgadapter.Client) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) UpsertByKeycloakSub(ctx context.Context, sub, email, displayName string) (domain.User, error) {
	now := time.Now().UTC()
	var u domain.User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (id, keycloak_sub, nickname, email, display_name, created_at, updated_at, last_seen_at)
		VALUES ($1, $2, '', $3, $4, $5, $5, $5)
		ON CONFLICT (keycloak_sub) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = CASE WHEN EXCLUDED.display_name <> '' THEN EXCLUDED.display_name ELSE users.display_name END,
			updated_at = EXCLUDED.updated_at,
			last_seen_at = EXCLUDED.last_seen_at
		RETURNING id, keycloak_sub, nickname, email, display_name, avatar_blob_id, created_at, updated_at, last_seen_at
	`, uuid.New(), sub, email, displayName, now).Scan(
		&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
		&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("upsert user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx, `
		SELECT id, keycloak_sub, nickname, email, display_name, avatar_blob_id, created_at, updated_at, last_seen_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
		&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, err
}

func (r *UserRepo) GetByKeycloakSub(ctx context.Context, sub string) (domain.User, bool, error) {
	var u domain.User
	err := r.db.QueryRow(ctx, `
		SELECT id, keycloak_sub, nickname, email, display_name, avatar_blob_id, created_at, updated_at, last_seen_at
		FROM users WHERE keycloak_sub = $1
	`, sub).Scan(&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
		&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, false, nil
	}
	if err != nil {
		return domain.User{}, false, err
	}
	return u, true, nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id domain.UserID, nickname, displayName string) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx, `
		UPDATE users SET nickname = $2,
			display_name = CASE WHEN $3 <> '' THEN $3 ELSE display_name END,
			updated_at = now()
		WHERE id = $1
		RETURNING id, keycloak_sub, nickname, email, display_name, avatar_blob_id, created_at, updated_at, last_seen_at
	`, id, nickname, displayName).Scan(&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
		&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.User{}, domain.ErrNicknameTaken
	}
	return u, err
}

func (r *UserRepo) SearchByNickname(ctx context.Context, query string, limit int) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, keycloak_sub, nickname, email, display_name, avatar_blob_id, created_at, updated_at, last_seen_at
		FROM users
		WHERE nickname <> '' AND lower(nickname) LIKE lower($1)
		ORDER BY nickname
		LIMIT $2
	`, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
			&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserRepo) TouchLastSeen(ctx context.Context, id domain.UserID) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET last_seen_at = now() WHERE id = $1`, id)
	return err
}

func (r *UserRepo) ListAll(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, keycloak_sub, nickname, email, display_name, avatar_blob_id, created_at, updated_at, last_seen_at FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
			&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserRepo) UpsertSynced(ctx context.Context, sub, email, displayName, nickname string) (domain.User, error) {
	u, err := r.UpsertByKeycloakSub(ctx, sub, email, displayName)
	if err != nil {
		return domain.User{}, err
	}
	if u.Nickname == "" && strings.TrimSpace(nickname) != "" {
		updated, err := r.UpdateProfile(ctx, u.ID, nickname, displayName)
		if err == domain.ErrNicknameTaken {
			return u, nil
		}
		return updated, err
	}
	return u, nil
}

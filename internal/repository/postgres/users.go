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

const (
	UserTable = "users"

	UserColID           = "id"
	UserColKeycloakSub  = "keycloak_sub"
	UserColNickname     = "nickname"
	UserColEmail        = "email"
	UserColDisplayName  = "display_name"
	UserColAvatarBlobID = "avatar_blob_id"
	UserColCreatedAt    = "created_at"
	UserColUpdatedAt    = "updated_at"
	UserColLastSeenAt   = "last_seen_at"
)

func UserColumns() []string {
	return []string{
		UserColID,
		UserColKeycloakSub,
		UserColNickname,
		UserColEmail,
		UserColDisplayName,
		UserColAvatarBlobID,
		UserColCreatedAt,
		UserColUpdatedAt,
		UserColLastSeenAt,
	}
}

func UserSelect(alias string) string {
	return selectList(alias, UserColumns())
}

type UserRepo struct {
	db *pgadapter.Client
}

func NewUserRepo(db *pgadapter.Client) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) UpsertByKeycloakSub(ctx context.Context, sub, email, displayName, nickname string) (domain.User, error) {
	now := time.Now().UTC()
	nick := strings.TrimSpace(nickname)
	candidates := uniqueNicknameCandidates(nick)

	var lastErr error
	for _, candidate := range candidates {
		u, err := r.upsertByKeycloakSubOnce(ctx, sub, email, displayName, candidate, now)
		if err == nil {
			return u, nil
		}
		lastErr = err
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// nickname (or rare race) taken — try next candidate
			continue
		}
		return domain.User{}, err
	}
	if lastErr != nil {
		return domain.User{}, fmt.Errorf("upsert user: %w", lastErr)
	}
	return domain.User{}, fmt.Errorf("upsert user: nickname unavailable")
}

func uniqueNicknameCandidates(base string) []string {
	base = strings.TrimSpace(base)
	out := make([]string, 0, 8)
	if base != "" {
		out = append(out, base)
		for i := 2; i <= 6; i++ {
			out = append(out, fmt.Sprintf("%s_%d", base, i))
		}
		out = append(out, fmt.Sprintf("%s_%s", truncateRunes(base, 48), uuid.New().String()[:8]))
	} else {
		out = append(out, "user_"+uuid.New().String()[:8])
	}
	return out
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	n := 0
	for i := range s {
		if n == max {
			return s[:i]
		}
		n++
	}
	return s
}

func (r *UserRepo) upsertByKeycloakSubOnce(ctx context.Context, sub, email, displayName, nickname string, now time.Time) (domain.User, error) {
	var u domain.User
	insertCols := []string{
		UserColID, UserColKeycloakSub, UserColNickname, UserColEmail, UserColDisplayName,
		UserColCreatedAt, UserColUpdatedAt, UserColLastSeenAt,
	}
	q := "INSERT INTO " + UserTable + " (" + selectList("", insertCols) + ")" +
		" VALUES ($1, $2, $3, $4, $5, $6, $6, $6)" +
		" ON CONFLICT (" + UserColKeycloakSub + ") DO UPDATE SET " +
		UserColEmail + " = EXCLUDED." + UserColEmail + ", " +
		UserColDisplayName + " = CASE WHEN EXCLUDED." + UserColDisplayName + " <> '' THEN EXCLUDED." + UserColDisplayName +
		" ELSE " + UserTable + "." + UserColDisplayName + " END, " +
		UserColNickname + " = CASE" +
		" WHEN EXCLUDED." + UserColNickname + " <> '' AND (" + UserTable + "." + UserColNickname + " = '' OR " +
		UserTable + "." + UserColNickname + " IS DISTINCT FROM EXCLUDED." + UserColNickname + ")" +
		" THEN EXCLUDED." + UserColNickname +
		" ELSE " + UserTable + "." + UserColNickname + " END, " +
		UserColUpdatedAt + " = EXCLUDED." + UserColUpdatedAt + ", " +
		UserColLastSeenAt + " = EXCLUDED." + UserColLastSeenAt +
		" RETURNING " + UserSelect("")
	err := r.db.QueryRow(ctx, q, uuid.New(), sub, nickname, email, displayName, now).Scan(
		&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
		&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt,
	)
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx,
		"SELECT "+UserSelect("")+" FROM "+UserTable+" WHERE "+UserColID+" = $1",
		id,
	).Scan(&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
		&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, err
}

func (r *UserRepo) GetByKeycloakSub(ctx context.Context, sub string) (domain.User, bool, error) {
	var u domain.User
	err := r.db.QueryRow(ctx,
		"SELECT "+UserSelect("")+" FROM "+UserTable+" WHERE "+UserColKeycloakSub+" = $1",
		sub,
	).Scan(&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
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
	q := "UPDATE " + UserTable + " SET " + UserColNickname + " = $2, " +
		UserColDisplayName + " = CASE WHEN $3 <> '' THEN $3 ELSE " + UserColDisplayName + " END, " +
		UserColUpdatedAt + " = now()" +
		" WHERE " + UserColID + " = $1" +
		" RETURNING " + UserSelect("")
	err := r.db.QueryRow(ctx, q, id, nickname, displayName).Scan(
		&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
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

func (r *UserRepo) UpdateAvatar(ctx context.Context, id domain.UserID, avatarBlobID *domain.BlobID) (domain.User, error) {
	var u domain.User
	q := "UPDATE " + UserTable + " SET " + UserColAvatarBlobID + " = $2, " +
		UserColUpdatedAt + " = now()" +
		" WHERE " + UserColID + " = $1" +
		" RETURNING " + UserSelect("")
	err := r.db.QueryRow(ctx, q, id, avatarBlobID).Scan(
		&u.ID, &u.KeycloakSub, &u.Nickname, &u.Email, &u.DisplayName, &u.AvatarBlobID,
		&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, err
}

func (r *UserRepo) SearchByNickname(ctx context.Context, query string, limit int) ([]domain.User, error) {
	rows, err := r.db.Query(ctx,
		"SELECT "+UserSelect("")+" FROM "+UserTable+
			" WHERE "+UserColNickname+" <> '' AND lower("+UserColNickname+") LIKE lower($1)"+
			" ORDER BY "+UserColNickname+
			" LIMIT $2",
		"%"+query+"%", limit)
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
	_, err := r.db.Exec(ctx,
		"UPDATE "+UserTable+" SET "+UserColLastSeenAt+" = now() WHERE "+UserColID+" = $1", id)
	return err
}

func (r *UserRepo) ListAll(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, "SELECT "+UserSelect("")+" FROM "+UserTable)
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
	return r.UpsertByKeycloakSub(ctx, sub, email, displayName, nickname)
}

package postgres

import (
	"context"

	pgadapter "voco/internal/adapters/postgres"
	"voco/internal/domain"

	"github.com/google/uuid"
)

type PushRepo struct {
	db *pgadapter.Client
}

func NewPushRepo(db *pgadapter.Client) *PushRepo {
	return &PushRepo{db: db}
}

func (r *PushRepo) SetPushEnabled(ctx context.Context, userID domain.UserID, enabled bool) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO notification_settings (user_id, push_enabled) VALUES ($1,$2)
		ON CONFLICT (user_id) DO UPDATE SET push_enabled = EXCLUDED.push_enabled`, userID, enabled)
	return err
}

func (r *PushRepo) GetPushEnabled(ctx context.Context, userID domain.UserID) (bool, error) {
	var enabled bool
	err := r.db.QueryRow(ctx, `SELECT push_enabled FROM notification_settings WHERE user_id=$1`, userID).Scan(&enabled)
	if err != nil {
		return false, nil
	}
	return enabled, nil
}

func (r *PushRepo) UpsertSubscription(ctx context.Context, userID domain.UserID, endpoint, p256dh, auth string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh, auth, enabled)
		VALUES ($1,$2,$3,$4,$5,true)
		ON CONFLICT (endpoint) DO UPDATE SET user_id=EXCLUDED.user_id, p256dh=EXCLUDED.p256dh, auth=EXCLUDED.auth, enabled=true`,
		uuid.New(), userID, endpoint, p256dh, auth)
	return err
}

func (r *PushRepo) DeleteSubscription(ctx context.Context, userID domain.UserID, endpoint string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM push_subscriptions WHERE user_id=$1 AND endpoint=$2`, userID, endpoint)
	return err
}

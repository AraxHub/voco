package postgres

import (
	"context"

	pgadapter "voco/internal/adapters/postgres"
	"voco/internal/domain"

	"github.com/google/uuid"
)

const (
	NotificationSettingsTable = "notification_settings"

	NotificationSettingsColUserID      = "user_id"
	NotificationSettingsColPushEnabled = "push_enabled"
)

func NotificationSettingsColumns() []string {
	return []string{
		NotificationSettingsColUserID,
		NotificationSettingsColPushEnabled,
	}
}

func NotificationSettingsSelect(alias string) string {
	return selectList(alias, NotificationSettingsColumns())
}

const (
	PushSubscriptionTable = "push_subscriptions"

	PushSubscriptionColID       = "id"
	PushSubscriptionColUserID   = "user_id"
	PushSubscriptionColEndpoint = "endpoint"
	PushSubscriptionColP256dh   = "p256dh"
	PushSubscriptionColAuth     = "auth"
	PushSubscriptionColEnabled  = "enabled"
)

func PushSubscriptionColumns() []string {
	return []string{
		PushSubscriptionColID,
		PushSubscriptionColUserID,
		PushSubscriptionColEndpoint,
		PushSubscriptionColP256dh,
		PushSubscriptionColAuth,
		PushSubscriptionColEnabled,
	}
}

func PushSubscriptionSelect(alias string) string {
	return selectList(alias, PushSubscriptionColumns())
}

type PushRepo struct {
	db *pgadapter.Client
}

func NewPushRepo(db *pgadapter.Client) *PushRepo {
	return &PushRepo{db: db}
}

func (r *PushRepo) SetPushEnabled(ctx context.Context, userID domain.UserID, enabled bool) error {
	cols := []string{NotificationSettingsColUserID, NotificationSettingsColPushEnabled}
	_, err := r.db.Exec(ctx,
		"INSERT INTO "+NotificationSettingsTable+" ("+selectList("", cols)+") VALUES ("+placeholders(len(cols))+")"+
			" ON CONFLICT ("+NotificationSettingsColUserID+") DO UPDATE SET "+
			NotificationSettingsColPushEnabled+" = EXCLUDED."+NotificationSettingsColPushEnabled,
		userID, enabled)
	return err
}

func (r *PushRepo) GetPushEnabled(ctx context.Context, userID domain.UserID) (bool, error) {
	var enabled bool
	err := r.db.QueryRow(ctx,
		"SELECT "+NotificationSettingsColPushEnabled+" FROM "+NotificationSettingsTable+
			" WHERE "+NotificationSettingsColUserID+" = $1",
		userID).Scan(&enabled)
	if err != nil {
		return false, nil
	}
	return enabled, nil
}

func (r *PushRepo) UpsertSubscription(ctx context.Context, userID domain.UserID, endpoint, p256dh, auth string) error {
	cols := []string{
		PushSubscriptionColID, PushSubscriptionColUserID, PushSubscriptionColEndpoint,
		PushSubscriptionColP256dh, PushSubscriptionColAuth, PushSubscriptionColEnabled,
	}
	_, err := r.db.Exec(ctx,
		"INSERT INTO "+PushSubscriptionTable+" ("+selectList("", cols)+")"+
			" VALUES ($1, $2, $3, $4, $5, true)"+
			" ON CONFLICT ("+PushSubscriptionColEndpoint+") DO UPDATE SET "+
			PushSubscriptionColUserID+" = EXCLUDED."+PushSubscriptionColUserID+", "+
			PushSubscriptionColP256dh+" = EXCLUDED."+PushSubscriptionColP256dh+", "+
			PushSubscriptionColAuth+" = EXCLUDED."+PushSubscriptionColAuth+", "+
			PushSubscriptionColEnabled+" = true",
		uuid.New(), userID, endpoint, p256dh, auth)
	return err
}

func (r *PushRepo) DeleteSubscription(ctx context.Context, userID domain.UserID, endpoint string) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM "+PushSubscriptionTable+
			" WHERE "+PushSubscriptionColUserID+" = $1 AND "+PushSubscriptionColEndpoint+" = $2",
		userID, endpoint)
	return err
}

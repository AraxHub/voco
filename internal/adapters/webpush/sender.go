package webpush

import (
	"context"
	"encoding/json"

	pgadapter "voco/internal/adapters/postgres"
	"voco/internal/domain"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type Config struct {
	VAPIDPublic  string `envconfig:"VAPID_PUBLIC"`
	VAPIDPrivate string `envconfig:"VAPID_PRIVATE"`
	Subscriber   string `envconfig:"VAPID_SUBJECT" default:"mailto:admin@voco-online.ru"`
}

type Sender struct {
	db  *pgadapter.Client
	cfg Config
}

func NewSender(db *pgadapter.Client, cfg Config) *Sender {
	return &Sender{db: db, cfg: cfg}
}

func (s *Sender) Send(ctx context.Context, userID domain.UserID, title, body string) error {
	if s.cfg.VAPIDPublic == "" || s.cfg.VAPIDPrivate == "" {
		return nil
	}
	rows, err := s.db.Query(ctx, `
		SELECT ps.endpoint, ps.p256dh, ps.auth
		FROM push_subscriptions ps
		JOIN notification_settings ns ON ns.user_id = ps.user_id
		WHERE ps.user_id = $1 AND ps.enabled AND ns.push_enabled`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	payload, _ := json.Marshal(map[string]string{"title": title, "body": body})
	for rows.Next() {
		var endpoint, p256dh, auth string
		if err := rows.Scan(&endpoint, &p256dh, &auth); err != nil {
			continue
		}
		sub := &webpush.Subscription{
			Endpoint: endpoint,
			Keys:     webpush.Keys{P256dh: p256dh, Auth: auth},
		}
		resp, err := webpush.SendNotificationWithContext(ctx, payload, sub, &webpush.Options{
			Subscriber:      s.cfg.Subscriber,
			VAPIDPublicKey:  s.cfg.VAPIDPublic,
			VAPIDPrivateKey: s.cfg.VAPIDPrivate,
			TTL:             60,
		})
		if err == nil && resp != nil {
			_ = resp.Body.Close()
		}
	}
	return nil
}

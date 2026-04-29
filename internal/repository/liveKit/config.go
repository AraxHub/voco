package liveKit

import "time"

type Cfg struct {
	LiveKitURL       string        `envconfig:"URL"`
	LiveKitAPIKey    string        `envconfig:"API_KEY" default:"devkey"`
	LiveKitAPISecret string        `envconfig:"API_SECRET" default:"devsecret"`
	TokenTTL         time.Duration `envconfig:"TOKEN_TTL" default:"6h"`
}

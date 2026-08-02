package app

import (
	"log"
	"os"
	"time"

	"voco/internal/adapters/keycloak"
	"voco/internal/adapters/livekit"
	"voco/internal/adapters/postgres"
	"voco/internal/adapters/webpush"
	"voco/internal/api/http"
	"voco/internal/pkg/auth"
	"voco/internal/pkg/logger"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const AppName = "VOCO"

type FeaturesConfig struct {
	UserSyncInterval time.Duration `envconfig:"USER_SYNC_INTERVAL" default:"15m"`
	MaxImageBytes    int64         `envconfig:"MAX_IMAGE_BYTES" default:"10485760"`
	MaxFileBytes     int64         `envconfig:"MAX_FILE_BYTES" default:"26214400"`
	MaxGroupMembers  int           `envconfig:"MAX_GROUP_MEMBERS" default:"100"`
	MaxMessageLen    int           `envconfig:"MAX_MESSAGE_LEN" default:"4000"`
}

type Config struct {
	Server        http.ServerConfig `envconfig:"server"`
	LiveKit       livekit.Cfg       `envconfig:"livekit"`
	Pg            postgres.Config   `envconfig:"pg"`
	Keycloak      auth.Config       `envconfig:"keycloak"`
	KeycloakAdmin keycloak.AdminConfig `envconfig:"keycloak_admin"`
	WebPush       webpush.Config    `envconfig:"webpush"`
	Features      FeaturesConfig    `envconfig:"features"`
	Log           logger.Config     `envconfig:"log"`
	DevMode       bool              `envconfig:"dev_mode"`
}

func MustLoadCfg(ReleaseMode string) (Config, error) {
	if ReleaseMode == "dev" {
		paths := []string{
			os.Getenv("VOCO_DOTENV_PATH"),
			"/app/.env",
			"deployment/voco-local/.env",
			".env",
		}
		for _, p := range paths {
			if p == "" {
				continue
			}
			if _, err := os.Stat(p); err == nil {
				if err := godotenv.Load(p); err != nil {
					return Config{}, err
				}
				log.Printf("config: loaded dotenv from %s", p)
				break
			}
		}
	}

	var cfg Config
	if err := envconfig.Process(AppName, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

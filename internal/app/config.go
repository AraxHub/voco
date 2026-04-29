package app

import (
	"log"
	"os"
	"voco/internal/api/http"
	"voco/internal/pkg/logger"
	"voco/internal/repository/cache/inmemory"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"voco/internal/repository/liveKit"
)

const AppName = "VOCO"

type Config struct {
	Server  http.ServerConfig `envconfig:"server"`
	Cache   inmemory.Config   `envconfig:"cache"`
	LiveKit liveKit.Cfg       `envconfig:"livekit"`
	Log     logger.Config     `envconfig:"log"`
	DevMode bool              `envconfig:"dev_mode"`
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

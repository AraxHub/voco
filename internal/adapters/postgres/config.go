package postgres

import (
	"fmt"
	"net/url"
	"time"
)

type Config struct {
	// URL, if set, overrides Host/Port/User/Password/Database/Schema/SSLMode.
	URL string `envconfig:"URL"`

	Host     string `envconfig:"HOST" default:"localhost"`
	Port     int    `envconfig:"PORT" default:"5432"`
	User     string `envconfig:"USER" default:"voco"`
	Password string `envconfig:"PASSWORD" default:"voco"`
	Database string `envconfig:"DATABASE" default:"voco"`
	Schema   string `envconfig:"SCHEMA" default:"voco"`
	SSLMode  string `envconfig:"SSLMODE" default:"disable"`

	MaxConns        int32         `envconfig:"MAX_CONNS" default:"10"`
	MinConns        int32         `envconfig:"MIN_CONNS" default:"0"`
	MaxConnLifetime time.Duration `envconfig:"MAX_CONN_LIFETIME" default:"1h"`
	MaxConnIdleTime time.Duration `envconfig:"MAX_CONN_IDLE_TIME" default:"30m"`
	ConnectTimeout  time.Duration `envconfig:"CONNECT_TIMEOUT" default:"5s"`

	MigrationsPath string `envconfig:"MIGRATIONS_PATH" default:"migrations"`
}

func (c Config) DSN() string {
	if c.URL != "" {
		return c.URL
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.Database,
	}

	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	if c.Schema != "" {
		q.Set("search_path", c.Schema)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// MigrateDSN — DSN без search_path: миграция 000001 создаёт schema voco,
// подключаться с search_path=voco до её применения нельзя.
func (c Config) MigrateDSN() string {
	raw := c.DSN()
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	q.Del("search_path")
	u.RawQuery = q.Encode()
	return u.String()
}

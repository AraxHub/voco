package auth

type Config struct {
	Enabled  bool   `envconfig:"ENABLED" default:"false"`
	URL      string `envconfig:"URL"`
	Realm    string `envconfig:"REALM" default:"voco"`
	ClientID string `envconfig:"CLIENT_ID" default:"voco-frontend"`
}

func (c Config) Issuer() string {
	base := trimRightSlash(c.URL)
	if base == "" {
		return ""
	}
	return base + "/realms/" + c.Realm
}

func trimRightSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

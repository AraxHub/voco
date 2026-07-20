package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type User struct {
	Sub   string
	Email string
	Name  string
}

type Service struct {
	enabled  bool
	clientID string
	verifier *oidc.IDTokenVerifier
}

func New(ctx context.Context, cfg Config) (*Service, error) {
	if !cfg.Enabled {
		return &Service{}, nil
	}

	issuer := cfg.Issuer()
	if issuer == "" {
		return nil, fmt.Errorf("keycloak: URL is required when ENABLED=true")
	}
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("keycloak provider: %w", err)
	}

	return &Service{
		enabled:  true,
		clientID: cfg.ClientID,
		verifier: provider.Verifier(&oidc.Config{
			SkipClientIDCheck: true,
		}),
	}, nil
}

func (s *Service) Enabled() bool {
	return s != nil && s.enabled
}

func (s *Service) Verify(ctx context.Context, rawToken string) (User, error) {
	if !s.Enabled() {
		return User{}, fmt.Errorf("auth is disabled")
	}

	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return User{}, fmt.Errorf("missing token")
	}

	token, err := s.verifier.Verify(ctx, rawToken)
	if err != nil {
		return User{}, err
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Azp   string `json:"azp"`
	}
	if err := token.Claims(&claims); err != nil {
		return User{}, err
	}

	if s.clientID != "" && claims.Azp != s.clientID {
		return User{}, fmt.Errorf("unexpected client")
	}

	return User{
		Sub:   token.Subject,
		Email: claims.Email,
		Name:  claims.Name,
	}, nil
}

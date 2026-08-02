package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"voco/internal/usecase/users"
)

type AdminConfig struct {
	Enabled      bool   `envconfig:"ENABLED" default:"false"`
	BaseURL      string `envconfig:"BASE_URL"` // e.g. http://keycloak:8080
	Realm        string `envconfig:"REALM" default:"voco"`
	ClientID     string `envconfig:"CLIENT_ID"`
	ClientSecret string `envconfig:"CLIENT_SECRET"`
}

type Directory struct {
	cfg    AdminConfig
	client *http.Client
}

func NewDirectory(cfg AdminConfig) *Directory {
	return &Directory{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (d *Directory) ListUsers(ctx context.Context) ([]users.DirectoryUser, error) {
	if d == nil || !d.cfg.Enabled {
		return nil, nil
	}
	token, err := d.token(ctx)
	if err != nil {
		return nil, err
	}
	base := strings.TrimRight(d.cfg.BaseURL, "/")
	endpoint := fmt.Sprintf("%s/admin/realms/%s/users?max=1000", base, url.PathEscape(d.cfg.Realm))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("keycloak users: %s: %s", res.Status, string(body))
	}
	var raw []struct {
		ID               string `json:"id"`
		Username         string `json:"username"`
		Email            string `json:"email"`
		FirstName        string `json:"firstName"`
		LastName         string `json:"lastName"`
		Enabled          bool   `json:"enabled"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	out := make([]users.DirectoryUser, 0, len(raw))
	for _, u := range raw {
		if !u.Enabled {
			continue
		}
		name := strings.TrimSpace(u.FirstName + " " + u.LastName)
		out = append(out, users.DirectoryUser{
			Sub: u.ID, Email: u.Email, DisplayName: name, Username: u.Username,
		})
	}
	return out, nil
}

func (d *Directory) token(ctx context.Context) (string, error) {
	base := strings.TrimRight(d.cfg.BaseURL, "/")
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", d.cfg.ClientID)
	form.Set("client_secret", d.cfg.ClientSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		base+"/realms/"+d.cfg.Realm+"/protocol/openid-connect/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("keycloak token: %s: %s", res.Status, string(body))
	}
	var parsed struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.AccessToken == "" {
		return "", fmt.Errorf("empty access_token")
	}
	return parsed.AccessToken, nil
}

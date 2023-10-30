package google

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	o2 "github.com/adoublef/past-swift/internal/oauth2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const api = "https://www.googleapis.com/oauth2/v2/userinfo"

type Config struct {
	oauth2.Config
}

// NewConfig will return an OAuth configuration to be used with oauth2.Authenticator
func NewConfig(opts ...o2.ConfigOption) *Config {
	c := oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		Scopes:       []string{"openid", "email", "profile"},
	}
	for _, o := range opts {
		o(&c)
	}
	return &Config{c}
}

func (p *Config) UserInfo(ctx context.Context, tok *oauth2.Token) (*o2.UserInfo, error) {
	r, err := p.Client(ctx, tok).Get(api)
	if err != nil {
		return nil, fmt.Errorf("response from github user api: %w", err)
	}
	defer r.Body.Close()
	// parse body
	var v User
	err = json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return nil, fmt.Errorf("decoding json body: %w", err)
	}
	u := o2.UserInfo{
		ID: o2.ID{
			Provider: "google",
			UserID:   v.ID},
		Photo: v.Picture,
		Login: strings.Split(v.Email, "@")[0],
		Name:  v.Name,
	}
	return &u, nil
}

// User is a provider agnostic representation of a user
//
// See (https://github.com/googleapis/google-api-go-client/blob/main/oauth2/v2/oauth2-gen.go).
type User struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
	Name    string `json:"name"`
}

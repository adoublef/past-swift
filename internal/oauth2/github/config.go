package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/adoublef/past-swift/env"
	o2 "github.com/adoublef/past-swift/internal/oauth2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const api = "https://api.github.com/user"

type Config struct {
	oauth2.Config
}

// NewConfig will return an OAuth configuration to be used with oauth2.Authenticator
func NewConfig(opts ...o2.ConfigOption) *Config {
	c := oauth2.Config{
		ClientID:     env.Must("GITHUB_CLIENT_ID"),
		ClientSecret: env.Must("GITHUB_CLIENT_SECRET"),
		// remove soon
		RedirectURL: env.WithValue("HOSTNAME", "http://localhost:8080") + "/callback",
		Endpoint:    github.Endpoint,
		Scopes:      []string{"openid"},
	}
	for _, o := range opts {
		o(&c)
	}
	return &Config{c}
}

func (c *Config) UserInfo(ctx context.Context, tok *oauth2.Token) (*o2.UserInfo, error) {
	r, err := c.Client(ctx, tok).Get(api)
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
			Provider: "github",
			UserID:   strconv.Itoa(v.ID)},
		Photo: v.AvatarUrl,
		Login: v.Login,
		Name:  v.Name,
	}
	return &u, nil
}

// User is a provider agnostic representation of a user
//
// See (https://github.com/google/go-github/blob/master/github/users.go)
type User struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	AvatarUrl string `json:"avatar_url"`
	Name      string `json:"name"`
}

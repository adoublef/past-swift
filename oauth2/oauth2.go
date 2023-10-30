package oauth2

import (
	"context"
	"errors"

	"golang.org/x/oauth2"
)

type ID struct {
	// Provider is the name of the oauth2 service
	Provider string
	// UserID is the id returned by the oauth2 service
	UserID string
}

type UserInfo struct {
	// ID is a compound of the auth provider and the associated id
	ID    ID     `json:"id"`
	Photo string `json:"photo"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

func RedirectURL(url string) ConfigOption {
	return func(c *oauth2.Config) {
		c.RedirectURL = url
	}
}

type ConfigOption func(*oauth2.Config)

type Config interface {
	UserInfo(ctx context.Context, tok *oauth2.Token) (*UserInfo, error)
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type Configs map[string]Config

func (pp Configs) Get(key string) (Config, error) {
	p, ok := pp[key]
	if !ok {
		return nil, errors.New("provider not found")
	}

	return p, nil
}

func (pp Configs) Set(key string, p Config) {
	if _, found := pp[key]; !found {
		pp[key] = p
	}
}

package oauth2

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/oauth2"
)

var (
	_ fmt.Stringer  = (ID{})
	_ driver.Valuer = (*ID)(nil)
)

type ID struct {
	// Provider is the name of the oauth2 service
	Provider string
	// UserID is the id returned by the oauth2 service
	UserID string
}

// TODO implement scanner

// Value implements driver.Valuer.
func (i ID) Value() (driver.Value, error) {
	p := strings.ToLower(i.Provider)
	// TODO validate that provider is valid
	ok := slices.Contains([]string{"github", "google"}, p)
	if !ok {
		return nil, errors.New("invalid provider")
	}
	return i.String(), nil
}

// String implements Stringer
func (i ID) String() string { return i.Provider + "|" + i.UserID }

func NewID(provider, value string) ID {
	return ID{Provider: provider, UserID: value}
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

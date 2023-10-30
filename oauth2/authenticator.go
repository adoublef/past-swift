package oauth2

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/adoublef/past-swift/internal/sessions"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	cs Configs
}

func (a *Authenticator) Configs() Configs {
	if a.cs == nil {
		a.cs = make(Configs)
	}
	return a.cs
}

// SignIn
func (a Authenticator) SignIn(w http.ResponseWriter, r *http.Request, c Config) (url string, err error) {
	var (
		ctx     = r.Context()
		session = sessions.FromContext(ctx)
		state   = oauth2.GenerateVerifier()
	)
	// get session store from context
	_, err = session.OAuth().Set(w, r, state)
	if err != nil {
		return "", err
	}
	return c.AuthCodeURL(state), nil
}

// HandleCallback
func (a Authenticator) HandleCallback(w http.ResponseWriter, r *http.Request, p Config) (u *UserInfo, err error) {
	var (
		ctx     = r.Context()
		session = sessions.FromContext(ctx)
	)
	// get cookie
	state, err := session.OAuth().Get(w, r)
	if err != nil {
		return nil, err
	}
	defer session.OAuth().Delete(w, r)
	// compare with url of state on request
	if !compare(state, r.FormValue("state")) {
		return nil, errors.New("state value mismatch")
	}
	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{Timeout: 2 * time.Second}
	ctx = context.WithValue(r.Context(), oauth2.HTTPClient, httpClient)
	// exchange `code` for `tok`
	tok, err := p.Exchange(ctx, r.FormValue("code"))
	if err != nil {
		return nil, fmt.Errorf("exchanging for token: %w", err)
	}
	// get `userinfo`
	u, err = p.UserInfo(ctx, tok)
	return
}

func compare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) != 0
}

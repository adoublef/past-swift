package sessions

import (
	"context"
	"fmt"
	"net/http"
)

var _ fmt.Stringer = (*contextKey)(nil)

type contextKey struct {
	name string
}

// String implements fmt.Stringer.
func (k *contextKey) String() string {
	return "sessions context key" + k.name
}

var (
	sessionKey = &contextKey{"session-key"}
)

func WithSession(parent context.Context, session *Session) context.Context {
	return context.WithValue(parent, sessionKey, session)
}

func FromContext(ctx context.Context) *Session {
	session, ok := ctx.Value(sessionKey).(*Session)
	if !ok {
		panic("session was not set")
	}
	return session
}

// Redirect will redirect the request if value exists inside the cookie
func Redirect(url string) func(f http.Handler) http.Handler {
	return func(f http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var s = FromContext(r.Context())
			if _, err := s.Site().Get(w, r); err != nil {
				f.ServeHTTP(w, r)
			} else {
				http.Redirect(w, r, url, http.StatusFound)
			}
		})
	}
}

// Protected will allow the request to pass-through if a `profile` exists in the cookie store
func Protected() func(f http.Handler) http.Handler {
	return func(f http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var s = FromContext(r.Context())
			if _, err := s.Site().Get(w, r); err != nil {
				http.Redirect(w, r, "/", http.StatusFound)
			} else {
				f.ServeHTTP(w, r)
			}
		})
	}
}

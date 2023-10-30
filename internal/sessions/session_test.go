package sessions_test

import (
	"context"
	"crypto/tls"
	"embed"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/adoublef/past-swift/internal/sessions"
	"github.com/adoublef/past-swift/sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/xid"
	is "github.com/stretchr/testify/require"
)

//go:embed all:migrations/*.up.sql
var migrations embed.FS

func TestSession(t *testing.T) {
	t.Run("Set", withSession(func(t *testing.T, s *sessions.Session) {
		t.Run("should set 'site' cookie", func(t *testing.T) {
			// set `oauth`
			var (
				w, r    = newTestServer(true)
				profile = xid.New()
			)
			// set
			session, err := s.Set(w, r, sessions.SessionSite, profile.String(), 24*time.Hour)
			is.NoError(t, err)

			c := w.Result().Cookies()[0]
			is.Equal(t, "_Host-site-session", c.Name)
			is.Equal(t, session.String(), c.Value)
		})
	}))

	t.Run("Get", withSession(func(t *testing.T, s *sessions.Session) {
		t.Run("should get cookie", func(t *testing.T) {
			var (
				w, r    = newTestServer(true)
				profile = xid.New()
			)
			// set
			_, err := s.Set(w, r, sessions.SessionSite, profile.String(), 24*time.Hour)
			is.NoError(t, err)
			r.AddCookie(w.Result().Cookies()[0])
			// get
			found, err := s.Get(w, r, sessions.SessionSite)
			is.NoError(t, err)
			is.Equal(t, found, profile.String())
		})
	}))

	t.Run("Delete", withSession(func(t *testing.T, s *sessions.Session) {
		t.Run("should delete session", func(t *testing.T) {
			var (
				profile = xid.New()
			)
			// set
			w, r := newTestServer(true)
			{
				_, err := s.Set(w, r, sessions.SessionSite, profile.String(), 24*time.Hour)
				is.NoError(t, err)
				c := w.Result().Cookies()[0]
				is.Equal(t, 86400, c.MaxAge)
			}
			// delete
			sc := w.Result().Cookies()[0]
			{
				w, r = newTestServer(true)
				r.AddCookie(sc)

				err := s.Delete(w, r, sessions.SessionSite)
				is.NoError(t, err)
				// check
				c := w.Result().Cookies()[0]
				is.True(t, strings.HasPrefix(c.Name, "_Host-"))
				is.Equal(t, "", c.Value)
				// not working
				is.Equal(t, -1, c.MaxAge)
			}
		})
	}))
}

// option to make secure secure
func newTestServer(secure bool) (w *httptest.ResponseRecorder, r *http.Request) {
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	if secure {
		r.TLS = &tls.ConnectionState{}
	}
	return w, r
}

func withSession(f func(t *testing.T, s *sessions.Session)) func(t *testing.T) {
	return func(t *testing.T) {
		dsn := path.Join(t.TempDir(), "cache.db")
		// run migration
		{
			db, err := sqlite3.Open(dsn)
			if err != nil {
				t.Fatalf("opening database: %v", err)
			}
			t.Cleanup(func() { db.Close() })

			err = sqlite3.Up(context.TODO(), db, migrations, "migrations")
			if err != nil {
				t.Fatalf("execute migrations: %v", err)
			}
		}
		// create session
		// both use similar APIs so only need to test one
		s, err := sessions.NewSession(context.TODO(), dsn)
		if err != nil {
			t.Fatalf("create session: %v", err)
		}
		f(t, s)
	}
}

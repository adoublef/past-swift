package sessions

import (
	"context"
	"database/sql"
	"embed"
	"io"
	"net/http"
	"time"

	"github.com/adoublef/past-swift/sqlite3"
	"github.com/gofrs/uuid"
)

//go:embed all:migrations/*.up.sql
var migrations embed.FS

var _ io.Closer = (*Session)(nil)

type Session struct {
	db *sql.DB
}

// Close implements io.Closer.
func (s *Session) Close() (err error) { return s.db.Close() }

func NewSession(ctx context.Context, dsn string) (s *Session, err error) {
	db, err := sqlite3.Open(dsn)
	if err != nil {
		return nil, err
	}
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	s = &Session{
		db: db,
	}
	return
}

func (s *Session) Up(ctx context.Context) (err error) {
	return sqlite3.Up(ctx, s.db, migrations, "migrations")
}

func (s Session) Set(w http.ResponseWriter, r *http.Request, name, value string, expiry time.Duration) (session uuid.UUID, err error) {
	var (
		ctx = r.Context()
		qry = "INSERT INTO sessions (id, name, value) VALUES (?, ?, ?)"
	)
	session, err = uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}
	_, err = s.db.ExecContext(ctx, qry, session, name, value)
	if err != nil {
		return uuid.Nil, err
	}
	setCookie(w, r, name, session.String(), expiry)
	return
}

func (s Session) Get(w http.ResponseWriter, r *http.Request, name string) (value string, err error) {
	var (
		ctx = r.Context()
		qry = "SELECT s.value FROM sessions AS s WHERE s.id = ?"
	)
	c, err := cookie(r, name)
	if err != nil {
		return "", err
	}
	session, err := uuid.FromString(c.Value)
	if err != nil {
		return "", err
	}
	err = s.db.QueryRowContext(ctx, qry, session).Scan(&value)
	if err != nil {
		return "", err
	}
	return
}

func (s Session) Delete(w http.ResponseWriter, r *http.Request, name string) (err error) {
	var (
		ctx = r.Context()
		qry = "DELETE FROM sessions AS s WHERE s.id = ?"
	)

	c, err := cookie(r, name)
	if err != nil {
		return err
	}
	session, err := uuid.FromString(c.Value)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, qry, session)
	if err != nil {
		return err
	}
	setCookie(w, r, name, "", -1)
	return
}

const (
	SessionSite  = "site-session"
	SessionOAuth = "oauth-session"
)

func cookie(r *http.Request, name string) (c *http.Cookie, err error) {
	var (
		secure = false
	)
	// this will need to change
	if secure = r.TLS != nil; secure {
		name = "_Host-" + name
	}
	c, err = r.Cookie(name)
	return
}

func setCookie(w http.ResponseWriter, r *http.Request, name string, value string, maxAge time.Duration) {
	var (
		secure = false
	)

	if secure = r.TLS != nil; secure {
		name = "_Host-" + name
	}

	var t int
	// TODO this is bad
	if maxAge >= 0 {
		t = int(maxAge.Seconds())
	} else {
		t = -1
	}

	c := http.Cookie{
		Name:     name,
		Secure:   secure,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   t,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &c)
}

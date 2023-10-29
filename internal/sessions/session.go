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
	"github.com/rs/xid"
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

func (s Session) Site() *SiteSession { return &SiteSession{db: s.db} }

func (s Session) OAuth() *OAuthSession { return &OAuthSession{db: s.db} }

var _ io.Closer = (*SiteSession)(nil)

type SiteSession struct {
	db *sql.DB
}

func (s *SiteSession) Close() error {
	return s.db.Close()
}

func NewSiteSession(ctx context.Context, dsn string) (s *SiteSession, err error) {
	db, err := sqlite3.Open(dsn)
	if err != nil {
		return nil, err
	}
	s = &SiteSession{db: db}
	err = s.db.PingContext(ctx)
	return
}

func (s *SiteSession) Set(w http.ResponseWriter, r *http.Request, profile xid.ID) (session uuid.UUID, err error) {
	var (
		ctx  = r.Context()
		name = "site-session"
		qry  = "INSERT INTO site (id, profile) VALUES (?, ?)"
	)
	session, err = uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}
	_, err = s.db.ExecContext(ctx, qry, session, profile)
	if err != nil {
		return uuid.Nil, err
	}
	setCookie(w, r, name, session.String(), 24*time.Hour)
	return
}

func (s SiteSession) Get(w http.ResponseWriter, r *http.Request) (profile xid.ID, err error) {
	var (
		ctx  = r.Context()
		name = "site-session"
		qry  = "SELECT s.profile FROM site AS s WHERE s.id = ?"
	)
	c, err := cookie(r, name)
	if err != nil {
		return xid.NilID(), err
	}
	session, err := uuid.FromString(c.Value)
	if err != nil {
		return xid.NilID(), err
	}
	err = s.db.QueryRowContext(ctx, qry, session).Scan(&profile)
	if err != nil {
		return xid.NilID(), err
	}
	return
}

func (s SiteSession) Delete(w http.ResponseWriter, r *http.Request) (err error) {
	var (
		ctx  = r.Context()
		name = "site-session"
		qry  = "DELETE FROM site AS s WHERE s.id = ?"
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

var _ io.Closer = (*OAuthSession)(nil)

type OAuthSession struct {
	db *sql.DB
}

// Close implements io.Closer.
func (s *OAuthSession) Close() error {
	return s.db.Close()
}

func NewOAuthSession(ctx context.Context, dsn string) (s *OAuthSession, err error) {
	db, err := sqlite3.Open(dsn)
	if err != nil {
		return nil, err
	}
	s = &OAuthSession{db: db}
	err = s.db.PingContext(ctx)
	return
}

func (s *OAuthSession) Set(w http.ResponseWriter, r *http.Request, state string) (session uuid.UUID, err error) {
	var (
		ctx  = r.Context()
		name = "oauth-session"
		qry  = "INSERT INTO oauth (id, state) VALUES (?, ?)"
	)
	session, err = uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}
	_, err = s.db.ExecContext(ctx, qry, session, state)
	if err != nil {
		return uuid.Nil, err
	}
	setCookie(w, r, name, session.String(), 10*time.Minute)
	return
}

func (s OAuthSession) Get(w http.ResponseWriter, r *http.Request) (state string, err error) {
	var (
		ctx  = r.Context()
		name = "oauth-session"
		qry  = "SELECT s.state FROM oauth AS s WHERE s.id = ?"
	)
	c, err := cookie(r, name)
	if err != nil {
		return "", err
	}
	session, err := uuid.FromString(c.Value)
	if err != nil {
		return "", err
	}
	err = s.db.QueryRowContext(ctx, qry, session).Scan(&state)
	if err != nil {
		return "", err
	}
	return
}

func (s OAuthSession) Delete(w http.ResponseWriter, r *http.Request) (err error) {
	var (
		ctx  = r.Context()
		name = "oauth-session"
		qry  = "DELETE FROM oauth AS s WHERE s.id = ?"
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

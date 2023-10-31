package http

import (
	"database/sql"
	"embed"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adoublef/past-swift/fly"
	"github.com/adoublef/past-swift/internal/iam"
	iamDB "github.com/adoublef/past-swift/internal/iam/sqlite3"
	o2 "github.com/adoublef/past-swift/internal/oauth2"
	"github.com/adoublef/past-swift/internal/oauth2/github"
	"github.com/adoublef/past-swift/internal/sessions"
	"github.com/adoublef/past-swift/sqlite3"
	"github.com/adoublef/past-swift/template"
	"github.com/go-chi/chi/v5"
)

//go:embed all:partials/*.html
var embedFS embed.FS
var T = template.NewFS(embedFS, "partials/*.html")

var _ http.Handler = (*Service)(nil)

type Service struct {
	m  *chi.Mux
	a  *o2.Authenticator
	db sqlite3.DB
	t  template.Template
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(dsn string, templates template.Template) (*Service, error) {
	db, err := sqlite3.Open(dsn)
	if err != nil {
		return nil, err
	}
	s := Service{
		m:  chi.NewMux(),
		a:  &o2.Authenticator{},
		db: db,
		t:  templates,
	}
	s.routes()
	return &s, nil
}

func (s *Service) routes() {
	s.a.Configs().Set("github", github.NewConfig())

	// site session exists
	ssh := s.m.With(sessions.Redirect("/projects"))
	ssh.Get("/", s.handleIndex())
	// replay on primary
	replay := s.m.With(fly.Replay(os.Getenv("DATABASE_URL")))
	replay.Get("/signin", s.handleSignIn())
	replay.Get("/callback", s.handleCallback())
	replay.Get("/signout", s.handleSignOut())
}

func (s *Service) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		template.ExecuteHTTP(w, r, s.t, "index.html", nil)
	}
}

func (s *Service) handleSignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := s.a.Configs().Get("github")
		url, err := s.a.SignIn(w, r, c)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to create auth code url", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func (s *Service) handleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx     = r.Context()
			session = sessions.FromContext(ctx)
			db      = s.db
		)
		c, _ := s.a.Configs().Get("github")
		info, err := s.a.HandleCallback(w, r, c)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to complete auth flow", http.StatusUnauthorized)
			return
		}
		u := iam.NewUser(info.ID, info.Login, info.Photo, info.Name)
		found, err := iamDB.ExistingProfile(ctx, db, info.ID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := iamDB.RegisterUser(ctx, db, u); err != nil {
				http.Error(w, "Failed to register user", http.StatusInternalServerError)
				return
			}
		case err != nil:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// id
		var id = u.ID()
		if found != nil {
			id = found.ID
		}
		_, err = session.Set(w, r, sessions.SessionSite, id.String(), 24*time.Hour)
		if err != nil {
			http.Error(w, "Failed to set site cookie", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (s *Service) handleSignOut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx     = r.Context()
			session = sessions.FromContext(ctx)
		)
		// delete site
		session.Delete(w, r, sessions.SessionSite)
		// redirect to home
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

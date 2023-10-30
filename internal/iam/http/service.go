package http

import (
	"database/sql"
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/adoublef/past-swift/fly"
	"github.com/adoublef/past-swift/internal/iam"
	"github.com/adoublef/past-swift/internal/iam/sqlite3"
	"github.com/adoublef/past-swift/internal/sessions"
	o2 "github.com/adoublef/past-swift/oauth2"
	"github.com/adoublef/past-swift/oauth2/github"
	s3 "github.com/adoublef/past-swift/sqlite3"
	tpl "github.com/adoublef/past-swift/template"
	"github.com/go-chi/chi/v5"
)

//go:embed all:partials/*.html
var embedFS embed.FS
var T = tpl.NewFS(embedFS, "partials")

// FIXME move to another package

var _ http.Handler = (*Service)(nil)

type Service struct {
	m  *chi.Mux
	a  *o2.Authenticator
	db *sql.DB
	T  *template.Template
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(dsn string, templates *template.Template) (*Service, error) {
	db, err := s3.Open(dsn)
	if err != nil {
		return nil, err
	}
	s := Service{
		m:  chi.NewMux(),
		a:  &o2.Authenticator{},
		db: db,
		T:  templates,
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
		err := s.T.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			// log.Println(err)
			http.Error(w, "Writing template error", http.StatusInternalServerError)
			return
		}
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
		found, err := sqlite3.ExistingProfile(ctx, db, info.ID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := sqlite3.RegisterUser(ctx, db, u); err != nil {
				http.Error(w, "Failed to register user", http.StatusInternalServerError)
				return
			}
		case err != nil:
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// id
		var id = u.ID()
		if found != nil {
			id = found.ID
		}
		_, err = session.Site().Set(w, r, id)
		if err != nil {
			log.Println(err)
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
		defer session.Site().Delete(w, r)
		// redirect to home
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

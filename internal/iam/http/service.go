package http

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adoublef/past-swift/fly"
	"github.com/adoublef/past-swift/internal/sessions"
	o2 "github.com/adoublef/past-swift/oauth2"
	"github.com/adoublef/past-swift/oauth2/github"
	"github.com/go-chi/chi/v5"
	"github.com/rs/xid"
)

var _ http.Handler = (*Service)(nil)

type Service struct {
	m *chi.Mux
	a *o2.Authenticator
	s *sessions.Session
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(sessions *sessions.Session) *Service {
	s := Service{
		m: chi.NewMux(),
		a: &o2.Authenticator{},
		s: sessions,
	}
	s.routes()
	return &s
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
		body := fmt.Sprintf(`
			<a href="/">Home</a>
			<a href="/signin">Sign in with GitHub</a>
			<p>Region: %s</p>
		`, os.Getenv("FLY_REGION"))

		w.Write([]byte(body))
	}
}

func (s *Service) handleSignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := s.a.Configs().Get("github")
		url, err := s.a.SignIn(w, r, c)
		if err != nil {
			http.Error(w, "Failed to create auth code url", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func (s *Service) handleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := s.a.Configs().Get("github")
		ou, err := s.a.HandleCallback(w, r, c)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to complete auth flow", http.StatusUnauthorized)
			return
		}
		fmt.Println(ou)
		_, err = s.s.Site().Set(w, r, xid.New())
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
		// delete site
		_ = s.s.Site().Delete(w, r)
		// redirect to home
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

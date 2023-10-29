package http

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adoublef/past-swift/fly"
	"github.com/adoublef/past-swift/internal/sessions"
	"github.com/go-chi/chi/v5"
	"github.com/rs/xid"
	"golang.org/x/oauth2"
)

var _ http.Handler = (*Service)(nil)

type Service struct {
	m *chi.Mux
	s *sessions.Session
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(sessions *sessions.Session) *Service {
	s := Service{
		m: chi.NewMux(),
		s: sessions,
	}
	s.routes()
	return &s
}

func (s *Service) routes() {
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
		// set oauth
		_, err := s.s.OAuth().Set(w, r, oauth2.GenerateVerifier())
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to create oauth cookie", http.StatusInternalServerError)
			return
		}
		// redirect to callback
		http.Redirect(w, r, "/callback", http.StatusFound)
	}
}

func (s *Service) handleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// delete oauth
		err := s.s.OAuth().Delete(w, r)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to delete oauth cookie", http.StatusInternalServerError)
			return
		}
		// set site
		_, err = s.s.Site().Set(w, r, xid.New())
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to set site cookie", http.StatusInternalServerError)
			return
		}
		// redirect to home
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

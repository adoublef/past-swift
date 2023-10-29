package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/adoublef/past-swift/internal/sessions"
	"github.com/go-chi/chi/v5"
)

var _ http.Handler = (*Service)(nil)

type Service struct {
	m *chi.Mux
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New() *Service {
	s := Service{
		m: chi.NewMux(),
	}
	s.routes()
	return &s
}

func (s *Service) routes() {
	// protected
	ssh := s.m.With(sessions.Protected())
	ssh.Get("/", s.handleIndex())
}

func (s *Service) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := fmt.Sprintf(`
			<a href="/">Home</a>
			<a href="/signout">Sign out</a>
			<p>Region: %s</p>
		`, os.Getenv("FLY_REGION"))

		w.Write([]byte(body))
	}
}
package http

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/adoublef/past-swift/internal/sessions"
	tpl "github.com/adoublef/past-swift/template"
	"github.com/go-chi/chi/v5"
)

//go:embed all:partials/*.html
var embedFS embed.FS
var T = tpl.NewFS(embedFS, "partials/*.html")

var _ http.Handler = (*Service)(nil)

type Service struct {
	m *chi.Mux
	t *template.Template
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func New(templates *template.Template) *Service {
	s := Service{
		m: chi.NewMux(),
		t: templates,
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
		tpl.ExecuteHTTP(w, r, s.t, "index.html", nil)
	}
}

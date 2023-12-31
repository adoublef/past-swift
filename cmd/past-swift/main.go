package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/adoublef/past-swift/env"
	iamHTTP "github.com/adoublef/past-swift/internal/iam/http"
	prjHTTP "github.com/adoublef/past-swift/internal/projects/http"
	"github.com/adoublef/past-swift/internal/sessions"
	"github.com/adoublef/past-swift/static"
	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-q
		cancel()
	}()

	if err := run(ctx); err != nil {
		log.Fatalf("past-swift: %s", err)
	}
}

func run(ctx context.Context) (err error) {
	// session
	ss, err := sessions.NewSession(ctx, os.Getenv("DATABASE_URL_SESSIONS"))
	if err != nil {
		return err
	}
	ctx = sessions.WithSession(ctx, ss)
	mux := chi.NewMux()
	// iam
	{
		t, err := iamHTTP.T.Funcs(static.FuncMap("/static")).Parse()
		if err != nil {
			return err
		}
		iam, err := iamHTTP.New(env.Must("DATABASE_URL"), t)
		if err != nil {
			return err
		}
		mux.Mount("/", iam)
	}
	// projects
	{
		t, err := prjHTTP.T.Funcs(static.FuncMap("/static")).Parse()
		if err != nil {
			return err
		}
		mux.Mount("/projects", prjHTTP.New(t))
	}
	// static
	{
		mux.Handle("/static/*", static.Handler("/static"))
	}
	// design
	{
		t, err := T.Funcs(static.FuncMap("/static")).Parse()
		if err != nil {
			return err
		}
		mux.Get("/design", handleDesign(t))
	}
	s := &http.Server{
		Addr:    ":" + env.WithValue("PORT", "8080"),
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
	sErr := make(chan error)
	go func() {
		sErr <- s.ListenAndServe()
	}()

	select {
	case err := <-sErr:
		return fmt.Errorf("main error: starting server: %w", err)
	case <-ctx.Done():
		// TODO
		return s.Shutdown(context.Background())
	}
}

/*
-- iam
/
/signin/{provider.id}
/callback/{provider.id}
/signout
/@{profile.login}
/profile/settings
-- projects
/projects
/projects/{project.id}
/projects/{project.id}/invite
/projects/{project.id}/join
-- static
/static/*
-- media
/media/track/{media.id}
*/

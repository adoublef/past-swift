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

	iamHTTP "github.com/adoublef/past-swift/internal/iam/http"
	prjHTTP "github.com/adoublef/past-swift/internal/projects/http"
	"github.com/adoublef/past-swift/internal/sessions"
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
	mux := chi.NewMux()
	// iam
	{
		iam, err := iamHTTP.New(os.Getenv("DATABASE_URL"))
		if err != nil {
			return err
		}
		mux.Mount("/", iam)
	}
	// projects
	{
		mux.Mount("/projects", prjHTTP.New())
	}
	s := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			return sessions.WithSession(ctx, ss)
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

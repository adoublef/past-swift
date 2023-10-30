package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	iamDB "github.com/adoublef/past-swift/internal/iam/sqlite3"
	"github.com/adoublef/past-swift/internal/sessions"
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
		log.Fatalf("migration: %s", err)
	}
}

func run(ctx context.Context) (err error) {
	// migration for `iam` module
	err = iamDB.Up(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	// migrations for `session` module
	ss, err := sessions.NewSession(ctx, os.Getenv("DATABASE_URL_SESSIONS"))
	if err != nil {
		return err
	}
	err = ss.Up(ctx)
	if err != nil {
		return err
	}
	return nil
}

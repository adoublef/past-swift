package sqlite3_test

import (
	"context"
	"database/sql"
	"embed"
	"path"
	"testing"

	"github.com/adoublef/past-swift/internal/iam"
	"github.com/adoublef/past-swift/internal/iam/sqlite3"
	"github.com/adoublef/past-swift/internal/oauth2"
	s3 "github.com/adoublef/past-swift/sqlite3"
	_ "github.com/mattn/go-sqlite3"
	is "github.com/stretchr/testify/require"
)

//go:embed all:migrations/*.up.sql
var migrations embed.FS

func TestSqlite3(t *testing.T) {
	t.Run("RegisterUser", withClient(func(t *testing.T, db *sql.DB) {
		// create a new user
		a := iam.NewUser(oauth2.NewID("google", "1"),
			"alpha@gmail.com", "https://avatar.google.com/alpha", "Alpha")

		err := sqlite3.RegisterUser(context.Background(), db, a)
		is.NoError(t, err, "register 'a")

		// fail to add duplicate user
		err = sqlite3.RegisterUser(context.Background(), db, a)
		is.Error(t, err, "duplicate 'a'")

		// add second user
		b := iam.NewUser(oauth2.NewID("google", "2"),
			"bravo@gmail.com", "https://avatar.google.com/bravo", "Bravo")

		err = sqlite3.RegisterUser(context.Background(), db, b)
		is.NoError(t, err, "register 'b'")
	}))

	t.Run("ExistingProfile", withClient(func(t *testing.T, db *sql.DB) {
		// create a new user
		a := iam.NewUser(oauth2.NewID("google", "1"),
			"alpha@gmail.com", "https://avatar.google.com/alpha", "Alpha")
		err := sqlite3.RegisterUser(context.Background(), db, a)
		is.NoError(t, err, "register 'a")

		// existing profile
		found, err := sqlite3.ExistingProfile(context.Background(), db, a.OAuth2ID)
		is.NoError(t, err, "lookup 'a' by 'oauth2'")
		is.Equal(t, a.Profile, found)
	}))
}

func withClient(f func(t *testing.T, db *sql.DB)) func(t *testing.T) {
	return func(t *testing.T) {
		dsn := path.Join(t.TempDir(), "test.db")
		db, err := s3.Open(dsn)
		if err != nil {
			t.Fatalf("opening database: %v", err)
		}
		t.Cleanup(func() { db.Close() })
		// run migration
		err = s3.Up(context.TODO(), db, migrations, "migrations")
		if err != nil {
			t.Fatalf("execute migrations: %v", err)
		}

		f(t, db)
	}
}

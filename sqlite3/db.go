package sqlite3

import (
	"context"
	"database/sql"
	"io/fs"
	"strings"

	"github.com/maragudk/migrate"
)

var args = strings.Join([]string{"_journal=wal", "_timeout=5000", "_synchronous=normal", "_fk=true"}, "&")

// Open opens a connection to the sqlite instance with "WAL" & foreign key support
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn+"?"+args)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Up replays the migration scripts from the current version
//
// NOTE https://github.com/maragudk/migrate/blob/main/migrate_test.go#L361
func Up(ctx context.Context, db *sql.DB, fsys fs.FS, dir string) (err error) {
	f, err := fs.Sub(fsys, dir)
	if err != nil {
		return err
	}
	err = migrate.Up(ctx, db, f)
	return
}

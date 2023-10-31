package sqlite3

import (
	"context"
	"database/sql"
	"io/fs"
	"strings"

	"github.com/maragudk/migrate"
)

type DB interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Ping() error
	PingContext(ctx context.Context) error
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

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

// lookup_profile.go
package sqlite3

import (
	"context"
	"database/sql"
	"embed"

	"github.com/adoublef/past-swift/internal/iam"
	"github.com/adoublef/past-swift/internal/oauth2"
	"github.com/adoublef/past-swift/sqlite3"
)

//go:embed all:migrations/*.up.sql
var migrations embed.FS

// Up replays migration scripts from current version
func Up(ctx context.Context, dsn string) (err error) {
	db, err := sqlite3.Open(dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return sqlite3.Up(ctx, db, migrations, "migrations")
}

// RegisterUser will add rows to the database to register a new user
func RegisterUser(ctx context.Context, db *sql.DB, u *iam.User) (err error) {
	// begin
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// try to insert profile
	_, err = tx.ExecContext(ctx, `
INSERT INTO profiles (id, login, photo_url, name)
VALUES (?, ?, ?, ?)
	`, u.Profile.ID, u.Profile.Login, u.Profile.Photo, u.Profile.Name)
	if err != nil {
		return err
	}
	// insert credentials
	_, err = tx.ExecContext(ctx, `
INSERT INTO accounts (oauth, profile)
VALUES (?, ?)
	`, u.OAuth2ID, u.Profile.ID)
	if err != nil {
		return err
	}
	// commit
	err = tx.Commit()
	if err != nil {
		return err
	}
	return
}

// ExistingProfile will return a profile if one exists for a given oauth id
func ExistingProfile(ctx context.Context, db *sql.DB, oauthId oauth2.ID) (*iam.Profile, error) {
	var p iam.Profile
	err := db.QueryRowContext(ctx, `
SELECT p.id, p.login, p.photo_url, p.name
FROM profiles AS p
JOIN accounts AS a ON p.id = a.profile
WHERE a.oauth = ?
	`, oauthId).Scan(&p.ID, &p.Login, &p.Photo, &p.Name)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

package repository

import (
	"database/sql"
	"errors"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Repositories struct {
	Users  *UserRepository
	Teams  *TeamRepository
	Tokens *TokenRepository
	Grants *GrantRepository
}

func New(db *sql.DB) Repositories {
	return Repositories{
		Users:  &UserRepository{db: db},
		Teams:  &TeamRepository{db: db},
		Tokens: &TokenRepository{db: db},
		Grants: &GrantRepository{db: db},
	}
}

// mapErr translates driver errors into the repository's error values.
func mapErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	var se *sqlite.Error
	if errors.As(err, &se) {
		switch se.Code() {
		case sqlite3.SQLITE_CONSTRAINT_UNIQUE, sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY:
			return ErrConflict
		}
	}
	return err
}

// requireAffected returns ErrNotFound when a write did not touch any row.
func requireAffected(res sql.Result, err error) error {
	if err != nil {
		return mapErr(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Timestamps are stored as unix seconds.

func toUnix(t time.Time) int64 {
	return t.Unix()
}

func toNullUnix(t *time.Time) sql.NullInt64 {
	if t == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: t.Unix(), Valid: true}
}

func fromUnix(s int64) time.Time {
	return time.Unix(s, 0).UTC()
}

func fromNullUnix(s sql.NullInt64) *time.Time {
	if !s.Valid {
		return nil
	}
	t := fromUnix(s.Int64)
	return &t
}

type scanner interface {
	Scan(dest ...any) error
}

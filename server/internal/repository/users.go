package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// bcrypt only uses the first 72 bytes of a password, so longer ones are rejected instead of silently truncated.
const (
	MinPasswordLength = 8
	MaxPasswordLength = 72
)

var ErrInvalidPassword = errors.New("password must be 8-72 bytes")

// dummyHash is compared against when a user doesn't exist, so a login takes as long as for an existing user.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("smr-dummy-password"), bcrypt.DefaultCost)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
	// HasPassword reports whether the user can log in to the web UI.
	HasPassword bool      `json:"hasPassword"`
	CreatedAt   time.Time `json:"createdAt"`
}

type UserRepository struct {
	db *sql.DB
}

const userColumns = `id, username, is_admin, password_hash IS NOT NULL, created_at`

// joinedUserColumns are the userColumns of a query that joins users as u.
const joinedUserColumns = `u.id, u.username, u.is_admin, u.password_hash IS NOT NULL, u.created_at`

// scanUser scans a row of userColumns. extra are scanned from columns selected before them.
func scanUser(s scanner, extra ...any) (User, error) {
	var u User
	var createdAt int64
	dest := append(extra, &u.ID, &u.Username, &u.IsAdmin, &u.HasPassword, &createdAt)
	if err := s.Scan(dest...); err != nil {
		return User{}, mapErr(err)
	}
	u.CreatedAt = fromUnix(createdAt)
	return u, nil
}

func (r *UserRepository) Create(ctx context.Context, username string, isAdmin bool) (User, error) {
	return scanUser(r.db.QueryRowContext(ctx,
		`INSERT INTO users (username, is_admin, created_at) VALUES (?, ?, ?) RETURNING `+userColumns,
		username, isAdmin, toUnix(time.Now())))
}

func (r *UserRepository) Get(ctx context.Context, username string) (User, error) {
	return scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE username = ?`, username))
}

func (r *UserRepository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+userColumns+` FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) SetAdmin(ctx context.Context, username string, isAdmin bool) (User, error) {
	return scanUser(r.db.QueryRowContext(ctx,
		`UPDATE users SET is_admin = ? WHERE username = ? RETURNING `+userColumns, isAdmin, username))
}

// ValidatePassword returns ErrInvalidPassword if the password is too short or too long.
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength || len(password) > MaxPasswordLength {
		return ErrInvalidPassword
	}
	return nil
}

// SetPassword replaces the user's password. An empty password removes it, so the user can't log in to the web UI.
func (r *UserRepository) SetPassword(ctx context.Context, userID int64, password string) error {
	var hash []byte
	if password != "" {
		if err := ValidatePassword(password); err != nil {
			return err
		}
		var err error
		if hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err != nil {
			return err
		}
	}
	return requireAffected(r.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, userID))
}

// CheckPassword returns the user if the password matches, and ErrInvalidCredentials if the user doesn't exist, has
// no password or the password is wrong.
func (r *UserRepository) CheckPassword(ctx context.Context, username, password string) (User, error) {
	var hash []byte
	u, err := scanUser(r.db.QueryRowContext(ctx,
		`SELECT password_hash, `+userColumns+` FROM users WHERE username = ?`, username), &hash)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return User{}, err
	}
	if hash == nil {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return User{}, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return User{}, ErrInvalidCredentials
	}
	return u, nil
}

// Delete removes the user together with their tokens, team memberships and grants.
func (r *UserRepository) Delete(ctx context.Context, username string) error {
	return requireAffected(r.db.ExecContext(ctx, `DELETE FROM users WHERE username = ?`, username))
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func (r *UserRepository) CountAdmins(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_admin`).Scan(&n)
	return n, err
}

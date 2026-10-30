package repository

import (
	"context"
	"database/sql"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"isAdmin"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserRepository struct {
	db *sql.DB
}

const userColumns = `id, username, is_admin, created_at`

func scanUser(s scanner) (User, error) {
	var u User
	var createdAt int64
	if err := s.Scan(&u.ID, &u.Username, &u.IsAdmin, &createdAt); err != nil {
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

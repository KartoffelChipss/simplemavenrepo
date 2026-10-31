package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"
)

const SessionLifetime = 30 * 24 * time.Hour

type SessionRepository struct {
	db *sql.DB
}

func (r *SessionRepository) Create(ctx context.Context, userID int64) (string, time.Time, error) {
	secret := make([]byte, 32)
	rand.Read(secret)
	plaintext := base64.RawURLEncoding.EncodeToString(secret)

	now := time.Now()
	expiresAt := now.Add(SessionLifetime)
	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= ?`, toUnix(now)); err != nil {
		return "", time.Time{}, err
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (user_id, session_hash, created_at, expires_at) VALUES (?, ?, ?, ?)`,
		userID, hashToken(plaintext), toUnix(now), toUnix(expiresAt))
	if err != nil {
		return "", time.Time{}, mapErr(err)
	}
	return plaintext, expiresAt, nil
}

func (r *SessionRepository) Authenticate(ctx context.Context, plaintext string) (User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, `
		SELECT `+joinedUserColumns+` FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.session_hash = ? AND s.expires_at > ?`,
		hashToken(plaintext), toUnix(time.Now())))
	if err == ErrNotFound {
		return User{}, ErrInvalidCredentials
	}
	return u, err
}

func (r *SessionRepository) Delete(ctx context.Context, plaintext string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE session_hash = ?`, hashToken(plaintext))
	return err
}

func (r *SessionRepository) DeleteForUser(ctx context.Context, userID int64, except string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE user_id = ? AND session_hash != ?`, userID, hashToken(except))
	return err
}

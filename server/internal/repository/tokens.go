package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"strings"
	"time"
)

const TokenPrefix = "smr_"

// Tokens are shown with this many characters so it can be identified by the user.
const displayPrefixLength = len(TokenPrefix) + 6

// lastUsedResolution limits how often last_used_at is written for a token that is used repeatedly.
const lastUsedResolution = time.Minute

type Token struct {
	ID         int64        `json:"id"`
	UserID     int64        `json:"userId"`
	Name       string       `json:"name"`
	Prefix     string       `json:"prefix"`
	CreatedAt  time.Time    `json:"createdAt"`
	ExpiresAt  *time.Time   `json:"expiresAt"`
	LastUsedAt *time.Time   `json:"lastUsedAt"`
	Scopes     []TokenScope `json:"scopes"`
}

type TokenScope struct {
	Pattern    string     `json:"pattern"`
	Permission Permission `json:"permission"`
}

func (s TokenScope) Allows(p string, perm Permission) bool {
	return s.Permission.Includes(perm) && MatchPattern(s.Pattern, p)
}

type TokenRepository struct {
	db *sql.DB
}

const tokenColumns = `id, user_id, name, prefix, created_at, expires_at, last_used_at`

func scanToken(s scanner) (Token, error) {
	var t Token
	var createdAt int64
	var expiresAt, lastUsedAt sql.NullInt64
	if err := s.Scan(&t.ID, &t.UserID, &t.Name, &t.Prefix, &createdAt, &expiresAt, &lastUsedAt); err != nil {
		return Token{}, mapErr(err)
	}
	t.CreatedAt = fromUnix(createdAt)
	t.ExpiresAt = fromNullUnix(expiresAt)
	t.LastUsedAt = fromNullUnix(lastUsedAt)
	t.Scopes = []TokenScope{}
	return t, nil
}

func hashToken(plaintext string) []byte {
	sum := sha256.Sum256([]byte(plaintext))
	return sum[:]
}

// Create generates a new token for the user. The plaintext token is only returned here. Only its hash is stored.
// scopes must already be normalized with NormalizePattern, no scopes create an unrestricted token.
func (r *TokenRepository) Create(ctx context.Context, userID int64, name string, expiresAt *time.Time, scopes []TokenScope) (Token, string, error) {
	secret := make([]byte, 32)
	rand.Read(secret)
	plaintext := TokenPrefix + base64.RawURLEncoding.EncodeToString(secret)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Token{}, "", err
	}
	defer tx.Rollback()

	t, err := scanToken(tx.QueryRowContext(ctx, `
		INSERT INTO tokens (user_id, name, prefix, token_hash, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?) RETURNING `+tokenColumns,
		userID, name, plaintext[:displayPrefixLength], hashToken(plaintext), toUnix(time.Now()), toNullUnix(expiresAt)))
	if err != nil {
		return Token{}, "", err
	}
	for _, s := range scopes {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO token_scopes (token_id, pattern, permission) VALUES (?, ?, ?)`,
			t.ID, s.Pattern, s.Permission); err != nil {
			return Token{}, "", mapErr(err)
		}
	}
	if t.Scopes, err = listScopes(ctx, tx, t.ID); err != nil {
		return Token{}, "", err
	}
	if err := tx.Commit(); err != nil {
		return Token{}, "", err
	}
	return t, plaintext, nil
}

type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// listScopes returns the scopes of a token, never nil.
func listScopes(ctx context.Context, q querier, tokenID int64) ([]TokenScope, error) {
	byToken, err := queryScopes(ctx, q, `SELECT token_id, pattern, permission FROM token_scopes
		WHERE token_id = ? ORDER BY pattern, permission`, tokenID)
	if err != nil {
		return nil, err
	}
	if byToken[tokenID] == nil {
		return []TokenScope{}, nil
	}
	return byToken[tokenID], nil
}

func queryScopes(ctx context.Context, q querier, query string, args ...any) (map[int64][]TokenScope, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byToken := map[int64][]TokenScope{}
	for rows.Next() {
		var tokenID int64
		var s TokenScope
		if err := rows.Scan(&tokenID, &s.Pattern, &s.Permission); err != nil {
			return nil, err
		}
		byToken[tokenID] = append(byToken[tokenID], s)
	}
	return byToken, rows.Err()
}

func (r *TokenRepository) ListForUser(ctx context.Context, userID int64) ([]Token, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+tokenColumns+` FROM tokens WHERE user_id = ? ORDER BY created_at, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := []Token{}
	for rows.Next() {
		t, err := scanToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	byToken, err := queryScopes(ctx, r.db, `SELECT s.token_id, s.pattern, s.permission FROM token_scopes s
		JOIN tokens t ON t.id = s.token_id
		WHERE t.user_id = ? ORDER BY s.pattern, s.permission`, userID)
	if err != nil {
		return nil, err
	}
	for i := range tokens {
		if scopes := byToken[tokens[i].ID]; scopes != nil {
			tokens[i].Scopes = scopes
		}
	}
	return tokens, nil
}

// Delete revokes a token. It is scoped to userID so a user can only revoke their own tokens.
func (r *TokenRepository) Delete(ctx context.Context, userID, tokenID int64) error {
	return requireAffected(r.db.ExecContext(ctx,
		`DELETE FROM tokens WHERE id = ? AND user_id = ?`, tokenID, userID))
}

// Authenticate returns the owner of a valid, unexpired token and its scopes, and records its use. Empty scopes mean
// the token has the full access of its user.
func (r *TokenRepository) Authenticate(ctx context.Context, plaintext string) (User, []TokenScope, error) {
	if !strings.HasPrefix(plaintext, TokenPrefix) {
		return User{}, nil, ErrInvalidCredentials
	}

	now := time.Now()
	var tokenID int64
	u, err := scanUser(r.db.QueryRowContext(ctx, `
		SELECT t.id, `+joinedUserColumns+` FROM tokens t
		JOIN users u ON u.id = t.user_id
		WHERE t.token_hash = ? AND (t.expires_at IS NULL OR t.expires_at > ?)`,
		hashToken(plaintext), toUnix(now)), &tokenID)
	if err != nil {
		if err == ErrNotFound {
			return User{}, nil, ErrInvalidCredentials
		}
		return User{}, nil, err
	}

	scopes, err := listScopes(ctx, r.db, tokenID)
	if err != nil {
		return User{}, nil, err
	}

	_, err = r.db.ExecContext(ctx,
		`UPDATE tokens SET last_used_at = ? WHERE id = ? AND (last_used_at IS NULL OR last_used_at <= ?)`,
		toUnix(now), tokenID, toUnix(now.Add(-lastUsedResolution)))
	return u, scopes, err
}

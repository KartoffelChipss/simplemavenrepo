package repository

import (
	"context"
	"database/sql"
	"time"
)

type Permission string

const (
	PermissionRead  Permission = "read"
	PermissionWrite Permission = "write"
)

func (p Permission) Valid() bool {
	return p == PermissionRead || p == PermissionWrite
}

type Grant struct {
	ID         int64      `json:"id"`
	User       string     `json:"user,omitempty"`
	Team       string     `json:"team,omitempty"`
	Pattern    string     `json:"pattern"`
	Permission Permission `json:"permission"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type GrantFilter struct {
	UserID int64
	TeamID int64
}

type GrantRepository struct {
	db *sql.DB
}

const grantSelect = `
	SELECT g.id, COALESCE(u.username, ''), COALESCE(t.name, ''), g.pattern, g.permission, g.created_at
	FROM grants g
	LEFT JOIN users u ON u.id = g.user_id
	LEFT JOIN teams t ON t.id = g.team_id`

func scanGrant(s scanner) (Grant, error) {
	var g Grant
	var createdAt int64
	if err := s.Scan(&g.ID, &g.User, &g.Team, &g.Pattern, &g.Permission, &createdAt); err != nil {
		return Grant{}, mapErr(err)
	}
	g.CreatedAt = fromUnix(createdAt)
	return g, nil
}

// CreateForUser grants a permission to a user. pattern must already be normalized with NormalizePattern.
func (r *GrantRepository) CreateForUser(ctx context.Context, userID int64, pattern string, perm Permission) (Grant, error) {
	return r.create(ctx, sql.NullInt64{Int64: userID, Valid: true}, sql.NullInt64{}, pattern, perm)
}

// CreateForTeam grants a permission to a team. pattern must already be normalized with NormalizePattern.
func (r *GrantRepository) CreateForTeam(ctx context.Context, teamID int64, pattern string, perm Permission) (Grant, error) {
	return r.create(ctx, sql.NullInt64{}, sql.NullInt64{Int64: teamID, Valid: true}, pattern, perm)
}

func (r *GrantRepository) create(ctx context.Context, userID, teamID sql.NullInt64, pattern string, perm Permission) (Grant, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO grants (user_id, team_id, pattern, permission, created_at)
		VALUES (?, ?, ?, ?, ?) RETURNING id`,
		userID, teamID, pattern, perm, toUnix(time.Now())).Scan(&id)
	if err != nil {
		return Grant{}, mapErr(err)
	}
	return r.Get(ctx, id)
}

func (r *GrantRepository) Get(ctx context.Context, id int64) (Grant, error) {
	return scanGrant(r.db.QueryRowContext(ctx, grantSelect+` WHERE g.id = ?`, id))
}

func (r *GrantRepository) List(ctx context.Context, f GrantFilter) ([]Grant, error) {
	return r.query(ctx, grantSelect+`
		WHERE (? = 0 OR g.user_id = ?) AND (? = 0 OR g.team_id = ?)
		ORDER BY g.pattern, g.id`, f.UserID, f.UserID, f.TeamID, f.TeamID)
}

// ListEffective returns the grants that apply to a user, both direct ones and those of their teams.
func (r *GrantRepository) ListEffective(ctx context.Context, userID int64) ([]Grant, error) {
	return r.query(ctx, grantSelect+`
		WHERE g.user_id = ? OR g.team_id IN (SELECT team_id FROM team_members WHERE user_id = ?)
		ORDER BY g.pattern, g.id`, userID, userID)
}

func (r *GrantRepository) query(ctx context.Context, query string, args ...any) ([]Grant, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grants := []Grant{}
	for rows.Next() {
		g, err := scanGrant(rows)
		if err != nil {
			return nil, err
		}
		grants = append(grants, g)
	}
	return grants, rows.Err()
}

func (r *GrantRepository) Delete(ctx context.Context, id int64) error {
	return requireAffected(r.db.ExecContext(ctx, `DELETE FROM grants WHERE id = ?`, id))
}

package repository

import (
	"context"
	"database/sql"
	"time"
)

type Team struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type TeamRepository struct {
	db *sql.DB
}

const teamColumns = `id, name, created_at`

func scanTeam(s scanner) (Team, error) {
	var t Team
	var createdAt int64
	if err := s.Scan(&t.ID, &t.Name, &createdAt); err != nil {
		return Team{}, mapErr(err)
	}
	t.CreatedAt = fromUnix(createdAt)
	return t, nil
}

func (r *TeamRepository) Create(ctx context.Context, name string) (Team, error) {
	return scanTeam(r.db.QueryRowContext(ctx,
		`INSERT INTO teams (name, created_at) VALUES (?, ?) RETURNING `+teamColumns, name, toUnix(time.Now())))
}

func (r *TeamRepository) Get(ctx context.Context, name string) (Team, error) {
	return scanTeam(r.db.QueryRowContext(ctx, `SELECT `+teamColumns+` FROM teams WHERE name = ?`, name))
}

func (r *TeamRepository) List(ctx context.Context) ([]Team, error) {
	return r.query(ctx, `SELECT `+teamColumns+` FROM teams ORDER BY name`)
}

// ListForUser returns the teams the user is a member of.
func (r *TeamRepository) ListForUser(ctx context.Context, userID int64) ([]Team, error) {
	return r.query(ctx, `
		SELECT t.id, t.name, t.created_at FROM teams t
		JOIN team_members m ON m.team_id = t.id
		WHERE m.user_id = ? ORDER BY t.name`, userID)
}

func (r *TeamRepository) query(ctx context.Context, query string, args ...any) ([]Team, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := []Team{}
	for rows.Next() {
		t, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

// Delete removes the team together with its memberships and grants.
func (r *TeamRepository) Delete(ctx context.Context, name string) error {
	return requireAffected(r.db.ExecContext(ctx, `DELETE FROM teams WHERE name = ?`, name))
}

func (r *TeamRepository) ListMembers(ctx context.Context, teamID int64) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.username, u.is_admin, u.created_at FROM users u
		JOIN team_members m ON m.user_id = u.id
		WHERE m.team_id = ? ORDER BY u.username`, teamID)
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

// AddMember adds the user to the team. Adding an existing member is a no-op.
func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO team_members (team_id, user_id) VALUES (?, ?)`, teamID, userID)
	return mapErr(err)
}

func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID int64) error {
	return requireAffected(r.db.ExecContext(ctx,
		`DELETE FROM team_members WHERE team_id = ? AND user_id = ?`, teamID, userID))
}

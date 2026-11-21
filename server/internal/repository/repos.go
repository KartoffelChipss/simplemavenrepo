package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// VersionPolicy decides which versions can be deployed to a repository.
type VersionPolicy string

const (
	PolicyRelease  VersionPolicy = "release"
	PolicySnapshot VersionPolicy = "snapshot"
	PolicyMixed    VersionPolicy = "mixed"
)

func (p VersionPolicy) Valid() bool {
	return p == PolicyRelease || p == PolicySnapshot || p == PolicyMixed
}

// Allows reports whether a version can be deployed under the policy.
func (p VersionPolicy) Allows(version string) bool {
	snapshot := strings.HasSuffix(version, "-SNAPSHOT")
	switch p {
	case PolicyRelease:
		return !snapshot
	case PolicySnapshot:
		return snapshot
	}
	return true
}

// Repo is a Maven repository, served under /<name>.
type Repo struct {
	ID            int64         `json:"id"`
	Name          string        `json:"name"`
	Description   *string       `json:"description"`
	VersionPolicy VersionPolicy `json:"versionPolicy"`
	CreatedAt     time.Time     `json:"createdAt"`
}

type RepoRepository struct {
	db *sql.DB
}

const repoColumns = `id, name, description, version_policy, created_at`

func scanRepo(s scanner) (Repo, error) {
	var r Repo
	var description sql.NullString
	var createdAt int64
	if err := s.Scan(&r.ID, &r.Name, &description, &r.VersionPolicy, &createdAt); err != nil {
		return Repo{}, mapErr(err)
	}
	if description.Valid {
		r.Description = &description.String
	}
	r.CreatedAt = fromUnix(createdAt)
	return r, nil
}

func (r *RepoRepository) Create(ctx context.Context, name string, policy VersionPolicy, description *string) (Repo, error) {
	return scanRepo(r.db.QueryRowContext(ctx, `
		INSERT INTO repositories (name, version_policy, description, created_at)
		VALUES (?, ?, ?, ?) RETURNING `+repoColumns,
		name, policy, toNullString(description), toUnix(time.Now())))
}

func (r *RepoRepository) Get(ctx context.Context, name string) (Repo, error) {
	return scanRepo(r.db.QueryRowContext(ctx, `SELECT `+repoColumns+` FROM repositories WHERE name = ?`, name))
}

func (r *RepoRepository) List(ctx context.Context) ([]Repo, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+repoColumns+` FROM repositories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	repos := []Repo{}
	for rows.Next() {
		repo, err := scanRepo(rows)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, rows.Err()
}

// Update changes the version policy and the description. nil leaves a field unchanged, an empty description clears it.
func (r *RepoRepository) Update(ctx context.Context, id int64, policy *VersionPolicy, description *string) (Repo, error) {
	return scanRepo(r.db.QueryRowContext(ctx, `
		UPDATE repositories SET
			version_policy = COALESCE(?, version_policy),
			description = CASE WHEN ? THEN ? ELSE description END
		WHERE id = ? RETURNING `+repoColumns,
		policy, description != nil, toNullString(description), id))
}

// Delete removes the repository. It fails with ErrConflict while artifacts are registered in it.
func (r *RepoRepository) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var inUse bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM artifacts JOIN repositories ON repositories.name = artifacts.repository WHERE repositories.id = ?)`,
		id).Scan(&inUse); err != nil {
		return err
	}
	if inUse {
		return ErrConflict
	}
	if err := requireAffected(tx.ExecContext(ctx, `DELETE FROM repositories WHERE id = ?`, id)); err != nil {
		return err
	}
	return tx.Commit()
}

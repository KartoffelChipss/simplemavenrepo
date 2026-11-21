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
	AutoRegister  bool          `json:"autoRegister"`
	DefaultPublic bool          `json:"defaultPublic"`
	CreatedAt     time.Time     `json:"createdAt"`
}

type RepoRepository struct {
	db *sql.DB
}

const repoColumns = `id, name, description, version_policy, auto_register, default_public, created_at`

func scanRepo(s scanner) (Repo, error) {
	var r Repo
	var description sql.NullString
	var createdAt int64
	if err := s.Scan(&r.ID, &r.Name, &description, &r.VersionPolicy, &r.AutoRegister, &r.DefaultPublic, &createdAt); err != nil {
		return Repo{}, mapErr(err)
	}
	if description.Valid {
		r.Description = &description.String
	}
	r.CreatedAt = fromUnix(createdAt)
	return r, nil
}

type RepoOptions struct {
	VersionPolicy *VersionPolicy
	Description   *string
	AutoRegister  *bool
	DefaultPublic *bool
}

// Create adds a repository. Unset options get their defaults: mixed versions and no auto registration.
func (r *RepoRepository) Create(ctx context.Context, name string, opts RepoOptions) (Repo, error) {
	policy := PolicyMixed
	if opts.VersionPolicy != nil {
		policy = *opts.VersionPolicy
	}
	return scanRepo(r.db.QueryRowContext(ctx, `
		INSERT INTO repositories (name, version_policy, description, auto_register, default_public, created_at)
		VALUES (?, ?, ?, ?, ?, ?) RETURNING `+repoColumns,
		name, policy, toNullString(opts.Description), opts.AutoRegister != nil && *opts.AutoRegister,
		opts.DefaultPublic != nil && *opts.DefaultPublic, toUnix(time.Now())))
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

// Update changes the options that are set.
func (r *RepoRepository) Update(ctx context.Context, id int64, opts RepoOptions) (Repo, error) {
	return scanRepo(r.db.QueryRowContext(ctx, `
		UPDATE repositories SET
			version_policy = COALESCE(?, version_policy),
			description = CASE WHEN ? THEN ? ELSE description END,
			auto_register = COALESCE(?, auto_register),
			default_public = COALESCE(?, default_public)
		WHERE id = ? RETURNING `+repoColumns,
		opts.VersionPolicy, opts.Description != nil, toNullString(opts.Description), opts.AutoRegister, opts.DefaultPublic, id))
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

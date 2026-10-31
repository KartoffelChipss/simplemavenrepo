package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type Artifact struct {
	ID         int64  `json:"id"`
	GroupID    string `json:"groupId"`
	ArtifactID string `json:"artifactId"`
	Path       string `json:"path"`
	Public     bool   `json:"public"`
	AllowReleaseOverwrite bool      `json:"allowReleaseOverwrite"`
	Description           *string   `json:"description"`
	CreatedAt             time.Time `json:"createdAt"`
}

type ArtifactRepository struct {
	db *sql.DB
}

const artifactColumns = `id, group_id, artifact_id, path, is_public, allow_release_overwrite, description, created_at`

func scanArtifact(s scanner) (Artifact, error) {
	var a Artifact
	var description sql.NullString
	var createdAt int64
	if err := s.Scan(&a.ID, &a.GroupID, &a.ArtifactID, &a.Path, &a.Public, &a.AllowReleaseOverwrite, &description, &createdAt); err != nil {
		return Artifact{}, mapErr(err)
	}
	if description.Valid {
		a.Description = &description.String
	}
	a.CreatedAt = fromUnix(createdAt)
	return a, nil
}

// toNullString stores nil and empty descriptions as NULL.
func toNullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func (r *ArtifactRepository) Create(ctx context.Context, groupID, artifactID string, public, allowReleaseOverwrite bool, description *string) (Artifact, error) {
	return scanArtifact(r.db.QueryRowContext(ctx, `
		INSERT INTO artifacts (group_id, artifact_id, is_public, allow_release_overwrite, description, created_at)
		VALUES (?, ?, ?, ?, ?, ?) RETURNING `+artifactColumns,
		groupID, artifactID, public, allowReleaseOverwrite, toNullString(description), toUnix(time.Now())))
}

func (r *ArtifactRepository) Get(ctx context.Context, groupID, artifactID string) (Artifact, error) {
	return scanArtifact(r.db.QueryRowContext(ctx,
		`SELECT `+artifactColumns+` FROM artifacts WHERE group_id = ? AND artifact_id = ?`, groupID, artifactID))
}

func (r *ArtifactRepository) List(ctx context.Context) ([]Artifact, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+artifactColumns+` FROM artifacts ORDER BY group_id, artifact_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artifacts := []Artifact{}
	for rows.Next() {
		a, err := scanArtifact(rows)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, a)
	}
	return artifacts, rows.Err()
}

// Update changes the visibility, the release overwrite setting and the description. nil leaves a field unchanged,
// an empty description clears it.
func (r *ArtifactRepository) Update(ctx context.Context, id int64, public, allowReleaseOverwrite *bool, description *string) (Artifact, error) {
	return scanArtifact(r.db.QueryRowContext(ctx, `
		UPDATE artifacts SET
			is_public = COALESCE(?, is_public),
			allow_release_overwrite = COALESCE(?, allow_release_overwrite),
			description = CASE WHEN ? THEN ? ELSE description END
		WHERE id = ? RETURNING `+artifactColumns,
		public, allowReleaseOverwrite, description != nil, toNullString(description), id))
}

// Delete unregisters the artifact. Its files stay on disk but can no longer be downloaded.
func (r *ArtifactRepository) Delete(ctx context.Context, id int64) error {
	return requireAffected(r.db.ExecContext(ctx, `DELETE FROM artifacts WHERE id = ?`, id))
}

// FindByPath returns the artifact whose directory contains the repository path p, e.g. com/example/demo for
// com/example/demo/1.0/demo-1.0.jar. If artifact directories are nested, the deepest one wins.
func (r *ArtifactRepository) FindByPath(ctx context.Context, p string) (Artifact, error) {
	segments := strings.Split(strings.Trim(p, "/"), "/")
	// An artifact directory has at least one group segment and the artifactId
	if len(segments) < 2 {
		return Artifact{}, ErrNotFound
	}

	candidates := make([]any, 0, len(segments)-1)
	for i := 2; i <= len(segments); i++ {
		candidates = append(candidates, strings.Join(segments[:i], "/"))
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(candidates)), ", ")
	return scanArtifact(r.db.QueryRowContext(ctx, `
		SELECT `+artifactColumns+` FROM artifacts
		WHERE path IN (`+placeholders+`) ORDER BY length(path) DESC LIMIT 1`, candidates...))
}

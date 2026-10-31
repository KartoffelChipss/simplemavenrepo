package repository

import (
	"context"
	"database/sql"
	"time"
)

// Deploy is one attempt to upload a file of an artifact, successful or not.
type Deploy struct {
	ID       int64  `json:"id"`
	Version  string `json:"version"`
	File     string `json:"file"`
	Username string `json:"username"`
	// Status is the HTTP status the upload was answered with
	Status    int       `json:"status"`
	Success   bool      `json:"success"`
	Error     *string   `json:"error"`
	Size      *int64    `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

type DeployRepository struct {
	db *sql.DB
}

const deployColumns = `id, version, file, username, status, error, size, created_at`

func scanDeploy(s scanner) (Deploy, error) {
	var d Deploy
	var errMsg sql.NullString
	var size sql.NullInt64
	var createdAt int64
	if err := s.Scan(&d.ID, &d.Version, &d.File, &d.Username, &d.Status, &errMsg, &size, &createdAt); err != nil {
		return Deploy{}, mapErr(err)
	}
	d.Success = d.Status >= 200 && d.Status < 300
	if errMsg.Valid {
		d.Error = &errMsg.String
	}
	if size.Valid {
		d.Size = &size.Int64
	}
	d.CreatedAt = fromUnix(createdAt)
	return d, nil
}

// Record logs an upload to the artifact directory artifactPath. errMsg and size are optional.
func (r *DeployRepository) Record(ctx context.Context, artifactPath, version, file, username string, status int, errMsg string, size *int64) error {
	var nullSize sql.NullInt64
	if size != nil {
		nullSize = sql.NullInt64{Int64: *size, Valid: true}
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO deploys (artifact_path, version, file, username, status, error, size, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		artifactPath, version, file, username, status, toNullString(&errMsg), nullSize, toUnix(time.Now()))
	return err
}

// List returns up to limit deploys of the artifact directory artifactPath, newest first, skipping the first offset
// ones. total is the number of deploys of the artifact overall.
func (r *DeployRepository) List(ctx context.Context, artifactPath string, offset, limit int) (deploys []Deploy, total int, err error) {
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM deploys WHERE artifact_path = ?`, artifactPath).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT `+deployColumns+` FROM deploys WHERE artifact_path = ?
		ORDER BY id DESC LIMIT ? OFFSET ?`, artifactPath, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	deploys = []Deploy{}
	for rows.Next() {
		d, err := scanDeploy(rows)
		if err != nil {
			return nil, 0, err
		}
		deploys = append(deploys, d)
	}
	return deploys, total, rows.Err()
}

package repository

import (
	"context"
	"database/sql"
	"time"
)

// DayDownloads is the number of downloads on one UTC day.
type DayDownloads struct {
	// Date is formatted as 2006-01-02
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type VersionDownloads struct {
	Version string `json:"version"`
	Count   int64  `json:"count"`
}

type PathDownloads struct {
	ArtifactPath string
	Count        int64
}

type DownloadRepository struct {
	db *sql.DB
}

const dateLayout = "2006-01-02"

// Day truncates t to midnight of its day in UTC.
func Day(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

// Record counts one download of a version of the artifact directory artifactPath at the given time.
func (r *DownloadRepository) Record(ctx context.Context, artifactPath, version string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO downloads (artifact_path, version, day, count) VALUES (?, ?, ?, 1)
		ON CONFLICT (artifact_path, day, version) DO UPDATE SET count = count + 1`,
		artifactPath, version, toUnix(Day(at)))
	return err
}

// Total returns the number of downloads of the artifact directory artifactPath, or of all artifacts if it is empty,
// on the days from since on. A zero since counts all downloads.
func (r *DownloadRepository) Total(ctx context.Context, artifactPath string, since time.Time) (int64, error) {
	var total int64
	err := r.db.QueryRowContext(ctx, `
		SELECT coalesce(sum(count), 0) FROM downloads
		WHERE (? = '' OR artifact_path = ?) AND day >= ?`,
		artifactPath, artifactPath, sinceUnix(since)).Scan(&total)
	return total, err
}

// Daily returns the downloads of the artifact directory artifactPath, or of all artifacts if it is empty, for every
// day from from to to, including days without downloads.
func (r *DownloadRepository) Daily(ctx context.Context, artifactPath string, from, to time.Time) ([]DayDownloads, error) {
	from, to = Day(from), Day(to)
	rows, err := r.db.QueryContext(ctx, `
		SELECT day, sum(count) FROM downloads
		WHERE (? = '' OR artifact_path = ?) AND day BETWEEN ? AND ?
		GROUP BY day`,
		artifactPath, artifactPath, toUnix(from), toUnix(to))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[int64]int64{}
	for rows.Next() {
		var day, count int64
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		counts[day] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	days := []DayDownloads{}
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		days = append(days, DayDownloads{Date: d.Format(dateLayout), Count: counts[toUnix(d)]})
	}
	return days, nil
}

// ByVersion returns the downloads of each version of the artifact directory artifactPath on the days from since on,
// most downloaded first. Versions without downloads are missing.
func (r *DownloadRepository) ByVersion(ctx context.Context, artifactPath string, since time.Time) ([]VersionDownloads, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT version, sum(count) AS total FROM downloads
		WHERE artifact_path = ? AND day >= ?
		GROUP BY version ORDER BY total DESC, version`,
		artifactPath, sinceUnix(since))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := []VersionDownloads{}
	for rows.Next() {
		var v VersionDownloads
		if err := rows.Scan(&v.Version, &v.Count); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// ByArtifact returns the downloads of each artifact directory on the days from since on, most downloaded first.
func (r *DownloadRepository) ByArtifact(ctx context.Context, since time.Time) ([]PathDownloads, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT artifact_path, sum(count) AS total FROM downloads
		WHERE day >= ?
		GROUP BY artifact_path ORDER BY total DESC, artifact_path`,
		sinceUnix(since))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	paths := []PathDownloads{}
	for rows.Next() {
		var p PathDownloads
		if err := rows.Scan(&p.ArtifactPath, &p.Count); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}

func sinceUnix(since time.Time) int64 {
	if since.IsZero() {
		return 0
	}
	return toUnix(Day(since))
}

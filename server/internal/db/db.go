package db

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "modernc.org/sqlite"
)

// Each entry upgrades the schema by one version. Never edit an existing entry, only append new ones.
var migrations = []string{
	`
	CREATE TABLE users (
		id         INTEGER PRIMARY KEY,
		username   TEXT    NOT NULL UNIQUE,
		is_admin   INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL
	);

	CREATE TABLE teams (
		id         INTEGER PRIMARY KEY,
		name       TEXT    NOT NULL UNIQUE,
		created_at INTEGER NOT NULL
	);

	CREATE TABLE team_members (
		team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		PRIMARY KEY (team_id, user_id)
	);
	CREATE INDEX team_members_user ON team_members(user_id);

	CREATE TABLE tokens (
		id           INTEGER PRIMARY KEY,
		user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name         TEXT    NOT NULL,
		prefix       TEXT    NOT NULL,
		token_hash   BLOB    NOT NULL UNIQUE,
		created_at   INTEGER NOT NULL,
		expires_at   INTEGER,
		last_used_at INTEGER
	);
	CREATE INDEX tokens_user ON tokens(user_id);

	CREATE TABLE grants (
		id         INTEGER PRIMARY KEY,
		user_id    INTEGER REFERENCES users(id) ON DELETE CASCADE,
		team_id    INTEGER REFERENCES teams(id) ON DELETE CASCADE,
		pattern    TEXT    NOT NULL,
		permission TEXT    NOT NULL CHECK (permission IN ('read', 'write')),
		created_at INTEGER NOT NULL,
		CHECK ((user_id IS NULL) != (team_id IS NULL))
	);
	-- Separate partial indexes because NULLs never collide in a UNIQUE constraint
	CREATE UNIQUE INDEX grants_user ON grants(user_id, pattern, permission) WHERE user_id IS NOT NULL;
	CREATE UNIQUE INDEX grants_team ON grants(team_id, pattern, permission) WHERE team_id IS NOT NULL;
	`,
	`
	CREATE TABLE artifacts (
		id          INTEGER PRIMARY KEY,
		group_id    TEXT    NOT NULL,
		artifact_id TEXT    NOT NULL,
		-- Directory of the artifact in the repository, e.g. com/example/demo
		path        TEXT    NOT NULL GENERATED ALWAYS AS (replace(group_id, '.', '/') || '/' || artifact_id) STORED UNIQUE,
		is_public   INTEGER NOT NULL DEFAULT 0,
		description TEXT,
		created_at  INTEGER NOT NULL
	);
	`,
	`
	ALTER TABLE users ADD COLUMN password_hash BLOB;

	CREATE TABLE sessions (
		id           INTEGER PRIMARY KEY,
		user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		session_hash BLOB    NOT NULL UNIQUE,
		created_at   INTEGER NOT NULL,
		expires_at   INTEGER NOT NULL
	);
	CREATE INDEX sessions_user ON sessions(user_id);
	`,
	`
	ALTER TABLE artifacts ADD COLUMN allow_release_overwrite INTEGER NOT NULL DEFAULT 0;
	`,
	`
	-- Keyed by the artifact directory instead of its id so the log survives deleting and re-registering the artifact, just like its files. The username is copied so entries stay readable after the user is deleted.
	CREATE TABLE deploys (
		id            INTEGER PRIMARY KEY,
		artifact_path TEXT    NOT NULL,
		version       TEXT    NOT NULL,
		file          TEXT    NOT NULL,
		username      TEXT    NOT NULL,
		status        INTEGER NOT NULL,
		error         TEXT,
		size          INTEGER,
		created_at    INTEGER NOT NULL
	);
	CREATE INDEX deploys_artifact ON deploys(artifact_path, id);
	`,
}

// Open opens the SQLite database at path and migrates it to the latest schema.
func Open(path string) (*sql.DB, error) {
	q := url.Values{}
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "journal_mode(WAL)")

	db, err := sql.Open("sqlite", "file:"+path+"?"+q.Encode())
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version > len(migrations) {
		return fmt.Errorf("database schema version %d is newer than this server supports (%d)", version, len(migrations))
	}

	for i := version; i < len(migrations); i++ {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(migrations[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
)

// The repositories migration moves everything of the single repository of older versions into "releases".
func TestMigrateToRepositories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	const before = 8
	for i, m := range migrations[:before] {
		if _, err := conn.Exec(m); err != nil {
			t.Fatalf("migration %d: %v", i+1, err)
		}
	}
	for _, stmt := range []string{
		fmt.Sprintf("PRAGMA user_version = %d", before),
		`INSERT INTO users (id, username, created_at) VALUES (1, 'alice', 0)`,
		`INSERT INTO artifacts (id, group_id, artifact_id, is_public, description, created_at) VALUES (7, 'com.example', 'demo', 1, 'Demo', 0)`,
		`INSERT INTO grants (user_id, pattern, permission, created_at) VALUES (1, 'com/example/**', 'write', 0), (1, '**', 'read', 0)`,
		`INSERT INTO tokens (id, user_id, name, prefix, token_hash, created_at) VALUES (1, 1, 'ci', 'smr_', x'00', 0)`,
		`INSERT INTO token_scopes (token_id, pattern, permission) VALUES (1, 'com/example/demo/**', 'write')`,
		`INSERT INTO deploys (artifact_path, version, file, username, status, created_at) VALUES ('com/example/demo', '1.0', 'demo-1.0.jar', 'alice', 201, 0)`,
		`INSERT INTO downloads (artifact_path, version, day, count) VALUES ('com/example/demo', '1.0', 0, 3)`,
	} {
		if _, err := conn.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	conn.Close()

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var repo, artifactPath, description string
	var id int64
	var public bool
	if err := db.QueryRow(`SELECT id, repository, path, is_public, description FROM artifacts`).Scan(&id, &repo, &artifactPath, &public, &description); err != nil {
		t.Fatal(err)
	}
	if id != 7 || repo != "releases" || artifactPath != "releases/com/example/demo" || !public || description != "Demo" {
		t.Errorf("artifact = %d %s %s %v %s", id, repo, artifactPath, public, description)
	}

	for query, want := range map[string]string{
		`SELECT group_concat(pattern, ' ') FROM (SELECT pattern FROM grants ORDER BY pattern)`: "** releases/com/example/**",
		`SELECT pattern FROM token_scopes`:                       "releases/com/example/demo/**",
		`SELECT artifact_path FROM deploys`:                      "releases/com/example/demo",
		`SELECT artifact_path FROM downloads`:                    "releases/com/example/demo",
		`SELECT name || ' ' || version_policy FROM repositories`: "releases mixed",
	} {
		var got string
		if err := db.QueryRow(query).Scan(&got); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
		if got != want {
			t.Errorf("%s = %q, want %q", query, got, want)
		}
	}
}

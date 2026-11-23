package ressourcehandler

import (
	"context"
	"encoding/base64"
	"io"
	"io/fs"
	"net/http/httptest"
	"os"
	"path/filepath"
	"simplemavenrepo/internal/db"
	"simplemavenrepo/internal/repository"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

type testServer struct {
	app   *fiber.App
	repos repository.Repositories
	dir   string
	auth  string
}

func newTestServer(t *testing.T) testServer {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	t.Setenv("SMR_REPO_PATH", dir)

	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	repos := repository.New(conn)
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })

	admin, err := repos.Users.Create(ctx, "admin", true)
	if err != nil {
		t.Fatal(err)
	}
	_, secret, err := repos.Tokens.Create(ctx, admin.ID, "test", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repos.Repos.Create(ctx, "snapshots", repository.RepoOptions{VersionPolicy: ptr(repository.PolicySnapshot)}); err != nil {
		t.Fatal(err)
	}
	for _, repo := range []string{"releases", "snapshots"} {
		if _, err := repos.Artifacts.Create(ctx, repo, "com.example", "demo", true, false, nil); err != nil {
			t.Fatal(err)
		}
	}

	app := fiber.New()
	RegisterRessourceRoutes(app, repos, root)
	// Stands in for the web UI, which gets every path that isn't a repository
	app.Get("/*", func(c fiber.Ctx) error { return c.SendString("web ui") })
	return testServer{app: app, repos: repos, dir: dir, auth: "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:"+secret))}
}

func (s testServer) do(t *testing.T, method, target, body string) (int, string) {
	t.Helper()
	status, _, got := s.doWithStatus(t, method, target, body)
	return status, got
}

// doWithStatus also returns the status line text after the code, the reason phrase.
func (s testServer) doWithStatus(t *testing.T, method, target, body string) (int, string, string) {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if method == fiber.MethodPut {
		req.Header.Set(fiber.HeaderAuthorization, s.auth)
	}
	res, err := s.app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	return res.StatusCode, strings.TrimSpace(strings.TrimPrefix(res.Status, strconv.Itoa(res.StatusCode))), string(data)
}

func TestRepositoryRoutes(t *testing.T) {
	s := newTestServer(t)

	const release = "/com/example/demo/1.0/demo-1.0.jar"
	const snapshot = "/com/example/demo/1.0-SNAPSHOT/demo-1.0-20260101.120000-1.jar"
	tests := []struct {
		method, target string
		wantStatus     int
		wantBody       string
	}{
		{fiber.MethodPut, "/releases" + release, fiber.StatusCreated, ""},
		{fiber.MethodGet, "/releases" + release, fiber.StatusOK, "release"},
		{fiber.MethodPut, "/snapshots" + release, fiber.StatusBadRequest, "only accepts SNAPSHOT versions"},
		{fiber.MethodPut, "/snapshots" + snapshot, fiber.StatusCreated, ""},
		{fiber.MethodGet, "/snapshots" + snapshot, fiber.StatusOK, "snapshot"},
		// Signatures are stored next to the files they sign
		{fiber.MethodPut, "/releases" + release + ".asc", fiber.StatusCreated, ""},
		{fiber.MethodGet, "/releases" + release + ".asc", fiber.StatusOK, "release"},
		{fiber.MethodPut, "/releases" + release + ".asc.sha1", fiber.StatusNoContent, ""},
		{fiber.MethodPut, "/releases" + release + ".asc", fiber.StatusConflict, "already published"},
		{fiber.MethodPut, "/snapshots" + snapshot + ".asc", fiber.StatusCreated, ""},
		{fiber.MethodPut, "/releases/com/example/demo/1.0/demo-1.0.txt.asc", fiber.StatusUnsupportedMediaType, ""},
		// The same coordinates in another repository are a different artifact
		{fiber.MethodGet, "/snapshots" + release, fiber.StatusNotFound, ""},
		{fiber.MethodGet, "/releases/com/example/demo/2.0/demo-2.0.jar", fiber.StatusNotFound, ""},
		{fiber.MethodGet, "/releases/com/example/other/1.0/other-1.0.jar", fiber.StatusNotFound, ""},
		// Uploads never get a 404, Maven would only say "Could not find artifact"
		{fiber.MethodPut, "/missing" + release, fiber.StatusBadRequest, "no repository named missing"},
		{fiber.MethodPut, "/releases/com/example/other/1.0/other-1.0.jar", fiber.StatusForbidden, "com.example:other is not registered in repository releases"},
		{fiber.MethodGet, "/missing" + release, fiber.StatusOK, "web ui"},
		{fiber.MethodGet, "/browse/releases/com/example/demo", fiber.StatusOK, "web ui"},
		{fiber.MethodGet, "/", fiber.StatusOK, "web ui"},
	}
	for _, tt := range tests {
		body := "release"
		if strings.Contains(tt.target, "SNAPSHOT") {
			body = "snapshot"
		}
		status, got := s.do(t, tt.method, tt.target, body)
		if status != tt.wantStatus || !strings.Contains(got, tt.wantBody) {
			t.Errorf("%s %s = %d %q, want %d containing %q", tt.method, tt.target, status, got, tt.wantStatus, tt.wantBody)
		}
	}

	// Maven and Gradle only show the reason phrase of a failed upload
	status, reason, _ := s.doWithStatus(t, fiber.MethodPut, "/snapshots"+release, "x")
	if status != fiber.StatusBadRequest || reason != "repository snapshots only accepts SNAPSHOT versions, not 1.0" {
		t.Errorf("reason phrase = %d %q", status, reason)
	}

	// Files and metadata are stored in the directory of their repository
	for _, f := range []string{
		"releases" + release,
		"releases/com/example/demo/maven-metadata.xml",
		"snapshots" + snapshot,
		"snapshots/com/example/demo/1.0-SNAPSHOT/maven-metadata.xml",
	} {
		if _, err := fs.Stat(os.DirFS(s.dir), f); err != nil {
			t.Errorf("%s: %v", f, err)
		}
	}
	snapshotMetadata, err := os.ReadFile(filepath.Join(s.dir, "snapshots/com/example/demo/1.0-SNAPSHOT/maven-metadata.xml"))
	if err != nil || !strings.Contains(string(snapshotMetadata), "<extension>jar.asc</extension>") {
		t.Errorf("the snapshot metadata must list the signature: %s, %v", snapshotMetadata, err)
	}
	metadata, err := os.ReadFile(filepath.Join(s.dir, "releases/com/example/demo/maven-metadata.xml"))
	if err != nil || !strings.Contains(string(metadata), "<groupId>com.example</groupId>") {
		t.Errorf("the metadata must not contain the repository in the groupId: %s, %v", metadata, err)
	}
}

func TestAutoRegister(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()
	if _, err := s.repos.Repos.Create(ctx, "public", repository.RepoOptions{AutoRegister: ptr(true), DefaultPublic: ptr(true)}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.repos.Repos.Create(ctx, "private", repository.RepoOptions{VersionPolicy: ptr(repository.PolicyRelease), AutoRegister: ptr(true)}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		method, target string
		wantStatus     int
		wantBody       string
	}{
		// Checksums, signatures and metadata don't register an artifact, only the files they belong to
		{fiber.MethodPut, "/public/org/acme/lib/1.0/lib-1.0.jar.sha1", fiber.StatusForbidden, "org.acme:lib is not registered"},
		{fiber.MethodPut, "/public/org/acme/lib/1.0/lib-1.0.jar.asc", fiber.StatusForbidden, "org.acme:lib is not registered"},
		{fiber.MethodPut, "/public/org/acme/lib/maven-metadata.xml", fiber.StatusForbidden, "is not registered"},
		// A rejected upload doesn't register it either
		{fiber.MethodPut, "/private/org/acme/lib/1.0-SNAPSHOT/lib-1.0-20260101.120000-1.jar", fiber.StatusBadRequest, "only accepts release versions"},
		{fiber.MethodPut, "/private/org/acme/lib/1.0/lib-1.0.txt", fiber.StatusUnsupportedMediaType, ""},
		{fiber.MethodPut, "/public/org/acme/lib/1.0/lib-1.0.jar", fiber.StatusCreated, ""},
		{fiber.MethodPut, "/public/org/acme/lib/1.0/lib-1.0.jar.sha1", fiber.StatusNoContent, ""},
		{fiber.MethodPut, "/public/org/acme/lib/1.0/lib-1.0.jar.asc", fiber.StatusCreated, ""},
		{fiber.MethodPut, "/public/org/acme/lib/1.0/lib-1.0.pom", fiber.StatusCreated, ""},
		{fiber.MethodGet, "/public/org/acme/lib/1.0/lib-1.0.jar", fiber.StatusOK, "release"},
		{fiber.MethodPut, "/private/org/acme/lib/1.0/lib-1.0.jar", fiber.StatusCreated, ""},
		// Registered as private, so downloading needs a token
		{fiber.MethodGet, "/private/org/acme/lib/1.0/lib-1.0.jar", fiber.StatusUnauthorized, ""},
		// Repositories without auto registration still require registering first
		{fiber.MethodPut, "/releases/org/acme/lib/1.0/lib-1.0.jar", fiber.StatusForbidden, "is not registered"},
	}
	for _, tt := range tests {
		status, got := s.do(t, tt.method, tt.target, "release")
		if status != tt.wantStatus || !strings.Contains(got, tt.wantBody) {
			t.Errorf("%s %s = %d %q, want %d containing %q", tt.method, tt.target, status, got, tt.wantStatus, tt.wantBody)
		}
	}

	for repo, wantPublic := range map[string]bool{"public": true, "private": false} {
		a, err := s.repos.Artifacts.Get(ctx, repo, "org.acme", "lib")
		if err != nil || a.Public != wantPublic || a.AllowReleaseOverwrite {
			t.Errorf("artifact in %s = %+v, %v, want public %v", repo, a, err, wantPublic)
		}
		if entries, _, err := s.repos.Deploys.List(ctx, a.Path, 0, 10); err != nil || len(entries) == 0 {
			t.Errorf("deploy log of %s = %v, %v, want the registering upload", a.Path, entries, err)
		}
	}
	if _, err := s.repos.Artifacts.Get(ctx, "releases", "org.acme", "lib"); err == nil {
		t.Error("the artifact must not be registered in releases")
	}
}

func ptr[T any](v T) *T { return &v }

func TestReasonPhrase(t *testing.T) {
	if got := reasonPhrase("line\r\nbreak – ü"); got != "line  break    " {
		t.Errorf("reasonPhrase = %q", got)
	}
	if got := reasonPhrase(strings.Repeat("a", 500)); len(got) != maxReasonLength {
		t.Errorf("reasonPhrase length = %d, want %d", len(got), maxReasonLength)
	}
}

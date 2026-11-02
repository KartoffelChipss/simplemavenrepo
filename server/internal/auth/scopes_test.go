package auth

import (
	"context"
	"io"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"simplemavenrepo/internal/db"
	"simplemavenrepo/internal/repository"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestTokenScopes(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	repos := repository.New(conn)
	ctx := context.Background()

	alice, err := repos.Users.Create(ctx, "alice", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repos.Grants.CreateForUser(ctx, alice.ID, "com/example/**", repository.PermissionWrite); err != nil {
		t.Fatal(err)
	}
	admin, err := repos.Users.Create(ctx, "admin", true)
	if err != nil {
		t.Fatal(err)
	}
	token := func(u repository.User, scopes ...repository.TokenScope) string {
		_, plaintext, err := repos.Tokens.Create(ctx, u.ID, "test", nil, scopes)
		if err != nil {
			t.Fatal(err)
		}
		return plaintext
	}
	full := token(alice)
	libOnly := token(alice, repository.TokenScope{Pattern: "com/example/lib/**", Permission: repository.PermissionWrite})
	outsideGrants := token(alice, repository.TokenScope{Pattern: "org/other/**", Permission: repository.PermissionWrite})
	adminFull := token(admin)
	adminScoped := token(admin, repository.TokenScope{Pattern: "com/example/api/**", Permission: repository.PermissionRead})

	app := fiber.New()
	app.Get("/account", Authenticate(repos), func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	app.Get("/admin", Authenticate(repos), RequireAdmin, func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	app.Get("/repo", AuthenticateScoped(repos), func(c fiber.Ctx) error {
		allowed, err := Authorize(c, repos.Grants, c.Query("path"), repository.Permission(c.Query("perm")))
		if err != nil {
			return err
		}
		return c.SendString(strconv.FormatBool(allowed))
	})

	get := func(target, token string) (int, string) {
		t.Helper()
		req := httptest.NewRequest(fiber.MethodGet, target, nil)
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		return res.StatusCode, string(body)
	}

	accountTests := []struct {
		name, target, token string
		want                int
	}{
		{"unscoped token manages the account", "/account", full, fiber.StatusOK},
		{"scoped token can't manage the account", "/account", libOnly, fiber.StatusForbidden},
		{"unscoped admin token", "/admin", adminFull, fiber.StatusOK},
		{"scoped admin token is no admin", "/admin", adminScoped, fiber.StatusForbidden},
	}
	for _, tt := range accountTests {
		if code, body := get(tt.target, tt.token); code != tt.want {
			t.Errorf("%s: got %d %s; want %d", tt.name, code, body, tt.want)
		}
	}

	repoTests := []struct {
		name, token, path string
		perm              repository.Permission
		want              bool
	}{
		{"unscoped token keeps the grants", full, "com/example/api/1.0/api-1.0.jar", repository.PermissionWrite, true},
		{"scope allows its artifact", libOnly, "com/example/lib/1.0/lib-1.0.jar", repository.PermissionWrite, true},
		{"write scope implies read", libOnly, "com/example/lib/maven-metadata.xml", repository.PermissionRead, true},
		{"scope blocks other granted artifacts", libOnly, "com/example/api/1.0/api-1.0.jar", repository.PermissionRead, false},
		{"scope can't exceed the grants", outsideGrants, "org/other/lib/1.0/lib-1.0.jar", repository.PermissionRead, false},
		{"unscoped admin token", adminFull, "org/other/lib/1.0/lib-1.0.jar", repository.PermissionWrite, true},
		{"scoped admin token within scope", adminScoped, "com/example/api/1.0/api-1.0.jar", repository.PermissionRead, true},
		{"scoped admin token can't write", adminScoped, "com/example/api/1.0/api-1.0.jar", repository.PermissionWrite, false},
		{"scoped admin token outside scope", adminScoped, "org/other/lib/1.0/lib-1.0.jar", repository.PermissionRead, false},
	}
	for _, tt := range repoTests {
		q := url.Values{"path": {tt.path}, "perm": {string(tt.perm)}}
		code, body := get("/repo?"+q.Encode(), tt.token)
		if code != fiber.StatusOK || body != strconv.FormatBool(tt.want) {
			t.Errorf("%s: got %d %s; want %v", tt.name, code, body, tt.want)
		}
	}
}

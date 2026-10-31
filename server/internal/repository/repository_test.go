package repository

import (
	"context"
	"errors"
	"path/filepath"
	"simplemavenrepo/internal/db"
	"testing"
	"time"
)

func newTestRepos(t *testing.T) Repositories {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return New(conn)
}

func TestRepositories(t *testing.T) {
	ctx := context.Background()
	r := newTestRepos(t)

	alice, err := r.Users.Create(ctx, "alice", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Users.Create(ctx, "alice", true); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate user: got %v, want ErrConflict", err)
	}

	team, err := r.Teams.Create(ctx, "devs")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Teams.AddMember(ctx, team.ID, alice.ID); err != nil {
		t.Fatal(err)
	}
	if err := r.Teams.AddMember(ctx, team.ID, alice.ID); err != nil {
		t.Fatalf("re-adding member should be a no-op: %v", err)
	}

	if _, err := r.Grants.CreateForUser(ctx, alice.ID, "com/alice/**", PermissionWrite); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Grants.CreateForTeam(ctx, team.ID, "com/example/**", PermissionRead); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Grants.CreateForTeam(ctx, team.ID, "com/example/**", PermissionRead); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate grant: got %v, want ErrConflict", err)
	}
	effective, err := r.Grants.ListEffective(ctx, alice.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(effective) != 2 || effective[0].User != "alice" || effective[1].Team != "devs" {
		t.Fatalf("unexpected effective grants: %+v", effective)
	}

	tok, plaintext, err := r.Tokens.Create(ctx, alice.ID, "laptop", nil)
	if err != nil {
		t.Fatal(err)
	}
	if u, err := r.Tokens.Authenticate(ctx, plaintext); err != nil || u.ID != alice.ID {
		t.Fatalf("Authenticate = %+v, %v", u, err)
	}
	if _, err := r.Tokens.Authenticate(ctx, plaintext+"x"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong token: got %v, want ErrInvalidCredentials", err)
	}
	tokens, err := r.Tokens.ListForUser(ctx, alice.ID)
	if err != nil || len(tokens) != 1 || tokens[0].LastUsedAt == nil {
		t.Fatalf("ListForUser = %+v, %v; want one used token", tokens, err)
	}

	past := time.Now().Add(-time.Hour)
	_, expired, err := r.Tokens.Create(ctx, alice.ID, "old", &past)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Tokens.Authenticate(ctx, expired); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expired token: got %v, want ErrInvalidCredentials", err)
	}

	if err := r.Tokens.Delete(ctx, alice.ID+1, tok.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleting another user's token: got %v, want ErrNotFound", err)
	}

	// Deleting the user cascades to tokens, memberships and grants
	if err := r.Users.Delete(ctx, "alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Tokens.Authenticate(ctx, plaintext); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("token of deleted user: got %v, want ErrInvalidCredentials", err)
	}
	if members, _ := r.Teams.ListMembers(ctx, team.ID); len(members) != 0 {
		t.Fatalf("members after user deletion: %+v", members)
	}
	if grants, _ := r.Grants.List(ctx, GrantFilter{}); len(grants) != 1 || grants[0].Team != "devs" {
		t.Fatalf("grants after user deletion: %+v", grants)
	}
}

func TestArtifacts(t *testing.T) {
	ctx := context.Background()
	r := newTestRepos(t)

	desc := "Demo library"
	demo, err := r.Artifacts.Create(ctx, "com.example", "demo", true, &desc)
	if err != nil {
		t.Fatal(err)
	}
	if demo.Path != "com/example/demo" || !demo.Public || demo.Description == nil || *demo.Description != desc {
		t.Fatalf("unexpected artifact: %+v", demo)
	}
	if _, err := r.Artifacts.Create(ctx, "com.example", "demo", false, nil); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate artifact: got %v, want ErrConflict", err)
	}
	// Different coordinates that map to the same directory
	if _, err := r.Artifacts.Create(ctx, "com", "example/demo", false, nil); !errors.Is(err, ErrConflict) {
		t.Fatalf("artifact with the same path: got %v, want ErrConflict", err)
	}
	nested, err := r.Artifacts.Create(ctx, "com.example.demo", "sub", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]int64{
		"com/example/demo/1.0/demo-1.0.jar":    demo.ID,
		"com/example/demo/maven-metadata.xml":  demo.ID,
		"/com/example/demo":                    demo.ID,
		"com/example/demo/sub/1.0/sub-1.0.pom": nested.ID,
		"com/example/other/1.0/other-1.0.jar":  0,
		"com/example":                          0,
		"com":                                  0,
	}
	for p, want := range tests {
		a, err := r.Artifacts.FindByPath(ctx, p)
		if want == 0 {
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("FindByPath(%q) = %+v, %v; want ErrNotFound", p, a, err)
			}
		} else if err != nil || a.ID != want {
			t.Errorf("FindByPath(%q) = %+v, %v; want id %d", p, a, err, want)
		}
	}

	private := false
	updated, err := r.Artifacts.Update(ctx, demo.ID, &private, nil)
	if err != nil || updated.Public || updated.Description == nil {
		t.Fatalf("Update visibility = %+v, %v", updated, err)
	}
	empty := ""
	updated, err = r.Artifacts.Update(ctx, demo.ID, nil, &empty)
	if err != nil || updated.Public || updated.Description != nil {
		t.Fatalf("Update clear description = %+v, %v", updated, err)
	}

	if err := r.Artifacts.Delete(ctx, demo.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Artifacts.FindByPath(ctx, "com/example/demo/1.0/demo-1.0.jar"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("FindByPath after delete: got %v, want ErrNotFound", err)
	}
}

func TestPasswordsAndSessions(t *testing.T) {
	ctx := context.Background()
	r := newTestRepos(t)

	alice, err := r.Users.Create(ctx, "alice", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Users.CheckPassword(ctx, "alice", ""); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("user without password: got %v, want ErrInvalidCredentials", err)
	}
	if err := r.Users.SetPassword(ctx, alice.ID, "short"); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("short password: got %v, want ErrInvalidPassword", err)
	}
	if err := r.Users.SetPassword(ctx, alice.ID, "correct horse"); err != nil {
		t.Fatal(err)
	}
	u, err := r.Users.CheckPassword(ctx, "alice", "correct horse")
	if err != nil || !u.HasPassword || u.ID != alice.ID {
		t.Fatalf("correct password: got %+v, %v", u, err)
	}
	if _, err := r.Users.CheckPassword(ctx, "alice", "wrong horse"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password: got %v, want ErrInvalidCredentials", err)
	}
	if _, err := r.Users.CheckPassword(ctx, "bob", "correct horse"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unknown user: got %v, want ErrInvalidCredentials", err)
	}

	first, _, err := r.Sessions.Create(ctx, alice.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := r.Sessions.Create(ctx, alice.ID)
	if err != nil {
		t.Fatal(err)
	}
	if u, err := r.Sessions.Authenticate(ctx, first); err != nil || u.Username != "alice" {
		t.Fatalf("session: got %+v, %v", u, err)
	}
	if err := r.Sessions.DeleteForUser(ctx, alice.ID, second); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Sessions.Authenticate(ctx, first); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("revoked session: got %v, want ErrInvalidCredentials", err)
	}
	if _, err := r.Sessions.Authenticate(ctx, second); err != nil {
		t.Fatalf("kept session: %v", err)
	}
	if err := r.Sessions.Delete(ctx, second); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Sessions.Authenticate(ctx, second); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("logged out session: got %v, want ErrInvalidCredentials", err)
	}
}

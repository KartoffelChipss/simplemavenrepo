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

package auth

import (
	"encoding/base64"
	"simplemavenrepo/internal/repository"
	"testing"
)

func TestAllowed(t *testing.T) {
	grants := []repository.Grant{
		{Pattern: "com/example/**", Permission: repository.PermissionRead},
		{Pattern: "com/example/lib/**", Permission: repository.PermissionWrite},
	}
	tests := []struct {
		path string
		perm repository.Permission
		want bool
	}{
		{"com/example/other/1.0/other-1.0.jar", repository.PermissionRead, true},
		{"com/example/other/1.0/other-1.0.jar", repository.PermissionWrite, false},
		{"com/example/lib/1.0/lib-1.0.jar", repository.PermissionWrite, true},
		{"com/example/lib/1.0/lib-1.0.jar", repository.PermissionRead, true},
		{"org/other/lib/1.0/lib-1.0.jar", repository.PermissionRead, false},
	}
	for _, tt := range tests {
		if got := Allowed(grants, tt.path, tt.perm); got != tt.want {
			t.Errorf("Allowed(%q, %s) = %v; want %v", tt.path, tt.perm, got, tt.want)
		}
	}
	if Allowed(nil, "com/example/lib", repository.PermissionRead) {
		t.Error("no grants should allow nothing")
	}
}

func TestParseAuthorization(t *testing.T) {
	basic := func(s string) string { return "Basic " + base64.StdEncoding.EncodeToString([]byte(s)) }
	tests := []struct {
		header, user, token string
		ok                  bool
	}{
		{"Bearer smr_abc", "", "smr_abc", true},
		{"bearer  smr_abc ", "", "smr_abc", true},
		{basic("alice:smr_abc"), "alice", "smr_abc", true},
		{basic("alice:smr:abc"), "alice", "smr:abc", true},
		{basic("alice"), "", "", false},
		{basic(":smr_abc"), "", "", false},
		{"Basic !!!", "", "", false},
		{"Bearer ", "", "", false},
		{"Digest x", "", "", false},
		{"", "", "", false},
	}
	for _, tt := range tests {
		user, token, ok := parseAuthorization(tt.header)
		if ok != tt.ok || (ok && (user != tt.user || token != tt.token)) {
			t.Errorf("parseAuthorization(%q) = %q, %q, %v; want %q, %q, %v", tt.header, user, token, ok, tt.user, tt.token, tt.ok)
		}
	}
}

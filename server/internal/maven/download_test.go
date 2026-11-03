package maven

import "testing"

func TestIsMainFile(t *testing.T) {
	exts := []string{".jar", ".aar", ".tar.gz"}
	for _, tc := range []struct {
		version, file string
		want          bool
	}{
		{"1.0", "demo-1.0.jar", true},
		{"1.0", "demo-1.0.aar", true},
		{"1.0", "demo-1.0.tar.gz", true},
		{"1.0", "demo-1.0-sources.jar", false},
		{"1.0", "demo-1.0-javadoc.jar", false},
		{"1.0", "demo-1.0.pom", false},
		{"1.0", "demo-1.0.jar.sha1", false},
		{"1.0", "demo-1.0.jar.asc", false},
		{"1.0", "other-1.0.jar", false},
		{"1.0", "demo-1.1.jar", false},
		{"1.0-SNAPSHOT", "demo-1.0-SNAPSHOT.jar", true},
		{"1.0-SNAPSHOT", "demo-1.0-20260101.120000-1.jar", true},
		{"1.0-SNAPSHOT", "demo-1.0-20260101.120000-12.jar", true},
		{"1.0-SNAPSHOT", "demo-1.0-20260101.120000-1-sources.jar", false},
		{"1.0-SNAPSHOT", "demo-1.1-20260101.120000-1.jar", false},
	} {
		if got := IsMainFile("demo", tc.version, tc.file, exts); got != tc.want {
			t.Errorf("IsMainFile(%q, %q) = %v, want %v", tc.version, tc.file, got, tc.want)
		}
	}
}

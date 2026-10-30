package repository

import "testing"

func TestNormalizePattern(t *testing.T) {
	valid := map[string]string{
		"**":                "**",
		"/**":               "**",
		"com/example":       "com/example/**",
		"com/example/**":    "com/example/**",
		"/com/example/lib/": "com/example/lib/**",
		"org/foo/my.lib/**": "org/foo/my.lib/**",
	}
	for in, want := range valid {
		got, err := NormalizePattern(in)
		if err != nil || got != want {
			t.Errorf("NormalizePattern(%q) = %q, %v; want %q", in, got, err, want)
		}
	}

	for _, in := range []string{"", "/", "com//example", "com/../etc", "com/*/lib", "com/**/lib", `com\example`, "./com"} {
		if got, err := NormalizePattern(in); err == nil {
			t.Errorf("NormalizePattern(%q) = %q; want error", in, got)
		}
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern, path string
		want          bool
	}{
		{"**", "com/example/lib/1.0/lib-1.0.jar", true},
		{"com/example/**", "com/example", true},
		{"com/example/**", "com/example/lib/1.0/lib-1.0.jar", true},
		{"com/example/**", "/com/example/lib/maven-metadata.xml", true},
		{"com/example/**", "com/examples/lib", false},
		{"com/example/**", "com", false},
		{"com/example/lib/**", "com/example/other/1.0/other-1.0.jar", false},
	}
	for _, tt := range tests {
		if got := MatchPattern(tt.pattern, tt.path); got != tt.want {
			t.Errorf("MatchPattern(%q, %q) = %v; want %v", tt.pattern, tt.path, got, tt.want)
		}
	}
}

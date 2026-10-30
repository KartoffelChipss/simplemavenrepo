package repository

import (
	"errors"
	"strings"
)

// MatchAll is the pattern that matches every path in the repository.
const MatchAll = "**"

// NormalizePattern validates a grant pattern and returns its canonical form.
// A pattern names a repository subtree: "com/example/**" or "com/example" both match com/example and everything below it.
func NormalizePattern(p string) (string, error) {
	p = strings.Trim(strings.TrimSpace(p), "/")
	if p == MatchAll {
		return MatchAll, nil
	}
	p = strings.TrimSuffix(p, "/"+MatchAll)
	if p == "" || strings.Contains(p, "\\") {
		return "", errors.New("invalid pattern")
	}

	for _, s := range strings.Split(p, "/") {
		if s == "" || s == "." || s == ".." || strings.Contains(s, "*") {
			return "", errors.New(`invalid pattern: use a path like "com/example/**"`)
		}
	}
	return p + "/" + MatchAll, nil
}

// MatchPattern reports whether a normalized pattern covers the repository path p.
func MatchPattern(pattern, p string) bool {
	if pattern == MatchAll {
		return true
	}
	prefix := strings.TrimSuffix(pattern, "/"+MatchAll)
	p = strings.Trim(p, "/")
	return p == prefix || strings.HasPrefix(p, prefix+"/")
}

package maven

import (
	"regexp"
	"strings"
)

var snapshotSuffix = regexp.MustCompile(`^-\d{8}\.\d{6}-\d+$`)

// IsMainFile reports whether file is the main file of a version of the artifact, i.e. has no classifier like -sources
// and one of the extensions. Snapshot files may carry a timestamp instead of -SNAPSHOT, e.g. demo-1.0-20260101.120000-1.jar
// in 1.0-SNAPSHOT.
func IsMainFile(artifactID, version, file string, extensions []string) bool {
	rest, ok := strings.CutPrefix(file, artifactID+"-")
	if !ok {
		return false
	}
	hasExt := false
	for _, ext := range extensions {
		if trimmed, ok := strings.CutSuffix(rest, ext); ok {
			rest, hasExt = trimmed, true
			break
		}
	}
	if !hasExt {
		return false
	}
	if rest == version {
		return true
	}
	base, ok := strings.CutSuffix(version, "-SNAPSHOT")
	if !ok {
		return false
	}
	stamp, ok := strings.CutPrefix(rest, base)
	return ok && snapshotSuffix.MatchString(stamp)
}

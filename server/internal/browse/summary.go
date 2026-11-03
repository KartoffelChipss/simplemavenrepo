package browse

import (
	"cmp"
	"encoding/xml"
	"io/fs"
	"path"
	"simplemavenrepo/internal/maven"
	"simplemavenrepo/internal/repository"
	"slices"
	"strings"
	"time"
)

type Summary struct {
	Artifact repository.Artifact `json:"artifact"`
	LatestVersion string `json:"latestVersion,omitempty"`
	VersionCount  int    `json:"versionCount"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

func Artifacts(fsys fs.FS, artifacts []repository.Artifact, canRead func(repository.Artifact) bool) []Summary {
	summaries := []Summary{}
	for _, a := range artifacts {
		if canRead(a) {
			summaries = append(summaries, summarize(fsys, a))
		}
	}
	return summaries
}

func summarize(fsys fs.FS, a repository.Artifact) Summary {
	s := Summary{Artifact: a}

	var versions []string
	entries, _ := fs.ReadDir(fsys, a.Path)
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			versions = append(versions, e.Name())
		}
	}
	s.VersionCount = len(versions)
	if len(versions) > 0 {
		s.LatestVersion = slices.MaxFunc(versions, maven.CompareVersions)
	}

	metadataPath := path.Join(a.Path, maven.MetadataFileName)
	if info, err := fs.Stat(fsys, metadataPath); err == nil {
		modified := info.ModTime().UTC()
		s.UpdatedAt = &modified
	}
	if data, err := fs.ReadFile(fsys, metadataPath); err == nil {
		var m maven.Metadata
		if xml.Unmarshal(data, &m) == nil {
			s.LatestVersion = cmp.Or(m.Versioning.Release, m.Versioning.Latest, s.LatestVersion)
		}
	}
	return s
}

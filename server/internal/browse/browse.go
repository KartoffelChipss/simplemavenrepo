package browse

import (
	"errors"
	"io/fs"
	"path"
	"simplemavenrepo/internal/maven"
	"simplemavenrepo/internal/repository"
	"slices"
	"strings"
	"time"
)

// ErrNotFound is returned for paths that don't exist or that the caller may not see.
var ErrNotFound = errors.New("not found")

type EntryType string

const (
	TypeDirectory EntryType = "directory"
	TypeFile      EntryType = "file"
)

type Entry struct {
	Name       string               `json:"name"`
	Type       EntryType            `json:"type"`
	Size       *int64               `json:"size,omitempty"`
	ModifiedAt *time.Time           `json:"modifiedAt,omitempty"`
	Artifact   *repository.Artifact `json:"artifact,omitempty"`
}

type Listing struct {
	// Path is the normalized path, without leading or trailing slashes. The repository root is "".
	Path          string               `json:"path"`
	Artifact      *repository.Artifact `json:"artifact,omitempty"`
	LatestVersion string               `json:"latestVersion,omitempty"`
	Entries       []Entry              `json:"entries"`
}

func List(fsys fs.FS, artifacts []repository.Artifact, canRead func(repository.Artifact) bool, p string) (Listing, error) {
	p, ok := normalize(p)
	if !ok {
		return Listing{}, ErrNotFound
	}

	var owner *repository.Artifact
	for _, a := range artifacts {
		if (p == a.Path || strings.HasPrefix(p, a.Path+"/")) && (owner == nil || len(a.Path) > len(owner.Path)) {
			owner = &a
		}
	}
	if owner == nil || !canRead(*owner) {
		return Listing{}, ErrNotFound
	}
	return listArtifactDir(fsys, artifacts, canRead, *owner, p)
}

// normalize cleans p and rejects hidden segments, which hold files that are still being uploaded.
func normalize(p string) (string, bool) {
	p = strings.Trim(path.Clean("/"+p), "/")
	if p == "" {
		return "", true
	}
	for _, s := range strings.Split(p, "/") {
		if strings.HasPrefix(s, ".") {
			return "", false
		}
	}
	return p, true
}

func listArtifactDir(fsys fs.FS, artifacts []repository.Artifact, canRead func(repository.Artifact) bool, owner repository.Artifact, p string) (Listing, error) {
	listing := Listing{Path: p, Artifact: &owner, Entries: []Entry{}}
	if p == owner.Path {
		listing.LatestVersion = summarize(fsys, owner).LatestVersion
	}

	dirEntries, err := fs.ReadDir(fsys, p)
	if errors.Is(err, fs.ErrNotExist) && p == owner.Path {
		return listing, nil
	}
	if err != nil {
		// Also covers p being a file, files are downloaded, not listed
		return Listing{}, ErrNotFound
	}

	nested := map[string]repository.Artifact{}
	for _, a := range artifacts {
		nested[a.Path] = a
	}

	for _, d := range dirEntries {
		if strings.HasPrefix(d.Name(), ".") {
			continue
		}
		info, err := d.Info()
		if err != nil {
			continue
		}
		modified := info.ModTime().UTC()
		entry := Entry{Name: d.Name(), Type: TypeFile, ModifiedAt: &modified}
		if d.IsDir() {
			entry.Type = TypeDirectory
			if a, ok := nested[path.Join(p, d.Name())]; ok {
				if !canRead(a) {
					continue
				}
				entry.Artifact = &a
			}
		} else {
			size := info.Size()
			entry.Size = &size
		}
		listing.Entries = append(listing.Entries, entry)
	}

	// Directories first. The directories of an artifact are its versions, newest first.
	versions := p == owner.Path
	slices.SortFunc(listing.Entries, func(x, y Entry) int {
		if x.Type != y.Type {
			if x.Type == TypeDirectory {
				return -1
			}
			return 1
		}
		if versions && x.Type == TypeDirectory {
			return maven.CompareVersions(y.Name, x.Name)
		}
		return strings.Compare(x.Name, y.Name)
	})
	return listing, nil
}

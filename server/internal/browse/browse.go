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
	Path     string               `json:"path"`
	Artifact *repository.Artifact `json:"artifact,omitempty"`
	Entries  []Entry              `json:"entries"`
}

func List(fsys fs.FS, artifacts []repository.Artifact, canRead func(repository.Artifact) bool, p string) (Listing, error) {
	p, ok := normalize(p)
	if !ok {
		return Listing{}, ErrNotFound
	}

	var readable []repository.Artifact
	for _, a := range artifacts {
		if canRead(a) {
			readable = append(readable, a)
		}
	}

	var owner *repository.Artifact
	for _, a := range artifacts {
		if (p == a.Path || strings.HasPrefix(p, a.Path+"/")) && (owner == nil || len(a.Path) > len(owner.Path)) {
			owner = &a
		}
	}
	if owner != nil {
		if !canRead(*owner) {
			return Listing{}, ErrNotFound
		}
		return listArtifactDir(fsys, artifacts, canRead, *owner, p)
	}
	return listGroups(readable, p)
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

func listGroups(readable []repository.Artifact, p string) (Listing, error) {
	prefix := ""
	if p != "" {
		prefix = p + "/"
	}

	listing := Listing{Path: p, Entries: []Entry{}}
	seen := map[string]bool{}
	for _, a := range readable {
		rest, found := strings.CutPrefix(a.Path, prefix)
		if !found {
			continue
		}
		name, _, _ := strings.Cut(rest, "/")
		if seen[name] {
			continue
		}
		seen[name] = true
		entry := Entry{Name: name, Type: TypeDirectory}
		if rest == name {
			entry.Artifact = &a
		}
		listing.Entries = append(listing.Entries, entry)
	}
	if len(listing.Entries) == 0 {
		return Listing{}, ErrNotFound
	}
	slices.SortFunc(listing.Entries, func(x, y Entry) int { return strings.Compare(x.Name, y.Name) })
	return listing, nil
}

func listArtifactDir(fsys fs.FS, artifacts []repository.Artifact, canRead func(repository.Artifact) bool, owner repository.Artifact, p string) (Listing, error) {
	listing := Listing{Path: p, Artifact: &owner, Entries: []Entry{}}

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
		entry := Entry{Name: d.Name(), Type: TypeFile}
		if d.IsDir() {
			entry.Type = TypeDirectory
			if a, ok := nested[path.Join(p, d.Name())]; ok {
				if !canRead(a) {
					continue
				}
				entry.Artifact = &a
			}
		} else {
			info, err := d.Info()
			if err != nil {
				continue
			}
			size, modified := info.Size(), info.ModTime().UTC()
			entry.Size, entry.ModifiedAt = &size, &modified
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

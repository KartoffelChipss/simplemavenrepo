package storage

import (
	"errors"
	"io/fs"
	"path"
)

type Usage struct {
	Total        int64
	Files        int64
	ByArtifact   map[string]int64
	Unregistered int64
}

// Measure walks dir and sums up the sizes of its regular files, attributing each file to the deepest of artifactPaths
// containing it. A missing dir is not an error.
func Measure(fsys fs.FS, dir string, artifactPaths []string) (Usage, error) {
	artifacts := make(map[string]bool, len(artifactPaths))
	for _, p := range artifactPaths {
		artifacts[p] = true
	}
	u := Usage{ByArtifact: map[string]int64{}}

	// owner caches the artifact a directory belongs to, "" for none
	owner := map[string]string{}
	var ownerOf func(dir string) string
	ownerOf = func(dir string) string {
		if o, ok := owner[dir]; ok {
			return o
		}
		o := ""
		if artifacts[dir] {
			o = dir
		} else if dir != "." && dir != "/" && dir != "" {
			o = ownerOf(path.Dir(dir))
		}
		owner[dir] = o
		return o
	}

	err := fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == dir && errors.Is(err, fs.ErrNotExist) {
				return fs.SkipAll
			}
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		info, err := d.Info()
		if errors.Is(err, fs.ErrNotExist) {
			// Removed while walking, e.g. a finished upload's temporary file
			return nil
		}
		if err != nil {
			return err
		}
		u.Total += info.Size()
		u.Files++
		if o := ownerOf(path.Dir(p)); o != "" {
			u.ByArtifact[o] += info.Size()
		} else {
			u.Unregistered += info.Size()
		}
		return nil
	})
	return u, err
}

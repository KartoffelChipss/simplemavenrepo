package javadoc

import (
	"archive/zip"
	"errors"
	"io"
	"io/fs"
	"path"
	"strings"
)

// FindJar returns the path of the javadoc jar in the version directory dir, or fs.ErrNotExist.
func FindJar(fsys fs.FS, dir, artifactID, version string) (string, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return "", err
	}

	exact := artifactID + "-" + version + "-javadoc.jar"
	snapshotPrefix := artifactID + "-" + strings.TrimSuffix(version, "-SNAPSHOT") + "-"
	isSnapshot := strings.HasSuffix(version, "-SNAPSHOT")

	var best fs.FileInfo
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, "-javadoc.jar") {
			continue
		}
		if name != exact && !(isSnapshot && strings.HasPrefix(name, snapshotPrefix)) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if best == nil || info.ModTime().After(best.ModTime()) {
			best = info
		}
	}
	if best == nil {
		return "", fs.ErrNotExist
	}
	return path.Join(dir, best.Name()), nil
}

// Jar is an opened javadoc jar. Close it when done.
type Jar struct {
	*zip.Reader
	f fs.File
}

func (j *Jar) Close() error {
	return j.f.Close()
}

// OpenJar opens the jar at p. The file system must return files that implement io.ReaderAt, like os.Root does.
func OpenJar(fsys fs.FS, p string) (*Jar, error) {
	f, err := fsys.Open(p)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	ra, ok := f.(io.ReaderAt)
	if !ok {
		f.Close()
		return nil, errors.New("file system does not support random access")
	}
	zr, err := zip.NewReader(ra, info.Size())
	if err != nil {
		f.Close()
		return nil, err
	}
	return &Jar{Reader: zr, f: f}, nil
}

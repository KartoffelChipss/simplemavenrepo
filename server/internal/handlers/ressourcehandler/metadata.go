package ressourcehandler

import (
	"bytes"
	"os"
	"path"
	"simplemavenrepo/internal/maven"
	"strings"
	"sync"
	"time"
)

// One lock per artifact directory
var metadataLocks sync.Map

// updateMetadata regenerates the maven-metadata.xml files affected by an upload to <group path>/<artifactId>/<version>/<file>.
func updateMetadata(root *os.Root, relPath string) error {
	versionDir := path.Dir(relPath)
	artifactDir := path.Dir(versionDir)
	version := path.Base(versionDir)
	artifactID := path.Base(artifactDir)
	groupID := strings.ReplaceAll(path.Dir(artifactDir), "/", ".")

	lock, _ := metadataLocks.LoadOrStore(artifactDir, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()

	now := time.Now()

	if strings.HasSuffix(version, "-SNAPSHOT") {
		files, err := listEntries(root, versionDir, false)
		if err != nil {
			return err
		}
		if m, ok := maven.NewSnapshotMetadata(groupID, artifactID, version, files, storedExtensions, now); ok {
			if err := writeMetadata(root, versionDir, m); err != nil {
				return err
			}
		}
	}

	versions, err := listVersions(root, artifactDir)
	if err != nil {
		return err
	}
	return writeMetadata(root, artifactDir, maven.NewArtifactMetadata(groupID, artifactID, versions, now))
}

func writeMetadata(root *os.Root, dir string, m maven.Metadata) error {
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	return writeFileAtomicWithHashes(root, path.Join(dir, maven.MetadataFileName), bytes.NewReader(data))
}

// listVersions returns the subdirectories of artifactDir that contain at least one stored artifact file.
func listVersions(root *os.Root, artifactDir string) ([]string, error) {
	dirs, err := listEntries(root, artifactDir, true)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, dir := range dirs {
		files, err := listEntries(root, path.Join(artifactDir, dir), false)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if hasStoredExtension(f) {
				versions = append(versions, dir)
				break
			}
		}
	}
	return versions, nil
}

// listEntries returns the names of the non-hidden directories (dirs = true) or files (dirs = false) in dir.
func listEntries(root *os.Root, dir string, dirs bool) ([]string, error) {
	f, err := root.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries, err := f.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") || e.IsDir() != dirs {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}

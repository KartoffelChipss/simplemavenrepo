package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
)

const (
	layoutMarker  = ".smr-layout"
	layoutVersion = "2"
	// Holds the files while they are moved, so a top-level directory named like the target can be moved too
	layoutStaging = ".smr-layout-staging"
)

// MigrateLayout moves a repository root from the single repository layout, where the group directories are at the top,
// into the directory of the repository named target. It is a no-op once the root has the current layout, and picks up
// where it left off if it was interrupted.
func MigrateLayout(root *os.Root, target string) error {
	data, err := fs.ReadFile(root.FS(), layoutMarker)
	if errors.Is(err, fs.ErrNotExist) {
		if err := stageTopLevel(root); err != nil {
			return err
		}
		// Written before the staged files are moved into place, so a restart doesn't stage them a second time
		if err := root.WriteFile(layoutMarker, []byte(layoutVersion+"\n"), 0o644); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if v := strings.TrimSpace(string(data)); v != layoutVersion {
		return fmt.Errorf("repository layout %q is not supported by this server", v)
	}

	if _, err := root.Stat(layoutStaging); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if err := root.Rename(layoutStaging, target); err != nil {
		return fmt.Errorf("moving the repository files into %s: %w", target, err)
	}
	log.Printf("Moved the repository files into the %q repository", target)
	return nil
}

// stageTopLevel moves every top-level entry except hidden ones into the staging directory.
func stageTopLevel(root *os.Root) error {
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || name == "lost+found" {
			continue
		}
		if err := root.MkdirAll(layoutStaging, 0o755); err != nil {
			return err
		}
		if err := root.Rename(name, layoutStaging+"/"+name); err != nil {
			return err
		}
	}
	return nil
}

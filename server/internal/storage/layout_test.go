package storage

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"testing"
)

func writeFiles(t *testing.T, root *os.Root, files ...string) {
	t.Helper()
	for _, f := range files {
		if err := root.MkdirAll(path.Dir(f), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := root.WriteFile(f, []byte(f), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func assertFiles(t *testing.T, root *os.Root, want map[string]bool) {
	t.Helper()
	for f, exists := range want {
		_, err := root.Stat(f)
		if exists && err != nil {
			t.Errorf("%s should exist: %v", f, err)
		} else if !exists && !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s should not exist: %v", f, err)
		}
	}
}

func TestMigrateLayout(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	// A group directory with the name of the target must end up inside it, not be mistaken for it
	writeFiles(t, root, "com/example/demo/1.0/demo-1.0.jar", "releases/odd/1.0/odd-1.0.jar", ".hidden")

	if err := MigrateLayout(root, "releases"); err != nil {
		t.Fatal(err)
	}
	assertFiles(t, root, map[string]bool{
		"releases/com/example/demo/1.0/demo-1.0.jar": true,
		"releases/releases/odd/1.0/odd-1.0.jar":      true,
		".hidden":                                    true,
		"com":                                        false,
		layoutStaging:                                false,
		layoutMarker:                                 true,
	})

	// Running again, e.g. on the next start, must not move anything
	if err := MigrateLayout(root, "releases"); err != nil {
		t.Fatal(err)
	}
	assertFiles(t, root, map[string]bool{
		"releases/com/example/demo/1.0/demo-1.0.jar": true,
		"releases/releases":                          true,
		"releases/releases/releases":                 false,
	})
}

func TestMigrateLayoutResumes(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	// Interrupted after the marker was written, before the staged files were moved into place
	writeFiles(t, root, layoutStaging+"/com/example/demo/1.0/demo-1.0.jar")
	if err := root.WriteFile(layoutMarker, []byte(layoutVersion+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := MigrateLayout(root, "releases"); err != nil {
		t.Fatal(err)
	}
	assertFiles(t, root, map[string]bool{
		"releases/com/example/demo/1.0/demo-1.0.jar": true,
		layoutStaging: false,
	})
}

func TestMigrateLayoutEmpty(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := MigrateLayout(root, "releases"); err != nil {
		t.Fatal(err)
	}
	assertFiles(t, root, map[string]bool{layoutMarker: true, "releases": false, layoutStaging: false})
}

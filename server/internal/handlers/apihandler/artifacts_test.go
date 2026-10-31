package apihandler

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"testing"
)

func TestRemoveTreeExcept(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	files := []string{
		"com/example/demo/maven-metadata.xml",
		"com/example/demo/1.0/demo-1.0.jar",
		"com/example/demo/sub/1.0/sub-1.0.jar", // com.example.demo:sub, nested inside com.example:demo
		"com/example/other/1.0/other-1.0.jar",
	}
	for _, f := range files {
		if err := root.MkdirAll(path.Dir(f), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := root.WriteFile(f, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := removeTreeExcept(root, "com/example/demo", []string{"com/example/demo/sub"}); err != nil {
		t.Fatal(err)
	}
	removeEmptyParents(root, "com/example/demo")

	for f, want := range map[string]bool{
		"com/example/demo/maven-metadata.xml":  false,
		"com/example/demo/1.0":                 false,
		"com/example/demo/sub/1.0/sub-1.0.jar": true,
		"com/example/other/1.0/other-1.0.jar":  true,
	} {
		_, err := root.Stat(f)
		if exists := err == nil; exists != want {
			t.Errorf("%s exists = %v, want %v (%v)", f, exists, want, err)
		}
	}

	// Without nested artifacts the whole directory goes, and so do parents that end up empty
	if err := removeTreeExcept(root, "com/example/other", nil); err != nil {
		t.Fatal(err)
	}
	removeEmptyParents(root, "com/example/other")
	if _, err := root.Stat("com/example/other"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("com/example/other still exists: %v", err)
	}
	if _, err := root.Stat("com/example"); err != nil {
		t.Errorf("com/example must stay, it still holds com/example/demo/sub: %v", err)
	}

	if err := removeTreeExcept(root, "com/missing/artifact", []string{"com/missing/artifact/x"}); err != nil {
		t.Errorf("missing directory: %v", err)
	}
}

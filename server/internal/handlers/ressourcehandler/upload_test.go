package ressourcehandler

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
)

func TestIsSnapshotPath(t *testing.T) {
	tests := map[string]bool{
		"com/example/demo/1.0-SNAPSHOT/demo-1.0-20260101.120000-1.jar": true,
		"com/example/demo/1.0/demo-1.0.jar":                            false,
		"com/example/demo/1.0/demo-1.0-SNAPSHOT.jar":                   false,
	}
	for p, want := range tests {
		if got := isSnapshotPath(p); got != want {
			t.Errorf("isSnapshotPath(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestWriteFileAtomicReplace(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	const p = "com/example/demo/1.0/demo-1.0.jar"
	if err := writeFileAtomicWithHashes(root, p, strings.NewReader("first"), false); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomicWithHashes(root, p, strings.NewReader("second"), false); !errors.Is(err, fs.ErrExist) {
		t.Fatalf("overwrite without replace: got %v, want fs.ErrExist", err)
	}
	assertContent(t, root, p, "first")

	if err := writeFileAtomicWithHashes(root, p, strings.NewReader("third"), true); err != nil {
		t.Fatal(err)
	}
	assertContent(t, root, p, "third")

	entries, err := fs.ReadDir(root.FS(), "com/example/demo/1.0")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			t.Errorf("temporary file %s was left behind", e.Name())
		}
	}
}

func assertContent(t *testing.T, root *os.Root, p, want string) {
	t.Helper()
	got, err := fs.ReadFile(root.FS(), p)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", p, got, want)
	}
}

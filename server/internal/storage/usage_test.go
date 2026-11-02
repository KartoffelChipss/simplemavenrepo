package storage

import (
	"testing"
	"testing/fstest"
)

func TestMeasure(t *testing.T) {
	fsys := fstest.MapFS{
		"com/example/demo/maven-metadata.xml":  {Data: make([]byte, 10)},
		"com/example/demo/1.0/demo-1.0.jar":    {Data: make([]byte, 100)},
		"com/example/demo/sub/1.0/sub-1.0.jar": {Data: make([]byte, 1000)}, // com.example.demo:sub, nested inside demo
		"com/example/other/1.0/other-1.0.jar":  {Data: make([]byte, 5)},
		"org/orphan/1.0/orphan-1.0.jar":        {Data: make([]byte, 7)},
	}
	artifacts := []string{"com/example/demo", "com/example/demo/sub", "com/example/other", "com/example/empty"}

	u, err := Measure(fsys, ".", artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if u.Total != 1122 || u.Files != 5 || u.Unregistered != 7 {
		t.Errorf("total %d, files %d, unregistered %d", u.Total, u.Files, u.Unregistered)
	}
	for p, want := range map[string]int64{
		"com/example/demo":     110,
		"com/example/demo/sub": 1000,
		"com/example/other":    5,
		"com/example/empty":    0,
	} {
		if got := u.ByArtifact[p]; got != want {
			t.Errorf("%s = %d, want %d", p, got, want)
		}
	}

	// Measuring a single artifact leaves out the nested one
	u, err = Measure(fsys, "com/example/demo", artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if got := u.ByArtifact["com/example/demo"]; got != 110 {
		t.Errorf("demo alone = %d, want 110", got)
	}

	u, err = Measure(fsys, "com/missing", artifacts)
	if err != nil || u.Total != 0 {
		t.Errorf("missing directory: %d, %v", u.Total, err)
	}
}

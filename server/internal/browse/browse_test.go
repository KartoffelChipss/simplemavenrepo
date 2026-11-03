package browse

import (
	"errors"
	"simplemavenrepo/internal/repository"
	"testing"
	"testing/fstest"
)

func names(l Listing) []string {
	var n []string
	for _, e := range l.Entries {
		n = append(n, e.Name)
	}
	return n
}

func TestList(t *testing.T) {
	file := &fstest.MapFile{Data: []byte("x")}
	fsys := fstest.MapFS{
		"org/example/lib/1.9/lib-1.9.jar":          file,
		"org/example/lib/1.10/lib-1.10.jar":        file,
		"org/example/lib/1.10-SNAPSHOT/a.jar":      file,
		"org/example/lib/maven-metadata.xml":       file,
		"org/example/lib/1.10/.lib.jar.123.upload": file,
		"org/example/secret/1.0/secret-1.0.jar":    file,
		"org/example/unregistered/1.0/x.jar":       file,
	}
	public := repository.Artifact{GroupID: "org.example", ArtifactID: "lib", Path: "org/example/lib", Public: true}
	private := repository.Artifact{GroupID: "org.example", ArtifactID: "secret", Path: "org/example/secret"}
	empty := repository.Artifact{GroupID: "com.empty", ArtifactID: "new", Path: "com/empty/new", Public: true}
	artifacts := []repository.Artifact{public, private, empty}

	anonymous := func(a repository.Artifact) bool { return a.Public }
	everyone := func(repository.Artifact) bool { return true }

	tests := []struct {
		name    string
		canRead func(repository.Artifact) bool
		path    string
		want    []string
		missing bool
	}{
		{name: "root", canRead: anonymous, path: "", missing: true},
		{name: "group", canRead: everyone, path: "org/example", missing: true},
		{name: "path is normalized", canRead: anonymous, path: "/org/example/lib/1.10/", want: []string{"lib-1.10.jar"}},
		{name: "versions newest first, then files", canRead: anonymous, path: "org/example/lib", want: []string{"1.10", "1.10-SNAPSHOT", "1.9", "maven-metadata.xml"}},
		{name: "upload temp files are hidden", canRead: anonymous, path: "org/example/lib/1.10", want: []string{"lib-1.10.jar"}},
		{name: "registered without files", canRead: anonymous, path: "com/empty/new", want: nil},
		{name: "private artifact", canRead: anonymous, path: "org/example/secret", missing: true},
		{name: "inside private artifact", canRead: anonymous, path: "org/example/secret/1.0", missing: true},
		{name: "private artifact with access", canRead: everyone, path: "org/example/secret/1.0", want: []string{"secret-1.0.jar"}},
		{name: "unregistered", canRead: everyone, path: "org/example/unregistered", missing: true},
		{name: "file", canRead: anonymous, path: "org/example/lib/maven-metadata.xml", missing: true},
		{name: "missing version", canRead: anonymous, path: "org/example/lib/2.0", missing: true},
		{name: "traversal", canRead: anonymous, path: "org/example/lib/../secret", missing: true},
		{name: "hidden segment", canRead: anonymous, path: "org/example/lib/1.10/.lib.jar.123.upload", missing: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := List(fsys, artifacts, tt.canRead, tt.path)
			if tt.missing {
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("got %v, %v, want ErrNotFound", names(l), err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := names(l)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestListMarksArtifacts(t *testing.T) {
	lib := repository.Artifact{GroupID: "org.example", ArtifactID: "lib", Path: "org/example/lib", Public: true}
	fsys := fstest.MapFS{"org/example/lib/1.0/lib-1.0.jar": &fstest.MapFile{Data: []byte("jar")}}
	canRead := func(repository.Artifact) bool { return true }

	nested := repository.Artifact{GroupID: "org.example.lib", ArtifactID: "ext", Path: "org/example/lib/ext", Public: true}
	fsys["org/example/lib/ext/1.0/ext-1.0.jar"] = &fstest.MapFile{Data: []byte("jar")}
	outer, err := List(fsys, []repository.Artifact{lib, nested}, canRead, "org/example/lib")
	if err != nil || outer.Entries[1].Name != "ext" || outer.Entries[1].Artifact == nil {
		t.Fatalf("listing should mark nested artifacts: %+v, %v", outer, err)
	}
	version, err := List(fsys, []repository.Artifact{lib}, canRead, "org/example/lib/1.0")
	if err != nil || version.Artifact == nil || version.Artifact.ArtifactID != "lib" {
		t.Fatalf("listing inside an artifact should name it: %+v, %v", version, err)
	}
	if f := version.Entries[0]; f.Type != TypeFile || f.Size == nil || *f.Size != 3 || f.ModifiedAt == nil {
		t.Fatalf("file entry: %+v", f)
	}
}

func TestArtifacts(t *testing.T) {
	file := &fstest.MapFile{Data: []byte("x")}
	metadata := &fstest.MapFile{Data: []byte(`<metadata><versioning><release>1.9</release></versioning></metadata>`)}
	fsys := fstest.MapFS{
		"org/example/lib/1.9/lib-1.9.jar":     file,
		"org/example/lib/1.10-SNAPSHOT/a.jar": file,
		"org/example/lib/.tmp/x":              file,
		"org/example/lib/maven-metadata.xml":  metadata,
		"org/example/raw/2.0/raw-2.0.jar":     file,
		"org/example/raw/10.0/raw-10.0.jar":   file,
	}
	lib := repository.Artifact{GroupID: "org.example", ArtifactID: "lib", Path: "org/example/lib", Public: true}
	raw := repository.Artifact{GroupID: "org.example", ArtifactID: "raw", Path: "org/example/raw", Public: true}
	secret := repository.Artifact{GroupID: "org.example", ArtifactID: "secret", Path: "org/example/secret"}
	empty := repository.Artifact{GroupID: "org.example", ArtifactID: "empty", Path: "org/example/empty", Public: true}

	got := Artifacts(fsys, []repository.Artifact{lib, raw, secret, empty}, func(a repository.Artifact) bool { return a.Public })
	if len(got) != 3 {
		t.Fatalf("private artifact should be left out: %+v", got)
	}
	if s := got[0]; s.LatestVersion != "1.9" || s.VersionCount != 2 || s.UpdatedAt == nil {
		t.Errorf("metadata release should win: %+v", s)
	}
	if s := got[1]; s.LatestVersion != "10.0" || s.VersionCount != 2 || s.UpdatedAt != nil {
		t.Errorf("without metadata the newest directory should be used: %+v", s)
	}
	if s := got[2]; s.LatestVersion != "" || s.VersionCount != 0 {
		t.Errorf("nothing published: %+v", s)
	}
}

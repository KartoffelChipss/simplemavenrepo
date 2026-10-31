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
		{name: "root", canRead: anonymous, path: "", want: []string{"com", "org"}},
		{name: "group hides private and unregistered", canRead: anonymous, path: "org/example", want: []string{"lib"}},
		{name: "group with read access", canRead: everyone, path: "/org/example/", want: []string{"lib", "secret"}},
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
		{name: "nothing readable", canRead: func(repository.Artifact) bool { return false }, path: "", missing: true},
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

	group, err := List(fsys, []repository.Artifact{lib}, canRead, "org/example")
	if err != nil || group.Entries[0].Artifact == nil || group.Artifact != nil {
		t.Fatalf("group listing should mark the artifact entry: %+v, %v", group, err)
	}
	version, err := List(fsys, []repository.Artifact{lib}, canRead, "org/example/lib/1.0")
	if err != nil || version.Artifact == nil || version.Artifact.ArtifactID != "lib" {
		t.Fatalf("listing inside an artifact should name it: %+v, %v", version, err)
	}
	if f := version.Entries[0]; f.Type != TypeFile || f.Size == nil || *f.Size != 3 || f.ModifiedAt == nil {
		t.Fatalf("file entry: %+v", f)
	}
}

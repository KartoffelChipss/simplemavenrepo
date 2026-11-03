package pom

import (
	"errors"
	"slices"
	"testing"
	"testing/fstest"
)

func file(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }

var repo = fstest.MapFS{
	"org/example/parent/1/parent-1.pom": file(`<project>
		<groupId>org.example</groupId><artifactId>parent</artifactId><version>1</version><packaging>pom</packaging>
		<url>https://example.org</url>
		<properties><guava.version>33.0</guava.version></properties>
		<licenses><license><name>MIT</name><url>https://mit.example</url></license></licenses>
		<developers><developer><id>jd</id><name>Jane Doe</name><organization>Example</organization></developer></developers>
		<scm><url>https://git.example/${project.artifactId}</url></scm>
		<dependencyManagement><dependencies>
			<dependency><groupId>com.google.guava</groupId><artifactId>guava</artifactId><version>${guava.version}</version></dependency>
			<dependency><groupId>org.example</groupId><artifactId>bom</artifactId><version>2</version><type>pom</type><scope>import</scope></dependency>
		</dependencies></dependencyManagement>
	</project>`),
	"org/example/bom/2/bom-2.pom": file(`<project>
		<groupId>org.example</groupId><artifactId>bom</artifactId><version>2</version>
		<dependencyManagement><dependencies>
			<dependency><groupId>org.junit</groupId><artifactId>junit</artifactId><version>5.10</version><scope>test</scope></dependency>
		</dependencies></dependencyManagement>
	</project>`),
	"org/example/lib/1.0/lib-1.0.pom": file(`<project>
		<parent><groupId>org.example</groupId><artifactId>parent</artifactId><version>1</version></parent>
		<artifactId>lib</artifactId><version>1.0</version>
		<name>Lib ${project.version}</name>
		<dependencies>
			<dependency><groupId>com.google.guava</groupId><artifactId>guava</artifactId></dependency>
			<dependency><groupId>org.junit</groupId><artifactId>junit</artifactId></dependency>
			<dependency><groupId>org.example</groupId><artifactId>other</artifactId><version>${project.version}</version><optional>true</optional></dependency>
			<dependency><groupId>x</groupId><artifactId>y</artifactId><version>${missing}</version></dependency>
		</dependencies>
	</project>`),
	"org/example/orphan/1.0/orphan-1.0.pom": file(`<project>
		<parent><groupId>org.apache</groupId><artifactId>apache</artifactId><version>30</version></parent>
		<artifactId>orphan</artifactId>
	</project>`),
	"org/example/snap/1.0-SNAPSHOT/snap-1.0-20240101.000000-1.pom": file(`<project><groupId>g</groupId><artifactId>snap</artifactId><version>old</version></project>`),
	"org/example/snap/1.0-SNAPSHOT/snap-1.0-20240102.000000-2.pom": file(`<project><groupId>g</groupId><artifactId>snap</artifactId><version>new</version></project>`),
}

func everything(string) bool { return true }

func TestResolveMergesParentsAndBoms(t *testing.T) {
	p, err := Resolve(repo, "org/example/lib/1.0", "lib", "1.0", everything)
	if err != nil {
		t.Fatal(err)
	}
	if p.GroupID != "org.example" || p.Packaging != "jar" || p.Name != "Lib 1.0" || p.URL != "https://example.org" {
		t.Errorf("metadata: %+v", p)
	}
	if len(p.Licenses) != 1 || p.Licenses[0].URL != "https://mit.example" {
		t.Errorf("licenses: %+v", p.Licenses)
	}
	if len(p.Developers) != 1 || p.Developers[0].Organization != "Example" {
		t.Errorf("developers: %+v", p.Developers)
	}
	if p.SCM == nil || p.SCM.URL != "https://git.example/lib" {
		t.Errorf("scm: %+v", p.SCM)
	}
	if p.Parent == nil || p.Parent.ArtifactID != "parent" || len(p.Unresolved) != 0 {
		t.Errorf("parent: %+v, unresolved %v", p.Parent, p.Unresolved)
	}

	want := []Dependency{
		{GroupID: "com.google.guava", ArtifactID: "guava", Version: "33.0", Scope: "compile"},
		{GroupID: "org.junit", ArtifactID: "junit", Version: "5.10", Scope: "test"},
		{GroupID: "org.example", ArtifactID: "other", Version: "1.0", Scope: "compile", Optional: true},
		{GroupID: "x", ArtifactID: "y", Version: "${missing}", Scope: "compile"},
	}
	if !slices.Equal(p.Dependencies, want) {
		t.Errorf("dependencies:\n got %+v\nwant %+v", p.Dependencies, want)
	}
}

func TestResolveMissingParent(t *testing.T) {
	p, err := Resolve(repo, "org/example/orphan/1.0", "orphan", "1.0", everything)
	if err != nil {
		t.Fatal(err)
	}
	if p.GroupID != "org.apache" || p.Version != "30" || !slices.Equal(p.Unresolved, []string{"org.apache:apache:30"}) {
		t.Errorf("got %+v", p)
	}
}

func TestResolveSkipsUnreadableParent(t *testing.T) {
	p, err := Resolve(repo, "org/example/lib/1.0", "lib", "1.0", func(path string) bool { return path != "org/example/parent" })
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Licenses) != 0 || !slices.Equal(p.Unresolved, []string{"org.example:parent:1"}) {
		t.Errorf("private parent should not be merged: %+v", p)
	}
}

func TestResolveSnapshotAndMissing(t *testing.T) {
	p, err := Resolve(repo, "org/example/snap/1.0-SNAPSHOT", "snap", "1.0-SNAPSHOT", everything)
	if err != nil || p.Version != "new" {
		t.Errorf("newest snapshot POM should be used: %+v, %v", p, err)
	}
	if _, err := Resolve(repo, "org/example/lib/2.0", "lib", "2.0", everything); !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

package javadoc

import (
	"archive/zip"
	"bytes"
	"errors"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
	"time"
)

func makeZip(t *testing.T, files map[string]string) *zip.Reader {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte(content))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	return zr
}

func zipBytes(t *testing.T, files map[string]string) string {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, _ := w.Create(name)
		f.Write([]byte(content))
	}
	w.Close()
	return buf.String()
}

func TestReadIndexModern(t *testing.T) {
	zr := makeZip(t, map[string]string{
		"index.html":                       "",
		"overview-tree.html":               "",
		"index-all.html":                   "",
		"com/example/package-summary.html": "",
		"com/example/Foo.html":             "",
		"com/example/Foo.Bar.html":         "",
		"Root.html":                        "",
		"package-search-index.js":          `packageSearchIndex = [{"l":"All Packages","u":"allpackages-index.html"},{"l":"com.example"},{"l":"com.missing"}];updateSearchResults();`,
		"type-search-index.js":             `typeSearchIndex = [{"l":"All Classes","u":"allclasses-index.html"},{"p":"com.example","l":"Foo.Bar"},{"p":"com.example","l":"Foo"},{"p":"<Unnamed>","l":"Root"}];updateSearchResults();`,
		// JDK 11 names the anchor field "url", later versions "u"
		"member-search-index.js": `memberSearchIndex = [{"p":"com.example","c":"Foo","l":"get(int)","url":"get-int-"},{"p":"com.example","c":"Foo.Bar","l":"run(String)","u":"run(java.lang.String)"},{"p":"com.example","c":"Foo","l":"SIZE"},{"p":"com.example","c":"Gone","l":"x()"}];updateSearchResults();`,
	})

	idx, err := ReadIndex(zr)
	if err != nil {
		t.Fatal(err)
	}
	if idx.StartPage != "index.html" {
		t.Errorf("StartPage = %q", idx.StartPage)
	}
	wantPages := []Page{{"Overview", "index.html"}, {"Tree", "overview-tree.html"}, {"Index", "index-all.html"}}
	if !reflect.DeepEqual(idx.Pages, wantPages) {
		t.Errorf("Pages = %v", idx.Pages)
	}
	if want := []Package{{"com.example", "com/example/package-summary.html"}}; !reflect.DeepEqual(idx.Packages, want) {
		t.Errorf("Packages = %v", idx.Packages)
	}
	wantTypes := []Type{
		{"<Unnamed>", "Root", "Root.html"},
		{"com.example", "Foo", "com/example/Foo.html"},
		{"com.example", "Foo.Bar", "com/example/Foo.Bar.html"},
	}
	if !reflect.DeepEqual(idx.Types, wantTypes) {
		t.Errorf("Types = %v", idx.Types)
	}
	wantMembers := []Member{
		{"com.example", "Foo", "get(int)", "com/example/Foo.html#get-int-"},
		{"com.example", "Foo.Bar", "run(String)", "com/example/Foo.Bar.html#run(java.lang.String)"},
		{"com.example", "Foo", "SIZE", "com/example/Foo.html#SIZE"},
	}
	if !reflect.DeepEqual(idx.Members, wantMembers) {
		t.Errorf("Members = %v", idx.Members)
	}
}

func TestReadIndexModules(t *testing.T) {
	zr := makeZip(t, map[string]string{
		"index.html": "",
		"my.mod/com/example/package-summary.html": "",
		"my.mod/com/example/Foo.html":             "",
		"package-search-index.js":                 `packageSearchIndex = [{"m":"my.mod","l":"com.example"}]`,
		"type-search-index.js":                    `typeSearchIndex = [{"m":"my.mod","p":"com.example","l":"Foo"}]`,
	})
	idx, err := ReadIndex(zr)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.Packages) != 1 || idx.Packages[0].URL != "my.mod/com/example/package-summary.html" {
		t.Errorf("Packages = %v", idx.Packages)
	}
	if len(idx.Types) != 1 || idx.Types[0].URL != "my.mod/com/example/Foo.html" {
		t.Errorf("Types = %v", idx.Types)
	}
}

func TestReadIndexZipped(t *testing.T) {
	zr := makeZip(t, map[string]string{
		"index.html":                       "",
		"com/example/package-summary.html": "",
		"com/example/Foo.html":             "",
		"package-search-index.zip":         zipBytes(t, map[string]string{"package-search-index.json": `[{"l":"com.example"}]`}),
		"type-search-index.zip":            zipBytes(t, map[string]string{"type-search-index.json": `[{"p":"com.example","l":"Foo"}]`}),
	})
	idx, err := ReadIndex(zr)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.Packages) != 1 || len(idx.Types) != 1 || len(idx.Members) != 0 {
		t.Errorf("Index = %+v", idx)
	}
}

func TestReadIndexLegacy(t *testing.T) {
	zr := makeZip(t, map[string]string{
		"index.html":                       "<frameset>",
		"overview-summary.html":            "",
		"index-files/index-1.html":         "",
		"package-list":                     "com.example\ncom.empty\n",
		"com/example/package-summary.html": "",
		"com/example/Foo.html":             "",
		"com/example/Foo.Bar.html":         "",
		"allclasses-noframe.html": `<ul>
<li><a href="com/example/Foo.html" title="class in com.example">Foo</a></li>
<li><a href="com/example/Foo.Bar.html" title="interface in com.example"><span class="interfaceName">Foo.Bar</span></a></li>
<li><a href="com/example/Missing.html" title="class in com.example">Missing</a></li>
</ul>`,
	})
	idx, err := ReadIndex(zr)
	if err != nil {
		t.Fatal(err)
	}
	if idx.StartPage != "overview-summary.html" {
		t.Errorf("StartPage = %q", idx.StartPage)
	}
	wantPages := []Page{{"Overview", "overview-summary.html"}, {"Index", "index-files/index-1.html"}}
	if !reflect.DeepEqual(idx.Pages, wantPages) {
		t.Errorf("Pages = %v", idx.Pages)
	}
	if want := []Package{{"com.example", "com/example/package-summary.html"}}; !reflect.DeepEqual(idx.Packages, want) {
		t.Errorf("Packages = %v", idx.Packages)
	}
	wantTypes := []Type{{"com.example", "Foo", "com/example/Foo.html"}, {"com.example", "Foo.Bar", "com/example/Foo.Bar.html"}}
	if !reflect.DeepEqual(idx.Types, wantTypes) {
		t.Errorf("Types = %v", idx.Types)
	}
}

func TestFindJar(t *testing.T) {
	at := func(d time.Duration) *fstest.MapFile {
		return &fstest.MapFile{ModTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(d)}
	}
	fsys := fstest.MapFS{
		"lib/1.0/lib-1.0.jar":                                    at(0),
		"lib/1.0/lib-1.0-javadoc.jar":                            at(0),
		"lib/1.0/other-1.0-javadoc.jar":                          at(time.Hour),
		"lib/2.0-SNAPSHOT/lib-2.0-20260101.000000-1-javadoc.jar": at(0),
		"lib/2.0-SNAPSHOT/lib-2.0-20260102.000000-2-javadoc.jar": at(time.Hour),
		"lib/3.0/lib-3.0.jar":                                    at(0),
	}
	tests := []struct {
		dir, version, want string
	}{
		{"lib/1.0", "1.0", "lib/1.0/lib-1.0-javadoc.jar"},
		{"lib/2.0-SNAPSHOT", "2.0-SNAPSHOT", "lib/2.0-SNAPSHOT/lib-2.0-20260102.000000-2-javadoc.jar"},
		{"lib/3.0", "3.0", ""},
		{"lib/4.0", "4.0", ""},
	}
	for _, tt := range tests {
		got, err := FindJar(fsys, tt.dir, "lib", tt.version)
		if tt.want == "" {
			if !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("FindJar(%s) = %q, %v, want fs.ErrNotExist", tt.dir, got, err)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Errorf("FindJar(%s) = %q, %v, want %q", tt.dir, got, err, tt.want)
		}
	}
}

func TestSigner(t *testing.T) {
	now := time.Date(2026, 10, 31, 15, 0, 0, 0, time.UTC)
	s := &Signer{key: []byte("0123456789abcdef0123456789abcdef"), now: func() time.Time { return now }}

	token := s.Sign("com/example/lib/1.0/lib-1.0-javadoc.jar")
	if got, ok := s.Verify(token); !ok || got != "com/example/lib/1.0/lib-1.0-javadoc.jar" {
		t.Fatalf("Verify = %q, %v", got, ok)
	}

	now = now.Add(10 * time.Hour)
	if again := s.Sign("com/example/lib/1.0/lib-1.0-javadoc.jar"); again == token {
		t.Error("token did not change on the next day")
	}
	if _, ok := s.Verify(token); !ok {
		t.Error("token expired after 10 hours")
	}
	now = now.Add(33 * time.Hour)
	if _, ok := s.Verify(token); ok {
		t.Error("token valid after it expired")
	}

	now = now.Add(-33 * time.Hour)
	other := &Signer{key: []byte("another key, another key, another"), now: s.now}
	if _, ok := other.Verify(token); ok {
		t.Error("token valid with another key")
	}
	if _, ok := s.Verify(token[:len(token)-2] + "AA"); ok {
		t.Error("tampered token valid")
	}
	if _, ok := s.Verify("garbage"); ok {
		t.Error("garbage valid")
	}
}

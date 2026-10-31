package javadoc

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
)

type Package struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Type struct {
	Package string `json:"package"`
	Name    string `json:"name"`
	URL     string `json:"url"`
}

type Member struct {
	Package string `json:"package"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	URL     string `json:"url"`
}

// Page is one of the pages javadoc generates for the whole documentation, like the overview or the index.
type Page struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Index lists the contents of a javadoc jar. URLs are paths of files in the jar, members have an anchor.
type Index struct {
	StartPage string    `json:"startPage"`
	Pages     []Page    `json:"pages"`
	Packages  []Package `json:"packages"`
	Types     []Type    `json:"types"`
	Members   []Member  `json:"members"`
}

// searchEntry is an entry of javadoc's search index. JDK 9 to 16 call the URL field "url", later versions "u".
type searchEntry struct {
	Module  string `json:"m"`
	Package string `json:"p"`
	Class   string `json:"c"`
	Label   string `json:"l"`
	U       string `json:"u"`
	URL     string `json:"url"`
}

func (e searchEntry) url() string {
	if e.U != "" {
		return e.U
	}
	return e.URL
}

const unnamedPackage = "<Unnamed>"

var globalPages = []Page{
	{Title: "Overview", URL: "overview-summary.html"},
	{Title: "Tree", URL: "overview-tree.html"},
	{Title: "Deprecated", URL: "deprecated-list.html"},
	{Title: "Index", URL: "index-all.html"},
	{Title: "Help", URL: "help-doc.html"},
}

// ReadIndex reads the search index javadoc generates since JDK 9. For older javadoc it falls back to the class and
// package lists, which have no members.
func ReadIndex(zr *zip.Reader) (Index, error) {
	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[f.Name] = f
	}
	j := jar{files}

	idx := Index{Pages: []Page{}, Packages: []Package{}, Types: []Type{}, Members: []Member{}}
	packages, err := j.searchIndex("package")
	if errors.Is(err, errNoSearchIndex) {
		if err := j.readLegacy(&idx); err != nil {
			return Index{}, err
		}
	} else if err != nil {
		return Index{}, err
	} else if err := j.readSearchIndex(&idx, packages); err != nil {
		return Index{}, err
	}

	// index.html is a frameset before JDK 9 and the overview afterwards.
	idx.StartPage = "index.html"
	if j.has("overview-summary.html") && !j.has("package-search-index.js") && !j.has("package-search-index.zip") {
		idx.StartPage = "overview-summary.html"
	}
	for _, p := range globalPages {
		if p.Title == "Overview" {
			p.URL = idx.StartPage
		} else if !j.has(p.URL) {
			// JDK 8 without -splitindex and later versions name the index index-all.html, split ones start here
			if p.Title != "Index" || !j.has("index-files/index-1.html") {
				continue
			}
			p.URL = "index-files/index-1.html"
		}
		idx.Pages = append(idx.Pages, p)
	}

	slices.SortFunc(idx.Packages, func(a, b Package) int { return strings.Compare(a.Name, b.Name) })
	slices.SortFunc(idx.Types, func(a, b Type) int {
		return cmpOr(strings.Compare(a.Package, b.Package), strings.Compare(a.Name, b.Name))
	})
	return idx, nil
}

func cmpOr(c ...int) int {
	for _, x := range c {
		if x != 0 {
			return x
		}
	}
	return 0
}

type jar struct {
	files map[string]*zip.File
}

func (j jar) has(name string) bool {
	f, ok := j.files[name]
	return ok && !f.FileInfo().IsDir()
}

func (j jar) read(name string) ([]byte, error) {
	f, ok := j.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

var errNoSearchIndex = errors.New("no search index")

// searchIndex reads the entries of one of the search indexes. JDK 11 and later write them as a script assigning an
// array, JDK 9 and 10 as JSON in a zip file in the jar.
func (j jar) searchIndex(kind string) ([]searchEntry, error) {
	name := kind + "-search-index"
	var data []byte
	if j.has(name + ".js") {
		js, err := j.read(name + ".js")
		if err != nil {
			return nil, err
		}
		start, end := bytes.IndexByte(js, '['), bytes.LastIndexByte(js, ']')
		if start < 0 || end < start {
			return nil, errors.New(name + ".js has no array")
		}
		data = js[start : end+1]
	} else if j.has(name + ".zip") {
		zipped, err := j.read(name + ".zip")
		if err != nil {
			return nil, err
		}
		inner, err := zip.NewReader(bytes.NewReader(zipped), int64(len(zipped)))
		if err != nil {
			return nil, err
		}
		f, err := inner.Open(name + ".json")
		if err != nil {
			return nil, err
		}
		data, err = io.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errNoSearchIndex
	}

	var entries []searchEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, errors.New(name + " is invalid: " + err.Error())
	}
	return entries, nil
}

// pageURL returns the first of the candidate files that exists. Javadoc puts files of named modules in a directory per
// module, but only when documenting more than one module, so both locations are tried.
func (j jar) pageURL(module string, file string) (string, bool) {
	if module != "" && j.has(module+"/"+file) {
		return module + "/" + file, true
	}
	return file, j.has(file)
}

func packageDir(pkg string) string {
	if pkg == unnamedPackage || pkg == "" {
		return ""
	}
	return strings.ReplaceAll(pkg, ".", "/") + "/"
}

func (j jar) readSearchIndex(idx *Index, packages []searchEntry) error {
	for _, e := range packages {
		if e.url() != "" {
			continue
		}
		if url, ok := j.pageURL(e.Module, packageDir(e.Label)+"package-summary.html"); ok {
			idx.Packages = append(idx.Packages, Package{Name: e.Label, URL: url})
		}
	}

	types, err := j.searchIndex("type")
	if err != nil && !errors.Is(err, errNoSearchIndex) {
		return err
	}
	typeURLs := map[string]string{}
	for _, e := range types {
		if e.Package == "" {
			continue
		}
		url, ok := j.pageURL(e.Module, packageDir(e.Package)+e.Label+".html")
		if !ok {
			continue
		}
		typeURLs[e.Package+" "+e.Label] = url
		idx.Types = append(idx.Types, Type{Package: e.Package, Name: e.Label, URL: url})
	}

	members, err := j.searchIndex("member")
	if err != nil && !errors.Is(err, errNoSearchIndex) {
		return err
	}
	for _, e := range members {
		url, ok := typeURLs[e.Package+" "+e.Class]
		if !ok {
			continue
		}
		anchor := e.url()
		if anchor == "" {
			anchor = e.Label
		}
		idx.Members = append(idx.Members, Member{Package: e.Package, Type: e.Class, Name: e.Label, URL: url + "#" + anchor})
	}
	return nil
}

var legacyTypeLink = regexp.MustCompile(`<a href="([^"#]+\.html)" title="[a-z ]+ in ([^"]+)"`)

func (j jar) readLegacy(idx *Index) error {
	for _, list := range []string{"element-list", "package-list"} {
		if !j.has(list) {
			continue
		}
		data, err := j.read(list)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(data), "\n") {
			name := strings.TrimSpace(line)
			if name == "" || strings.HasPrefix(name, "module:") {
				continue
			}
			if url := packageDir(name) + "package-summary.html"; j.has(url) {
				idx.Packages = append(idx.Packages, Package{Name: name, URL: url})
			}
		}
		break
	}

	for _, list := range []string{"allclasses-noframe.html", "allclasses-frame.html", "allclasses.html"} {
		if !j.has(list) {
			continue
		}
		data, err := j.read(list)
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, m := range legacyTypeLink.FindAllStringSubmatch(string(data), -1) {
			url := path.Clean(m[1])
			if seen[url] || !j.has(url) {
				continue
			}
			seen[url] = true
			name := strings.TrimSuffix(path.Base(url), ".html")
			idx.Types = append(idx.Types, Type{Package: m[2], Name: name, URL: url})
		}
		break
	}
	return nil
}

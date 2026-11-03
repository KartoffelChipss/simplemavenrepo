package pom

import (
	"encoding/xml"
	"errors"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"

	"deps.dev/util/maven"
)

var ErrNotFound = errors.New("no pom found")

const maxParents = 16

type Coordinates struct {
	GroupID    string `json:"groupId"`
	ArtifactID string `json:"artifactId"`
	Version    string `json:"version"`
}

type License struct {
	Name string `xml:"name" json:"name"`
	URL  string `xml:"url" json:"url,omitempty"`
}

type Developer struct {
	ID           string `xml:"id" json:"id,omitempty"`
	Name         string `xml:"name" json:"name,omitempty"`
	Email        string `xml:"email" json:"email,omitempty"`
	URL          string `xml:"url" json:"url,omitempty"`
	Organization string `xml:"organization" json:"organization,omitempty"`
}

type SCM struct {
	URL string `json:"url,omitempty"`
	Tag string `json:"tag,omitempty"`
}

type IssueManagement struct {
	System string `json:"system,omitempty"`
	URL    string `json:"url"`
}

type Dependency struct {
	GroupID    string `json:"groupId"`
	ArtifactID string `json:"artifactId"`
	Version    string `json:"version,omitempty"`
	Scope      string `json:"scope"`
	Type       string `json:"type,omitempty"`
	Classifier string `json:"classifier,omitempty"`
	Optional   bool   `json:"optional"`
}

type Project struct {
	Coordinates
	Packaging       string           `json:"packaging"`
	Name            string           `json:"name,omitempty"`
	Description     string           `json:"description,omitempty"`
	URL             string           `json:"url,omitempty"`
	Parent          *Coordinates     `json:"parent,omitempty"`
	Licenses        []License        `json:"licenses"`
	Developers      []Developer      `json:"developers"`
	SCM             *SCM             `json:"scm,omitempty"`
	IssueManagement *IssueManagement `json:"issueManagement,omitempty"`
	Dependencies    []Dependency     `json:"dependencies"`
	Unresolved      []string         `json:"unresolved"`
}

// extra holds what deps.dev's model leaves out.
type extra struct {
	Licenses   []License   `xml:"licenses>license"`
	Developers []Developer `xml:"developers>developer"`
}

type resolver struct {
	fsys       fs.FS
	canRead    func(artifactPath string) bool
	unresolved []string
}

func Resolve(fsys fs.FS, dir, artifactID, version string, canRead func(artifactPath string) bool) (Project, error) {
	r := &resolver{fsys: fsys, canRead: canRead, unresolved: []string{}}
	file, ok := findPom(fsys, dir, artifactID, version)
	if !ok {
		return Project{}, ErrNotFound
	}
	p, ex, err := r.load(file)
	if err != nil {
		return Project{}, err
	}
	r.mergeParents(&p, &ex)

	props := properties(p)
	p.ProcessDependencies(func(g, a, v maven.String) (maven.DependencyManagement, error) {
		return r.dependencyManagement(string(g), string(a), expand(string(v), props))
	})

	result := Project{
		Coordinates: Coordinates{GroupID: string(p.GroupID), ArtifactID: string(p.ArtifactID), Version: string(p.Version)},
		Packaging:   expand(string(p.Packaging), props),
		Name:        expand(string(p.Name), props),
		Description: expand(string(p.Description), props),
		URL:         expand(string(p.URL), props),
		Licenses:    []License{},
		Developers:  []Developer{},
		Unresolved:  r.unresolved,
	}
	if result.Packaging == "" {
		result.Packaging = "jar"
	}
	if p.Parent.ArtifactID != "" {
		result.Parent = &Coordinates{GroupID: string(p.Parent.GroupID), ArtifactID: string(p.Parent.ArtifactID), Version: string(p.Parent.Version)}
	}
	for _, l := range ex.Licenses {
		if l.Name = expand(l.Name, props); l.Name != "" {
			l.URL = expand(l.URL, props)
			result.Licenses = append(result.Licenses, l)
		}
	}
	for _, d := range ex.Developers {
		d = Developer{ID: expand(d.ID, props), Name: expand(d.Name, props), Email: expand(d.Email, props), URL: expand(d.URL, props), Organization: expand(d.Organization, props)}
		if d != (Developer{}) {
			result.Developers = append(result.Developers, d)
		}
	}
	if s := p.SCM; s.URL != "" {
		result.SCM = &SCM{URL: expand(string(s.URL), props), Tag: expand(string(s.Tag), props)}
	}
	if im := p.IssueManagement; im.URL != "" {
		result.IssueManagement = &IssueManagement{System: expand(string(im.System), props), URL: expand(string(im.URL), props)}
	}
	result.Dependencies = []Dependency{}
	for _, d := range p.Dependencies {
		if d.GroupID == "" || d.ArtifactID == "" {
			continue
		}
		dep := Dependency{
			GroupID:    expand(string(d.GroupID), props),
			ArtifactID: expand(string(d.ArtifactID), props),
			Version:    expand(string(d.Version), props),
			Scope:      expand(string(d.Scope), props),
			Type:       expand(string(d.Type), props),
			Classifier: expand(string(d.Classifier), props),
			Optional:   d.Optional.Boolean(),
		}
		if dep.Scope == "" {
			dep.Scope = "compile"
		}
		if dep.Type == "jar" {
			dep.Type = ""
		}
		result.Dependencies = append(result.Dependencies, dep)
	}
	return result, nil
}

func (r *resolver) load(file string) (maven.Project, extra, error) {
	data, err := fs.ReadFile(r.fsys, file)
	if err != nil {
		return maven.Project{}, extra{}, err
	}
	var p maven.Project
	if err := xml.Unmarshal(data, &p); err != nil {
		return maven.Project{}, extra{}, err
	}
	var ex extra
	_ = xml.Unmarshal(data, &ex)
	for i, l := range ex.Licenses {
		ex.Licenses[i] = License{Name: strings.TrimSpace(l.Name), URL: strings.TrimSpace(l.URL)}
	}
	for i, d := range ex.Developers {
		ex.Developers[i] = Developer{
			ID:           strings.TrimSpace(d.ID),
			Name:         strings.TrimSpace(d.Name),
			Email:        strings.TrimSpace(d.Email),
			URL:          strings.TrimSpace(d.URL),
			Organization: strings.TrimSpace(d.Organization),
		}
	}
	_ = p.MergeProfiles("", maven.ActivationOS{})
	return p, ex, nil
}

func (r *resolver) find(g, a, v string) (maven.Project, extra, bool) {
	artifactPath := path.Join(strings.ReplaceAll(g, ".", "/"), a)
	if g != "" && a != "" && v != "" && !strings.Contains(v, "${") && r.canRead(artifactPath) {
		if file, ok := findPom(r.fsys, path.Join(artifactPath, v), a, v); ok {
			if p, ex, err := r.load(file); err == nil {
				return p, ex, true
			}
		}
	}
	r.unresolved = append(r.unresolved, g+":"+a+":"+v)
	return maven.Project{}, extra{}, false
}

func (r *resolver) mergeParents(p *maven.Project, ex *extra) {
	key := p.Parent.ProjectKey
	seen := map[maven.ProjectKey]bool{}
	for i := 0; key.ArtifactID != "" && i < maxParents && !seen[key]; i++ {
		seen[key] = true
		parent, parentExtra, ok := r.find(string(key.GroupID), string(key.ArtifactID), string(key.Version))
		if !ok {
			break
		}
		p.MergeParent(parent)
		if len(ex.Licenses) == 0 {
			ex.Licenses = parentExtra.Licenses
		}
		if len(ex.Developers) == 0 {
			ex.Developers = parentExtra.Developers
		}
		key = parent.Parent.ProjectKey
	}
	// The coordinates are inherited from the parent element even when the parent itself is missing.
	if p.GroupID == "" {
		p.GroupID = p.Parent.GroupID
	}
	if p.Version == "" {
		p.Version = p.Parent.Version
	}
}

func (r *resolver) dependencyManagement(g, a, v string) (maven.DependencyManagement, error) {
	bom, ex, ok := r.find(g, a, v)
	if !ok {
		return maven.DependencyManagement{}, ErrNotFound
	}
	r.mergeParents(&bom, &ex)
	_ = bom.Interpolate()
	return bom.DependencyManagement, nil
}

func findPom(fsys fs.FS, dir, artifactID, version string) (string, bool) {
	exact := artifactID + "-" + version + ".pom"
	if info, err := fs.Stat(fsys, path.Join(dir, exact)); err == nil && !info.IsDir() {
		return path.Join(dir, exact), true
	}
	base, isSnapshot := strings.CutSuffix(version, "SNAPSHOT")
	if !isSnapshot {
		return "", false
	}
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return "", false
	}
	var candidates []string
	for _, e := range entries {
		if n := e.Name(); !e.IsDir() && strings.HasPrefix(n, artifactID+"-"+base) && strings.HasSuffix(n, ".pom") {
			candidates = append(candidates, n)
		}
	}
	if len(candidates) == 0 {
		return "", false
	}
	return path.Join(dir, slices.Max(candidates)), true
}

func properties(p maven.Project) map[string]string {
	m := map[string]string{}
	for _, prop := range p.Properties.Properties {
		m[prop.Name] = prop.Value
	}
	for k, v := range map[string]maven.String{
		"groupId":        p.GroupID,
		"artifactId":     p.ArtifactID,
		"version":        p.Version,
		"name":           p.Name,
		"url":            p.URL,
		"parent.groupId": p.Parent.GroupID,
		"parent.version": p.Parent.Version,
	} {
		if v != "" {
			m["project."+k] = string(v)
			m["pom."+k] = string(v)
		}
	}
	return m
}

var placeholder = regexp.MustCompile(`\$\{([^}]+)}`)

func expand(s string, props map[string]string) string {
	for range 8 {
		next := placeholder.ReplaceAllStringFunc(s, func(m string) string {
			if v, ok := props[m[2:len(m)-1]]; ok {
				return v
			}
			return m
		})
		if next == s {
			break
		}
		s = next
	}
	return s
}

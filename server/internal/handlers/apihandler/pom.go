package apihandler

import (
	"cmp"
	"errors"
	"log"
	"path"
	"simplemavenrepo/internal/pom"
	"simplemavenrepo/internal/repository"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type pomDependency struct {
	pom.Dependency
	Path string `json:"path,omitempty"`
}

type pomResponse struct {
	pom.Project
	Dependencies []pomDependency `json:"dependencies"`
}

func (h *handler) getPom(c fiber.Ctx) error {
	res, _, err := h.resolvePom(c)
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(res)
}

// resolvePom reads the POM of the version at the path of the request. registered holds the coordinates, like
// com.example:demo, of every artifact on this server, including those the caller can't read.
func (h *handler) resolvePom(c fiber.Ctx) (res pomResponse, registered map[string]bool, err error) {
	p := strings.Trim(path.Clean("/"+c.Params("*")), "/")
	notFound := fiber.NewError(fiber.StatusNotFound, "no pom found")

	canRead, err := h.readFilter(c)
	if err != nil {
		return res, nil, err
	}
	artifacts, err := h.repos.Artifacts.List(c.Context())
	if err != nil {
		return res, nil, repoError(err, "artifact")
	}
	readable := map[string]repository.Artifact{}
	byCoordinates := map[string][]repository.Artifact{}
	registered = map[string]bool{}
	for _, a := range artifacts {
		registered[a.GroupID+":"+a.ArtifactID] = true
		if canRead(a) {
			readable[a.Path] = a
			coordinatesPath := strings.TrimPrefix(a.Path, a.Repository+"/")
			byCoordinates[coordinatesPath] = append(byCoordinates[coordinatesPath], a)
		}
	}

	artifact, ok := readable[path.Dir(p)]
	version := path.Base(p)
	if !ok || strings.HasPrefix(version, ".") {
		return res, nil, notFound
	}
	// Prefers the copy in the repository of the artifact, then the others in order of their names
	locate := func(coordinatesPath string) []repository.Artifact {
		copies := slices.Clone(byCoordinates[coordinatesPath])
		slices.SortStableFunc(copies, func(x, y repository.Artifact) int {
			return cmp.Compare(boolRank(x.Repository != artifact.Repository), boolRank(y.Repository != artifact.Repository))
		})
		return copies
	}

	project, err := pom.Resolve(h.repoFS, p, artifact.ArtifactID, version, func(coordinatesPath string) []string {
		var paths []string
		for _, a := range locate(coordinatesPath) {
			paths = append(paths, a.Path)
		}
		return paths
	})
	if errors.Is(err, pom.ErrNotFound) {
		return res, nil, notFound
	}
	if err != nil {
		log.Printf("Reading the POM of %s failed: %v", p, err)
		return res, nil, fiber.NewError(fiber.StatusUnprocessableEntity, "the pom can't be read")
	}

	res = pomResponse{Project: project, Dependencies: []pomDependency{}}
	for _, d := range project.Dependencies {
		dep := pomDependency{Dependency: d}
		if copies := locate(path.Join(strings.ReplaceAll(d.GroupID, ".", "/"), d.ArtifactID)); len(copies) > 0 {
			dep.Path = copies[0].Path
		}
		res.Dependencies = append(res.Dependencies, dep)
	}
	return res, registered, nil
}

func boolRank(b bool) int {
	if b {
		return 1
	}
	return 0
}

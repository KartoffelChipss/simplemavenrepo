package apihandler

import (
	"errors"
	"log"
	"path"
	"simplemavenrepo/internal/pom"
	"simplemavenrepo/internal/repository"
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
	p := strings.Trim(path.Clean("/"+c.Params("*")), "/")
	notFound := fiber.NewError(fiber.StatusNotFound, "no pom found")

	canRead, err := h.readFilter(c)
	if err != nil {
		return err
	}
	artifacts, err := h.repos.Artifacts.List(c.Context())
	if err != nil {
		return repoError(err, "artifact")
	}
	readable := map[string]repository.Artifact{}
	for _, a := range artifacts {
		if canRead(a) {
			readable[a.Path] = a
		}
	}

	artifact, ok := readable[path.Dir(p)]
	version := path.Base(p)
	if !ok || strings.HasPrefix(version, ".") {
		return notFound
	}
	project, err := pom.Resolve(h.repoFS, p, artifact.ArtifactID, version, func(artifactPath string) bool {
		_, ok := readable[artifactPath]
		return ok
	})
	if errors.Is(err, pom.ErrNotFound) {
		return notFound
	}
	if err != nil {
		log.Printf("Reading the POM of %s failed: %v", p, err)
		return fiber.NewError(fiber.StatusUnprocessableEntity, "the pom can't be read")
	}

	res := pomResponse{Project: project, Dependencies: []pomDependency{}}
	for _, d := range project.Dependencies {
		dep := pomDependency{Dependency: d}
		if a, ok := readable[path.Join(strings.ReplaceAll(d.GroupID, ".", "/"), d.ArtifactID)]; ok {
			dep.Path = a.Path
		}
		res.Dependencies = append(res.Dependencies, dep)
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(res)
}

package apihandler

import (
	"cmp"
	"errors"
	"io/fs"
	"log"
	"path"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/handlers/javadochandler"
	"simplemavenrepo/internal/javadoc"
	"simplemavenrepo/internal/maven"
	"simplemavenrepo/internal/repository"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type javadocResponse struct {
	Artifact repository.Artifact `json:"artifact"`
	Version  string              `json:"version"`
	Versions []string            `json:"versions"`
	BaseURL  string              `json:"baseUrl"`
	javadoc.Index
}

type javadocArtifact struct {
	Artifact repository.Artifact `json:"artifact"`
	// Versions that have javadoc, newest first.
	Versions []string `json:"versions"`
}

// listJavadoc lists the artifacts the caller can read that have javadoc for at least one version.
func (h *handler) listJavadoc(c fiber.Ctx) error {
	canRead, err := h.readFilter(c)
	if err != nil {
		return err
	}
	artifacts, err := h.repos.Artifacts.List(c.Context())
	if err != nil {
		return repoError(err, "artifact")
	}

	result := []javadocArtifact{}
	for _, a := range artifacts {
		if !canRead(a) {
			continue
		}
		if versions := h.javadocVersions(a); len(versions) > 0 {
			result = append(result, javadocArtifact{Artifact: a, Versions: versions})
		}
	}
	slices.SortFunc(result, func(x, y javadocArtifact) int {
		return cmp.Or(strings.Compare(x.Artifact.GroupID, y.Artifact.GroupID), strings.Compare(x.Artifact.ArtifactID, y.Artifact.ArtifactID))
	})
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(result)
}

// getJavadoc returns the contents of the javadoc jar of an artifact version, given as its repository path.
func (h *handler) getJavadoc(c fiber.Ctx) error {
	p := strings.Trim(path.Clean("/"+c.Params("*")), "/")
	notFound := fiber.NewError(fiber.StatusNotFound, "no javadoc found")

	artifact, err := h.repos.Artifacts.FindByPath(c.Context(), p)
	if errors.Is(err, repository.ErrNotFound) {
		return notFound
	}
	if err != nil {
		return repoError(err, "artifact")
	}
	version := path.Base(p)
	if path.Dir(p) != artifact.Path || strings.HasPrefix(version, ".") {
		return notFound
	}

	if !artifact.Public {
		_, loggedIn, err := auth.IdentifyOptional(c, h.repos)
		if err != nil {
			return err
		}
		if !loggedIn {
			return notFound
		}
		allowed, err := auth.Authorize(c, h.repos.Grants, p, repository.PermissionRead)
		if err != nil {
			return repoError(err, "grant")
		}
		if !allowed {
			return notFound
		}
	}

	jarPath, err := javadoc.FindJar(h.repoFS, p, artifact.ArtifactID, version)
	if errors.Is(err, fs.ErrNotExist) {
		return notFound
	}
	if err != nil {
		return err
	}
	jar, err := javadoc.OpenJar(h.repoFS, jarPath)
	if err != nil {
		log.Printf("Opening javadoc %s failed: %v", jarPath, err)
		return fiber.NewError(fiber.StatusUnprocessableEntity, "the javadoc jar can't be read")
	}
	defer jar.Close()
	index, err := javadoc.ReadIndex(jar.Reader)
	if err != nil {
		log.Printf("Reading javadoc index of %s failed: %v", jarPath, err)
		return fiber.NewError(fiber.StatusUnprocessableEntity, "the javadoc jar can't be read")
	}

	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(javadocResponse{
		Artifact: artifact,
		Version:  version,
		Versions: h.javadocVersions(artifact),
		BaseURL:  javadochandler.Prefix + h.signer.Sign(jarPath) + "/",
		Index:    index,
	})
}

func (h *handler) javadocVersions(artifact repository.Artifact) []string {
	entries, err := fs.ReadDir(h.repoFS, artifact.Path)
	if err != nil {
		return []string{}
	}
	versions := []string{}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if _, err := javadoc.FindJar(h.repoFS, path.Join(artifact.Path, e.Name()), artifact.ArtifactID, e.Name()); err == nil {
			versions = append(versions, e.Name())
		}
	}
	slices.SortFunc(versions, func(a, b string) int { return maven.CompareVersions(b, a) })
	return versions
}

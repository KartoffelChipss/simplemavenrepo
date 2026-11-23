package apihandler

import (
	"errors"
	"io/fs"
	"log"
	"path"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/handlers/ressourcehandler"
	"simplemavenrepo/internal/repository"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// deleteVersion deletes a version of an artifact, given as its path, e.g. releases/com/example/demo/1.0. It requires
// write access to the version.
func (h *handler) deleteVersion(c fiber.Ctx) error {
	p := strings.Trim(path.Clean("/"+c.Params("*")), "/")
	notFound := fiber.NewError(fiber.StatusNotFound, "version not found")

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

	allowed, err := auth.Permissions(c, h.repos.Grants)
	if err != nil {
		return repoError(err, "grant")
	}
	if !allowed(p, repository.PermissionWrite) {
		if !artifact.Public && !allowed(artifact.Path, repository.PermissionRead) {
			return notFound
		}
		return fiber.NewError(fiber.StatusForbidden, "no write permission for "+artifact.GroupID+":"+artifact.ArtifactID)
	}

	all, err := h.repos.Artifacts.List(c.Context())
	if err != nil {
		return repoError(err, "artifact")
	}
	for _, other := range all {
		if strings.HasPrefix(other.Path, p+"/") {
			return fiber.NewError(fiber.StatusConflict,
				"the version contains the artifact "+other.GroupID+":"+other.ArtifactID+", delete that one first")
		}
	}

	repoRoot, err := h.root.OpenRoot(artifact.Repository)
	if errors.Is(err, fs.ErrNotExist) {
		return notFound
	}
	if err != nil {
		log.Printf("Opening the directory of repository %s failed: %v", artifact.Repository, err)
		return fiber.ErrInternalServerError
	}
	defer repoRoot.Close()

	artifactDir := strings.TrimPrefix(artifact.Path, artifact.Repository+"/")
	err = ressourcehandler.DeleteVersion(repoRoot, artifact.Repository, artifactDir, artifact.GroupID, artifact.ArtifactID, version)
	if errors.Is(err, fs.ErrNotExist) {
		return notFound
	}
	if err != nil {
		log.Printf("Deleting %s failed: %v", p, err)
		return fiber.ErrInternalServerError
	}
	log.Printf("%s deleted %s", auth.CurrentUser(c).Username, p)
	return c.SendStatus(fiber.StatusNoContent)
}

package apihandler

import (
	"errors"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/browse"
	"simplemavenrepo/internal/repository"

	"github.com/gofiber/fiber/v3"
)

// permissions returns a function that reports whether the caller, who may be anonymous, has a permission on a
// repository path. Anonymous callers have none.
func (h *handler) permissions(c fiber.Ctx) (func(string, repository.Permission) bool, error) {
	_, loggedIn, err := auth.IdentifyOptional(c, h.repos)
	if err != nil {
		return nil, err
	}
	if !loggedIn {
		return func(string, repository.Permission) bool { return false }, nil
	}
	allowed, err := auth.Permissions(c, h.repos.Grants)
	if err != nil {
		return nil, repoError(err, "grant")
	}
	return allowed, nil
}

// readFilter returns a function that reports whether the caller, who may be anonymous, can read an artifact.
func (h *handler) readFilter(c fiber.Ctx) (func(repository.Artifact) bool, error) {
	allowed, err := h.permissions(c)
	if err != nil {
		return nil, err
	}
	return func(a repository.Artifact) bool {
		return a.Public || allowed(a.Path, repository.PermissionRead)
	}, nil
}

// listBrowsableArtifacts lists the artifacts the caller can read, with their latest version.
func (h *handler) listBrowsableArtifacts(c fiber.Ctx) error {
	canRead, err := h.readFilter(c)
	if err != nil {
		return err
	}
	artifacts, err := h.repos.Artifacts.List(c.Context())
	if err != nil {
		return repoError(err, "artifact")
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(browse.Artifacts(h.repoFS, artifacts, canRead))
}

// browseRepository lists a directory inside an artifact.
func (h *handler) browseRepository(c fiber.Ctx) error {
	allowed, err := h.permissions(c)
	if err != nil {
		return err
	}
	canRead := func(a repository.Artifact) bool { return a.Public || allowed(a.Path, repository.PermissionRead) }
	artifacts, err := h.repos.Artifacts.List(c.Context())
	if err != nil {
		return repoError(err, "artifact")
	}

	listing, err := browse.List(h.repoFS, artifacts, canRead, c.Params("*"))
	if errors.Is(err, browse.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}
	if err != nil {
		return err
	}
	listing.CanWrite = allowed(listing.Artifact.Path, repository.PermissionWrite)
	// The listing depends on who is asking.
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(listing)
}

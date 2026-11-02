package apihandler

import (
	"errors"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/browse"
	"simplemavenrepo/internal/repository"

	"github.com/gofiber/fiber/v3"
)

// readFilter returns a function that reports whether the caller, who may be anonymous, can read an artifact.
func (h *handler) readFilter(c fiber.Ctx) (func(repository.Artifact) bool, error) {
	_, loggedIn, err := auth.IdentifyOptional(c, h.repos)
	if err != nil {
		return nil, err
	}
	var allowed func(string, repository.Permission) bool
	if loggedIn {
		if allowed, err = auth.Permissions(c, h.repos.Grants); err != nil {
			return nil, repoError(err, "grant")
		}
	}
	return func(a repository.Artifact) bool {
		return a.Public || (loggedIn && allowed(a.Path, repository.PermissionRead))
	}, nil
}

// browseRepository lists a directory of the repository.
func (h *handler) browseRepository(c fiber.Ctx) error {
	canRead, err := h.readFilter(c)
	if err != nil {
		return err
	}
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
	// The listing depends on who is asking.
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(listing)
}

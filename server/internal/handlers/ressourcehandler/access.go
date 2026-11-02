package ressourcehandler

import (
	"errors"
	"log"
	"path"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/repository"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// newReadGuard only serves files of registered artifacts. Private artifacts additionally require a token with read access.
func newReadGuard(repos repository.Repositories) fiber.Handler {
	return func(c fiber.Ctx) error {
		relPath := path.Clean("/" + strings.TrimPrefix(c.Path(), "/releases"))

		artifact, err := repos.Artifacts.FindByPath(c.Context(), relPath)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return fiber.ErrNotFound
			}
			log.Printf("Looking up artifact for %s failed: %v", relPath, err)
			return fiber.ErrInternalServerError
		}
		if artifact.Public {
			return c.Next()
		}

		if err := auth.Identify(c, repos); err != nil {
			return err
		}
		allowed, err := auth.Authorize(c, repos.Grants, relPath, repository.PermissionRead)
		if err != nil {
			log.Printf("Checking permissions for %s failed: %v", relPath, err)
			return fiber.ErrInternalServerError
		}
		if !allowed {
			return fiber.NewError(fiber.StatusForbidden, "no read permission for "+artifact.GroupID+":"+artifact.ArtifactID)
		}

		if err := c.Next(); err != nil {
			return err
		}
		// Shared caches must not keep files of private artifacts
		if cc := string(c.Response().Header.Peek(fiber.HeaderCacheControl)); cc != "" {
			c.Set(fiber.HeaderCacheControl, strings.Replace(cc, "public", "private", 1))
		}
		return nil
	}
}

package ressourcehandler

import (
	"context"
	"errors"
	"log"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/maven"
	"simplemavenrepo/internal/repository"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// readGuard is called with the path of a requested file, e.g. releases/com/example/demo/1.0/demo-1.0.jar, and only lets
// serve send files of registered artifacts. Private artifacts additionally require a token with read access.
type readGuard func(c fiber.Ctx, relPath string, serve fiber.Handler) error

func newReadGuard(repos repository.Repositories) readGuard {
	return func(c fiber.Ctx, relPath string, serve fiber.Handler) error {
		artifact, err := repos.Artifacts.FindByPath(c.Context(), relPath)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return fiber.ErrNotFound
			}
			log.Printf("Looking up artifact for %s failed: %v", relPath, err)
			return fiber.ErrInternalServerError
		}
		if artifact.Public {
			if err := serve(c); err != nil {
				return err
			}
			countDownload(c, repos, artifact, relPath)
			return nil
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

		if err := serve(c); err != nil {
			return err
		}
		// Shared caches must not keep files of private artifacts
		if cc := string(c.Response().Header.Peek(fiber.HeaderCacheControl)); cc != "" {
			c.Set(fiber.HeaderCacheControl, strings.Replace(cc, "public", "private", 1))
		}
		countDownload(c, repos, artifact, relPath)
		return nil
	}
}

// downloadExtensions are the extensions of files counted as downloads. POMs and Gradle module files are left out since
// builds fetch them along with the main file, which would count every download twice.
var downloadExtensions = slices.DeleteFunc(slices.Clone(storedExtensions), func(ext string) bool {
	return ext == ".pom" || ext == ".module"
})

// countDownload records a download if relPath was served in full and is the main file of a version, so fetching
// checksums, signatures, sources or javadoc doesn't count. A failure to count doesn't fail the download.
func countDownload(c fiber.Ctx, repos repository.Repositories, artifact repository.Artifact, relPath string) {
	if c.Method() != fiber.MethodGet || c.Response().StatusCode() != fiber.StatusOK {
		return
	}
	version, file, ok := strings.Cut(strings.TrimPrefix(relPath, artifact.Path+"/"), "/")
	if !ok || strings.Contains(file, "/") || !maven.IsMainFile(artifact.ArtifactID, version, file, downloadExtensions) {
		return
	}
	ctx := context.WithoutCancel(c.Context())
	if err := repos.Downloads.Record(ctx, artifact.Path, version, time.Now()); err != nil {
		log.Printf("Counting download of %s failed: %v", relPath, err)
	}
}

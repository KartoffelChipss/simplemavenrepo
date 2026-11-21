package ressourcehandler

import (
	"errors"
	"log"
	"os"
	"path"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/config"
	"simplemavenrepo/internal/repository"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

const repoKey = "smr.repository"

// RegisterRessourceRoutes serves every repository under /<name>. Requests whose first segment is not a repository fall
// through to the routes registered after these, i.e. the web UI.
func RegisterRessourceRoutes(app *fiber.App, repos repository.Repositories, root *os.Root) {
	notFound := func(c fiber.Ctx) error { return fiber.ErrNotFound }
	volatile := static.New(config.GetRepoPath(), static.Config{
		ByteRange:       true,
		MaxAge:          0,
		CacheDuration:   -1,
		NotFoundHandler: notFound,
	})
	immutable := static.New(config.GetRepoPath(), static.Config{
		ByteRange:       true,
		MaxAge:          31536000, // 1 year
		CacheDuration:   10 * time.Minute,
		NotFoundHandler: notFound,
	})
	readGuard := newReadGuard(repos)

	app.Get("/*", func(c fiber.Ctx) error {
		if ok, err := resolveRepo(c, repos); err != nil {
			return err
		} else if !ok {
			return c.Next()
		}
		relPath := strings.Trim(path.Clean("/"+c.Path()), "/")
		serve := immutable
		if isVolatilePath(relPath) {
			serve = volatile
		}
		return readGuard(c, relPath, serve)
	})

	app.Put("/*", func(c fiber.Ctx) error {
		ok, err := resolveRepo(c, repos)
		if err == nil && !ok {
			name, _, _ := strings.Cut(strings.TrimPrefix(c.Path(), "/"), "/")
			// Not 404: Maven reports that as "Could not find artifact" and drops the message
			err = fiber.NewError(fiber.StatusBadRequest, "there is no repository named "+name+" on this server")
		}
		if err == nil {
			err = c.Next()
		}
		if err != nil {
			return sendUploadError(c, err)
		}
		return nil
	}, auth.AuthenticateScoped(repos), newUploadHandler(root, repos))
}

// maxReasonLength keeps the status line short, some clients limit its length
const maxReasonLength = 200

// sendUploadError responds with the message of err as the body and as the reason phrase of the status line.
func sendUploadError(c fiber.Ctx, err error) error {
	code, msg := fiber.StatusInternalServerError, "internal server error"
	var fe *fiber.Error
	if errors.As(err, &fe) {
		code, msg = fe.Code, fe.Message
	}
	if code == fiber.StatusInternalServerError {
		msg = "internal server error"
	}
	c.Status(code)
	c.Response().Header.SetStatusMessage([]byte(reasonPhrase(msg)))
	return c.SendString(msg)
}

// reasonPhrase makes msg safe for the status line, which can't hold line breaks and should be plain ASCII.
func reasonPhrase(msg string) string {
	var b strings.Builder
	for _, r := range msg {
		if b.Len() >= maxReasonLength {
			break
		}
		if r < 0x20 || r > 0x7e {
			r = ' '
		}
		b.WriteRune(r)
	}
	return b.String()
}

// resolveRepo finds the repository named by the first path segment and stores it for the next handlers.
func resolveRepo(c fiber.Ctx, repos repository.Repositories) (bool, error) {
	name, _, _ := strings.Cut(strings.TrimPrefix(c.Path(), "/"), "/")
	if name == "" || strings.HasPrefix(name, ".") {
		return false, nil
	}
	repo, err := repos.Repos.Get(c.Context(), name)
	if errors.Is(err, repository.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		log.Printf("Looking up repository %s failed: %v", name, err)
		return false, fiber.ErrInternalServerError
	}
	c.Locals(repoKey, repo)
	return true, nil
}

func currentRepo(c fiber.Ctx) repository.Repo {
	repo, _ := c.Locals(repoKey).(repository.Repo)
	return repo
}

// isVolatilePath reports whether the file at p can change, so it must not be cached: metadata and snapshot files.
func isVolatilePath(p string) bool {
	return strings.HasPrefix(path.Base(p), "maven-metadata.xml") || strings.HasSuffix(path.Base(path.Dir(p)), "-SNAPSHOT")
}

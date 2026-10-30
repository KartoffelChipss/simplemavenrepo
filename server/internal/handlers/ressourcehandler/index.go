package ressourcehandler

import (
	"log"
	"os"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/config"
	"simplemavenrepo/internal/repository"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func RegisterRessourceRoutes(app *fiber.App, repos repository.Repositories) {
	releasesGroup := app.Group("/releases")

	volatile := static.New(config.GetRepoPath(), static.Config{
		ByteRange:     true,
		MaxAge:        0,
		CacheDuration: -1,
	})
	readGuard := newReadGuard(repos)
	releasesGroup.Get("/*/maven-metadata.xml*", readGuard, volatile)
	releasesGroup.Get("/*-SNAPSHOT/*", readGuard, volatile)

	releasesGroup.Get("/*", readGuard, static.New(config.GetRepoPath(), static.Config{
		ByteRange:     true,
		MaxAge:        31536000, // 1 year
		CacheDuration: 10 * time.Minute,
	}))

	root, err := os.OpenRoot(config.GetRepoPath())
	if err != nil {
		log.Fatalf("Cannot open repository path: %v", err)
	}
	releasesGroup.Put("/*", auth.Authenticate(repos.Tokens), newUploadHandler(root, repos))
}

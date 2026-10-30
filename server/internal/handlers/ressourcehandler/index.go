package ressourcehandler

import (
	"log"
	"os"
	"simplemavenrepo/internal/config"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func RegisterRessourceRoutes(app *fiber.App) {
	releasesGroup := app.Group("/releases")

	volatile := static.New(config.GetRepoPath(), static.Config{
		ByteRange:     true,
		MaxAge:        0,
		CacheDuration: -1,
	})
	releasesGroup.Get("/*/maven-metadata.xml*", volatile)
	releasesGroup.Get("/*-SNAPSHOT/*", volatile)

	releasesGroup.Get("/*", static.New(config.GetRepoPath(), static.Config{
		ByteRange:     true,
		MaxAge:        31536000, // 1 year
		CacheDuration: 10 * time.Minute,
	}))

	root, err := os.OpenRoot(config.GetRepoPath())
	if err != nil {
		log.Fatalf("Cannot open repository path: %v", err)
	}
	releasesGroup.Put("/*", newUploadHandler(root))
}

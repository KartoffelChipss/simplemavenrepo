package webui

import (
	"embed"
	"io/fs"
	"log"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

//go:embed all:dist
var files embed.FS

func RegisterRoutes(app *fiber.App) {
	dist, err := fs.Sub(files, "dist")
	if err != nil {
		log.Fatalf("Web UI: %v", err)
	}
	assets, err := fs.Sub(dist, "assets")
	if err != nil {
		log.Fatalf("Web UI: %v", err)
	}
	index, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		log.Printf("Web UI is not built, run `pnpm build` in web/ to include it")
		return
	}

	serveIndex := func(c fiber.Ctx) error {
		c.Status(fiber.StatusOK)
		c.Set(fiber.HeaderCacheControl, "no-cache")
		c.Type("html")
		return c.Send(index)
	}
	isBackendPath := func(c fiber.Ctx) bool {
		p := c.Path()
		return strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/releases/") || strings.HasPrefix(p, "/javadoc/")
	}

	app.Get("/assets/*", static.New("", static.Config{
		FS:              assets,
		MaxAge:          31536000, // 1 year
		NotFoundHandler: func(c fiber.Ctx) error { return fiber.ErrNotFound },
	}))
	app.Get("/", serveIndex)
	app.Get("/*", static.New("", static.Config{
		FS:              dist,
		Next:            isBackendPath,
		NotFoundHandler: serveIndex,
		IndexNames:      []string{},
	}))
}

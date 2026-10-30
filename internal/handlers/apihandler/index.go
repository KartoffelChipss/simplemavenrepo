package apihandler

import "github.com/gofiber/fiber/v3"

func RegisterApiRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Use(func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Next()
	})

	api.Get("/about", aboutHandler)
}

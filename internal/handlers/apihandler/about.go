package apihandler

import "github.com/gofiber/fiber/v3"

func aboutHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"version": "1.0.0",
	})
}

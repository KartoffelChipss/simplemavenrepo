package apihandler

import (
	"errors"
	"simplemavenrepo/internal/repository"

	"github.com/gofiber/fiber/v3"
)

type handler struct {
	repos repository.Repositories
}

func RegisterApiRoutes(app *fiber.App, repos repository.Repositories) {
	h := &handler{repos: repos}
	api := app.Group("/api")

	api.Use(func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		err := c.Next()
		if err == nil {
			return nil
		}
		code := fiber.StatusInternalServerError
		var fe *fiber.Error
		if errors.As(err, &fe) {
			code = fe.Code
		}
		msg := err.Error()
		if code == fiber.StatusInternalServerError {
			msg = "internal server error"
		}
		return c.Status(code).JSON(fiber.Map{"error": msg})
	})

	api.Get("/about", aboutHandler)

	me := api.Group("/me", h.authenticate)
	me.Get("", h.getMe)
	me.Get("/tokens", h.listMyTokens)
	me.Post("/tokens", h.createMyToken)
	me.Delete("/tokens/:id", h.deleteMyToken)

	users := api.Group("/users", h.authenticate, requireAdmin)
	users.Get("", h.listUsers)
	users.Post("", h.createUser)
	users.Get("/:username", h.getUser)
	users.Patch("/:username", h.updateUser)
	users.Delete("/:username", h.deleteUser)
	users.Get("/:username/tokens", h.listUserTokens)
	users.Post("/:username/tokens", h.createUserToken)
	users.Delete("/:username/tokens/:id", h.deleteUserToken)

	teams := api.Group("/teams", h.authenticate, requireAdmin)
	teams.Get("", h.listTeams)
	teams.Post("", h.createTeam)
	teams.Get("/:team", h.getTeam)
	teams.Delete("/:team", h.deleteTeam)
	teams.Put("/:team/members/:username", h.addTeamMember)
	teams.Delete("/:team/members/:username", h.removeTeamMember)

	grants := api.Group("/grants", h.authenticate, requireAdmin)
	grants.Get("", h.listGrants)
	grants.Post("", h.createGrant)
	grants.Delete("/:id", h.deleteGrant)
}

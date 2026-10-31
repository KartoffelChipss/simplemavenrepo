package apihandler

import (
	"errors"
	"io/fs"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/javadoc"
	"simplemavenrepo/internal/repository"

	"github.com/gofiber/fiber/v3"
)

type handler struct {
	repos  repository.Repositories
	repoFS fs.FS
	signer *javadoc.Signer
}

func RegisterApiRoutes(app *fiber.App, repos repository.Repositories, repoFS fs.FS, signer *javadoc.Signer) {
	h := &handler{repos: repos, repoFS: repoFS, signer: signer}
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
		// The UI handles 401 itself, browsers must not show their own Basic auth prompt.
		if code == fiber.StatusUnauthorized {
			c.Response().Header.Del(fiber.HeaderWWWAuthenticate)
		}
		if code == fiber.StatusInternalServerError {
			msg = "internal server error"
		}
		return c.Status(code).JSON(fiber.Map{"error": msg})
	})

	api.Get("/about", aboutHandler)
	api.Post("/auth/login", loginLimiter, h.login)
	api.Post("/auth/logout", h.logout)
	api.Get("/browse", h.browseRepository)
	api.Get("/browse/*", h.browseRepository)
	api.Get("/javadoc", h.listJavadoc)
	api.Get("/javadoc/*", h.getJavadoc)

	authenticate := auth.Authenticate(repos)

	me := api.Group("/me", authenticate)
	me.Get("", h.getMe)
	me.Put("/password", h.changeMyPassword)
	me.Get("/tokens", h.listMyTokens)
	me.Post("/tokens", h.createMyToken)
	me.Delete("/tokens/:id", h.deleteMyToken)

	users := api.Group("/users", authenticate, auth.RequireAdmin)
	users.Get("", h.listUsers)
	users.Post("", h.createUser)
	users.Get("/:username", h.getUser)
	users.Patch("/:username", h.updateUser)
	users.Delete("/:username", h.deleteUser)
	users.Get("/:username/tokens", h.listUserTokens)
	users.Post("/:username/tokens", h.createUserToken)
	users.Delete("/:username/tokens/:id", h.deleteUserToken)

	teams := api.Group("/teams", authenticate, auth.RequireAdmin)
	teams.Get("", h.listTeams)
	teams.Post("", h.createTeam)
	teams.Get("/:team", h.getTeam)
	teams.Delete("/:team", h.deleteTeam)
	teams.Put("/:team/members/:username", h.addTeamMember)
	teams.Delete("/:team/members/:username", h.removeTeamMember)

	grants := api.Group("/grants", authenticate, auth.RequireAdmin)
	grants.Get("", h.listGrants)
	grants.Post("", h.createGrant)
	grants.Delete("/:id", h.deleteGrant)

	artifacts := api.Group("/artifacts", authenticate, auth.RequireAdmin)
	artifacts.Get("", h.listArtifacts)
	artifacts.Post("", h.createArtifact)
	artifacts.Get("/:groupId/:artifactId", h.getArtifact)
	artifacts.Patch("/:groupId/:artifactId", h.updateArtifact)
	artifacts.Delete("/:groupId/:artifactId", h.deleteArtifact)
}

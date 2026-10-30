package apihandler

import (
	"simplemavenrepo/internal/repository"
	"time"

	"github.com/gofiber/fiber/v3"
)

type createdToken struct {
	repository.Token
	// The plaintext token is only returned once, on creation
	Secret string `json:"token"`
}

func (h *handler) listTokens(c fiber.Ctx, u repository.User) error {
	tokens, err := h.repos.Tokens.ListForUser(c.Context(), u.ID)
	if err != nil {
		return repoError(err, "token")
	}
	return c.JSON(tokens)
}

func (h *handler) createToken(c fiber.Ctx, u repository.User) error {
	var req struct {
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if req.Name == "" || len(req.Name) > 100 {
		return fiber.NewError(fiber.StatusBadRequest, "name must be 1-100 characters")
	}
	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now()) {
		return fiber.NewError(fiber.StatusBadRequest, "expiresAt must be in the future")
	}

	t, secret, err := h.repos.Tokens.Create(c.Context(), u.ID, req.Name, req.ExpiresAt)
	if err != nil {
		return repoError(err, "token")
	}
	return c.Status(fiber.StatusCreated).JSON(createdToken{Token: t, Secret: secret})
}

func (h *handler) deleteToken(c fiber.Ctx, u repository.User) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.repos.Tokens.Delete(c.Context(), u.ID, id); err != nil {
		return repoError(err, "token")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *handler) listMyTokens(c fiber.Ctx) error  { return h.listTokens(c, currentUser(c)) }
func (h *handler) createMyToken(c fiber.Ctx) error { return h.createToken(c, currentUser(c)) }
func (h *handler) deleteMyToken(c fiber.Ctx) error { return h.deleteToken(c, currentUser(c)) }

func (h *handler) withUserParam(c fiber.Ctx, next func(fiber.Ctx, repository.User) error) error {
	u, err := h.repos.Users.Get(c.Context(), c.Params("username"))
	if err != nil {
		return repoError(err, "user")
	}
	return next(c, u)
}

func (h *handler) listUserTokens(c fiber.Ctx) error  { return h.withUserParam(c, h.listTokens) }
func (h *handler) createUserToken(c fiber.Ctx) error { return h.withUserParam(c, h.createToken) }
func (h *handler) deleteUserToken(c fiber.Ctx) error { return h.withUserParam(c, h.deleteToken) }

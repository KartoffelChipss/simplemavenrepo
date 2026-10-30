package apihandler

import (
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/repository"

	"github.com/gofiber/fiber/v3"
)

type userDetails struct {
	repository.User
	Teams  []repository.Team  `json:"teams"`
	Grants []repository.Grant `json:"grants"`
}

func (h *handler) userDetails(c fiber.Ctx, u repository.User) error {
	teams, err := h.repos.Teams.ListForUser(c.Context(), u.ID)
	if err != nil {
		return repoError(err, "user")
	}
	grants, err := h.repos.Grants.ListEffective(c.Context(), u.ID)
	if err != nil {
		return repoError(err, "user")
	}
	return c.JSON(userDetails{User: u, Teams: teams, Grants: grants})
}

func (h *handler) getMe(c fiber.Ctx) error {
	return h.userDetails(c, auth.CurrentUser(c))
}

func (h *handler) listUsers(c fiber.Ctx) error {
	users, err := h.repos.Users.List(c.Context())
	if err != nil {
		return repoError(err, "user")
	}
	return c.JSON(users)
}

func (h *handler) getUser(c fiber.Ctx) error {
	u, err := h.repos.Users.Get(c.Context(), c.Params("username"))
	if err != nil {
		return repoError(err, "user")
	}
	return h.userDetails(c, u)
}

func (h *handler) createUser(c fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		IsAdmin  bool   `json:"isAdmin"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := validateName("username", req.Username); err != nil {
		return err
	}

	u, err := h.repos.Users.Create(c.Context(), req.Username, req.IsAdmin)
	if err != nil {
		return repoError(err, "user")
	}
	return c.Status(fiber.StatusCreated).JSON(u)
}

func (h *handler) updateUser(c fiber.Ctx) error {
	var req struct {
		IsAdmin *bool `json:"isAdmin"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}

	u, err := h.repos.Users.Get(c.Context(), c.Params("username"))
	if err != nil {
		return repoError(err, "user")
	}
	if req.IsAdmin == nil || *req.IsAdmin == u.IsAdmin {
		return c.JSON(u)
	}
	if !*req.IsAdmin {
		if err := h.ensureNotLastAdmin(c, u); err != nil {
			return err
		}
	}

	u, err = h.repos.Users.SetAdmin(c.Context(), u.Username, *req.IsAdmin)
	if err != nil {
		return repoError(err, "user")
	}
	return c.JSON(u)
}

func (h *handler) deleteUser(c fiber.Ctx) error {
	u, err := h.repos.Users.Get(c.Context(), c.Params("username"))
	if err != nil {
		return repoError(err, "user")
	}
	if err := h.ensureNotLastAdmin(c, u); err != nil {
		return err
	}
	if err := h.repos.Users.Delete(c.Context(), u.Username); err != nil {
		return repoError(err, "user")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ensureNotLastAdmin prevents locking everyone out of the management API.
func (h *handler) ensureNotLastAdmin(c fiber.Ctx, u repository.User) error {
	if !u.IsAdmin {
		return nil
	}
	admins, err := h.repos.Users.CountAdmins(c.Context())
	if err != nil {
		return repoError(err, "user")
	}
	if admins <= 1 {
		return fiber.NewError(fiber.StatusConflict, "cannot remove the last admin")
	}
	return nil
}

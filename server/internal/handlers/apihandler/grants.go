package apihandler

import (
	"simplemavenrepo/internal/repository"

	"github.com/gofiber/fiber/v3"
)

// listGrants lists all grants, optionally filtered with ?user=<username> or ?team=<name>.
func (h *handler) listGrants(c fiber.Ctx) error {
	var f repository.GrantFilter
	if name := c.Query("user"); name != "" {
		u, err := h.repos.Users.Get(c.Context(), name)
		if err != nil {
			return repoError(err, "user")
		}
		f.UserID = u.ID
	}
	if name := c.Query("team"); name != "" {
		t, err := h.repos.Teams.Get(c.Context(), name)
		if err != nil {
			return repoError(err, "team")
		}
		f.TeamID = t.ID
	}

	grants, err := h.repos.Grants.List(c.Context(), f)
	if err != nil {
		return repoError(err, "grant")
	}
	return c.JSON(grants)
}

func (h *handler) createGrant(c fiber.Ctx) error {
	var req struct {
		User       string                `json:"user"`
		Team       string                `json:"team"`
		Pattern    string                `json:"pattern"`
		Permission repository.Permission `json:"permission"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if (req.User == "") == (req.Team == "") {
		return fiber.NewError(fiber.StatusBadRequest, "exactly one of user or team is required")
	}
	if !req.Permission.Valid() {
		return fiber.NewError(fiber.StatusBadRequest, `permission must be "read" or "write"`)
	}
	pattern, err := repository.NormalizePattern(req.Pattern)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var g repository.Grant
	if req.User != "" {
		u, uErr := h.repos.Users.Get(c.Context(), req.User)
		if uErr != nil {
			return repoError(uErr, "user")
		}
		g, err = h.repos.Grants.CreateForUser(c.Context(), u.ID, pattern, req.Permission)
	} else {
		t, tErr := h.repos.Teams.Get(c.Context(), req.Team)
		if tErr != nil {
			return repoError(tErr, "team")
		}
		g, err = h.repos.Grants.CreateForTeam(c.Context(), t.ID, pattern, req.Permission)
	}
	if err != nil {
		return repoError(err, "grant")
	}
	return c.Status(fiber.StatusCreated).JSON(g)
}

func (h *handler) deleteGrant(c fiber.Ctx) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.repos.Grants.Delete(c.Context(), id); err != nil {
		return repoError(err, "grant")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

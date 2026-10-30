package apihandler

import (
	"simplemavenrepo/internal/repository"

	"github.com/gofiber/fiber/v3"
)

type teamDetails struct {
	repository.Team
	Members []repository.User  `json:"members"`
	Grants  []repository.Grant `json:"grants"`
}

func (h *handler) listTeams(c fiber.Ctx) error {
	teams, err := h.repos.Teams.List(c.Context())
	if err != nil {
		return repoError(err, "team")
	}
	return c.JSON(teams)
}

func (h *handler) getTeam(c fiber.Ctx) error {
	t, err := h.repos.Teams.Get(c.Context(), c.Params("team"))
	if err != nil {
		return repoError(err, "team")
	}
	members, err := h.repos.Teams.ListMembers(c.Context(), t.ID)
	if err != nil {
		return repoError(err, "team")
	}
	grants, err := h.repos.Grants.List(c.Context(), repository.GrantFilter{TeamID: t.ID})
	if err != nil {
		return repoError(err, "team")
	}
	return c.JSON(teamDetails{Team: t, Members: members, Grants: grants})
}

func (h *handler) createTeam(c fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := validateName("team name", req.Name); err != nil {
		return err
	}

	t, err := h.repos.Teams.Create(c.Context(), req.Name)
	if err != nil {
		return repoError(err, "team")
	}
	return c.Status(fiber.StatusCreated).JSON(t)
}

func (h *handler) deleteTeam(c fiber.Ctx) error {
	if err := h.repos.Teams.Delete(c.Context(), c.Params("team")); err != nil {
		return repoError(err, "team")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *handler) teamAndUserParams(c fiber.Ctx) (repository.Team, repository.User, error) {
	t, err := h.repos.Teams.Get(c.Context(), c.Params("team"))
	if err != nil {
		return t, repository.User{}, repoError(err, "team")
	}
	u, err := h.repos.Users.Get(c.Context(), c.Params("username"))
	if err != nil {
		return t, u, repoError(err, "user")
	}
	return t, u, nil
}

func (h *handler) addTeamMember(c fiber.Ctx) error {
	t, u, err := h.teamAndUserParams(c)
	if err != nil {
		return err
	}
	if err := h.repos.Teams.AddMember(c.Context(), t.ID, u.ID); err != nil {
		return repoError(err, "team member")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *handler) removeTeamMember(c fiber.Ctx) error {
	t, u, err := h.teamAndUserParams(c)
	if err != nil {
		return err
	}
	if err := h.repos.Teams.RemoveMember(c.Context(), t.ID, u.ID); err != nil {
		return repoError(err, "team member")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

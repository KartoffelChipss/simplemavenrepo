package apihandler

import (
	"regexp"
	"simplemavenrepo/internal/repository"

	"github.com/gofiber/fiber/v3"
)

var (
	groupIDPattern    = regexp.MustCompile(`^[A-Za-z0-9_-]+(\.[A-Za-z0-9_-]+)*$`)
	artifactIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-][A-Za-z0-9_.-]*$`)
)

const maxDescriptionLength = 1000

func validateDescription(d *string) error {
	if d != nil && len(*d) > maxDescriptionLength {
		return fiber.NewError(fiber.StatusBadRequest, "description must be at most 1000 characters")
	}
	return nil
}

func (h *handler) listArtifacts(c fiber.Ctx) error {
	artifacts, err := h.repos.Artifacts.List(c.Context())
	if err != nil {
		return repoError(err, "artifact")
	}
	return c.JSON(artifacts)
}

func (h *handler) artifactParam(c fiber.Ctx) (repository.Artifact, error) {
	a, err := h.repos.Artifacts.Get(c.Context(), c.Params("groupId"), c.Params("artifactId"))
	if err != nil {
		return a, repoError(err, "artifact")
	}
	return a, nil
}

func (h *handler) getArtifact(c fiber.Ctx) error {
	a, err := h.artifactParam(c)
	if err != nil {
		return err
	}
	return c.JSON(a)
}

func (h *handler) createArtifact(c fiber.Ctx) error {
	var req struct {
		GroupID     string  `json:"groupId"`
		ArtifactID  string  `json:"artifactId"`
		Public      bool    `json:"public"`
		Description *string `json:"description"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if !groupIDPattern.MatchString(req.GroupID) {
		return fiber.NewError(fiber.StatusBadRequest, "groupId must be dot-separated parts of letters, digits, '_' or '-'")
	}
	if !artifactIDPattern.MatchString(req.ArtifactID) {
		return fiber.NewError(fiber.StatusBadRequest, "artifactId must consist of letters, digits, '.', '_' or '-' and must not start with '.'")
	}
	if err := validateDescription(req.Description); err != nil {
		return err
	}

	a, err := h.repos.Artifacts.Create(c.Context(), req.GroupID, req.ArtifactID, req.Public, req.Description)
	if err != nil {
		return repoError(err, "artifact")
	}
	return c.Status(fiber.StatusCreated).JSON(a)
}

func (h *handler) updateArtifact(c fiber.Ctx) error {
	var req struct {
		Public      *bool   `json:"public"`
		Description *string `json:"description"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := validateDescription(req.Description); err != nil {
		return err
	}

	a, err := h.artifactParam(c)
	if err != nil {
		return err
	}
	a, err = h.repos.Artifacts.Update(c.Context(), a.ID, req.Public, req.Description)
	if err != nil {
		return repoError(err, "artifact")
	}
	return c.JSON(a)
}

func (h *handler) deleteArtifact(c fiber.Ctx) error {
	a, err := h.artifactParam(c)
	if err != nil {
		return err
	}
	if err := h.repos.Artifacts.Delete(c.Context(), a.ID); err != nil {
		return repoError(err, "artifact")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

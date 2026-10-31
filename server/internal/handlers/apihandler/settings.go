package apihandler

import (
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
)

const (
	maxServerNameLength        = 64
	maxServerDescriptionLength = 20000
)

func (h *handler) getSettings(c fiber.Ctx) error {
	s, err := h.repos.Settings.Get(c.Context())
	if err != nil {
		return repoError(err, "settings")
	}
	return c.JSON(s)
}

func (h *handler) updateSettings(c fiber.Ctx) error {
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if utf8.RuneCountInString(name) > maxServerNameLength {
			return fiber.NewError(fiber.StatusBadRequest, "name must be at most 64 characters")
		}
		req.Name = &name
	}
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if utf8.RuneCountInString(description) > maxServerDescriptionLength {
			return fiber.NewError(fiber.StatusBadRequest, "description must be at most 20000 characters")
		}
		req.Description = &description
	}

	s, err := h.repos.Settings.Update(c.Context(), req.Name, req.Description)
	if err != nil {
		return repoError(err, "settings")
	}
	return c.JSON(s)
}

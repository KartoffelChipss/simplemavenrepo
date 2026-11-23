package apihandler

import (
	"net/url"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/repository"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
)

const (
	maxServerNameLength        = 64
	maxServerDescriptionLength = 20000
	maxOSVURLLength            = 2048
)

func (h *handler) getSettings(c fiber.Ctx) error {
	s, err := h.repos.Settings.Get(c.Context())
	if err != nil {
		return repoError(err, "settings")
	}
	// The URL may point to an internal mirror, only admins need to see it.
	if u, ok, err := auth.IdentifyOptional(c, h.repos); err != nil || !ok || !u.IsAdmin {
		s.OSVURL = nil
	}
	return c.JSON(s)
}

func (h *handler) updateSettings(c fiber.Ctx) error {
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		OSVEnabled  *bool   `json:"osvEnabled"`
		OSVURL      *string `json:"osvUrl"`
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
	if req.OSVURL != nil {
		osvURL := strings.TrimRight(strings.TrimSpace(*req.OSVURL), "/")
		if osvURL != "" {
			if err := validateOSVURL(osvURL); err != nil {
				return err
			}
		}
		req.OSVURL = &osvURL
	}

	s, err := h.repos.Settings.Update(c.Context(), repository.SettingsUpdate{
		Name:        req.Name,
		Description: req.Description,
		OSVEnabled:  req.OSVEnabled,
		OSVURL:      req.OSVURL,
	})
	if err != nil {
		return repoError(err, "settings")
	}
	h.osv.Reset()
	return c.JSON(s)
}

func validateOSVURL(raw string) error {
	invalid := fiber.NewError(fiber.StatusBadRequest, "OSV URL must be an http or https URL without credentials, query or fragment")
	if len(raw) > maxOSVURLLength {
		return invalid
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return invalid
	}
	return nil
}

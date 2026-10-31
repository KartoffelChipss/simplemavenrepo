package apihandler

import (
	"errors"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/repository"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

var loginLimiter = limiter.New(limiter.Config{
	Max:                    10,
	Expiration:             5 * time.Minute,
	SkipSuccessfulRequests: true,
	KeyGenerator: func(c fiber.Ctx) string {
		var req struct {
			Username string `json:"username"`
		}
		_ = c.Bind().JSON(&req)
		return c.IP() + "|" + strings.ToLower(req.Username)
	},
	LimitReached: func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusTooManyRequests, "too many failed logins, try again later")
	},
})

func setSessionCookie(c fiber.Ctx, value string, expires time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     auth.SessionCookie,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HTTPOnly: true,
		Secure:   c.Secure(),
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}

func (h *handler) login(c fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}

	u, err := h.repos.Users.CheckPassword(c.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCredentials) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid username or password")
		}
		return repoError(err, "user")
	}

	session, expiresAt, err := h.repos.Sessions.Create(c.Context(), u.ID)
	if err != nil {
		return repoError(err, "session")
	}
	setSessionCookie(c, session, expiresAt)
	return h.userDetails(c, u)
}

func (h *handler) logout(c fiber.Ctx) error {
	if session := c.Cookies(auth.SessionCookie); session != "" {
		if err := h.repos.Sessions.Delete(c.Context(), session); err != nil {
			return repoError(err, "session")
		}
	}
	setSessionCookie(c, "", time.Unix(0, 0))
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *handler) changeMyPassword(c fiber.Ctx) error {
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}

	u := auth.CurrentUser(c)
	if u.HasPassword {
		if _, err := h.repos.Users.CheckPassword(c.Context(), u.Username, req.CurrentPassword); err != nil {
			if errors.Is(err, repository.ErrInvalidCredentials) {
				return fiber.NewError(fiber.StatusForbidden, "current password is wrong")
			}
			return repoError(err, "user")
		}
	}
	if err := h.setPassword(c, u, req.NewPassword); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// setPassword changes the password and logs the user out everywhere except in the current session.
func (h *handler) setPassword(c fiber.Ctx, u repository.User, password string) error {
	if err := h.repos.Users.SetPassword(c.Context(), u.ID, password); err != nil {
		if errors.Is(err, repository.ErrInvalidPassword) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return repoError(err, "user")
	}
	var current string
	if u.ID == auth.CurrentUser(c).ID {
		current = c.Cookies(auth.SessionCookie)
	}
	if err := h.repos.Sessions.DeleteForUser(c.Context(), u.ID, current); err != nil {
		return repoError(err, "session")
	}
	return nil
}

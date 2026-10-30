package apihandler

import (
	"encoding/base64"
	"errors"
	"log"
	"simplemavenrepo/internal/repository"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type userKey struct{}

// authenticate accepts a token either as "Bearer <token>" or as the password of HTTP Basic auth, which is what Maven and Gradle send.
func (h *handler) authenticate(c fiber.Ctx) error {
	username, token, ok := parseAuthorization(c.Get(fiber.HeaderAuthorization))
	if !ok {
		c.Set(fiber.HeaderWWWAuthenticate, `Basic realm="simplemavenrepo"`)
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
	}

	user, err := h.repos.Tokens.Authenticate(c.Context(), token)
	if err == nil && username != "" && username != user.Username {
		err = repository.ErrInvalidCredentials
	}
	if err != nil {
		if !errors.Is(err, repository.ErrInvalidCredentials) {
			log.Printf("Authentication failed: %v", err)
			return fiber.ErrInternalServerError
		}
		c.Set(fiber.HeaderWWWAuthenticate, `Basic realm="simplemavenrepo"`)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	fiber.Locals(c, userKey{}, user)
	return c.Next()
}

func requireAdmin(c fiber.Ctx) error {
	if !currentUser(c).IsAdmin {
		return fiber.NewError(fiber.StatusForbidden, "admin access required")
	}
	return c.Next()
}

func currentUser(c fiber.Ctx) repository.User {
	return fiber.Locals[repository.User](c, userKey{})
}

// parseAuthorization returns the username (empty for bearer tokens) and token from an Authorization header.
func parseAuthorization(header string) (username, token string, ok bool) {
	scheme, value, found := strings.Cut(header, " ")
	if !found {
		return "", "", false
	}
	switch strings.ToLower(scheme) {
	case "bearer":
		token = strings.TrimSpace(value)
		return "", token, token != ""
	case "basic":
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
		if err != nil {
			return "", "", false
		}
		username, token, found = strings.Cut(string(decoded), ":")
		return username, token, found && username != "" && token != ""
	}
	return "", "", false
}

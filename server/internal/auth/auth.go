package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"simplemavenrepo/internal/repository"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type userKey struct{}

// Authenticate is a middleware that only lets authenticated requests through. See Identify.
func Authenticate(tokens *repository.TokenRepository) fiber.Handler {
	return func(c fiber.Ctx) error {
		if err := Identify(c, tokens); err != nil {
			return err
		}
		return c.Next()
	}
}

// Identify authenticates the request and stores the user for CurrentUser. It accepts a token either as
// "Bearer <token>" or as the password of HTTP Basic auth, which is what Maven and Gradle send.
func Identify(c fiber.Ctx, tokens *repository.TokenRepository) error {
	username, token, ok := parseAuthorization(c.Get(fiber.HeaderAuthorization))
	if !ok {
		c.Set(fiber.HeaderWWWAuthenticate, `Basic realm="simplemavenrepo"`)
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
	}

	user, err := tokens.Authenticate(c.Context(), token)
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
	return nil
}

func RequireAdmin(c fiber.Ctx) error {
	if !CurrentUser(c).IsAdmin {
		return fiber.NewError(fiber.StatusForbidden, "admin access required")
	}
	return c.Next()
}

// CurrentUser returns the user set by Authenticate.
func CurrentUser(c fiber.Ctx) repository.User {
	return fiber.Locals[repository.User](c, userKey{})
}

// Authorize reports whether the user has perm on the repository path p. Admins may access everything.
func Authorize(ctx context.Context, grants *repository.GrantRepository, u repository.User, p string, perm repository.Permission) (bool, error) {
	if u.IsAdmin {
		return true, nil
	}
	effective, err := grants.ListEffective(ctx, u.ID)
	if err != nil {
		return false, err
	}
	return Allowed(effective, p, perm), nil
}

// Allowed reports whether any of the grants gives perm on the repository path p. A write grant implies read.
func Allowed(grants []repository.Grant, p string, perm repository.Permission) bool {
	for _, g := range grants {
		if (g.Permission == perm || g.Permission == repository.PermissionWrite) && repository.MatchPattern(g.Pattern, p) {
			return true
		}
	}
	return false
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

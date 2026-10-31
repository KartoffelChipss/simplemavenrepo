package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net/url"
	"simplemavenrepo/internal/repository"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type userKey struct{}

const SessionCookie = "smr_session"

// Authenticate is a middleware that only lets authenticated requests through. See Identify.
func Authenticate(repos repository.Repositories) fiber.Handler {
	return func(c fiber.Ctx) error {
		if err := Identify(c, repos); err != nil {
			return err
		}
		return c.Next()
	}
}

// Identify authenticates the request and stores the user for CurrentUser. It accepts a token either as
// "Bearer <token>" or as the password of HTTP Basic auth, which is what Maven and Gradle send. Without an
// Authorization header it falls back to the session cookie of the web UI.
func Identify(c fiber.Ctx, repos repository.Repositories) error {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		if session := c.Cookies(SessionCookie); session != "" {
			return identifySession(c, repos.Sessions, session)
		}
	}

	username, token, ok := parseAuthorization(header)
	if !ok {
		c.Set(fiber.HeaderWWWAuthenticate, `Basic realm="simplemavenrepo"`)
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
	}

	user, err := repos.Tokens.Authenticate(c.Context(), token)
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

// IdentifyOptional is Identify for endpoints that also serve anonymous requests. ok is false for requests without
// credentials and for an expired session cookie, which the web UI treats as logged out. Invalid tokens still fail.
func IdentifyOptional(c fiber.Ctx, repos repository.Repositories) (user repository.User, ok bool, err error) {
	if c.Get(fiber.HeaderAuthorization) == "" {
		session := c.Cookies(SessionCookie)
		if session == "" {
			return repository.User{}, false, nil
		}
		if _, err := repos.Sessions.Authenticate(c.Context(), session); errors.Is(err, repository.ErrInvalidCredentials) {
			return repository.User{}, false, nil
		}
	}
	if err := Identify(c, repos); err != nil {
		return repository.User{}, false, err
	}
	return CurrentUser(c), true, nil
}

func identifySession(c fiber.Ctx, sessions *repository.SessionRepository, session string) error {
	if !isSameOrigin(c) {
		return fiber.NewError(fiber.StatusForbidden, "cross-origin request")
	}

	user, err := sessions.Authenticate(c.Context(), session)
	if err != nil {
		if !errors.Is(err, repository.ErrInvalidCredentials) {
			log.Printf("Session authentication failed: %v", err)
			return fiber.ErrInternalServerError
		}
		return fiber.NewError(fiber.StatusUnauthorized, "session expired")
	}

	fiber.Locals(c, userKey{}, user)
	return nil
}

func isSameOrigin(c fiber.Ctx) bool {
	switch c.Method() {
	case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
		return true
	}
	if site := c.Get("Sec-Fetch-Site"); site != "" {
		return site == "same-origin" || site == "none"
	}
	if origin := c.Get(fiber.HeaderOrigin); origin != "" {
		u, err := url.Parse(origin)
		return err == nil && u.Host == c.Host()
	}
	return true
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

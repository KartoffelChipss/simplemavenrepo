package auth

import (
	"encoding/base64"
	"errors"
	"log"
	"net/url"
	"simplemavenrepo/internal/repository"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type userKey struct{}
type scopesKey struct{}

const SessionCookie = "smr_session"

// Authenticate is a middleware that only lets authenticated requests through. See Identify. Scoped tokens are
// rejected: they only grant access to repository paths, so they must not reach routes that manage the account, e.g.
// creating an unrestricted token. Use AuthenticateScoped for routes that check access with Authorize.
func Authenticate(repos repository.Repositories) fiber.Handler {
	return func(c fiber.Ctx) error {
		if err := Identify(c, repos); err != nil {
			return err
		}
		if Scopes(c) != nil {
			return fiber.NewError(fiber.StatusForbidden, "this token is restricted to repository access")
		}
		return c.Next()
	}
}

// AuthenticateScoped is Authenticate that also accepts scoped tokens. The route must check access with Authorize.
func AuthenticateScoped(repos repository.Repositories) fiber.Handler {
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

	user, scopes, err := repos.Tokens.Authenticate(c.Context(), token)
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
	if len(scopes) > 0 {
		fiber.Locals(c, scopesKey{}, scopes)
	}
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
	if !IsAdmin(c) {
		return fiber.NewError(fiber.StatusForbidden, "admin access required")
	}
	return c.Next()
}

// CurrentUser returns the user set by Authenticate.
func CurrentUser(c fiber.Ctx) repository.User {
	return fiber.Locals[repository.User](c, userKey{})
}

// Scopes returns the scopes of the token that authenticated the request, or nil when access is not restricted by
// the token: for unscoped tokens, sessions and anonymous requests.
func Scopes(c fiber.Ctx) []repository.TokenScope {
	return fiber.Locals[[]repository.TokenScope](c, scopesKey{})
}

// IsAdmin reports whether the request has admin access. Scoped tokens of admins don't.
func IsAdmin(c fiber.Ctx) bool {
	return CurrentUser(c).IsAdmin && Scopes(c) == nil
}

// Authorize reports whether the authenticated request has perm on the repository path p. See Permissions.
func Authorize(c fiber.Ctx, grants *repository.GrantRepository, p string, perm repository.Permission) (bool, error) {
	allowed, err := Permissions(c, grants)
	if err != nil {
		return false, err
	}
	return allowed(p, perm), nil
}

// Permissions returns a function that reports whether the authenticated request has a permission on a repository
// path, loading the grants only once. Admins may access everything. A scoped token is additionally limited to its
// scopes, also for admins.
func Permissions(c fiber.Ctx, grants *repository.GrantRepository) (func(p string, perm repository.Permission) bool, error) {
	u := CurrentUser(c)
	scopes := Scopes(c)
	var effective []repository.Grant
	if !u.IsAdmin {
		var err error
		if effective, err = grants.ListEffective(c.Context(), u.ID); err != nil {
			return nil, err
		}
	}
	return func(p string, perm repository.Permission) bool {
		if scopes != nil && !slices.ContainsFunc(scopes, func(s repository.TokenScope) bool { return s.Allows(p, perm) }) {
			return false
		}
		return u.IsAdmin || Allowed(effective, p, perm)
	}, nil
}

// Allowed reports whether any of the grants gives perm on the repository path p. A write grant implies read.
func Allowed(grants []repository.Grant, p string, perm repository.Permission) bool {
	for _, g := range grants {
		if g.Permission.Includes(perm) && repository.MatchPattern(g.Pattern, p) {
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

package apihandler

import (
	"errors"
	"log"
	"regexp"
	"simplemavenrepo/internal/repository"
	"slices"

	"github.com/gofiber/fiber/v3"
)

var repoNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// reservedRepoNames are the top-level paths of the server and the web UI, which a repository would hide.
var reservedRepoNames = []string{
	"api", "assets", "javadoc", "livez", "readyz",
	"login", "browse", "docs", "tokens", "artifacts", "repositories", "users", "teams", "stats", "settings",
}

func validateRepoName(name string) error {
	if !repoNamePattern.MatchString(name) {
		return fiber.NewError(fiber.StatusBadRequest,
			"name must be 1-64 lowercase letters, digits, '_' or '-' and start with a letter or digit")
	}
	if slices.Contains(reservedRepoNames, name) {
		return fiber.NewError(fiber.StatusBadRequest, `"`+name+`" is reserved for the server`)
	}
	return nil
}

func validatePolicy(p repository.VersionPolicy) error {
	if !p.Valid() {
		return fiber.NewError(fiber.StatusBadRequest, `versionPolicy must be "release", "snapshot" or "mixed"`)
	}
	return nil
}

// listRepos lists all repositories. Anyone may see them, so clients can show where to fetch artifacts from.
func (h *handler) listRepos(c fiber.Ctx) error {
	repos, err := h.repos.Repos.List(c.Context())
	if err != nil {
		return repoError(err, "repository")
	}
	return c.JSON(repos)
}

func (h *handler) repoParam(c fiber.Ctx) (repository.Repo, error) {
	r, err := h.repos.Repos.Get(c.Context(), c.Params("repo"))
	if err != nil {
		return r, repoError(err, "repository")
	}
	return r, nil
}

func (h *handler) getRepo(c fiber.Ctx) error {
	r, err := h.repoParam(c)
	if err != nil {
		return err
	}
	return c.JSON(r)
}

func (h *handler) createRepo(c fiber.Ctx) error {
	var req struct {
		Name          string                   `json:"name"`
		VersionPolicy repository.VersionPolicy `json:"versionPolicy"`
		Description   *string                  `json:"description"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := validateRepoName(req.Name); err != nil {
		return err
	}
	if err := validatePolicy(req.VersionPolicy); err != nil {
		return err
	}
	if err := validateDescription(req.Description); err != nil {
		return err
	}

	r, err := h.repos.Repos.Create(c.Context(), req.Name, req.VersionPolicy, req.Description)
	if err != nil {
		return repoError(err, "repository")
	}
	if err := h.root.MkdirAll(r.Name, 0o755); err != nil {
		log.Printf("Creating the directory of repository %s failed: %v", r.Name, err)
	}
	return c.Status(fiber.StatusCreated).JSON(r)
}

func (h *handler) updateRepo(c fiber.Ctx) error {
	var req struct {
		VersionPolicy *repository.VersionPolicy `json:"versionPolicy"`
		Description   *string                   `json:"description"`
	}
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if req.VersionPolicy != nil {
		if err := validatePolicy(*req.VersionPolicy); err != nil {
			return err
		}
	}
	if err := validateDescription(req.Description); err != nil {
		return err
	}

	r, err := h.repoParam(c)
	if err != nil {
		return err
	}
	r, err = h.repos.Repos.Update(c.Context(), r.ID, req.VersionPolicy, req.Description)
	if err != nil {
		return repoError(err, "repository")
	}
	return c.JSON(r)
}

// deleteRepo removes a repository without artifacts. With ?deleteFiles=true its directory is removed as well, which
// holds files of artifacts that were unregistered without deleting them.
func (h *handler) deleteRepo(c fiber.Ctx) error {
	r, err := h.repoParam(c)
	if err != nil {
		return err
	}
	if err := h.repos.Repos.Delete(c.Context(), r.ID); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return fiber.NewError(fiber.StatusConflict, "the repository still has artifacts, delete them first")
		}
		return repoError(err, "repository")
	}
	if fiber.Query[bool](c, "deleteFiles") {
		if err := h.root.RemoveAll(r.Name); err != nil {
			log.Printf("Deleting files of repository %s failed: %v", r.Name, err)
			return fiber.NewError(fiber.StatusInternalServerError)
		}
	}
	return c.SendStatus(fiber.StatusNoContent)
}

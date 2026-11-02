package apihandler

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"path"
	"regexp"
	"simplemavenrepo/internal/repository"
	"simplemavenrepo/internal/storage"
	"strings"

	"github.com/gofiber/fiber/v3"
)

var (
	groupIDPattern    = regexp.MustCompile(`^[A-Za-z0-9_-]+(\.[A-Za-z0-9_-]+)*$`)
	artifactIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-][A-Za-z0-9_.-]*$`)
)

const maxDescriptionLength = 1000

const (
	defaultDeployPageSize = 10
	maxDeployPageSize     = 100
)

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

// artifactWithSize is an artifact with the disk space its files take, not counting artifacts nested inside it.
type artifactWithSize struct {
	repository.Artifact
	Size int64 `json:"size"`
}

func (h *handler) withSize(ctx context.Context, a repository.Artifact) (artifactWithSize, error) {
	all, err := h.repos.Artifacts.List(ctx)
	if err != nil {
		return artifactWithSize{}, repoError(err, "artifact")
	}
	paths := make([]string, len(all))
	for i, other := range all {
		paths[i] = other.Path
	}
	u, err := storage.Measure(h.repoFS, a.Path, paths)
	if err != nil {
		log.Printf("Measuring the size of %s:%s failed: %v", a.GroupID, a.ArtifactID, err)
		return artifactWithSize{}, fiber.NewError(fiber.StatusInternalServerError)
	}
	return artifactWithSize{Artifact: a, Size: u.ByArtifact[a.Path]}, nil
}

func (h *handler) getArtifact(c fiber.Ctx) error {
	a, err := h.artifactParam(c)
	if err != nil {
		return err
	}
	res, err := h.withSize(c.Context(), a)
	if err != nil {
		return err
	}
	return c.JSON(res)
}

func (h *handler) createArtifact(c fiber.Ctx) error {
	var req struct {
		GroupID               string  `json:"groupId"`
		ArtifactID            string  `json:"artifactId"`
		Public                bool    `json:"public"`
		AllowReleaseOverwrite bool    `json:"allowReleaseOverwrite"`
		Description           *string `json:"description"`
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

	a, err := h.repos.Artifacts.Create(c.Context(), req.GroupID, req.ArtifactID, req.Public, req.AllowReleaseOverwrite, req.Description)
	if err != nil {
		return repoError(err, "artifact")
	}
	return c.Status(fiber.StatusCreated).JSON(a)
}

func (h *handler) updateArtifact(c fiber.Ctx) error {
	var req struct {
		Public                *bool   `json:"public"`
		AllowReleaseOverwrite *bool   `json:"allowReleaseOverwrite"`
		Description           *string `json:"description"`
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
	a, err = h.repos.Artifacts.Update(c.Context(), a.ID, req.Public, req.AllowReleaseOverwrite, req.Description)
	if err != nil {
		return repoError(err, "artifact")
	}
	res, err := h.withSize(c.Context(), a)
	if err != nil {
		return err
	}
	return c.JSON(res)
}

func (h *handler) deleteArtifact(c fiber.Ctx) error {
	a, err := h.artifactParam(c)
	if err != nil {
		return err
	}
	if fiber.Query[bool](c, "deleteFiles") {
		if err := h.deleteArtifactFiles(c.Context(), a); err != nil {
			log.Printf("Deleting files of %s:%s failed: %v", a.GroupID, a.ArtifactID, err)
			return fiber.NewError(fiber.StatusInternalServerError)
		}
	}
	if err := h.repos.Artifacts.Delete(c.Context(), a.ID); err != nil {
		return repoError(err, "artifact")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// deleteArtifactFiles removes the artifact's directory, except for the directories of other artifacts nested inside it
func (h *handler) deleteArtifactFiles(ctx context.Context, a repository.Artifact) error {
	all, err := h.repos.Artifacts.List(ctx)
	if err != nil {
		return err
	}
	var nested []string
	for _, other := range all {
		if strings.HasPrefix(other.Path, a.Path+"/") {
			nested = append(nested, other.Path)
		}
	}
	if err := removeTreeExcept(h.root, a.Path, nested); err != nil {
		return err
	}
	removeEmptyParents(h.root, a.Path)
	return nil
}

// removeTreeExcept deletes dir and everything below it except the keep directories and the directories leading to
// them. A missing dir is not an error.
func removeTreeExcept(root *os.Root, dir string, keep []string) error {
	if len(keep) == 0 {
		return root.RemoveAll(dir)
	}
	entries, err := fs.ReadDir(root.FS(), dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		child := path.Join(dir, e.Name())
		var below []string
		kept := false
		for _, k := range keep {
			if k == child {
				kept = true
			} else if strings.HasPrefix(k, child+"/") {
				below = append(below, k)
			}
		}
		if kept {
			continue
		}
		if len(below) > 0 && e.IsDir() {
			if err := removeTreeExcept(root, child, below); err != nil {
				return err
			}
			continue
		}
		if err := root.RemoveAll(child); err != nil {
			return err
		}
	}
	return nil
}

// removeEmptyParents removes dir and its parents up to the repository root for as long as they are empty, so deleting
// com/example/demo doesn't leave an empty com/example behind.
func removeEmptyParents(root *os.Root, dir string) {
	for ; dir != "." && dir != "/" && dir != ""; dir = path.Dir(dir) {
		if err := root.Remove(dir); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return
		}
	}
}

func (h *handler) listDeploys(c fiber.Ctx) error {
	a, err := h.artifactParam(c)
	if err != nil {
		return err
	}
	page := fiber.Query[int](c, "page", 1)
	pageSize := fiber.Query[int](c, "pageSize", defaultDeployPageSize)
	if page < 1 || pageSize < 1 || pageSize > maxDeployPageSize {
		return fiber.NewError(fiber.StatusBadRequest, "invalid page or pageSize")
	}

	deploys, total, err := h.repos.Deploys.List(c.Context(), a.Path, (page-1)*pageSize, pageSize)
	if err != nil {
		return repoError(err, "deploy")
	}
	return c.JSON(fiber.Map{"deploys": deploys, "total": total, "page": page, "pageSize": pageSize})
}

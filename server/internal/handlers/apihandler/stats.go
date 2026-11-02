package apihandler

import (
	"log"
	"simplemavenrepo/internal/storage"
	"sort"

	"github.com/gofiber/fiber/v3"
)

const (
	defaultLargestArtifacts = 10
	maxLargestArtifacts     = 100
)

// getStorageStats reports the disk space used by the repository and the artifacts taking up the most of it.
func (h *handler) getStorageStats(c fiber.Ctx) error {
	limit := fiber.Query[int](c, "limit", defaultLargestArtifacts)
	if limit < 1 || limit > maxLargestArtifacts {
		return fiber.NewError(fiber.StatusBadRequest, "invalid limit")
	}

	all, err := h.repos.Artifacts.List(c.Context())
	if err != nil {
		return repoError(err, "artifact")
	}
	paths := make([]string, len(all))
	for i, a := range all {
		paths[i] = a.Path
	}
	u, err := storage.Measure(h.repoFS, ".", paths)
	if err != nil {
		log.Printf("Measuring the repository size failed: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	largest := make([]artifactWithSize, len(all))
	for i, a := range all {
		largest[i] = artifactWithSize{Artifact: a, Size: u.ByArtifact[a.Path]}
	}
	sort.SliceStable(largest, func(i, j int) bool { return largest[i].Size > largest[j].Size })
	largest = largest[:min(limit, len(largest))]

	return c.JSON(fiber.Map{
		"totalSize":        u.Total,
		"fileCount":        u.Files,
		"artifactCount":    len(all),
		"unregisteredSize": u.Unregistered,
		"largestArtifacts": largest,
	})
}

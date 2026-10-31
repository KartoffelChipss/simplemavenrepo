package healthhandler

import (
	"context"
	"database/sql"
	"log"
	"simplemavenrepo/internal/config"
	"simplemavenrepo/internal/util"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

const probeTimeout = 2 * time.Second

func RegisterHealthRoutes(app *fiber.App, database *sql.DB) {
	methods := []string{fiber.MethodGet, fiber.MethodHead}
	app.Add(methods, healthcheck.LivenessEndpoint, healthcheck.New())
	app.Add(methods, healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
		Probe: func(c fiber.Ctx) bool {
			return ready(c.Context(), database)
		},
	}))
}

func ready(ctx context.Context, database *sql.DB) bool {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		log.Printf("Readiness check: database unavailable: %v", err)
		return false
	}
	if err := util.CheckDirAccess(config.GetRepoPath()); err != nil {
		log.Printf("Readiness check: repository path unavailable: %v", err)
		return false
	}
	return true
}

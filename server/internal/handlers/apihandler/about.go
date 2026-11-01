package apihandler

import (
	"runtime"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/config"
	"time"

	"github.com/gofiber/fiber/v3"
)

var startedAt = time.Now()

func (h *handler) about(c fiber.Ctx) error {
	res := fiber.Map{"status": "ok"}
	user, loggedIn, err := auth.IdentifyOptional(c, h.repos)
	if err != nil {
		return err
	}
	if loggedIn && user.IsAdmin {
		version := config.GetVersion()
		if version == "" {
			version = "dev"
		}
		res["version"] = version
		res["goVersion"] = runtime.Version()
		res["platform"] = runtime.GOOS + "/" + runtime.GOARCH
		res["startedAt"] = startedAt
	}
	return c.JSON(res)
}

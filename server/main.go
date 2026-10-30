package main

import (
	"log"
	"net"
	"simplemavenrepo/internal/config"
	"simplemavenrepo/internal/handlers/apihandler"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	app := fiber.New()

	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} (${latency}) - ${ip}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	apihandler.RegisterApiRoutes(app)

	addr := net.JoinHostPort(config.GetHost(), config.GetPort())
	log.Println("Server starting on " + addr)
	log.Fatal(app.Listen(addr, fiber.ListenConfig{DisableStartupMessage: true}))
}

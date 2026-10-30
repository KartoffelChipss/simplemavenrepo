package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"simplemavenrepo/internal/config"
	"simplemavenrepo/internal/handlers/apihandler"
	"simplemavenrepo/internal/handlers/ressourcehandler"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func checkDirAccess(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	f, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", dir, err)
	}
	_, err = f.ReadDir(1)
	f.Close()
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("cannot read %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".write-test-*")
	if err != nil {
		return fmt.Errorf("cannot write to %s: %w", dir, err)
	}
	name := tmp.Name()
	tmp.Close()
	return os.Remove(name)
}

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit:         config.GetBodyLimit(),
		StreamRequestBody: true,
	})

	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} (${latency}) - ${ip}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	repoPath := config.GetRepoPath()
	err := checkDirAccess(repoPath)
	if err != nil {
		log.Fatalf("Repository path error: %v", err)
	}
	log.Printf("Using repository path: %s", repoPath)

	apihandler.RegisterApiRoutes(app)
	ressourcehandler.RegisterRessourceRoutes(app)

	addr := net.JoinHostPort(config.GetHost(), config.GetPort())
	log.Println("Server starting on " + addr)
	log.Fatal(app.Listen(addr, fiber.ListenConfig{DisableStartupMessage: true}))
}

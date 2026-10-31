package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"simplemavenrepo/internal/config"
	"simplemavenrepo/internal/db"
	"simplemavenrepo/internal/handlers/apihandler"
	"simplemavenrepo/internal/handlers/ressourcehandler"
	"simplemavenrepo/internal/repository"
	"simplemavenrepo/internal/webui"

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

// bootstrapAdmin creates an initial admin user when the database has no users, so the management API can be used.
func bootstrapAdmin(repos repository.Repositories) error {
	ctx := context.Background()
	n, err := repos.Users.Count(ctx)
	if err != nil || n > 0 {
		return err
	}

	admin, err := repos.Users.Create(ctx, "admin", true)
	if err != nil {
		return err
	}
	pw := make([]byte, 18)
	rand.Read(pw)
	password := base64.RawURLEncoding.EncodeToString(pw)
	if err := repos.Users.SetPassword(ctx, admin.ID, password); err != nil {
		return err
	}
	_, secret, err := repos.Tokens.Create(ctx, admin.ID, "bootstrap", nil)
	if err != nil {
		return err
	}
	log.Printf("Created initial admin user %q with password %s and token %s - store them now, they will not be shown again",
		admin.Username, password, secret)
	return nil
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

	configPath := config.GetConfigPath()
	err = os.MkdirAll(configPath, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}
	log.Printf("Using config path: %s", configPath)

	database, err := db.Open(config.GetDBPath())
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer database.Close()
	log.Printf("Using database: %s", config.GetDBPath())

	repos := repository.New(database)
	if err := bootstrapAdmin(repos); err != nil {
		log.Fatalf("Creating initial admin failed: %v", err)
	}

	apihandler.RegisterApiRoutes(app, repos)
	ressourcehandler.RegisterRessourceRoutes(app, repos)
	webui.RegisterRoutes(app)

	addr := net.JoinHostPort(config.GetHost(), config.GetPort())
	log.Println("Server starting on " + addr)
	log.Fatal(app.Listen(addr, fiber.ListenConfig{DisableStartupMessage: true}))
}

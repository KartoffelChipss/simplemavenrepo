package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net"
	"os"
	"path/filepath"
	"simplemavenrepo/internal/config"
	"simplemavenrepo/internal/db"
	"simplemavenrepo/internal/handlers/apihandler"
	"simplemavenrepo/internal/handlers/healthhandler"
	"simplemavenrepo/internal/handlers/javadochandler"
	"simplemavenrepo/internal/handlers/ressourcehandler"
	"simplemavenrepo/internal/javadoc"
	"simplemavenrepo/internal/repository"
	"simplemavenrepo/internal/util"
	"simplemavenrepo/internal/webui"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

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

	repoPath := config.GetRepoPath()
	err := util.CheckDirAccess(repoPath)
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

	root, err := os.OpenRoot(repoPath)
	if err != nil {
		log.Fatalf("Cannot open repository path: %v", err)
	}
	key, err := javadoc.LoadOrCreateKey(filepath.Join(configPath, "javadoc.key"))
	if err != nil {
		log.Fatalf("Javadoc signing key error: %v", err)
	}
	signer := javadoc.NewSigner(key)

	// Registered before the logger so probes every few seconds don't flood the request log
	healthhandler.RegisterHealthRoutes(app, database)

	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} (${latency}) - ${ip}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	apihandler.RegisterApiRoutes(app, repos, root.FS(), signer)
	javadochandler.RegisterJavadocRoutes(app, root.FS(), signer)
	ressourcehandler.RegisterRessourceRoutes(app, repos)
	webui.RegisterRoutes(app)

	addr := net.JoinHostPort(config.GetHost(), config.GetPort())
	log.Println("Server starting on " + addr)
	log.Fatal(app.Listen(addr, fiber.ListenConfig{DisableStartupMessage: true}))
}

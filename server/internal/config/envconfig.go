package config

import (
	"log"
	"os"
	"path/filepath"
	"simplemavenrepo/internal/util"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetVersion() string {
	return getEnv("SMR_VERSION", "")
}

func GetPort() string {
	return getEnv("SMR_PORT", "4321")
}

func GetHost() string {
	return getEnv("SMR_HOST", "127.0.0.1")
}

func GetRepoPath() string {
	return getEnv("SMR_REPO_PATH", "./repo")
}

func GetDBPath() string {
	return filepath.Join(GetConfigPath(), "smr.db")
}

func GetConfigPath() string {
	return getEnv("SMR_CONFIG_PATH", "./config")
}

func GetBodyLimit() int {
	limit, err := util.ParseByteSize(getEnv("SMR_BODY_LIMIT", "512MB"))
	if err != nil {
		log.Printf("Error parsing body limit. Falling back to default 512MB. Error: %v", err)
		return 512 * 1024 * 1024 // 512MB
	}
	return limit
}

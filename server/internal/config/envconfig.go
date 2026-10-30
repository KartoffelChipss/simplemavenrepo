package config

import (
	"os"
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

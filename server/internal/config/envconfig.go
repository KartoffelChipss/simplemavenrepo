package config

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"simplemavenrepo/internal/util"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
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

func GetShutdownTimeout() time.Duration {
	timeout, err := time.ParseDuration(getEnv("SMR_SHUTDOWN_TIMEOUT", "10s"))
	if err != nil || timeout <= 0 {
		log.Printf("Error parsing shutdown timeout. Falling back to default 10s. Error: %v", err)
		return 10 * time.Second
	}
	return timeout
}

func GetTrustProxyConfig() (cfg fiber.TrustProxyConfig, ok bool) {
	for _, entry := range strings.Split(getEnv("SMR_TRUSTED_PROXIES", ""), ",") {
		entry = strings.TrimSpace(entry)
		switch strings.ToLower(entry) {
		case "":
			continue
		case "loopback":
			cfg.Loopback = true
		case "private":
			cfg.Private = true
		case "linklocal":
			cfg.LinkLocal = true
		default:
			if _, _, err := net.ParseCIDR(entry); err != nil && net.ParseIP(entry) == nil {
				log.Printf("Ignoring invalid trusted proxy %q, expected an IP, a CIDR range, loopback, private or linklocal", entry)
				continue
			}
			cfg.Proxies = append(cfg.Proxies, entry)
		}
		ok = true
	}
	return cfg, ok
}

func GetProxyHeader() string {
	return getEnv("SMR_PROXY_HEADER", fiber.HeaderXForwardedFor)
}

func GetBodyLimit() int {
	limit, err := util.ParseByteSize(getEnv("SMR_BODY_LIMIT", "512MB"))
	if err != nil {
		log.Printf("Error parsing body limit. Falling back to default 512MB. Error: %v", err)
		return 512 * 1024 * 1024 // 512MB
	}
	return limit
}

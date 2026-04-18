package config

import (
	"fmt"
	"os"
	"strings"
)

// LegacyConfig holds all application configuration loaded from environment variables.
// Deprecated: use the new nested Config struct instead. Removed in task 1.2.
type LegacyConfig struct {
	DBDriver    string
	DBPath      string
	DatabaseURL string
	JWTSecret   string
	CORSOrigins []string
	Port        string
	GinMode     string
}

// LoadConfig reads environment variables and returns a validated LegacyConfig.
// Deprecated: use the new config loading in task 1.2.
func LoadConfig() (*LegacyConfig, error) {
	cfg := &LegacyConfig{}

	// DBDriver with default
	cfg.DBDriver = os.Getenv("DB_DRIVER")
	if cfg.DBDriver == "" {
		cfg.DBDriver = "sqlite"
	}

	// DBPath with default
	cfg.DBPath = os.Getenv("DB_PATH")
	if cfg.DBPath == "" {
		cfg.DBPath = "./data/dev.db"
	}

	// DatabaseURL (required for mysql, optional otherwise)
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")

	// JWTSecret (required, min 32 bytes)
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 bytes, got %d", len(cfg.JWTSecret))
	}

	// CORSOrigins parsed from comma-separated env var
	if raw := os.Getenv("CORS_ORIGINS"); raw != "" {
		for _, origin := range strings.Split(raw, ",") {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				cfg.CORSOrigins = append(cfg.CORSOrigins, trimmed)
			}
		}
	}

	// Port with default
	cfg.Port = os.Getenv("PORT")
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// GinMode (release mode)
	cfg.GinMode = os.Getenv("GIN_MODE")

	return cfg, nil
}

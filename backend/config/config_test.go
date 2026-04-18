package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to set multiple env vars and return a cleanup function
func setEnvs(vars map[string]string) func() {
	for k, v := range vars {
		os.Setenv(k, v)
	}
	return func() {
		for k := range vars {
			os.Unsetenv(k)
		}
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	cleanup := setEnvs(map[string]string{
		"JWT_SECRET": "this-is-exactly-thirty-two-bytes-long!!",
	})
	defer cleanup()

	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "sqlite", cfg.DBDriver)
	assert.Equal(t, "./data/dev.db", cfg.DBPath)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "this-is-exactly-thirty-two-bytes-long!!", cfg.JWTSecret)
	assert.Nil(t, cfg.CORSOrigins)
}

func TestLoadConfig_CustomValues(t *testing.T) {
	cleanup := setEnvs(map[string]string{
		"DB_DRIVER":    "mysql",
		"DB_PATH":      "/data/prod.db",
		"DATABASE_URL": "user:pass@tcp(localhost:3306)/db",
		"JWT_SECRET":   "a-very-long-secret-that-is-at-least-32-bytes!!",
		"CORS_ORIGINS": "http://localhost:3000,https://example.com",
		"PORT":         "9090",
	})
	defer cleanup()

	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "mysql", cfg.DBDriver)
	assert.Equal(t, "/data/prod.db", cfg.DBPath)
	assert.Equal(t, "user:pass@tcp(localhost:3306)/db", cfg.DatabaseURL)
	assert.Equal(t, "a-very-long-secret-that-is-at-least-32-bytes!!", cfg.JWTSecret)
	assert.Equal(t, []string{"http://localhost:3000", "https://example.com"}, cfg.CORSOrigins)
	assert.Equal(t, "9090", cfg.Port)
}

func TestLoadConfig_JWTSecretTooShort(t *testing.T) {
	cleanup := setEnvs(map[string]string{
		"JWT_SECRET": "too-short",
	})
	defer cleanup()

	cfg, err := LoadConfig()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
	assert.Contains(t, err.Error(), "32")
}

func TestLoadConfig_JWTSecretMissing(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	cfg, err := LoadConfig()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestLoadConfig_JWTSecretExactly32Bytes(t *testing.T) {
	secret := "12345678901234567890123456789012" // exactly 32 bytes
	cleanup := setEnvs(map[string]string{
		"JWT_SECRET": secret,
	})
	defer cleanup()

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, secret, cfg.JWTSecret)
}

func TestLoadConfig_CORSOrigins_Empty(t *testing.T) {
	cleanup := setEnvs(map[string]string{
		"JWT_SECRET":  "a-very-long-secret-that-is-at-least-32-bytes!!",
		"CORS_ORIGINS": "",
	})
	defer cleanup()

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Nil(t, cfg.CORSOrigins)
}

func TestLoadConfig_CORSOrigins_WithSpaces(t *testing.T) {
	cleanup := setEnvs(map[string]string{
		"JWT_SECRET":  "a-very-long-secret-that-is-at-least-32-bytes!!",
		"CORS_ORIGINS": " http://a.com , https://b.com , ",
	})
	defer cleanup()

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, []string{"http://a.com", "https://b.com"}, cfg.CORSOrigins)
}

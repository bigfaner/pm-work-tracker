package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to write a temp YAML file and return its path + cleanup
func writeTempYAML(t *testing.T, content string) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path, func() {}
}

func TestLoadConfig_DefaultsWhenNoFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.yaml")
	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Server defaults
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "", cfg.Server.GinMode)
	assert.Equal(t, Duration(30*time.Second), cfg.Server.ReadTimeout)
	assert.Equal(t, Duration(30*time.Second), cfg.Server.WriteTimeout)
	assert.Equal(t, int64(10*1024*1024), cfg.Server.MaxBodySize)

	// Database defaults
	assert.Equal(t, "sqlite", cfg.Database.Driver)
	assert.Equal(t, "./data/dev.db", cfg.Database.Path)
	assert.Equal(t, 10, cfg.Database.MaxOpenConns)
	assert.Equal(t, 5, cfg.Database.MaxIdleConns)
	assert.Equal(t, Duration(time.Hour), cfg.Database.ConnMaxLifetime)

	// Auth defaults
	assert.Equal(t, "", cfg.Auth.JWTSecret)
	assert.Equal(t, Duration(24*time.Hour), cfg.Auth.JWTExpiry)

	// Logging defaults
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestLoadConfig_YAMLParsingSuccess(t *testing.T) {
	yaml := `
server:
  port: "9090"
  gin_mode: release
  read_timeout: 60s
  write_timeout: 45s
  max_body_size: 20971520
database:
  driver: mysql
  path: /data/prod.db
  url: "user:pass@tcp(localhost:3306)/db"
  max_open_conns: 20
  max_idle_conns: 10
  conn_max_lifetime: 2h
auth:
  jwt_secret: "this-is-exactly-thirty-two-bytes-long!!"
  jwt_expiry: 48h
  initial_admin:
    username: admin
    password: secret123
cors:
  origins:
    - http://localhost:3000
    - https://example.com
logging:
  level: debug
  format: text
`
	path, cleanup := writeTempYAML(t, yaml)
	defer cleanup()

	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Server overrides
	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, "release", cfg.Server.GinMode)
	assert.Equal(t, Duration(60*time.Second), cfg.Server.ReadTimeout)
	assert.Equal(t, Duration(45*time.Second), cfg.Server.WriteTimeout)
	assert.Equal(t, int64(20971520), cfg.Server.MaxBodySize)

	// Database overrides
	assert.Equal(t, "mysql", cfg.Database.Driver)
	assert.Equal(t, "/data/prod.db", cfg.Database.Path)
	assert.Equal(t, "user:pass@tcp(localhost:3306)/db", cfg.Database.URL)
	assert.Equal(t, 20, cfg.Database.MaxOpenConns)
	assert.Equal(t, 10, cfg.Database.MaxIdleConns)
	assert.Equal(t, Duration(2*time.Hour), cfg.Database.ConnMaxLifetime)

	// Auth overrides
	assert.Equal(t, "this-is-exactly-thirty-two-bytes-long!!", cfg.Auth.JWTSecret)
	assert.Equal(t, Duration(48*time.Hour), cfg.Auth.JWTExpiry)
	assert.Equal(t, "admin", cfg.Auth.InitialAdmin.Username)
	assert.Equal(t, "secret123", cfg.Auth.InitialAdmin.Password)

	// CORS overrides
	assert.Equal(t, []string{"http://localhost:3000", "https://example.com"}, cfg.CORS.Origins)

	// Logging overrides
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
}

func TestLoadConfig_YAMLSyntaxError(t *testing.T) {
	yaml := `
server:
  port: "9090"
  this is not valid yaml: [
`
	path, cleanup := writeTempYAML(t, yaml)
	defer cleanup()

	cfg, err := LoadConfig(path)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse config file")
}

func TestLoadConfig_SpecificFieldOverride(t *testing.T) {
	// Only override one field, rest should stay default
	yaml := `
server:
  port: "3000"
`
	path, cleanup := writeTempYAML(t, yaml)
	defer cleanup()

	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Overridden
	assert.Equal(t, "3000", cfg.Server.Port)
	// Defaults preserved
	assert.Equal(t, Duration(30*time.Second), cfg.Server.ReadTimeout)
	assert.Equal(t, "sqlite", cfg.Database.Driver)
	assert.Equal(t, "./data/dev.db", cfg.Database.Path)
	assert.Equal(t, "info", cfg.Logging.Level)
}

func TestLoadConfig_NestedStructParsing(t *testing.T) {
	yaml := `
auth:
  jwt_secret: "this-is-exactly-thirty-two-bytes-long!!"
  jwt_expiry: 12h
  initial_admin:
    username: superadmin
    password: mypassword
`
	path, cleanup := writeTempYAML(t, yaml)
	defer cleanup()

	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "this-is-exactly-thirty-two-bytes-long!!", cfg.Auth.JWTSecret)
	assert.Equal(t, Duration(12*time.Hour), cfg.Auth.JWTExpiry)
	assert.Equal(t, "superadmin", cfg.Auth.InitialAdmin.Username)
	assert.Equal(t, "mypassword", cfg.Auth.InitialAdmin.Password)
}

// --- applyEnvOverrides tests ---

func TestApplyEnvOverrides_StringField(t *testing.T) {
	cfg := defaultConfig()
	t.Setenv("SERVER_PORT", "3000")
	t.Setenv("LOGGING_LEVEL", "debug")

	err := applyEnvOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, "3000", cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestApplyEnvOverrides_IntField(t *testing.T) {
	cfg := defaultConfig()
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "25")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "12")

	err := applyEnvOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
	assert.Equal(t, 12, cfg.Database.MaxIdleConns)
}

func TestApplyEnvOverrides_Int64Field(t *testing.T) {
	cfg := defaultConfig()
	t.Setenv("SERVER_MAX_BODY_SIZE", "20971520")

	err := applyEnvOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, int64(20971520), cfg.Server.MaxBodySize)
}

func TestApplyEnvOverrides_DurationField(t *testing.T) {
	cfg := defaultConfig()
	t.Setenv("SERVER_READ_TIMEOUT", "60s")
	t.Setenv("AUTH_JWT_EXPIRY", "48h")

	err := applyEnvOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, Duration(60*time.Second), cfg.Server.ReadTimeout)
	assert.Equal(t, Duration(48*time.Hour), cfg.Auth.JWTExpiry)
}

func TestApplyEnvOverrides_SliceField(t *testing.T) {
	cfg := defaultConfig()
	t.Setenv("CORS_ORIGINS", "http://localhost:3000, https://example.com")

	err := applyEnvOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, []string{"http://localhost:3000", "https://example.com"}, cfg.CORS.Origins)
}

func TestApplyEnvOverrides_NestedStructField(t *testing.T) {
	cfg := defaultConfig()
	t.Setenv("AUTH_INITIAL_ADMIN_USERNAME", "superadmin")
	t.Setenv("AUTH_INITIAL_ADMIN_PASSWORD", "secret123")

	err := applyEnvOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, "superadmin", cfg.Auth.InitialAdmin.Username)
	assert.Equal(t, "secret123", cfg.Auth.InitialAdmin.Password)
}

func TestApplyEnvOverrides_EmptyEnvVar(t *testing.T) {
	cfg := defaultConfig()
	cfg.Server.Port = "9090"
	// Set env var to empty — should NOT override
	t.Setenv("SERVER_PORT", "")

	err := applyEnvOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, "9090", cfg.Server.Port, "empty env var should not override existing value")
}

func TestApplyEnvOverrides_UnsetEnvVar(t *testing.T) {
	cfg := defaultConfig()
	cfg.Server.Port = "9090"
	// Ensure env var is not set at all
	os.Unsetenv("SERVER_PORT")

	err := applyEnvOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, "9090", cfg.Server.Port, "unset env var should not override existing value")
}

func TestApplyEnvOverrides_InvalidInt(t *testing.T) {
	cfg := defaultConfig()
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "not-a-number")

	err := applyEnvOverrides(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_MAX_OPEN_CONNS")
	assert.Contains(t, err.Error(), "cannot parse")
}

func TestApplyEnvOverrides_InvalidInt64(t *testing.T) {
	cfg := defaultConfig()
	t.Setenv("SERVER_MAX_BODY_SIZE", "not-a-number")

	err := applyEnvOverrides(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SERVER_MAX_BODY_SIZE")
	assert.Contains(t, err.Error(), "cannot parse")
}

func TestApplyEnvOverrides_InvalidDuration(t *testing.T) {
	cfg := defaultConfig()
	t.Setenv("SERVER_READ_TIMEOUT", "not-a-duration")

	err := applyEnvOverrides(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SERVER_READ_TIMEOUT")
	assert.Contains(t, err.Error(), "cannot parse")
}

func TestLoadConfig_IntegratedWithEnvOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.yaml")
	t.Setenv("SERVER_PORT", "7070")
	t.Setenv("LOGGING_LEVEL", "warn")

	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "7070", cfg.Server.Port)
	assert.Equal(t, "warn", cfg.Logging.Level)
	// Defaults for non-overridden fields should still be present
	assert.Equal(t, "sqlite", cfg.Database.Driver)
}

func TestLoadConfig_DurationFieldFromString(t *testing.T) {
	yaml := `
server:
  read_timeout: 1h30m
  write_timeout: 500ms
database:
  conn_max_lifetime: 30m
auth:
  jwt_expiry: 72h
`
	path, cleanup := writeTempYAML(t, yaml)
	defer cleanup()

	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, Duration(90*time.Minute), cfg.Server.ReadTimeout)
	assert.Equal(t, Duration(500*time.Millisecond), cfg.Server.WriteTimeout)
	assert.Equal(t, Duration(30*time.Minute), cfg.Database.ConnMaxLifetime)
	assert.Equal(t, Duration(72*time.Hour), cfg.Auth.JWTExpiry)
}

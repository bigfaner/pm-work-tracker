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

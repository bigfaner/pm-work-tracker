package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	// Server defaults
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "", cfg.Server.GinMode)
	assert.Equal(t, Duration(30*time.Second), cfg.Server.ReadTimeout)
	assert.Equal(t, Duration(30*time.Second), cfg.Server.WriteTimeout)
	assert.Equal(t, int64(10*1024*1024), cfg.Server.MaxBodySize)

	// Database defaults
	assert.Equal(t, "sqlite", cfg.Database.Driver)
	assert.Equal(t, "./data/dev.db", cfg.Database.Path)
	assert.Equal(t, "", cfg.Database.URL)
	assert.Equal(t, 10, cfg.Database.MaxOpenConns)
	assert.Equal(t, 5, cfg.Database.MaxIdleConns)
	assert.Equal(t, Duration(time.Hour), cfg.Database.ConnMaxLifetime)

	// Auth defaults
	assert.Equal(t, "", cfg.Auth.JWTSecret)
	assert.Equal(t, Duration(24*time.Hour), cfg.Auth.JWTExpiry)
	assert.Equal(t, "", cfg.Auth.InitialAdmin.Username)
	assert.Equal(t, "", cfg.Auth.InitialAdmin.Password)

	// CORS defaults
	assert.Nil(t, cfg.CORS.Origins)

	// Logging defaults
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestDuration_UnmarshalYAML_String(t *testing.T) {
	var d Duration
	err := yaml.Unmarshal([]byte(`"30s"`), &d)
	require.NoError(t, err)
	assert.Equal(t, Duration(30*time.Second), d)
}

func TestDuration_UnmarshalYAML_HourString(t *testing.T) {
	var d Duration
	err := yaml.Unmarshal([]byte(`"1h"`), &d)
	require.NoError(t, err)
	assert.Equal(t, Duration(time.Hour), d)
}

func TestDuration_UnmarshalYAML_ComplexString(t *testing.T) {
	var d Duration
	err := yaml.Unmarshal([]byte(`"1h30m"`), &d)
	require.NoError(t, err)
	assert.Equal(t, Duration(90*time.Minute), d)
}

func TestDuration_UnmarshalYAML_Integer(t *testing.T) {
	var d Duration
	// 30 seconds in nanoseconds
	err := yaml.Unmarshal([]byte(`30000000000`), &d)
	require.NoError(t, err)
	assert.Equal(t, Duration(30*time.Second), d)
}

func TestDuration_UnmarshalYAML_InvalidString(t *testing.T) {
	var d Duration
	err := yaml.Unmarshal([]byte(`"not-a-duration"`), &d)
	assert.Error(t, err)
}

func TestDuration_UnmarshalYAML_UnquotedString(t *testing.T) {
	// YAML: unquoted 30s should parse as a string
	var d Duration
	err := yaml.Unmarshal([]byte(`30s`), &d)
	require.NoError(t, err)
	assert.Equal(t, Duration(30*time.Second), d)
}

func TestDuration_UnmarshalYAML_UnquotedHour(t *testing.T) {
	var d Duration
	err := yaml.Unmarshal([]byte(`1h`), &d)
	require.NoError(t, err)
	assert.Equal(t, Duration(time.Hour), d)
}

func TestDuration_Value(t *testing.T) {
	d := Duration(5 * time.Minute)
	assert.Equal(t, 5*time.Minute, d.Value())
}

func TestDuration_Value_Zero(t *testing.T) {
	var d Duration
	assert.Equal(t, time.Duration(0), d.Value())
}

func TestConfig_YAMLTags(t *testing.T) {
	// Verify the struct has correct yaml tags by marshaling and checking keys
	cfg := defaultConfig()
	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	// The marshaled YAML should contain our nested keys
	s := string(data)
	assert.Contains(t, s, "server:")
	assert.Contains(t, s, "database:")
	assert.Contains(t, s, "auth:")
	assert.Contains(t, s, "cors:")
	assert.Contains(t, s, "logging:")
}

func TestConfig_YAMLRoundTrip(t *testing.T) {
	original := defaultConfig()
	data, err := yaml.Marshal(original)
	require.NoError(t, err)

	var parsed Config
	err = yaml.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, original.Server.Port, parsed.Server.Port)
	assert.Equal(t, original.Server.ReadTimeout, parsed.Server.ReadTimeout)
	assert.Equal(t, original.Server.WriteTimeout, parsed.Server.WriteTimeout)
	assert.Equal(t, original.Server.MaxBodySize, parsed.Server.MaxBodySize)
	assert.Equal(t, original.Database.Driver, parsed.Database.Driver)
	assert.Equal(t, original.Database.Path, parsed.Database.Path)
	assert.Equal(t, original.Database.ConnMaxLifetime, parsed.Database.ConnMaxLifetime)
	assert.Equal(t, original.Auth.JWTExpiry, parsed.Auth.JWTExpiry)
	assert.Equal(t, original.Logging.Level, parsed.Logging.Level)
	assert.Equal(t, original.Logging.Format, parsed.Logging.Format)
}

func TestConfig_YAMLParseCustom(t *testing.T) {
	input := `
server:
  port: "9090"
  read_timeout: 60s
  write_timeout: 60s
database:
  driver: mysql
  url: "user:pass@tcp(localhost:3306)/db"
auth:
  jwt_secret: "this-is-exactly-thirty-two-bytes-long!!"
  jwt_expiry: 48h
logging:
  level: debug
  format: text
`
	var cfg Config
	err := yaml.Unmarshal([]byte(input), &cfg)
	require.NoError(t, err)

	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, Duration(60*time.Second), cfg.Server.ReadTimeout)
	assert.Equal(t, "mysql", cfg.Database.Driver)
	assert.Equal(t, "user:pass@tcp(localhost:3306)/db", cfg.Database.URL)
	assert.Equal(t, "this-is-exactly-thirty-two-bytes-long!!", cfg.Auth.JWTSecret)
	assert.Equal(t, Duration(48*time.Hour), cfg.Auth.JWTExpiry)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
}

func TestConfig_YAMLParseWithInitialAdmin(t *testing.T) {
	input := `
auth:
  jwt_secret: "this-is-exactly-thirty-two-bytes-long!!"
  initial_admin:
    username: admin
    password: secret123
`
	var cfg Config
	err := yaml.Unmarshal([]byte(input), &cfg)
	require.NoError(t, err)

	assert.Equal(t, "admin", cfg.Auth.InitialAdmin.Username)
	assert.Equal(t, "secret123", cfg.Auth.InitialAdmin.Password)
}

func TestConfig_YAMLParseCORSOrigins(t *testing.T) {
	input := `
cors:
  origins:
    - http://localhost:3000
    - https://example.com
`
	var cfg Config
	err := yaml.Unmarshal([]byte(input), &cfg)
	require.NoError(t, err)

	assert.Equal(t, []string{"http://localhost:3000", "https://example.com"}, cfg.CORS.Origins)
}

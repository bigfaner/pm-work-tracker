package config

import (
	"time"
)

// Duration is a custom type wrapping time.Duration for YAML unmarshaling.
// It supports string format ("30s", "1h") with integer (nanoseconds) fallback.
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler for Duration.
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		parsed, err := time.ParseDuration(s)
		if err == nil {
			*d = Duration(parsed)
			return nil
		}
		// String parsed but not a valid duration; fall through to int.
	}
	// fallback: integer (nanoseconds)
	var ns int64
	if err := unmarshal(&ns); err != nil {
		return err
	}
	*d = Duration(time.Duration(ns))
	return nil
}

// Value converts Duration back to time.Duration for consumers.
func (d Duration) Value() time.Duration {
	return time.Duration(d)
}

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	CORS     CORSConfig     `yaml:"cors"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Port         string   `yaml:"port" env:"SERVER_PORT"`
	GinMode      string   `yaml:"gin_mode" env:"SERVER_GIN_MODE"`
	BasePath     string   `yaml:"base_path" env:"SERVER_BASE_PATH"`
	ReadTimeout  Duration `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT"`
	WriteTimeout Duration `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT"`
	MaxBodySize  int64    `yaml:"max_body_size" env:"SERVER_MAX_BODY_SIZE"`
}

// DatabaseConfig holds database-related configuration.
type DatabaseConfig struct {
	Driver          string   `yaml:"driver" env:"DATABASE_DRIVER"`
	Path            string   `yaml:"path" env:"DATABASE_PATH"`
	URL             string   `yaml:"url" env:"DATABASE_URL"`
	MaxOpenConns    int      `yaml:"max_open_conns" env:"DATABASE_MAX_OPEN_CONNS"`
	MaxIdleConns    int      `yaml:"max_idle_conns" env:"DATABASE_MAX_IDLE_CONNS"`
	ConnMaxLifetime Duration `yaml:"conn_max_lifetime" env:"DATABASE_CONN_MAX_LIFETIME"`
}

// AuthConfig holds authentication-related configuration.
type AuthConfig struct {
	JWTSecret    string             `yaml:"jwt_secret" env:"AUTH_JWT_SECRET"`
	JWTExpiry    Duration           `yaml:"jwt_expiry" env:"AUTH_JWT_EXPIRY"`
	InitialAdmin InitialAdminConfig `yaml:"initial_admin"`
}

// InitialAdminConfig holds initial admin user configuration.
type InitialAdminConfig struct {
	Username string `yaml:"username" env:"AUTH_INITIAL_ADMIN_USERNAME"`
	Password string `yaml:"password" env:"AUTH_INITIAL_ADMIN_PASSWORD"`
}

// CORSConfig holds CORS-related configuration.
type CORSConfig struct {
	Origins []string `yaml:"origins" env:"CORS_ORIGINS"`
}

// LoggingConfig holds logging-related configuration.
type LoggingConfig struct {
	Level  string `yaml:"level" env:"LOGGING_LEVEL"`
	Format string `yaml:"format" env:"LOGGING_FORMAT"`
}

// defaultConfig returns a Config with sensible defaults.
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         "8080",
			ReadTimeout:  Duration(30 * time.Second),
			WriteTimeout: Duration(30 * time.Second),
			MaxBodySize:  10 * 1024 * 1024,
		},
		Database: DatabaseConfig{
			Driver:          "sqlite",
			Path:            "./data/dev.db",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: Duration(time.Hour),
		},
		Auth: AuthConfig{
			JWTExpiry: Duration(24 * time.Hour),
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

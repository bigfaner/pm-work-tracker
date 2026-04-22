package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// validate checks all config fields and returns the first error encountered.
func (c *Config) validate() error {
	// auth.jwt_secret >= 32 bytes
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("auth.jwt_secret must be at least 32 bytes, got %d", len(c.Auth.JWTSecret))
	}
	if strings.HasPrefix(c.Auth.JWTSecret, "CHANGE_ME") {
		return fmt.Errorf("auth.jwt_secret is still a placeholder — replace it before starting")
	}
	// server.port 1024-65535
	port, _ := strconv.Atoi(c.Server.Port)
	if port < 1024 || port > 65535 {
		return fmt.Errorf("server.port must be between 1024 and 65535, got %s", c.Server.Port)
	}
	// database.driver
	if c.Database.Driver != "sqlite" && c.Database.Driver != "mysql" {
		return fmt.Errorf("database.driver must be sqlite or mysql, got %q", c.Database.Driver)
	}
	// logging.level
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error")
	}
	// logging.format
	if c.Logging.Format != "json" && c.Logging.Format != "text" {
		return fmt.Errorf("logging.format must be json or text")
	}
	// positive durations
	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server.read_timeout must be positive")
	}
	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server.write_timeout must be positive")
	}
	if c.Database.ConnMaxLifetime <= 0 {
		return fmt.Errorf("database.conn_max_lifetime must be positive")
	}
	if c.Auth.JWTExpiry <= 0 {
		return fmt.Errorf("auth.jwt_expiry must be positive")
	}
	return nil
}

// LoadConfig reads a YAML config file from the given path and returns a Config.
// If the file does not exist, defaults are used. If the file exists but has
// invalid YAML syntax, an error is returned.
func LoadConfig(path string) (*Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		// File doesn't exist; use defaults, but still apply env overrides.
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config file: %w", err)
		}
	}

	if err := applyEnvOverrides(cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation: %s", err)
	}

	return cfg, nil
}

// applyEnvOverrides recursively walks struct fields and overrides values from
// environment variables based on the `env` struct tag.
func applyEnvOverrides(cfg *Config) error {
	return walkStruct(reflect.ValueOf(cfg).Elem())
}

// walkStruct recursively processes struct fields. For leaf fields with an `env`
// tag, it reads the environment variable and parses the value according to the
// field's type. For nested structs, it recurses into their fields.
func walkStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		envTag := fieldType.Tag.Get("env")

		// For struct fields, recurse into them to process their env-tagged fields
		if field.Kind() == reflect.Struct {
			if err := walkStruct(field); err != nil {
				return err
			}
			continue
		}

		// Only process fields with an env tag
		if envTag == "" {
			continue
		}

		// Check if env var is set and non-empty using LookupEnv
		value, ok := os.LookupEnv(envTag)
		if !ok || value == "" {
			continue
		}

		if err := setField(field, envTag, value); err != nil {
			return err
		}
	}

	return nil
}

// setField parses the environment variable value and sets the reflect.Value.
func setField(field reflect.Value, envKey, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		return nil
	case reflect.Int:
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("env %s: cannot parse %q as int: %w", envKey, value, err)
		}
		field.SetInt(int64(n))
		return nil
	case reflect.Int64:
		// Check for Duration type
		if field.Type() == reflect.TypeOf(Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("env %s: cannot parse %q as duration: %w", envKey, value, err)
			}
			field.SetInt(int64(d))
			return nil
		}
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("env %s: cannot parse %q as int64: %w", envKey, value, err)
		}
		field.SetInt(n)
		return nil
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			field.Set(reflect.ValueOf(parts))
			return nil
		}
	}

	return fmt.Errorf("env %s: unsupported type %s", envKey, field.Kind())
}

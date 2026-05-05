---
scope: backend
source: feature/config-yaml TECH-001–005
---

# Config Struct Convention

Pattern for structured configuration in Go backend services.

## TECH-config-001: Nested Config Struct

Configuration uses nested structs with `yaml` and `env` struct tags. Each domain is a separate sub-struct.

```go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Auth     AuthConfig     `yaml:"auth"`
    CORS     CORSConfig     `yaml:"cors"`
    Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
    Port string `yaml:"port" env:"SERVER_PORT"`
}
```

Adding a new config domain: create a new sub-struct, add as field to `Config`, add `yaml` tag.

## TECH-config-002: Duration YAML Unmarshaling

Custom `Duration` type supports human-readable strings ("30s", "1h") while falling back to integer nanoseconds.

```go
type Duration time.Duration

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
    var s string
    if err := unmarshal(&s); err == nil {
        parsed, err := time.ParseDuration(s)
        if err != nil { return err }
        *d = Duration(parsed)
        return nil
    }
    var ns int64
    if err := unmarshal(&ns); err != nil { return err }
    *d = Duration(time.Duration(ns))
    return nil
}
```

## TECH-config-003: Env Override via Reflection

`applyEnvOverrides` recursively traverses struct fields with `env` tags, parsing by type:

| Go type | Parsing |
|---------|---------|
| `string` | direct assignment |
| `int` | `strconv.Atoi` |
| `int64` | `strconv.ParseInt` |
| `time.Duration` | `time.ParseDuration` |
| `[]string` | comma split + trim spaces |

## TECH-config-004: Validation Rules

| Field | Constraint |
|-------|-----------|
| JWT secret | ≥ 32 bytes |
| Server port | 1024–65535 |
| Database driver | "sqlite" or "mysql" |
| Logging level | debug/info/warn/error |
| Logging format | json or text |
| Duration fields | > 0 |

## TECH-config-005: InitDB with Connection Pool

`InitDB(*DatabaseConfig)` configures connection pool from config:

```go
db.SetMaxOpenConns(cfg.MaxOpenConns)
db.SetMaxIdleConns(cfg.MaxIdleConns)
db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
```

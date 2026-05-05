---
feature: "config-yaml"
generated: "2026-05-04"
status: draft
---

# Technical Specifications: config-yaml

## Config Struct Design

### TECH-001: Nested Config Struct Pattern

**Requirement**: Configuration uses nested structs with `yaml` and `env` struct tags. Each domain (server, database, auth, cors, logging) is a separate sub-struct.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section "Data Models"

Pattern:
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
    // ...
}
```

### TECH-002: Duration YAML Unmarshaling

**Requirement**: Custom `Duration` type with `UnmarshalYAML` supports human-readable strings ("30s", "1h") while falling back to integer nanoseconds.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section "Duration YAML Unmarshaling"

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

### TECH-003: Env Override via Reflection

**Requirement**: `applyEnvOverrides` recursively traverses struct fields with `env` tags, parsing values by type: string, int, int64, time.Duration, []string (comma-separated).

**Scope**: [CROSS]

**Source**: design/tech-design.md Section "Env Override"

Type mapping:
- `string` → direct assignment
- `int` → `strconv.Atoi`
- `int64` → `strconv.ParseInt`
- `time.Duration` → `time.ParseDuration`
- `[]string` → comma split + trim spaces

## Validation

### TECH-004: Config Validation Rules

**Requirement**: Field-level validation constraints applied at startup.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section "Validation Rules"

| Field | Constraint | Error Message |
|-------|-----------|---------------|
| `auth.jwt_secret` | ≥ 32 bytes | "auth.jwt_secret must be at least 32 bytes, got N" |
| `server.port` | 1024–65535 | "server.port must be between 1024 and 65535" |
| `database.driver` | "sqlite" or "mysql" | "database.driver must be sqlite or mysql" |
| `logging.level` | debug/info/warn/error | "logging.level must be one of: debug, info, warn, error" |
| `logging.format` | json or text | "logging.format must be json or text" |
| duration fields | > 0 | "{field} must be positive" |

## Database

### TECH-005: InitDB with Connection Pool

**Requirement**: `InitDB(*DatabaseConfig)` configures connection pool parameters from config struct: MaxOpenConns, MaxIdleConns, ConnMaxLifetime.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section "InitDB"

---
created: 2026-04-18
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: config-yaml

## Overview

将后端配置从扁平环境变量迁移到结构化 YAML 配置文件，按功能域分组，支持环境变量覆盖。YAML 文件路径通过 `-config` flag 指定，默认 `config.yaml`（相对于工作目录）。

配置优先级：**环境变量 > config.yaml > 硬编码默认值**

## Architecture

### Layer Placement

配置属于基础设施层（`backend/config/`），所有上层（handler、service、repository）通过依赖注入接收配置。

### Component Diagram

```
cmd/server/main.go
  │  flag "-config" → path
  ↓
config.LoadConfig(path)
  ├── defaultConfig()          → 硬编码默认值
  ├── yaml.Unmarshal(file)     → 解析 YAML
  ├── applyEnvOverrides(cfg)   → 反射遍历 env tag，os.Getenv 覆盖
  └── cfg.validate()           → 校验所有字段
        │
        ↓
  ┌─────┼──────────┐
  ↓     ↓          ↓
InitDB()  SetupRouter()  SeedAdmin()
(dbCfg)   (server, cors,  (authCfg)
           auth)
```

### Dependencies

| 依赖 | 类型 | 说明 |
|------|------|------|
| `gopkg.in/yaml.v3` | 新增直接依赖 | 已有间接依赖，标准 YAML v3 解析 |
| `golang.org/x/crypto/bcrypt` | 已有 | SeedAdmin 密码哈希 |

## Interfaces

### LoadConfig

```go
func LoadConfig(path string) (*Config, error)
```

`path` 由 `main.go` 通过 flag 传入：
```go
configPath := flag.String("config", "config.yaml", "path to config file")
flag.Parse()
cfg, err := config.LoadConfig(*configPath)
```

### InitDB

```go
func InitDB(cfg *DatabaseConfig) (*gorm.DB, error)
```

签名变更：接受 `*DatabaseConfig` 替代无参数（原实现内部读 env vars）。

### SeedAdmin

```go
func SeedAdmin(db *gorm.DB, authCfg *AuthConfig) error
```

幂等操作：用户已存在则跳过。失败仅警告，不阻止启动。

## Data Models

### Config Struct

```go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Auth     AuthConfig     `yaml:"auth"`
    CORS     CORSConfig     `yaml:"cors"`
    Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
    Port         string        `yaml:"port" env:"SERVER_PORT"`
    GinMode      string        `yaml:"gin_mode" env:"SERVER_GIN_MODE"`
    ReadTimeout  time.Duration `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT"`
    WriteTimeout time.Duration `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT"`
    MaxBodySize  int64         `yaml:"max_body_size" env:"SERVER_MAX_BODY_SIZE"`
}

type DatabaseConfig struct {
    Driver          string        `yaml:"driver" env:"DATABASE_DRIVER"`
    Path            string        `yaml:"path" env:"DATABASE_PATH"`
    URL             string        `yaml:"url" env:"DATABASE_URL"`
    MaxOpenConns    int           `yaml:"max_open_conns" env:"DATABASE_MAX_OPEN_CONNS"`
    MaxIdleConns    int           `yaml:"max_idle_conns" env:"DATABASE_MAX_IDLE_CONNS"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DATABASE_CONN_MAX_LIFETIME"`
}

type AuthConfig struct {
    JWTSecret    string              `yaml:"jwt_secret" env:"AUTH_JWT_SECRET"`
    JWTExpiry    time.Duration       `yaml:"jwt_expiry" env:"AUTH_JWT_EXPIRY"`
    InitialAdmin InitialAdminConfig  `yaml:"initial_admin"`
}

type InitialAdminConfig struct {
    Username string `yaml:"username" env:"AUTH_INITIAL_ADMIN_USERNAME"`
    Password string `yaml:"password" env:"AUTH_INITIAL_ADMIN_PASSWORD"`
}

type CORSConfig struct {
    Origins []string `yaml:"origins" env:"CORS_ORIGINS"`
}

type LoggingConfig struct {
    Level  string `yaml:"level" env:"LOGGING_LEVEL"`
    Format string `yaml:"format" env:"LOGGING_FORMAT"`
}
```

### YAML Example (config.yaml)

```yaml
server:
  port: "8080"
  gin_mode: ""
  read_timeout: 30s
  write_timeout: 30s
  max_body_size: 10485760

database:
  driver: sqlite
  path: ./data/dev.db
  url: ""
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 1h

auth:
  jwt_secret: ""
  jwt_expiry: 24h
  initial_admin:
    username: admin
    password: admin123

cors:
  origins: []

logging:
  level: info
  format: json
```

### defaultConfig()

```go
func defaultConfig() *Config {
    return &Config{
        Server: ServerConfig{
            Port:         "8080",
            ReadTimeout:  30 * time.Second,
            WriteTimeout: 30 * time.Second,
            MaxBodySize:  10 * 1024 * 1024,
        },
        Database: DatabaseConfig{
            Driver:       "sqlite",
            Path:         "./data/dev.db",
            MaxOpenConns: 10,
            MaxIdleConns: 5,
            ConnMaxLifetime: time.Hour,
        },
        Auth: AuthConfig{
            JWTExpiry: 24 * time.Hour,
        },
        Logging: LoggingConfig{
            Level:  "info",
            Format: "json",
        },
    }
}
```

### Env Override (Reflection)

`applyEnvOverrides(cfg *Config) error` 递归遍历结构体：

1. 对每个有 `env` tag 的叶子字段，取 `os.Getenv(tagValue)`
2. 如果 env var 非空，按字段类型解析：
   - `string` → 直接赋值
   - `int` → `strconv.Atoi`
   - `int64` → `strconv.ParseInt`
   - `time.Duration` → `time.ParseDuration`
   - `[]string` → 逗号分割 + trim 空格
3. 对嵌套结构体（如 `Auth.InitialAdmin`），递归处理其子字段的 `env` tag

### Duration YAML Unmarshaling

`time.Duration` 默认以纳秒整数解析。自定义 `UnmarshalYAML` 支持 "30s"、"1h" 字符串：

```go
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
    var s string
    if err := unmarshal(&s); err == nil {
        parsed, err := time.ParseDuration(s)
        if err != nil {
            return err
        }
        *d = Duration(parsed)
        return nil
    }
    // fallback: integer (nanoseconds)
    var ns int64
    if err := unmarshal(&ns); err != nil {
        return err
    }
    *d = Duration(time.Duration(ns))
    return nil
}
```

使用自定义类型 `type Duration time.Duration` 以附加 `UnmarshalYAML` 方法，同时保持与 `time.Duration` 兼容。

## Error Handling

### Strategy

所有错误在启动阶段立即报告并退出（`log.Fatalf`）。不做错误累积。

| 场景 | 处理 |
|------|------|
| 指定的 config 文件不存在 | 静默跳过，使用默认值 |
| 指定的 config 文件存在但语法错误 | `fmt.Errorf("parse config file: %w", err)` → 退出 |
| 环境变量值无法解析 | `fmt.Errorf("env %s: cannot parse %q as %s: %w", ...)` → 退出 |
| 验证失败 | `fmt.Errorf("config validation: %s", desc)` → 退出 |
| SeedAdmin 失败 | `log.Printf("warning: seed admin: %v", err)` → 仅警告，不阻止启动 |

### Validation Rules

```go
func (c *Config) validate() error {
    // auth.jwt_secret >= 32 bytes
    if len(c.Auth.JWTSecret) < 32 {
        return fmt.Errorf("auth.jwt_secret must be at least 32 bytes, got %d", len(c.Auth.JWTSecret))
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
    if c.Server.ReadTimeout <= 0 { return fmt.Errorf("server.read_timeout must be positive") }
    if c.Server.WriteTimeout <= 0 { return fmt.Errorf("server.write_timeout must be positive") }
    if c.Database.ConnMaxLifetime <= 0 { return fmt.Errorf("database.conn_max_lifetime must be positive") }
    if c.Auth.JWTExpiry <= 0 { return fmt.Errorf("auth.jwt_expiry must be positive") }
    return nil
}
```

## Testing Strategy

### Unit Tests

| 测试文件 | 覆盖范围 |
|---------|---------|
| `config_test.go` | `LoadConfig`: 默认值、YAML 解析、env 覆盖、验证规则 |
| `config_test.go` | `applyEnvOverrides`: 各类型字段的解析和错误 |
| `config_test.go` | `validate`: 每条规则的 pass/fall |
| `database_test.go` | `InitDB`: 连接池参数正确设置 |
| `seed_test.go` (新) | `SeedAdmin`: 创建新用户、跳过已存在用户、空配置跳过 |

### Integration Tests

更新现有 3 个集成测试文件中 `config.Config` 构造方式：
- `tests/integration/auth_isolation_test.go`
- `tests/integration/progress_completion_test.go`
- `internal/handler/router_test.go`

### Coverage Target

新增代码 ≥ 90% 覆盖率。

## Security Considerations

### Threat Model

| 风险 | 影响 |
|------|------|
| config.yaml 含敏感值被提交到版本控制 | JWT secret、管理员密码泄露 |
| 配置文件权限过宽 | 其他用户可读取敏感配置 |

### Mitigations

- `config.yaml` 加入 `.gitignore`
- 仅提供 `config.yaml.example`（不含真实密钥）作为模板
- `SeedAdmin()` 使用 bcrypt 哈希密码，不明文存储
- `applyEnvOverrides` 不在日志中输出任何 env var 值

## Breaking Changes

| 变更 | 影响文件 | 迁移 |
|------|---------|------|
| `config.Config` → 嵌套结构体 | `main.go`, `router.go`, 3 个测试文件 | `cfg.JWTSecret` → `cfg.Auth.JWTSecret` |
| `InitDB()` → `InitDB(*DatabaseConfig)` | `main.go` | 传入 `&cfg.Database` |
| 环境变量名变更 | 部署配置 | `JWT_SECRET` → `AUTH_JWT_SECRET` 等 |

## File Changes Summary

| 文件 | 操作 | 说明 |
|------|------|------|
| `config/config.go` | 重写 | LoadConfig + Config struct + applyEnvOverrides + validate + Duration type |
| `config/database.go` | 修改 | InitDB 接受 DatabaseConfig，加入连接池配置 |
| `config/seed.go` | 新建 | SeedAdmin 函数 |
| `config/config_test.go` | 重写 | 适配新 struct + 新测试 |
| `config/database_test.go` | 修改 | InitDB 新签名测试 |
| `config/seed_test.go` | 新建 | SeedAdmin 测试 |
| `cmd/server/main.go` | 修改 | flag 解析 + 新 Config 用法 |
| `internal/handler/router.go` | 修改 | 嵌套字段访问 |
| `internal/handler/router_test.go` | 修改 | Config 构造适配 |
| `tests/integration/*.go` | 修改 | Config 构造适配 |
| `config.yaml.example` | 新建 | 配置模板 |
| `.gitignore` | 修改 | 添加 config.yaml |

## Open Questions

None — all decisions resolved through PRD and design review.

## Appendix

### Alternatives Considered

| 方案 | 优势 | 劣势 | 未选原因 |
|------|------|------|---------|
| `goccy/go-yaml` | 性能更好 | API 兼容性不如 yaml.v3 | 已有 yaml.v3 间接依赖，够用 |
| 手动 env 映射 | 无反射开销 | 每增一个字段需手动维护 | ~20 个字段不值得用手动映射 |
| Viper | 功能全面（多格式、远程配置） | 重型依赖、过度抽象 | 远超当前需求 |

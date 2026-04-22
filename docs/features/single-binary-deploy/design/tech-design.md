---
created: 2026-04-22
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: 前后端单体构建 (Single Binary Deploy)

## Overview

使用 Go 标准库 `embed` 将前端 `dist/` 打包进后端二进制。单个二进制同时提供 API 路由、静态资源和 SPA 兜底。`BASE_PATH` 从配置文件读取，控制所有路由的统一前缀，支持多服务共享域名部署。

关键决策：
- **API 路径去掉 `/api` 前缀**：`/api/v1/*` → `{BASE_PATH}/v1/*`，与前端路由用版本前缀区分
- **`embed.FS` 而非运行时读文件**：编译时锁定前后端版本，零外部依赖
- **`fstest.MapFS` 用于测试**：无需真实文件，单元测试可覆盖所有静态服务逻辑

## Architecture

### Layer Placement

本 feature 横跨三层：

| 层 | 变更 |
|----|------|
| Config | 新增 `BasePath` 字段 + 占位符校验 |
| Handler/Router | 路由前缀重构 + 静态文件 handler |
| Build | 新增构建脚本 + embed 包 |

### Component Diagram

```
scripts/build.sh
  ├── npm ci && npm run build  →  frontend/dist/
  ├── cp frontend/dist → backend/web/dist/
  └── go build ./cmd/server   →  bin/pm-work-tracker-{env}
                                        │
                              ┌─────────▼──────────────┐
                              │  Go Binary              │
                              │  ┌─────────────────┐   │
                              │  │ web.FS (embed)  │   │
                              │  │  dist/index.html│   │
                              │  │  dist/assets/*  │   │
                              │  └────────┬────────┘   │
                              │           │             │
                              │  ┌────────▼────────┐   │
                              │  │   Gin Router    │   │
                              │  │  /health        │   │
                              │  │  {BP}/v1/*      │   │
                              │  │  {BP}/assets/*  │   │
                              │  │  {BP}/**        │   │
                              │  └─────────────────┘   │
                              └────────────────────────┘

BP = BasePath (empty in local dev, e.g. "/pm" in production)
```

### New Files

| 文件 | 说明 |
|------|------|
| `backend/web/embed.go` | `package web`，导出 `var FS embed.FS` 和 `ValidateAssets()` |
| `backend/web/dist/` | 构建产物，gitignore |
| `backend/internal/handler/static.go` | `ServeStatic` + `ServeSPA` handler |
| `scripts/build.sh` | 构建脚本 |
| `backend/config.yaml.example` | 含占位符的配置模板，提交到 git |

### Dependencies

无新增外部依赖。使用标准库：
- `embed` — 嵌入静态文件
- `io/fs` — 操作 embed.FS
- `testing/fstest` — 测试用内存 FS（仅 test）
- `net/http` — MIME 类型检测

## Interfaces

### web package

```go
// backend/web/embed.go
package web

import "embed"

//go:embed dist
var FS embed.FS

// ValidateAssets checks that dist/index.html exists in the embedded FS.
// Returns an error with a hint to run the build script if missing.
func ValidateAssets(fs embed.FS) error
```

### Static Handler

```go
// backend/internal/handler/static.go

// ServeStatic serves files from dist/assets/* with long-lived cache headers.
// Returns 404 JSON {"error":"not found"} for missing files.
func ServeStatic(fs embed.FS) gin.HandlerFunc

// ServeSPA always returns dist/index.html with Cache-Control: no-cache.
func ServeSPA(fs embed.FS) gin.HandlerFunc
```

### Router

```go
// backend/internal/handler/router.go

// SetupRouter signature change: accepts fs embed.FS.
// When called from main, passes web.FS.
// When called from tests without static serving, passes empty embed.FS.
func SetupRouter(deps *Dependencies, fs embed.FS) *gin.Engine
```

## Data Models

### Config 变更

```go
// backend/config/types.go
type ServerConfig struct {
    Port         string   `yaml:"port"          env:"SERVER_PORT"`
    GinMode      string   `yaml:"gin_mode"       env:"SERVER_GIN_MODE"`
    BasePath     string   `yaml:"base_path"      env:"SERVER_BASE_PATH"`  // NEW
    ReadTimeout  Duration `yaml:"read_timeout"   env:"SERVER_READ_TIMEOUT"`
    WriteTimeout Duration `yaml:"write_timeout"  env:"SERVER_WRITE_TIMEOUT"`
    MaxBodySize  int64    `yaml:"max_body_size"  env:"SERVER_MAX_BODY_SIZE"`
}
```

### config.yaml.example

```yaml
server:
  port: "8080"
  gin_mode: ""
  base_path: ""              # e.g. "/pm" for shared-domain deployments
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
  jwt_secret: "CHANGE_ME_jwt_secret_min_32_bytes_long"
  jwt_expiry: 24h
  initial_admin:
    username: admin
    password: CHANGE_ME_admin_password

cors:
  origins: []

logging:
  level: info
  format: json
```

## Route Changes

### API 路由前缀重构

| Before | After | 说明 |
|--------|-------|------|
| `/api/v1/*` | `{BP}/v1/*` | BP 为空时本地路径不变（去掉 `/api`） |
| `/health` | `/health` | 不加 BP 前缀，供负载均衡探活 |

### 新增路由（注册顺序）

```
1. GET  /health                    → healthCheck (no BP, no auth)
2. POST {BP}/v1/auth/login         → rate-limited
3. POST {BP}/v1/auth/logout        → auth
4. GET  {BP}/v1/teams              → auth
5. POST {BP}/v1/teams              → auth + perm
6. ...  {BP}/v1/teams/:teamId/*    → auth + team scope + perm
7. ...  {BP}/v1/admin/*            → auth + perm
8. GET  {BP}/v1/me/permissions     → auth
9. GET  {BP}/assets/*filepath      → ServeStatic (no auth)
10. GET {BP}/                      → ServeSPA (no auth)
11. GET {BP}/*path                 → ServeSPA (no auth, catch-all)
```

静态路由注册在 API 路由之后，确保 API 优先匹配。

### 缓存策略

| 路径 | Cache-Control |
|------|---------------|
| `{BP}/assets/*` | `max-age=31536000, immutable` |
| `{BP}/` | `no-cache` |
| `{BP}/*path` (SPA fallback) | `no-cache` |

## Error Handling

### 静态资源 404

```json
HTTP 404
{"error": "not found"}
```

不走 `apperrors.RespondError`，直接 `c.JSON(404, gin.H{"error": "not found"})`，与 API 错误格式保持一致但不引入 AppError 依赖。

### 启动校验失败

```
# 构建产物缺失
FATAL: embedded assets missing dist/index.html — did the build script succeed?

# 配置占位符未替换
FATAL: config validation: auth.jwt_secret is still a placeholder
```

两种情况均 `log.Fatalf` 退出，退出码 1。

### 构建脚本错误

`set -e` 保证任意步骤失败立即退出，打印失败步骤信息，非零退出码。

## Testing Strategy

### Unit Tests

| 测试 | 文件 | 验证点 |
|------|------|--------|
| `TestServeStatic_Found` | `static_test.go` | 返回文件内容 + `max-age=31536000, immutable` |
| `TestServeStatic_NotFound` | `static_test.go` | 返回 404 + `{"error":"not found"}` |
| `TestServeSPA` | `static_test.go` | 返回 index.html + `no-cache` |
| `TestValidateAssets_Missing` | `embed_test.go` | index.html 缺失时返回 error |
| `TestValidateAssets_Present` | `embed_test.go` | index.html 存在时返回 nil |
| `TestConfigValidate_Placeholder` | `config_test.go` | `CHANGE_ME_*` jwt_secret 被拒绝 |
| `TestRouterBasePath` | `router_test.go` | 路由注册在正确前缀下 |

所有测试使用 `testing/fstest.MapFS` 构造内存 FS，无需真实构建产物。

### Integration Tests

现有 `backend/cmd/server/main_test.go` 继续覆盖完整启动流程。

### Coverage Target

新增代码覆盖率目标：≥ 80%

## Security Considerations

### 敏感配置不入 git

- `backend/config.yaml` 加入 `.gitignore`
- `config.yaml.example` 中敏感字段使用 `CHANGE_ME_` 前缀占位符
- 启动时校验占位符未被留存，防止误用模板文件直接启动

### 路径穿越防护

使用 `fs.Open` + `embed.FS` 访问文件，Go 标准库自动拒绝 `..` 路径穿越，无需额外处理。

### 静态文件 MIME 类型

使用 `http.DetectContentType` 或文件扩展名推断 MIME，防止 Content-Type 嗅探攻击（配合 `X-Content-Type-Options: nosniff`）。

## Frontend Changes

| 文件 | 变更 |
|------|------|
| `vite.config.ts` | 新增 `base: process.env.VITE_BASE_PATH \|\| '/'` |
| `src/main.tsx` | `BrowserRouter` 新增 `basename={import.meta.env.VITE_BASE_PATH \|\| '/'}` |
| `src/api/client.ts` | `baseURL` 改为 `` `${import.meta.env.VITE_BASE_PATH ?? ''}/v1` `` |
| `vite.config.ts` proxy | `/api` → `/v1` |

本地开发：`VITE_BASE_PATH` 未设置，basename 为 `/`，baseURL 为 `/v1`，proxy `/v1 → localhost:8080`。

## Build Script

```bash
# scripts/build.sh
# Usage: ./scripts/build.sh <dev|prod>
#
# Steps:
# 1. Validate ENV arg (dev|prod)
# 2. Check git branch (dev→dev, prod→main)
# 3. cd frontend && npm ci && npm run build
# 4. rm -rf backend/web/dist && cp -r frontend/dist backend/web/dist
# 5. cd backend && go build -o ../bin/pm-work-tracker-{env} ./cmd/server
# 6. Print success message
#
# set -e: any step failure stops immediately with non-zero exit
```

## .gitignore Additions

```
backend/config.yaml
backend/web/dist/
bin/
```

## Open Questions

- [x] API 路径是否去掉 `/api` 前缀 → 确认去掉
- [x] 配置模板方案 → config.yaml.example + gitignore
- [x] 构建脚本位置 → scripts/build.sh

## Appendix

### Alternatives Considered

| 方案 | Pros | Cons | 未选原因 |
|------|------|------|---------|
| 运行时读取 dist/ 目录 | 无需 embed | 部署需携带目录，破坏单体目标 | 违反零外部依赖需求 |
| 保留 `/api` 前缀 | 无破坏性变更 | 路径层级重复，与 PRD 设计不符 | PRD 明确要求去掉 |
| Nginx 反向代理 | 成熟方案 | 引入外部依赖，违反零依赖需求 | Out of scope |

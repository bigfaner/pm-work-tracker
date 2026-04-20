---
created: 2026-04-20
author: fanhuifeng
status: Draft
---

# Proposal: 前后端单体部署（Go embed.FS）

## Problem

当前项目前后端代码分离，但缺少生产部署方案：

- 后端 Go 服务仅提供 `/api/*` 接口，不服务任何静态文件
- 前端 `npm run build` 产出的 `dist/` 目录无人托管
- 没有 Dockerfile、docker-compose 或任何部署脚本
- 线上若要运行，需要额外配置 Nginx 或其他静态文件服务器，增加运维复杂度

本地开发体验良好（Vite dev server + proxy 已配置），无需改动。

## Proposed Solution

### 方案 A（推荐）：Go embed.FS 内嵌前端产物

利用 Go 标准库 `embed` 包，将 `frontend/dist/` 目录在编译时打包进 Go 二进制文件。生产部署只需一个二进制文件，无需额外的静态文件服务器。

**工作原理：**

```
构建流程：
  1. cd frontend && npm run build   → 生成 frontend/dist/
  2. cd backend && go build         → embed.FS 将 dist/ 打包进二进制

运行时路由：
  /api/*   → Gin handlers（现有逻辑不变）
  /*       → 从 embed.FS 提供静态文件，404 fallback 到 index.html（SPA 路由）
```

**变更范围：**
- `backend/internal/handler/router.go`：增加静态文件路由，`/api/*` 之外的请求 fallback 到 `index.html`
- `backend/cmd/server/main.go` 或新增 `backend/static/static.go`：声明 `//go:embed` 指令
- `scripts/build.sh`：编译与打包脚本，接受 `ENV` 参数（`dev` / `prod`），输出对应二进制
- 本地开发流程：**不变**，继续分两个终端启动

**环境与配置策略：**

| 环境 | 分支 | 数据库 | 配置文件 |
|------|------|--------|----------|
| 本地 | 任意 | SQLite | `config.yaml`（不提交，加入 .gitignore） |
| dev  | `dev` | MySQL（dev 库） | `config.dev.yaml`（提交，敏感字段留占位符） |
| prod | `main` | MySQL（prod 库） | `config.prod.yaml`（提交，敏感字段留占位符） |

二进制产物相同，通过 `-config` 参数指定配置文件。k8s 部署时：
- ConfigMap 挂载对应 config 文件（非敏感配置）
- Secret 以环境变量形式注入敏感字段（DB 密码、JWT secret），由 env override 机制覆盖 config 中的占位符

**业务价值：**
- 部署极简：scp 一个二进制 + config.yaml 即可上线
- 无运维依赖：不需要 Nginx、不需要容器编排
- 前后端版本强绑定：不会出现前后端版本不匹配的问题

**代价：**
- 每次前端改动都需要重新 go build（CI 中自动化后无感知）
- 二进制体积增大（前端 dist 通常 1-5 MB，可接受）
- 本地开发时不使用 embed 路径，需要用 build tag 区分（或直接接受本地不走 embed）

---

### 方案 B：Nginx 反向代理

Nginx 服务 `frontend/dist/` 静态文件，`/api/*` 反向代理到 Go 后端。

**业务价值：** 前后端独立部署，可分别扩缩容。  
**代价：** 需要维护 Nginx 配置，部署多一个进程，与"单体部署"目标不符。

---

### 方案 C：Docker 多阶段构建

Dockerfile 多阶段：node 构建前端 → go 构建后端 → 最终镜像包含二进制 + dist。

**业务价值：** 环境隔离，适合容器化部署。  
**代价：** 引入 Docker 依赖，当前项目规模下偏重。

---

## Recommended Approach

**方案 A**，理由：
1. 与"单体部署"目标完全吻合
2. 零运维依赖，部署最简单
3. 本地开发流程完全不受影响
4. Go embed 是标准库特性，无额外依赖

## Scope

**In scope：**
- `router.go` 增加静态文件服务（SPA fallback）
- embed 声明文件（`backend/static/static.go` 或 `main.go` 内）
- `scripts/build.sh`：接受 `ENV=dev|prod` 参数，校验分支、构建前端、构建后端、输出二进制
- `config.dev.yaml` / `config.prod.yaml`：提交到 git，敏感字段留占位符
- `.gitignore` 加入 `config.yaml`（本地开发用，不提交）

**Out of scope：**
- Docker / CI/CD 流水线
- 本地开发流程改动
- Nginx 配置
- 服务器上的进程管理（systemd / supervisor）

## Build Script Design

`scripts/build.sh` 核心逻辑：

```
用法: ./scripts/build.sh [dev|prod]

1. 读取 ENV 参数（默认 dev）
2. 校验当前 git 分支：
   - dev  → 必须在 dev 分支
   - prod → 必须在 main 分支
3. cd frontend && npm ci && npm run build
4. cd backend && go build -o ../dist/pm-work-tracker ./cmd/server
5. 输出产物路径提示
```

输出产物：`dist/pm-work-tracker`。

部署时指定对应配置文件：
```
./pm-work-tracker -config config.dev.yaml   # dev 环境
./pm-work-tracker -config config.prod.yaml  # prod 环境
```
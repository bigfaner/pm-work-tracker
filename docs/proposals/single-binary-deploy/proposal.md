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

**触发事件：** 2026-04-15 产品经理在团队群中提出"需要一个大家都能访问的测试环境"，原因是远程验收只能通过屏幕共享或本地录屏进行，无法独立操作和回归验证。同日测试人员反馈无法在自己的机器上独立运行前后端（缺少 Node.js 环境）。该请求直接推动了本提案。

**为什么现在解决：** 项目进入功能验收阶段，远程团队成员（产品经理、测试）需要访问 dev 环境。当前每次验证需本地启动前后端两个进程，阻塞跨地域验收。部署方案是上线前的必要前置条件。

## Proposed Solution

### 方案 A（推荐）：Go embed.FS 内嵌前端产物

利用 Go 标准库 `embed` 包，将 `frontend/dist/` 目录在编译时打包进 Go 二进制文件。生产部署只需一个二进制文件，无需额外的静态文件服务器。

**工作原理：**

```
构建流程：
  1. cd frontend && npm run build   → 生成 frontend/dist/
  2. cd backend && go build         → embed.FS 将 dist/ 打包进二进制

运行时路由：
  /api/*    → Gin handlers（现有逻辑不变）
  /         → 从 embed.FS 提供 index.html（SPA 兜底）
  /assets/* → 从 embed.FS 提供带 hash 的静态文件（JS/CSS/图片等）

路径映射：Vite 默认产出 dist/assets/<hash>.js，base 为 '/'。
Go embed 声明 `//go:embed dist/* dist/assets/*`，Gin 注册 /assets/* 路由指向 embed.FS 的 assets/ 子目录。
无需重命名 Vite 产出文件。
```

**错误状态行为：**

| 场景 | HTTP 状态码 | 响应体 | 日志行为 |
|------|------------|--------|----------|
| `/assets/<existing-hash>.js` | 200 | 文件内容 | 无日志（正常请求） |
| `/assets/nonexistent.js` | 404 | `{"error":"not found"}` JSON | `WARN: static asset not found path=/assets/nonexistent.js` |
| `/teams/1/items`（前端路由） | 200 | index.html 内容 | 无日志（SPA 兜底） |
| embed.FS 为空（build 失败但 go build 成功） | N/A | N/A | **启动拒绝**：`main()` 校验 embed.FS 中 index.html 存在，不存在则 `log.Fatal("index.html not found in embed.FS, did npm run build succeed?")` 并退出码 1 |
| 必填配置为占位符 | N/A | N/A | **启动拒绝**：`log.Fatal("config field X is placeholder, must override via env var")` 并退出码 1（同 R3 缓解措施） |

实现要点：`/assets/*` 路由在 embed.FS 中查找文件，找到则返回，未找到返回 404 JSON。`NoRoute` handler 仅处理非 `/assets/*` 的未匹配路径（即 SPA 路由），返回 index.html。这样 `/assets/nonexistent.js` 不会触发 SPA 兜底。

**变更范围：**
- `backend/internal/handler/router.go`：增加 `/` 路由返回 index.html，`/assets/*` 路由提供静态文件
- `backend/cmd/server/main.go` 或新增 `backend/static/static.go`：声明 `//go:embed` 指令
- `scripts/build.sh`：编译与打包脚本，接受 `ENV` 参数（`dev` / `prod`），输出对应二进制

**环境与配置策略：**

| 环境 | 分支 | 数据库 | 配置文件 |
|------|------|--------|----------|
| 本地 | 任意 | SQLite | `config.yaml`（不提交，加入 .gitignore） |
| dev  | `dev` | MySQL（dev 库） | `config.dev.yaml`（提交，敏感字段留占位符） |
| prod | `main` | MySQL（prod 库） | `config.prod.yaml`（提交，敏感字段留占位符） |

二进制产物相同，通过 `-config` 参数指定配置文件。敏感字段（DB 密码、JWT secret）在 config 中留占位符，部署时通过环境变量覆盖。

**业务价值：**
- 部署步骤：scp 二进制 + config 文件，执行一个命令即可启动
- 运行时依赖：仅 Linux 内核，不需要安装 Nginx、Node.js 或容器运行时
- 前后端版本强绑定：编译时锁定，不会出现前后端版本不匹配

**代价：**
- 前端改动需重新 `npm run build && go build` 并重启服务。CI 不在范围内，部署需手动执行 build.sh 并 scp。接受理由：当前仅一名开发者，发布频率约每周 1-2 次，构建耗时约 2 分钟。若频率增加应引入 CI 或迁移至方案 B。
- 二进制体积增大（前端 dist 通常 1-5 MB，可接受）
- 本地开发时不使用 embed 路径，需要用 build tag 区分（或直接接受本地不走 embed）

---

### 方案 B：维持现状（无部署方案）

不引入任何部署机制，继续以本地开发方式运行。远程团队通过屏幕共享或 VPN+本地启动进行验收。

**业务价值：** 零实现成本，不引入新的技术复杂度。
**代价：** 远程团队成员无法独立访问应用，每次验收需开发者在线配合；测试人员需自行安装 Node.js 和 Go 环境才能本地运行；前后端版本通过人工口头协调，易出现不一致。按当前每周 1-2 次验收节奏估算，每次协调耗时约 15-30 分钟，月累计约 2 小时开发者时间被占用。随着功能增加和测试回归范围扩大，该成本会持续增长。

---

### 方案 C：Nginx 反向代理

Nginx 服务 `frontend/dist/` 静态文件，`/api/*` 反向代理到 Go 后端。

**业务价值：** 前后端独立部署和扩缩容；Nginx 提供 gzip 压缩、静态资源缓存、CDN 集成等成熟能力。
**代价：** 需要维护 Nginx 配置文件，部署两个进程，运维复杂度高于单二进制方案。适合前后端团队独立发版、需要 CDN 或流量管理的场景。

---

### 方案 D：Docker 多阶段构建

Dockerfile 多阶段：node 构建前端 → go 构建后端 → 最终镜像包含二进制 + dist。

**业务价值：** 环境隔离，适合容器化部署和编排。
**代价：** 引入 Docker 依赖，镜像构建约 2-3 分钟（首次或依赖变更时）。当前项目只有一名开发者、一台服务器，Docker 抽象层收益不大。

---

## Recommended Approach

**方案 A**，理由：
1. 与"单体部署"目标完全吻合
2. 部署只需二进制 + config，无需额外软件
3. 本地开发流程完全不受影响
4. Go embed 是标准库特性，无额外依赖

## Scope

**In scope：**
- `router.go` 增加 `/` 返回 index.html、`/assets/*` 静态文件路由
- embed 声明文件（`backend/static/static.go` 或 `main.go` 内）
- `scripts/build.sh`：接受 `ENV=dev|prod` 参数，校验分支、构建前端、构建后端、输出二进制
- `config.dev.yaml` / `config.prod.yaml`：提交到 git，敏感字段留占位符
- `.gitignore` 加入 `config.yaml`（本地开发用，不提交）

**Out of scope：**
- Docker / CI/CD 流水线
- 本地开发流程改动
- Nginx 配置
- 服务器上的进程管理（systemd / supervisor）
- 结构化日志框架（如 zap/zerolog）引入或日志格式规范化——错误状态表中的 `log.Fatal` / `log.Printf` 使用 Go 标准 `log` 包，不引入额外依赖

## Build Script Design

`scripts/build.sh [dev|prod]`：校验 git 分支（dev→dev 分支，prod→main 分支），执行 `npm ci && npm run build`，再 `go build -o ../dist/pm-work-tracker`。部署时 `./pm-work-tracker -config config.dev.yaml`。

## Risk Assessment

| # | 风险 | 可能性 | 影响 | 缓解措施 |
|---|------|--------|------|----------|
| R1 | SPA 路由冲突：浏览器直接访问 `/teams/1/items` 等前端路由时，Go 路由找不到对应 API 路径，返回 404 而非 index.html | 高 | 高 | router.go 中 API 路由和 `/assets/*` 静态文件路由注册完成后，最后注册 `NoRoute` handler 返回 embed.FS 中的 index.html，让前端 Router 接管客户端路由。`/assets/*` 路由独立处理，缺失文件返回 404 而非 index.html，避免缺失静态资源被 SPA 兜底吞掉 |
| R2 | 静态资源缓存：浏览器缓存旧的 JS/CSS，前端更新后用户看到旧版本或白屏 | 中 | 中 | 按文件类型设置 Cache-Control：`/` 和 `/index.html` 响应头设为 `Cache-Control: no-cache`（每次拉取最新入口）；`/assets/*` 下带 hash 文件名的响应设为 `Cache-Control: max-age=31536000, immutable`（长期缓存，文件名含 hash 时内容变更即路径变更） |
| R3 | 配置误部署：config 文件中 DB 密码等占位符未被环境变量覆盖，导致服务启动失败或连接错误数据库 | 中 | 高 | 服务启动时校验必填字段（DB 密码、JWT secret）非空非占位值，若为空则拒绝启动并打印明确错误信息；占位符使用明显标记如 `<PLACEHOLDER>` |

## Success Criteria

以下每项均为可验证的验收条件：

1. **embed 静态文件路由**：`scripts/build.sh prod` 构建后，启动二进制，`GET /` 返回 200 和 index.html 内容；`GET /assets/*.js` 返回对应 JS 文件且 Content-Type 为 `application/javascript`
2. **SPA 路由兜底**：浏览器直接访问 `/teams/1/items`（非 API 路径），返回 200 和 index.html，前端 Router 正确渲染对应页面
3. **缺失静态资源返回 404**：`GET /assets/nonexistent.js` 返回 404 和 JSON `{"error":"not found"}`，而非 index.html 兜底；服务端日志输出包含请求路径的 WARN 级别日志
4. **API 路由不受影响**：所有现有 `/api/*` 接口的集成测试全部通过，无回归
5. **构建脚本 — dev**：`scripts/build.sh dev` 在 dev 分支执行成功，输出 `dist/pm-work-tracker` 二进制；在非 dev 分支执行时脚本以非零退出码退出并打印分支错误
6. **构建脚本 — prod**：`scripts/build.sh prod` 在 main 分支执行成功；在非 main 分支执行时脚本以非零退出码退出并打印分支错误
7. **缓存头**：`GET /` 响应包含 `Cache-Control: no-cache`；`GET /assets/<hash>.js` 响应包含 `Cache-Control: max-age=31536000, immutable`
8. **配置文件**：`config.dev.yaml` 和 `config.prod.yaml` 提交到 git，其中 DB 密码和 JWT secret 字段为占位符 `<PLACEHOLDER>`；`.gitignore` 包含 `config.yaml`
9. **配置校验**：使用未修改的 config 文件（占位符未覆盖）启动服务时，进程退出并打印包含 "PLACEHOLDER" 的错误信息
10. **embed 校验**：若 embed.FS 中 index.html 不存在（如 npm build 失败但 go build 成功），服务启动时以退出码 1 退出并打印 "index.html not found in embed.FS" 错误信息
11. **本地开发不受影响**：`make dev`（或等价命令）启动前后端开发服务器，现有测试套件全部通过，无需执行 build.sh

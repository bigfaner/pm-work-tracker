# PM Work Tracker

项目管理工作追踪系统。前端基于 React + Vite，后端基于 Go + Gin，支持本地分离开发和单二进制生产部署。

---

## 本地开发

### 环境要求

- Go 1.21+
- Node.js 18+

### 第一步：后端配置

复制配置文件并修改：

```bash
cp backend/config.yaml.example backend/config.yaml
```

编辑 `backend/config.yaml`，至少修改以下字段：

```yaml
server:
  port: "8080"
  gin_mode: "debug"   # 本地开发建议改为 debug

auth:
  jwt_secret: "替换为至少32字节的随机字符串"
  initial_admin:
    username: "admin"
    password: "替换为初始管理员密码"

cors:
  origins:
    - "http://localhost:5173"   # 前端开发服务器地址，保持默认即可
```

数据库默认使用 SQLite，文件路径为 `backend/data/app.db`，首次启动自动创建，无需额外配置。

### 第二步：启动后端

```bash
cd backend
go run cmd/server/main.go -dev
```

`-dev` 标志跳过前端静态资源校验，后端以纯 API 模式运行，监听 `http://localhost:8080`。

### 第三步：启动前端

```bash
cd frontend
npm install
npm run dev
```

前端开发服务器启动在 `http://localhost:5173`，Vite 已配置代理将 `/v1` 请求转发到后端：

```ts
// vite.config.ts
server: {
  proxy: {
    '/v1': 'http://localhost:8080',
  },
}
```

前端无需额外配置 API 地址，代理自动处理跨域。

### 前端环境变量（可选）

在 `frontend/` 目录创建 `.env.local`：

```env
# 若后端部署在子路径下（生产环境用），本地开发通常留空
VITE_BASE_PATH=
```

---

## 生产部署

生产部署使用单二进制模式：前端静态资源编译后嵌入 Go 二进制，只需分发一个可执行文件。

### 构建

构建脚本要求在对应分支上执行：

```bash
# 生产构建（必须在 main 分支）
./scripts/build.sh prod

# 开发环境构建（必须在 dev 分支）
./scripts/build.sh dev
```

构建流程：
1. `npm ci && npm run build` — 编译前端到 `frontend/dist/`
2. 将 `frontend/dist/` 复制到 `backend/web/dist/`（Go embed 目录）
3. `go build` — 将前端资源嵌入，输出到 `bin/pm-work-tracker`

### 部署

将以下文件上传到服务器：

```
bin/pm-work-tracker   # 可执行文件
backend/config.yaml        # 配置文件（需提前准备好）
```

### 服务器配置

在服务器上创建 `config.yaml`（参考 `backend/config.yaml.example`），重点修改：

```yaml
server:
  port: "8080"
  gin_mode: "release"
  base_path: "/pm"       # 若部署在子路径如 /app，填写 /app；默认为 /pm

database:
  driver: "sqlite"
  path: "/data/app.db"   # 建议使用绝对路径，确保目录存在且可写

auth:
  jwt_secret: "生产环境随机密钥，至少32字节"
  jwt_expiry: "24h"
  initial_admin:
    username: "admin"
    password: "强密码"

cors:
  origins:
    - "https://your-domain.com"   # 替换为实际域名；若前后端同域可留空列表
```

### 启动

```bash
./pm-work-tracker --config config.yaml
```

启动时会自动校验嵌入的前端资源是否完整，若缺失会立即报错退出。

### 子路径部署

应用默认挂载在 `/pm` 路径下（`https://example.com/pm`）。`build.sh` 会自动读取 `backend/config.yaml` 中的 `server.base_path` 作为前端构建的 `VITE_BASE_PATH`，两者始终保持一致。

若需要修改子路径，只需修改 `backend/config.yaml`：

```yaml
server:
  base_path: "/pm"
```

然后重新执行 `./scripts/build.sh prod` 即可。

### 反向代理（Nginx 示例）

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 配置参考

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `server.port` | 监听端口 | `8080` |
| `server.gin_mode` | Gin 模式，`debug` 或 `release` | `release` |
| `server.base_path` | 子路径前缀，根路径留空 | `/pm` |
| `server.read_timeout` | HTTP 读超时 | `30s` |
| `server.write_timeout` | HTTP 写超时 | `30s` |
| `database.driver` | 数据库驱动，`sqlite` 或 `mysql` | `sqlite` |
| `database.path` | SQLite 文件路径 | `./data/app.db` |
| `database.url` | MySQL 连接串（使用 mysql 时填写） | `""` |
| `auth.jwt_secret` | JWT 签名密钥，**必须修改**，≥32字节 | — |
| `auth.jwt_expiry` | Token 有效期 | `24h` |
| `auth.initial_admin.username` | 初始管理员用户名 | `admin` |
| `auth.initial_admin.password` | 初始管理员密码，**必须修改** | — |
| `cors.origins` | 允许的跨域来源列表 | `["http://localhost:5173"]` |

---

## FAQ

**Q: 线上部署时，`vite.config.ts` 中的 `server.proxy` 会有影响吗？**

不会。`vite.config.ts` 的 `server` 配置块（包括 `proxy`）只在 `npm run dev` 启动开发服务器时生效，`npm run build` 编译时完全忽略这部分配置，不会打包进产物。

线上部署时，前端静态文件已嵌入 Go 二进制，由 Go 直接提供服务。前端发出的 `/v1/...` 请求打到同一个 Go 进程，不存在跨域问题，也不需要任何代理。

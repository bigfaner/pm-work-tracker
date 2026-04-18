---
feature: "config-yaml"
---

# User Stories: config-yaml

## Story 1: 开发者使用 YAML 管理本地配置

**As a** 开发者
**I want to** 通过编辑 config.yaml 文件来设置所有后端配置（数据库、端口、日志等）
**So that** 我不需要记住每个环境变量的名字，而是通过清晰的分组结构快速理解和修改配置

**Acceptance Criteria:**
- Given 项目根目录存在 config.yaml 文件
- When 应用启动
- Then 应用读取 YAML 配置并按功能域分组加载所有配置项

---

## Story 2: 运维通过环境变量覆盖敏感配置

**As a** 运维/部署人员
**I want to** 通过环境变量覆盖 config.yaml 中的任意字段
**So that** 敏感信息（JWT secret、管理员密码）不需要写入配置文件，可以通过部署平台的安全变量管理

**Acceptance Criteria:**
- Given config.yaml 中 auth.jwt_secret 为空或弱值
- When 设置环境变量 AUTH_JWT_SECRET 为 >= 32 字节的值
- Then 应用使用环境变量中的值，覆盖 YAML 配置

---

## Story 3: 启动时配置验证

**As a** 开发者/运维
**I want to** 在应用启动时立即发现配置错误（JWT secret 过短、端口无效、格式错误等）
**So that** 我不会带着错误配置运行服务，避免运行时出现难以排查的问题

**Acceptance Criteria:**
- Given auth.jwt_secret 少于 32 字节
- When 应用启动
- Then 应用拒绝启动并输出明确的错误信息 "auth.jwt_secret must be at least 32 bytes"

---

## Story 4: 首次启动自动创建管理员

**As a** 开发者
**I want to** 在 config.yaml 中配置初始管理员账号，首次启动时自动创建
**So that** 我不需要手动注册或通过数据库插入第一个管理员用户

**Acceptance Criteria:**
- Given config.yaml 中 auth.initial_admin.username 和 password 均非空
- When 应用首次启动（数据库中不存在该用户）
- Then 自动创建该管理员账号，密码哈希存储
- When 应用再次启动（用户已存在）
- Then 跳过创建，不影响现有数据

---

## Story 5: 数据库连接池调优

**As a** 运维/部署人员
**I want to** 通过配置文件控制数据库连接池参数（最大连接数、空闲连接数、连接存活时间）
**So that** 我可以根据实际负载调整数据库连接策略，避免连接耗尽或资源浪费

**Acceptance Criteria:**
- Given config.yaml 中 database.max_open_conns 设为 20
- When 应用启动并连接数据库
- Then 数据库连接池的最大打开连接数为 20

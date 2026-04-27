---
feature: "db-dialect-compat"
---

# User Stories: db-dialect-compat

## Story 1: 需求池转主事项在 MySQL 下正常工作

**As a** 运维人员
**I want to** 在 MySQL 生产环境下将需求池条目转换为主事项
**So that** MySQL 部署可行性得到验证，生产环境可正常使用该功能

**Acceptance Criteria:**
- Given 本地 MySQL 8.0 实例已导入应用 schema 且应用以 MySQL 模式启动
- When 通过 API 将需求池条目转换为主事项（`POST /v1/teams/:id/item-pool/:id/convert-to-main`）
- Then 接口返回 200，主事项编号和子事项编号分别按规则正确生成：
  - 主事项编号格式为 `{teamCode}-{seq:05d}`（如 `TEAM-00042`）
  - 子事项编号格式为 `{mainCode}-{seq:02d}`（如 `TEAM-00042-03`）

---

## Story 2: 首次部署 MySQL 时 RBAC 迁移顺利完成

**As a** 运维人员
**I want to** 在空 MySQL 数据库上启动应用并自动完成 RBAC 数据迁移
**So that** 预设角色（superadmin、pm、member）和权限数据正确初始化，无需手动干预

**Acceptance Criteria:**
- Given 本地 MySQL 8.0 实例已导入 schema 但无迁移记录
- When 启动应用触发 RBAC 迁移
- Then `roles` 表包含 3 条预设角色（superadmin、pm、member），`role_permissions` 表包含对应的权限码，无 SQL 语法错误

---

## Story 3: 开发者编写新 SQL 时被自动化检查拦截

**As a** 开发者
**I want to** 在提交包含硬编码 SQLite 专属 SQL 的代码时收到明确提示
**So that** 新的 MySQL 不兼容问题在提交前被发现，不会流入生产

**Acceptance Criteria:**
- Given 开发者在 `backend/internal/repository/` 下的 `.go` 文件中写入了包含 `SUBSTR(` 或 `CAST(` 或 `datetime(` 或 `pragma_` 的原始 SQL
- When 执行 git commit 触发 lint-staged
- Then 提交被拦截，提示使用 `dialect` 包生成方言安全的 SQL
- Given 开发者在 repo 层使用 `dialect.CastInt()` / `dialect.Substr()` / `dialect.Now()` 生成 SQL（而非硬编码 SQLite 关键字）
- When 执行 git commit
- Then 提交正常通过，不被拦截（无假阳性）

---

## Story 4: MySQL 下列存在检查正常工作

**As a** 运维人员
**I want to** 在 MySQL 生产环境下执行 RBAC 迁移时 HasColumn 函数正确检测列是否存在
**So that** 迁移逻辑能根据列是否存在（如 `team_members.role`）选择正确的迁移路径，不会因 pragma_table_info 语法错误而中断

**Acceptance Criteria:**
- Given 本地 MySQL 8.0 实例已导入 schema 且 RBAC 迁移尚未执行
- When 启动应用触发 RBAC 迁移，流程执行到 `HasColumn` 调用
- Then `HasColumn` 通过 `information_schema.columns` 查询返回正确的布尔值（如 `HasColumn(db, 'pmw_team_members', 'role_key')` 返回 `true`，`HasColumn(db, 'pmw_team_members', 'nonexistent')` 返回 `false`），迁移顺利完成，无 SQL 语法错误

---

## Story 5: SQLite 环境下所有现有测试通过方言改造

**As a** 开发者
**I want to** 在 SQLite 环境下运行所有现有测试并全部通过
**So that** 方言改造不引入任何回归，日常开发流程保持正常

**Acceptance Criteria:**
- Given 应用以 SQLite 模式启动（默认配置）
- When 运行 `go test ./internal/... ./config/... ./cmd/...`
- Then 所有现有测试通过，无回归

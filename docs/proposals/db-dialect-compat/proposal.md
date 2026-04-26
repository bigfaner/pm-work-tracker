# Proposal: 数据库方言兼容性

## Problem

项目需长期同时支持 SQLite（开发/测试）和 MySQL（生产）。当前代码中存在 4 处硬编码的 SQLite 专属 SQL，在 MySQL 下会报语法错误。目前没有系统性机制防止新的不兼容 SQL 被引入。

**来源：** P1/P2 由生产事故发现（"需求池转主事项"接口返回 500），P3/P4 由后续代码审查发现。审查结论：不兼容点不止一处，需要系统性解决而非逐个修补。对应的紧急修复提交为 `86fd7c7`（`refactor(migration): separate DDL from DML, fix MySQL compatibility`），该提交在 `backend/internal/migration/rbac.go` 中以 if-else 分支逐点修补了 P3/P4，在 `backend/internal/repository/gorm/` 中修复了 P1/P2。本提案将此类临时修复重构为可维护的方言抽象层。

**为何现在重构：** 当前目标是在本地环境验证使用 MySQL 部署的完整可行性——确保所有功能（建表、迁移、业务操作）在 MySQL 下端到端可用。紧急修复 `86fd7c7` 解决了建表和迁移阶段的兼容性问题，但 repo 层的 P1/P2（`CAST AS INTEGER`）尚未修复，导致"需求池转主事项"等业务功能在 MySQL 下仍然 500。系统性修复全部兼容性问题后，才能完成 MySQL 部署可行性验证。

**已知的 4 个不兼容点：**

| # | 文件 | 行 | 问题 SQL | MySQL 正确写法 |
|---|------|----|---------|--------------|
| P1 | `repository/gorm/main_item_repo.go` | L95 | `CAST(SUBSTR(code,?) AS INTEGER)` | `CAST(SUBSTRING(code,?) AS SIGNED)` |
| P2 | `repository/gorm/sub_item_repo.go` | L116 | 同上 | 同上 |
| P3 | `migration/rbac.go` | L235, L261 | `INTEGER PRIMARY KEY AUTOINCREMENT` | `BIGINT UNSIGNED NOT NULL AUTO_INCREMENT` + `PRIMARY KEY(id)` |
| P4 | `migration/rbac.go` | L321 | `pragma_table_info` | `information_schema.columns` |

**影响：** P1/P2 导致"需求池转主事项"功能 500；P3 在 MySQL 下首次 RBAC 迁移会失败；P4 是导出函数，调用方在 MySQL 环境下会崩溃。

## Solution

### 用户可感知的行为变化

完成后，在 MySQL 环境下：

- "需求池转主事项"操作正常返回 200，不再出现 500 错误（修复 P1/P2）
- 首次 RBAC 迁移顺利完成，不再因 DDL 语法错误中断（修复 P3）
- 调用 `HasColumn` 的导出函数正常返回列信息，不再因 `pragma_table_info` 崩溃（修复 P4）
- SQLite（开发/测试）环境下所有现有功能不受影响

### 1. 引入方言辅助模块

在 `internal/pkg/db` 下新增 `dialect.go`，封装数据库方言差异判断和常用 SQL 片段生成：

- `IsMySQL(db) bool` — 复用现有 `migration.isMySQL` 逻辑，提升为可跨包使用
- `CastInt(expr) string` — 返回 `CAST(expr AS INTEGER)` 或 `CAST(expr AS SIGNED)`
- `Substr(str, start) string` — 返回 `SUBSTR(...)` 或 `SUBSTRING(...)`
- `Now() string` — 返回 `datetime('now')` 或 `CURRENT_TIMESTAMP`

导出为 `Dialect` 结构体（`type Dialect struct { mysql bool }`），由 `NewDialect(db *gorm.DB) *Dialect` 工厂函数创建。Repository 层通过构造函数接收 `*Dialect` 实例，调用其方法生成方言安全的 SQL。后续新增方言差异只需在 `Dialect` 上添加方法，不破坏现有调用签名。

### 2. 修复 4 个已知不兼容点

- P1/P2：repo 构造函数注入 `*Dialect`，`NextCode` 查询使用 `dialect.CastInt` + `dialect.Substr`
- P3：`rebuildTeamMembersTable` 增加 `isMySQL()` 分支，分别生成对应 DDL
- P4：`HasColumn` 增加 `isMySQL()` 分支，与 `columnExists` 保持一致

### 3. 防复发：代码规范

在 `docs/lessons/` 新增一条经验教训，并更新 `.claude/rules/` ：

> **规则：禁止在 repository 层直接拼接原始 SQL 字符串。** 需要原始 SQL 时，必须通过 `internal/pkg/db/dialect` 包生成方言安全的 SQL 片段。

## Alternatives

### A. 维持现状（do nothing）—— 保留 `86fd7c7` 的临时补丁

- 优点：零额外改动，无需重构
- 缺点：P1/P2 未修复，MySQL 下"需求池转主事项"等业务功能仍返回 500，无法完成 MySQL 部署可行性验证；`rbac.go` 中的散落 if-else 分支和无统一方言管理的 repo 层 cast 逻辑永久保留；每次发现新的 MySQL 不兼容点，需在代码库中搜索所有 SQL 拼接处并逐个添加 ad-hoc 分支
- **评估：** 不推荐——P1/P2 阻塞了 MySQL 部署验证目标，不做则目标无法达成

### B. GORM 回调/插件自动改写 SQL

- 优点：对业务代码透明
- 缺点：GORM 没有内置的 SQL 改写钩子，需实现 `Callback` 注册（~150 行 boilerplate 注册 `gorm.Callback().Query().Before("gorm:query")` 拦截器 + 正则匹配替换 SQL 片段），且 GORM v2 的 `Callbacks` API 变动频繁（近 3 个 minor 版本有 breaking change）；正则替换脆弱，无法处理动态拼接的表达式
- **评估：** 不推荐——预估 ~150 行框架胶水代码 + 正则维护成本，而当前仅 4 个不兼容点，投入产出比不合理

### C. 引入 dialect 辅助模块（推荐）

- 优点：显式、可测试、侵入性低（仅 `dialect.go` + `dialect_test.go` 两个新文件，约 60 行实现代码）；新增方言函数时有统一的地方去加
- 缺点：需要每个 repo 显式调用
- **评估：** 推荐——约 60 行实现代码覆盖 4 个已知不兼容点，每个 repo 仅需构造函数增加 1 个参数，不引入框架胶水层

## Scope

### In-scope

- 新增 `internal/pkg/db/dialect.go` 及测试
- 修复 P1-P4 四个不兼容点
- 修复 `rebuildTeamMembersTable` 的 DDL 方言分支
- 更新代码规范（rules 文件）
- 在 `scripts/lint-staged.sh` 中增加方言 SQL 关键字检查（~10 行），对 `backend/internal/repository/` 下的 `.go` 文件 grep 标记 `SUBSTR(`、`CAST(`、`datetime(`、`pragma_` 等 SQLite 专属关键字，发现则报错提示使用 `dialect` 包

### Out-of-scope

- 现有测试文件中的 SQLite 语法（测试用 SQLite 内存库，无需改）
- `SQLite-schema.sql` / `MySql-schema.sql`（已各自正确）
- GORM ORM 调用（自动适配，无问题）

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `Dialect` 初始化错误导致生产环境方言判断反转 | Low | **High** — 所有原始 SQL 语句使用错误方言，P1/P2 复现 500 错误，P3 迁移失败 | `NewDialect` 工厂函数从 `*gorm.DB` 自动检测方言（复用现有 `isMySQL` 的 `Dialector` 名称检查），不暴露 `mysql bool` 字段；新增测试用例模拟 MySQL/SQLite Dialector 验证检测结果 |
| dialect 模块累积过多条件分支，成为不可维护的 if-else 集合 | Medium | Medium — 模块复现其试图解决的问题，每次新增方言差异需读懂更多分支 | 当前仅 4 个已知差异，预计增长缓慢（SQLite 与 MySQL 的常见差异约 8-10 个，见 https://www.sqlite.org/quirks.html）；若超过 15 个方法则考虑拆分为 `sqlite.go` / `mysql.go` 两个文件，每文件只包含一种方言实现 |
| 遗漏新的方言差异 | Medium | Medium — 新增原始 SQL 时仍可能硬编码方言 | dialect 包集中管理 + 代码规范要求 repo 层原始 SQL 必须通过 dialect 包生成 + 自动化 lint 检查（见 scope） |
| Repo 构造函数变更影响测试 | Low | Low — 测试需补传 `*Dialect` 参数 | 测试中构造 `&Dialect{mysql: false}`（SQLite）即可；新增测试用例验证 `mysql: true`（MySQL）分支 |

## Success Criteria

1. MySQL 环境下"需求池转主事项"功能返回 200（覆盖 P1/P2）
2. MySQL 环境下首次 RBAC 迁移成功完成（覆盖 P3）
3. MySQL 环境下 `HasColumn` 对 `information_schema.columns` 查询返回正确结果，无 panic（覆盖 P4）
4. SQLite 环境下所有现有测试继续通过
5. `dialect.go` 单元测试覆盖所有导出函数（每个函数至少 2 组用例：SQLite / MySQL）
6. `.claude/rules/` 中包含禁止 repo 层直接拼接原始 SQL 的规范，且 `docs/lessons/` 中有对应的经验教训文档

**验证环境：** 成功标准 1-3 通过本地连接 MySQL 实例进行集成测试验证；标准 4-6 通过 `go test` 单元测试验证。

**测试环境前置条件（成功标准 1-3）：** 本地 MySQL 8.0 实例，已通过 `mysql -u root < backend/migrations/MySql-schema.sql` 导入应用 schema；应用配置文件中 `database.driver` 设为 `"mysql"`、`database.url` 设为指向该实例的 DSN（如 `root:@tcp(127.0.0.1:3306)/pm_work_tracker?parseTime=true`）、`auto_schema` 设为 `false`（schema 已手动导入，无需应用自动建表）。验证 schema 导入成功：`mysql -u root -e 'SHOW TABLES' pm_work_tracker | grep pmw_users` 应返回表名；若无输出则说明 schema 未正确导入，需重新执行导入命令。

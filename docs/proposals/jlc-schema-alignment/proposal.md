---
title: Schema 对齐嘉立创数据库开发规范（MySQL 兼容）
slug: jlc-schema-alignment
status: draft
created: 2026-04-24
---

## Problem

当前 `backend/migrations/schema.sql` 基于 SQLite 语法编写，存在以下问题：

1. **无法直接迁移到 MySQL**：缺少字符集声明、使用 SQLite 专属语法（`AUTOINCREMENT`、`REAL`、`BOOLEAN`）、无 `ON UPDATE CURRENT_TIMESTAMP`。将当前 `schema.sql` 直接在 MySQL 8.0 执行会立即报错，例如第 6 行 `INTEGER PRIMARY KEY AUTOINCREMENT` 会触发：
   ```
   ERROR 1064 (42000): You have an error in your SQL syntax; check the manual
   that corresponds to your MySQL 8.0 server version for the right syntax to use
   near 'AUTOINCREMENT'
   ```
   `BOOLEAN` 在 SQLite 中是文本别名，在 MySQL 中是 `TINYINT(1)` 别名，语义不同；`REAL` 在 MySQL 中无对应类型，需显式映射为 `DECIMAL`。

2. **命名不符合 JLC 规范**：软删字段用 `deleted_at`（应为 `deleted_flag + deleted_time`），更新时间用 `updated_at`（应为 `db_update_time`），缺少业务唯一键 `biz_key`。以下为规范原文（JLCZD-03-016，字段命名规范 [强制]）：

   > 软删除关键字：`deleted_flag`；删除时间戳：`deleted_time`

   以及通用字段规范 [强制]：

   > 软删：表示软删除状态的关键字: `deleted_flag`，业务字段+删除状态字段+删除时间戳字段=唯一键，删除时间默认值 `1970-01-01 08:00:00`

   时间字段规范 [强制]：

   > `create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`，`db_update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`

3. **TEXT 字段未受控**：`description`、`background`、`achievement` 等 7 个 TEXT 字段无长度约束，在 MySQL 中会导致行存储膨胀、临时表内存分配过大。
4. **使用 MySQL 关键字作字段名**：`status` 是 MySQL 8.0 关键字，直接用作字段名存在 SQL 注入风险和语法冲突隐患（规范原文 [强制]：禁止使用 MySQL 关键字和保留字作为表名和字段名）。
5. **数值类型不精确**：`completion REAL` 在 MySQL 中映射为浮点数，存在精度问题；应使用 `DECIMAL`。

**为何现在解决**：项目正在准备迁移至 MySQL 生产环境，当前 SQLite schema 无法在 MySQL 上直接执行，是部署的硬性阻塞项。每推迟一个迭代，新增的表和字段都会在迁移时产生额外的命名转换工作量。

## Proposal

产出一份 MySQL 兼容的 `schema.sql`，完整对齐 JLC 数据库开发规范，作为后续迁移 MySQL 的基准文档。

### 变更清单

#### 1. 全局：表级声明

每张表末尾统一加：

```sql
ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='...'
```

#### 2. 全局：通用字段重命名

| 当前字段 | 新字段 | 类型 | 说明 |
|---------|--------|------|------|
| `created_at DATETIME` | `create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP` | DATETIME | JLC 强制 |
| `updated_at DATETIME` | `db_update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | DATETIME | JLC 强制 |
| `deleted_at DATETIME` | `deleted_flag TINYINT(1) NOT NULL DEFAULT 0` + `deleted_time DATETIME NOT NULL DEFAULT '1970-01-01 08:00:00'` | TINYINT + DATETIME | JLC 强制 |
| *(缺失)* | `biz_key BIGINT NOT NULL` | BIGINT | JLC 强制，业务关联键 |

#### 3. 全局：主键类型

`INTEGER PRIMARY KEY AUTOINCREMENT` → `BIGINT UNSIGNED NOT NULL AUTO_INCREMENT`

#### 4. 全局：BOOLEAN → TINYINT(1)

`BOOLEAN` 在 MySQL 中是 `TINYINT(1)` 的别名，显式声明更清晰：

| 字段 | 变更 |
|------|------|
| `is_super_admin BOOLEAN` | `is_super_admin TINYINT(1) NOT NULL DEFAULT 0` |
| `is_key_item BOOLEAN` | `is_key_item TINYINT(1) NOT NULL DEFAULT 0` |
| `is_pm_correct BOOLEAN` | `is_pm_correct TINYINT(1) NOT NULL DEFAULT 0` |
| `is_auto BOOLEAN` | `is_auto TINYINT(1) NOT NULL DEFAULT 0` |

#### 5. 全局：REAL → DECIMAL

`completion REAL` → `completion DECIMAL(5,2) NOT NULL DEFAULT 0.00`（范围 0.00~100.00）

#### 6. 全局：`status` 关键字重命名

| 表 | 当前 | 新字段名 |
|----|------|---------|
| `users` | `status VARCHAR(10)` | `user_status VARCHAR(20)` |
| `main_items` | `status VARCHAR(20)` | `item_status VARCHAR(20)` |
| `sub_items` | `status VARCHAR(20)` | `item_status VARCHAR(20)` |
| `item_pools` | `status VARCHAR(20)` | `pool_status VARCHAR(20)` |

#### 7. TEXT 字段 → VARCHAR（按业务估算）

| 表 | 字段 | 当前 | 新类型 | 估算依据 |
|----|------|------|--------|---------|
| `main_items` | `description` | TEXT | `VARCHAR(2000)` | 事项背景描述，设计上限 500 字 |
| `sub_items` | `description` | TEXT | `VARCHAR(2000)` | 同上 |
| `item_pools` | `background` | TEXT | `VARCHAR(2000)` | 需求背景，设计上限 500 字 |
| `item_pools` | `expected_output` | TEXT | `VARCHAR(1000)` | 预期产出，设计上限 200 字 |
| `progress_records` | `achievement` | TEXT | `VARCHAR(1000)` | 每日进展，设计上限 200 字 |
| `progress_records` | `blocker` | TEXT | `VARCHAR(1000)` | 阻塞描述，设计上限 200 字 |
| `progress_records` | `lesson` | TEXT | `VARCHAR(1000)` | 经验教训，设计上限 200 字 |

**当前数据库实测**：对 `backend/data/dev.db` 执行 `SELECT MAX(LENGTH(field))` 查询，各字段最大值均为 0~15 字符（各表行数 10~20 行），数据量不具备统计意义。上述 VARCHAR 上限作为设计约束而非观测值设定。

**超限策略**：若上线前测试环境导入真实数据后，任一字段实际内容超出 VARCHAR 上限，MySQL 严格模式下将报错（`ERROR 1406: Data too long for column`）。处理方式：将该字段升级为 TEXT 并移入独立 detail 表（符合 JLC 大字段规范），不影响主表行存储。

#### 8. 索引命名对齐

当前索引命名已基本符合 `idx_` / `uk_` 前缀规范，无需大改。唯一调整：

- `CONSTRAINT idx_team_user UNIQUE` → `UNIQUE KEY uk_team_user (team_id, user_id)`（MySQL 语法）

#### 9. 软删唯一键调整

JLC 规范要求：`业务字段 + deleted_flag + deleted_time = 唯一键`。

受影响的唯一索引：

| 表 | 当前 | 新唯一键 |
|----|------|---------|
| `users` | `idx_users_username` | `uk_username_deleted (username, deleted_flag, deleted_time)` |
| `teams` | `idx_teams_code` | `uk_teams_code_deleted (code, deleted_flag, deleted_time)` |
| `main_items` | `idx_main_items_team_code` | `uk_main_items_team_code_deleted (team_id, code, deleted_flag, deleted_time)` |

#### 10. 后端 Go 代码适配

**Model 层**：

| 文件 | 变更 |
|------|------|
| `model/base.go`（或等效 BaseModel） | 移除 `gorm.Model` 嵌入；改为手动声明 `CreateTime`、`DbUpdateTime`、`DeletedFlag`、`DeletedTime`、`BizKey` |
| 各 model 文件 | `Status` → 对应业务字段名；`Completion float64` → `Completion decimal.Decimal` 或 `float64`（保持精度） |

**GORM 软删适配**：

当前 GORM 软删依赖 `gorm.DeletedAt`（`deleted_at IS NULL` 自动过滤）。改为 `deleted_flag` 后需：

1. 实现自定义 `Scope`：`db.Where("deleted_flag = 0")`，替代 GORM 内置软删过滤
2. 软删操作改为显式 UPDATE：`UPDATE ... SET deleted_flag=1, deleted_time=NOW() WHERE id=?`
3. 所有 repo 层查询统一通过 `NotDeleted()` scope，不再依赖 GORM 自动过滤

**Repository 层**：
- 所有 `Find`/`First`/`Count` 调用加 `.Scopes(NotDeleted)` 或等效条件
- 每个 repo 接口封闭 `SoftDelete(ctx, id)` 方法，内部执行 `UPDATE ... SET deleted_flag=1, deleted_time=NOW()`；禁止在 repo 外部直接调用 `db.Delete()`
- repo 接口不暴露硬删除方法，从接口层面杜绝误操作

**Service 层**：`Delete` 方法统一调用 repo 的 `SoftDelete(ctx, id)`，不感知底层实现。

### Observable Impact

本次变更涉及字段重命名，API 响应的 JSON 结构会发生以下变化：

| 受影响字段 | 当前 JSON key | 变更后 JSON key | 涉及接口 |
|-----------|--------------|----------------|---------|
| `status` (users) | `status` | `userStatus` | `/api/v1/users/*` |
| `status` (main_items) | `status` | `itemStatus` | `/api/v1/teams/:id/main-items/*` |
| `status` (sub_items) | `status` | `itemStatus` | `/api/v1/teams/:id/sub-items/*` |
| `status` (item_pools) | `status` | `poolStatus` | `/api/v1/teams/:id/item-pools/*` |
| `created_at` | `createdAt` | `createTime` | 所有资源接口 |
| `updated_at` | `updatedAt` | `dbUpdateTime` | 所有资源接口 |
| `deleted_at` | `deletedAt` | 字段消失（不对外暴露） | 所有资源接口 |

**这是一次破坏性 API 变更（breaking change）**。前端所有消费上述字段的代码（组件、API 模块、E2E 测试）必须与后端同步更新，不能分批部署。部署顺序：前端与后端在同一次发布中同时上线，不存在向后兼容窗口。

### 范围边界

| 影响项 | 分类 | 说明 |
|--------|------|------|
| `schema.sql` MySQL 化 | 本次范围 | 产出可在 MySQL 8.0 直接执行的 schema |
| 后端 Go model / repo / service 适配 | 本次范围 | 约 15~20 个文件，见第 10 节 |
| E2E 测试中涉及上述字段的断言更新 | 本次范围 | 字段名变更后测试必须同步修改才能通过 |
| 前端 API 模块及组件中的字段名更新 | 本次范围 | 与后端同一发布，不可延后 |
| API 合同文档（OpenAPI/Swagger）更新 | 延后 | 当前项目无正式 API 文档，待引入后补充 |
| 数据迁移脚本（从 SQLite 导出到 MySQL） | 不在本次范围 | 独立任务，依赖本次 schema 产出 |

**工作量估算**：预计 2~3 个迭代完成（schema + 后端适配 1 个迭代，前端 + E2E 同步 1 个迭代，联调验证 0~1 个迭代）。

**协调约束**：frontend JSON 字段重命名必须与 backend 同步部署，不可延后独立上线。若前端资源在目标迭代内不可用，本次变更需整体延后，不可拆分为仅后端先上线。

## Alternatives

### 方案 A（推荐）：完整对齐 JLC 规范，产出 MySQL 兼容 Schema

完整应用上述所有变更，产出可直接在 MySQL 8.0 执行的 `schema.sql`。

- **优点**：一次性对齐规范，后续迁移无历史包袱；字段名与 JLC 规范一一对应，code review 时无需查阅映射表
- **代价**：后端 model 层字段名需同步修改（约 15~20 个文件）

### 方案 B：仅做类型和字符集对齐，保留现有命名

保留 `deleted_at`、`updated_at`、`status` 等现有命名，只改 MySQL 语法和类型。

- **优点**：后端代码改动最小
- **代价**：不符合 JLC 规范；`status` 关键字风险未消除；下次需要对齐规范时，命名转换工作量与届时已积累的文件数成正比（当前已有 15~20 个文件，每个迭代递增）
- **B 的实际成本**：语法/类型修复（`AUTOINCREMENT`、`BOOLEAN`、`REAL`、字符集声明）仍需改动 `schema.sql` + 各 model 文件，约 5~8 个文件。方案 A 在此基础上额外增加字段重命名，约再涉及 10~12 个文件。A 相对 B 的增量成本是 10~12 个文件的命名变更，换取一次性规范对齐。

### 方案 C：维持现状

不做任何变更，继续使用 SQLite。

- **代价**：无法满足生产环境 MySQL 部署需求

## Risks

| # | 风险 | 可能性 | 影响 |
|---|------|--------|------|
| 1 | 后端代码改动量大 | 高 | 中 |
| 2 | VARCHAR 长度估算不足 | 中 | 低 |
| 3 | biz_key 生成策略未确定 | 低 | 中 |
| 4 | 软删唯一键变更与现有数据冲突 | 中 | 高 |

1. **后端代码改动量大**（可能性：高 / 影响：中）：`deleted_at` → `deleted_flag + deleted_time` 涉及 GORM 软删机制，需自定义 `DeletedAt` 替换为手动软删逻辑，影响所有 repo 层查询。改动量约 15~20 个文件，但每个文件的变更模式固定，不涉及业务逻辑。
   - 缓解：在 tech-design 阶段明确 GORM 软删适配方案，单独作为一个任务。

2. **VARCHAR 长度估算可能不足**（可能性：中 / 影响：低）：若实际内容超出 VARCHAR 上限，MySQL 会截断或报错。
   - 缓解：上线前在测试环境导入真实数据验证；保守估算时取 2 倍预期长度。

3. **biz_key 生成策略**（可能性：低 / 影响：中）：采用雪花算法（Snowflake）生成 64-bit BIGINT，需在后端引入或实现雪花算法库。
   - 缓解：Go 生态有成熟实现（如 `bwmarrin/snowflake`），引入成本低；需确定 worker-id 分配策略（单机可固定为 1）。

4. **软删唯一键变更影响现有数据**（可能性：中 / 影响：高）：若已有数据中存在 `deleted_flag=0` 且重复的 username / code，添加唯一键时会直接报错，迁移中断。
   - 脏数据检测（迁移前执行）：
     ```sql
     -- 检测 users 表重复 username
     SELECT username, COUNT(*) FROM users WHERE deleted_flag = 0 GROUP BY username HAVING COUNT(*) > 1;
     -- 检测 teams 表重复 code
     SELECT code, COUNT(*) FROM teams WHERE deleted_flag = 0 GROUP BY code HAVING COUNT(*) > 1;
     -- 检测 main_items 表重复 team_id+code
     SELECT team_id, code, COUNT(*) FROM main_items WHERE deleted_flag = 0 GROUP BY team_id, code HAVING COUNT(*) > 1;
     ```
   - 回滚方案：迁移脚本在事务中执行；若唯一键添加失败，`ROLLBACK` 回到迁移前状态，原表数据不受影响。若脏数据无法清理，改用部分索引（`WHERE deleted_flag = 0`）替代复合唯一键。

## Success Criteria

- [ ] 新 `schema.sql` 可在 MySQL 8.0 无报错执行
- [ ] 所有表包含 `create_time`、`db_update_time`、`deleted_flag`、`deleted_time`、`biz_key`
- [ ] 无 TEXT 字段，全部替换为有长度约束的 VARCHAR
- [ ] 无 `status` 关键字直接用作字段名
- [ ] 所有索引符合 `idx_` / `uk_` 命名规范
- [ ] 每张表有 `COMMENT` 说明
- [ ] `model/base.go` 不再嵌入 `gorm.Model`，改为手动声明 `CreateTime`、`DbUpdateTime`、`DeletedFlag`、`DeletedTime`、`BizKey`
- [ ] 每个 repo 接口包含 `SoftDelete(ctx, id)` 方法，且无 `db.Delete()` 直接调用出现在 repo 层外部（`grep -r "db.Delete" --include="*.go" backend/internal/service` 返回空）
- [ ] 所有 repo 层查询通过 `NotDeleted()` scope 过滤软删记录，`go test ./...` 全部通过
- [ ] 前端 API 模块中所有字段引用已更新（`userStatus`、`itemStatus`、`poolStatus`、`createTime`、`dbUpdateTime`），`npm test` 全部通过
- [ ] E2E 测试中涉及上述字段的断言已同步更新，E2E 套件全部通过

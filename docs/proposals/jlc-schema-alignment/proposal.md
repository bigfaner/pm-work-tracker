---
title: Schema 对齐嘉立创数据库开发规范（MySQL 兼容）
slug: jlc-schema-alignment
status: draft
created: 2026-04-24
---

## Problem

当前 `backend/migrations/schema.sql` 基于 SQLite 语法编写，存在以下问题：

1. **无法直接迁移到 MySQL**：缺少字符集声明、使用 SQLite 专属语法（`AUTOINCREMENT`、`REAL`、`BOOLEAN`）、无 `ON UPDATE CURRENT_TIMESTAMP`。
2. **命名不符合 JLC 规范**：软删字段用 `deleted_at`（应为 `deleted_flag + deleted_time`），更新时间用 `updated_at`（应为 `db_update_time`），缺少业务唯一键 `biz_key`。
3. **TEXT 字段未受控**：`description`、`background`、`achievement` 等 7 个 TEXT 字段无长度约束，在 MySQL 中会导致行存储膨胀、临时表内存分配过大。
4. **使用 MySQL 关键字作字段名**：`status` 是 MySQL 8.0 关键字，直接用作字段名存在 SQL 注入风险和语法冲突隐患。
5. **数值类型不精确**：`completion REAL` 在 MySQL 中映射为浮点数，存在精度问题；应使用 `DECIMAL`。

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
| `main_items` | `description` | TEXT | `VARCHAR(2000)` | 事项背景描述，通常 < 500 字 |
| `sub_items` | `description` | TEXT | `VARCHAR(2000)` | 同上 |
| `item_pools` | `background` | TEXT | `VARCHAR(2000)` | 需求背景，通常 < 500 字 |
| `item_pools` | `expected_output` | TEXT | `VARCHAR(1000)` | 预期产出，通常 < 200 字 |
| `progress_records` | `achievement` | TEXT | `VARCHAR(1000)` | 每日进展，通常 < 200 字 |
| `progress_records` | `blocker` | TEXT | `VARCHAR(1000)` | 阻塞描述，通常 < 200 字 |
| `progress_records` | `lesson` | TEXT | `VARCHAR(1000)` | 经验教训，通常 < 200 字 |

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

### 不在本次范围内

- 数据迁移脚本（从 SQLite 导出到 MySQL）

## Alternatives

### 方案 A（推荐）：完整对齐 JLC 规范，产出 MySQL 兼容 Schema

完整应用上述所有变更，产出可直接在 MySQL 8.0 执行的 `schema.sql`。

- **优点**：一次性对齐规范，后续迁移无历史包袱；字段语义清晰
- **代价**：后端 model 层字段名需同步修改（约 15~20 个文件）

### 方案 B：仅做类型和字符集对齐，保留现有命名

保留 `deleted_at`、`updated_at`、`status` 等现有命名，只改 MySQL 语法和类型。

- **优点**：后端代码改动最小
- **代价**：不符合 JLC 规范，技术债留存；`status` 关键字风险未消除

### 方案 C：维持现状

不做任何变更，继续使用 SQLite。

- **代价**：无法满足生产环境 MySQL 部署需求

## Risks

1. **后端代码改动量大**：`deleted_at` → `deleted_flag + deleted_time` 涉及 GORM 软删机制，需自定义 `DeletedAt` 替换为手动软删逻辑，影响所有 repo 层查询。
   - 缓解：在 tech-design 阶段明确 GORM 软删适配方案，单独作为一个任务。

2. **VARCHAR 长度估算可能不足**：若实际内容超出 VARCHAR 上限，MySQL 会截断或报错。
   - 缓解：上线前在测试环境导入真实数据验证；保守估算时取 2 倍预期长度。

3. **biz_key 生成策略**：采用雪花算法（Snowflake）生成 64-bit BIGINT，需在后端引入或实现雪花算法库。
   - 缓解：Go 生态有成熟实现（如 `bwmarrin/snowflake`），引入成本低；需确定 worker-id 分配策略（单机可固定为 1）。

4. **软删唯一键变更影响现有数据**：若已有数据 `deleted_flag=0` 且存在重复 username，唯一键会冲突。
   - 缓解：迁移前先清理脏数据，或将唯一键改为部分索引（`WHERE deleted_flag = 0`）。

## Success Criteria

- [ ] 新 `schema.sql` 可在 MySQL 8.0 无报错执行
- [ ] 所有表包含 `create_time`、`db_update_time`、`deleted_flag`、`deleted_time`、`biz_key`
- [ ] 无 TEXT 字段，全部替换为有长度约束的 VARCHAR
- [ ] 无 `status` 关键字直接用作字段名
- [ ] 所有索引符合 `idx_` / `uk_` 命名规范
- [ ] 每张表有 `COMMENT` 说明

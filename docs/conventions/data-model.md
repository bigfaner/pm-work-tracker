---
scope: global
source: feature/jlc-schema-alignment, feature/schema-alignment-cleanup
verified: "2026-05-04"
---

# Data Model Conventions

Rules governing database schema, model structs, and field naming. All new tables and model changes must comply.

## DM-001: Table Naming

**Rule**: All tables use the `pmw_` prefix. Engine is InnoDB with utf8mb4 charset.

**Why**: Aligns with JLC database development spec (JLCZD-03-016). The `pmw_` prefix prevents name collisions in shared MySQL instances.

**Example**:
```sql
CREATE TABLE pmw_users (
    ...
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

CREATE TABLE pmw_roles (
    ...
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';
```

**Go model**:
```go
func (User) TableName() string          { return "pmw_users" }
func (Role) TableName() string          { return "pmw_roles" }
func (RolePermission) TableName() string { return "pmw_role_permissions" }
```

## DM-002: BaseModel Fields

**Rule**: All business entities embed `BaseModel` which provides five standard fields. Two append-only tables (ProgressRecord, StatusHistory) are exceptions and declare their own fields without BaseModel.

**Why**: Centralizes audit columns, soft-delete support, and the external identifier. Eliminates per-table duplication of common DDL.

**BaseModel definition**:
```go
type BaseModel struct {
    ID           uint      `gorm:"primarykey;autoIncrement" json:"-"`
    BizKey       int64     `gorm:"not null" json:"bizKey"`
    CreateTime   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
    DbUpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"dbUpdateTime"`
    DeletedFlag  int       `gorm:"not null;default:0;index" json:"-"`
    DeletedTime  time.Time `gorm:"not null;default:'1970-01-01 08:00:00'" json:"-"`
}
```

**Key constraints**:
- `ID` is `json:"-"` -- never exposed in API responses. Used only for internal FK joins.
- `BizKey` is `int64` (snowflake), exposed as `json:"bizKey"` -- the sole external resource identifier.
- `DeletedFlag` and `DeletedTime` are `json:"-"` -- soft-delete metadata is never sent to clients.
- Primary key DDL: `BIGINT UNSIGNED NOT NULL AUTO_INCREMENT`.

## DM-003: FK Column Naming

**Rule**: FK columns storing external bizKey values use the `*_key` suffix. FK columns storing internal IDs use `*_id` suffix. No FOREIGN KEY constraints in DDL -- only indexes.

**Why**: After ID stopped being exposed, all FK relationships must reference bizKey. The `*_key` suffix distinguishes bizKey references from internal auto-increment IDs. DDL-level FK constraints are omitted to avoid cross-table cascading dependencies and simplify data migration.

**Examples**:
```sql
-- bizKey FK columns (*_key suffix)
team_key          BIGINT NOT NULL     -- references teams.biz_key
main_item_key     BIGINT NOT NULL     -- references main_items.biz_key
user_key          BIGINT NOT NULL     -- references users.biz_key
assignee_key      BIGINT              -- nullable bizKey reference
pm_key            BIGINT NOT NULL     -- team's PM user bizKey

-- Internal ID lookups (uint in Go, not in DDL columns)
FindByID(ctx, id uint)                -- repo-internal only
```

## DM-004: Status Field Naming

**Rule**: Status fields must be prefixed with the entity name: `item_status`, `pool_status`, `user_status`. Never use bare `status`.

**Why**: `status` is a MySQL 8.0 reserved keyword and causes SQL syntax errors. Entity-prefixed names also improve query readability in multi-table JOINs.

**Examples**:
```sql
-- pmw_users
user_status     VARCHAR(10) NOT NULL DEFAULT 'enabled'

-- pmw_main_items, pmw_sub_items
item_status     VARCHAR(20) NOT NULL DEFAULT 'pending'

-- pmw_item_pools
pool_status     VARCHAR(20) NOT NULL DEFAULT 'pending'
```

**Go model**:
```go
type User struct {
    BaseModel
    UserStatus string `gorm:"type:varchar(10);not null;default:'enabled'" json:"userStatus"`
}
type MainItem struct {
    BaseModel
    ItemStatus string `gorm:"type:varchar(20);not null;default:'pending'" json:"itemStatus"`
}
```

**Validation**: Enum values are validated at the service layer. Column type is always `VARCHAR(20)`, never MySQL `ENUM`.

## DM-005: Text Field Type Selection

**Rule**: Use VARCHAR for short fields. Use TEXT for long content fields (descriptions, backgrounds).

**Why**: TEXT is appropriate for unbounded content like descriptions and backgrounds. VARCHAR remains suitable for bounded fields with known max lengths.

**Mapping**:
| Content type | Type | Examples |
|-------------|------|----------|
| Long text (description, background, content, tags) | `type:text` | MainItem.ItemDesc, SubItem.ItemDesc, ItemPool.Background, ItemPool.ExpectedOutput, DecisionLog.Tags, DecisionLog.Content |
| Short string (title, code, name, status) | VARCHAR(100-200) | title VARCHAR(100), code VARCHAR(12), name VARCHAR(64) |
| Status / enum-like | VARCHAR(5-20) | priority VARCHAR(5), user_status VARCHAR(10), item_status VARCHAR(20) |

## DM-006: Completion Field Uses DECIMAL

**Rule**: The `completion` column uses `DECIMAL(5,2)`, not `REAL` or `FLOAT`.

**Why**: Floating-point types lose precision on repeated arithmetic. DECIMAL ensures exact storage of percentage values (0.00 to 100.00).

**Example**:
```sql
completion DECIMAL(5,2) NOT NULL DEFAULT 0.00
```

```go
Completion float64 `gorm:"type:decimal(5,2);not null" json:"completion"`
```

---

## DM-007: Core Entity Hierarchy

_Source: feature/pm-work-tracker_

```
MainItem → SubItem → ProgressRecord
                ↑
ItemPool ───────┘ (assign creates SubItem atomically)
```

- A **MainItem** belongs to one Team and contains zero or more SubItems.
- A **SubItem** belongs to one MainItem and has zero or more ProgressRecords.
- A **ProgressRecord** belongs to one SubItem and is append-only.
- An **ItemPool** entry, when assigned, atomically creates a SubItem under a MainItem (single DB transaction in `ItemPoolService.Assign`).

## DM-008: Completion Calculation (Weighted Average)

_Source: feature/pm-work-tracker_

`MainItem.Completion` is the weighted average of all its SubItems' completion values:

```
MainItem.Completion = sum(SubItem.Completion * SubItem.Weight) / sum(SubItem.Weight)
```

**Weight rule (v1):** All SubItems have `Weight = 1.0`, making this a simple average. The Weight field is reserved for future use.

**Recalculation trigger:** `MainItemService.RecalcCompletion` is called synchronously after every ProgressRecord append and after any SubItem completion change.

**Edge cases:** If a MainItem has zero SubItems, `Completion = 0`. If all SubItem weights are zero, `Completion = 0`.

**Why synchronous recalc:** No message queue, no cache layer, no background workers in v1. Progress rollup is computed synchronously. Data volume is small enough that this is fast.

**Recalculation triggers** (complete list):
1. ProgressRecord append (original trigger)
2. Sub-item completion change
3. Sub-item deletion -- recalculates parent main item
4. Sub-item move -- recalculates both source and target main items within the same transaction

## DM-009: ProgressRecord Append-Only

_Source: feature/pm-work-tracker_

ProgressRecords are append-only. Once created:

- **No DELETE** -- records are never removed.
- **No UPDATE** -- except for PM completion correction.
- No `UpdatedAt` or `DeletedAt` fields on the model.
- `ProgressRepo` has no delete or update methods (except `CorrectCompletion`).

**PM Correction Exception:** A PM may update the `Completion` field of an existing ProgressRecord, setting `IsPMCorrect = true` as an audit flag. No other field may be modified. This keeps the timeline clean while providing an audit trail.

**Why:** Append-only design preserves a complete history of progress updates for accountability and weekly reporting (PRD 5.4).

## DM-010: BizKey Snowflake Pattern

_Source: feature/pm-work-tracker_

MainItem uses a `Code` field (e.g. `"MI-0001"`) as a human-readable business key, auto-generated at creation time. This is separate from the database primary key (`ID`) and the `BizKey` snowflake field from BaseModel.

- `Code` has a unique index: `gorm:"size:10;not null;uniqueIndex"`
- Generation and collision handling managed in `MainItemService.Create`

## DM-011: Team Data Isolation at Query Layer

_Source: feature/pm-work-tracker_

Every repository method querying team-scoped data accepts `teamID uint` as a required parameter and applies `.Where("team_id = ?", teamID)` unconditionally. The `teamID` always comes from middleware-injected context, never from user-supplied request body or query params. Enforced by `TeamScopeMiddleware`.

## DM-012: Index Naming and Strategy

**Rule**: Follow established index naming and structure conventions from existing tables.

**Naming**:

| Index Type | MySQL Pattern | SQLite Pattern | Example |
|------------|---------------|----------------|---------|
| BizKey unique | `uk_biz_key` | `uk_<table>_biz_key` | `uk_biz_key (biz_key)` |
| Business unique with soft-delete | `uk_<desc>_deleted` | `uk_<table>_<desc>_deleted` | `uk_teams_code_deleted (team_code, deleted_flag, deleted_time)` |
| Non-unique composite | `idx_<table>_<desc>` | `idx_<table>_<desc>` | `idx_main_items_team_status (team_key, item_status)` |
| Deleted flag filter | `idx_<table>_deleted_flag` | `idx_<table>_deleted_flag` | `idx_main_items_deleted_flag (deleted_flag)` |

**Why**: MySQL uses shorter UK names (`uk_biz_key`) since indexes are table-scoped in DDL. SQLite CREATE INDEX requires explicit table name, so the `<table>_` prefix avoids ambiguity across tables.

**Strategy — prefer composite over standalone indexes**:

1. Use composite indexes `(col_a, col_b)` where possible. A composite index covers queries on its leftmost prefix, so `idx_t_status (team_key, status)` also serves `WHERE team_key = ?` queries.
2. Business unique indexes with `(deleted_flag, deleted_time)` trailing columns double as composite indexes for their leading columns. Example: `uk_milestone_maps_team_name_deleted (team_key, map_name, deleted_flag, deleted_time)` covers `WHERE team_key = ? AND map_name = ?` lookups.
3. Keep a standalone `idx_<table>_deleted_flag (deleted_flag)` for soft-delete filtering.
4. Do NOT create standalone single-column indexes when a composite index already covers that column as its leading prefix.

**Source**: /learn entry 2026-06-07 (milestone-map schema alignment)

## DM-013: Date Column Type

**Rule**: All date/time columns use `DATETIME` type, not `DATE`. Do not specify `DEFAULT NULL` on nullable columns.

**Why**: Existing tables use `DATETIME` consistently (e.g., `plan_start_date DATETIME`, `expected_end_date DATETIME`). Mixing `DATE` and `DATETIME` across tables creates inconsistency. MySQL nullable columns default to NULL implicitly — `DEFAULT NULL` is redundant.

**Example**:
```sql
-- Correct
plan_start_date DATETIME COMMENT '计划开始日期'
expected_end_date  DATETIME COMMENT '预计结束日期'

-- Wrong
planned_start_date DATE DEFAULT NULL
planned_end_date  DATE
```

**Source**: /learn entry 2026-06-07 (milestone-map schema alignment)

## DM-014: Schema Design Reference-First

**Rule**: Before writing any new table DDL, read the existing schema files to apply established conventions.

**Required reading**:
1. `backend/migrations/MySql-schema.sql`
2. `backend/migrations/SQLite-schema.sql`

**Checklist** (derive from existing tables, not from memory):
- UK naming: `uk_biz_key` (MySQL) / `uk_<table>_biz_key` (SQLite)
- Business unique: `uk_<desc>_deleted (..., deleted_flag, deleted_time)` pattern
- Date columns: `DATETIME`, no `DEFAULT NULL`
- Index naming: `idx_<table>_<desc>`
- Composite indexes over standalone where possible
- Nullable FK columns: type only, no `DEFAULT NULL`
- Every column has COMMENT (MySQL)

**Why**: Design-time schemas in `docs/features/<slug>/design/schema.sql` are often written without referencing runtime migrations, leading to naming and type drift. The reference-first rule prevents cascading rework when the schema reaches implementation.

**Source**: /learn entry 2026-06-07 (milestone-map schema alignment)

---
created: 2026-04-24
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: Schema 对齐嘉立创数据库开发规范（MySQL 兼容）

## Overview

本次变更分三层并行推进：

1. **Schema 层**：重写 `backend/migrations/schema.sql`，产出 MySQL 8.0 兼容的 DDL
2. **后端 Go 层**：替换 `BaseModel`，移除 GORM 软删依赖，为每个 repo 封闭 `SoftDelete()` 方法，引入雪花算法生成 `biz_key`
3. **前端层**：更新 `frontend/src/types/index.ts` 字段名，同步修改所有 API 模块和消费字段的组件

三层变更必须在同一次发布中同时上线（breaking change）。

## Architecture

### Layer Placement

```
schema.sql (DDL)
    ↓
model/base.go + model/*.go   ← BaseModel 替换，字段重命名
    ↓
repository/gorm/*.go         ← NotDeleted scope，SoftDelete 方法，raw SQL 更新
    ↓
service/*.go                 ← Delete 调用改为 SoftDelete，biz_key 赋值
    ↓
frontend/src/types/index.ts  ← 字段名更新
frontend/src/api/*.ts        ← 无需改动（字段名由 types 驱动）
frontend/src/pages/*.tsx     ← 消费 .status 的组件更新字段名
```

### Component Diagram

```
+------------------+     +------------------+     +------------------+
|  schema.sql      |     |  model/base.go   |     |  pkg/snowflake/  |
|  (MySQL DDL)     |     |  (new BaseModel) |     |  generator.go    |
+------------------+     +------------------+     +------------------+
                                  |                        |
                    +-------------+------------------------+
                    |
          +------------------+
          |  repo/gorm/*.go  |
          |  NotDeleted()    |
          |  SoftDelete()    |
          +------------------+
                    |
          +------------------+
          |  service/*.go    |
          |  biz_key assign  |
          +------------------+
                    |
          +------------------+
          |  frontend/types  |
          |  + pages/api     |
          +------------------+
```

### Dependencies

| 依赖 | 类型 | 说明 |
|------|------|------|
| `gorm.io/gorm` | 现有 | 移除 `gorm.Model` 嵌，保留 GORM 其他功能 |
| `gorm.io/driver/mysql` | 现有（go.mod 已有） | 已在 go.mod 中，无需新增 |
| `bwmarrin/snowflake` | 新增 | 生成 64-bit biz_key，单机 worker-id=1 |

## Interfaces

### 1. 新 BaseModel

```go
// backend/internal/model/base.go
type BaseModel struct {
    ID          uint      `gorm:"primarykey;autoIncrement" json:"id"`
    BizKey      int64     `gorm:"not null;uniqueIndex:uk_biz_key" json:"-"`
    CreateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
    DbUpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"dbUpdateTime"`
    DeletedFlag int       `gorm:"not null;default:0;index" json:"-"`
    DeletedTime time.Time `gorm:"not null;default:'1970-01-01 08:00:00'" json:"-"`
}
```

> `BizKey` 和软删字段均 `json:"-"`，不对外暴露。

### 1b. Deviation Model Structs（不嵌入 BaseModel）

`ProgressRecord` 和 `StatusHistory` 是 append-only 表，无 `biz_key`、无软删字段，不嵌入 `BaseModel`。`TeamMember` 有软删，嵌入 `BaseModel`。

类型映射：`BIGINT UNSIGNED` → `uint`，`BIGINT`（biz_key）→ `int64`，`DECIMAL(5,2)` → `float64`，`TINYINT(1)` → `int`。

```go
// backend/internal/model/progress_record.go
// append-only：无 biz_key，无软删字段
type ProgressRecord struct {
    ID          uint      `gorm:"primarykey;autoIncrement" json:"id"`
    SubItemID   uint      `gorm:"not null" json:"subItemId"`
    TeamID      uint      `gorm:"not null" json:"teamId"`
    AuthorID    uint      `gorm:"not null" json:"authorId"`
    Completion  float64   `gorm:"type:decimal(5,2);not null" json:"completion"`
    Achievement string    `gorm:"type:varchar(1000)" json:"achievement"`
    Blocker     string    `gorm:"type:varchar(1000)" json:"blocker"`
    Lesson      string    `gorm:"type:varchar(1000)" json:"lesson"`
    IsPmCorrect int       `gorm:"not null;default:0" json:"isPmCorrect"`
    CreateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
}

func (ProgressRecord) TableName() string { return "pmw_progress_records" }

// backend/internal/model/status_history.go
// append-only：无 biz_key，无软删字段
type StatusHistory struct {
    ID         uint      `gorm:"primarykey;autoIncrement" json:"id"`
    ItemType   string    `gorm:"type:varchar(20);not null" json:"itemType"`
    ItemID     uint      `gorm:"not null" json:"itemId"`
    FromStatus string    `gorm:"type:varchar(20);not null" json:"fromStatus"`
    ToStatus   string    `gorm:"type:varchar(20);not null" json:"toStatus"`
    ChangedBy  uint      `gorm:"not null" json:"changedBy"`
    IsAuto     int       `gorm:"not null;default:0" json:"isAuto"`
    Remark     string    `gorm:"type:varchar(200)" json:"remark"`
    CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
}

func (StatusHistory) TableName() string { return "pmw_status_histories" }

// backend/internal/model/team_member.go
// 嵌入 BaseModel：成员可被移出团队（软删）
type TeamMember struct {
    BaseModel
    TeamID   uint      `gorm:"not null" json:"teamId"`
    UserID   uint      `gorm:"not null" json:"userId"`
    RoleID   *uint     `json:"roleId"`
    JoinedAt time.Time `gorm:"not null" json:"joinedAt"`
}

func (TeamMember) TableName() string { return "pmw_team_members" }
```

### 2. NotDeleted Scope

```go
// backend/internal/repository/gorm/scopes.go
func NotDeleted(db *gorm.DB) *gorm.DB {
    return db.Where("deleted_flag = 0")
}
```

所有 repo 的 `Find`/`First`/`Count`/`List` 调用统一加 `.Scopes(NotDeleted)`。

### 3. SoftDelete 方法（每个 repo 接口）

受影响的 repo 接口（当前有 `Delete` 方法的）：

| Repo 接口 | 变更 |
|-----------|------|
| `TeamRepo` | `Delete` → `SoftDelete(ctx, id uint) error` |
| `SubItemRepo` | `Delete` → `SoftDelete(ctx, id uint) error` |
| `TeamMemberRepo` | 新增 `SoftDelete(ctx, id uint) error`（成员移出团队场景） |
| `RoleRepo` | `Delete` 保留（Role 无软删需求，硬删可接受） |

`SoftDelete` 实现：

```go
// backend/internal/repository/gorm/team_repo.go
func (r *teamRepo) SoftDelete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).
        Model(&model.Team{}).
        Where("id = ? AND deleted_flag = 0", id).
        Updates(map[string]any{
            "deleted_flag": 1,
            "deleted_time": time.Now(),
        }).Error
}
```

### 4. Snowflake Generator

```go
// backend/internal/pkg/snowflake/generator.go
package snowflake

import "github.com/bwmarrin/snowflake"

var node *snowflake.Node

func Init(workerID int64) error {
    var err error
    node, err = snowflake.NewNode(workerID)
    return err
}

func Generate() int64 {
    return node.Generate().Int64()
}
```

在 `main.go` 启动时调用 `snowflake.Init(1)`（单机固定 worker-id=1）。

### 5. Service 层 biz_key 赋值 & SoftDelete 调用

受影响的 service 文件及具体变更：

| Service 文件 | 变更 |
|-------------|------|
| `main_item_service.go` | `Create()` 赋值 `BizKey: snowflake.Generate()` |
| `sub_item_service.go` | `Create()` 赋值 `BizKey`；`Delete()` 改调 `subItemRepo.SoftDelete()` |
| `team_service.go` | `Create()` 赋值 `BizKey`；`Delete()` 改调 `teamRepo.SoftDelete()`；`RemoveMember()` 改调 `teamMemberRepo.SoftDelete()` |
| `item_pool_service.go` | `Create()` 赋值 `BizKey` |
| `auth_service.go` | `Register()` 赋值 `BizKey` |
| `progress_service.go` | `Create()` 赋值 `BizKey`（progress_records 无软删） |

```go
// 示例：main_item_service.go
item := &model.MainItem{
    BizKey: snowflake.Generate(),
    // ... 其他字段
}
```

### 6. 前端类型更新

```typescript
// frontend/src/types/index.ts — 受影响字段

// User
status?: 'enabled' | 'disabled'  →  userStatus?: 'enabled' | 'disabled'
createdAt: string                 →  createTime: string

// Team
createdAt: string                 →  createTime: string
updatedAt: string                 →  dbUpdateTime: string

// MainItem / SubItem
status: string                    →  itemStatus: string
createdAt: string                 →  createTime: string
updatedAt: string                 →  dbUpdateTime: string

// ItemPool
status: string                    →  poolStatus: string
createdAt: string                 →  createTime: string
updatedAt: string                 →  dbUpdateTime: string

// ProgressRecord
createdAt: string                 →  createTime: string

// Role
createdAt: string                 →  createTime: string
```

## Data Models

### 完整 schema.sql DDL（所有表，含 pmw_ 前缀）

```sql
-- pmw_users
CREATE TABLE pmw_users (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key         BIGINT          NOT NULL,
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    username        VARCHAR(64)     NOT NULL,
    display_name    VARCHAR(64)     NOT NULL,
    password_hash   VARCHAR(255)    NOT NULL,
    is_super_admin  TINYINT(1)      NOT NULL DEFAULT 0,
    email           VARCHAR(100)             DEFAULT '',
    user_status     VARCHAR(20)     NOT NULL DEFAULT 'enabled',
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_username_deleted (username, deleted_flag, deleted_time),
    KEY idx_users_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- pmw_teams
CREATE TABLE pmw_teams (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key         BIGINT          NOT NULL,
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    name            VARCHAR(100)    NOT NULL,
    description     VARCHAR(500),
    pm_id           BIGINT UNSIGNED NOT NULL,
    code            VARCHAR(6)      NOT NULL DEFAULT '',
    item_seq        INT             NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_teams_code_deleted (code, deleted_flag, deleted_time),
    KEY idx_teams_pm_id (pm_id),
    KEY idx_teams_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队表';

-- pmw_team_members（有软删：成员可被移出团队）
CREATE TABLE pmw_team_members (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key         BIGINT          NOT NULL,
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_id         BIGINT UNSIGNED NOT NULL,
    user_id         BIGINT UNSIGNED NOT NULL,
    role_id         BIGINT UNSIGNED,
    joined_at       DATETIME        NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_team_user_deleted (team_id, user_id, deleted_flag, deleted_time),
    KEY idx_team_members_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队成员表';

-- pmw_main_items
CREATE TABLE pmw_main_items (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key           BIGINT          NOT NULL,
    create_time       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag      TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time      DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_id           BIGINT UNSIGNED NOT NULL,
    code              VARCHAR(12)     NOT NULL,
    title             VARCHAR(100)    NOT NULL,
    description       VARCHAR(2000)   NOT NULL DEFAULT '',
    priority          VARCHAR(5)      NOT NULL,
    proposer_id       BIGINT UNSIGNED NOT NULL,
    assignee_id       BIGINT UNSIGNED,
    start_date        DATETIME,
    expected_end_date DATETIME,
    actual_end_date   DATETIME,
    item_status       VARCHAR(20)     NOT NULL DEFAULT '待开始',
    completion        DECIMAL(5,2)    NOT NULL DEFAULT 0.00,
    is_key_item       TINYINT(1)      NOT NULL DEFAULT 0,
    delay_count       INT             NOT NULL DEFAULT 0,
    archived_at       DATETIME,
    sub_item_seq      INT             NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_main_items_team_code_deleted (team_id, code, deleted_flag, deleted_time),
    KEY idx_main_items_team_id (team_id),
    KEY idx_main_items_assignee_id (assignee_id),
    KEY idx_main_items_expected_end_date (expected_end_date),
    KEY idx_main_items_team_status (team_id, item_status),
    KEY idx_main_items_team_priority (team_id, priority),
    KEY idx_main_items_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主事项表';

-- pmw_sub_items
CREATE TABLE pmw_sub_items (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key           BIGINT          NOT NULL,
    create_time       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag      TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time      DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_id           BIGINT UNSIGNED NOT NULL,
    main_item_id      BIGINT UNSIGNED NOT NULL,
    code              VARCHAR(15)     NOT NULL DEFAULT '',
    title             VARCHAR(100)    NOT NULL,
    description       VARCHAR(2000),
    priority          VARCHAR(5)      NOT NULL,
    assignee_id       BIGINT UNSIGNED,
    start_date        DATETIME,
    expected_end_date DATETIME,
    actual_end_date   DATETIME,
    item_status       VARCHAR(20)     NOT NULL DEFAULT '待开始',
    completion        DECIMAL(5,2)    NOT NULL DEFAULT 0.00,
    is_key_item       TINYINT(1)      NOT NULL DEFAULT 0,
    delay_count       INT             NOT NULL DEFAULT 0,
    weight            DECIMAL(5,2)    NOT NULL DEFAULT 1.00,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_sub_items_main_code (main_item_id, code),
    KEY idx_sub_items_main_item_id (main_item_id),
    KEY idx_sub_items_team_id (team_id),
    KEY idx_sub_items_assignee_id (assignee_id),
    KEY idx_sub_items_team_status (team_id, item_status),
    KEY idx_sub_items_team_priority (team_id, priority),
    KEY idx_sub_items_expected_end_date (expected_end_date),
    KEY idx_sub_items_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子事项表';

-- pmw_item_pools
CREATE TABLE pmw_item_pools (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key          BIGINT          NOT NULL,
    create_time      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag     TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time     DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_id          BIGINT UNSIGNED NOT NULL,
    title            VARCHAR(100)    NOT NULL,
    background       VARCHAR(2000),
    expected_output  VARCHAR(1000),
    submitter_id     BIGINT UNSIGNED NOT NULL,
    pool_status      VARCHAR(20)     NOT NULL DEFAULT '待分配',
    assigned_main_id BIGINT UNSIGNED,
    assigned_sub_id  BIGINT UNSIGNED,
    assignee_id      BIGINT UNSIGNED,
    reject_reason    VARCHAR(200),
    reviewed_at      DATETIME,
    reviewer_id      BIGINT UNSIGNED,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    KEY idx_item_pools_team_id (team_id),
    KEY idx_item_pools_team_status (team_id, pool_status),
    KEY idx_item_pools_submitter_id (submitter_id),
    KEY idx_item_pools_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='需求池表';

-- pmw_progress_records（append-only：无软删，无 biz_key）
CREATE TABLE pmw_progress_records (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    sub_item_id   BIGINT UNSIGNED NOT NULL,
    team_id       BIGINT UNSIGNED NOT NULL,
    author_id     BIGINT UNSIGNED NOT NULL,
    completion    DECIMAL(5,2)    NOT NULL,
    achievement   VARCHAR(1000),
    blocker       VARCHAR(1000),
    lesson        VARCHAR(1000),
    is_pm_correct TINYINT(1)      NOT NULL DEFAULT 0,
    create_time   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_progress_records_sub_item_id (sub_item_id),
    KEY idx_progress_records_sub_item_created (sub_item_id, create_time),
    KEY idx_progress_records_team_id (team_id),
    KEY idx_progress_records_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='进度记录表（追加写入）';

-- pmw_status_histories（append-only：无软删，无 biz_key）
CREATE TABLE pmw_status_histories (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    item_type    VARCHAR(20)     NOT NULL,
    item_id      BIGINT UNSIGNED NOT NULL,
    from_status  VARCHAR(20)     NOT NULL,
    to_status    VARCHAR(20)     NOT NULL,
    changed_by   BIGINT UNSIGNED NOT NULL,
    is_auto      TINYINT(1)      NOT NULL DEFAULT 0,
    remark       VARCHAR(200),
    create_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_item (item_type, item_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='状态变更历史表（追加写入）';
```

### TableName() 更新

所有 model 的 `TableName()` 方法需同步更新：

| Model | 当前 TableName | 新 TableName |
|-------|---------------|-------------|
| `User` | `users` | `pmw_users` |
| `Team` | `teams` | `pmw_teams` |
| `TeamMember` | `team_members` | `pmw_team_members` |
| `MainItem` | `main_items` | `pmw_main_items` |
| `SubItem` | `sub_items` | `pmw_sub_items` |
| `ItemPool` | `item_pools` | `pmw_item_pools` |
| `ProgressRecord` | `progress_records` | `pmw_progress_records` |
| `StatusHistory` | `status_histories` | `pmw_status_histories` |

## Error Handling

### Error Types & Codes

| Error Code | Name | Description | HTTP Status |
|------------|------|-------------|-------------|
| ERR_VALIDATION | ErrValidation | 字段值不在枚举范围内（如 item_status 非法值） | 400 |
| ERR_NOT_FOUND | ErrNotFound | SoftDelete 目标记录不存在或已删除 | 404 |

### Propagation Strategy

- `SoftDelete` 若 `RowsAffected == 0`（记录不存在或已软删），返回 `apperrors.ErrNotFound`
- service 层不捕获 `SoftDelete` 错误，直接透传给 handler
- handler 通过 `apperrors.RespondError` 统一响应

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| model | Unit | go test | BaseModel 字段 JSON tag 正确性 | 100% |
| repo | Unit (mock DB) | go test + sqlmock | NotDeleted scope 过滤、SoftDelete UPDATE 语句 | 90% |
| service | Unit (mock repo) | go test | biz_key 赋值、Delete 调用 SoftDelete | 90% |
| frontend types | Unit | vitest | 字段名变更后类型编译通过 | 100% |
| E2E | Integration | Playwright | 状态字段显示、软删后不可见 | 5 个场景全部通过 |

### Key Test Scenarios

1. `NotDeleted` scope：插入 `deleted_flag=1` 记录，List 不返回该记录
2. `SoftDelete`：调用后 `deleted_flag=1`，`deleted_time != '1970-01-01 08:00:00'`
3. `SoftDelete` 幂等：对已软删记录再次调用返回 `ErrNotFound`
4. `biz_key` 唯一性：并发创建 100 条记录，biz_key 无重复
5. 前端：`item.itemStatus` 正确渲染状态徽章（原 `item.status`）
6. 前端：`item.createTime` 正确显示时间（原 `item.createdAt`）

### E2E Scenarios（5 个，全部必须通过）

1. 软删后不可见：删除团队成员后，成员列表不再显示该成员
2. 状态字段显示：`itemStatus` 字段正确渲染状态徽章（原 `status`）
3. `createTime` 显示：事项详情页正确显示 `createTime`（原 `createdAt`）
4. 成员移出团队：`RemoveMember` 操作后，成员软删，团队成员数减 1
5. 需求池状态：`poolStatus` 字段正确显示需求池状态（原 `status`）

### Overall Coverage Target

后端：90%；前端：≥70%（当前基线，维持不降低）

## Security Considerations

### Threat Model

| 威胁 | 对策 |
|------|------|
| `status` 关键字冲突导致 DDL/DML parse error | 字段重命名为 `user_status`/`item_status`/`pool_status` |
| `biz_key` 暴露泄露雪花时间戳和 worker-id | `json:"-"` 阻止序列化；日志中禁止出现 biz_key（见下） |
| auto-increment `id` 暴露，攻击者可顺序枚举资源 | auth middleware 在所有资源端点校验 team membership，枚举越权 id 返回 403，不泄露数据 |
| biz_key 通过日志或错误响应间接泄露 | biz_key 不得出现在日志、error message、API 响应中；`json:"-"` 保证序列化层不泄露，logging middleware 不得打印完整 model struct（应只记录 id） |

> 注：worker-id=1（硬编码单机），雪花值编码了创建时间和 worker-id。一旦 biz_key 泄露，可反推创建时间和机器标识，因此日志管控是必要的。

### Mitigations

- repo 接口不暴露硬删除方法，从接口层面杜绝误操作
- `SoftDelete` 实现加 `deleted_flag = 0` 条件，防止重复软删并确保幂等性
- logging middleware 记录请求/响应时只记录 `id`，不打印完整 model struct

## PRD Coverage Map

| PRD AC | Design Component | Interface / Model |
|--------|-----------------|-------------------|
| schema.sql 在 MySQL 8.0 无报错执行 | schema.sql 重写 | DDL 全量替换 |
| 所有表含 create_time/db_update_time/deleted_flag/deleted_time/biz_key | BaseModel 替换 | `model/base.go` |
| 无 TEXT 字段 | schema.sql + model GORM tag | VARCHAR(1000/2000) |
| 无 status 关键字 | model 字段重命名 | UserStatus/ItemStatus/PoolStatus |
| 索引符合 idx_/uk_ 规范 | schema.sql | 索引命名全量检查 |
| 每张表有 COMMENT | schema.sql | 表级 COMMENT |
| model/base.go 不嵌入 gorm.Model | BaseModel 替换 | `model/base.go` |
| repo 封闭 SoftDelete，无 db.Delete 外部调用 | SoftDelete 接口 | `TeamRepo`/`SubItemRepo` |
| 所有 repo 查询通过 NotDeleted scope | NotDeleted scope | `scopes.go` |
| go test ./... 全部通过 | 各层单元测试 | 见 Testing Strategy |
| 前端字段引用更新，npm test 通过 | types/index.ts + pages | 字段名全量替换 |
| E2E 测试通过 | E2E 断言更新 | `__tests__/e2e/*.spec.ts` |

## Open Questions

- [x] biz_key 生成策略：雪花算法，worker-id=1（单机）
- [x] TEXT 字段处理：改为 VARCHAR，超限升级为独立 detail 表
- [x] 软删接口设计：repo 封闭 SoftDelete，禁止外部 db.Delete
- [ ] `progress_records` 和 `status_histories` 是否需要 biz_key？（当前设计：不需要，append-only 表无业务关联需求）

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| 保留 gorm.DeletedAt，仅改字段名 | 改动最小 | 不符合 JLC 规范，deleted_at 语义与 deleted_flag 不同 | 规范强制要求 |
| 使用 GORM 自定义软删插件 | 自动过滤 | 引入额外依赖，与 JLC 字段名不兼容 | 简单 scope 已足够 |
| biz_key 用 UUID | 无需外部库 | VARCHAR 类型，不符合 JLC BIGINT 要求 | JLC 规范要求 BIGINT |

### References

- JLC 规范：`docs/references/《嘉立创集团数据库开发规范》JLCZD-03-016【传阅】`
- PRD：`docs/features/jlc-schema-alignment/prd/prd-spec.md`
- 现有 BaseModel：`backend/internal/model/base.go`
- 现有 repo helpers：`backend/internal/pkg/repo/helpers.go`

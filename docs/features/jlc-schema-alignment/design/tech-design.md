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
handler/*.go                 ← 路径参数解析 uint→int64，GetByID→GetByBizKey
    ↓
frontend/src/types/index.ts  ← 字段名更新
frontend/src/api/*.ts        ← 无需改动（字段名由 types 驱动）
frontend/src/pages/*.tsx     ← 消费 .status 的组件更新字段名
```

### Component Diagram

```
+------------------+     +------------------+
|  schema.sql      |---->|  model/base.go   |
|  (MySQL DDL)     |     |  (new BaseModel) |
+------------------+     +------------------+
                                  |
                    +-------------+
                    |
          +------------------+
          |  repo/gorm/*.go  |
          |  NotDeleted()    |
          |  SoftDelete()    |
          +------------------+
                    |
          +------------------+     +------------------+
          |  service/*.go    |<----|  pkg/snowflake/  |
          |  biz_key assign  |     |  generator.go    |
          +------------------+     +------------------+
                    |
          +------------------+
          |  handler/*.go    |  ← 路径参数 uint→int64
          |  GetByBizKey     |    ParseInt(c.Param)
          +------------------+
                    |
          +-----------------------------+
          |  frontend/src/types/        |
          |  index.ts                   |
          +-----------------------------+
                    |
          +-----------------------------+
          |  frontend/src/api/*.ts      |
          +-----------------------------+
                    |
          +-----------------------------+
          |  frontend/src/pages/*.tsx   |
          +-----------------------------+
```

### Dependencies

| 依赖 | 类型 | 说明 |
|------|------|------|
| `gorm.io/gorm` | 现有 | 移除 `gorm.Model` 嵌，保留 GORM 其他功能 |
| `gorm.io/driver/mysql` | 现有（go.mod 已有） | 已在 go.mod 中，无需新增 |
| `bwmarrin/snowflake` | 新增 | 生成 64-bit biz_key，单机 worker-id=1 |
| `github.com/stretchr/testify` | 现有 | Go 测试断言库（assert/require） |
| `github.com/DATA-DOG/go-sqlmock` | 现有 | repo 层 mock DB 测试 |
| `vitest` | 现有 | 前端单元测试框架 |
| `@playwright/test` | 现有 | E2E 测试框架 |

### Deployment Constraints

- **Single-node only**: Current design hardcodes `worker-id=1` for snowflake generation. Multi-node deployment requires worker-id coordination (etcd/Redis) — out of scope for this iteration.
- **Violation consequence**: biz_key collisions will occur if multiple nodes share the same worker-id.

## Interfaces

### 1. 新 BaseModel

```go
// backend/internal/model/base.go
type BaseModel struct {
    ID          uint      `gorm:"primarykey;autoIncrement" json:"-"`
    BizKey      int64     `gorm:"not null;uniqueIndex:uk_biz_key" json:"bizKey"`
    CreateTime  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
    DbUpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"dbUpdateTime"`
    DeletedFlag int       `gorm:"not null;default:0;index" json:"-"`
    DeletedTime time.Time `gorm:"not null;default:'1970-01-01 08:00:00'" json:"-"`
}
```

> BizKey 对外暴露（`json:"bizKey"`）作为资源标识符；ID 不对外暴露（`json:"-"`），仅用于数据库内部关联。

### 1b. Deviation Model Structs（不嵌入 BaseModel）

`ProgressRecord` 和 `StatusHistory` 是 append-only 表，无 `biz_key`、无软删字段，不嵌入 `BaseModel`。`TeamMember` 有软删，嵌入 `BaseModel`。

类型映射：`BIGINT UNSIGNED` → `uint`，`BIGINT`（biz_key）→ `int64`，`DECIMAL(5,2)` → `float64`，`TINYINT(1)` → `int`。

```go
// backend/internal/model/progress_record.go
// append-only：无 biz_key，无软删字段
type ProgressRecord struct {
    ID          uint      `gorm:"primarykey;autoIncrement" json:"-"`
    SubItemKey  int64     `gorm:"not null" json:"subItemKey"`
    TeamKey     int64     `gorm:"not null" json:"teamKey"`
    AuthorKey   int64     `gorm:"not null" json:"authorKey"`
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
    ID         uint      `gorm:"primarykey;autoIncrement" json:"-"`
    ItemType   string    `gorm:"type:varchar(20);not null" json:"itemType"`
    ItemKey    int64     `gorm:"not null" json:"itemKey"`
    FromStatus string    `gorm:"type:varchar(20);not null" json:"fromStatus"`
    ToStatus   string    `gorm:"type:varchar(20);not null" json:"toStatus"`
    ChangedBy  int64     `gorm:"not null" json:"changedBy"`
    IsAuto     int       `gorm:"not null;default:0" json:"isAuto"`
    Remark     string    `gorm:"type:varchar(200)" json:"remark"`
    CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
}

func (StatusHistory) TableName() string { return "pmw_status_histories" }

// backend/internal/model/team_member.go
// 嵌入 BaseModel：成员可被移出团队（软删）
type TeamMember struct {
    BaseModel
    TeamKey  int64     `gorm:"not null" json:"teamKey"`
    UserKey  int64     `gorm:"not null" json:"userKey"`
    RoleKey  *int64    `json:"roleKey"`
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

### 3b. FindByBizKey 方法（每个 repo 接口）

`id` 不再对外暴露后，API 路径参数（`:itemId`）的值改为 `bizKey`（int64）。所有业务表 repo 接口新增 `FindByBizKey`，作为外部查询入口；`FindByID` 保留供内部关联使用。

受影响的 repo 接口：

| Repo 接口 | 新增方法 |
|-----------|---------|
| `MainItemRepo` | `FindByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error)` |
| `SubItemRepo` | `FindByBizKey(ctx context.Context, bizKey int64) (*model.SubItem, error)` |
| `TeamRepo` | `FindByBizKey(ctx context.Context, bizKey int64) (*model.Team, error)` |
| `ItemPoolRepo` | `FindByBizKey(ctx context.Context, bizKey int64) (*model.ItemPool, error)` |
| `UserRepo` | `FindByBizKey(ctx context.Context, bizKey int64) (*model.User, error)` |

实现模式（以 MainItem 为例）：

```go
func (r *mainItemRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error) {
    var item model.MainItem
    err := r.db.WithContext(ctx).Scopes(NotDeleted).
        Where("biz_key = ?", bizKey).First(&item).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrNotFound
        }
        return nil, err
    }
    return &item, nil
}
```

**Handler 层变更**：路径参数解析从 `uint` 改为 `int64`，调用 `FindByBizKey`：

```go
// 变更前
itemID, _ := strconv.ParseUint(c.Param("itemId"), 10, 64)
item, err := h.svc.GetByID(ctx, uint(itemID))

// 变更后
bizKey, _ := strconv.ParseInt(c.Param("itemId"), 10, 64)
item, err := h.svc.GetByBizKey(ctx, bizKey)
```

DELETE handler 需要两步：先用 `bizKey` 查出记录，再用内部 `id` 执行删除（service `Delete` 方法签名不变，仍接受 `uint` id）：

```go
// 变更后 DELETE handler
bizKey, _ := strconv.ParseInt(c.Param("itemId"), 10, 64)
item, err := h.svc.FindByBizKey(ctx, bizKey)  // step 1: resolve bizKey → record
if err != nil { apperrors.RespondError(c, err); return }
err = h.svc.Delete(ctx, item.ID)               // step 2: delete by internal id
```

### 3c. Repo Interface Definitions

`MainItemRepo` 完整接口（含新增 `FindByBizKey`）：

```go
// backend/internal/repository/main_item_repo.go
type MainItemRepo interface {
    Create(ctx context.Context, item *model.MainItem) error
    FindByID(ctx context.Context, id uint) (*model.MainItem, error)
    FindByIDs(ctx context.Context, ids []uint) (map[uint]*model.MainItem, error)
    FindByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error) // new
    Update(ctx context.Context, item *model.MainItem, fields map[string]interface{}) error
    List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error)
    ListByTeamAndStatus(ctx context.Context, teamID uint, status string) ([]model.MainItem, error)
    NextCode(ctx context.Context, teamID uint) (string, error)
    CountByTeam(ctx context.Context, teamID uint) (int64, error)
    ListNonArchivedByTeam(ctx context.Context, teamID uint) ([]model.MainItem, error)
}
```

`TeamMemberRepo` 完整接口（`SoftDelete` 为本次新增方法）：

```go
// backend/internal/repository/team_repo.go
type TeamMemberRepo interface {
    AddMember(ctx context.Context, member *model.TeamMember) error
    RemoveMember(ctx context.Context, teamID, userID uint) error
    FindMember(ctx context.Context, teamID, userID uint) (*model.TeamMember, error)
    ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error)
    CountMembers(ctx context.Context, teamID uint) (int64, error)
    UpdateMember(ctx context.Context, member *model.TeamMember) error
    FindPMMembers(ctx context.Context, teamIDs []uint) (map[uint]string, error)
    SoftDelete(ctx context.Context, id uint) error // new: soft-delete a team member record
}
```

`SubItemRepo` 完整接口（`Delete` 替换为 `SoftDelete`，新增 `FindByBizKey`）：

```go
// backend/internal/repository/sub_item_repo.go
type SubItemRepo interface {
    Create(ctx context.Context, item *model.SubItem) error
    FindByID(ctx context.Context, id uint) (*model.SubItem, error)
    FindByBizKey(ctx context.Context, bizKey int64) (*model.SubItem, error) // new
    Update(ctx context.Context, item *model.SubItem, fields map[string]interface{}) error
    SoftDelete(ctx context.Context, id uint) error                          // new: replaces Delete
    List(ctx context.Context, teamID uint, mainItemID uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error)
    ListByMainItem(ctx context.Context, mainItemID uint) ([]*model.SubItem, error)
    ListByTeam(ctx context.Context, teamID uint) ([]model.SubItem, error)
    NextSubCode(ctx context.Context, mainItemID uint) (string, error)
}
```

`TeamRepo` 完整接口（`Delete` 替换为 `SoftDelete`，新增 `FindByBizKey`）：

```go
// backend/internal/repository/team_repo.go
type TeamRepo interface {
    Create(ctx context.Context, team *model.Team) error
    FindByID(ctx context.Context, teamID uint) (*model.Team, error)
    FindByBizKey(ctx context.Context, bizKey int64) (*model.Team, error) // new
    List(ctx context.Context) ([]*model.Team, error)
    Update(ctx context.Context, team *model.Team) error
    SoftDelete(ctx context.Context, id uint) error                       // new: replaces Delete

    // TeamMember operations
    AddMember(ctx context.Context, member *model.TeamMember) error
    RemoveMember(ctx context.Context, teamID, userID uint) error
    FindMember(ctx context.Context, teamID, userID uint) (*model.TeamMember, error)
    ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error)
    CountMembers(ctx context.Context, teamID uint) (int64, error)
    UpdateMember(ctx context.Context, member *model.TeamMember) error
    FindPMMembers(ctx context.Context, teamIDs []uint) (map[uint]string, error)

    // Paginated list with optional search
    ListFiltered(ctx context.Context, search string, offset, limit int) ([]*model.Team, int64, error)

    // Admin operations
    ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error)
    FindTeamsByUserIDs(ctx context.Context, userIDs []uint) (map[uint][]dto.TeamSummary, error)
}
```

`ItemPoolRepo` 完整接口（新增 `FindByBizKey`）：

```go
// backend/internal/repository/item_pool_repo.go
type ItemPoolRepo interface {
    Create(ctx context.Context, item *model.ItemPool) error
    FindByID(ctx context.Context, id uint) (*model.ItemPool, error)
    FindByBizKey(ctx context.Context, bizKey int64) (*model.ItemPool, error) // new
    Update(ctx context.Context, item *model.ItemPool, fields map[string]interface{}) error
    List(ctx context.Context, teamID uint, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error)
}
```

`UserRepo` 完整接口（新增 `FindByBizKey`）：

```go
// backend/internal/repository/user_repo.go
type UserRepo interface {
    FindByID(ctx context.Context, id uint) (*model.User, error)
    FindByIDs(ctx context.Context, ids []uint) (map[uint]*model.User, error)
    FindByBizKey(ctx context.Context, bizKey int64) (*model.User, error) // new
    FindByUsername(ctx context.Context, username string) (*model.User, error)
    List(ctx context.Context) ([]*model.User, error)
    ListFiltered(ctx context.Context, search string, offset, limit int) ([]*model.User, int64, error)
    SearchAvailable(ctx context.Context, teamID uint, search string, limit int) ([]*model.User, error)
    Create(ctx context.Context, user *model.User) error
    Update(ctx context.Context, user *model.User) error
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
| `main_item_service.go` | `Create()` 赋值 `BizKey: snowflake.Generate()`；新增 `GetByBizKey(ctx, bizKey int64)` |
| `sub_item_service.go` | `Create()` 赋值 `BizKey`；`Delete()` 改调 `subItemRepo.SoftDelete()`；新增 `GetByBizKey()` |
| `team_service.go` | `Create()` 赋值 `BizKey`；`Delete()` 改调 `teamRepo.SoftDelete()`；`RemoveMember()` 改调 `teamMemberRepo.SoftDelete()`；新增 `GetByBizKey()` |
| `item_pool_service.go` | `Create()` 赋值 `BizKey`；新增 `GetByBizKey()` |
| `auth_service.go` | `Register()` 赋值 `BizKey` |

```go
// 示例：main_item_service.go
item := &model.MainItem{
    BizKey: snowflake.Generate(),
    // ... 其他字段
}
```

### 5b. Service Interface Definitions（含新增方法）

以下为受影响的四个 service 接口完整定义，新增方法标注 `// new`。

```go
// backend/internal/service/main_item_service.go
type MainItemService interface {
    Create(ctx context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error)
    Update(ctx context.Context, teamID, itemID uint, req dto.MainItemUpdateReq) error
    Archive(ctx context.Context, teamID, itemID uint) error
    List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error)
    Get(ctx context.Context, itemID uint) (*model.MainItem, error)
    GetByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error) // new
    RecalcCompletion(ctx context.Context, mainItemID uint) error
    ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) (*model.MainItem, error)
    AvailableTransitions(ctx context.Context, teamID, callerID, itemID uint) ([]string, error)
    EvaluateLinkage(ctx context.Context, mainItemID uint, changedBy uint) (*LinkageResult, error)
}
```

```go
// backend/internal/service/sub_item_service.go
type SubItemService interface {
    Create(ctx context.Context, teamID, callerID uint, req dto.SubItemCreateReq) (*model.SubItem, error)
    Update(ctx context.Context, teamID, itemID uint, req dto.SubItemUpdateReq) error
    ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) (*SubItemChangeResult, error)
    Delete(ctx context.Context, teamID, callerID, itemID uint) error // delegates to subItemRepo.SoftDelete internally
    Get(ctx context.Context, teamID, itemID uint) (*model.SubItem, error)
    GetByBizKey(ctx context.Context, bizKey int64) (*model.SubItem, error) // new
    List(ctx context.Context, teamID uint, mainItemID *uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error)
    Assign(ctx context.Context, teamID, pmID, itemID, assigneeID uint) error
    AvailableTransitions(ctx context.Context, teamID, subID uint) ([]string, error)
}
```

```go
// backend/internal/service/team_service.go
type TeamService interface {
    CreateTeam(ctx context.Context, creatorID uint, req dto.CreateTeamReq) (*model.Team, error)
    GetTeam(ctx context.Context, teamID uint) (*model.Team, error)
    GetTeamByBizKey(ctx context.Context, bizKey int64) (*model.Team, error) // new
    GetTeamDetail(ctx context.Context, teamID uint) (*dto.TeamDetailResp, error)
    ListTeams(ctx context.Context, callerID uint, isSuperAdmin bool, search string, page, pageSize int) ([]*dto.TeamListResp, int64, error)
    UpdateTeam(ctx context.Context, pmID, teamID uint, req dto.UpdateTeamReq) (*model.Team, error)
    InviteMember(ctx context.Context, pmID, teamID uint, req dto.InviteMemberReq) error
    RemoveMember(ctx context.Context, pmID, teamID, targetUserID uint) error
    TransferPM(ctx context.Context, currentPMID, teamID, newPMID uint) error
    DisbandTeam(ctx context.Context, callerID uint, teamID uint, confirmName string) error
    UpdateMemberRole(ctx context.Context, pmID, teamID, targetUserID, roleID uint) error
    ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error)
    SearchAvailableUsers(ctx context.Context, teamID uint, search string) ([]*dto.UserSearchDTO, error)
}
```

```go
// backend/internal/service/item_pool_service.go
type ItemPoolService interface {
    Submit(ctx context.Context, teamID, submitterID uint, req dto.SubmitItemPoolReq) (*model.ItemPool, error)
    Assign(ctx context.Context, teamID, pmID, poolItemID uint, req dto.AssignItemPoolReq) error
    ConvertToMain(ctx context.Context, teamID, pmID, poolItemID uint, req dto.ConvertToMainItemReq) (*model.MainItem, error)
    Reject(ctx context.Context, teamID, pmID, poolItemID uint, reason string) error
    List(ctx context.Context, teamID uint, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error)
    Get(ctx context.Context, teamID, poolItemID uint) (*model.ItemPool, error)
    GetByBizKey(ctx context.Context, teamID uint, bizKey int64) (*model.ItemPool, error) // new
}
```

```go
// backend/internal/service/auth_service.go
type AuthService interface {
    Register(ctx context.Context, req dto.RegisterReq) (*model.User, error) // modified: adds BizKey assignment
    Login(ctx context.Context, username, password string) (token string, user *model.User, err error)
    Logout(ctx context.Context, token string) error
    ParseToken(ctx context.Context, token string) (*appjwt.Claims, error)
}
```

### 6. 前端类型更新

以下为所有受影响类型变更后的完整 TypeScript 接口定义（`frontend/src/types/index.ts`）。

```typescript
// 变更：id 移除（json:"-"），bizKey 新增，status → userStatus，createdAt → createTime
export interface User {
  bizKey: string
  username: string
  displayName: string
  email?: string
  isSuperAdmin: boolean
  userStatus?: 'enabled' | 'disabled'
  teams?: TeamSummary[]
  createTime: string
}

// 变更：id 移除，bizKey 新增，createdAt → createTime，updatedAt → dbUpdateTime
// FK 字段：pmId → pmKey
// 字段重命名：name → teamName，description → teamDesc（避免 MySQL 保留字）
export interface Team {
  bizKey: string
  teamName: string
  code?: string
  teamDesc: string
  pmKey: string
  pmDisplayName?: string
  createTime: string
  dbUpdateTime: string
}

// 变更：id 移除，bizKey 新增，status → itemStatus，createdAt → createTime，updatedAt → dbUpdateTime
// FK 字段：teamId → teamKey，proposerId → proposerKey，assigneeId → assigneeKey
// 字段重命名：description → itemDesc，startDate → planStartDate（避免 MySQL 保留字）
export interface MainItem {
  bizKey: string
  teamKey: string
  code: string
  title: string
  itemDesc?: string
  priority: string
  proposerKey: string
  assigneeKey: string | null
  planStartDate: string | null
  expectedEndDate: string | null
  actualEndDate: string | null
  itemStatus: string
  statusName?: string
  completion: number
  createTime: string
  dbUpdateTime: string
}

// 变更：id 移除，bizKey 新增，status → itemStatus，createdAt → createTime，updatedAt → dbUpdateTime
// FK 字段：teamId → teamKey，mainItemId → mainItemKey，assigneeId → assigneeKey
// 字段重命名：description → itemDesc，startDate → planStartDate（避免 MySQL 保留字）
export interface SubItem {
  bizKey: string
  teamKey: string
  mainItemKey: string
  code: string
  title: string
  itemDesc: string
  priority: string
  assigneeKey: string | null
  planStartDate: string | null
  expectedEndDate: string | null
  actualEndDate: string | null
  itemStatus: string
  statusName?: string
  completion: number
  weight: number
  createTime: string
  dbUpdateTime: string
}

// 变更：id 移除，bizKey 新增，status → poolStatus，createdAt → createTime，updatedAt → dbUpdateTime
// FK 字段：teamId → teamKey，submitterId → submitterKey，assignedMainId → assignedMainKey，assignedSubId → assignedSubKey，assigneeId → assigneeKey，reviewerId → reviewerKey
export interface ItemPool {
  bizKey: string
  teamKey: string
  title: string
  background: string
  expectedOutput: string
  submitterKey: string
  submitterName?: string
  poolStatus: string
  assignedMainKey: string | null
  assignedSubKey: string | null
  assignedMainCode: string
  assignedMainTitle: string
  assigneeKey: string | null
  rejectReason: string
  reviewedAt: string | null
  reviewerKey: string | null
  createTime: string
  dbUpdateTime: string
}

// 变更：id 移除（append-only 表无 biz_key），createdAt → createTime
// FK 字段：subItemId → subItemKey，teamId → teamKey，authorId → authorKey
export interface ProgressRecord {
  subItemKey: string
  teamKey: string
  authorKey: string
  authorName?: string
  completion: number
  achievement: string
  blocker: string
  lesson: string
  isPMCorrect: boolean
  createTime: string
}

// 变更：createdAt → createTime（Role 无软删，保留 id）
// 字段重命名：name → roleName，description → roleDesc（避免 MySQL 保留字）
export interface Role {
  id: number
  roleName: string
  roleDesc: string
  isPreset: boolean
  permissionCount: number
  memberCount: number
  createTime: string
}
```

> **FK 字段策略**：所有外键字段从存储内部 `id` 改为存储 `biz_key`，列名从 `*_id` 改为 `*_key`（如 `main_item_key`、`sub_item_key`）。原因：`id` 不再对外暴露后，关联关系也必须通过 `biz_key` 建立，确保整个系统只使用 `biz_key` 作为对外标识符。

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
    team_name       VARCHAR(100)    NOT NULL,
    team_desc       VARCHAR(500),
    pm_key          BIGINT          NOT NULL,
    code            VARCHAR(6)      NOT NULL DEFAULT '',
    item_seq        INT             NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_teams_code_deleted (code, deleted_flag, deleted_time),
    KEY idx_teams_pm_key (pm_key),
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
    team_key        BIGINT          NOT NULL,
    user_key        BIGINT          NOT NULL,
    role_key        BIGINT,
    joined_at       DATETIME        NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_team_user_deleted (team_key, user_key, deleted_flag, deleted_time),
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
    team_key          BIGINT          NOT NULL,
    code              VARCHAR(12)     NOT NULL,
    title             VARCHAR(100)    NOT NULL,
    item_desc         VARCHAR(2000)   NOT NULL DEFAULT '',
    priority          VARCHAR(5)      NOT NULL,
    proposer_key      BIGINT          NOT NULL,
    assignee_key      BIGINT,
    plan_start_date   DATETIME,
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
    UNIQUE KEY uk_main_items_team_code_deleted (team_key, code, deleted_flag, deleted_time),
    KEY idx_main_items_team_key (team_key),
    KEY idx_main_items_assignee_key (assignee_key),
    KEY idx_main_items_expected_end_date (expected_end_date),
    KEY idx_main_items_team_status (team_key, item_status),
    KEY idx_main_items_team_priority (team_key, priority),
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
    team_key          BIGINT          NOT NULL,
    main_item_key     BIGINT          NOT NULL,
    code              VARCHAR(15)     NOT NULL DEFAULT '',
    title             VARCHAR(100)    NOT NULL,
    item_desc         VARCHAR(2000),
    priority          VARCHAR(5)      NOT NULL,
    assignee_key      BIGINT,
    plan_start_date   DATETIME,
    expected_end_date DATETIME,
    actual_end_date   DATETIME,
    item_status       VARCHAR(20)     NOT NULL DEFAULT '待开始',
    completion        DECIMAL(5,2)    NOT NULL DEFAULT 0.00,
    is_key_item       TINYINT(1)      NOT NULL DEFAULT 0,
    delay_count       INT             NOT NULL DEFAULT 0,
    weight            DECIMAL(5,2)    NOT NULL DEFAULT 1.00,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    UNIQUE KEY uk_sub_items_main_code (main_item_key, code),
    KEY idx_sub_items_main_item_key (main_item_key),
    KEY idx_sub_items_team_key (team_key),
    KEY idx_sub_items_assignee_key (assignee_key),
    KEY idx_sub_items_team_status (team_key, item_status),
    KEY idx_sub_items_team_priority (team_key, priority),
    KEY idx_sub_items_expected_end_date (expected_end_date),
    KEY idx_sub_items_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子事项表';

-- pmw_item_pools
CREATE TABLE pmw_item_pools (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    biz_key           BIGINT          NOT NULL,
    create_time       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_update_time    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_flag      TINYINT(1)      NOT NULL DEFAULT 0,
    deleted_time      DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
    team_key          BIGINT          NOT NULL,
    title             VARCHAR(100)    NOT NULL,
    background        VARCHAR(2000),
    expected_output   VARCHAR(1000),
    submitter_key     BIGINT          NOT NULL,
    pool_status       VARCHAR(20)     NOT NULL DEFAULT '待分配',
    assigned_main_key BIGINT,
    assigned_sub_key  BIGINT,
    assignee_key      BIGINT,
    reject_reason     VARCHAR(200),
    reviewed_at       DATETIME,
    reviewer_key      BIGINT,
    PRIMARY KEY (id),
    UNIQUE KEY uk_biz_key (biz_key),
    KEY idx_item_pools_team_key (team_key),
    KEY idx_item_pools_team_status (team_key, pool_status),
    KEY idx_item_pools_submitter_key (submitter_key),
    KEY idx_item_pools_deleted_flag (deleted_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='需求池表';

-- pmw_progress_records（append-only：无软删，无 biz_key）
CREATE TABLE pmw_progress_records (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    sub_item_key    BIGINT          NOT NULL,
    team_key        BIGINT          NOT NULL,
    author_key      BIGINT          NOT NULL,
    completion      DECIMAL(5,2)    NOT NULL,
    achievement     VARCHAR(1000),
    blocker         VARCHAR(1000),
    lesson          VARCHAR(1000),
    is_pm_correct   TINYINT(1)      NOT NULL DEFAULT 0,
    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_progress_records_sub_item_key (sub_item_key),
    KEY idx_progress_records_sub_item_created (sub_item_key, create_time),
    KEY idx_progress_records_team_key (team_key),
    KEY idx_progress_records_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='进度记录表（追加写入）';

-- pmw_status_histories（append-only：无软删，无 biz_key）
CREATE TABLE pmw_status_histories (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    item_type    VARCHAR(20)     NOT NULL,
    item_key     BIGINT          NOT NULL,
    from_status  VARCHAR(20)     NOT NULL,
    to_status    VARCHAR(20)     NOT NULL,
    changed_by   BIGINT          NOT NULL,
    is_auto      TINYINT(1)      NOT NULL DEFAULT 0,
    remark       VARCHAR(200),
    create_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_item (item_type, item_key)
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
| ERR_DUPLICATE_BIZ_KEY | ErrDuplicateBizKey | biz_key 唯一键冲突（雪花碰撞或并发重复提交） | 409 |

以下 sentinel 已存在于 `backend/internal/pkg/errors/errors.go`：

```go
// backend/internal/pkg/errors/errors.go (existing)
var ErrNotFound  = &AppError{Code: "NOT_FOUND",        Status: 404, Message: "resource not found"}
var ErrValidation = &AppError{Code: "VALIDATION_ERROR", Status: 400, Message: "request validation failed"}
```

新增 `ErrDuplicateBizKey`：

```go
// backend/internal/pkg/errors/errors.go (new)
var ErrDuplicateBizKey = &AppError{Code: "DUPLICATE_BIZ_KEY", Status: 409, Message: "biz_key uniqueness violation"}
```

**Service 层处理**：在 `Create()` 方法中捕获 MySQL Error 1062（duplicate entry），转换为 `ErrDuplicateBizKey`：

```go
func (s *mainItemService) Create(ctx context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error) {
    item := &model.MainItem{
        BizKey: snowflake.Generate(),
        // ... 其他字段
    }
    err := s.repo.Create(ctx, item)
    if err != nil {
        if isMySQLDuplicateError(err) {
            return nil, apperrors.ErrDuplicateBizKey
        }
        return nil, err
    }
    return item, nil
}

func isMySQLDuplicateError(err error) bool {
    var mysqlErr *mysql.MySQLError
    if errors.As(err, &mysqlErr) {
        return mysqlErr.Number == 1062
    }
    return false
}
```

`FindByBizKey` 和 `SoftDelete` 实现直接返回 `apperrors.ErrNotFound`；handler 通过 `apperrors.RespondError` 统一响应。

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
| auto-increment `id` 暴露，攻击者可顺序枚举资源 | `json:"-"` 阻止 `id` 序列化，`id` 不出现在任何 API 响应中，顺序枚举攻击面消除 |
| `biz_key` 暴露泄露雪花时间戳和 worker-id | `biz_key` 通过 `json:"bizKey"` 对外暴露；雪花 ID 非顺序，枚举难度高；但仍编码了创建时间和 worker-id，日志中禁止出现原始雪花值（见下） |
| biz_key 通过日志或错误响应间接泄露创建时间/机器标识 | logging middleware 不得打印完整 model struct；error message 不得包含 biz_key 原始值 |
| 多节点部署时 worker-id 碰撞导致 biz_key 重复 | **部署约束**：本设计假设单节点部署，worker-id 硬编码为 1。多节点部署需引入 worker-id 协调机制（如 etcd/Redis 分布式锁），超出本次迭代范围。若未来扩展为多节点，必须确保每个节点分配唯一 worker-id（范围 0-1023），否则会产生 biz_key 碰撞。 |

> 注：worker-id=1（硬编码单机），雪花值编码了创建时间和 worker-id。biz_key 通过 `json:"bizKey"` 对外暴露，客户端可见；但日志和 error message 中不得打印原始雪花值，防止在非预期渠道泄露机器标识。
>
> **多节点部署约束**：当前设计仅支持单节点部署。如需水平扩展，必须先实现 worker-id 协调机制（推荐方案：启动时从 etcd/Redis 原子分配唯一 worker-id，节点下线时释放）。未实现协调前，禁止多节点部署，否则会导致 biz_key 碰撞。

### Mitigations

- repo 接口不暴露硬删除方法，从接口层面杜绝误操作
- `SoftDelete` 实现加 `deleted_flag = 0` 条件，防止重复软删并确保幂等性
- logging middleware 记录请求/响应时只记录 `bizKey`，不打印完整 model struct

### Logging Implementation

- Use structured logging (e.g., zap) with field filtering
- Implement custom `MarshalJSON()` for BaseModel that omits internal fields:
  ```go
  func (b BaseModel) MarshalJSON() ([]byte, error) {
      return json.Marshal(struct {
          BizKey int64 `json:"bizKey"`
      }{BizKey: b.BizKey})
  }
  ```
- Or use logging middleware that explicitly extracts `bizKey` from context, never logs raw model

## PRD Coverage Map

| PRD AC | Design Component | Interface / Model |
|--------|-----------------|-------------------|
| schema.sql 在 MySQL 8.0 无报错执行 | schema.sql 重写 | DDL 全量替换 |
| 所有表含 create_time/db_update_time/deleted_flag/deleted_time/biz_key | BaseModel 替换 | `model/base.go` |
| biz_key 对外暴露（json:"bizKey"），id 不对外暴露（json:"-"） | BaseModel json tag | `model/base.go` |
| 无 TEXT 字段 | schema.sql + model GORM tag | VARCHAR(1000/2000) |
| 无 status 关键字 | model 字段重命名 | UserStatus/ItemStatus/PoolStatus |
| 索引符合 idx_/uk_ 规范 | schema.sql | 索引命名全量检查 |
| 每张表有 COMMENT | schema.sql | 表级 COMMENT |
| 无外键约束（DDL 层面） | schema.sql | 所有表仅保留索引，无 FOREIGN KEY 约束 |
| model/base.go 不嵌入 gorm.Model | BaseModel 替换 | `model/base.go` |
| repo 封闭 SoftDelete，无 db.Delete 外部调用 | SoftDelete 接口 | `TeamRepo`/`SubItemRepo`/`TeamMemberRepo` |
| 所有 repo 查询通过 NotDeleted scope | NotDeleted scope | `scopes.go` |
| bizKey 对外暴露，id 不对外暴露 | BaseModel json tag | `BizKey json:"bizKey"`, `ID json:"-"` |
| API 路径参数改用 bizKey，后端通过 FindByBizKey 定位记录 | FindByBizKey 接口 | 所有业务 repo + handler 层 |
| FK 字段从 id 改为 biz_key，列名 *_id → *_key | schema.sql + model structs | 所有 FK 字段（team_key, user_key, main_item_key 等） |
| FK 数据迁移：现有数据 *_id 列值需回填为 biz_key | Migration Strategy（见下） | 两阶段迁移：ADD COLUMN → backfill → DROP COLUMN |
| go test ./... 全部通过 | 各层单元测试 | 见 Testing Strategy |
| 前端字段引用更新，npm test 通过 | types/index.ts + pages | 字段名全量替换 |
| E2E 测试通过 | E2E 断言更新 | `__tests__/e2e/*.spec.ts` |

## Migration Strategy

### FK Data Migration: id → biz_key

现有生产数据中，所有 FK 字段（`team_id`、`user_id`、`main_item_id` 等）存储的是内部 `id` 值。本次变更需将这些字段改为存储 `biz_key`，列名从 `*_id` 改为 `*_key`。

**两阶段迁移方案**：

#### Phase 1: Add New Columns & Backfill

```sql
-- 1. 为每个 FK 字段添加新的 *_key 列
ALTER TABLE pmw_teams ADD COLUMN pm_key BIGINT;
ALTER TABLE pmw_team_members ADD COLUMN team_key BIGINT, ADD COLUMN user_key BIGINT, ADD COLUMN role_key BIGINT;
ALTER TABLE pmw_main_items ADD COLUMN team_key BIGINT, ADD COLUMN proposer_key BIGINT, ADD COLUMN assignee_key BIGINT;
ALTER TABLE pmw_sub_items ADD COLUMN team_key BIGINT, ADD COLUMN main_item_key BIGINT, ADD COLUMN assignee_key BIGINT;
ALTER TABLE pmw_item_pools ADD COLUMN team_key BIGINT, ADD COLUMN submitter_key BIGINT, ADD COLUMN assigned_main_key BIGINT, ADD COLUMN assigned_sub_key BIGINT, ADD COLUMN assignee_key BIGINT, ADD COLUMN reviewer_key BIGINT;
ALTER TABLE pmw_progress_records ADD COLUMN sub_item_key BIGINT, ADD COLUMN team_key BIGINT, ADD COLUMN author_key BIGINT;
ALTER TABLE pmw_status_histories ADD COLUMN item_key BIGINT, ADD COLUMN changed_by BIGINT;

-- 2. 通过 JOIN 回填 biz_key 值
UPDATE pmw_teams t JOIN pmw_users u ON t.pm_id = u.id SET t.pm_key = u.biz_key;
UPDATE pmw_team_members tm 
  JOIN pmw_teams t ON tm.team_id = t.id 
  JOIN pmw_users u ON tm.user_id = u.id 
  SET tm.team_key = t.biz_key, tm.user_key = u.biz_key;
UPDATE pmw_team_members tm LEFT JOIN pmw_roles r ON tm.role_id = r.id SET tm.role_key = r.biz_key;
UPDATE pmw_main_items mi 
  JOIN pmw_teams t ON mi.team_id = t.id 
  JOIN pmw_users proposer ON mi.proposer_id = proposer.id 
  LEFT JOIN pmw_users assignee ON mi.assignee_id = assignee.id 
  SET mi.team_key = t.biz_key, mi.proposer_key = proposer.biz_key, mi.assignee_key = assignee.biz_key;
-- 类似模式适用于其他表...
```

#### Phase 2: Drop Old Columns & Add Constraints

```sql
-- 3. 删除旧的 *_id 列
ALTER TABLE pmw_teams DROP COLUMN pm_id;
ALTER TABLE pmw_team_members DROP COLUMN team_id, DROP COLUMN user_id, DROP COLUMN role_id;
-- 其他表类似...

-- 4. 添加 NOT NULL 约束和索引（回填完成后）
ALTER TABLE pmw_teams MODIFY pm_key BIGINT NOT NULL;
ALTER TABLE pmw_main_items MODIFY team_key BIGINT NOT NULL, MODIFY proposer_key BIGINT NOT NULL;
-- 其他表类似...
```

**迁移执行顺序**：

1. 先迁移被引用表（users、teams、roles）的 `biz_key` 生成
2. 再迁移引用表（team_members、main_items、sub_items 等）的 FK 字段
3. 最后迁移 append-only 表（progress_records、status_histories）

**回滚方案**：保留旧 `*_id` 列直到 Phase 2 确认无误后再删除，如需回滚可从 `*_key` 反向回填。

## Open Questions

- [x] biz_key 生成策略：雪花算法，worker-id=1（单机）
- [x] TEXT 字段处理：改为 VARCHAR，超限升级为独立 detail 表
- [x] 软删接口设计：repo 封闭 SoftDelete，禁止外部 db.Delete
- [x] `progress_records` 和 `status_histories` 是否需要 biz_key？（结论：不需要，append-only 表无业务关联需求）
- [x] FK 字段存储策略：从存储内部 `id` 改为存储 `biz_key`，列名 `*_id` → `*_key`（原因：id 不对外暴露后，关联关系也必须通过 biz_key 建立）
- [x] 外键约束：DDL 层面不创建 FOREIGN KEY 约束，仅保留索引（原因：避免跨表级联依赖，便于数据迁移和分库分表）
- [x] FK 数据迁移：两阶段迁移（ADD COLUMN → backfill → DROP COLUMN），保留旧列直到确认无误

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

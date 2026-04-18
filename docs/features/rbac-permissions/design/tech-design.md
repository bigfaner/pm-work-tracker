---
created: 2026-04-18
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: RBAC 权限体系

## Overview

在现有 Go/Gin 后端中引入标准 RBAC 权限体系。核心变更：

- 新增 `roles` + `role_permissions` 两张表，替换硬编码布尔标记
- 权限码以 Go 常量定义，角色-权限绑定存数据库
- 中间件从角色名检查（`RequireRole`/`RequireTeamRole`）切换到权限码检查（`RequirePermission`）
- JWT 精简为 `user_id` + `username` + `iat`，权限由后端实时查询
- 移除 `users.is_super_admin`、`users.can_create_team`、`teams.pm_id`、`team_members.role`

## Architecture

### Layer Placement

```
Handler  →  Service  →  Repository  →  Model/DB
   ↓           ↓
Middleware   PermissionCache (in-memory)
```

新增组件分布在现有分层中：

| 组件 | 层 | 说明 |
|------|----|------|
| `pkg/permissions` | pkg | 权限码常量定义 + 注册表 |
| `model/role.go` | model | Role、RolePermission 模型 |
| `repository/role_repo.go` | repository | 角色 CRUD + 权限绑定接口 |
| `service/role_service.go` | service | 角色管理业务逻辑 |
| `handler/role_handler.go` | handler | 角色 CRUD API |
| `middleware/rbac.go` | middleware | 重写为 RequirePermission |
| `pkg/permcache` | pkg | 权限缓存（启动加载，变更刷新） |

### Component Diagram

```
┌──────────────┐     ┌──────────────┐
│  role_handler│────>│ role_service │
└──────────────┘     └──────┬───────┘
                            │
                     ┌──────▼───────┐
                     │  role_repo   │
                     └──────┬───────┘
                            │
┌──────────────┐     ┌──────▼───────┐     ┌──────────────┐
│ RequirePerm  │────>│ permcache    │<────│ role_perms   │
│  middleware  │     │ (in-memory)  │     │    table     │
└──────────────┘     └──────────────┘     └──────────────┘
```

### Dependencies

无新增外部依赖。权限缓存使用 `sync.RWMutex` + map 实现，不引入缓存库。

## Interfaces

### Permission Registry (`internal/pkg/permissions/permissions.go`)

```go
type PermissionInfo struct {
    Code        string
    Resource    string
    Action      string
    Description string
}

// 权限码常量
const (
    TeamCreate      = "team:create"
    TeamRead        = "team:read"
    TeamUpdate      = "team:update"
    TeamDelete      = "team:delete"
    TeamInvite      = "team:invite"
    TeamRemove      = "team:remove"
    TeamTransfer    = "team:transfer"

    MainItemCreate  = "main_item:create"
    MainItemRead    = "main_item:read"
    MainItemUpdate  = "main_item:update"
    MainItemArchive = "main_item:archive"

    SubItemCreate      = "sub_item:create"
    SubItemRead        = "sub_item:read"
    SubItemUpdate      = "sub_item:update"
    SubItemAssign      = "sub_item:assign"
    SubItemChangeStatus = "sub_item:change_status"

    ProgressCreate = "progress:create"
    ProgressRead   = "progress:read"
    ProgressUpdate = "progress:update"

    ItemPoolSubmit = "item_pool:submit"
    ItemPoolReview = "item_pool:review"

    ViewWeekly = "view:weekly"
    ViewGantt  = "view:gantt"
    ViewTable  = "view:table"

    ReportExport = "report:export"

    UserRead      = "user:read"
    UserUpdate    = "user:update"
    UserManageRole = "user:manage_role"
)

var AllPermissions []PermissionInfo  // 启动时初始化，27 个权限码
var AllCodes map[string]bool         // 快速查找
```

### Role Repository (`internal/repository/role_repo.go`)

```go
type RoleRepo interface {
    Create(ctx context.Context, role *model.Role, permissionCodes []string) error
    FindByID(ctx context.Context, id uint) (*model.Role, error)
    List(ctx context.Context, params ListRolesParams) ([]*model.Role, int64, error)
    Update(ctx context.Context, role *model.Role, permissionCodes []string) error
    Delete(ctx context.Context, id uint) error
    GetPermissions(ctx context.Context, roleID uint) ([]string, error)
    CountMembers(ctx context.Context, roleID uint) (int64, error)
    FindTeamMemberRole(ctx context.Context, teamID, userID uint) (uint, error)
}

type ListRolesParams struct {
    Search   string
    Preset   *bool  // nil=全部, true=预置, false=自定义
    Page     int
    PageSize int
}
```

### Permission Cache (`internal/pkg/permcache/cache.go`)

```go
type Cache interface {
    GetPermissions(roleID uint) map[string]bool
    HasPermission(roleID uint, code string) bool
    Refresh(ctx context.Context) error     // 全量刷新
    Invalidate(roleID uint)                // 单角色失效（触发重新加载）
}
```

### Role Service (`internal/service/role_service.go`)

```go
type RoleService interface {
    ListRoles(ctx context.Context, params repository.ListRolesParams) ([]*model.Role, int64, error)
    GetRole(ctx context.Context, id uint) (*model.RoleDetail, error)
    CreateRole(ctx context.Context, req CreateRoleReq) (*model.Role, error)
    UpdateRole(ctx context.Context, id uint, req UpdateRoleReq) (*model.Role, error)
    DeleteRole(ctx context.Context, id uint) error
    ListPermissions(ctx context.Context) []permissions.PermissionInfo
    GetUserPermissions(ctx context.Context, userID uint) (*UserPermissionsResp, error)
}

type UserPermissionsResp struct {
    IsSuperAdmin    bool              `json:"is_superadmin"`
    TeamRoles       map[uint]string   `json:"team_roles"`        // teamID → role name
    TeamPermissions map[uint][]string `json:"team_permissions"`  // teamID → codes
}

type CreateRoleReq struct {
    Name            string   `json:"name" binding:"required,min=2,max=50"`
    Description     string   `json:"description" binding:"max=200"`
    PermissionCodes []string `json:"permission_codes" binding:"required,min=1"`
}

type UpdateRoleReq struct {
    Name            *string  `json:"name" binding:"omitempty,min=2,max=50"`
    Description     *string  `json:"description" binding:"omitempty,max=200"`
    PermissionCodes []string `json:"permission_codes" binding:"omitempty,min=1"`
}
```

> `UpdateRoleReq` 字段均为指针类型，区分"未提交"与"清空"。`PermissionCodes` 为 nil 时不更新权限，非 nil 时替换整个权限列表。

### Middleware (`internal/middleware/rbac.go`)

```go
// RequireSuperAdmin 检查用户是否为超级管理员
func RequireSuperAdmin(cache permcache.Cache, userRepo repository.UserRepo) gin.HandlerFunc

// RequirePermission 检查用户在当前团队是否拥有指定权限码
// 必须在 TeamScopeMiddleware 之后使用
func RequirePermission(code string, cache permcache.Cache) gin.HandlerFunc

// CanCreateTeam 检查用户是否有权创建团队（全局权限）
// 超级管理员 bypass；其他用户检查是否在任意团队中拥有 team:create 权限
func CanCreateTeam(cache permcache.Cache, teamRepo repository.TeamRepo) gin.HandlerFunc
```

## Data Models

### New: Role (`internal/model/role.go`)

```go
type Role struct {
    ID          uint   `gorm:"primaryKey"`
    Name        string `gorm:"uniqueIndex;size:50;not null"`
    Description string `gorm:"size:200"`
    IsPreset    bool   `gorm:"default:false;not null"`
    IsGlobal    bool   `gorm:"default:false;not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

Table: `roles`

### New: RolePermission (`internal/model/role.go`)

```go
type RolePermission struct {
    ID             uint   `gorm:"primaryKey"`
    RoleID         uint   `gorm:"not null;uniqueIndex:idx_role_perm"`
    PermissionCode string `gorm:"size:50;not null;uniqueIndex:idx_role_perm"`
}
```

Table: `role_permissions`，复合唯一索引 `(role_id, permission_code)`

### Modified: User

```go
type User struct {
    gorm.Model
    Username     string `gorm:"uniqueIndex;size:64;not null"`
    DisplayName  string `gorm:"size:64"`
    PasswordHash string `gorm:"size:255;-:json"`
    GlobalRoleID *uint  `gorm:"index"`                    // NEW: FK → roles.id
    GlobalRole   *Role  `gorm:"foreignKey:GlobalRoleID"`  // NEW
}
```

移除: `IsSuperAdmin bool`, `CanCreateTeam bool`

### Modified: TeamMember

```go
type TeamMember struct {
    ID        uint      `gorm:"primaryKey"`
    TeamID    uint      `gorm:"not null;uniqueIndex:idx_team_user"`
    UserID    uint      `gorm:"not null;uniqueIndex:idx_team_user"`
    RoleID    uint      `gorm:"not null;index"`           // NEW: FK → roles.id
    Role      *Role     `gorm:"foreignKey:RoleID"`        // NEW
    JoinedAt  time.Time `gorm:"autoCreateTime"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

移除: `Role string` (varchar 20)

### Modified: Team

```go
type Team struct {
    gorm.Model
    Name        string `gorm:"size:100;not null"`
    Description string `gorm:"size:500"`
}
```

移除: `PmID uint`

### ER Diagram

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────┐
│    users     │     │   team_members   │     │    teams     │
├─────────────┤     ├──────────────────┤     ├─────────────┤
│ id          │     │ id               │     │ id          │
│ username    │     │ team_id    ──────┼────>│ id          │
│ display_name│     │ user_id    ──────┼──┐  │ name        │
│ password_h. │     │ role_id    ──┐   │  │  │ description │
│ global_role.┼──┐  │ joined_at      │  │  └─────────────┘
└─────────────┘  │  │ created_at     │  │
                 │  │ updated_at     │  │
                 │  └──────────────────┘  │
                 │                        │  ┌─────────────┐
                 │  ┌─────────────┐       │  │    users     │
                 │  │    roles     │       │  └─────────────┘
                 │  ├─────────────┤       │
                 └─>│ id          │       │
                    │ name        │<──────┘
                    │ description │
                    │ is_preset   │
                    │ is_global   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────────┐
                    │ role_permissions│
                    ├─────────────────┤
                    │ id              │
                    │ role_id         │
                    │ permission_code │
                    └─────────────────┘
```

## Middleware Route Mapping

每个路由的中间件从角色名检查切换到权限码检查：

### 团队外路由

| Route | Current | New |
|-------|---------|-----|
| `POST /teams` | Handler 内检查 `CanCreateTeam` | `CanCreateTeam` 中间件 |
| `GET /teams` | AuthMiddleware | AuthMiddleware（不变） |

### 团队内路由（TeamScopeMiddleware 之后）

| Route | Current | New |
|-------|---------|-----|
| `GET /teams/:teamId` | any member | `RequirePermission(team:read)` |
| `PUT /teams/:teamId` | RequireTeamRole("pm") | `RequirePermission(team:update)` |
| `DELETE /teams/:teamId` | RequireTeamRole("pm") | `RequirePermission(team:delete)` |
| `GET /teams/:teamId/members` | any member | `RequirePermission(team:read)` |
| `POST /teams/:teamId/members` | RequireTeamRole("pm") | `RequirePermission(team:invite)` |
| `DELETE /teams/:teamId/members/:userId` | RequireTeamRole("pm") | `RequirePermission(team:remove)` |
| `PUT /teams/:teamId/pm` | RequireTeamRole("pm") | `RequirePermission(team:transfer)` |
| `POST /main-items` | any member | `RequirePermission(main_item:create)` |
| `GET /main-items` | any member | `RequirePermission(main_item:read)` |
| `GET /main-items/:itemId` | any member | `RequirePermission(main_item:read)` |
| `PUT /main-items/:itemId` | any member | `RequirePermission(main_item:update)` |
| `POST /main-items/:itemId/archive` | any member | `RequirePermission(main_item:archive)` |
| `POST /main-items/:itemId/sub-items` | any member | `RequirePermission(sub_item:create)` |
| `GET /main-items/:itemId/sub-items` | any member | `RequirePermission(sub_item:read)` |
| `GET /sub-items/:subId` | any member | `RequirePermission(sub_item:read)` |
| `PUT /sub-items/:subId` | any member | `RequirePermission(sub_item:update)` |
| `PUT /sub-items/:subId/status` | any member | `RequirePermission(sub_item:change_status)` |
| `PUT /sub-items/:subId/assignee` | any member | `RequirePermission(sub_item:assign)` |
| `POST /sub-items/:subId/progress` | any member | `RequirePermission(progress:create)` |
| `GET /sub-items/:subId/progress` | any member | `RequirePermission(progress:read)` |
| `PATCH /progress/:recordId/completion` | any member | `RequirePermission(progress:update)` |
| `POST /item-pool` | any member | `RequirePermission(item_pool:submit)` |
| `GET /item-pool` | any member | `RequirePermission(item_pool:submit)` |
| `GET /item-pool/:poolId` | any member | `RequirePermission(item_pool:submit)` |
| `POST /item-pool/:poolId/assign` | RequireTeamRole("pm") | `RequirePermission(item_pool:review)` |
| `POST /item-pool/:poolId/reject` | RequireTeamRole("pm") | `RequirePermission(item_pool:review)` |
| `GET /views/weekly` | any member | `RequirePermission(view:weekly)` |
| `GET /views/gantt` | any member | `RequirePermission(view:gantt)` |
| `GET /views/table` | any member | `RequirePermission(view:table)` |
| `GET /views/table/export` | any member | `RequirePermission(view:table)` |
| `GET /reports/weekly/preview` | any member | `RequirePermission(report:export)` |
| `GET /reports/weekly/export` | any member | `RequirePermission(report:export)` |

### Admin 路由

| Route | Current | New |
|-------|---------|-----|
| `GET /admin/users` | RequireRole("superadmin") | `RequireSuperAdmin()` |
| `GET /admin/teams` | RequireRole("superadmin") | `RequireSuperAdmin()` |
| `PUT /admin/users/:userId/can-create-team` | RequireRole("superadmin") | **删除**（由角色管理替代） |
| `GET /admin/roles` | — | `RequireSuperAdmin()` |
| `POST /admin/roles` | — | `RequireSuperAdmin()` |
| `GET /admin/roles/:roleId` | — | `RequireSuperAdmin()` |
| `PUT /admin/roles/:roleId` | — | `RequireSuperAdmin()` |
| `DELETE /admin/roles/:roleId` | — | `RequireSuperAdmin()` |
| `GET /admin/permissions` | — | `RequireSuperAdmin()` |

### 新增用户路由

| Route | Auth | 说明 |
|-------|------|------|
| `GET /me/permissions` | AuthMiddleware | 获取当前用户权限 |
| `PUT /teams/:teamId/members/:userId/role` | RequirePermission(team:invite) | 变更成员角色 |

## JWT Changes

### Current Claims

```go
type Claims struct {
    UserID uint   `json:"user_id"`
    Role   string `json:"role"`      // "superadmin" or "member"
    jwtv5.RegisteredClaims              // contains exp, iat
}
```

### New Claims

```go
type Claims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    jwtv5.RegisteredClaims              // contains iat (no exp)
}
```

变更点：
- 移除 `Role` 字段（权限由后端实时查询）
- 新增 `Username` 字段
- 移除 `exp`，token 有效期通过后端检查 `time.Now() - iat` 是否超过配置的 `MaxTokenAge`
- `jwt.Sign()` 不再设置 `ExpiresAt`，`jwt.Verify()` 不再检查 `ExpiresAt`

### SuperAdmin 判断

当前：从 JWT `Role == "superadmin"` 判断
新方案：中间件从 DB 查询 `user.GlobalRoleID` 是否指向 superadmin 角色。通过 `RequireSuperAdmin` 中间件在请求级别查询并缓存到 context。

## Permission Cache

### 设计

```go
type cache struct {
    mu     sync.RWMutex
    store  map[uint]map[string]bool  // roleID → permission code set
    repo   repository.RoleRepo
}
```

### 生命周期

| 事件 | 行为 |
|------|------|
| 启动 | `Refresh()` 全量加载所有角色权限到内存 |
| 角色创建/更新/删除 | `Invalidate(roleID)` 触发单角色重新加载 |
| 查询权限 | `GetPermissions(roleID)` 读内存，无 IO |

### 一致性

角色变更后，缓存立即刷新。已在进行中的请求使用旧权限集（可接受的微秒级窗口）。前端通过 `/api/me/permissions` 获取最新权限。

## Data Migration

### 迁移步骤（单事务）

```sql
BEGIN;

-- 1. 创建 roles 表
CREATE TABLE roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(200),
    is_preset BOOLEAN NOT NULL DEFAULT FALSE,
    is_global BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME,
    updated_at DATETIME
);

-- 2. 创建 role_permissions 表
CREATE TABLE role_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_code VARCHAR(50) NOT NULL,
    UNIQUE(role_id, permission_code)
);

-- 3. 插入预置角色
INSERT INTO roles (name, description, is_preset, is_global, created_at, updated_at) VALUES
    ('superadmin', '系统最高权限角色', 1, 1, datetime('now'), datetime('now')),
    ('pm', '团队管理权限', 1, 0, datetime('now'), datetime('now')),
    ('member', '基础操作权限', 1, 0, datetime('now'), datetime('now'));

-- 4. 插入预置权限（pm 和 member 的权限码，superadmin 不插入记录，代码中 bypass）
-- pm: 21 个权限码（除 user:update, user:manage_role 外全部）
-- member: 11 个权限码

-- 5. users 表新增 global_role_id
ALTER TABLE users ADD COLUMN global_role_id INTEGER REFERENCES roles(id);

-- 6. 迁移 superadmin 用户
UPDATE users SET global_role_id = (SELECT id FROM roles WHERE name = 'superadmin')
    WHERE is_super_admin = 1;

-- 7. team_members 表新增 role_id
ALTER TABLE team_members ADD COLUMN role_id INTEGER REFERENCES roles(id);

-- 8. 迁移团队成员角色
UPDATE team_members SET role_id = (SELECT id FROM roles WHERE name = 'pm')
    WHERE role = 'pm';
UPDATE team_members SET role_id = (SELECT id FROM roles WHERE name = 'member')
    WHERE role = 'member';

-- 9. 移除旧字段（后续 migration 或手动执行）
-- ALTER TABLE users DROP COLUMN is_super_admin;
-- ALTER TABLE users DROP COLUMN can_create_team;
-- ALTER TABLE teams DROP COLUMN pm_id;
-- ALTER TABLE team_members DROP COLUMN role;

COMMIT;
```

### PM 身份迁移

`teams.pm_id` → `team_members.role_id = pm_role_id`

迁移后 PM 身份完全由 `team_members.role_id` 决定。迁移脚本确保每个团队的 `pm_id` 对应的 `team_members` 记录已正确设置 `role_id`。

### 回滚方案

事务失败自动回滚。如需手动回滚已提交的迁移，提供逆向脚本：删除新列、新表。旧字段在迁移期间保留（步骤 9 为独立操作），降低回滚复杂度。

## Error Handling

### Error Type Definition

沿用现有 `internal/pkg/errors/errors.go` 中的 `AppError`：

```go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }
```

统一通过 `RespondError(c, err)` 返回错误响应，非 `*AppError` 类型降级为 500。

### New Error Types

```go
var (
    ErrRoleNotFound     = &AppError{Code: "ROLE_NOT_FOUND", Message: "角色不存在", Status: 404}
    ErrRoleInUse        = &AppError{Code: "ROLE_IN_USE", Message: "角色正在被使用，无法删除", Status: 409}
    ErrRolePreset       = &AppError{Code: "ROLE_PRESET", Message: "预置角色不可删除", Status: 409}
    ErrRoleDuplicate    = &AppError{Code: "ROLE_DUPLICATE", Message: "角色名称已存在", Status: 409}
    ErrPermissionDenied = &AppError{Code: "PERMISSION_DENIED", Message: "权限不足", Status: 403}
)
```

### Superadmin 检查失败

当前 `RequireRole("superadmin")` → 新 `RequireSuperAdmin()`，行为一致：返回 403。

### 权限码检查失败

`RequirePermission(code)` 不通过时返回 403，body 包含 `{ "code": "PERMISSION_DENIED" }`。

## Testing Strategy

### Test Framework & Tools

沿用项目现有测试工具：

| 工具 | 用途 |
|------|------|
| `testing` (stdlib) | 测试入口 |
| `github.com/stretchr/testify` | 断言 (`assert`/`require`) 和 mock (`mock`) |
| `net/http/httptest` | Handler 层 HTTP 测试 |
| `github.com/gin-gonic/gin` | `gin.CreateTestContext` 构造测试上下文 |

### Mock 策略

- **Repository 层**：使用 `testify/mock` 生成 mock 实现。所有 `*_repo.go` 接口均定义 mock（如 `MockRoleRepo`）。
- **Permission Cache**：提供 `NewMockCache()` 返回预设权限集，用于中间件测试。
- **Service 层**：Handler 测试中 mock Service 接口；Service 测试中 mock Repository 接口。
- **数据库测试**：Repository 层使用真实 SQLite 内存数据库（`sqlite.Open(":memory:")`）+ GORM AutoMigrate，不 mock GORM。

### Unit Tests

| 组件 | 测试内容 |
|------|---------|
| `permissions` | AllPermissions 包含 27 个权限码，AllCodes 查找正确 |
| `permcache` | 缓存命中/失效/刷新逻辑 |
| `RequireSuperAdmin` | superadmin 通过、普通用户 403 |
| `RequirePermission` | 有权限通过、无权限 403、superadmin bypass |
| `CanCreateTeam` | superadmin 通过、pm 在任意团队有 team:create 通过、普通用户 403 |
| `role_service` | CRUD 业务规则（名称重复、预置不可删、使用中不可删） |
| `role_repo` | CRUD + 权限绑定 + 成员计数 |

### Integration Tests

| 场景 | 测试内容 |
|------|---------|
| 角色 CRUD API | 创建→编辑→验证权限→删除完整流程 |
| 团队邀请带角色 | 邀请时指定角色 → 成员角色正确 → 权限生效 |
| 权限变更传播 | 修改角色权限 → 缓存刷新 → 中间件使用新权限 |
| 数据迁移 | 旧数据 → 执行迁移 → 验证角色映射正确 |
| 预置角色保护 | 尝试删除/编辑 superadmin → 409 |

### Coverage Target

核心包覆盖率 ≥ 80%：`permissions`、`permcache`、`middleware/rbac`、`service/role_service`。

## Security Considerations

### Threat Model

| 威胁 | 影响 | 缓解措施 |
|------|------|---------|
| JWT 泄露导致越权 | 高 | JWT 不含权限信息，后端实时校验；权限变更即时生效 |
| 权限提升攻击 | 高 | superadmin 检查从 DB 查询，不信任 JWT |
| 角色管理接口未授权访问 | 高 | RequireSuperAdmin 中间件保护所有 /admin/roles 路由 |
| 权限缓存不一致 | 低 | 角色变更立即刷新缓存；微秒级窗口可接受 |

### Mitigations

1. **前端隐藏仅为 UX**：所有权限检查在后端中间件层执行，前端不能绕过
2. **权限码不可动态创建**：权限码是 Go 常量，API 不能创建新权限码
3. **superadmin bypass 在中间件层**：不依赖业务代码手动检查
4. **迁移在事务中执行**：失败自动回滚，不破坏现有数据

## Open Questions

- [ ] 是否需要 rate limiting 在 `/api/me/permissions` 端点上？（当前无，建议与现有 login rate limiter 策略统一）

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| 权限码存数据库 | 管理员可在线创建权限码 | 权限码与路由绑定脱节，拼写错误风险 | PRD 明确禁止：权限码由代码定义 |
| 权限写入 JWT | 减少后端查询 | JWT 泄露暴露权限、变更需等 token 过期 | PRD 安全需求：JWT 不含权限 |
| Redis 缓存权限 | 分布式缓存 | 引入新依赖、角色数少无需分布式 | 角色数量少（< 50），内存缓存足够 |

### Preset Role Permission Assignments

**superadmin**: 代码中 bypass，不存 role_permissions 记录

**pm** (21 permissions):
```
team:create, team:read, team:update, team:invite, team:remove, team:transfer,
main_item:create, main_item:read, main_item:update, main_item:archive,
sub_item:create, sub_item:read, sub_item:update, sub_item:assign, sub_item:change_status,
progress:create, progress:read, progress:update,
item_pool:submit, item_pool:review,
view:weekly, view:gantt, view:table,
report:export,
user:read
```

**member** (11 permissions):
```
main_item:read,
sub_item:create, sub_item:read, sub_item:update, sub_item:change_status,
progress:create, progress:read,
item_pool:submit,
view:weekly, view:table,
report:export
```

### References

- PRD: `docs/features/rbac-permissions/prd/prd-spec.md`
- User Stories: `docs/features/rbac-permissions/prd/prd-user-stories.md`
- UI Functions: `docs/features/rbac-permissions/prd/prd-ui-functions.md`

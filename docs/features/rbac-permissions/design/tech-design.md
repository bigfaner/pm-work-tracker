---
created: 2026-04-19
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: RBAC 权限体系

## Overview

将现有三层分散权限检查（`RequireRole` / `RequireTeamRole` / `isPMOrSuperAdmin`）替换为统一的 `RequirePermission(code)` 中间件。权限码在 Go 代码中定义，角色-权限绑定存数据库，SuperAdmin 保留 `User.IsSuperAdmin` 布尔标记。

**改动范围**：

| 层 | 改动 |
|----|------|
| Model | 新增 `Role`、`RolePermission` 表；`TeamMember` 加 `RoleID`；`User` 移除 `CanCreateTeam` |
| Middleware | 新增 `RequirePermission`；移除 `RequireRole`/`RequireTeamRole`；改 `TeamScopeMiddleware` 加载权限码 |
| Handler | 移除所有 `isPMOrSuperAdmin`；新增 Role CRUD handler；改 JWT Claims |
| Service | 新增 `RoleService`；改造 `TeamService`（invite 用 RoleID）；改造 `AuthService`（JWT Claims） |
| Repository | 新增 `RoleRepo`；改造 `TeamRepo`（按 RoleID 查权限） |
| Frontend | Zustand store 加 permission map；新增 `<PermissionGuard>` 组件和 `useHasPermission` hook |

## Architecture

### Layer Placement

```
handler/router.go          → 路由注册，绑定 RequirePermission(code)
middleware/
  auth.go                  → AuthMiddleware（JWT → user_id, username）
  team_scope.go            → TeamScopeMiddleware（加载团队成员 + 角色权限）
  permission.go            → RequirePermission(code)（权限校验）
handler/
  role_handler.go          → 角色 CRUD（superadmin only）
  permission_handler.go    → 权限码列表 + 用户权限查询
service/
  role_service.go          → 角色业务逻辑
repository/
  role_repo.go             → 角色 + 角色权限数据访问
pkg/permissions/
  codes.go                 → 权限码注册表（Go 代码定义）
```

### Component Diagram

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Router      │────>│ RequirePermission│────>│  Handler        │
│  (gin)       │     │  Middleware      │     │  (CRUD logic)   │
└─────────────┘     └────────┬─────────┘     └────────┬────────┘
                             │                         │
                    ┌────────▼─────────┐      ┌───────▼────────┐
                    │ TeamScopeMid     │      │  Service        │
                    │ (load perm codes)│      │  (business)     │
                    └────────┬─────────┘      └───────┬────────┘
                             │                         │
                    ┌────────▼─────────────────────────▼────────┐
                    │            Repository Layer                │
                    │  RoleRepo  │  TeamRepo  │  UserRepo       │
                    └────────┬──────────────────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   Database      │
                    │  roles          │
                    │  role_permissions│
                    │  team_members   │
                    │  users          │
                    └─────────────────┘
```

### Middleware Chain

```
Request → AuthMiddleware → [TeamScopeMiddleware] → RequirePermission(code) → Handler

AuthMiddleware:
  1. Extract JWT → {user_id, username}
  2. Set context: userID, username
  3. Load User from DB → set isSuperAdmin in context

TeamScopeMiddleware (team-scoped routes only):
  1. Parse :teamId from URL
  2. If isSuperAdmin → set callerTeamRole="superadmin", empty permCodes → next
  3. Load TeamMember(teamID, userID) → get roleID
  4. Load RolePermissions(roleID) → get []permissionCode
  5. Set context: teamID, roleID, permCodes

RequirePermission(code):
  1. If isSuperAdmin → c.Next()
  2. If team context → code in permCodes → c.Next() / 403
  3. If non-team context → query DB: any role of this user has the code → c.Next() / 403
```

### Dependencies

| Dependency | Purpose | Status |
|---|---|---|
| `github.com/gin-gonic/gin` | HTTP framework | Existing |
| `gorm.io/gorm` | ORM | Existing |
| `github.com/golang-jwt/jwt/v5` | JWT signing | Existing |
| `github.com/glebarez/sqlite` | SQLite driver | Existing |
| No new external dependencies needed | — | — |

## Interfaces

### Permission Code Registry

```go
// internal/pkg/permissions/codes.go
package permissions

// ActionDef describes a single permission action within a resource.
type ActionDef struct {
    Code        string // e.g. "team:create"
    Description string // e.g. "创建团队"
}

// Resource → Actions mapping. Single source of truth for all permission codes.
var Registry = map[string][]ActionDef{
    "team": {
        {Code: "team:create", Description: "创建团队"},
        {Code: "team:read", Description: "查看团队信息"},
        {Code: "team:update", Description: "编辑团队信息"},
        {Code: "team:delete", Description: "解散团队"},
        {Code: "team:invite", Description: "邀请成员加入"},
        {Code: "team:remove", Description: "移除团队成员"},
        {Code: "team:transfer", Description: "转让PM身份"},
    },
    "main_item": { /* ... */ },
    "sub_item":  { /* ... */ },
    "progress":  { /* ... */ },
    "item_pool": { /* ... */ },
    "view":      { /* ... */ },
    "report":    { /* ... */ },
    "user":      { /* ... */ },
}

// AllCodes returns a flat set of all valid permission code strings.
func AllCodes() map[string]bool

// ValidateCode checks if a code exists in the registry.
func ValidateCode(code string) bool
```

### RoleService Interface

```go
// internal/service/role_service.go
type RoleService interface {
    ListRoles(ctx context.Context) ([]Role, error)
    GetRole(ctx context.Context, roleID uint) (*Role, error)
    CreateRole(ctx context.Context, req CreateRoleReq) (*Role, error)
    UpdateRole(ctx context.Context, roleID uint, req UpdateRoleReq) (*Role, error)
    DeleteRole(ctx context.Context, roleID uint) error
    ListPermissionCodes(ctx context.Context) map[string][]ActionDef
    GetUserPermissions(ctx context.Context, userID uint) (*UserPermissions, error)
}
```

### RoleRepo Interface

```go
// internal/repository/role_repo.go
type RoleRepo interface {
    List(ctx context.Context) ([]model.Role, error)
    FindByID(ctx context.Context, id uint) (*model.Role, error)
    FindByName(ctx context.Context, name string) (*model.Role, error)
    Create(ctx context.Context, role *model.Role) error
    Update(ctx context.Context, role *model.Role) error
    Delete(ctx context.Context, id uint) error

    // Permission bindings
    ListPermissions(ctx context.Context, roleID uint) ([]string, error)
    SetPermissions(ctx context.Context, roleID uint, codes []string) error

    // Usage count
    CountMembersByRoleID(ctx context.Context, roleID uint) (int64, error)

    // Non-team-context: check if any of user's roles has the given permission code
    HasPermission(ctx context.Context, userID uint, code string) (bool, error)
}
```

### UserPermissions (API response struct)

```go
type UserPermissions struct {
    IsSuperAdmin    bool                `json:"is_superadmin"`
    TeamPermissions map[uint][]string   `json:"team_permissions"` // teamID → codes
}
```

### Migration Function

```go
// internal/migration/rbac.go
// MigrateToRBAC 在单个数据库事务中完成 RBAC 数据迁移。
// 步骤：创建新表 → 种子预置角色 → 迁移 team_members.role → 移除 users.can_create_team。
// 幂等：通过 schema_migrations 表跟踪，重复执行不产生副作用。
func MigrateToRBAC(db *gorm.DB) error
```

## Data Models

### New: Role

```sql
CREATE TABLE roles (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL UNIQUE,
    description TEXT   NOT NULL DEFAULT '',
    is_preset  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);
```

### New: RolePermission

```sql
CREATE TABLE role_permissions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    role_id         INTEGER NOT NULL REFERENCES roles(id),
    permission_code TEXT    NOT NULL,
    UNIQUE(role_id, permission_code)
);
```

### Modified: TeamMember

```sql
-- Add column
ALTER TABLE team_members ADD COLUMN role_id INTEGER REFERENCES roles(id);

-- After migration, remove old column
-- ALTER TABLE team_members DROP COLUMN role;  -- SQLite 不支持 DROP COLUMN，迁移时重建表
```

**迁移策略**：SQLite 不支持 DROP COLUMN。迁移脚本将：
1. 创建新表 `team_members_new`（含 `role_id`，不含 `role`）
2. 数据迁移（`role` → `role_id` 映射）
3. 删除旧表，重命名新表

### Modified: User

```sql
-- Remove after migration (同样需要重建表 for SQLite)
-- ALTER TABLE users DROP COLUMN can_create_team;
-- 保留 is_super_admin 字段
```

### Preset Roles Seed Data

| role_id | name | is_preset | permissions |
|---------|------|-----------|-------------|
| 1 | superadmin | true | 全部权限码（通过 IsSuperAdmin 标记绕过，实际不绑定权限码到 role_permissions） |
| 2 | pm | true | team:create, team:read, team:update, team:delete, team:invite, team:remove, team:transfer, main_item:*, sub_item:*, progress:*, item_pool:*, view:*, report:*, user:read |
| 3 | member | true | main_item:read, sub_item:create, sub_item:read, sub_item:update, sub_item:change_status, progress:create, progress:read, item_pool:submit, view:weekly, view:table, report:export |

> **注**：superadmin 角色存在于 roles 表中供展示，但实际权限绕过通过 `User.IsSuperAdmin` 标记实现，不需要在 `role_permissions` 中存储其权限码。

### ER Diagram

```
users                    teams                   roles
┌───────────────┐       ┌──────────────┐       ┌──────────────────┐
│ id            │       │ id           │       │ id               │
│ username      │       │ name         │       │ name             │
│ display_name  │       │ description  │       │ description      │
│ password_hash │       │ pm_id (FK)   │──┐    │ is_preset        │
│ email         │       │ created_at   │  │    │ created_at       │
│ is_super_admin│       │ updated_at   │  │    │ updated_at       │
│ status        │       │ deleted_at   │  │    │ deleted_at       │
└───────────────┘       └──────────────┘  │    └──────────────────┘
                                           │           │
team_members                              │    role_permissions
┌───────────────┐                         │    ┌──────────────────┐
│ id            │                         │    │ id               │
│ team_id (FK)  │─────────────────────────┘    │ role_id (FK)     │
│ user_id (FK)  │                              │ permission_code  │
│ role_id (FK)  │──────────────────────────────┤                  │
│ joined_at     │                              └──────────────────┘
│ UNIQUE(team_id, user_id) │
└───────────────┘
```

## Handler Changes

### Removed Code

| File | What to Remove |
|------|---------------|
| `handler/main_item_handler.go` | `isPMOrSuperAdmin()` method |
| `handler/sub_item_handler.go` | `isPMOrSuperAdmin()` method |
| `handler/progress_handler.go` | `isPMOrSuperAdmin()` method |
| `middleware/rbac.go` | `RequireRole()`, `RequireTeamRole()` (replaced by `RequirePermission`) |
| `pkg/jwt/jwt.go` | `Role` field from `Claims` struct |

### Router Changes (handler/router.go)

```go
// Before:
adminRoutes := api.Group("/admin", authMiddleware, rbac.RequireRole("superadmin"))
teamRoutes := api.Group("/teams/:teamId", authMiddleware, teamScopeMiddleware)
teamRoutes.PUT("", rbac.RequireTeamRole("pm"), h.teamHandler.Update)

// After:
adminRoutes := api.Group("/admin", authMiddleware, perm.RequirePermission("user:manage_role"))
teamRoutes := api.Group("/teams/:teamId", authMiddleware, teamScopeMiddleware)
teamRoutes.PUT("", perm.RequirePermission("team:update"), h.teamHandler.Update)
```

**完整路由权限映射**：

| Method | Path | Permission Code |
|--------|------|----------------|
| POST | /api/v1/teams | `team:create` |
| GET | /api/v1/teams | (auth only) |
| GET | /api/v1/teams/:teamId | `team:read` |
| PUT | /api/v1/teams/:teamId | `team:update` |
| DELETE | /api/v1/teams/:teamId | `team:delete` |
| POST | /api/v1/teams/:teamId/members | `team:invite` |
| DELETE | /api/v1/teams/:teamId/members/:memberId | `team:remove` |
| PUT | /api/v1/teams/:teamId/pm | `team:transfer` |
| PUT | /api/v1/teams/:teamId/members/:memberId/role | `team:invite` |
| POST | /api/v1/teams/:teamId/main-items | `main_item:create` |
| GET | /api/v1/teams/:teamId/main-items | `main_item:read` |
| PUT | /api/v1/teams/:teamId/main-items/:itemId | `main_item:update` |
| PUT | /api/v1/teams/:teamId/main-items/:itemId/archive | `main_item:archive` |
| POST | /api/v1/teams/:teamId/main-items/:mainId/sub-items | `sub_item:create` |
| PUT | /api/v1/teams/:teamId/sub-items/:itemId | `sub_item:update` |
| PUT | /api/v1/teams/:teamId/sub-items/:itemId/assignee | `sub_item:assign` |
| PUT | /api/v1/teams/:teamId/sub-items/:itemId/status | `sub_item:change_status` |
| POST | /api/v1/teams/:teamId/sub-items/:subItemId/progress | `progress:create` |
| PUT | /api/v1/teams/:teamId/progress/:recordId/completion | `progress:update` |
| POST | /api/v1/teams/:teamId/item-pool | `item_pool:submit` |
| PUT | /api/v1/teams/:teamId/item-pool/:poolId/assign | `item_pool:review` |
| PUT | /api/v1/teams/:teamId/item-pool/:poolId/convert | `item_pool:review` |
| PUT | /api/v1/teams/:teamId/item-pool/:poolId/reject | `item_pool:review` |
| GET | /api/v1/teams/:teamId/views/weekly | `view:weekly` |
| GET | /api/v1/teams/:teamId/views/gantt | `view:gantt` |
| GET | /api/v1/teams/:teamId/views/table | `view:table` |
| GET | /api/v1/teams/:teamId/reports/weekly/preview | `report:export` |
| GET | /api/v1/teams/:teamId/reports/weekly/export | `report:export` |
| GET | /api/v1/admin/users | `user:read` |
| POST | /api/v1/admin/users | `user:manage_role` |
| GET | /api/v1/admin/users/:id | `user:read` |
| PUT | /api/v1/admin/users/:id | `user:update` |
| PUT | /api/v1/admin/users/:id/status | `user:update` |
| GET | /api/v1/admin/roles | `user:manage_role` |
| POST | /api/v1/admin/roles | `user:manage_role` |
| GET | /api/v1/admin/roles/:id | `user:manage_role` |
| PUT | /api/v1/admin/roles/:id | `user:manage_role` |
| DELETE | /api/v1/admin/roles/:id | `user:manage_role` |
| GET | /api/v1/admin/permissions | `user:manage_role` |
| GET | /api/v1/me/permissions | (auth only) |

> **注**：`sub_item:update` 和 `sub_item:change_status` 在中间件层检查权限码后，handler 层仍保留 assignee 附加检查（业务规则，非权限码范畴）。

### Assignee Pattern (保留的业务规则)

`sub_item:update` 和 `sub_item:change_status` 有额外的 assignee 访问规则，在 handler 层处理：

```go
// sub_item_handler.go
func (h *SubItemHandler) Update(c *gin.Context) {
    // 中间件已检查 sub_item:update 权限码
    // handler 额外检查：PM/SuperAdmin 可编辑所有，其他只能编辑分配给自己的
    if !middleware.IsSuperAdmin(c) && !middleware.IsPM(c) {
        item := /* load sub_item */
        if item.AssigneeID != middleware.GetUserID(c) {
            errors.RespondError(c, errors.ErrForbidden)
            return
        }
    }
    // ... proceed
}
```

### JWT Claims Change

```go
// Before
type Claims struct {
    jwt.RegisteredClaims
    UserID uint
    Role   string  // "superadmin" or "user"
}

// After
type Claims struct {
    jwt.RegisteredClaims
    UserID   uint
    Username string
}
```

- `Sign()` 签名时只传入 `userID` 和 `username`
- `AuthMiddleware` 从 JWT 解析后设置 context: `userID`, `username`
- 不再从 JWT 读取 Role；改为在 `TeamScopeMiddleware` 中从 DB 加载

### Invite Member Change

```go
// Before: InviteMemberReq
type InviteMemberReq struct {
    Username string `json:"username" binding:"required"`
    Role     string `json:"role" binding:"required,oneof=member"` // 硬编码只能邀请 member
}

// After: InviteMemberReq
type InviteMemberReq struct {
    Username string `json:"username" binding:"required"`
    RoleID   uint   `json:"roleId" binding:"required"` // 从角色列表选择
}
```

Service 层校验：RoleID 必须存在、不能是 superadmin 角色。

## Frontend Design

### Auth Store Extension

```typescript
// frontend/src/store/auth.ts additions

interface PermissionMap {
  isSuperadmin: boolean;
  teamPermissions: Record<number, string[]>; // teamId → permission codes
}

interface AuthState {
  // ... existing fields ...
  permissions: PermissionMap | null;
  permissionsLoadedAt: number | null; // timestamp

  // New methods
  fetchPermissions(): Promise<void>;
  hasPermission(code: string, teamId?: number): boolean;
}
```

- `fetchPermissions()` calls `GET /api/v1/me/permissions`
- `hasPermission(code, teamId?)`: if `isSuperadmin` → true; if `teamId` given → check `teamPermissions[teamId]`; else check if any team has the code

### Permission Guard Component

```tsx
// frontend/src/components/PermissionGuard.tsx
interface Props {
  code: string;
  teamId?: number;
  children: React.ReactNode;
}

function PermissionGuard({ code, teamId, children }: Props) {
  const hasPermission = useAuthStore(s => s.hasPermission(code, teamId));
  if (!hasPermission) return null;
  return <>{children}</>;
}
```

Usage:
```tsx
<PermissionGuard code="team:invite" teamId={currentTeamId}>
  <Button>邀请成员</Button>
</PermissionGuard>
```

### useHasPermission Hook

```typescript
// frontend/src/hooks/usePermission.ts
function useHasPermission(code: string, teamId?: number): boolean {
  return useAuthStore(s => s.hasPermission(code, teamId));
}
```

### Permission Refresh Strategy

| 触发方式 | 说明 |
|----------|------|
| 登录成功 | 立即 fetchPermissions |
| 路由切换 | 若距上次加载 >5 分钟，重新 fetchPermissions |
| 收到 403 响应 | 立即 fetchPermissions（权限可能已变更） |
| 手动刷新 | 在 team member 管理页面提供刷新按钮 |

### Admin Route Guard Change

```tsx
// Before
function AdminRoute() {
  const isSuperAdmin = useAuthStore(s => s.isSuperAdmin);
  // ...
}

// After
function AdminRoute() {
  const isSuperadmin = useAuthStore(s => s.permissions?.isSuperadmin ?? false);
  // ...same logic, source changes
}
```

### Sidebar Navigation Change

```tsx
// Before
const allNavItems = isSuperAdmin ? [...navItems, ...adminItems, teamItem] : [...navItems, teamItem];

// After — use permission codes
const allNavItems = [
  ...navItems,   // items, weekly, gantt, item-pool, report — 每个受对应 view:* 权限码控制
  teamItem,      // 始终显示
  ...(hasPermission('user:read') ? adminItems : []),  // User Management
];
// 甘特图入口受 view:gantt 权限码控制
```

### Type Changes

```typescript
// frontend/src/types/index.ts

// Remove from User type:
// canCreateTeam: boolean;

// Add:
interface User {
  // ... existing ...
  // Remove: canCreateTeam
  // Keep: isSuperAdmin (renamed from isSuperAdmin for clarity)
}

// New types:
interface Role {
  id: number;
  name: string;
  description: string;
  isPreset: boolean;
  permissionCount: number;
  memberCount: number;
  createdAt: string;
}

interface PermissionGroup {
  resource: string;
  actions: { code: string; description: string }[];
}
```

## Error Handling

### New Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `ERR_ROLE_NOT_FOUND` | 404 | 角色不存在 |
| `ERR_ROLE_NAME_EXISTS` | 409 | 角色名称已存在 |
| `ERR_ROLE_IN_USE` | 422 | 角色正在被使用，无法删除 |
| `ERR_PRESET_ROLE_IMMUTABLE` | 403 | 预置角色不可编辑/删除 |
| `ERR_INVALID_PERMISSION_CODE` | 400 | 无效的权限码 |

### Permission Denied Response

统一通过 `RequirePermission` 中间件返回 403：

```json
{
  "code": "ERR_FORBIDDEN",
  "message": "权限不足：缺少 team:invite 权限"
}
```

## Testing Strategy

### Unit Tests

| 测试对象 | 测试要点 | 文件 |
|----------|----------|------|
| Permission Code Registry | AllCodes 完整性、ValidateCode 正确性 | `pkg/permissions/codes_test.go` |
| RequirePermission 中间件 | superadmin 绕过、有权限通过、无权限 403 | `middleware/permission_test.go` |
| TeamScopeMiddleware | 加载 roleID 和 permCodes 到 context | `middleware/team_scope_test.go` |
| RoleService | CRUD 逻辑、预置角色保护、重名检测 | `service/role_service_test.go` |
| RoleRepo | 数据库 CRUD、权限绑定、使用人数计数 | `repository/gorm/role_repo_test.go` |
| JWT Claims | 新 Claims 结构签名/验证 | `pkg/jwt/jwt_test.go` |

### Integration Tests

使用 `net/http/httptest` 创建测试 HTTP 服务器，配合测试 SQLite 数据库（`:memory:`）进行全链路测试。

| 测试场景 | 覆盖范围 |
|----------|----------|
| 角色 CRUD 全流程 | POST/GET/PUT/DELETE /admin/roles |
| 邀请成员 + 角色选择 | POST /teams/:id/members with roleId |
| 权限驱动的 API 访问 | 不同角色访问各 endpoint 的 200/403 |
| 数据迁移脚本 | 迁移前后数据一致性、幂等性 |

### Coverage Target

- 新增代码：≥ 85%
- 修改的中间件：≥ 90%
- 测试工具：`testing`（Go 标准）、`github.com/stretchr/testify`（断言）、`net/http/httptest`（HTTP 集成测试）、`gorm.io/driver/sqlite`（`:memory:` 测试数据库）

## Security Considerations

### Threat Model

| 威胁 | 风险 | 缓解措施 |
|------|------|----------|
| 前端绕过 | 用户通过开发者工具调用无权限 API | 后端 RequirePermission 中间件强制校验 |
| 权限码注入 | 管理员分配不存在的权限码 | Go 代码校验 permission_code 合法性 |
| 提权攻击 | 用户修改自己的角色 | 角色 CRUD 仅 superadmin 可操作；TeamScopeMiddleware 从 DB 加载 |
| 水平越权 | A 团队 PM 操作 B 团队资源 | TeamScopeMiddleware 按 teamID 隔离 |
| 角色删除风险 | 删除正在使用的角色 | 先检查使用人数，有用户则拒绝删除 |

### Permission Check Flow (Defense in Depth)

```
Layer 1: Router Middleware (RequirePermission)  ← 强制，不可绕过
Layer 2: Handler (assignee check)                ← 业务规则补充
Layer 3: Service (ownership check)               ← 纵深防御
```

## Open Questions

- [ ] SQLite 重建表迁移方案确认（SQLite 不支持 DROP COLUMN，需要 CREATE + COPY + DROP + RENAME）
- [ ] 前端 permission 轮询间隔确认（建议 5 分钟，过长会导致权限变更延迟感知）

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| `user_global_roles` 表追踪 SuperAdmin | 完全消除 User 上的权限标记 | 多一张表、多一次 JOIN、增加复杂度 | SuperAdmin 是唯一的全局角色，用布尔标记更简单 |
| Redis 缓存权限码 | 高性能 | 引入外部依赖 | SQLite 本地查询 <3ms，无需 Redis |
| JWT 内携带权限码 | 减少后端查询 | 权限变更需等 token 过期、增加 token 体积 | PRD 明确要求 JWT 不含权限字段 |
| 权限码存数据库 | 可在线扩展 | 违背 PRD"由代码定义"原则，且权限码与代码逻辑耦合 | PRD 要求权限码由代码定义 |

### References

- PRD: `docs/features/rbac-permissions/prd/prd-spec.md`
- User Stories: `docs/features/rbac-permissions/prd/prd-user-stories.md`
- UI Functions: `docs/features/rbac-permissions/prd/prd-ui-functions.md`

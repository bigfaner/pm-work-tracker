---
created: 2026-04-28
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: 细化 user / role 权限粒度

## Overview

本次变更是一次纯粹的权限码重构，不涉及新增 API 接口或数据库表结构变更。核心工作分四个层面：

1. **权限码注册表**（`permissions/codes.go`）：更新 `user` 资源 4 个码，新增 `role` 资源 4 个码
2. **路由中间件绑定**（`router.go`）：14 条路由重新绑定新权限码
3. **数据迁移**（`migration/rbac.go`）：新增幂等迁移函数，将存量 `pmw_role_permissions` 中的旧码转换为新码
4. **前端权限守卫**（`lib/permissions.ts`、`App.tsx`、`Sidebar.tsx`、`TeamManagementPage.tsx`）：替换旧码引用

## Architecture

### Layer Placement

| Layer | 变更内容 |
|-------|---------|
| `pkg/permissions/` | 权限码常量注册表（单一真实来源） |
| `migration/` | 数据迁移函数（幂等，事务执行） |
| `handler/router.go` | 路由中间件绑定（`deps.perm(code)` 调用） |
| `handler/router_test.go` | 测试种子数据更新 |
| `frontend/src/lib/permissions.ts` | 前端权限码常量（UI 多选框数据源） |
| `frontend/src/App.tsx` | 路由级权限守卫 |
| `frontend/src/components/layout/Sidebar.tsx` | 导航菜单权限控制 |
| `frontend/src/pages/TeamManagementPage.tsx` | 角色下拉选择器权限控制 |

### Component Diagram

```
permissions/codes.go  ←── 单一真实来源
        │
        ├── router.go (deps.perm("role:read") 等)
        │       └── middleware/permission.go (RequirePermission)
        │
        ├── migration/rbac.go (seedPresetRoles + MigratePermissionGranularity)
        │       └── pmw_role_permissions 表
        │
        └── frontend/src/lib/permissions.ts
                ├── App.tsx (PermissionRoute)
                ├── Sidebar.tsx (菜单项 permission 字段)
                └── TeamManagementPage.tsx (角色下拉 role:read 检查)
```

### Dependencies

无新增外部依赖。所有变更在现有代码结构内完成。

## Interfaces

### 1. 权限码注册表变更（`backend/internal/pkg/permissions/codes.go`）

`user` 资源从 3 个码变为 4 个码，新增 `role` 资源 4 个码：

```go
// user 资源（更新）
{Code: "user:list",        Description: "列出用户（成员选择器、用户列表页）"},
{Code: "user:read",        Description: "查看用户详情（含敏感字段，用户管理页）"},
{Code: "user:update",      Description: "编辑用户信息"},          // 不变
{Code: "user:assign_role", Description: "给用户分配角色"},         // 原 user:manage_role

// role 资源（新增）
{Resource: "role", Permissions: []Permission{
    {Code: "role:read",   Description: "查看角色列表和详情"},
    {Code: "role:create", Description: "创建新角色"},
    {Code: "role:update", Description: "编辑角色名称、描述、权限码"},
    {Code: "role:delete", Description: "删除自定义角色"},
}}
```

移除：`user:manage_role`

### 2. 路由绑定变更（`backend/internal/handler/router.go`）

仅修改 `deps.perm(...)` 参数，不改动路由路径或 handler：

```go
// admin 路由组变更（adminGroup）
adminGroup.GET("/users",                deps.perm("user:list"),        deps.Admin.ListUsers)
adminGroup.POST("/users",               deps.perm("user:assign_role"), deps.Admin.CreateUser)
adminGroup.GET("/users/:userId",        deps.perm("user:read"),        deps.Admin.GetUser)
// PUT /users/:userId, status, password, DELETE /users/:userId — user:update 不变
adminGroup.GET("/teams",                deps.perm("user:list"),        deps.Admin.ListTeams)

adminGroup.GET("/roles",                deps.perm("role:read"),        deps.Role.ListRoles)
adminGroup.POST("/roles",               deps.perm("role:create"),      deps.Role.CreateRole)
adminGroup.GET("/roles/:id",            deps.perm("role:read"),        deps.Role.GetRole)
adminGroup.PUT("/roles/:id",            deps.perm("role:update"),      deps.Role.UpdateRole)
adminGroup.DELETE("/roles/:id",         deps.perm("role:delete"),      deps.Role.DeleteRole)
adminGroup.GET("/permissions",          deps.perm("role:read"),        deps.Permission.ListPermissionCodes)
```

### 3. 数据迁移函数（`backend/internal/migration/rbac.go`）

新增 `MigratePermissionGranularity` 函数，幂等执行，通过 `schema_migrations` 表追踪版本：

```go
const permGranularityVersion = "permission_granularity_001"

// MigratePermissionGranularity 将存量 pmw_role_permissions 中的旧权限码
// 转换为新权限码。幂等：重复执行无副作用。
func MigratePermissionGranularity(db *gorm.DB) error
```

**迁移逻辑（事务内执行）：**

```
Step 1: 快照备份（SELECT INTO 临时表或应用层记录）
Step 2: user:manage_role → role:create + role:update + role:delete
  - 查找所有 role_id WHERE permission_code = 'user:manage_role'
  - 对每个 role_id INSERT role:create, role:update, role:delete（忽略已存在）
  - DELETE WHERE permission_code = 'user:manage_role'
Step 3: 旧 user:read → user:list + 新 user:read（原子授予）
  - 查找所有 role_id WHERE permission_code = 'user:read'
  - 对每个 role_id INSERT user:list（忽略已存在）
  - 对每个 role_id INSERT user:read（新语义，忽略已存在）— 原子授予，避免空窗
  - DELETE WHERE permission_code = 'user:read'（旧码已被新码替代）
Step 4: 标记 schema_migrations version = 'permission_granularity_001'
```

> **注意**：步骤 3 中旧 `user:read` 和新 `user:read` 的 `permission_code` 字符串相同，但语义不同。迁移的实际操作是：先插入 `user:list`（新码），再确保 `user:read` 行存在（语义已变为"查看详情"）。由于旧 `user:read` 行本身不删除（语义已更新为新含义），此步骤实际只需插入 `user:list`。

**修正后的步骤 3（更准确）：**

```
Step 3: 旧 user:read 语义迁移
  - 查找所有 role_id WHERE permission_code = 'user:read'（这些角色原来有"列出用户"权限）
  - 对每个 role_id INSERT user:list（新码，忽略已存在）
  - 旧 user:read 行保留（其语义自动变为"查看用户详情"，因为路由绑定已更新）
  - 结果：原持有 user:read 的角色同时持有 user:list 和 user:read（新语义）
```

### 4. `seedPresetRoles` 更新（`backend/internal/migration/rbac.go`）

pm 角色权限码列表更新：

```go
pmCodes := []string{
    // team (不变)
    "team:create", "team:read", "team:update", "team:delete",
    "team:invite", "team:remove", "team:transfer",
    // main_item (不变)
    "main_item:create", "main_item:read", "main_item:update", "main_item:archive",
    "main_item:change_status",
    // sub_item (不变)
    "sub_item:create", "sub_item:read", "sub_item:update", "sub_item:assign", "sub_item:change_status",
    // progress (不变)
    "progress:create", "progress:read", "progress:update",
    // item_pool (不变)
    "item_pool:submit", "item_pool:review",
    // view (不变)
    "view:weekly", "view:gantt", "view:table",
    // report (不变)
    "report:export",
    // user (更新：user:read→user:list+user:read，user:manage_role→user:assign_role)
    "user:list", "user:read", "user:update", "user:assign_role",
    // role (新增)
    "role:read", "role:create", "role:update", "role:delete",
}
// member 不变
```

### 5. 前端权限码常量（`frontend/src/lib/permissions.ts`）

`PERMISSION_GROUPS` 中 `user` 组更新，新增 `role` 组：

```ts
// user 组（更新）
{ value: 'user:list',        label: '列出用户' },
{ value: 'user:read',        label: '查看用户详情' },
{ value: 'user:update',      label: '编辑用户' },
{ value: 'user:assign_role', label: '分配角色' },

// role 组（新增）
{
  key: 'role',
  label: '角色管理',
  permissions: [
    { value: 'role:read',   label: '查看角色' },
    { value: 'role:create', label: '创建角色' },
    { value: 'role:update', label: '编辑角色' },
    { value: 'role:delete', label: '删除角色' },
  ],
}
```

### 6. 前端路由守卫（`frontend/src/App.tsx`）

```tsx
// 变更前
<Route element={<PermissionRoute code="user:read" />}>        // 用户管理页
<Route element={<PermissionRoute code="user:manage_role" />}> // 角色管理页

// 变更后
<Route element={<PermissionRoute code="user:list" />}>        // 用户管理页
<Route element={<PermissionRoute code="role:read" />}>        // 角色管理页
```

### 7. 导航菜单（`frontend/src/components/layout/Sidebar.tsx`）

```ts
// 变更前
{ key: '/users', label: '用户管理', icon: UserCog, permission: 'user:read' },
{ key: '/roles', label: '角色管理', icon: Shield,  permission: 'user:manage_role' },

// 变更后
{ key: '/users', label: '用户管理', icon: UserCog, permission: 'user:list' },
{ key: '/roles', label: '角色管理', icon: Shield,  permission: 'role:read' },
```

### 8. 团队管理页角色下拉（`frontend/src/pages/TeamManagementPage.tsx`）

当前 `useQuery(['roles'], listRolesApi)` 无权限检查，直接调用。变更后：

```tsx
const { hasPermission } = useAuthStore()
const canReadRoles = hasPermission('role:read')

const { data: rolesData } = useQuery({
  queryKey: ['roles'],
  queryFn: listRolesApi,
  enabled: canReadRoles,   // 新增：无 role:read 时不发请求
})
```

下拉选择器在 `canReadRoles === false` 时显示禁用状态（见 prd-ui-functions.md UI Function 1）。

## Data Models

无新增数据库表或字段。`pmw_role_permissions.permission_code` 列的值集合发生变化，由迁移脚本处理。

## Error Handling

### Error Types & Codes

| Error Code | Name | Description | HTTP Status |
|------------|------|-------------|-------------|
| ERR_FORBIDDEN | AppError | 缺少对应权限码时，`RequirePermission` 中间件返 | 403 |

无新增错误类型。现有 `RequirePermission` 中间件已处理所有权限不足场景。

### Propagation Strategy

不变：`RequirePermission` → `apperrors.RespondError` → JSON `{code: "ERR_FORBIDDEN", message: "权限不足：缺少 X 权限"}`。

## Cross-Layer Data Map

| Field Name | Storage Layer | Backend Model | API/DTO | Frontend Type | Validation Rule |
|------------|---------------|---------------|---------|---------------|-----------------|
| `permission_code` (user:list) | `pmw_role_permissions.permission_code VARCHAR(50)` | `model.RolePermission.PermissionCode string` | `json:"code"` | `string` in `PermissionData.teamPermissions` | 必须在 `permissions.Registry` 中注册 |
| `permission_code` (user:read 新) | 同上 | 同上 | 同上 | 同上 | 同上 |
| `permission_code` (user:assign_role) | 同上 | 同上 | 同上 | 同上 | 同上 |
| `permission_code` (role:read) | 同上 | 同上 | 同上 | 同上 | 同上 |
| `permission_code` (role:create) | 同上 | 同上 | 同上 | 同上 | 同上 |
| `permission_code` (role:update) | 同上 | 同上 | 同上 | 同上 | 同上 |
| `permission_code` (role:delete) | 同上 | 同上 | 同上 | 同上 | 同上 |

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| 路由中间件 | 单元测试 | Go `httptest` + Gin | 4 个新权限码（role:read/create/update/delete）各自在无权限时返回 403 | 4 个新测试用例 |
| 迁移函数 | 单元测试 | Go + SQLite in-memory | `MigratePermissionGranularity` 幂等性、旧码转换正确性、事务回滚 | 3 个测试用例 |
| 预置角色 | 单元测试 | Go + SQLite in-memory | `VerifyPresetRoleCodes` 更新后通过 | 更新现有测试 |
| 前端权限守卫 | 单元测试 | vitest + testing-library | `user:list` 控制用户管理页入口，`role:read` 控制角色管理页入口 | 更新现有 2 个测试 |

### Key Test Scenarios

**后端（新增 4 个路由中间件测试）：**

```go
// 1. GET /admin/roles — 无 role:read → 403
// 2. POST /admin/roles — 无 role:create → 403
// 3. PUT /admin/roles/:id — 无 role:update → 403
// 4. DELETE /admin/roles/:id — 无 role:delete → 403
```

**迁移测试（新增 3 个）：**

```go
// 1. user:manage_role → role:create+update+delete（原码消失，3 个新码存在）
// 2. 旧 user:read → user:list 插入，user:read 行保留（语义更新）
// 3. 幂等：重复执行 MigratePermissionGranularity 无副作用
```

**前端（更新现有测试）：**

```ts
// permission-driven-ui.test.tsx
// 原: user:read → 用户管理页可见
// 改: user:list → 用户管理页可见
// 原: user:manage_role → 角色管理页可见
// 改: role:read → 角色管理页可见
```

**`router_test.go` 种子数据更新：**

```go
// 将 "user:read", "user:manage_role" 替换为
// "user:list", "user:read", "user:assign_role",
// "role:read", "role:create", "role:update", "role:delete"
```

### Overall Coverage Target

本次变更无新增业务逻辑，测试目标为：所有受影响的现有测试通过 + 新增 7 个测试用例（4 路由 + 3 迁移）。

## Security Considerations

### Threat Model

1. **权限码语义碰撞**：旧 `user:read`（列出用户）和新 `user:read`（查看详情）同名不同义，迁移期间可能出现短暂的权限混乱
2. **迁移脚本部分执行**：若迁移中途失败，可能出现旧码已删除但新码未插入的状态

### Mitigations

1. 迁移在单一数据库事务中执行，失败自动回滚；迁移前对 `pmw_role_permissions` 做快照
2. 步骤 3 中旧 `user:read` 行不删除（语义随路由绑定更新而更新），避免空窗期
3. CI grep 断言：`user:manage_role` 引用数为零才允许合并
4. 后端是安全防线，前端权限隐藏仅为 UX 优化

## PRD Coverage Map

| PRD Requirement / AC | Design Component | Interface / Model |
|----------------------|------------------|-------------------|
| Story 1: role:read → GET /admin/roles 返回 200 | `router.go` 路由绑定 | `deps.perm("role:read")` |
| Story 1: 无 role:read → GET /admin/roles 返回 403 | `middleware/permission.go` | `RequirePermission` |
| Story 2: user:list → GET /admin/users 返回 200 | `router.go` 路由绑定 | `deps.perm("user:list")` |
| Story 2: 无 user:read → GET /admin/users/:userId 返回 403 | `router.go` 路由绑定 | `deps.perm("user:read")` |
| Story 3: role:create → POST /admin/roles 返回 201 | `router.go` 路由绑定 | `deps.perm("role:create")` |
| Story 3: 只有 role:read → POST /admin/roles 返回 403 | `middleware/permission.go` | `RequirePermission` |
| Story 4: pm 有完整 user:* + role:* | `migration/rbac.go` | `seedPresetRoles` pmCodes |
| Story 5: user:manage_role → role:create+update+delete | `migration/rbac.go` | `MigratePermissionGranularity` |
| Story 5: 旧 user:read → user:list（原子授予新 user:read） | `migration/rbac.go` | `MigratePermissionGranularity` |
| Story 5: 事务回滚 | `migration/rbac.go` | `MigratePermissionGranularity` |
| Story 6: 无 user:list → 用户管理页入口不显示 | `App.tsx` + `Sidebar.tsx` | `PermissionRoute code="user:list"` |
| Story 6: 无 role:create → 创建角色按钮不显示 | `RoleManagementPage.tsx` | `usePermission("role:create")` |
| Story 6: 后端是安全防线 | `middleware/permission.go` | `RequirePermission` |

## Open Questions

- [x] `user:read` 旧码行是否删除？→ 不删除，语义随路由绑定更新（见迁移步骤 3 修正）
- [x] member 角色是否自动获得 `role:read`？→ 否，由管理员按需授予

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| 新增独立迁移文件 `permission_granularity.go` | 职责分离 | 与 `rbac.go` 中的 `SyncPresetRoles` 调用链分离，需要在 `main.go` 额外注册 | 改动量更大，`rbac.go` 已有完整的迁移基础设施 |
| 向后兼容别名（旧码继续有效） | 消费方无需立即迁移 | 别名与拆分目标相悖，中间件需特殊分支 | 已在 proposal 中否决 |

### References

- `backend/internal/pkg/permissions/codes.go` — 权限码注册表
- `backend/internal/migration/rbac.go` — 迁移基础设施
- `backend/internal/handler/router.go` — 路由绑定
- `frontend/src/lib/permissions.ts` — 前端权限码常量
- `docs/proposals/permission-granularity/proposal.md` — 原始提案

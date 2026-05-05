---
scope: backend, frontend, database
source: feature/permission-granularity
verified: "2026-05-04"
---

# Permission Code Conventions

## TECH-perm-001: Permission Code Split/Expand Migration Pattern

When permission codes are split or renamed, the migration follows this pattern:

1. **Idempotent DB migration**: A dedicated migration function (e.g., `MigratePermissionGranularity`) performs the code transformation within a single database transaction. The `schema_migrations` table tracks whether the migration has already run (by version string). Re-running the migration is a no-op.

2. **Atomic grants**: When a code is split (e.g., `user:read` splits into `user:list` + new `user:read`), both new codes are written atomically to avoid access gaps. The old code row is either preserved (if its semantics change) or deleted only after all replacement codes are inserted.

3. **CI grep assertion**: After the migration, a CI grep step confirms zero residual references to the old code string in the codebase. Only when the grep returns zero matches is the change allowed to merge.

**Why**: Permission code changes affect runtime authorization. A partially-applied migration could leave users without required permissions or grant unintended ones. Idempotency ensures safe re-runs after failures.

**Example**:
```
Migration step 1: user:manage_role -> INSERT role:create, role:update, role:delete; DELETE user:manage_role
Migration step 2: user:read -> INSERT user:list (keep user:read row, semantics updated by route binding)
Migration step 3: Mark version 'permission_granularity_001' in schema_migrations
CI grep: grep -r "user:manage_role" src/ backend/ --include="*.go" --include="*.ts" --include="*.tsx" must return 0 matches
```

## TECH-perm-002: schema_migrations Table for Idempotency

The `schema_migrations` table stores completed migration version strings. Before executing a migration, the function checks whether its version already exists in the table. If found, the migration is skipped.

**Why**: Prevents duplicate permission code inserts on application restart. The seeding function (`SyncPresetRoles`) also uses INSERT-IGNORE semantics, but migration logic needs explicit version tracking because it transforms existing data, not just inserts.

## TECH-perm-003: Migration Must Run Before SyncPresetRoles

In `main.go`, the migration function (e.g., `MigratePermissionGranularity`) must be called **before** `SyncPresetRoles`. If called after, `SyncPresetRoles` would write the old permission codes, and the migration would not find the expected old codes to transform.

**Why**: Startup order dependency. `SyncPresetRoles` writes preset role permissions from the current code definition. If migration runs after seeding, the seed may write old-format codes that the migration then incorrectly transforms or misses entirely.

## TECH-perm-004: CI Grep Assertion for Zero Residual Old Codes

After a permission code migration, the CI pipeline includes a grep assertion that scans all source files (`.go`, `.ts`, `.tsx`) for references to deprecated permission code strings. The assertion fails the build if any matches are found.

**Why**: Static analysis catches stale references that tests might miss — especially in UI components, comments, or config files that aren't exercised by unit tests. The grep is the final gate before merge.

---

## Permission Code Format and Registry

_Source: feature/rbac-permissions_

All permission codes follow `resource:action` format. Codes are defined in Go code (`internal/pkg/permissions/codes.go`), not configurable via API. The `Registry` map is the single source of truth.

**Why:** Codes couple to code logic (handler routing, middleware checks). Allowing runtime creation of new codes would create orphaned codes with no enforcement. PRD explicitly scopes out custom permission codes.

### Complete Permission Code List (29 codes)

| Resource | Code | Description |
|----------|------|-------------|
| team | `team:create` | 创建团队（全局权限，非团队上下文） |
| team | `team:read` | 查看团队信息 |
| team | `team:update` | 编辑团队信息 |
| team | `team:delete` | 解散团队 |
| team | `team:invite` | 邀请成员加入 |
| team | `team:remove` | 移除团队成员 |
| team | `team:transfer` | 转让PM身份 |
| main_item | `main_item:create` | 创建主事项 |
| main_item | `main_item:read` | 查看主事项 |
| main_item | `main_item:update` | 编辑主事项 |
| main_item | `main_item:archive` | 归档主事项 |
| main_item | `main_item:change_status` | 变更主事项状态 |
| sub_item | `sub_item:create` | 创建子事项 |
| sub_item | `sub_item:read` | 查看子事项 |
| sub_item | `sub_item:update` | 编辑子事项 |
| sub_item | `sub_item:assign` | 分配子事项负责人 |
| sub_item | `sub_item:change_status` | 变更子事项状态 |
| progress | `progress:create` | 追加进度记录 |
| progress | `progress:read` | 查看进度记录 |
| progress | `progress:update` | 修正进度记录 |
| item_pool | `item_pool:submit` | 提交事项到事项池 |
| item_pool | `item_pool:review` | 审核/分配/拒绝事项池事项 |
| view | `view:weekly` | 查看周视图 |
| view | `view:gantt` | 查看甘特图 |
| view | `view:table` | 查看表格视图 |
| report | `report:export` | 导出周报 |
| user | `user:read` | 查看用户信息 |
| user | `user:update` | 编辑用户信息 |
| user | `user:manage_role` | 管理角色定义 |

> Note: `sub_item:update`, `sub_item:change_status`, and `main_item:change_status` have additional assignee business rules at the handler/service layer.

### SuperAdmin Bypass Rule

SuperAdmin bypasses all team-level permission checks. Implementation: `RequirePermission` middleware checks `isSuperAdmin` from context first; if true, immediately calls `c.Next()`. SuperAdmin does NOT store codes in `role_permissions` -- bypass is via `User.IsSuperAdmin` boolean flag.

### Preset Roles (Immutable)

Three preset roles seeded at migration:

| Role | is_preset | Behavior |
|------|-----------|----------|
| superadmin | true | All codes (bypass via IsSuperAdmin flag). Cannot be edited or deleted. |
| pm | true | All team/main_item/sub_item/progress/item_pool/view/report codes + user:read. Can be edited by SuperAdmin. |
| member | true | Limited: main_item:read, sub_item:create/read/update/change_status, progress:create/read, item_pool:submit, view:weekly/table, report:export. Can be edited by SuperAdmin. |

### JWT Claims Structure

JWT contains minimal identity only -- no permissions or roles:

| Field | Type | Description |
|-------|------|-------------|
| user_id | number | User ID |
| username | string | User display name |
| iat | number | Issued-at timestamp |

Permissions fetched via `GET /api/v1/me/permissions` returning `{ is_superadmin: bool, team_permissions: { [teamId: string]: string[] } }`.

---
feature: "统一权限码校验，移除 bypass 模式"
status: Draft
---

# 统一权限码校验，移除 bypass 模式 — PRD Spec

> PRD Spec: defines WHAT the feature is and why it exists.

## Background

### Why (Reason)

RBAC 权限系统已上线，但存在两层 bypass 逻辑绕过权限码校验：
1. **SuperAdmin bypass**：`User.IsSuperAdmin` 布尔字段使 `RequirePermission` 中间件直接短路，跳过所有权限码检查。SuperAdmin 角色在 DB 中无任何权限码记录。
2. **Handler-level bypass**：`isPMOrSuperAdmin()` 函数在 sub_item_handler 和 progress_handler 中，根据 SuperAdmin 标志或 `sub_item:assign` 权限码决定是否跳过 assignee 所有权检查。这导致自定义角色（如 ext-member）即使拥有 `sub_item:update` 也可能被 403 拒绝。

team_handler 中还有 5 处通过 `IsSuperAdmin` 标志获取团队 PM BizKey 来绕过 PM 身份校验的逻辑。

### What (Target)

移除所有 bypass 逻辑，统一采用权限码校验：
- 删除 `User.IsSuperAdmin` 字段
- superadmin 预设角色获得全部 29 个权限码
- 所有权限判断统一通过权限码完成
- 前端同步移除 `isSuperAdmin` 引用

### Who (Users)

| 用户角色 | 影响 |
|---------|------|
| SuperAdmin | 无功能性变化，仍可执行所有操作（通过权限码而非硬编码短路） |
| PM | 无变化 |
| 自定义角色用户（如 ext-member） | **核心受益方** -- 拥有的权限码不再被 bypass 逻辑阻塞 |
| Member | 无变化 |

## Goals

| Goal | Metric | Notes |
|------|--------|-------|
| 自定义角色权限码生效 | 自定义角色拥有 `sub_item:update` 后编辑子事项成功率 100%（当前 0%） | 核心修复目标 |
| 消除 bypass 代码 | `IsSuperAdmin`/`isPMOrSuperAdmin` 引用数从 ~40 降为 0 | 后端 + 前端合计 |
| SuperAdmin 功能无损 | SuperAdmin 角色可执行全部 14 类操作，与移除前一致 | 回归验证 |

## Scope

### In Scope

- [x] 删除 `User.IsSuperAdmin` 字段（model、migration、VO/DTO）
- [x] superadmin 预设角色 seed 全部 29 个权限码
- [x] 移除 `RequirePermission` 中间件的 SuperAdmin 短路
- [x] 移除 `isPMOrSuperAdmin()` 函数及所有调用点（sub_item_handler、progress_handler）
- [x] 移除 sub_item_handler 中的 assignee 所有权检查
- [x] 替换 team_handler 中 5 处 "SuperAdmin 充当 PM" 逻辑为权限码驱动
- [x] TeamScopeMiddleware 为 superadmin 角色自动注入团队成员身份
- [x] 服务层（team_service 等）将 PM 身份校验改为权限码校验
- [x] DB migration：删除 `users.is_super_admin` 列，将原 SuperAdmin 用户绑定到 superadmin 角色
- [x] 前端：从 types、auth store、PermissionGuard、页面组件中移除 `isSuperAdmin`，改用权限码
- [x] 更新 `docs/conventions/permission-codes.md` 中的 SuperAdmin Bypass Rule 章节

### Out of Scope

- 新增权限码（当前 29 个够用）
- 角色管理 UI 的变更
- 数据范围（data scope）的变更（all_teams / own_teams / self_only 不变）
- member 角色权限调整
- `todos.txt` #65（member 角色获取权限问题）

## Flow Description

### Business Flow Description

#### 当前流程（有 bypass）

1. 用户发起请求 → AuthMiddleware 加载 User，设置 `isSuperAdmin` 到 context
2. TeamScopeMiddleware 检查：若 SuperAdmin → 注入空 permCodes，跳过团队成员检查
3. RequirePermission 检查：若 SuperAdmin → 直接 `c.Next()`，跳过权限码检查
4. Handler 中：`isPMOrSuperAdmin()` 检查 SuperAdmin 标志或 `sub_item:assign` → 决定是否跳过 assignee 检查
5. team_handler 中：SuperAdmin 获取团队 PM BizKey 绕过 PM 身份校验

自定义角色（ext-member）在步骤 4 被 403 拒绝：拥有 `sub_item:update` 但没有 `sub_item:assign`，且不是 assignee。

#### 目标流程（无 bypass）

1. 用户发起请求 → AuthMiddleware 加载 User（无 IsSuperAdmin 字段）
2. TeamScopeMiddleware 检查：若用户角色为 superadmin → 自动注入团队成员身份（无需 team_members 记录）；否则正常加载 permCodes
3. RequirePermission 检查：所有用户统一走 permCodes 线性扫描
4. Handler 中：不再有 `isPMOrSuperAdmin()` 调用，不再有 assignee 检查
5. team_handler 中：不再有 "充当 PM" 逻辑，服务层校验权限码而非 PM 身份

自定义角色（ext-member）在步骤 3 通过 `sub_item:update` 检查，步骤 4 不再有额外检查，请求成功。

### Business Flow Diagram

```mermaid
flowchart TD
    Start([用户发起 API 请求]) --> Auth[AuthMiddleware: 加载 User]
    Auth --> AuthValid{Token 有效?}
    AuthValid -->|No| Unauthorized[返回 401 UNAUTHORIZED]
    AuthValid -->|Yes| TeamScope[TeamScopeMiddleware]

    TeamScope --> IsSuperAdminRole{用户角色 = superadmin?}
    IsSuperAdminRole -->|Yes| AutoInject[自动注入团队成员身份 + 加载全部 permCodes]
    IsSuperAdminRole -->|No| IsTeamMember{是该团队成员?}
    IsTeamMember -->|No| TeamForbidden[返回 403 非团队成员]
    IsTeamMember -->|Yes| LoadPerms[从 DB 加载该用户角色的 permCodes]

    AutoInject --> PermCheck[RequirePermission: 检查 permCodes 是否包含所需权限码]
    LoadPerms --> PermCheck

    PermCheck --> HasCode{permCodes 包含所需码?}
    HasCode -->|No| Forbidden[返回 403 FORBIDDEN]
    HasCode -->|Yes| Handler[进入 Handler 执行业务逻辑]

    Handler --> Success[返回成功响应]
    Unauthorized --> End([End])
    TeamForbidden --> End
    Forbidden --> End
    Success --> End

    style AuthValid fill:#f9f,stroke:#333
    style IsSuperAdminRole fill:#f9f,stroke:#333
    style IsTeamMember fill:#f9f,stroke:#333
    style HasCode fill:#f9f,stroke:#333
```

## Functional Specs

### 5.1 DB Migration

| # | 变更项 | 详情 |
|---|--------|------|
| 1 | 删除 `users.is_super_admin` 列 | `ALTER TABLE users DROP COLUMN is_super_admin` |
| 2 | 绑定 SuperAdmin 用户到 superadmin 角色 | 查找原 `is_super_admin=true` 的用户，为其在每个已有团队中创建 `team_members` 记录并关联 superadmin 角色。事务保证原子性 |
| 3 | superadmin 角色 seed 全部 29 个权限码 | 在 `role_permissions` 表中为 superadmin 角色插入全部 29 条权限码记录 |

### 5.2 Backend Middleware Changes

| # | 文件 | 变更 |
|---|------|------|
| 1 | `middleware/auth.go` | 移除 `IsSuperAdmin()` 函数；AuthMiddleware 不再设置 `isSuperAdmin` 到 context |
| 2 | `middleware/permission.go` | `RequirePermission` 移除 SuperAdmin 短路：删除 `if IsSuperAdmin(c) { c.Next(); return }` |
| 3 | `middleware/team_scope.go` | SuperAdmin bypass 替换为角色检测：若用户在任何团队中的角色为 superadmin，自动注入团队成员身份。不再依赖 `isSuperAdmin` context 值 |

### 5.3 Backend Handler Changes

| # | 文件 | 变更 |
|---|------|------|
| 1 | `handler/sub_item_handler.go` | 删除 `isPMOrSuperAdmin()` 函数；Update 和 ChangeStatus handler 中移除 assignee 所有权检查 |
| 2 | `handler/progress_handler.go` | 移除 `pmFlag` 参数和 `isPMOrSuperAdmin()` 调用 |
| 3 | `handler/team_handler.go` L99 | Update: 移除 `IsSuperAdmin` 判断，改为依赖 `team:update` 权限码 |
| 4 | `handler/team_handler.go` L128 | Disband: 同上，依赖 `team:delete` 权限码 |
| 5 | `handler/team_handler.go` L194 | RemoveMember: 移除 `IsSuperAdmin` 判断，依赖 `team:remove` 权限码 |
| 6 | `handler/team_handler.go` L233 | UpdateMemberRole: 同上 |
| 7 | `handler/team_handler.go` L276 | TransferPM: 移除 `IsSuperAdmin` 判断，依赖 `team:transfer` 权限码 |
| 8 | `handler/team_handler.go` L56 | List: `isSuperAdmin` 参数移除，改为通过权限码判断是否可见所有团队 |

### 5.4 Backend Service Changes

| # | 文件 | 变更 |
|---|------|------|
| 1 | `service/team_service.go` | `ListTeams` 移除 `isSuperAdmin` 参数；PM 身份校验改为权限码校验（`team:update` 等） |
| 2 | `service/role_service.go` | `UserInfo` 移除 `IsSuperAdmin` 字段 |
| 3 | `model/user.go` | 移除 `IsSuperAdmin` 字段 |

### 5.5 Backend VO/DTO Changes

| # | 文件 | 变更 |
|---|------|------|
| 1 | `vo/user_vo.go` | 移除 `IsSuperAdmin` 字段 |
| 2 | `dto/auth.go` | 移除 `IsSuperAdmin` 字段 |
| 3 | `service/role_service.go` (UserInfo) | 移除 `IsSuperAdmin` 字段 |

### 5.6 Frontend Changes

| # | 文件/模块 | 变更 |
|---|-----------|------|
| 1 | `types/index.ts` | 移除 `isSuperAdmin` 字段 |
| 2 | `store/auth.ts` | 移除 `isSuperAdmin` 状态和相关逻辑 |
| 3 | `components/PermissionGuard` | 移除 `isSuperAdmin` 快捷路径，统一使用 `hasPermission()` |
| 4 | 页面组件（25 个文件） | 移除所有 `isSuperAdmin` 引用，改用 `hasPermission()` |
| 5 | `mocks/handlers.ts` | 移除 `isSuperAdmin` mock 数据 |

### 5.7 Related Changes

| # | Project | Module | Change Point | Updated Logic |
|---|---------|--------|-------------|---------------|
| 1 | backend | `config/seed.go` | 创建初始 admin 用户 | 移除 `IsSuperAdmin: true`，改为创建 team_members 记录关联 superadmin 角色 |
| 2 | backend | `migration/rbac.go` | seedPresetRoles | superadmin 角色写入全部 29 个权限码 |
| 3 | docs | `conventions/permission-codes.md` | SuperAdmin Bypass Rule 章节 | 删除 bypass 描述，改为 "superadmin 角色拥有全部权限码" |

## Other Notes

### Performance Requirements

- 权限码检查仍为 permCodes 切片线性扫描（当前方案，29 个码足够小），无需变更
- TeamScopeMiddleware 对 superadmin 角色的自动注入不应增加 DB 查询

### Data Requirements

- **Migration 必须是事务性的**：`is_super_admin` 列删除和 team_members 记录创建在同一事务中
- **Seed 数据更新**：superadmin 角色的权限码记录在 migration 中创建
- **SQLite 和 MySQL schema 同步**：两份 schema 文件均需移除 `is_super_admin` 列

### Failure Scenarios

| 场景 | 影响 | 应对策略 |
|------|------|----------|
| Migration 部分失败（列已删除但 team_members 记录未创建） | SuperAdmin 用户无法通过权限码校验，所有操作返回 403 | Migration 使用单事务包裹：`DROP COLUMN` 和 `INSERT INTO team_members` 在同一事务中，失败时整体回滚。回滚后系统保持原状态（IsSuperAdmin 字段仍存在） |
| 部署期间并发请求 | 后端已部署新代码但 migration 尚未执行，或 migration 已执行但前端仍发送 `isSuperAdmin` 参数 | 部署顺序：先执行 DB migration → 再部署后端 → 最后部署前端。Migration 是向后兼容的（新代码不依赖旧字段），前端发送多余字段被后端忽略 |
| 孤儿 SuperAdmin 用户（`is_super_admin=true` 但在 users 表中无有效记录） | Migration 查找不到对应用户 | Migration 在事务开头校验：`SELECT id FROM users WHERE is_super_admin = true AND deleted_at IS NULL`，若结果集为空则跳过 team_members 插入步骤并记录 warning 日志 |

### Monitoring Requirements

- **Rollout 期间 403 率基线监控**：部署后 1 小时内，对比同一时间窗口的 403 响应数量与部署前 7 天平均值。若 403 率增长超过 20%，触发回滚评估
- 监控指标来源：现有 API 日志中的 HTTP 状态码统计，无需新增埋点

### Security Requirements

- 权限码校验统一后，SuperAdmin 不再有代码层面的 "超级权限" -- 全部通过 DB 中的权限码控制
- superadmin 角色的权限码为 seed 数据，不可通过 API 修改（is_preset = true）
- 若需要撤销 SuperAdmin 的某些权限，可直接从 `role_permissions` 中删除对应记录

---

## Quality Checklist

- [x] Is the requirement title accurate and descriptive
- [x] Does the background include all three elements: reason, target, users
- [x] Are the goals quantified
- [x] Is the flow description complete
- [x] Does the business flow diagram exist (Mermaid format)
- [x] Is the list page description complete — N/A (no list page changes)
- [x] Are button actions described completely — N/A (no new buttons)
- [x] Are form descriptions complete — N/A (no new forms)
- [x] Are related changes thoroughly analyzed
- [x] Are non-functional requirements considered (performance / data / monitoring / security)
- [x] Are all tables filled completely
- [x] Is there any ambiguous or vague wording
- [x] Is the spec actionable and verifiable

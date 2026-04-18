---
created: 2026-04-18
author: "faner"
status: Draft
---

# Proposal: 扩展用户与权限体系 — 标准 RBAC + 三级数据范围

## Problem

当前系统的权限体系有三个结构性问题：

1. **角色硬编码**：`User` 模型用 `IsSuperAdmin` 和 `CanCreateTeam` 两个布尔标记控制权限，无法扩展。新增任何权限都需要改模型、改中间件、发版。
2. **无权限实体**：操作权限散落在中间件 `RequireRole` / `RequireTeamRole` 和 handler 的 if-else 中，前端对用户能做什么毫无感知，无法按权限渲染 UI。
3. **数据权限隐式**：team_scope 中间件硬编码 "PM 看本队、superadmin 看全部"，无显式配置，无法灵活调整数据可见范围。

这导致：管理员无法在线调整权限策略，前端按钮/菜单无法按用户能力动态展示，新增功能时权限代码到处散落。

## Proposed Solution

引入标准 RBAC（Role-Based Access Control）体系，配合三级数据范围控制：

### 权限模型：资源×操作矩阵

权限码由代码定义（常量/枚举），是权限的**唯一真实来源**。数据库只存储角色→权限码的绑定关系。

**职责划分：**

| 职责 | 执行者 | 说明 |
|------|--------|------|
| 定义权限码 | 开发者 | 在代码中声明常量，新增功能时同步新增 |
| 绑定权限码到路由 | 开发者 | 路由注册时用 `RequirePermission("team:create")` 声明 |
| 创建角色、勾选权限 | 管理员 | 从系统定义的权限列表中选择，不能凭空创建权限码 |
| 分配角色给用户 | 管理员 | 在线操作，即时生效 |

管理员管理的是"角色"（哪些权限组合），而非"权限码"本身。新增权限码需要开发者改代码发版。

权限按「资源:操作」格式定义，覆盖当前系统的 8 个核心资源：

| 资源 | 操作 | 示例权限码 |
|------|------|-----------|
| team | create, read, update, delete, invite, remove, transfer, dissolve | `team:create` |
| main_item | create, read, update, archive | `main_item:create` |
| sub_item | create, read, update, assign, change_status | `sub_item:assign` |
| progress | create, read, update | `progress:create` |
| item_pool | submit, review, assign, reject | `item_pool:review` |
| view | weekly, gantt, table | `view:gantt` |
| report | export | `report:export` |
| user | read, update, manage_role, manage_permission | `user:manage_role` |

### 角色管理

- 管理员可在线创建/编辑/删除角色
- 每个角色勾选一组权限码
- 系统预置三个默认角色：superadmin（全部权限）、pm（团队管理权限）、member（基础操作权限）
- 一个用户绑定一个全局角色

### 数据权限：三级数据范围

每个角色附带一个数据范围属性，控制可见数据边界：

| 数据范围 | 说明 | 适用角色 |
|---------|------|---------|
| all_teams | 可查看所有团队数据 | superadmin |
| own_teams | 仅查看所属团队的数据 | pm, member |
| self_only | 仅查看自己相关的数据 | 受限成员 |

### 前端权限渲染

- 登录后获取用户的权限列表和数据范围
- 按权限动态显示/隐藏按钮、菜单项、操作入口
- 管理员页面提供角色 CRUD 和权限勾选界面

### 数据迁移

- 迁移脚本将现有 `IsSuperAdmin=true` 用户映射到 superadmin 角色
- `CanCreateTeam=true` 用户映射到包含 `team:create` 权限的角色
- 移除 User 模型上的 `IsSuperAdmin` 和 `CanCreateTeam` 字段

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| **方案 A：标准 RBAC（本方案）** | 权限码由代码定义、角色在线可配，职责清晰 | 新增权限码需开发者改代码发版 | ✅ 采用 |
| 方案 B：权限码也存数据库 | 管理员可完全在线创建权限码，无需发版 | 权限码与路由绑定脱节，容易拼写错误和遗漏 | ❌ 放弃 |
| 方案 C：直接赋权（无角色） | 最灵活，每个用户独立权限 | 用户量大时管理困难，无法批量调整 | ❌ 放弃 |

## Scope

### In Scope

- 权限模型设计（资源×操作矩阵，权限码由代码定义，角色-权限绑定存数据库）
- 角色管理（CRUD 角色，为角色分配权限）
- 用户-角色绑定（为用户分配角色）
- 三级数据范围（all_teams / own_teams / self_only）
- 后端 API：角色管理接口、权限校验中间件改造
- 前端：权限渲染（按钮/菜单按权限显示/隐藏）
- 前端：管理员页面（角色管理、权限分配）
- 数据迁移：从布尔标记迁移到 RBAC
- JWT claims 更新（携带权限列表和数据范围）

### Out of Scope

- 用户自助注册 / 改密码 / 个人信息编辑（用户生命周期管理）
- 审计日志（权限变更记录，后续版本）
- 组织/部门层级结构
- 团队级权限作用域（本次为全局统一权限）
- 权限继承或角色层级

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 迁移脚本对现有用户数据破坏 | Low | High | 迁移前备份，迁移脚本在事务中执行，提供回滚脚本 |
| 权限矩阵遗漏导致功能不可用 | Medium | Medium | 对照现有 PRD 逐项梳理权限，测试覆盖所有角色×操作组合 |
| 前端权限遗漏导致越权操作 | Medium | High | 前后端双重校验，前端隐藏只是 UX 优化，后端中间件是安全防线 |
| 预置角色权限调整影响现有用户 | Low | Medium | 预置角色权限变更需通过迁移脚本，不直接修改 |

## Success Criteria

- [ ] 管理员可通过 UI 创建/编辑/删除角色并分配权限
- [ ] 前端按钮和菜单根据用户权限动态显示/隐藏
- [ ] 现有用户数据无缝迁移，迁移后行为与迁移前一致
- [ ] 后端中间件基于权限码而非角色名进行访问控制
- [ ] `User` 模型上不再有 `IsSuperAdmin` / `CanCreateTeam` 布尔字段
- [ ] JWT token 包含用户权限列表和数据范围

## Next Steps

- Proceed to `/write-prd` to formalize requirements

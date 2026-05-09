---
feature: "统一权限码校验，移除 bypass 模式"
---

# User Stories: 统一权限码校验，移除 bypass 模式

## Story 1: 自定义角色用户编辑子事项

**As a** 自定义角色用户（如 ext-member）
**I want to** 使用被授予的 `sub_item:update` 权限码编辑任何子事项
**So that** 我能完成分配给我的工作，而不是因为 assignee 检查被 403 拒绝

**Acceptance Criteria:**
- Given 一个自定义角色拥有 `sub_item:update` 权限码
- When 该角色用户发送 `PUT /teams/:teamId/sub-items/:subId` 编辑一个非自己负责的子事项
- Then 返回 200，编辑成功（不再触发 assignee 所有权检查）

- Given 一个自定义角色**未拥有** `sub_item:update` 权限码（仅有 `sub_item:view`）
- When 该角色用户发送 `PUT /teams/:teamId/sub-items/:subId` 编辑子事项
- Then 返回 403 FORBIDDEN，响应体包含权限不足提示（bypass 移除后权限拒绝仍正确生效）

---

## Story 2: 自定义角色用户变更子事项状态

**As a** 自定义角色用户（如 ext-member）
**I want to** 使用被授予的 `sub_item:change_status` 权限码变更任何子事项的状态
**So that** 我能推进工作流，而不是被 assignee 检查阻塞

**Acceptance Criteria:**
- Given 一个自定义角色拥有 `sub_item:change_status` 权限码
- When 该角色用户发送 `PUT /teams/:teamId/sub-items/:subId/status` 变更一个非自己负责的子事项状态
- Then 返回 200，状态变更成功（不再触发 assignee 所有权检查）

- Given 一个自定义角色**未拥有** `sub_item:change_status` 权限码
- When 该角色用户发送 `PUT /teams/:teamId/sub-items/:subId/status` 变更子事项状态
- Then 返回 403 FORBIDDEN（权限拒绝在 bypass 移除后仍正确生效）

---

## Story 3: SuperAdmin 功能无损

**As a** SuperAdmin 用户
**I want to** 在 bypass 移除后，通过加载全部 29 个权限码执行所有操作
**So that** 我的工作流不受影响，校验路径与其他用户统一

**Acceptance Criteria:**

**AC 3a — 团队管理操作:**
- Given 一个 SuperAdmin 用户（`is_super_admin=true`）
- When 该用户依次执行以下操作：创建团队（201）、更新团队信息（200）、邀请成员（200）、修改成员角色（200）、移除成员（200）、转让 PM（200）、解散团队（200）
- Then 每项操作返回括号中标注的 HTTP 状态码（TeamScopeMiddleware 注入全部 29 码，RequirePermission 统一走 permCodes 检查）

**AC 3b — 事项与子事项操作:**
- Given 一个 SuperAdmin 用户
- When 该用户依次执行以下操作：创建主事项（201）、编辑主事项（200）、归档主事项（200）、创建子事项（201）、编辑子事项（200）、分配子事项负责人（200）、变更子事项状态（200）
- Then 每项操作返回括号中标注的 HTTP 状态码

**AC 3c — 其他操作:**
- Given 一个 SuperAdmin 用户
- When 该用户依次执行以下操作：添加进度（201）、提交待办事项（201）、审核待办事项（200）、查看周报（200）、查看甘特图（200）、查看表格视图（200）、导出报表（200）、查看用户列表（200）
- Then 每项操作返回括号中标注的 HTTP 状态码

---

## Story 4: SuperAdmin 跨团队访问资源

**As a** SuperAdmin 用户
**I want to** 访问任何团队的资源（即使我不是该团队的显式成员）
**So that** 我能管理系统中所有团队而无需逐个加入

**Acceptance Criteria:**
- Given 一个 SuperAdmin 用户（`is_super_admin=true`），且该用户在 `team_members` 表中无该团队的记录
- When 该用户发送 `GET /teams/:teamId/main-items` 访问该团队资源
- Then 返回 200，TeamScopeMiddleware 为 SuperAdmin 注入全部 29 个权限码（跳过团队成员检查），不返回 403

- Given 一个普通用户（非 SuperAdmin）不是该团队成员
- When 该用户发送 `GET /teams/:teamId/main-items` 访问该团队资源
- Then 返回 403 FORBIDDEN，TeamScopeMiddleware 拒绝非成员访问（跨团队访问仅限 SuperAdmin）

---

## Story 5: 前端移除 isSuperAdmin 依赖

**As a** 前端开发者
**I want to** 所有权限相关逻辑统一使用 `hasPermission()` 而非 `isSuperAdmin` 布尔值
**So that** UI 可见性由权限码驱动，不需要维护 SuperAdmin 特殊路径

**Acceptance Criteria:**

- Given 代码库中不再存在 `isSuperAdmin` 引用
- When TypeScript 编译和所有测试运行
- Then 编译通过且所有测试通过

- Given 一个 SuperAdmin 用户登录后
- When 前端通过 `/api/v1/me/permissions` 获取权限码列表
- Then 返回包含全部 29 个权限码（`GetUserPermissions` SuperAdmin 路径），所有基于权限码的 UI 元素（导航菜单、操作按钮）可见性与变更前一致

---

## Story 6: PM 团队管理操作不受影响

**As a** PM 用户
**I want to** 在服务层移除 PM 身份校验后，仍能正常管理团队（邀请成员、移除成员、转让 PM）
**So that** 服务层变更不会中断 PM 的日常工作流

**Acceptance Criteria:**
- Given 一个用户在团队中的角色为 PM（pm 预设角色持有 `team:invite`、`team:remove`、`team:transfer` 权限码）
- When 该用户依次执行邀请成员（200）、移除成员（200）、转让 PM（200）
- Then 每项操作返回括号中标注的 HTTP 状态码（middleware `RequirePermission` 校验权限码，服务层不再做 PM 身份校验）

- Given 一个用户在团队中的角色为 PM
- When 该用户执行团队管理操作，且服务层已移除 PM 身份校验
- Then 操作成功，行为与移除前一致

---

## Story 7: 自定义角色用户添加进度

**As a** 自定义角色用户（如 ext-member）
**I want to** 使用被授予的 `progress:create` 权限码为任何子事项添加进度
**So that** 我能更新工作进展，而不是被 `isPMOrSuperAdmin()` 的 pmFlag 检查阻塞

**Acceptance Criteria:**
- Given 一个自定义角色拥有 `progress:create` 权限码
- When 该角色用户发送 `POST /teams/:teamId/sub-items/:subId/progress` 为非自己负责的子事项添加进度
- Then 返回 200，进度创建成功（不再触发 `isPMOrSuperAdmin()` 检查）

- Given 一个自定义角色**未拥有** `progress:create` 权限码
- When 该角色用户发送 `POST /teams/:teamId/sub-items/:subId/progress` 添加进度
- Then 返回 403 FORBIDDEN（权限拒绝在 `isPMOrSuperAdmin()` 移除后仍正确生效）

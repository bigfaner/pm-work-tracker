---
feature: "rbac-permissions"
status: Draft
---

# RBAC 权限体系 — PRD Spec

> PRD Spec: defines WHAT the feature is and why it exists.

## 需求背景

### 为什么做（原因）

当前系统权限体系存在三个结构性问题：

1. **角色硬编码**：`User` 模型用 `IsSuperAdmin` 和 `CanCreateTeam` 两个布尔标记控制权限。新增权限需改模型、改中间件、发版。
2. **权限散落**：操作权限分散在中间件 `RequireRole`/`RequireTeamRole` 和三个 handler 中重复的 `isPMOrSuperAdmin` 方法中。前端对用户能力无感知，无法按权限渲染 UI。
3. **角色名绑定**：权限检查依赖角色名字符串（"pm"/"member"/"superadmin"），而非权限能力码。PM 拥有的具体能力无法灵活组合或拆分。

这导致：管理员无法在线调整权限策略，前端按钮/菜单无法按用户能力动态展示，新增功能时权限代码到处散落。

### 要做什么（对象）

引入标准 RBAC 权限体系。权限码由代码定义（`resource:action` 格式），角色在线可配（由管理员在 UI 上创建/编辑）。角色在邀请用户进入团队时绑定，用户在不同团队可拥有不同角色。SuperAdmin 是系统预置角色，独立于团队，绕过所有团队级权限检查。

### 用户是谁（人员）

| 角色 | 说明 | 核心诉求 |
|------|------|---------|
| 超级管理员 | 系统预置最高权限角色，独立于团队 | 管理角色定义、用户权限分配，查看所有团队 |
| PM（团队负责人） | 团队内拥有管理权限的角色 | 管理团队、分配工作、查看团队进度 |
| 团队成员 | 团队内基础操作角色 | 提交事项、更新进度、查看团队数据 |

## 需求目标

| 目标 | 量化指标 | 说明 |
|------|----------|------|
| 权限可在线调整 | 角色编辑后即时生效 | 管理员从系统定义的权限列表勾选，无需开发者改代码 |
| 前端权限感知 | 100% 按钮/菜单按权限渲染 | 登录后获取各团队权限码列表，动态显示/隐藏 UI 元素 |
| 数据迁移无损 | 迁移后所有用户行为与迁移前一致 | 迁移脚本在事务中执行 |
| 权限检查统一 | 消除散落的 if-else 和重复的 `isPMOrSuperAdmin` | 中间件统一基于权限码校验 |

## Scope

### In Scope

- 权限模型（`resource:action` 权限码，代码定义，角色-权限绑定存数据库）
- 角色管理（superadmin 创建/编辑/删除角色，为角色分配权限码）
- 团队级角色绑定（邀请用户进团队时指定角色，用户可拥有多个团队的不同角色）
- SuperAdmin 系统预置角色（不可编辑/删除，绕过团队级检查）
- 权限校验中间件改造（基于权限码而非角色名）
- 前端权限渲染（按钮/菜单按权限显示/隐藏）
- 前端角色管理页面
- 数据迁移
- 预置默认角色（superadmin、pm、member）

### Out of Scope

- 用户自助注册 / 改密码 / 个人信息编辑
- 审计日志（权限变更记录，后续版本）
- 组织/部门层级结构
- 权限继承或角色层级
- 自定义权限码（管理员不能创建新权限码，只能从系统定义的权限列表中勾选）
- 显式数据范围控制（all_teams/own_teams/self_only）

## 流程说明

### 业务流程说明

系统引入两层角色模型：全局预置超级管理员和团队级角色。

**超级管理员流程**：超级管理员是系统预置角色，独立于团队。可创建/编辑/删除角色定义，为角色分配权限码，查看所有团队数据。SuperAdmin 不可编辑其权限配置、不可删除。

**团队角色流程**：PM 或超级管理员邀请用户进入团队时，从系统定义的角色中选择一个角色赋予该用户。用户在不同团队可拥有不同角色。用户在团队中的可见数据范围由团队成员身份自然界定——只能看到所属团队的数据，权限码控制在团队内能执行的操作。

**权限检查流程**：后端中间件基于权限码（如 `team:invite`）而非角色名进行访问控制。SuperAdmin 角色绕过所有团队级权限检查。前端根据权限列表动态渲染 UI。

### 角色管理流程图

```mermaid
flowchart TD
    A[超级管理员登录] --> B[进入角色管理页面]
    B --> C{操作类型}
    C -->|创建角色| D[填写角色名称和描述]
    D --> E[从权限列表中勾选权限]
    E --> F[保存角色]
    F --> F1{名称是否重复}
    F1 -->|是| F2[提示角色名称已存在]
    F2 --> D
    F1 -->|否| M[角色定义生效]
    C -->|编辑角色| G[选择已有角色]
    G --> H[调整权限勾选]
    H --> F
    C -->|删除角色| I[选择已有角色]
    I --> J{是否有用户使用该角色}
    J -->|是| K[提示无法删除，需先迁移用户]
    J -->|否| L[确认删除]
    L --> M
```

### 团队角色分配流程图

```mermaid
flowchart TD
    A[PM 或超级管理员邀请成员] --> B[输入用户账号搜索]
    B --> B1{搜索结果}
    B1 -->|无匹配| B2[提示未找到用户]
    B2 --> B
    B1 -->|有结果| C[选择用户]
    C --> D[从角色列表中选择角色]
    D --> E[确认邀请]
    E --> E1{用户是否已在团队中}
    E1 -->|是| E2[提示用户已在团队中]
    E1 -->|否| F[用户加入团队，角色生效]
    F --> G[前端根据新角色刷新权限]

    H[PM 管理团队成员] --> I[查看团队成员列表]
    I --> J{操作类型}
    J -->|变更角色| K[选择新角色]
    K --> L[确认变更]
    L --> G
    J -->|移除成员| M[确认移除]
    M --> N[用户退出团队，权限失效]
```

### 权限检查流程图

```mermaid
flowchart TD
    A[用户请求 API] --> B[中间件提取 JWT user_id]
    B --> C{路由需要权限码?}
    C -->|否| J[放行]
    C -->|是| D{是否为 superadmin 角色?}
    D -->|是| J
    D -->|否| E{请求属于团队上下文?}
    E -->|否| F{用户任意团队角色包含该权限码?}
    F -->|是| J
    F -->|否| G[返回 403]
    E -->|是| H{用户在该团队的角色包含该权限码?}
    H -->|是| J
    H -->|否| G
```

## 功能描述

### 5.1 角色管理（超级管理员）

**数据来源**：系统数据库中的角色表。

**数据权限**：仅超级管理员可访问。

**页面类型**：列表页 + 表单页

**列表字段**：

| 字段名称 | 类型 | 说明 |
|---------|------|------|
| 角色名称 | string | 角色的显示名称 |
| 描述 | string | 角色的功能描述 |
| 权限数量 | number | 该角色绑定的权限码数量 |
| 使用人数 | number | 使用该角色的团队成员数量 |
| 是否预置 | boolean | 系统预置角色不可删除 |
| 创建时间 | datetime | 角色创建时间 |

**排序方式**：默认按创建时间升序。

**翻页设置**：每页 20 条，支持翻页。

**搜索条件**：

| 序号 | 搜索项 | 控件类型 | 说明 | 默认提示 |
|------|--------|----------|------|----------|
| 1 | 角色名称 | 输入框 | 模糊匹配 | 搜索角色名称 |
| 2 | 是否预置 | 下拉单选 | 全部 / 预置 / 自定义 | 全部 |

**功能说明**：

| 功能 | 说明 | 可操作角色 |
|------|------|-----------|
| 查看角色列表 | 查看所有系统角色及其权限配置 | 超级管理员 |
| 创建角色 | 填写名称、描述，从权限列表勾选权限 | 超级管理员 |
| 编辑角色 | 修改角色名称、描述、权限勾选（superadmin 不可编辑，pm/member 可编辑） | 超级管理员 |
| 删除角色 | 删除自定义角色（预置角色不可删除，有用户的角色不可删除） | 超级管理员 |
| 查看权限列表 | 查看系统所有可用权限码（只读，按资源分组展示） | 超级管理员 |

**表单字段（创建/编辑角色）**：

| 字段名称 | 控件类型 | 必填 | 规则说明 |
|---------|----------|------|----------|
| 角色名称 | 单行文本 | 是 | 2-50 字符，不可与已有角色重名 |
| 描述 | 多行文本 | 否 | 最多 200 字符 |
| 权限勾选 | 复选框组 | 至少 1 个 | 按资源分组展示权限码，可多选 |

**权限码矩阵（系统定义，按资源分组）**：

| 资源 | 操作 | 权限码 | 当前对应的后端检查 | 说明 |
|------|------|--------|-------------------|------|
| team | create | `team:create` | `CanCreateTeam` flag / SuperAdmin | 创建团队（全局权限，非团队上下文） |
| team | read | `team:read` | Team membership | 查看团队信息 |
| team | update | `team:update` | `RequireTeamRole("pm")` | 编辑团队信息 |
| team | delete | `team:delete` | `RequireTeamRole("pm")` | 解散团队 |
| team | invite | `team:invite` | `RequireTeamRole("pm")` | 邀请成员加入 |
| team | remove | `team:remove` | `RequireTeamRole("pm")` | 移除团队成员 |
| team | transfer | `team:transfer` | `RequireTeamRole("pm")` | 转让 PM 身份 |
| main_item | create | `main_item:create` | `isPMOrSuperAdmin` | 创建主事项 |
| main_item | read | `main_item:read` | Any team member | 查看主事项 |
| main_item | update | `main_item:update` | `isPMOrSuperAdmin` | 编辑主事项 |
| main_item | archive | `main_item:archive` | `isPMOrSuperAdmin` | 归档主事项 |
| sub_item | create | `sub_item:create` | Any team member | 创建子事项 |
| sub_item | read | `sub_item:read` | Any team member | 查看子事项 |
| sub_item | update | `sub_item:update` | `isPMOrSuperAdmin` / assignee | 编辑子事项（PM 或被分配者） |
| sub_item | assign | `sub_item:assign` | `isPMOrSuperAdmin` | 分配子事项负责人 |
| sub_item | change_status | `sub_item:change_status` | `isPMOrSuperAdmin` / assignee | 变更子事项状态 |
| progress | create | `progress:create` | Any team member | 追加进度记录 |
| progress | read | `progress:read` | Any team member | 查看进度记录 |
| progress | update | `progress:update` | `isPMOrSuperAdmin` | 修正进度记录 |
| item_pool | submit | `item_pool:submit` | Any team member | 提交事项到事项池 |
| item_pool | review | `item_pool:review` | `RequireTeamRole("pm")` | 审核/分配/拒绝事项池事项 |
| view | weekly | `view:weekly` | Any team member | 查看周视图 |
| view | gantt | `view:gantt` | Any team member | 查看甘特图 |
| view | table | `view:table` | Any team member | 查看表格视图 |
| report | export | `report:export` | Any team member | 导出周报 |
| user | read | `user:read` | `RequireRole("superadmin")` | 查看用户信息 |
| user | update | `user:update` | `RequireRole("superadmin")` | 编辑用户信息 |
| user | manage_role | `user:manage_role` | `RequireRole("superadmin")` | 管理角色定义 |

> **注**：`sub_item:update` 和 `sub_item:change_status` 有额外的业务规则——被分配者可操作自己负责的子事项，即使没有 PM 级别权限。这是在服务层处理的业务规则，不影响权限码定义。

### 5.2 团队成员角色管理

**数据来源**：team_members 表（含角色关联）。

**数据权限**：PM 可管理本团队成员角色，超级管理员可管理所有团队。

**功能说明**：

| 功能 | 说明 | 可操作角色 |
|------|------|-----------|
| 邀请成员 | 搜索用户并邀请，指定角色 | PM / 超级管理员 |
| 变更成员角色 | 修改团队成员的角色 | PM / 超级管理员 |
| 移除成员 | 将成员移出团队 | PM / 超级管理员 |
| 查看成员角色 | 查看团队成员及其角色 | 团队内所有成员 |

**邀请成员表单字段**：

| 字段名称 | 控件类型 | 必填 | 规则说明 |
|---------|----------|------|----------|
| 用户账号 | 搜索框 | 是 | 输入关键字搜索用户 |
| 角色 | 下拉选择 | 是 | 从系统定义的角色列表中选择（排除 superadmin） |

### 5.3 前端权限渲染

| 功能 | 说明 |
|------|------|
| 权限列表获取 | 登录后从后端获取用户在每个团队的权限码列表 |
| 按钮控制 | 根据权限码显示/隐藏操作按钮 |
| 菜单控制 | 根据权限码显示/隐藏菜单项 |
| 页面控制 | 无权限访问的页面显示提示或重定向 |
| 权限变更刷新 | 角色变更后前端自动刷新权限列表 |

### 5.4 预置角色定义

系统预置三个默认角色：

| 角色 | 权限码 | 说明 |
|------|--------|------|
| superadmin | 全部权限码 | 系统预置，独立于团队，绕过团队级检查，不可编辑/删除 |
| pm | team:create, team:read, team:update, team:delete, team:invite, team:remove, team:transfer, main_item:\*, sub_item:\*, progress:\*, item_pool:\*, view:\*, report:\*, user:read | 团队管理权限，可管理团队、事项、进度 |
| member | main_item:read, sub_item:create, sub_item:read, sub_item:update, sub_item:change_status, progress:create, progress:read, item_pool:submit, view:weekly, view:table, report:\* | 基础操作权限 |

### 5.5 数据迁移

| 迁移项 | 源 | 目标 | 规则 |
|--------|----|------|------|
| 超级管理员 | users.is_super_admin = true | 绑定 superadmin 预置角色 | 直接映射 |
| 创建团队权限 | users.can_create_team = true | pm 角色（含 team:create）或自定义含 team:create 的角色 | 映射到角色 |
| 团队 PM | team_members.role = "pm" | team_members 角色绑定 pm 预置角色 | 直接映射 |
| 团队成员 | team_members.role = "member" | team_members 角色绑定 member 预置角色 | 直接映射 |
| 布尔字段移除 | users.is_super_admin, users.can_create_team | 删除 | 迁移完成后移除 |

**迁移执行要求**：
- 迁移脚本在数据库事务中执行，任何步骤失败则整体回滚
- 回滚后系统保持原有行为，可排查错误后重新执行迁移
- 迁移脚本需幂等——重复执行不会产生重复数据或错误

### 5.6 JWT Claims 与权限获取

**JWT Claims（仅含最小标识）：**

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | number | 用户 ID |
| username | string | 用户名称 |
| iat | number | Token 签发时间 |

> JWT 不包含 `is_super_admin` 或 `role` 等权限相关字段。权限信息通过 API 实时获取，避免信息泄露和权限撤销延迟。

**权限获取方式：**

| 场景 | 方式 | 说明 |
|------|------|------|
| 前端渲染 | `GET /api/me/permissions` | 返回 { is_superadmin, team_roles: { team_id: [codes] } } |
| 后端权限检查 | 中间件从数据库/缓存实时查询 | 每次请求验证实际权限 |

### 5.7 关联改动

| 序号 | 功能模块 | 改动点 | 更改后逻辑 |
|------|----------|--------|-----------|
| 1 | 团队管理 | 邀请成员流程 | 增加角色选择步骤（当前只有 PM/member 选择） |
| 2 | 认证中间件 | RequireRole / RequireTeamRole / isPMOrSuperAdmin | 替换为 RequirePermission(permission_code) |
| 3 | 前端导航 | 菜单/按钮渲染 | 从 `/api/me/permissions` 获取权限码，按权限渲染 |
| 4 | 用户管理 | admin 接口权限 | 从 RequireRole("superadmin") 改为 RequirePermission("user:manage_role") |
| 5 | 团队创建 | CanCreateTeam 检查 | 从布尔标记改为 RequirePermission("team:create") |
| 6 | 前端 Store | auth store | 移除 isSuperAdmin/canCreateTeam，改为权限码列表 |

## 其他说明

### 性能需求

- 权限检查响应时间：< 10ms（中间件级别，使用缓存）
- 角色管理页面加载：< 2 秒
- 权限 API 响应时间：< 200ms（`/api/me/permissions`）
- 兼容性：支持主流现代浏览器（Chrome、Edge、Safari 最新两个版本）

### 数据需求

- 角色定义持久化到数据库
- 团队成员-角色绑定持久化到 team_members 表
- 预置角色通过数据库种子数据初始化
- 迁移脚本在事务中执行，提供回滚方案

### 安全性需求

- 后端中间件是权限安全的最终防线，前端隐藏仅为 UX 优化
- 权限码由代码定义，不可通过 API 创建新的权限码
- 角色管理接口仅超级管理员可访问
- JWT 不携带权限标记，仅用于身份识别
- 后端权限检查从数据库/缓存实时查询，不依赖 JWT 中的字段
- 权限变更即时生效（不依赖 token 过期周期）

---

## 质量检查

- [x] 需求标题是否概括准确
- [x] 需求背景是否包含原因、对象、人员三要素
- [x] 需求目标是否量化
- [x] 流程说明是否完整
- [x] 业务流程图是否包含（Mermaid 格式）
- [x] 列表页描述是否完整
- [x] 按钮描述是否完整
- [x] 表单描述是否完整
- [x] 关联性需求是否全面分析
- [x] 非功能性需求（性能/数据/安全）是否考虑
- [x] 所有表格是否填写完整

---
feature: "rbac-permissions"
---

# RBAC 权限体系 — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

- 角色管理页面（超级管理员专用）
- 权限码浏览视图（超级管理员专用）
- 团队成员角色选择（邀请流程中的角色选择步骤）
- 权限驱动的 UI 渲染（按钮/菜单/页面的条件显示）

## UI Function 1: 角色列表页

### Description

超级管理员查看所有系统角色，支持创建、编辑、删除角色。

### User Interaction Flow

1. 超级管理员点击导航中的"角色管理"
2. 系统展示角色列表（名称、描述、权限数量、使用人数、是否预置）
3. 点击"创建角色" → 进入角色编辑表单
4. 点击角色行 → 进入角色详情/编辑
5. 点击"删除"（仅自定义角色、无用户使用时可用）→ 确认弹窗 → 删除

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 角色名称 | string | roles 表 | |
| 描述 | string | roles 表 | |
| 权限数量 | number | role_permissions 关联计数 | |
| 使用人数 | number | team_members 关联计数 | |
| 是否预置 | boolean | roles.is_preset | 预置角色不可删除 |
| 创建时间 | datetime | roles.created_at | |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Loading | 骨架屏/加载动画 | 页面首次加载 |
| Populated | 角色列表表格 | 数据加载完成 |
| Empty | 空状态提示"暂无角色，请创建第一个角色" | 系统初始化未完成 |
| Error | 错误提示 | 加载失败 |

### Validation Rules

- 预置角色（superadmin、pm、member）不可删除，删除按钮置灰
- 有用户使用的角色不可删除，删除按钮置灰并显示 Tooltip 提示

---

## UI Function 2: 角色编辑表单

### Description

创建或编辑角色，填写名称、描述，勾选权限码。

### User Interaction Flow

1. 点击"创建角色"或编辑某角色 → 打开角色编辑表单
2. 填写角色名称（必填）和描述（选填）
3. 权限码按资源分组展示，每组可展开/折叠
4. 勾选权限码（至少选一个）
5. 点击保存 → 校验通过 → 保存成功，返回列表

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 角色名称 | string | 用户输入 | 2-50 字符 |
| 描述 | string | 用户输入 | 最多 200 字符 |
| 权限码列表 | string[] | 系统权限码定义 | 按 resource 分组展示 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Creating | 空表单，标题"创建角色" | 新建模式 |
| Editing | 预填已有数据，标题"编辑角色" | 编辑模式 |
| Saving | 保存按钮 loading 状态 | 提交中 |
| Error | 表单上方错误提示 | 保存失败 |

### Validation Rules

- 角色名称：必填，2-50 字符，不可与已有角色重名
- 描述：选填，最多 200 字符
- 权限勾选：至少选择一个权限码
- 预置角色：superadmin 不可编辑（无编辑入口），pm/member 可编辑权限勾选

---

## UI Function 3: 权限码浏览视图

### Description

展示系统所有可用权限码，按资源分组，只读展示。供管理员了解系统权限全貌。

### User Interaction Flow

1. 超级管理员在角色管理页面点击"查看权限列表"
2. 系统展示所有权限码，按资源分组
3. 每个权限码显示：权限码字符串、操作描述
4. 只读，不可编辑

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 资源名称 | string | 权限码分组键 | 如 team, main_item |
| 权限码 | string | 系统定义 | 如 team:create |
| 操作描述 | string | 系统定义 | 如"创建团队" |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Loading | 加载动画 | 首次加载 |
| Populated | 分组展示的权限码列表 | 加载完成 |

---

## UI Function 4: 邀请成员时的角色选择

### Description

在邀请用户加入团队的流程中，增加角色选择步骤。

### User Interaction Flow

1. PM 点击"邀请成员"
2. 搜索用户 → 选择用户
3. 从角色下拉列表中选择一个角色（默认选中 member）
4. 确认邀请

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 用户搜索 | string | 用户输入 | 按账号/用户名搜索 |
| 角色列表 | array | 系统角色列表 | 排除 superadmin |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Searching | 搜索结果下拉 | 输入搜索关键字 |
| No Results | "未找到匹配用户" | 搜索无结果 |
| Selected | 用户名 + 角色选择 | 选中用户 |

### Validation Rules

- 必须选择一个角色才能提交
- superadmin 角色不出现在可选列表中

---

## UI Function 5: 权限驱动的 UI 渲染

### Description

前端根据用户权限动态显示/隐藏按钮、菜单项和操作入口。

### User Interaction Flow

1. 用户登录成功 → 前端请求用户权限数据
2. 前端存储权限映射 { team_id: [permission_codes] }
3. 页面渲染时，每个受控 UI 元素检查当前上下文的权限码
4. 有权限则显示，无权限则隐藏（不是禁用/置灰）
5. 跨团队操作时使用对应团队的权限集

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| is_superadmin | boolean | `/api/me/permissions` | 全局权限标记 |
| team_permissions | map | `/api/me/permissions` | { team_id: [codes] } |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Permitted | 正常显示按钮/菜单 | 用户拥有所需权限码 |
| Not Permitted | 完全隐藏（不占空间） | 用户缺少所需权限码 |
| Superadmin | 显示所有 UI 元素 | is_superadmin = true |

### Validation Rules

- superadmin 用户显示所有 UI 元素，无需逐项检查
- 非团队上下文的操作（如创建团队）使用全局权限检查
- 权限数据缓存在前端，角色变更时需刷新

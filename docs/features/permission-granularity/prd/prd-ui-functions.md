---
feature: "permission-granularity"
---

# 细化 user / role 权限粒度 — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

- 团队管理页：添加成员对话框中的角色下拉选择器
- 用户管理页：页面入口权限控制
- 角色管理页：创建/编辑/删除按钮权限控制

---

## UI Function 1: 角色下拉选择器（团队管理页）

### Description

在"添加成员"对话框中，角色下拉选择器根据当前用户是否持有 `role:read` 权限决定是否加载角色列表。

### User Interaction Flow

1. 用户点击"添加成员"按钮，对话框打开
2. 系统检查当前用户是否持有 `role:read`
3. 若有：调用 `GET /admin/roles`，将角色列表填充到下拉选择器
4. 若无：下拉选择器显示为禁用状态，提示"无权限查看角色列表"

### Form Spec — "添加成员"对话框

| Field | Type | Required | Source | Notes |
|-------|------|----------|--------|-------|
| memberSearch | string | 否 | 用户输入 | 搜索框，按用户名或邮箱过滤成员候选列表 |
| memberId | string (bizKey) | 是 | 用户从搜索结果中选择 | 选中成员的唯一标识，提交时传给后端 |
| roleId | string (bizKey) | 是 | GET /admin/roles | 下拉选择器选中的角色 bizKey |
| roleName | string | — | GET /admin/roles | 下拉显示文本，不提交 |

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| roleId | string (bizKey) | GET /admin/roles | 角色唯一标识 |
| roleName | string | GET /admin/roles | 下拉显示文本 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| loading | 下拉框显示加载中 | 请求发出后 |
| populated | 显示角色列表 | 请求成功，用户有 role:read |
| disabled | 禁用，显示无权限提示 | 用户无 role:read |
| error | 显示加载失败，可重试 | 请求失败 |

### Validation Rules

- `memberId` 必填：未选择成员时，提交按钮禁用
- 成员搜索无结果：搜索框下方显示"未找到匹配成员"，`memberId` 保持未选中，提交按钮禁用
- `roleId` 必填：未选择角色时，提交按钮禁用，表单提示"请选择角色"
- 角色列表为空（`GET /admin/roles` 返回空数组）：下拉显示"暂无可用角色"，提交按钮禁用

---

## UI Function 2: 用户管理页入口权限控制

### Description

用户管理页的导航入口根据当前用户是否持有 `user:list` 权限决定是否显示。

### User Interaction Flow

1. 用户登录后，系统加载权限列表
2. 若用户持有 `user:list`：导航菜单中显示"用户管理"入口
3. 若用户不持有 `user:list`：导航菜单中不显示"用户管理"入口

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| permissions | string[] | GET /me/permissions | 当前用户权限码列表 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| visible | 显示"用户管理"菜单项 | 用户持有 user:list |
| hidden | 不显示菜单项 | 用户不持有 user:list |

### Validation Rules

- 权限列表加载期间（`GET /me/permissions` 未返回）：菜单项默认隐藏，不显示加载占位，避免权限加载完成后菜单项闪现
- 权限接口请求失败：菜单项保持隐藏，不因接口异常而意外暴露入口

---

## UI Function 3: 角色管理页操作按钮权限控制

### Description

角色管理页的创建、编辑、删除按钮分别根据 `role:create`、`role:update`、`role:delete` 权限独立控制显示/隐藏。

### User Interaction Flow

1. 用户进入角色管理页
2. 系统根据权限码列表独立控制每个操作按钮的可见性：
   - 持有 `role:create`：显示"创建角色"按钮
   - 持有 `role:update`：显示每行的"编辑"按钮
   - 持有 `role:delete`：显示每行的"删除"按钮
3. 用户点击操作按钮，系统调用对应接口

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| permissions | string[] | GET /me/permissions | 当前用户权限码列表 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| create-visible | 显示"创建角色"按钮 | 用户持有 role:create |
| create-hidden | 不显示"创建角色"按钮 | 用户不持有 role:create |
| edit-visible | 显示行内"编辑"按钮 | 用户持有 role:update |
| delete-visible | 显示行内"删除"按钮 | 用户持有 role:delete |

### Form Spec — 创建/编辑角色表单

| Field | Type | Required | Constraints | Notes |
|-------|------|----------|-------------|-------|
| roleName | string | 是 | 1–50 个字符，不可与现有角色名重复 | 创建和编辑均必填 |
| description | string | 否 | 最多 200 个字符 | 角色描述，可为空 |
| permissionCodes | string[] | 否 | 多选，选项来自 GET /admin/permissions | 空数组合法（无权限角色） |

权限码多选行为：
- 选项列表来自 `GET /admin/permissions`；若接口返回空，显示"暂无可用权限码"，多选框禁用
- 编辑时，已授予的权限码默认勾选
- 提交时传递完整的选中列表（全量覆盖，非增量）

### Validation Rules

- 预置角色（superadmin/pm/member）的"删除"按钮始终禁用，不受 `role:delete` 权限影响
- 有成员绑定的角色，"删除"按钮禁用，悬停显示 Tooltip："该角色下有成员，无法删除"
- `roleName` 为空时，提交按钮禁用
- `roleName` 超过 50 字符时，输入框下方显示"角色名称不能超过 50 个字符"
- `roleName` 与现有角色名重复时，提交后后端返回 409，表单显示"角色名称已存在"
- `description` 超过 200 字符时，输入框下方显示"角色描述不能超过 200 个字符"，提交按钮禁用

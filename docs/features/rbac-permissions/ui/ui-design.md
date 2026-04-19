---
created: 2026-04-19
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: RBAC 权限体系

## Design System

- **Framework**: React 18 + TypeScript
- **Component Library**: shadcn/ui 风格组件（Radix UI 基础 + Tailwind CSS）
- **已有 UI 原语**: `Table`, `Dialog`, `Input`, `Select`, `Button`, `Badge`, `Pagination`, `PaginationPageSize`, `Breadcrumb`, `Tooltip`, `Tabs`
- **已有共享组件**: `ConfirmDialog`, `UserAvatar`, `StatusBadge`
- **状态管理**: zustand (auth store, team store)
- **数据获取**: @tanstack/react-query (useQuery, useMutation)
- **样式**: Tailwind CSS v4，主题变量定义在 `src/index.css`

### 新增 UI 原语

| 组件 | 说明 |
|------|------|
| `CheckboxGroup` | 分组复选框，用于权限码勾选 |
| `CollapsibleSection` | 可折叠区块，用于权限码分组展示 |

---

## Component 1: 角色列表页 (RoleManagementPage)

> 对应 UI Function 1，路由 `/roles`

### Layout Structure

```
AppLayout
└── main content area
    ├── Page Header
    │   ├── Breadcrumb: "首页 > 角色管理"
    │   ├── Title: "角色管理"
    │   └── Button: "创建角色" (right-aligned)
    ├── Filter Bar
    │   ├── Input (搜索角色名称)
    │   └── Select (预置筛选: 全部/预置/自定义)
    ├── Table (角色列表)
    │   ├── Header: 角色名称 | 描述 | 权限数量 | 使用人数 | 类型 | 创建时间 | 操作
    │   └── Row: {name} | {description} | {permCount} | {userCount} | Badge(preset/custom) | {createdAt} | EditBtn DeleteBtn
    └── Pagination
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 表格行显示骨架屏（3-5 行） | 筛选栏不可操作 |
| Populated | 完整数据表格 | 支持排序、翻页、筛选 |
| Empty | 居中提示"暂无自定义角色" + 创建按钮 | 仅在筛选无结果时显示 |
| Error | 表格上方红色错误提示条 | 提供"重试"按钮 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击"创建角色" | 打开角色编辑 Dialog（创建模式） | Dialog 弹出 |
| 点击角色名称/编辑按钮 | 打开角色编辑 Dialog（编辑模式） | Dialog 弹出，预填数据 |
| 点击"删除"按钮 | 打开 ConfirmDialog | 显示角色名称和使用人数，有用户时提示无法删除 |
| 输入搜索关键字 | 防抖 300ms 后筛选列表 | 表格实时更新 |
| 切换预置筛选 | 立即筛选列表 | 表格更新 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 角色名称列 | `role.name` | GET /api/roles |
| 描述列 | `role.description` | GET /api/roles |
| 权限数量列 | `role.permissionCount` | GET /api/roles (后端关联计数) |
| 使用人数列 | `role.memberCount` | GET /api/roles (后端关联计数) |
| 类型 Badge | `role.isPreset` | true → "预置" Badge variant=secondary; false → "自定义" |
| 创建时间列 | `role.createdAt` | 格式化 YYYY/MM/DD |
| 删除按钮 | disabled when `role.isPreset \|\| role.memberCount > 0` | Tooltip 提示原因 |

---

## Component 2: 角色编辑表单 (RoleEditDialog)

> 对应 UI Function 2，Dialog 组件

### Layout Structure

```
Dialog (size="lg")
├── DialogHeader
│   └── DialogTitle: "创建角色" | "编辑角色: {name}"
├── DialogBody
│   ├── Form Field: 角色名称
│   │   ├── label + 必填标记
│   │   └── Input
│   ├── Form Field: 描述
│   │   ├── label
│   │   └── textarea (native)
│   ├── Form Field: 权限配置
│   │   ├── label + "至少选择 1 个权限"
│   │   └── Permission Checkbox Groups
│   │       ├── CollapsibleSection "团队管理"
│   │       │   └── Checkbox: team:create, team:read, ...
│   │       ├── CollapsibleSection "主事项"
│   │       │   └── Checkbox: main_item:create, ...
│   │       └── ... (按资源分组)
│   └── Error Message (条件渲染)
└── DialogFooter
    ├── Button (variant="secondary"): 取消
    └── Button (loading when saving): 保存
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Creating | 空表单，标题"创建角色"，所有权限未勾选 | 用户从零填写 |
| Editing (预置角色) | 名称/描述只读，权限可勾选 | 标题"编辑角色: {name}"，名称字段 disabled |
| Editing (自定义角色) | 全部可编辑，预填已有数据 | 标题"编辑角色: {name}" |
| Saving | 保存按钮显示 loading spinner | 不可重复点击 |
| Error | DialogBody 顶部显示红色错误文本 | 保留表单数据不丢失 |
| Name Conflict | 角色名称输入框下方错误提示"角色名称已存在" | 不关闭 Dialog |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 输入角色名称 | 实时校验长度（2-50 字符） | 不符合时显示校验提示 |
| 勾选/取消权限 | 更新权限列表 | 分组标题显示已选数量 "团队管理 (3/5)" |
| 点击折叠标题 | 展开/折叠该分组 | 箭头旋转动画 |
| 点击"保存" | 校验 → 提交 API → 成功关闭 Dialog | 成功后 invalidate roles query 并 toast 提示 |
| 点击"取消" | 关闭 Dialog | 不保存，不提示（表单无确认丢失逻辑） |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 角色名称 Input | `form.name: string` | useState |
| 描述 textarea | `form.description: string` | useState |
| 权限复选框组 | `form.permissionCodes: string[]` | useState |
| 创建模式 | POST /api/roles `{ name, description, permissionCodes }` | useMutation |
| 编辑模式 | PUT /api/roles/:id `{ name, description, permissionCodes }` | useMutation |

### 权限分组定义

| 分组名 | 权限码 |
|--------|--------|
| 团队管理 | team:create, team:read, team:update, team:delete, team:invite, team:remove, team:transfer |
| 主事项 | main_item:create, main_item:read, main_item:update, main_item:archive |
| 子事项 | sub_item:create, sub_item:read, sub_item:update, sub_item:assign, sub_item:change_status |
| 进度管理 | progress:create, progress:read, progress:update |
| 事项池 | item_pool:submit, item_pool:review |
| 视图 | view:weekly, view:gantt, view:table |
| 周报 | report:export |
| 用户管理 | user:read, user:update, user:manage_role |

---

## Component 3: 权限码浏览视图 (PermissionBrowseDialog)

> 对应 UI Function 3，Dialog 组件

### Layout Structure

```
Dialog (size="md")
├── DialogHeader
│   └── DialogTitle: "系统权限列表"
├── DialogBody
│   └── Permission Groups (只读)
│       ├── CollapsibleSection "团队管理" (默认展开)
│       │   └── Table (2 列)
│       │       ├── Header: 权限码 | 说明
│       │       └── Row: team:create | 创建团队
│       ├── CollapsibleSection "主事项"
│       │   └── ...
│       └── ... (8 个分组)
└── DialogFooter
    └── Button (variant="secondary"): 关闭
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 分组区域骨架屏 | — |
| Populated | 完整权限列表，按分组折叠展示 | 默认全部展开 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击分组标题 | 展开/折叠该分组 | 箭头旋转 |
| 点击"关闭" | 关闭 Dialog | — |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 分组标题 | 资源名称（如"团队管理"） | 前端硬编码（权限码由代码定义） |
| 权限码列 | 权限码字符串 | 前端硬编码 |
| 说明列 | 操作描述 | 前端硬编码 |

> 权限码列表是系统代码定义的常量，不需要从 API 获取。前端维护一份权限码元数据即可。

---

## Component 4: 邀请成员角色选择

> 对应 UI Function 4，嵌入现有邀请成员 Dialog

### Layout Structure

在现有 TeamDetailPage 的邀请 Dialog 中增加角色选择字段：

```
Dialog (inviteMember)
├── DialogHeader
│   └── DialogTitle: "邀请成员"
├── DialogBody
│   ├── Form Field: 搜索用户
│   │   └── Input (搜索框)
│   ├── 搜索结果列表（条件渲染）
│   │   └── 搜索结果项 (用户名 + 选择按钮)
│   ├── Form Field: 角色 ← 新增
│   │   └── Select (角色下拉)
│   │       ├── SelectTrigger
│   │       └── SelectContent
│   │           └── SelectItem: {role.name} (排除 superadmin)
│   └── Error Message
└── DialogFooter
    ├── Button (variant="secondary"): 取消
    └── Button: 确认邀请
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Searching | Input 下方显示搜索结果下拉列表 | 防抖 300ms 搜索 |
| No Results | "未找到匹配用户" 提示 | Input 下方 |
| Selected | 用户名显示 + 角色下拉可用 | 角色默认选中第一个非 superadmin 角色 |
| Saving | 确认按钮 loading | 不可重复点击 |
| Error | 错误提示 | 保留表单数据 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 输入搜索关键字 | 防抖 300ms 搜索用户 | 下拉列表更新 |
| 选择用户 | 显示用户名，激活角色选择 | 角色下拉默认选中 member |
| 选择角色 | 更新选中角色 | — |
| 点击"确认邀请" | 校验 → 提交 API → 成功关闭 Dialog | 成功后 invalidate members query 并 toast |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 搜索结果 | `searchResults: User[]` | GET /api/users/search?q={keyword} |
| 选中用户 | `selectedUser: User` | useState |
| 角色下拉 | `roles` (排除 superadmin) | GET /api/roles |
| 选中角色 | `selectedRoleId: number` | useState, 默认 member 角色ID |
| 提交 | POST /api/teams/:id/members `{ userId, roleId }` | useMutation |

---

## Component 5: 权限驱动的 UI 渲染

> 对应 UI Function 5，全局渲染逻辑

### Layout Structure

不新增页面组件，而是通过权限 Hook 和 Wrapper 影响现有 UI。

### 权限数据结构

```typescript
interface PermissionData {
  isSuperadmin: boolean
  teamPermissions: Record<number, string[]>  // teamId → permissionCodes
}
```

### 新增 Hook: usePermission

```typescript
function usePermission(code: string): boolean
```

逻辑：
1. 从 auth store 读取 `permissionData`
2. 若 `isSuperadmin === true`，返回 true
3. 若当前在团队上下文（team store 中有 currentTeamId），检查 `teamPermissions[currentTeamId]` 是否包含 code
4. 若不在团队上下文（如创建团队按钮），检查用户所有团队的权限合集是否包含 code

### 新增组件: PermissionGuard

```tsx
<PermissionGuard code="team:invite">
  <Button>邀请成员</Button>
</PermissionGuard>
```

- 有权限：渲染子组件
- 无权限：不渲染（不占空间）

### 受控 UI 元素改动

| 现有位置 | UI 元素 | 权限码 | 改动方式 |
|---------|---------|--------|---------|
| Sidebar.tsx | "用户管理"菜单项 | `user:read` | PermissionGuard 包裹 |
| Sidebar.tsx | "甘特图"菜单项 | `view:gantt` | PermissionGuard 包裹 |
| TeamManagementPage.tsx | "创建团队"按钮 | `team:create` | PermissionGuard 包裹（非团队上下文检查） |
| TeamDetailPage.tsx | "邀请成员"按钮 | `team:invite` | PermissionGuard 包裹 |
| TeamDetailPage.tsx | "移除成员"按钮 | `team:remove` | PermissionGuard 包裹 |
| TeamDetailPage.tsx | "转让 PM"按钮 | `team:transfer` | PermissionGuard 包裹 |
| TeamDetailPage.tsx | "编辑团队"按钮 | `team:update` | PermissionGuard 包裹 |
| TeamDetailPage.tsx | "解散团队"按钮 | `team:delete` | PermissionGuard 包裹 |
| ItemViewPage.tsx | "创建主事项"按钮 | `main_item:create` | PermissionGuard 包裹 |
| MainItemDetailPage.tsx | "编辑"按钮 | `main_item:update` | PermissionGuard 包裹 |
| MainItemDetailPage.tsx | "归档"按钮 | `main_item:archive` | PermissionGuard 包裹 |
| SubItemDetailPage.tsx | "分配负责人"按钮 | `sub_item:assign` | PermissionGuard 包裹 |
| ItemPoolPage.tsx | "审核"按钮 | `item_pool:review` | PermissionGuard 包裹 |
| 进度详情 | "修正进度"按钮 | `progress:update` | PermissionGuard 包裹 |
| ReportPage.tsx | "导出周报"按钮 | `report:export` | PermissionGuard 包裹 |

### 权限获取与缓存

| 时机 | 行为 |
|------|------|
| 登录成功后 | 调用 `GET /api/me/permissions`，存入 auth store |
| 路由切换时 | 检查缓存时间，超过 5 分钟则重新请求 |
| 团队切换时 | 立即刷新权限数据（不同团队权限可能不同） |
| 角色变更后（被变更用户） | 5 分钟轮询间隔检测到权限变化，自动刷新 |

---

## Component 6: 变更成员角色

> 对应 UI Function 6，嵌入现有 TeamDetailPage

### Layout Structure

在现有团队成员列表的每行中修改角色显示：

```
TableRow (团队成员)
├── TableCell: UserAvatar + 用户名
├── TableCell: 角色 ← 改造
│   ├── 文本: {roleName}
│   └── Button (variant="ghost", size="sm"): "变更" ← 仅 PM/超管可见
│       └── 点击 → 内联角色 Select
└── TableCell: 操作 (移除按钮)
```

内联变更交互：

```
点击"变更"
  → 角色文本替换为 Select (排除 superadmin)
  → 选择新角色
  → 自动提交（或显示确认按钮）
  → 成功：恢复文本显示，toast 提示
  → 失败：错误提示，恢复原角色文本
```

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Viewing | 角色名称文本 + "变更"按钮 | 默认状态 |
| Selecting | 角色文本替换为 Select 下拉 | 聚焦到 Select |
| Saving | Select 置灰 + loading spinner | 等待 API 响应 |
| Error | 原角色文本恢复 + toast 错误提示 | 自动退出 Selecting 状态 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击"变更" | 行内显示角色 Select | Select 自动展开 |
| 选择新角色 | 立即提交变更 API | 行内 loading |
| 变更成功 | 恢复文本显示新角色名 | toast "角色已更新" |
| 变更失败 | 恢复原角色名 | toast 错误信息 |
| 点击 Select 外部 | 取消变更，恢复原文本 | 无 API 调用 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 当前角色 | `member.roleName` | teamMembers query |
| 变更按钮 | visible when `hasPermission('team:invite')` | usePermission |
| 角色 Select | `roles` (排除 superadmin) | GET /api/roles |
| 提交变更 | PUT /api/teams/:id/members/:userId/role `{ roleId }` | useMutation |

### Validation Rules

- PM 不能变更自己的角色（"变更"按钮不显示在 PM 自己的行上）
- superadmin 角色不出现在下拉列表中
- 必须选择与当前角色不同的新角色才触发提交

---

## 路由与导航变更

### 新增路由

| 路由 | 组件 | 权限 |
|------|------|------|
| `/roles` | RoleManagementPage | `user:manage_role` |

### Sidebar 变更

```
现有结构:
  标准导航项 (事项清单, 每周进展, ...)
  --- 分隔线 ---
  adminItems (用户管理) ← 仅 isSuperAdmin

变更后:
  标准导航项 (事项清单, 每周进展, ...)
  --- 分隔线 ---
  管理区:
    角色管理 ← PermissionGuard(user:manage_role)
    用户管理 ← PermissionGuard(user:read)
```

"角色管理"菜单项使用 `Shield` 图标 (lucide-react)。

### AdminRoute 替换

现有 `AdminRoute` 组件基于 `isSuperAdmin` 布尔值。RBAC 上线后：
- `/users` 路由的权限检查改为 `RequirePermission("user:read")`
- `/roles` 路由的权限检查改为 `RequirePermission("user:manage_role")`
- `AdminRoute` 组件替换为 `PermissionRoute` 组件，接收 `code` prop

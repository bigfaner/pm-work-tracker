---
created: 2026-04-18
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: RBAC 权限体系

## Design System

Tailwind UI — Indigo 主色，Inter 字体，专业温暖 SaaS 风格。沿用现有 `styles.css` 中的 CSS 变量和组件样式。

核心 token：
- Primary: `#4f46e5` (primary-600)，按钮/链接
- 背景: `#ffffff` / `#f8fafc` 交替
- 文字: `#0f172a` (primary) / `#475569` (secondary) / `#94a3b8` (tertiary)
- 圆角: 6px (input) / 8px (button) / 12px (card)
- 阴影: `0 1px 3px rgba(0,0,0,0.1)` (card)

## Component: 角色列表页

> UI Function 1 — 超级管理员在管理后台查看和管理所有角色。

### Layout Structure

在现有 `admin.html` 的 tabs 中新增"角色管理"tab，与"用户管理""团队列表"并列。

```
┌──────────────────────────────────────────────┐
│ [用户管理]  [团队列表]  [角色管理]            │  ← tabs
├──────────────────────────────────────────────┤
│ 筛选栏: [搜索角色名]  [预置/自定义 ▾]        │
│                            [+ 创建角色]      │
├──────────────────────────────────────────────┤
│ 表格:                                        │
│ 角色名称 | 描述 | 权限数 | 使用人数 | 操作   │
│ ─────────────────────────────────────────    │
│ superadmin | 系统最高权限 | — | 1 | —       │  ← 预置，无操作
│ pm         | 团队管理权限 | 8  | 5 | 编辑   │  ← 预置，仅可编辑
│ member     | 基础成员权限 | 4  | 12| 编辑   │
│ viewer     | 只读查看者   | 3  | 0 | 编辑 删│  ← 自定义
├──────────────────────────────────────────────┤
│ 分页: < 1 2 3 >                              │
└──────────────────────────────────────────────┘
```

角色名称列：预置角色后附加 `badge-slate` 标记"预置"，自定义角色无标记。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 表格行显示骨架屏（3-4 行灰色条纹，pulse 动画） | 首次进入 tab 或刷新 |
| Populated | 完整数据表格 | 数据加载完成 |
| Empty | 居中插图 + "暂无自定义角色" + "创建角色"按钮 | 仅有预置角色时 |
| Error | 红色 alert 横幅 "加载失败，请重试" + 重试按钮 | API 返回错误 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击"创建角色" | 打开角色编辑表单（模态框） | 模态框从右侧滑入或 fade-in |
| 点击"编辑" | 打开预填的角色编辑表单（模态框） | 同上 |
| 点击"删除" | 显示确认弹窗 | 弹窗标题"删除角色"，内容"确定删除角色「{name}」？此操作不可撤销"，红色"删除"+ 灰色"取消" |
| 删除成功 | 关闭弹窗，刷新列表 | 右上角绿色 toast "角色已删除" |
| 搜索输入 | 防抖 300ms 后筛选列表 | 列表实时更新 |
| 预置/自定义筛选 | 切换下拉，刷新列表 | 立即生效 |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 角色名称 | `name` | `GET /api/v1/admin/roles` → `items[].name` |
| "预置"标记 | `is_preset` | `items[].is_preset` |
| 描述 | `description` | `items[].description` |
| 权限数 | `permission_count` | `items[].permission_count` |
| 使用人数 | `member_count` | `items[].member_count` |
| 编辑按钮 | `is_preset` + 排除 superadmin | `is_preset && name !== "superadmin"` → 可编辑 |
| 删除按钮 | `is_preset === false && member_count === 0` | 不可用时隐藏（非禁用） |

---

## Component: 角色编辑表单

> UI Function 2 — 创建或编辑角色的模态框表单。

### Layout Structure

模态框（overlay + 居中卡片），最大宽度 640px，最大高度 80vh，内容可滚动。

```
┌─────────────────────────────────────┐
│ 创建角色 / 编辑角色            [✕]  │
├─────────────────────────────────────┤
│ 角色名称 *                          │
│ [________________________]          │
│                                     │
│ 描述                                │
│ [________________________]          │
│ [________________________]          │
│                                     │
│ 权限配置 *                          │
│ ┌─ team ──────────────────────┐     │
│ │ ☑ team:create  创建团队     │     │
│ │ ☑ team:read    查看团队信息 │     │
│ │ ☐ team:update  编辑团队信息 │     │
│ │ ☐ team:invite  邀请成员     │     │
│ └─────────────────────────────┘     │
│ ┌─ main_item ─────────────────┐     │
│ │ ☐ main_item:create 创建事项 │     │
│ │ ...                         │     │
│ └─────────────────────────────┘     │
│ ...更多资源组（折叠/展开）          │
├─────────────────────────────────────┤
│                     [取消] [保存]   │
└─────────────────────────────────────┘
```

权限分组区域：每个资源为一个 collapsible section，标题显示资源名和已选数量（如 "team (2/4)"），默认全部展开。编辑预置角色时，名称字段 readonly。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Creating | 空表单，标题"创建角色"，名称和描述为空 | 从"创建角色"按钮进入 |
| Editing | 预填已有数据，标题"编辑角色" | 从角色列表"编辑"进入 |
| Saving | "保存"按钮显示 spinner + "保存中..."，表单不可编辑 | 提交请求进行中 |
| Error | 表单顶部红色 alert 横幅 | 保存失败（409 重名等） |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 名称输入 | 实时校验长度 | 超出范围时输入框变红 |
| 权限勾选 | 更新资源组的已选计数 | 资源标题 "(2/4)" 实时更新 |
| 点击资源组标题 | 折叠/展开该组 | 180° 箭头旋转动画 |
| 点击"保存" | 校验 → 提交 | 校验失败高亮错误字段 |
| 保存成功 | 关闭模态框，刷新角色列表 | 绿色 toast "角色已创建"/"角色已更新" |
| 点击"取消"或"✕" | 关闭模态框 | 表单有修改时弹出确认"放弃编辑？" |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 名称输入 | `name` | 创建时空；编辑时 `GET /api/v1/admin/roles/:id` → `name` |
| 描述输入 | `description` | 同上 |
| 权限勾选 | `permission_codes` | 编辑时预填 `permissions[].code` |
| 资源分组 | `resource` | `GET /api/v1/admin/permissions` → 按 `resource` 分组 |
| 保存提交 | POST/PUT body | `{ name, description, permission_codes }` |

---

## Component: 权限码浏览视图

> UI Function 3 — 只读展示所有系统权限码，按资源分组。

### Layout Structure

在角色管理 tab 内，"创建角色"按钮旁放"查看权限列表"按钮。点击后切换到权限浏览视图（同 tab 内切换，非新页面）。

```
┌──────────────────────────────────────────────┐
│ [← 返回角色列表]              查看权限列表    │
├──────────────────────────────────────────────┤
│ ┌─ team (4 项) ─────────────────────────┐    │
│ │ 权限码              │ 操作描述         │    │
│ │ team:create         │ 创建团队         │    │
│ │ team:read           │ 查看团队信息     │    │
│ │ team:update         │ 编辑团队信息     │    │
│ │ team:invite         │ 邀请成员         │    │
│ └───────────────────────────────────────┘    │
│ ┌─ main_item (4 项) ────────────────────┐    │
│ │ ...                                   │    │
│ └───────────────────────────────────────┘    │
└──────────────────────────────────────────────┘
```

每个资源组为一个 card，权限码用 `font-mono` 样式显示。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 卡片骨架屏 | 首次加载 |
| Populated | 分组卡片列表 | 加载完成 |

无需 empty 和 error 状态（权限码由系统定义，始终存在）。

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 点击"← 返回角色列表" | 切换回角色列表视图 | 无动画，直接切换 |
| 点击"查看权限列表" | 切换到权限浏览视图 | 加载 `GET /api/v1/admin/permissions` |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 资源组标题 | `resource` | `GET /api/v1/admin/permissions` → `resource` |
| 权限码 | `actions[].code` | `font-mono` 显示 |
| 操作描述 | `actions[].description` | 普通文本 |

---

## Component: 邀请成员时的角色选择

> UI Function 4 — 在邀请流程中增加角色选择下拉。

### Layout Structure

在现有团队管理页面的邀请流程中，"邀请成员"按钮打开的模态框/面板内，新增角色选择下拉框。

```
┌─────────────────────────────────┐
│ 邀请成员到「产品研发团队」  [✕] │
├─────────────────────────────────┤
│ 用户账号                        │
│ [搜索用户名/账号________]       │
│ ┌ 搜索结果 ──────────────┐     │
│ │ 张三 (zhangsan)         │     │
│ │ 李四 (lisi)             │     │
│ └─────────────────────────┘     │
│                                 │
│ 角色 *                          │
│ [▼ PM - 团队管理权限_____]      │  ← 新增
│                                 │
│              [取消] [发送邀请]   │
└─────────────────────────────────┘
```

角色下拉显示 `name — description` 格式，默认选中"member"角色。superadmin 角色不出现在列表中。

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Searching | 下拉列表显示 spinner | 输入搜索关键字时 |
| No Results | "未找到匹配用户" | 搜索无结果 |
| Selected | 用户名 + 角色选择 | 选中用户后 |
| Inviting | "发送邀请"按钮 loading | 提交中 |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| 搜索用户 | 防抖搜索，展示下拉结果 | 结果列表实时更新 |
| 选择用户 | 填入用户名，启用角色下拉 | 角色下拉变为可操作 |
| 切换角色 | 更新选中角色 | 无特殊反馈 |
| 点击"发送邀请" | 提交 `{ username, role_id }` | 成功后关闭模态框 + toast |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| 角色下拉列表 | `id`, `name`, `description` | `GET /api/v1/admin/roles`（排除 superadmin） |
| 提交数据 | `username`, `role_id` | `POST /api/v1/teams/:teamId/members` |

---

## Component: 权限驱动的 UI 渲染

> UI Function 5 — 前端根据用户权限动态显示/隐藏 UI 元素。

### Layout Structure

非可见组件，属于前端渲染逻辑层。

### 受控 UI 元素映射

| UI 元素 | 所需权限 | 所在页面 |
|---------|----------|----------|
| "创建团队"按钮 | `team:create` | 主页 / 导航 |
| "邀请成员"按钮 | `team:invite` | 团队管理 |
| "编辑团队信息"按钮 | `team:update` | 团队管理 |
| "创建事项"按钮 | `main_item:create` | 事项列表 |
| "编辑事项"按钮 | `main_item:update` | 事项详情 |
| "删除事项"按钮 | `main_item:delete` | 事项详情 |
| "创建子项"按钮 | `sub_item:create` | 事项详情 |
| "编辑子项"按钮 | `sub_item:update` | 子项详情 |
| "删除子项"按钮 | `sub_item:delete` | 子项详情 |
| "管理后台"导航入口 | `is_superadmin` | 侧边栏 |
| "导出周报"按钮 | `report:export` | 周视图 |
| "甘特图"导航入口 | `view:gantt` | 侧边栏 |
| "变更成员角色"按钮 | `team:invite` | 团队管理 |

### Implementation Rules

1. **完全隐藏**（`display: none`），不是禁用或置灰。用户看不到无权限的 UI 元素
2. **superadmin 绕过**：`is_superadmin === true` 时显示所有 UI 元素，无需逐项检查权限码
3. **团队上下文**：使用当前选中团队的 `team_permissions[teamId]` 权限集
4. **全局操作**：创建团队等非团队上下文的操作，检查所有团队权限的并集或 `is_superadmin`
5. **缓存策略**：登录后请求 `GET /api/v1/me/permissions` 并缓存至内存；角色变更时重新请求刷新

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Permitted | 正常显示 | 用户拥有所需权限码 |
| Not Permitted | `display: none`（不占空间） | 用户缺少权限码 |
| Superadmin | 显示所有 | `is_superadmin = true` |

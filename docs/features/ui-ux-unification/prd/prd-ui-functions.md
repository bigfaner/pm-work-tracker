---
feature: "ui-ux-unification"
---

# UI/UX 统一与优化 — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

- Sidebar（左侧导航菜单）
- WeekPicker 共享组件（新建）
- WeeklyViewPage（每周进展）
- ReportPage（周报导出）
- MainItemDetailPage（主事项详情）
- SubItemDetailPage（子事项详情）
- ItemViewPage（事项清单）
- ItemPoolPage（待办事项）
- UserManagementPage（用户管理）
- RoleManagementPage（角色管理）
- TeamDetailPage（团队详情）
- 全站可跳转链接样式

---

## UI Function 1: 左侧导航菜单重排

### Description
调整 Sidebar 菜单项顺序，增加业务组与管理组的视觉分隔线。

### User Interaction Flow
1. 用户登录后，左侧菜单按新顺序渲染
2. 业务组：事项清单 → 待办事项 → 每周进展 → 整体进度 → 周报导出
3. 分隔线
4. 管理组：团队管理 → 用户管理（需权限）→ 角色管理（需权限）
5. 无权限的菜单项由 PermissionGuard 控制隐藏，不影响分组结构

### States

| State | Display | Trigger |
|-------|---------|---------|
| 有权限 | 菜单项正常显示 | 用户拥有对应权限 |
| 无权限 | 菜单项隐藏 | PermissionGuard 判断无权限 |
| 当前页 | 菜单项高亮（`bg-primary-50 text-primary-700`） | 路由匹配 |

---

## UI Function 2: WeekPicker 组件

### Description
跨浏览器兼容的周次选择器，单行紧凑布局，支持上一周/下一周快速切换。替换 WeeklyViewPage 和 ReportPage 中的原生 `<input type="week">`。

### User Interaction Flow
1. 组件初始化，显示当前周：`‹   2026年第16周  04/13 ~ 04/19   ›`
2. 用户点击「‹」→ 切换到上一周，更新显示文字，触发 `onChange` 回调
3. 用户点击「›」→ 若非最大周，切换到下一周；若已是最大周，按钮禁用不响应
4. 父组件接收 `onChange(newWeekStart)` 后刷新数据

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| weekStart | string (YYYY-MM-DD) | 父组件传入 | 当前周的周一日期 |
| weekEnd | string (YYYY-MM-DD) | 组件内计算 | weekStart + 6天 |
| weekNumber | number | 组件内计算 | ISO 周次 |
| year | number | 组件内计算 | 周次所属年份 |
| maxWeek | string (YYYY-MM-DD) | 父组件传入（可选） | 默认当前周周一 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 正常 | `‹ 2026年第16周 04/13~04/19 ›`，两侧箭头均可点击 | weekStart < maxWeek |
| 已到最大周 | `›` 按钮 `opacity-40 cursor-not-allowed` | weekStart >= maxWeek |

### Validation Rules
- `weekStart` 必须是周一（ISO 周起始日）
- 不允许选择超过 `maxWeek` 的周次

---

## UI Function 3: 表格行内操作按钮图标

### Description
为用户管理、角色管理、团队详情、待办事项页面的表格行内操作按钮添加图标，统一为「图标 + 文字」风格。

### User Interaction Flow
1. 用户查看表格，操作列显示带图标的按钮
2. 图标在文字左侧，间距 `gap-1.5`，图标尺寸 `w-3.5 h-3.5`
3. 点击行为与现有逻辑完全一致，仅视觉变化

### Data Requirements
无新增数据需求，仅视觉改动。

### States

| State | Display | Trigger |
|-------|---------|---------|
| 正常 | 图标 + 文字，`variant="ghost" size="sm"` | 默认 |
| 禁用 | 图标 + 文字，`disabled` 样式 | 如角色管理中预置角色的删除按钮 |

---

## UI Function 4: 子事项详情页「追加进度」按钮位置

### Description
将「追加进度」按钮从页面标题栏右侧移至「进度记录」卡片的 CardHeader 右侧。

### User Interaction Flow
1. 用户进入子事项详情页
2. 页面标题栏仅显示子事项标题，无操作按钮
3. 滚动到「进度记录」卡片，CardHeader 右侧显示「追加进度」按钮（需 `progress:update` 权限）
4. 点击按钮，弹出追加进度对话框（行为不变）

### States

| State | Display | Trigger |
|-------|---------|---------|
| 有权限 | 按钮显示在进度记录卡片右上角 | 用户拥有 `progress:update` 权限 |
| 无权限 | 按钮隐藏 | PermissionGuard 判断 |

---

## UI Function 5: 每周进展子事项完成度显示

### Description
在 WeeklyViewPage 的 SubItemRow 组件中，子事项标题旁展示当前完成度百分比数字。

### User Interaction Flow
1. 用户查看每周进展页的主事项卡片
2. 每个子事项行在标题文字之后、负责人之前显示完成度数字
3. 数字格式：`65%`，完成时显示绿色

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| completion | number (0-100) | SubItemSnapshot.completion | 已有字段 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 进行中 | `65%`，`text-[11px] font-semibold text-secondary` | completion < 100 |
| 已完成 | `100%`，`text-[11px] font-semibold text-success-text` | completion === 100 |

---

## UI Function 6: 可跳转文字链接高亮

### Description
全站所有 `<Link>` 跳转文字统一使用主色高亮样式，与普通文字明显区分。

### User Interaction Flow
1. 用户浏览任意页面
2. 可跳转的文字（事项标题、团队名称、面包屑等）显示为主色高亮
3. 鼠标悬停时显示下划线，确认可点击
4. 点击后正常跳转

### States

| State | Display | Trigger |
|-------|---------|---------|
| 默认 | `text-primary-600`，无下划线 | 正常渲染 |
| Hover | `text-primary-700 underline` | 鼠标悬停 |

---

## UI Function 7: 详情页信息布局统一

### Description
子事项详情页基本信息卡片与主事项详情页采用相同的 label+value 展示规范，统一标签字体、值字体和间距。

### User Interaction Flow
1. 用户进入子事项详情页
2. 基本信息卡片使用与主事项详情页一致的 label（`text-xs text-tertiary`）+ value（`text-[13px]`）网格布局
3. 视觉风格与主事项详情页保持一致

### States

| State | Display | Trigger |
|-------|---------|---------|
| 加载中 | 卡片骨架或空白 | 数据请求中 |
| 已加载 | 完整信息网格 | 数据返回 |

---

## UI Function 8: 事项清单按钮文案

### Description
将事项清单页（ItemViewPage）的「创建主事项」按钮文案更改为「新增主事项」。

### User Interaction Flow
1. 用户进入事项清单页
2. 页面右上角按钮显示「新增主事项」
3. 点击行为与原来完全一致

### States

| State | Display | Trigger |
|-------|---------|---------|
| 有权限 | 显示「新增主事项」按钮 | 用户拥有创建权限 |
| 无权限 | 按钮隐藏 | PermissionGuard 判断 |

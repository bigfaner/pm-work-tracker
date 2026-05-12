---
feature: "里程碑图"
---

# 里程碑图 — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

新增 1 个独立页面（/milestones）+ 修改 3 个现有页面（事项清单、主事项编辑、表格视图），提供里程碑的创建、管理、可视化展示和与现有事项的集成。

## Navigation Architecture

- **Platform**: web

### Primary Navigation (shared across pages)

| # | Label | Target Page | Icon Keyword |
|---|-------|-------------|-------------|
| 1 | 里程碑图 | /milestones | milestone/timeline |

### Secondary Pages (navigated from a parent page)

| Page | Entry Point (UF# or action) | Return Target |
|------|-----------------------------|---------------|
| 里程碑详情面板 | UF-1 点击里程碑节点 | /milestones |
| 主事项详情 | UF-1 点击关联的 MI 条目 | /items/:mainItemId |

### Navigation Rules

- 里程碑图页面从主导航进入，与事项清单/甘特图/周报/表格平级
- 里程碑详情面板为 overlay（不离开当前页面）
- 从详情面板点击 MI 条目跳转到主事项详情页

---

## UI Function 1: 里程碑时间线页面

### Placement

- **Mode**: new-page
- **Target Page**: /milestones
- **Position**: 独立页面，主导航"事项清单"和"甘特图"之间

### Description

展示团队所有里程碑及其关联 MainItem 的横向时间线视图。里程碑作为时间轴上的节点，关联的 MI 按时间排列在下方并连线到对应里程碑。支持缩放、拖拽归属变更、点击交互。

### User Interaction Flow

1. 用户从主导航进入 /milestones 页面
2. 系统渲染时间线：横向时间轴，里程碑节点按计划日期排列，关联 MI 展示在对应里程碑下方
3. 用户可缩放时间轴（周/月/季切换控件）
4. 用户点击里程碑节点 → 弹出详情面板（名称、日期、状态、完成度、关联 MI 列表）
5. 用户点击空白处或"创建里程碑"按钮 → 弹出创建表单
6. 用户拖拽 MI 条目到不同里程碑 → 改变 MI 归属
7. 用户拖拽 MI 到空白区域 → 解绑 MI

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 里程碑名称 | string | milestone.name | 节点上显示 |
| 计划日期 | date | milestone.planned_date | 节点上显示，决定时间轴位置 |
| 状态 | enum | milestone.status | not_started/in_progress/completed/cancelled |
| 完成度 | decimal | 计算值 | 关联 MI completion 的平均值，空里程碑为 0 |
| 关联 MI 数量 | int | 计算值 | 悬停时显示 |
| MI 标题 | string | main_item.title | 时间线上 MI 条目 |
| MI 编号 | string | main_item.code | 如 MI-0001 |
| MI 状态 | enum | main_item.status | MI 条目上显示 |
| MI 完成度 | int | main_item.completion | MI 条目上显示 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Loading | 骨架屏/加载动画 | 页面初始加载 |
| Empty | 空状态提示"暂无里程碑，点击创建" + 创建按钮 | 团队无里程碑 |
| Populated | 时间线+里程碑节点+MI 条目 | 有里程碑数据 |
| No Permission | 403 提示页 | 用户缺少 milestone:read 权限 |
| Error | 错误提示+重试按钮 | API 请求失败 |

### Validation Rules

- 创建里程碑：名称必填（1-100 字符），计划日期必填
- 编辑里程碑：名称和日期不可同时为空
- 拖拽归属：目标里程碑必须与 MI 属于同一团队

---

## UI Function 2: 创建/编辑里程碑弹窗

### Placement

- **Mode**: existing-page:/milestones
- **Target Page**: /milestones
- **Position**: 页面中央 modal overlay，从时间线页面触发

### Description

弹窗表单用于创建新里程碑或编辑现有里程碑。包含名称、计划日期字段。

### User Interaction Flow

1. 用户点击时间线页面的"创建里程碑"按钮（或点击空白区域）→ 弹出创建弹窗
2. 用户填写名称和计划日期 → 点击确认 → 创建成功 → 弹窗关闭 → 时间线刷新
3. 用户在里程碑详情面板点击"编辑" → 弹出编辑弹窗（预填当前值）
4. 用户修改后点击保存 → 保存成功 → 弹窗关闭 → 时间线刷新

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 名称 | string | milestone.name | 必填，1-100 字符 |
| 计划日期 | date | milestone.planned_date | 必填，日期选择器 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Create Mode | 标题"创建里程碑"，表单为空 | 点击创建按钮 |
| Edit Mode | 标题"编辑里程碑"，表单预填值 | 点击编辑按钮 |
| Submitting | 确认按钮 loading 状态 | 提交中 |
| Validation Error | 字段下方红色错误提示 | 校验失败 |

### Validation Rules

- 名称：必填，1-100 字符
- 计划日期：必填

---

## UI Function 3: 里程碑详情面板

### Placement

- **Mode**: existing-page:/milestones
- **Target Page**: /milestones
- **Position**: 页面右侧 slide-over panel，从时间线点击里程碑节点触发

### Description

展示单个里程碑的完整信息：名称、计划日期、状态、完成度、关联 MI 列表。提供编辑、删除、状态切换操作。

### User Interaction Flow

1. 用户在时间线上点击里程碑节点 → 右侧弹出详情面板
2. 面板显示：名称、计划日期、状态标签、完成度进度条、关联 MI 列表
3. 用户点击"编辑" → 弹出编辑弹窗（UF-2）
4. 用户点击状态切换按钮 → 状态变更 → 面板和时间线刷新
5. 用户点击"删除" → 确认弹窗 → 删除成功 → 面板关闭，时间线刷新
6. 用户点击关联 MI 列表中的某条 → 跳转到主事项详情页

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 名称 | string | milestone.name | 标题 |
| 计划日期 | date | milestone.planned_date | 日期格式 |
| 状态 | enum | milestone.status | 状态标签样式区分 |
| 完成度 | decimal | 计算值 | 进度条 + 百分比 |
| 关联 MI 列表 | list | main_items where milestone_key = this | 编号+标题+状态+完成度 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Loading | 骨架屏 | 面板打开 |
| Populated | 完整信息展示 | 数据加载完成 |
| Cancelled | 灰色样式，关联 MI 列表为空 | 里程碑已取消 |

### Validation Rules

- 状态切换：明确合法转换方向如下：
  - `not_started` → `in_progress` ✓ / `cancelled` ✓；→ `completed` ✗（按钮禁用，tooltip 提示"未开始的里程碑不可直接标记为已完成"）
  - `in_progress` → `completed` ✓ / `cancelled` ✓；→ `not_started` ✗（按钮禁用，tooltip 提示"进行中不可回退为未开始"）
  - `completed` → `cancelled` ✓；→ `in_progress` ✗ / `not_started` ✗（按钮禁用，tooltip 提示"已完成状态不可回退"）
  - `cancelled` → 任何状态 ✗（所有切换按钮禁用，tooltip 提示"已取消的里程碑不可恢复"）
- 删除：点击删除按钮后弹出确认弹窗（非 alert），用户确认后执行删除，取消则关闭弹窗不做操作

---

## UI Function 4: 事项清单页里程碑筛选

### Placement

- **Mode**: existing-page:/items
- **Target Page**: /items
- **Position**: 筛选栏区域，现有"状态"和"负责人"筛选器右侧

### Description

在事项清单页面的筛选栏增加里程碑下拉筛选器，在事项列表中显示里程碑标签。

### User Interaction Flow

1. 用户在事项清单页看到新增的"里程碑"筛选下拉框
2. 用户选择某个里程碑 → 列表过滤只显示该里程碑下的 MI
3. 用户选择"未分配" → 列表过滤只显示 milestone_key 为空的 MI
4. 用户选择"全部" → 显示所有 MI（默认值）
5. 每条 MI 记录旁边显示所属里程碑名称标签（如有）

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 里程碑列表 | list | 团队所有里程碑 | 下拉框选项 + "未分配"选项 |
| MI 的里程碑名称 | string | 关联里程碑的 name | MI 条目旁的标签 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Default | 筛选器显示"里程碑：全部" | 页面加载 |
| Filtered | 列表仅显示匹配的 MI | 用户选择筛选值 |

### Validation Rules

- 筛选值必须是当前团队内已存在的里程碑 bizKey 或 `all`（全部）或 `unassigned`（未分配），其他值不产生筛选效果并回退到 `all`
- 切换团队时筛选器重置为 `all`，下拉选项刷新为新团队的里程碑列表
- 下拉选项排除 `cancelled` 状态的里程碑

---

## UI Function 5: 主事项编辑弹窗里程碑选择器

### Placement

- **Mode**: existing-page:/items/:mainItemId
- **Target Page**: /items/:mainItemId
- **Position**: 编辑主事项 modal 中，现有"负责人"字段下方

### Description

在主事项编辑弹窗中增加"所属里程碑"下拉选择框，用于指定或修改 MI 的里程碑归属。

### User Interaction Flow

1. 用户打开主事项编辑弹窗 → 看到"所属里程碑"下拉框（显示当前值或"未分配"）
2. 用户选择某个里程碑 → 保存后 MI 的 milestone_key 更新
3. 用户选择"未分配" → 保存后 MI 的 milestone_key 置空
4. 用户不做修改 → 保持原值

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 里程碑列表 | list | 团队所有非 cancelled 里程碑 | 下拉框选项 |
| 当前里程碑 | string | main_item.milestone_key | 预填当前值 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Default | 下拉框显示当前里程碑名称 | 弹窗打开 |
| Empty | 下拉框显示"未分配" | MI 未绑定里程碑 |

### Validation Rules

- 里程碑选项排除 cancelled 状态的里程碑
- 下拉框为可选字段（允许"未分配"）

---

## UI Function 6: 表格视图里程碑列

### Placement

- **Mode**: existing-page:/table
- **Target Page**: /table
- **Position**: 表格列，在"状态"列和"预期完成"列之间

### Description

在表格视图中增加"里程碑"列，显示每条记录所属里程碑的名称。

### User Interaction Flow

1. 用户进入表格视图 → 看到新增的"里程碑"列
2. 未分配里程碑的记录显示"-"
3. 用户可按里程碑列排序和筛选

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 里程碑名称 | string | 关联里程碑的 name | 表格单元格，未分配则显示"-" |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Assigned | 里程碑名称 | MI 已绑定里程碑 |
| Unassigned | "-" | MI 未绑定里程碑 |

### Validation Rules

- 排序方向仅接受 `asc`（升序，默认）或 `desc`（降序），按里程碑名称字母序排列
- 筛选值必须是有效里程碑 bizKey 或 `all`（全部）或 `unassigned`（未分配），其他值不产生筛选效果并回退到 `all`
- 未分配里程碑的记录（milestone_key 为空）在筛选 `unassigned` 时显示，排序时按空值处理（排末尾）

---

## Page Composition

| Page | Type | UI Functions | Position Notes |
|------|------|-------------|----------------|
| /milestones | new | UF-1, UF-2, UF-3 | 新页面：时间线(UF-1) + 创建/编辑弹窗(UF-2) + 详情面板(UF-3) |
| /items | existing | UF-4 | 筛选栏增加里程碑筛选器 + MI 条目显示里程碑标签 |
| /items/:mainItemId | existing | UF-5 | 编辑弹窗增加里程碑选择器 |
| /table | existing | UF-6 | 表格增加里程碑列 |

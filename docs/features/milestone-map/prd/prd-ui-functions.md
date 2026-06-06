---
feature: "里程碑图"
---

# 里程碑图 — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

新增 1 个独立页面（/milestones，含两级视图）+ 修改 3 个现有页面（事项清单、主事项编辑、表格视图），提供里程碑图的创建管理、里程碑的可视化展示和与现有事项的集成。

## Navigation Architecture

- **Platform**: web

### Primary Navigation (shared across pages)

| # | Label | Target Page | Icon Keyword |
|---|-------|-------------|-------------|
| 1 | 里程碑图 | /milestones | milestone/timeline |

### Secondary Pages (navigated from a parent page)

| Page | Entry Point (UF# or action) | Return Target |
|------|-----------------------------|---------------|
| 时间线详情 | UF-1 点击里程碑图卡片 | /milestones（列表视图） |
| 里程碑详情面板 | UF-1 时间线中点击里程碑节点 | /milestones（时间线视图） |
| 主事项详情 | UF-1 点击关联的 MI 条目 | /items/:mainItemId |

### Navigation Rules

- 里程碑图页面从主导航进入，与事项清单/甘特图/周报/表格平级
- 页面默认显示里程碑图列表视图，点击卡片进入时间线视图
- 时间线视图通过面包屑导航返回列表视图
- 里程碑详情面板为 overlay（不离开当前页面）
- 从详情面板点击 MI 条目跳转到主事项详情页

---

## UI Function 1: 里程碑图页面（两级视图）

### Placement

- **Mode**: new-page
- **Target Page**: /milestones
- **Position**: 独立页面，主导航"事项清单"和"甘特图"之间

### Description

两级视图页面：
- **第一级（列表视图）**：展示团队所有里程碑图的卡片列表，支持按名称、负责人、状态三种方式筛选。每张卡片分四行展示：第一行名称+状态 Badge，第二行里程碑数量+事项数量+负责人（左右对齐），第三行计划时间跨度（左）+"整体进度"+进度条+百分比（右），底部为里程碑节点缩略图。
- **第二级（时间线视图）**：点击卡片进入该里程碑图的时间线视图。页面顶部为详情标题区（名称+可点击的状态 Badge，右上角编辑和删除按钮位于卡片外部），紧接基本信息卡片（负责人、计划开始、计划完成、整体进度四字段同一行左右对齐，下方分隔线后显示描述，描述最多三行溢出截断），再下方展示里程碑节点及关联 MI 的横向时间线。支持按名称、状态两种方式筛选。支持缩放、点击交互。

### User Interaction Flow

**列表视图：**
1. 用户从主导航进入 /milestones 页面
2. 系统渲染里程碑图卡片列表，支持三种筛选（按名称搜索、按负责人下拉、按状态下拉），依次排列
3. 用户点击某张卡片 → 进入该里程碑图的时间线视图

**时间线视图：**
4. 系统渲染时间线：顶部详情标题区（名称+可点击的状态 Badge，右上角编辑和删除按钮位于卡片外部），下方基本信息卡片（负责人、计划开始、计划完成、整体进度四字段同一行左右对齐，下方分隔线后显示描述，描述最多三行溢出截断，悬浮 Tooltip 展示完整内容），再下方横向时间轴，里程碑节点按计划完成时间排列，关联 MI 展示在对应里程碑下方。支持两种筛选（按名称搜索、按状态下拉）
5. 用户点击标题区的状态 Badge → 下拉显示可用状态转换选项（不可用的状态以灰色不可点击显示）→ 选择目标状态 → 状态变更 → 页面刷新
6. 用户可缩放时间轴（周/月/季切换控件）：改变时间轴刻度标签的粒度（周=每7天一条刻度线、月=每30天、季=每90天），不改变节点位置，不隐藏任何节点。节点过密时自动出现水平滚动条
7. 用户点击里程碑节点 → 弹出详情面板（UF-3）
8. 用户点击"+ 创建里程碑"按钮 → 弹出创建弹窗（UF-2）
9. 用户点击面包屑"里程碑图" → 返回列表视图

### Data Requirements

**列表视图：**

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 里程碑图名称 | string | milestone_map.map_name | 卡片标题；支持名称搜索筛选 |
| 负责人（PM） | string | milestone_map.assignee_key | 支持负责人下拉筛选 |
| 里程碑图状态 | enum | milestone_map.map_status | 规划中/已评审/待实施/实施中/已完成；支持状态筛选 |
| 里程碑数量 | int | 计算值 | 关联里程碑计数 |
| 事项数量 | int | 计算值 | 所有关联 MI 计数 |
| 整体进度 | decimal | 计算值 | 所有关联 MI completion 的平均值 |
| 计划开始时间 | date | milestone_map.planned_start_date | 卡片上显示日期跨度 |
| 计划完成时间 | date | milestone_map.planned_end_date | 卡片上显示日期跨度 |

**时间线视图（信息卡）：**

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 里程碑图名称 | string | milestone_map.map_name | 信息卡标题 |
| 状态 | enum | milestone_map.map_status | 信息卡标题旁 Badge |
| 负责人 | string | milestone_map.assignee_key | 信息卡显示 |
| 计划开始时间 | date | milestone_map.planned_start_date | 信息卡显示 |
| 计划完成时间 | date | milestone_map.planned_end_date | 信息卡显示 |
| 整体进度 | decimal | 计算值 | 信息卡显示进度条+百分比 |
| 描述 | string | milestone_map.map_desc | 信息卡显示，最多三行溢出截断，鼠标悬浮 Tooltip 展示完整内容 |

**时间线视图（时间轴）：**

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 里程碑名称 | string | milestone.milestone_name | 节点上显示；支持名称搜索筛选 |
| 计划完成时间 | date | milestone.expected_end_date | 决定时间轴位置 |
| 状态 | enum | milestone.milestone_status | not_started/in_progress/completed/cancelled；支持状态筛选 |
| 完成度 | decimal | 计算值 | 关联 MI completion 的平均值，空里程碑为 0 |
| 关联 MI 数量 | int | 计算值 | 悬停时显示 |
| MI 标题/编号/状态/完成度 | various | main_item.* | MI 条目上显示 |

### States

**列表视图：**

| State | Display | Trigger |
|-------|---------|---------|
| Loading | 骨架屏卡片 | 页面初始加载 |
| Empty | 空状态提示"暂无里程碑图" + 创建按钮 | 团队无里程碑图 |
| Populated | 卡片网格 | 有里程碑图数据 |
| Error | 错误提示+重试按钮 | API 请求失败 |

**时间线视图：**

| State | Display | Trigger |
|-------|---------|---------|
| Loading | 骨架屏 | 进入时间线 |
| Populated | 时间线+里程碑节点+MI 条目 | 有里程碑数据 |
| No Permission | 403 提示页 | 用户缺少 milestone:read 权限 |
| Error | 错误提示+重试按钮 | API 请求失败 |

### Validation Rules

- 创建里程碑图：名称必填（1-100 字符），描述可选
- 列表视图筛选：名称搜索为客户端模糊匹配；负责人下拉选项为团队成员列表；状态筛选值必须是有效状态枚举值或 `all`
- 时间线视图筛选：名称搜索为客户端模糊匹配；状态筛选值必须是有效状态枚举值或 `all`
- 里程碑节点不可重叠，节点间必须保持最小间距
- 删除里程碑图：仅 `planning` 状态的里程碑图可被删除；点击详情标题区右上角的删除按钮后弹出确认弹窗，用户确认后执行删除并返回列表视图

---

## UI Function 2: 创建/编辑里程碑弹窗

### Placement

- **Mode**: existing-page:/milestones
- **Target Page**: /milestones
- **Position**: 页面中央 modal overlay，从时间线视图触发

### Description

弹窗表单用于创建新里程碑或编辑现有里程碑。包含名称、计划完成时间、描述字段。

### User Interaction Flow

1. 用户点击时间线视图的"+ 创建里程碑"按钮 → 弹出创建弹窗
2. 用户填写名称和计划完成时间 → 点击确认 → 创建成功 → 弹窗关闭 → 时间线刷新
3. 用户在里程碑详情面板点击"编辑" → 弹出编辑弹窗（预填当前值）
4. 用户修改后点击保存 → 保存成功 → 弹窗关闭 → 时间线刷新

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 名称 | string | milestone.milestone_name | 必填，1-100 字符 |
| 计划完成时间 | date | milestone.expected_end_date | 必填，日期选择器 |
| 描述 | string | milestone.milestone_desc | 可选，多行文本 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Create Mode | 标题"创建里程碑"，表单为空 | 点击创建按钮 |
| Edit Mode | 标题"编辑里程碑"，表单预填值 | 点击编辑按钮 |
| Submitting | 确认按钮 loading 状态 | 提交中 |
| Validation Error | 字段下方红色错误提示 | 校验失败 |

### Validation Rules

- 名称：必填，1-100 字符
- 计划完成时间：必填
- 描述：可选

---

## UI Function 3: 里程碑详情面板

### Placement

- **Mode**: existing-page:/milestones
- **Target Page**: /milestones
- **Position**: 页面右侧 slide-over panel，从时间线点击里程碑节点触发

### Description

展示单个里程碑的完整信息：名称、计划完成时间、状态（Badge+下拉切换）、完成度、关联 MI 列表（每行含解绑按钮）。提供编辑、删除、快速添加事项操作。

### User Interaction Flow

1. 用户在时间线上点击里程碑节点 → 右侧弹出详情面板
2. 面板顶部：计划完成时间（左） + 状态 Badge（可点击切换） + 编辑按钮（右），三者同行
3. 用户点击状态 Badge → 下拉显示可用状态转换 → 选择目标状态 → 状态变更 → 面板和时间线刷新
4. 用户点击"编辑" → 弹出编辑弹窗（UF-2）
5. 用户点击"删除" → 确认弹窗 → 删除成功 → 面板关闭，时间线刷新
6. 用户点击关联 MI 列表中的某条 → 跳转到主事项详情页
7. 用户点击关联 MI 列表行右侧的 × 按钮 → 解除该 MI 与里程碑的绑定 → MI 列表刷新
8. 用户点击"关联事项"标题行右侧的"+ 添加"按钮 → 弹出快速添加事项弹窗（UF-3a） → 填写完整主事项表单并确认 → 创建 MI 并绑定当前里程碑 → 详情面板 MI 列表刷新

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 名称 | string | milestone.milestone_name | 面板标题 |
| 计划完成时间 | date | milestone.expected_end_date | 同行显示 |
| 状态 | enum | milestone.milestone_status | Badge 显示，点击弹出下拉 |
| 完成度 | decimal | 计算值 | 进度条 + 百分比 |
| 描述 | string | milestone.milestone_desc | 面板内显示 |
| 关联 MI 列表 | list | main_items where milestone_key = this | 编号+标题+状态+完成度+解绑按钮 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Loading | 骨架屏 | 面板打开 |
| Populated | 完整信息展示 | 数据加载完成 |
| Cancelled | 灰色样式，关联 MI 列表为空 | 里程碑已取消 |

### Validation Rules

- 状态切换：点击状态 Badge 弹出下拉菜单，仅显示合法转换选项（不可用的状态以灰色+不可点击显示）：
  - `not_started` → `in_progress` ✓ / `cancelled` ✓
  - `in_progress` → `completed` ✓ / `cancelled` ✓
  - `completed` → `cancelled` ✓
  - `cancelled` → 任何状态 ✗（无下拉选项）
- 删除：仅 `not_started` 和 `cancelled` 状态的里程碑可被删除；点击删除按钮后弹出确认弹窗（非 alert），用户确认后执行删除，取消则关闭弹窗不做操作
- 解绑：点击 MI 行右侧 × 按钮即解除绑定，显示撤销 toast

---

## UI Function 3a: 快速添加事项弹窗

### Placement

- **Mode**: existing-page:/milestones
- **Target Page**: /milestones
- **Position**: 页面中央 modal overlay，从详情面板"+ 添加"按钮触发

### Description

复用与事项清单页一致的主事项创建表单（CreateMainItemDialog），所属里程碑自动预填为当前里程碑且不可修改。

### User Interaction Flow

1. 用户在详情面板点击"+ 添加"按钮 → 弹出创建主事项弹窗
2. 表单字段与事项清单创建表单完全一致：标题、优先级+负责人、开始时间+预期完成时间、所属里程碑、描述
3. 所属里程碑字段自动预填当前里程碑名称，disabled 不可修改
4. 用户填写表单并点击"确认" → 创建 MI 并绑定当前里程碑 → 弹窗关闭 → 详情面板 MI 列表刷新

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 标题 | string | main_item.title | 必填，1-100 字符 |
| 优先级 | enum | main_item.priority | 必填，默认 P2，下拉选择 |
| 负责人 | string | main_item.assignee_key | 必填，团队成员下拉选择 |
| 开始时间 | date | main_item.start_date | 必填，默认今天 |
| 预期完成时间 | date | main_item.expected_end_date | 必填 |
| 所属里程碑 | string | 自动预填 | disabled，不可修改 |
| 描述 | string | main_item.description | 可选 |

### Validation Rules

- 与事项清单创建主事项表单校验规则一致：标题、负责人、开始时间、预期完成时间为必填
- 所属里程碑不可修改，自动传入当前里程碑 bizKey

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
- **Position**: 表格列，在"标题"列和"优先级"列之间

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

## UI Function 7: 创建/编辑里程碑图弹窗

### Placement

- **Mode**: existing-page:/milestones
- **Target Page**: /milestones
- **Position**: 页面中央 modal overlay，从列表视图触发

### Description

弹窗表单用于创建新里程碑图或编辑现有里程碑图。包含名称、负责人、计划开始时间、计划完成时间、描述字段。

### User Interaction Flow

1. 用户点击列表视图的"+ 创建里程碑图"按钮或虚线创建卡片 → 弹出创建弹窗
2. 用户填写名称、负责人、计划开始时间、计划完成时间和描述 → 点击创建 → 创建成功 → 弹窗关闭 → 列表刷新
3. 用户在里程碑图卡片上点击编辑 → 弹出编辑弹窗（预填当前值）
4. 用户修改后点击保存 → 保存成功 → 弹窗关闭 → 列表刷新

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 名称 | string | milestone_map.map_name | 必填，1-100 字符 |
| 负责人 | string | milestone_map.assignee_key | 必填，团队成员下拉选择 |
| 计划开始时间 | date | milestone_map.planned_start_date | 可选，日期选择器 |
| 计划完成时间 | date | milestone_map.planned_end_date | 可选，日期选择器 |
| 描述 | string | milestone_map.map_desc | 可选 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Create Mode | 标题"创建里程碑图"，表单为空 | 点击创建按钮 |
| Edit Mode | 标题"编辑里程碑图"，表单预填值 | 点击编辑按钮 |
| Submitting | 确认按钮 loading 状态 | 提交中 |

### Validation Rules

- 名称：必填，1-100 字符
- 负责人：必填，必须是当前团队成员
- 计划开始时间：可选；若填写，计划完成时间不得早于计划开始时间
- 计划完成时间：可选；若填写，不得早于计划开始时间

---

## Status Transition Business Rules

> 跨 UI Function 共用的状态流转约束。各 UF 的 Validation Rules 节引用此处规则。

| Rule ID | Rule | Enforced At |
|---------|------|-------------|
| BR-1 | 里程碑（Milestone）不可切换至终态（`completed`/`cancelled`），除非其下所有 MI 均已处于终态（`completed`/`cancelled`） | UF-3 状态切换 |
| BR-2 | 里程碑图（Milestone Map）不可切换至终态（`completed`），除非其下所有里程碑均已处于终态 | UF-1 信息卡状态切换 |
| BR-3 | 处于终态的 MI 不可变更 milestone_key；处于终态（cancelled）的里程碑不可接收新 MI | UF-5 里程碑选择器 |
| BR-4 | 删除约束：仅 `planning` 状态的里程碑图可删除（UF-1）；仅 `not_started` 和 `cancelled` 状态的里程碑可删除（UF-3） | UF-1、UF-3 删除操作 |

---

## Page Composition

| Page | Type | UI Functions | Position Notes |
|------|------|-------------|----------------|
| /milestones | new | UF-1, UF-2, UF-3, UF-3a, UF-7 | 新页面：列表+时间线(UF-1) + 创建/编辑里程碑弹窗(UF-2) + 详情面板(UF-3) + 快速添加事项(UF-3a) + 创建/编辑里程碑图弹窗(UF-7) |
| /items | existing | UF-4 | 筛选栏增加里程碑筛选器 + MI 条目显示里程碑标签 |
| /items/:mainItemId | existing | UF-5 | 编辑弹窗增加里程碑选择器 |
| /table | existing | UF-6 | 表格增加里程碑列 |

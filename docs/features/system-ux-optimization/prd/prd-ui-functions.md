---
feature: "System UX Optimization Batch"
---

# System UX Optimization Batch — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

本功能涉及以下 UI 变更：
- 事项清单页面（卡片视图 + 表格视图）的过滤器增强和终态排序
- 主事项详情页的子事项列表排序和删除按钮
- 子事项编辑弹窗新增开始时间字段
- 子事项详情页的移动和删除功能
- 待办转换表单的交互优化（描述置灰、必填校验、表单清空）
- 整体进度页面的状态过滤和甘特图修复
- 每周进展页面的后端过滤（无 UI 变更）
- 全局左侧菜单栏的团队选择器过滤
- 状态流转操作的错误提示

## Navigation Architecture

### Primary Navigation (shared across pages)

无新增页面。所有变更为现有页面的功能增强。

### Navigation Rules

- 所有变更为 existing-page 模式
- 无新增路由

## UI Function 1: 状态流转错误提示

### Placement

- **Mode**: existing-page
- **Target Page**: 主事项详情页 / 事项清单页（卡片视图、表格视图）
- **Position**: 状态流转按钮区域下方

### Description

状态流转失败时，在操作区域下方展示行内 Alert 错误消息，替代当前的2秒 tooltip。消息内容来自后端返回的具体原因。

### User Interaction Flow

1. 用户点击状态流转按钮
2. 后端返回不可流转的具体原因
3. 操作区域下方展示 Alert 行内错误消息（如"该主事项下还有未完成的子事项，无法关闭"）
4. 终态流转仍保留确认对话框

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 错误消息 | string | 后端流转接口返回 | 具体不可流转原因 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 正常 | 无提示 | 流转成功 |
| 错误 | Alert 行内消息 | 后端返回流转失败 |
| 确认 | 确认对话框 | 终态流转 |

### Alert 生命周期

- 错误 Alert 在展示后 **不会自动消失**，持续显示直到用户执行以下任一操作：再次点击状态流转按钮、手动关闭 Alert（提供关闭图标）、或离开当前页面
- 手动关闭 Alert 后，Alert 区域隐藏，不影响其他功能使用
- 重复的状态流转失败会更新 Alert 内容（替换为最新错误消息）

### Validation Rules

- 错误消息必须来自后端，前端不硬编码提示文案
- tooltip 提示方式被完全替代

---

## UI Function 2: 子事项编辑弹窗开始时间

### Placement

- **Mode**: existing-page
- **Target Page**: 子事项编辑弹窗
- **Position**: 现有字段之间，添加"开始时间"日期选择器

### Description

子事项编辑弹窗中新增"开始时间"字段，允许用户直接修改开始时间。

### User Interaction Flow

1. 用户打开子事项编辑弹窗
2. 看到"开始时间"字段，显示当前值
3. 修改开始时间
4. 点击保存，成功更新

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 开始时间 | date | SubItemUpdateReq.StartDate | 后端已支持 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 展示 | 当前开始时间 | 弹窗打开 |
| 编辑 | 日期选择器 | 用户点击字段 |
| 保存 | 保存成功提示 | 用户提交 |

### Validation Rules

- 日期格式由日期选择器组件保证
- 开始时间不得晚于结束时间（若结束时间已设置）
- 开始时间允许设置为过去日期（允许补录历史信息）
- 开始时间无需与父主事项时间范围对齐（子事项时间可独立于主事项）

---

## UI Function 3: 事项删除

### Placement

- **Mode**: existing-page
- **Target Page**: 主事项详情页 / 子事项详情页 / 子事项编辑界面
- **Position**: 操作按钮区域

### Description

PM 角色用户可见删除按钮，点击后弹出确认对话框，确认后软删除。主事项删除级联删除所有子事项。

### User Interaction Flow

1. PM 用户点击删除按钮
2. 弹出确认对话框，主事项删除提示"将同时删除 N 个子事项"
3. 用户确认
4. 执行软删除，页面更新

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 子事项数量 | number | 主事项详情 | 确认对话框中提示 |
| 权限码 | string | RBAC | main_item:delete, sub_item:delete |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 隐藏 | 无删除按钮 | 非 PM 角色 |
| 可用 | 删除按钮 | PM 角色 |
| 确认 | 确认对话框 | 用户点击删除 |
| 完成 | 删除成功，页面更新 | 确认后 |

### Validation Rules

- 删除按钮仅 PM 角色可见（前端基于权限码判断）
- 确认对话框必须显示子事项数量

---

## UI Function 4: 转换表单优化

### Placement

- **Mode**: existing-page
- **Target Page**: 待办→子事项转换表单 / 待办→主事项转换表单
- **Position**: 表单字段区域和提交按钮

### Description

三部分优化：(a) 待办→子事项描述字段置灰；(b) 负责人和优先级必填校验；(c) 关闭表单时清空字段。

### User Interaction Flow

1. 用户打开待办→子事项转换表单
2. 描述字段显示为灰色 disabled 状态，不可编辑
3. 用户打开待办→主事项/子事项转换表单
4. 未选负责人或优先级时提交按钮禁用，两字段标签显示必填标记（*）
5. 用户填写必填项后提交按钮启用
6. 用户提交或关闭表单
7. 所有字段清空，下次打开为空白表单

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 负责人 | select | 用户列表 | 必填 |
| 优先级 | select | 优先级选项 | 必填 |
| 描述 | text | 待办事项 | 待办→子事项时 disabled |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 描述置灰 | disabled 灰色样式 | 待办→子事项表单 |
| 提交禁用 | 提交按钮不可点击 | 必填项未填 |
| 提交可用 | 提交按钮可点击 | 必填项已填 |
| 清空 | 所有字段为空 | 表单关闭（含取消） |

### Validation Rules

- 负责人字段必填，标签显示 * 标记
- 优先级字段必填，标签显示 * 标记
- 描述字段在待办→子事项场景下 disabled
- 提交失败时字段内容保留不清空，便于用户修正后重试；仅在提交成功或用户手动关闭/取消表单时清空所有字段

---

## UI Function 5: 子事项排序

### Placement

- **Mode**: existing-page
- **Target Page**: 主事项详情页
- **Position**: 子事项列表区域

### Description

子事项列表按 id 倒序排列，最新创建的子事项排在最前。

### User Interaction Flow

1. 用户打开主事项详情页
2. 子事项列表按创建时间倒序展示

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 子事项列表 | list | 后端排序后返回 | ORDER BY id DESC |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 排序展示 | 最新子事项在前 | 页面加载 |

### Validation Rules

- 排序在服务端完成

---

## UI Function 6: 子事项移动

### Placement

- **Mode**: existing-page
- **Target Page**: 子事项详情页 / 子事项编辑界面
- **Position**: 操作按钮区域

### Description

提供"移动到其他主事项"功能，用户选择目标主事项后移动子事项。

### User Interaction Flow

1. 用户点击"移动到其他主事项"按钮
2. 弹出目标主事项选择器（搜索或列表）
3. 选择目标主事项
4. 系统校验：已关闭状态的目标禁止移动、同一主事项禁止移动
5. 弹出确认对话框
6. 用户确认后执行移动，编号自动更新

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 目标主事项 | select/search | 主事项列表 | 排除已关闭和当前主事项 |
| 新编号 | string | NextSubCode | 自动生成 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 选择目标 | 目标选择器 | 用户点击移动按钮 |
| 校验失败 | 错误提示 | 目标已关闭或同一主事项 |
| 确认 | 确认对话框 | 目标有效 |
| 完成 | 移动成功，页面更新 | 确认后 |

### Validation Rules

- 目标主事项不能为已关闭状态
- 目标主事项不能与源主事项相同

### 移动子事项 API 规格

**请求：**
- **Method**: `PUT /api/sub-items/{bizKey}/move`
- **Body**: `{ "targetMainItemBizKey": "string" }`

**响应：**
- **200 OK**: `{ "newSubCode": "string", "mainItemBizKey": "string" }` — 移动成功，返回新编号和目标主事项 bizKey
- **400 Bad Request**: `{ "message": "目标主事项已关闭" }` 或 `{ "message": "不能移动到同一主事项" }`
- **404 Not Found**: 子事项或目标主事项不存在

前端根据响应更新子事项所属主事项和编号，刷新页面数据。

---

## UI Function 7: 过滤穿透

### Placement

- **Mode**: existing-page
- **Target Page**: 事项清单页面（卡片视图 + 表格视图）
- **Position**: 现有过滤器区域

### Description

状态过滤器改为多选（Checkbox Group），负责人过滤穿透子事项。因子事项匹配的主事项展示视觉标识。

### User Interaction Flow

1. 用户在状态过滤器选择多个状态（Checkbox Group 多选）
2. 用户在负责人过滤器选择某负责人
3. 系统展示匹配的主事项 + 因子事项匹配而连带展示的主事项
4. 因子事项匹配的主事项卡片/行上显示"因子事项匹配"标识
5. 该主事项下仅展示匹配的子事项，非匹配子事项折叠隐藏
6. 未选择任何过滤器时展示全部事项

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 状态 | multi-select | 状态列表 | Checkbox Group |
| 负责人 | select | 用户列表 | 穿透到子事项 |
| 匹配标识 | badge | 过滤结果 | "因子事项匹配" |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 无过滤 | 全部事项 | 未选择任何过滤器 |
| 状态过滤 | 匹配主事项 | 选择了状态 |
| 穿透过滤 | 匹配主事项+连带主事项 | 选择了负责人 |
| 标识展示 | "因子事项匹配"标签 | 因子事项匹配 |

### Validation Rules

- 卡片视图和表格视图统一支持
- 状态和负责人过滤器可独立使用或组合使用
- 组合过滤时采用 **AND 逻辑**：主事项须同时满足状态条件 AND（直接负责人匹配 OR 子事项负责人匹配）。具体规则：(1) 仅选状态 → 展示状态匹配的主事项；(2) 仅选负责人 → 展示负责人直接匹配的主事项 + 因子事项负责人匹配而连带展示的主事项；(3) 同时选状态和负责人 → 主事项自身状态须匹配 AND（负责人直接匹配 OR 含负责人匹配的子事项），不满足状态条件的主事项即使子事项匹配负责人也不展示

### 过滤穿透 API 响应结构

后端过滤穿透查询的响应中，每个主事项须包含以下字段供前端区分匹配类型：

| Field | Type | Description |
|-------|------|-------------|
| `matchType` | enum | `direct`（主事项自身匹配）或 `indirect`（因子事项匹配而连带展示） |
| `matchedSubItemIds` | string[] | 当 matchType 为 `indirect` 时，列出匹配的子事项 bizKey 列表；前端仅展示这些子事项，其余折叠隐藏 |

前端根据 `matchType` 决定是否展示"因子事项匹配"标识，根据 `matchedSubItemIds` 决定展示哪些子事项。

---

## UI Function 8: 终态排序和进度页面过滤

### Placement

- **Mode**: existing-page
- **Target Page**: 事项清单页面（#11）/ 整体进度页面（#12）
- **Position**: 列表区域 / 过滤器区域

### Description

事项清单页面终态主事项排序到最后。整体进度页面新增状态过滤器（多选），首次加载默认选中"进行中"。

### User Interaction Flow

1. 用户查看事项清单页面，终态主事项自动排在最后
2. 用户打开整体进度页面，首次加载"进行中"复选框默认选中
3. 用户可修改状态过滤（多选）
4. 取消所有复选框后展示全部事项

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 状态过滤器 | multi-select | 状态列表 | Checkbox Group |
| 默认状态 | string | "进行中" | 首次加载自动选中 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 默认加载 | "进行中"选中 | 首次访问 |
| 自定义过滤 | 用户选择的状态 | 用户操作 |
| 零选择 | 全部事项 | 取消所有复选框 |

### Validation Rules

- 整体进度页面仅支持状态过滤，不支持负责人过滤
- 终态主事项始终排在过滤结果最后

---

## UI Function 9: 甘特图修复

### Placement

- **Mode**: existing-page
- **Target Page**: 整体进度页面甘特图区域
- **Position**: 时间轴和滚动容器

### Description

修复甘特图时间起点空白和 macOS 滚动条不显示两个问题。

### User Interaction Flow

1. 甘特图时间轴起始日期取最早开始时间前1天，终止日期取最晚结束后1天
2. macOS 用户 hover 容器时水平滚动条可见

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 时间范围 | date range | 可见主事项数据 | 自动计算 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 正常 | 无空白区域 | 甘特图加载 |
| 滚动条可见 | 水平滚动条 | macOS hover 容器 |

### Validation Rules

- 时间范围自动计算，不留空白
- 滚动条至少在 hover 时显示

---

## UI Function 10: 团队选择器过滤

### Placement

- **Mode**: existing-page
- **Target Page**: 全局左侧菜单栏
- **Position**: 菜单顶部团队下拉选择器

### Description

团队下拉选择器仅展示当前用户有权限的团队。

### User Interaction Flow

1. 用户登录后，团队列表 API 仅返回有权限的团队
2. 下拉选择器仅展示有权限的团队
3. 若仅有一个团队，可省略下拉或固定展示

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 团队列表 | list | 后端过滤后返回 | 基于用户权限 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 单团队 | 固定展示团队名 | 仅有一个团队 |
| 多团队 | 下拉选择器 | 多个团队 |

### Validation Rules

- 无特殊校验

---

## Page Composition

| Page | Type | UI Functions | Position Notes |
|------|------|-------------|----------------|
| 主事项详情页 | existing | UF-1（状态错误提示）, UF-3（删除按钮）, UF-5（子事项排序） | 操作按钮区域 + 子事项列表 |
| 子事项编辑弹窗 | existing | UF-2（开始时间）, UF-4（转换表单优化）, UF-6（移动按钮） | 表单字段 + 操作按钮 |
| 子事项详情页 | existing | UF-3（删除按钮）, UF-6（移动按钮） | 操作按钮区域 |
| 待办转换表单 | existing | UF-4（置灰/必填/清空） | 表单字段 + 提交按钮 |
| 事项清单页（卡片/表格） | existing | UF-1（状态错误提示）, UF-7（过滤穿透）, UF-8（终态排序） | 过滤器区域 + 列表 |
| 整体进度页面 | existing | UF-8（状态过滤）, UF-9（甘特图修复） | 过滤器区域 + 甘特图 |
| 每周进展页面 | existing | 无 UI 变更（后端过滤） | — |
| 全局左侧菜单栏 | existing | UF-10（团队选择器过滤） | 菜单顶部 |

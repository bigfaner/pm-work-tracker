---
feature: "weekly-stats-optimization"
status: Draft
---

# 每周进展统计优化 — PRD Spec

> PRD Spec: defines WHAT the feature is and why it exists.

## 需求背景

### 为什么做（原因）

每周进展页顶部统计栏存在两个持续问题：

**规则不透明**：统计数字缺乏说明，用户无法直接理解口径。典型混淆：`justCompleted`（本周新完成）的事项同时被计入活跃总数，该行为未文档化。近 5 次 sprint review 中有 3 次出现团队成员无法解释"活跃"数字构成的情况（"新完成的事项为什么也算活跃？"）。

**状态覆盖不全**：当前仅统计 4 个维度（活跃、新完成、进行中、阻塞中），缺少 `pending`（未开始）、`pausing`（暂停中）、逾期中三种状态的可见性。PM 每周须逐一展开子事项才能判断是否存在未启动或逾期事项，统计栏无法支撑快速决策。以当前团队规模（约 20 个活跃子事项）为例，每次周会前手动逐项检查约需 3–5 分钟。

### 要做什么（对象）

将每周进展页顶部统计栏从 4 个卡片扩展为 7 个，覆盖所有决策相关状态，并为每个卡片添加 hover tooltip 说明统计口径，使规则自解释。

### 用户是谁（人员）

- **PM / 项目负责人**：每周通过统计栏快速判断团队整体进展状态，识别风险信号（未启动、逾期、阻塞）
- **团队成员**：查看自己负责的子事项在统计中的归属，理解数字含义

## 需求目标

| 目标 | 量化指标 | 说明 |
|------|----------|------|
| 消除统计规则混淆 | 周会前无需解释统计口径（0 次混淆/sprint） | tooltip 使规则自解释 |
| 减少手动检查时间 | 周会前检查时间从 3–5 分钟降至 <1 分钟 | 统计栏直接呈现未启动/逾期/暂停数量 |
| 状态全覆盖 | 7 个卡片覆盖所有 6 种子事项状态的决策相关维度 | pending/pausing/overdue 新增可见 |

## Scope

### In Scope
- [x] 后端：`WeeklyStats` DTO 新增 `pending`、`pausing`、`overdue` 三个字段
- [x] 后端：`buildWeeklyGroups` 补充三种状态的计数逻辑
- [x] 后端：新增 `buildWeeklyGroups` 单元测试，覆盖 7 个卡片的计数规则（含条件优先级 fixture）
- [x] 前端：`StatsBar` 组件从 4 列扩展为 7 列，响应式布局
- [x] 前端：每个统计卡片添加 hover tooltip，展示统计规则说明
- [x] 前端：tooltip 键盘可访问性（Tab 聚焦 + `aria-describedby`）
- [x] 前端：`StatsBar` tooltip 渲染单元测试
- [x] 前端：`isOverdue` 函数签名更新（`referenceDate` 改为必传参数，移除 `today` 默认值）及所有现有调用方同步更新

### Out of Scope
- 统计数据的历史趋势图表
- 按主事项维度的分组统计
- 统计数据的导出功能
- 主事项（MainItem）的状态统计

## 流程说明

### 业务流程说明

用户访问每周进展页时，系统根据所选周次计算统计数据并渲染统计栏。统计栏展示 7 个卡片，每个卡片对应一类风险信号。用户可通过 hover（桌面端）或点击（移动端）查看每个卡片的统计规则说明。

**"本周活跃"判定逻辑**（所有状态统计的前提）：

子事项满足以下任一条件即为本周活跃：
1. 本周内有进展记录（progressRecord.createdAt ∈ [weekStart, weekEnd)）
2. 日期范围与本周重叠：创建于本周结束前，且未在本周开始前完成/关闭（actualEndDate < weekStart）

**条件优先级**：条件 1 优先于条件 2。若子事项本周内有进展记录，则无论其 actualEndDate 是否早于 weekStart，该事项均计为本周活跃。

### 业务流程图

```mermaid
flowchart TD
    Start([用户访问每周进展页]) --> SelectWeek[选择周次]
    SelectWeek --> FetchData[后端获取子事项 + 进展记录]
    FetchData --> FetchOK{请求成功?}
    FetchOK -->|否 - 超时 / 500 / 权限拒绝| FetchError[统计栏所有卡片显示"-"并提示加载失败]
    FetchOK -->|是| ComputeActive{判断本周活跃}
    FetchError --> End
    ComputeActive -->|有进展记录 OR 日期重叠| Active[计入活跃]
    ComputeActive -->|不满足任一条件| Inactive[不计入统计]
    Active --> ClassifyStatus{按状态分类}
    ClassifyStatus -->|progressing| InProgress[进行中 +1]
    ClassifyStatus -->|blocking| Blocked[阻塞中 +1]
    ClassifyStatus -->|pending| NotStarted[未开始 +1]
    ClassifyStatus -->|pausing| Paused[暂停中 +1]
    ClassifyStatus -->|completed + actualEndDate ∈ 本周| NewlyDone[本周新完成 +1]
    Active --> CheckOverdue{expectedEndDate < weekEnd AND status ∉ completed/closed}
    CheckOverdue -->|是| Overdue[逾期中 +1]
    CheckOverdue -->|否| NoOverdue[不计入逾期]
    InProgress & Blocked & NotStarted & Paused & NewlyDone & Overdue --> RenderStats[渲染 7 个统计卡片]
    RenderStats --> UserHover{用户 hover / 点击卡片}
    UserHover -->|桌面端 hover| ShowTooltip[显示 tooltip 说明]
    UserHover -->|移动端点击| ToggleTooltip[展开/收起 tooltip]
    ShowTooltip & ToggleTooltip --> End([结束])
    Inactive --> End
    NoOverdue --> End
```

## 功能描述

### 5.1 统计栏（StatsBar）

**数据来源**：后端 `GET /v1/teams/:teamId/views/weekly?weekStart=YYYY-MM-DD` 返回的 `stats` 字段

**显示范围**：当前所选周次内，团队所有子事项的聚合统计

**数据权限**：与每周进展页一致，仅展示当前团队数据

**页面类型**：仪表盘（统计卡片区域）

**统计卡片字段**：

| # | 卡片名 | 字段名 | 统计规则 | Tooltip 说明文本 |
|---|--------|--------|----------|-----------------|
| 1 | 本周活跃 | `activeSubItems` | 本周活跃的子事项总数（含本周新完成） | 本周有进展记录，或计划周期与本周重叠的子事项总数 |
| 2 | 本周新完成 | `newlyCompleted` | status=completed AND actualEndDate ∈ [weekStart, weekEnd] | 本周内实际完成（actualEndDate 落在本周）的子事项数 |
| 3 | 进行中 | `inProgress` | status=progressing AND 本周活跃 | 状态为"进行中"且本周活跃的子事项数 |
| 4 | 阻塞中 | `blocked` | status=blocking AND 本周活跃 | 状态为"阻塞中"且本周活跃的子事项数 |
| 5 | 未开始 | `pending` | status=pending AND 本周活跃 | 已创建但尚未启动（状态为 pending）且本周活跃的子事项数 |
| 6 | 暂停中 | `pausing` | status=pausing AND 本周活跃 | 状态为"暂停中"且本周活跃的子事项数 |
| 7 | 逾期中 | `overdue` | expectedEndDate < weekEnd AND status ∉ {completed, closed} AND 本周活跃 | 计划截止日在本周结束前已过、尚未完成/关闭且本周活跃的子事项数 |

**数学不变式**：
- 卡片 2（新完成）⊆ 卡片 1（活跃）：新完成的事项也计入活跃总数
- 卡片 3+4+5+6 之和 ≤ 卡片 1（活跃总数）：completed 状态不计入 3-6
- 卡片 7（逾期）与卡片 3-6 可能重叠（一个事项可以同时是 progressing 且逾期）

### 5.2 Tooltip 交互

**桌面端**：鼠标悬停卡片 300ms 后显示 tooltip，鼠标移出后隐藏

**移动端 fallback**：触摸设备无 hover 事件，点击卡片展开 tooltip，再次点击收起

**键盘可访问性**：Tab 键可依次聚焦所有 7 个卡片，聚焦时 tooltip 可见；每个卡片元素携带 `aria-describedby` 属性，其值指向对应 tooltip 内容节点的 `id`

**Tooltip 位置**：默认显示在卡片正上方（placement=top）；若上方空间不足则自动翻转至下方

### 5.3 响应式布局

| 断点 | 布局 |
|------|------|
| ≥1280px | 7 列单行 |
| 768px–1279px | 4+3 两行（flex-wrap） |
| <768px | 2 列多行 |

### 5.4 关联性需求改动

| 序号 | 涉及项目 | 功能模块 | 关联改动点 | 更改后逻辑说明 |
|------|----------|----------|------------|----------------|
| 1 | 后端 | `WeeklyStats` DTO | 新增 3 个字段 | 新增 `pending int`、`pausing int`、`overdue int`，Go struct 新增字段默认零值，JSON 序列化向后兼容 |
| 2 | 后端 | `buildWeeklyGroups` | 补充计数逻辑 | 在现有 `inProgress`/`blocked` 计数基础上，新增 `pending`/`pausing`/`overdue` 三种状态的计数分支 |
| 3 | 前端 | `WeeklyStats` 类型定义 | 新增 3 个字段 | `types/index.ts` 中 `WeeklyStats` 接口新增 `pending`、`pausing`、`overdue` 字段 |
| 4 | 前端 | `isOverdue` 函数 | 参数签名更新 | `lib/status.ts` 中 `isOverdue(item, referenceDate)` 将 `referenceDate` 改为必传参数（移除 `today` 默认值）；调用方统一传入 `weekEnd`；`weekEnd` 为 `undefined` 时调用方须抛出错误，不允许静默回退到 `today`；需同步更新所有现有调用方 |

## 其他说明

### 性能需求
- 响应时间：统计计算在现有 `WeeklyComparison` 接口内完成，不新增接口调用；新增计数逻辑不得使该接口 P99 响应时间超出现有基线 +50ms
- 并发量：与现有每周进展页一致
- 兼容性：支持主流桌面浏览器（Chrome/Firefox/Safari/Edge 最新版）；移动端支持 iOS Safari / Android Chrome

### 数据需求
- 数据埋点：无新增埋点需求
- 数据初始化：无
- 数据迁移：无（仅新增 DTO 字段，无数据库变更）

### 监控需求
- 无新增监控需求，复用现有 `/views/weekly` 接口监控

### 安全性需求
- 无新增安全需求，复用现有团队权限校验

---

## 质量检查

- [x] 需求标题是否概括准确
- [x] 需求背景是否包含原因、对象、人员三要素
- [x] 需求目标是否量化
- [x] 流程说明是否完整
- [x] 业务流程图是否包含（Mermaid 格式）
- [x] 功能描述是否完整（统计卡片字段/tooltip 交互/响应式布局/关联改动）
- [x] 关联性需求是否全面分析
- [x] 非功能性需求（性能/数据/监控/安全）是否考虑
- [x] 所有表格是否填写完整
- [x] 是否有歧义或模糊表述
- [x] 是否可执行、可验收

---
created: 2026-04-23
author: fanhuifeng
status: Draft
---

# Proposal: 每周进展统计优化

## Problem

每周进展页顶部统计栏存在两个问题：

**规则不透明**：用户无法直接看到每个数字的统计口径，容易对数字产生疑惑。**已知实例**：`justCompleted`（本周新完成）的事项同时被计入 `activeSubItems`，该行为未文档化；近 5 次 sprint review（2026-02-21、2026-03-07、2026-03-21、2026-04-04、2026-04-18）中有 3 次（2026-03-07、2026-03-21、2026-04-04）出现团队成员无法解释"活跃"数字构成的情况（典型反应："新完成的事项为什么也算活跃？"；2026-03-07 会议记录见 docs/sprint-reviews/2026-03-07.md；2026-03-21 和 2026-04-04 两次的混淆发生在会后口头讨论中，未形成书面记录），确认了规则不透明导致的持续混淆。

**状态覆盖不全**：当前只统计 4 个维度（活跃、新完成、进行中、阻塞中），缺少 `pending`（未开始）、`pausing`（暂停中）、逾期中三种状态的可见性。子事项共有 6 种状态（pending / progressing / blocking / pausing / completed / closed），但统计栏无法反映全貌。

**推迟代价**：PM 每周须逐一展开子事项才能判断是否存在未启动或逾期事项，统计栏无法支撑快速决策。以当前团队规模（约 20 个活跃子事项）为例，每次周会前手动逐项检查约需 3–5 分钟；子事项数量每翻倍，该成本等比增长。

## Proposed Solution

**扩展统计栏至 7 个卡片，并为每个卡片添加 hover tooltip 说明统计规则。**

### 统计规则定义

首先定义"本周活跃"的基础条件（所有状态统计的前提）：

> **本周活跃** = 子事项满足以下任一条件：
> 1. 本周内有进展记录（progressRecord.createdAt ∈ [weekStart, weekEnd)）
> 2. 日期范围与本周重叠：创建于本周结束前，且未在本周开始前完成/关闭（actualEndDate < weekStart）
>
> **条件优先级**：条件 1 优先于条件 2。若子事项本周内有进展记录（满足条件 1），则无论其 `actualEndDate` 是否早于 `weekStart`，该事项均计为本周活跃。典型场景：某事项已在上周标记完成（`actualEndDate < weekStart`），但本周补录了一条进展记录——该事项应计入本周活跃，因为它在本周产生了实际活动。

| # | 卡片名 | 统计规则 | Tooltip 说明 |
|---|--------|----------|-------------|
| 1 | 本周活跃 | 本周活跃的子事项总数（含本周新完成） | 本周有进展记录，或计划周期与本周重叠的子事项总数 |
| 2 | 本周新完成 | status=completed AND actualEndDate ∈ [weekStart, weekEnd] | 本周内实际完成（actualEndDate 落在本周）的子事项数 |
| 3 | 进行中 | status=progressing AND 本周活跃 | 状态为"进行中"且本周活跃的子事项数 |
| 4 | 阻塞中 | status=blocking AND 本周活跃 | 状态为"阻塞中"且本周活跃的子事项数 |
| 5 | 未开始 | status=pending AND 本周活跃 | 已创建但尚未启动（状态为 pending）且本周活跃的子事项数 |
| 6 | 暂停中 | status=pausing AND 本周活跃 | 状态为"暂停中"且本周活跃的子事项数 |
| 7 | 逾期中 | expectedEndDate < weekEnd AND status ∉ {completed, closed} AND 本周活跃 | 计划截止日在本周结束前已过、尚未完成/关闭且本周活跃的子事项数 |

**说明**：
- 卡片 2（本周新完成）与卡片 1（本周活跃）存在包含关系：新完成的事项也计入活跃总数，这是预期行为。
- 卡片 3-6 之和 ≤ 卡片 1（活跃总数），因为 completed 状态的事项不计入 3-6。
- 卡片 7（逾期中）与卡片 3-6 可能重叠（一个事项可以同时是 progressing 且逾期）。

### UI 行为规格

| 属性 | 规格 |
|------|------|
| 卡片顺序 | 本周活跃 → 本周新完成 → 进行中 → 阻塞中 → 未开始 → 暂停中 → 逾期中（与规则表 #1–7 一致） |
| 响应式断点 | ≥1280px：7 列单行；768px–1279px：4+3 两行（flex-wrap）；<768px：2 列多行 |
| Hover 触发延迟 | 300ms（防止鼠标划过时频繁弹出） |
| Tooltip 位置 | 默认显示在卡片正上方（`placement=top`）；若上方空间不足则自动翻转至下方 |
| 移动端 fallback | 触摸设备无 hover，改为点击卡片展开/收起 tooltip |

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| 按状态全覆盖（6 个状态各一个卡片） | 完整、对称 | completed/closed 的全量历史计数单调递增，不反映本周变化，对周会决策无增量信息（本周完成情况已由卡片 2 覆盖）；且 closed 状态在周视图中无需关注 | 否决 |
| 保持 4 个卡片，只加 tooltip | 改动最小，后端零变更 | 仍缺少 pending/pausing/overdue 可见性，PM 无法从统计栏判断是否有未启动或逾期事项 | 否决 |
| 7 个卡片 + tooltip（本方案） | 覆盖所有决策相关状态，tooltip 使规则自解释，解决已知混淆实例 | 布局从 4 列变 7 列，需调整 UI；后端 DTO 新增 3 个字段；tooltip 需实现键盘焦点触发与 `aria-describedby` 关联（屏幕阅读器可访问性）；需在三个断点（≥1280px / 768–1279px / <768px）分别验证布局；移动端需额外实现点击展开/收起 fallback | 采纳 |

**方案选择说明**：方案 1 的 completed/closed 卡片显示的是全量历史累计数，该数字在每次周会时只会单调增长，无法回答"本周完成了多少"这一核心问题——而这正是卡片 2（本周新完成）已经覆盖的内容。方案 1 用一个信息量更低的数字替换了卡片 2，是降级而非扩展。至于"7 个卡片是否稀释注意力"：方案 1 的 6 个卡片中有 2 个（completed/closed）对周会决策无用，实际有效信息密度低于本方案的 7 个卡片；本方案每个卡片对应一类需要 PM 关注的风险信号（未启动、逾期、阻塞等），卡片数量由决策维度决定，而非越少越好。方案 2 的缺陷是阻断性的——tooltip 只解决"数字含义不清"的问题，无法补全缺失的维度，PM 仍须手动展开子事项才能判断是否存在未启动或逾期风险。方案 3 的代价（7 列布局 + DTO 新增 3 字段）是有界且可控的：布局通过 flex-wrap 降级，DTO 变更向后兼容。

## Scope

**预计交付**：1 个 sprint（约 5 个工作日）。

### In Scope
- 后端：`WeeklyStats` DTO 新增 `pending`、`pausing`、`overdue` 三个字段
- 后端：`buildWeeklyGroups` 补充三种状态的计数逻辑
- 后端：新增 `buildWeeklyGroups` 单元测试，覆盖 7 个卡片的计数规则（含条件优先级 fixture，见成功标准）
- 前端：`StatsBar` 组件从 4 列扩展为 7 列
- 前端：每个统计卡片添加 hover tooltip，展示统计规则说明
- 前端：`StatsBar` tooltip 渲染单元测试，覆盖 7 个卡片的 tooltip 文本与方案表格字符串完全一致
- 文档：在本提案中明确统计规则，作为后续实现的参考

### Out of Scope
- 统计数据的历史趋势图表
- 按主事项维度的分组统计
- 统计数据的导出功能
- 主事项（MainItem）的状态统计

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 7 列布局在小屏幕上溢出 | 中 | 低 | 使用 flex-wrap 或 grid 自适应，必要时折叠为两行 |
| "逾期中"定义与前端 `isOverdue` 函数不一致 | 低 | 中 | 更新 `lib/status.ts` 中的 `isOverdue(item, referenceDate)` 签名，将 `referenceDate` 默认值从 `today` 改为调用方传入的 `weekEnd`，保持后端与前端判断基准一致 |
| DTO 向后兼容性：`WeeklyStats` 新增字段可能影响现有调用方 | 低 | 中 | Go struct 新增字段默认零值，JSON 序列化向后兼容；上线前 grep 所有 `WeeklyStats` 引用，确认无按字段位置解构的调用方 |
| 测试覆盖缺口：`buildWeeklyGroups` 新增计数逻辑无单元测试，存在回归风险 | 高 | 中 | 遵循项目 TDD 规范，先写覆盖 7 个卡片规则的单元测试（含 fixture，见成功标准），再实现逻辑 |

## Success Criteria

- [ ] 统计栏显示 7 个卡片，布局在 1280px 宽度下不溢出，在 768px 宽度下卡片自动折行且无内容截断，在 <768px 宽度下呈现 2 列多行布局且无内容截断
- [ ] 触摸设备上点击卡片可展开 tooltip，再次点击收起（无 hover 事件时的 fallback 行为）
- [ ] 后端计数逻辑通过以下 fixture 验证：给定 7 个子事项（A: progressing + 本周有进展记录；B: pending + 日期与本周重叠、无进展记录；C: pausing + 本周有进展记录 + expectedEndDate < weekEnd；D: completed + actualEndDate ∈ [weekStart, weekEnd]；E: blocking + 本周有进展记录 + expectedEndDate < weekEnd；F: completed + actualEndDate < weekStart + 本周有进展记录；G: closed + actualEndDate < weekStart + 无进展记录），期望输出：活跃=6，新完成=1，进行中=1，阻塞中=1，未开始=1，暂停中=1，逾期中=2。其中 F 验证条件优先级：actualEndDate 早于本周但因本周有进展记录（条件 1）仍计为活跃，且不计入新完成；G 验证 closed 排除路径：actualEndDate < weekStart 且无进展记录，不计入活跃（活跃仍为 6，而非 7）
- [ ] 上述 fixture 同时验证数学不变式：进行中(1) + 阻塞中(1) + 未开始(1) + 暂停中(1) = 4 ≤ 活跃(6)
- [ ] 每个卡片的 tooltip 文本与方案表格"Tooltip 说明"列的字符串完全一致（字符串匹配，非仅检查存在性）
- [ ] 每个统计卡片元素携带 `aria-describedby` 属性，其值指向对应 tooltip 内容节点的 `id`；通过 Tab 键可依次聚焦所有 7 个卡片，聚焦时 tooltip 可见（键盘可访问性）
- [ ] 现有 4 个统计的数字与优化前保持一致（无回归）

## Next Steps

- Proceed to `/write-prd` to formalize requirements

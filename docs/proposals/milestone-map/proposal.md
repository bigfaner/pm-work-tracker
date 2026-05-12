---
created: 2026-05-12
author: "fanhuifeng"
status: Draft
---

# Proposal: 里程碑图

## Problem

当前 MainItem 是扁平列表，PM 缺乏跨事项的阶段/节点视角，无法按关键里程碑跟踪多个事项的整体进展。Gantt 视图虽然有时间线，但以单个事项为中心，无法体现"一组事项共同达成某个目标"的分组关系。

### Evidence

- 现有 4 个视图（事项列表/甘特图/周报/表格）均为事项维度，无里程碑维度
- MainItem 之间无分组机制，PM 只能通过命名约定或记忆来管理阶段关系
- 甘特图展示单个 MI 的时间线，缺少"一组 MI 对应一个交付节点"的视图
- **证据来源说明**：以上痛点基于开发者对 PM 日常工作流的观察推断（命名约定"v1.1-需求阶段"等实际存在于数据中），尚未进行正式用户访谈或问卷调查。方案开发前将在 PRD 阶段安排 PM 验证环节，确认需求优先级

### Urgency

当前单团队 MainItem 数量已达 30–50 个级别，部分项目分 3–4 个交付阶段（如需求确认、开发、测试、上线）。PM 目前通过命名约定（如"v1.1-需求阶段"）和记忆管理阶段关系，每次跨阶段状态盘点需要逐条扫描甘特图或列表视图，单个项目每周约需 20–30 分钟人工对齐进度。缺少里程碑维度导致：无法一眼识别哪个阶段有延期风险，跨阶段汇报依赖手工整理。

成本量化：假设 5 个活跃项目，每个项目每周 25 分钟手动阶段跟踪，月浪费约 8 小时 PM 工时。里程碑视图可将阶段盘点时间压缩至 5 分钟/项目。

## Proposed Solution

新增 Milestone 实体，作为团队级别的阶段节点。一个里程碑可绑定多个 MainItem（多对一关系），通过独立的时间线图页面可视化展示。

核心要素：
- **Milestone 实体**：名称、计划日期、四态状态机（not_started/in_progress/completed/cancelled）、自动计算完成度
- **关联模型**：MainItem.milestone_key 外键，一个 MI 只属于一个里程碑（可空）
- **时间线图**：横向时间轴，里程碑为节点，关联 MI 按时间排列并连线到对应里程碑。性能目标：支持单个团队最多 20 个里程碑、200 个 MI（~10 个 MI/里程碑）在视口内渲染，初始绘制 < 500ms、交互帧率 ≥ 30fps。超出 200 个 MI 时启用分页加载（按里程碑分组，默认展开前 5 个里程碑，其余折叠）
- **交互**：点击查看详情/创建、拖拽改变 MI 归属、时间轴缩放（周/月/季）
- **独立权限**：milestone:create/update/delete/read 权限码

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Do nothing | 零开发成本 | 无法解决跨事项分组跟踪问题，PM 只能用命名约定 | Rejected: 核心功能缺失 |
| 只做数据层，不做可视化 | 开发量小（~2 天），能先分组 | 没有可视化，分组价值大幅缩水，PM 体验差 | Rejected: 不做可视化价值有限 |
| 用标签替代里程碑 | 零开发，灵活 | 标签无日期/状态/完成度，无法做时间线可视化 | Rejected: 能力不足 |
| 分阶段交付：v1 数据层+列表视图，v2 时间线可视化 | 风险低，v1 约 3 天可交付数据层和分组列表，快速验证里程碑概念 | 分两次交付增加上下文切换成本；列表视图无法体现时间关系，PM 仍需手动判断阶段进展；v1→v2 间隔期间用户体验割裂 | Rejected: 列表视图无法验证核心价值（时间线视角），两次交付的总工时 > 一次性交付 |
| 完整里程碑图（本方案） | 一次性交付完整体验（数据+可视化+交互），用户无需经历功能残缺的过渡期 | 开发量约 5–7 天（数据模型+API 2 天，时间线页面 2–3 天，权限+测试 1–2 天） | Selected: 完整方案总工时仅比分阶段 v1 多 2–4 天，但避免了两轮部署和用户适应成本。时间线可视化是里程碑概念的核心价值，拆分交付无法验证这一价值 |

## Scope

### In Scope

- 里程碑数据模型（新表 pmw_milestones + MainItem 加 milestone_key 外键）
- CRUD API（创建/读取/更新/删除/列表）
- 独立 RBAC 权限码（milestone:create/update/delete/read）
- 时间线可视化页面（/milestones）
- 交互：点击查看详情、点击创建里程碑、拖拽改变 MI 归属、时间轴缩放
- 完成度自动计算（关联 MI 完成度的简单平均值）
- 四态状态机（not_started/in_progress/completed/cancelled）
- 软删除 + 解绑关联 MI
- 数据库迁移（SQLite + MySQL 双 schema）

### Out of Scope

- 里程碑报表/导出（CSV、PDF）
- 通知提醒（里程碑到期/延期通知）
- 甘特图集成（现有甘特图不显示里程碑标记）
- 状态变更历史记录（类似 StatusHistory）
- 里程碑级别的进度记录（类似 ProgressRecord）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 时间线可视化复杂度高，拖拽+缩放实现困难 | Medium | High | 优先评估 vis-timeline（活跃维护、支持拖拽+缩放+分组、MIT 许可）；备选 react-calendar-timeline。评估标准：(1) 支持横向时间轴+节点分组 (2) 拖拽 API 可捕获目标节点 (3) 包体积 < 200KB gzip。若两库均不满足，v1 回退到静态卡片布局（无拖拽、无缩放），拖拽作为 v1.1 增强 |
| MainItem 加外键影响现有 API/查询性能 | Low | Medium | milestone_key 可空且无 DDL 外键约束，仅加索引。迁移步骤：(1) ALTER TABLE 加列（NULL, 默认空）(2) 加索引 (3) 部署新代码。回滚方案：DROP COLUMN。现有查询不受影响因为 WHERE 条件不涉及此列 |
| 里程碑与 MI 的绑定/解绑操作与现有编辑流程冲突 | Low | Medium | MI 编辑页增加里程碑选择器（下拉框），复用现有 MainItem.update API 扩展 milestone_key 字段。时间线拖拽仅作为快捷入口，最终走同一 update API，避免两套写入路径。冲突场景：MI 已属里程碑 A，拖到 B 时，API 先解绑 A 再绑 B，单一事务保证一致性 |

## Success Criteria

- [ ] CRUD 可验证：通过 API 创建里程碑后，GET 列表返回该记录（status=not_started, completion=0）；编辑名称/日期后再次 GET 验证变更；删除（软删除）后列表不再返回
- [ ] 关联/解绑：创建里程碑 M1 后，通过 MI 编辑 API 设置 milestone_key 绑定 3 个 MI（A/B/C），解绑 B 后 GET M1 返回 2 个关联 MI。空里程碑（0 个 MI）creation=0，正常展示
- [ ] 时间线图：渲染包含 3+ 里程碑的页面，每个里程碑节点可见且显示关联 MI 数量。缩放控件切换周/月/季后，时间轴刻度标签对应变化。cancelled 状态里程碑以灰色样式区分
- [ ] 完成度计算：里程碑绑定 3 个 MI（completion 分别为 100/50/0），milestone completion 显示 50.00（DECIMAL(5,2) 平均值）。新增进度记录更新某 MI completion 后，里程碑 completion 在下次 GET 时同步更新
- [ ] 拖拽归属变更：将 MI-A 从里程碑 M1 拖至 M2，操作完成后 GET M1 关联列表不含 A、GET M2 关联列表包含 A。拖至空白区域（无里程碑）等效于解绑，MI milestone_key 置空
- [ ] 权限控制：拥有 milestone:create 权限的用户可创建，无此权限的用户点击创建按钮后返回 403，按钮显示为禁用态（灰色+tooltip "无权限"）。milestone:read 缺失时，/milestones 页面返回 403。权限码（create/update/delete/read）在 permission-codes.md 注册，PM preset role 默认包含全部 4 个，member 默认仅 read
- [ ] 软删除级联：DELETE 里程碑后，其关联的所有 MI 的 milestone_key 在同一事务内置空（NULL），其他视图（列表/甘特图）中这些 MI 不再显示里程碑标签

## Next Steps

- Proceed to `/write-prd` to formalize requirements

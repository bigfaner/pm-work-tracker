---
feature: "里程碑图"
status: Draft
db-schema: "yes"
---

# 里程碑图 — PRD Spec

> PRD Spec: defines WHAT the feature is and why it exists.

## Background

### Why (Reason)

当前 MainItem 是扁平列表，PM 缺乏跨事项的阶段/节点视角。现有 4 个视图（事项列表/甘特图/周报/表格）均为事项维度，无里程碑维度。MainItem 之间无分组机制，PM 只能通过命名约定（如"v1.1-需求阶段"）和记忆管理阶段关系。

单团队 MainItem 数量已达 30–50 个，分 3–4 个交付阶段。每次跨阶段状态盘点需逐条扫描甘特图或列表视图，单个项目每周约需 20–30 分钟人工对齐进度。假设 5 个活跃项目，月浪费约 8 小时 PM 工时。

### What (Target)

新增 Milestone 实体作为团队级别的阶段节点，支持：
- 一个里程碑绑定多个 MainItem（多对一关系）
- 独立的时间线图页面可视化展示里程碑及关联事项
- 四态状态机管理里程碑生命周期
- 完成度自动计算
- 在现有事项页面中集成里程碑维度

### Who (Users)

| 角色 | 使用场景 |
|------|----------|
| PM（项目经理） | 创建/管理里程碑，将事项分配到阶段，按里程碑跟踪进度，进行阶段盘点和汇报 |
| 团队成员 | 查看里程碑时间线，了解当前所处阶段和交付节点 |
| 管理层 | 通过时间线图快速了解项目整体进度和阶段分布 |

## Goals

| Goal | Metric | Notes |
|------|--------|-------|
| 提供里程碑维度的分组跟踪能力 | 支持创建里程碑并关联 MainItem | 核心功能 |
| 减少跨阶段进度盘点时间 | 从 25 分钟/项目降至 5 分钟/项目 | 80% 效率提升 |
| 可视化展示阶段进展 | 时间线图渲染里程碑+关联事项 | 含缩放和拖拽交互 |
| 与现有页面集成 | 3 个现有页面增加里程碑维度 | 事项清单/详情/表格 |

## Scope

### In Scope
- [ ] 里程碑数据模型（新表 pmw_milestones + MainItem 加 milestone_key 外键）
- [ ] CRUD API（创建/读取/更新/删除/列表）
- [ ] 独立 RBAC 权限码（milestone:create/update/delete/read）
- [ ] 时间线可视化页面（/milestones，含缩放和拖拽）
- [ ] 现有页面集成：事项清单增加里程碑筛选、主事项编辑增加里程碑选择器、表格视图增加里程碑列
- [ ] 完成度自动计算（关联 MI 完成度的简单平均值）
- [ ] 四态状态机（not_started/in_progress/completed/cancelled）
- [ ] 软删除 + 事务内解绑关联 MI
- [ ] 数据库迁移（SQLite + MySQL 双 schema）

### Out of Scope
- 里程碑报表/导出（CSV、PDF）
- 通知提醒（里程碑到期/延期通知）
- 甘特图集成（现有甘特图不显示里程碑标记）
- 状态变更历史记录（类似 StatusHistory）
- 里程碑级别的进度记录（类似 ProgressRecord）

## Flow Description

### Business Flow Description

#### 里程碑生命周期

1. **创建**：PM 在时间线页面或通过 API 创建里程碑，填写名称和计划日期，初始状态为 `not_started`
2. **绑定事项**：PM 通过时间线拖拽或在主事项编辑页选择里程碑，将 MainItem 关联到里程碑
3. **启动**：PM 手动将里程碑状态切换为 `in_progress`，或第一个关联 MI 进入 progressing 时提示切换
4. **跟踪**：系统自动计算完成度（关联 MI 完成度平均值），PM 在时间线图查看阶段进展
5. **完成**：PM 手动标记为 `completed`，或所有关联 MI 均为 completed/closed 时可自动完成
6. **取消**：任意状态下 PM 可将里程碑标记为 `cancelled`，关联 MI 自动解绑

#### 里程碑-MainItem 关联流程

1. MainItem 的 milestone_key 为空表示未分配到任何里程碑
2. 绑定：设置 milestone_key 为目标里程碑的 bizKey
3. 换绑：更新 milestone_key 为新里程碑的 bizKey（自动解绑旧里程碑）
4. 解绑：设置 milestone_key 为空

#### 状态机转换规则

| 当前状态 | 可转换到 | 触发条件 |
|----------|----------|----------|
| not_started | in_progress | PM 手动切换 |
| not_started | cancelled | PM 手动取消 |
| in_progress | completed | PM 手动标记 / 所有关联 MI 完成 |
| in_progress | cancelled | PM 手动取消 |
| completed | cancelled | PM 手动取消 |
| cancelled | — | 终态，不可恢复 |

### Business Flow Diagram

```mermaid
flowchart TD
    Start([创建里程碑]) --> Create[填写名称+计划日期]
    Create --> Validate{校验名称 1-100 字符\n+ 计划日期必填?}
    Validate -->|校验失败| ValidationErr[返回字段错误提示]
    Validate -->|校验通过| APICreate[调用创建 API]
    APICreate --> APICheck{API 响应?}
    APICheck -->|500/网络错误| CreateErr[显示"创建失败，请重试"\n保留表单数据]
    APICheck -->|成功| NotStarted[状态: not_started]

    NotStarted --> BindMI[绑定 MainItem]
    BindMI --> BindCheck{MI 与里程碑\n同团队?}
    BindCheck -->|否| BindErr[提示"事项与里程碑\n不属于同一团队"]
    BindCheck -->|是| CalcCompletion[自动计算完成度]

    NotStarted -->|PM 手动| InProgress[状态: in_progress]
    NotStarted -->|PM 取消| Cancelled[状态: cancelled]

    InProgress --> BindMI
    InProgress --> AllDone{所有关联 MI\n均为 completed?}
    AllDone -->|是| Completed[状态: completed]
    AllDone -->|否| InProgress
    InProgress -->|PM 手动标记| Completed
    InProgress -->|PM 取消| Cancelled

    Completed -->|PM 取消| Cancelled

    Cancelled --> UnbindAll[事务内解绑所有 MI]
    UnbindAll --> End([结束])

    CalcCompletion --> InProgress

    style Cancelled fill:#999,color:#fff
    style Completed fill:#4caf50,color:#fff
    style InProgress fill:#2196f3,color:#fff
    style ValidationErr fill:#f44336,color:#fff
    style CreateErr fill:#f44336,color:#fff
    style BindErr fill:#ff9800,color:#fff
```

## Functional Specs

> UI 功能规格详见 [prd-ui-functions.md](./prd-ui-functions.md)。

### Related Changes

| # | Project | Module | Change Point | Updated Logic |
|------|----------|----------|------------|----------------|
| 1 | backend | MainItem model | 新增 milestone_key 字段 | 可空字符串，引用 pmw_milestones.biz_key |
| 2 | backend | MainItem API | 更新/查询接口支持 milestone_key | create/update 接受 milestone_key，list/detail 返回该字段 |
| 3 | backend | RBAC | 新增 4 个权限码 | milestone:create/update/delete/read |
| 4 | backend | database | SQLite + MySQL schema | 新增 pmw_milestones 表，pmw_main_items 加列 |
| 5 | frontend | 事项清单页 /items | 增加里程碑筛选和标签显示 | 新增筛选下拉框和列表中的里程碑标签 |
| 6 | frontend | 主事项详情页 /items/:id | 编辑弹窗增加里程碑选择器 | 新增下拉框字段 |
| 7 | frontend | 表格视图 /table | 增加里程碑列 | 新增表格列 |

## Other Notes

### Performance Requirements
- 时间线页面初始渲染 < 500ms（20 里程碑 + 200 MI）
- 拖拽交互帧率 ≥ 30fps
- 里程碑列表 API 响应 < 200ms
- 超出 200 个 MI 时启用分页加载（按里程碑分组折叠）

### Data Requirements
- 完成度计算：DECIMAL(5,2)，关联 MI completion 的简单平均值（空里程碑为 0）
- 完成度在 GET 时实时计算，不持久化
- 软删除里程碑时，关联 MI 的 milestone_key 在同一事务内置空
- MainItem 的 milestone_key 无 DDL 外键约束，仅加索引

### Monitoring Requirements
- 里程碑 CRUD 操作日志复用现有审计机制

### Security Requirements
- 里程碑权限独立于事项权限，通过 RBAC 控制
- milestone:read 缺失时 /milestones 页面返回 403
- 无权限操作返回 403，前端按钮禁用态

---

## Quality Checklist

- [x] Is the requirement title accurate and descriptive
- [x] Does the background include all three elements: reason, target, users
- [x] Are the goals quantified
- [x] Is the flow description complete
- [x] Does the business flow diagram exist (Mermaid format)
- [x] Is prd-ui-functions.md referenced and UI specs complete
- [x] Are related changes thoroughly analyzed
- [x] Are non-functional requirements considered (performance / data / monitoring / security)
- [x] Are all tables filled completely
- [x] Is there any ambiguous or vague wording
- [x] Is the spec actionable and verifiable

---
feature: "milestone-map"
status: tasks
---

# Feature: milestone-map

<!-- Status flow: prd → design → tasks → in-progress → completed -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 新增 MilestoneMap（五态）+ Milestone（四态）两个层级实体，里程碑图 CRUD、里程碑 CRUD、完成度自动计算、两级视图页面(/milestones)、现有 3 个页面集成里程碑维度 |
| User Stories | prd/prd-user-stories.md | 14 个故事（~100 条 AC）覆盖 PM/团队成员/管理层三个角色的里程碑图 CRUD、里程碑 CRUD、状态切换（含 BR-1~4 校验）、详情面板管理、列表/时间线查看（含空态/错误态/骨架屏）、拖拽换绑、3 个现有页面集成、权限只读控制、确认弹窗/撤销 toast/并发冲突等交互细节 |
| UI Functions | prd/prd-ui-functions.md | 8 个 UI Function：里程碑图两级视图(UF-1)、创建/编辑里程碑弹窗(UF-2)、详情面板(UF-3)、快速添加事项(UF-3a)、事项清单筛选(UF-4)、主事项编辑选择器(UF-5)、表格列(UF-6)、创建/编辑里程碑图弹窗(UF-7) |
| UI Design | ui/ui-design.md | 里程碑图列表组件+时间线视图+详情面板（Badge状态切换+行内解绑+完整表单快速添加）+其他集成组件。基于项目 DESIGN.md 设计系统 |
| Prototype | ui/prototype/ | HTML 原型（milestones.html 含两级视图，items.html 含里程碑筛选，item-detail.html 含选择器，table.html 含里程碑列） |
| Tech Design | design/tech-design.md | 完整技术设计：5层架构、接口签名、状态机、权限码、完成度计算、跨层数据映射、集成规格、测试策略 |
| API Handbook | design/api-handbook.md | 14 个 API 端点（MilestoneMap 7个 + Milestone 7个），含状态码、错误码、权限映射 |
| ER Diagram | design/er-diagram.md | pmw_milestone_maps + pmw_milestones 两张新表 + pmw_main_items ALTER，含索引设计 |
| Schema | design/schema.sql | MySQL 8.0 + SQLite 双方言 DDL |

## Traceability

| PRD Section | Design Section | UI Component | Placement | Tasks |
|-------------|----------------|--------------|-----------|-------|
| UF-1 里程碑图两级视图 | Component: 里程碑图列表视图 + 时间线视图 | MilestoneMapListPage + TimelineView | new-page:/milestones | 1.1, 1.2, 2.1, 2.2, 3.1, 3.2 |
| UF-2 创建/编辑里程碑弹窗 | Component: 创建/编辑里程碑弹窗 | MilestoneDialog | existing-page:/milestones | 2.2, 3.3 |
| UF-3 详情面板 | Component: 里程碑详情面板 | MilestoneDetailPanel | existing-page:/milestones | 2.2, 3.3 |
| UF-3a 快速添加事项 | Component: Quick Add MainItem Dialog | QuickAddDialog（复用 CreateMainItemDialog） | existing-page:/milestones | 2.3, 3.3 |
| UF-4 事项清单筛选 | Component: 事项清单页里程碑筛选 | MilestoneFilter | existing-page:/items | 2.2, 3.4 |
| UF-5 编辑选择器 | Component: 主事项编辑弹窗里程碑选择器 | MilestoneSelector | existing-page:/items/:mainItemId | 2.3, 3.5 |
| UF-6 表格列 | Component: 表格视图里程碑列 | MilestoneColumn | existing-page:/table | 2.3, 3.6 |
| UF-7 创建/编辑里程碑图弹窗 | Component: 创建/编辑里程碑图弹窗 | MilestoneMapDialog | existing-page:/milestones | 2.1, 3.2 |

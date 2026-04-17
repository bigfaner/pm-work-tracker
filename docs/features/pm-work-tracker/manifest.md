---
feature: "pm-work-tracker"
status: design
---

# Feature: pm-work-tracker

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | Web 端 PM 工作事项追踪系统，以主事项→子事项为核心数据结构，叠加周视图、甘特图、事项池和周报导出（Markdown），支持多团队隔离、三级角色权限（超级管理员/PM/成员）、追加式进度记录 |
| User Stories | prd/prd-user-stories.md | 12 个用户故事，覆盖成员提交事项池、PM 拆解分配子事项、成员追加进度记录、PM 周视图对比计划与实际、PM 导出周报、超级管理员管理权限、团队管理、事项池审核、成员自行添加子事项、子事项状态管理（阻塞/挂起/延期）、甘特图视图 |
| UI Functions | prd/prd-ui-functions.md | 11 个 UI 功能模块：登录页、主事项列表（事项视图）、主事项详情页、子事项详情与进度记录、事项池、周视图、甘特图视图、表格视图、周报导出、团队管理页、超级管理员后台 |
| PRD Eval | prd-eval.md | 总评 A；背景与目标 B、流程说明 A、功能描述 B、用户故事 A、范围清晰度 A、UI Functions A |
| Tech Design | design/tech-design.md | React SPA + Go/Gin REST API + GORM（SQLite dev / MySQL prod）；四层架构（Transport/Service/Repository/DB）；6 个核心数据模型；JWT + RBAC + 团队隔离中间件 |
| API Handbook | design/api-handbook.md | 40+ REST 端点，覆盖 Auth/Teams/MainItems/SubItems/ProgressRecords/ItemPool/Views/Reports/Admin；统一响应信封；16 个错误码 |
| UI Design | ui/ui-design.md | React 18 + Ant Design v5；11 个页面完整布局/状态/交互/数据绑定规格；全局筛选栏模式；甘特图自定义 CSS Grid 实现 |

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| 5.1 用户认证与权限 | tech-design: AuthService, JWT Flow; api-handbook: Auth endpoints; ui-design: Page 1 登录页 | — |
| 5.2 团队管理 | tech-design: TeamService; api-handbook: Teams endpoints; ui-design: Page 10 团队管理页, Page 11 超级管理员后台 | — |
| 5.3 主事项管理 | tech-design: MainItemService, MainItem model; api-handbook: MainItems endpoints; ui-design: Page 2 事项视图, Page 3 主事项详情页 | — |
| 5.4 子事项管理 | tech-design: SubItemService, SubItem model, ProgressService; api-handbook: SubItems + ProgressRecords endpoints; ui-design: Page 4 子事项详情页 | — |
| 5.5 事项池 | tech-design: ItemPoolService, ItemPool model; api-handbook: ItemPool endpoints; ui-design: Page 5 事项池 | — |
| 5.6 视图（周/甘特/表格） | tech-design: ViewService; api-handbook: Views endpoints; ui-design: Page 6 周视图, Page 7 甘特图, Page 8 表格视图 | — |
| 5.7 周报导出 | tech-design: ReportService; api-handbook: Reports endpoints; ui-design: Page 9 周报导出 | — |

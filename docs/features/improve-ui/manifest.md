---
feature: "improve-ui"
status: tasks
---

# Feature: improve-ui

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Consolidated Specs

| Spec File | Items | Target |
|-----------|-------|--------|
| specs/biz-specs.md | BIZ-001–008 (8 rules) | docs/conventions/ |
| specs/tech-specs.md | TECH-001–005 (5 specs) | docs/conventions/ |

Integrated: 2026-05-04

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 全面 UI 重做需求：13 页面视觉风格替换 + 5 页面交互结构变更 + 设计系统统一化 + 后端 API 适配 |
| User Stories | prd/prd-user-stories.md | 6 个用户故事：视图切换、用户管理独立页、团队详情独立页、每周进展对比、全量表格独立页、设计系统一致性 |
| UI Functions | prd/prd-ui-functions.md | 13 个 UI 功能模块详细规格，含 4 个结构变更页面的完整交互说明 |
| Tech Design | design/tech-design.md | shadcn/ui + Radix UI + Tailwind CSS 大爆炸替换；后端 User 模型扩展 + Admin CRUD + 每周进展 API 重构 |
| API Handbook | design/api-handbook.md | 4 个新增端点 + 2 个修改端点（用户 CRUD、每周进展对比） |
| Prototype | ui/prototype/ | 15 个原型文件（13 HTML + app.js + styles.css），作为精确实现蓝图 |
| Proposal | ../../proposals/improve-ui/proposal.md | 原始提案：原型驱动替换方案，前后端联动 |

## Traceability

<!-- Maps PRD sections to prototype pages and expected design/task artifacts. -->

| PRD Section | Prototype Page | Change Level | Tasks |
|-------------|----------------|--------------|-------|
| 5.1 用户管理 | user-management.html | 结构变更（从超级管理员 Tab 拆分） | 5.4 |
| 5.2 团队详情 | team-detail.html | 结构变更（从内嵌视图变为独立路由） | 5.5 |
| 5.3 事项清单 | main-items.html | 结构变更（新增 Summary/Detail 切换） | 5.1 |
| 5.4 全量表格 | table-view.html | 新增独立页 | 5.2 |
| 5.5 每周进展 | weekly-view.html | 重点变更（双列对比 + 进度增量标记） | 3.3, 5.3 |
| 5.6 设计系统 | styles.css | 全局统一化 | 2.1, 2.2 |
| 5.7 后端适配 | — | API 增量调整 | 1.2, 3.1, 3.2 |
| 登录 | login.html | 视觉重做 | 4.1 |
| 主事项详情 | mainitem-detail.html | 视觉重做 | 4.2 |
| 子事项详情 | subitem-detail.html | 视觉重做 | 4.2 |
| 事项池 | item-pool.html | 视觉重做 | 4.3 |
| 甘特图 | gantt-view.html | 视觉重做 | 4.4 |
| 周报导出 | weekly-report.html | 视觉重做 | 4.1 |
| 团队管理 | team-management.html | 视觉重做 | 4.5 |
| — | — | 工具链迁移 + 测试 | 1.1, 6.1 |

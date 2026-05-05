---
feature: "code-conventions"
status: tasks
---

# Feature: code-conventions

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 建立项目级编码规范 + 清理现有代码，确保 AI 和开发者遵循一致风格 |
| User Stories | prd/prd-user-stories.md | 5 个用户故事：规范加载、命名统一、去重、组件化、TDD 流程 |
| Tech Design | design/tech-design.md | 3-phase plan: docs/lint → backend cleanup → frontend cleanup |
| API Handbook | design/api-handbook.md | Internal interfaces: MapNotFound, ApplyPaginationDefaults, ParseDate, Textarea, PrioritySelect |

## Consolidated Specs

| Spec File | Items | Target |
|-----------|-------|--------|
| specs/biz-specs.md | BIZ-001–003 (3 rules) | docs/conventions/ |
| specs/tech-specs.md | TECH-001–008 (8 specs) | docs/conventions/ |

Integrated: 2026-05-04

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| 5.1 规范文档体系 | Architecture, Section 5 | 1.1, 1.2 |
| 5.5 Lint 规则配置 | Section 6 Lint Configuration | 1.3 |
| 5.3 后端去重 | Section 1 Backend Shared Helpers, Section 2 Constructor | 2.1, 2.2, 2.3 |
| 5.4 前端组件化 | Section 3 Frontend Components | 3.1, 3.2 |
| 5.4 UI 风格统一 | Section 4 Color Token Mapping | 3.3 |
| 5.6 开发流程规范 | Section 5 rules/testing.md | 1.2 (testing.md) |

---
feature: "integration-test-coverage"
status: tasks
---

# Feature: integration-test-coverage

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 为 36 个未测试 API 端点编写按用户流程组织的集成测试，补全 6 个单元测试缺口，目标 100% 端点覆盖率、150+ 测试用例 |
| User Stories | prd/prd-user-stories.md | 8 个用户故事：Item 生命周期、Item Pool 审查、团队管理、后台用户管理、视图报表、共享辅助函数提取、代码审查者、单元测试缺口 |
| Tech Design | design/tech-design.md | 5 个集成测试文件 + helpers.go 提取 + 6 个单元测试缺口，使用现有内存 SQLite + DI 路由模式，无新生产代码 |
| Tasks | tasks/index.json | 21 tasks across 5 phases: F1 lifecycle → F7 helpers → F2/F3/F4 core flows → F5 views → F6 unit gaps |

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| F1: Item Lifecycle (17 endpoints) | Test Request Pattern, Status Machine Reference, Helper Signatures | 1.1, 1.2 |
| F2: Item Pool (6 endpoints) | Helper Signatures, DB Isolation, Test Naming | 3.1 |
| F3: Team Management (9 endpoints) | Helper Signatures, DB Isolation, Test Naming | 3.2 |
| F4: Admin User Management (6 endpoints) | Helper Signatures, DB Isolation, Test Naming | 3.3 |
| F5: Views & Reports (6 endpoints) | Helper Signatures, DB Isolation, Test Naming | 4.1 |
| F6: Unit Test Gaps (6 methods) | F6 Unit Test Gap Strategy | 5.1, 5.2 |
| F7: Shared Helpers (10 functions) | Unified Setup Function, Helper Signatures | 2.1 |

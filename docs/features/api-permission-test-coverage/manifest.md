---
feature: "api-permission-test-coverage"
status: tasks
---

# Feature: api-permission-test-coverage

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 为 12 个权限敏感端点补充 Handler 单元测试（24 case），并新增集成测试覆盖预设角色矩阵、自定义角色权限组合、权限边界场景，预计 ~3 个工作日 |
| User Stories | prd/prd-user-stories.md | 4 个用户故事：权限差异端点单元测试、预设角色矩阵集成验证、自定义角色权限组合验证、权限边界场景验证 |
| Tech Design | design/tech-design.md | 3 个测试文件：handler/permission_matrix_test.go（U1，12端点×2 case，无DB）、tests/integration/rbac_permission_test.go（I-A/B/C，真实SQLite）、tests/integration/permission_coverage_test.go（I-D，覆盖率断言）；无新生产代码，无新CI步骤 |

## Traceability

| PRD Section | Design Section | UI Component | Tasks |
|-------------|----------------|--------------|-------|
| U1: Handler 单元测试 | Tech Design § U1 — `permission_matrix_test.go` | — | 1.1, 1.2 |
| I-A: 预设角色矩阵 | Tech Design § I-A — `TestRBACPermMatrix_PresetRoles` | — | 2.1 |
| I-B: 自定义角色 | Tech Design § I-B — `TestCustomRole_PartialPermissions` | — | 2.2 |
| I-C: 权限边界 | Tech Design § I-C — `TestPermBoundary_*` | — | 2.3 |
| I-D: 权限码覆盖率 | Tech Design § I-D — `TestPermissionCodeCoverage` | — | 2.4 |

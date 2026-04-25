---
status: "completed"
started: "2026-04-25 21:45"
completed: "2026-04-25 21:48"
time_spent: "~3m"
---

# Task Record: fix-e2e-3-1 修复 e2e 测试失败: unknown

## Summary
修复 001_init_test.go 中 Test001Init_CreatesAllTables 测试失败：测试期望 pmw_ 前缀表名，但 001_init.sql 创建的是无前缀表名。将测试中的期望表名从 pmw_users/pmw_teams 等改为 users/teams 等，与实际 SQL 保持一致。

## Changes

### Files Created
无

### Files Modified
- backend/migrations/001_init_test.go

### Key Decisions
- 001_init.sql 使用无前缀表名（SQLite 迁移历史），schema.sql 使用 pmw_ 前缀（MySQL 目标 schema）；测试应与 001_init.sql 保持一致

## Test Results
- **Passed**: 19
- **Failed**: 0
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] 已定位失败的根本原因
- [x] 已修复代码或测试脚本
- [x] 单元测试全部通过

## Notes
无

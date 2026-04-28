---
id: "1.gate"
title: "Phase 1 Exit Gate"
priority: "P0"
estimated_time: "1h"
dependencies: ["1.summary"]
status: pending
breaking: true
---

# 1.gate: Phase 1 Exit Gate

## Description

Exit verification gate for Phase 1 (后端权限码变更). Confirms that all backend outputs are complete, internally consistent, and match the design specification before Phase 2 (frontend) begins.

## Verification Checklist

1. [ ] `permissions/codes.go` 包含新 `user` 资源（4 个码）和新 `role` 资源（4 个码），`user:manage_role` 不存在
2. [ ] `router.go` 中 14 条路由全部绑定新权限码，`user:manage_role` 零残留
3. [ ] `MigratePermissionGranularity` 函数存在，幂等测试通过
4. [ ] `seedPresetRoles` pm 角色权限码列表已更新
5. [ ] `go build ./backend/...` 通过
6. [ ] `go test ./backend/internal/pkg/permissions/ ./backend/internal/migration/ ./backend/internal/handler/` 全部通过
7. [ ] 新增 7 个路由中间件测试用例全部通过
8. [ ] `grep -r "user:manage_role" backend/` 零结果

## Reference Files

- `docs/features/permission-granularity/design/tech-design.md` — Cross-Layer Data Map
- `docs/features/permission-granularity/tasks/records/1-summary.md`

## Acceptance Criteria

- [ ] 所有 Verification Checklist 项通过
- [ ] 任何偏离设计的决策已记录在 record 中
- [ ] Record 通过 `task record` 创建，含测试证据

## Implementation Notes

This is a verification-only task. No new feature code should be written.
If issues are found:
1. Fix inline if trivial
2. Document non-trivial issues as decisions in the record
3. Set status to `blocked` if a blocking issue cannot be resolved

---
id: "2.gate"
title: "Phase 2 Exit Gate"
priority: "P0"
estimated_time: "1h"
dependencies: ["2.summary"]
status: pending
breaking: true
---

# 2.gate: Phase 2 Exit Gate

## Description

Exit verification gate for Phase 2 (前端权限守卫更新). Confirms that all frontend outputs are complete, internally consistent, and match the design specification before test generation begins.

## Verification Checklist

1. [ ] `lib/permissions.ts` 包含新 `user` 组（4 个码）和新 `role` 组（4 个码），`user:manage_role` 不存在
2. [ ] `App.tsx` 路由守卫使用 `user:list`（用户管理）和 `role:read`（角色管理）
3. [ ] `Sidebar.tsx` 菜单项使用 `user:list` 和 `role:read`
4. [ ] `TeamManagementPage.tsx` 角色下拉 `useQuery` 添加 `enabled: canReadRoles`
5. [ ] `RoleManagementPage.tsx` 创建/编辑/删除按钮分别由 `role:create/update/delete` 独立控制
6. [ ] `permission-driven-ui.test.tsx` 旧码引用全部替换，新增守卫测试通过
7. [ ] `npm test` 全部通过
8. [ ] `grep -r "user:manage_role" frontend/src/` 零结果

## Reference Files

- `docs/features/permission-granularity/design/tech-design.md` — Cross-Layer Data Map
- `docs/features/permission-granularity/tasks/records/2-summary.md`

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

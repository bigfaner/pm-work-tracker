---
iteration: 0
title: "Pre-Revision (Freeform Findings)"
---

# Pre-Revision Eval Report (Iteration 0)

## ATTACK_POINTS

### Factual Corrections (direct edits)

- **[high]** 权限 bug 根因被误判为不明，实际代码分析已可定位具体原因 | quote: "#8 权限 bug 根因不明（可能是 seed 数据或中间件逻辑）" | improvement: Replace ambiguous root-cause statement with specific diagnosis: RoleKey is nil on team member records. Commit to investigation step (query pmw_team_members for role_key IS NULL) before fix.

- **[high]** 排除了 seed 数据更新但删除功能需要新权限码，构成直接矛盾 | quote: "新增权限码的 seed 数据更新（由 RBAC 提案覆盖）" out of scope, but "删除需新增权限码（main_item:delete, sub_item:delete）" in Constraints | improvement: Either move seed data update for delete permission codes into Phase 1 scope (4-line change in migration/rbac.go), or defer delete to Phase 2 and use existing permissions as temporary guards.

### Structural/Architectural Suggestions

- **[high]** 过滤穿透的反向查询模式缺乏实现策略 | quote: "过滤穿透：负责人筛选时，如果某子事项匹配，则连带展示其主事项。这需要后端支持子→父反向查询，卡片视图从客户端过滤改为服务端过滤" | improvement: Specify implementation as enhancement to existing in-memory filtering in ViewService.TableView, not a new SQL query pattern. Describe the algorithm: collect matching sub-items → extract parent keys → include parents where own assignee matches OR biz_key in parent set.

- **[high]** 级联软删除未指定事务边界 | quote: "PM 点击删除按钮 → 确认对话框 → 软删除主事项及其所有子事项" | improvement: Add explicit constraint: cascade delete executes in single DB transaction, soft-delete sub-items first (bulk), then main item. Skip linkage evaluation for deleted items.

- **[medium]** 子事项移动后旧父项下剩余子项是否重编号未明确 | quote: "子事项移动到其他主事项时自动重新编号，保持编号与父事项的一致性" | improvement: State explicitly: moved sub-item gets new code via NextSubCode on target parent, remaining sub-items on source parent are NOT renumbered (gap acceptable, avoids cascading code changes).

- **[medium]** Phase 2 的移动对话框依赖 Phase 1 的表单清空模式 | quote: "所有新增/转换表单关闭时清空字段" | improvement: Acknowledge the cross-phase dependency. Either extract shared form-reset utility in Phase 1, or note that Phase 2 must independently implement reset for move dialog.

- **[medium]** 软删除主事项后无撤销路径 | quote: "PM 点击删除按钮 → 确认对话框 → 软删除主事项及其所有子事项" | improvement: Add minimal traceability by recording deletion event in status_history (one extra call within delete transaction). No new table or UI needed. Update Out of Scope to clarify that status history recording IS included but dedicated audit UI is not.

### Subjective Preferences (skipped — not actionable for pre-revision)

- (建议 items are prescriptive improvements, not findings to edit into the document — the structural suggestions above incorporate their substance where the concern identifies a verifiable gap)

## BORDERLINE_FINDINGS

- Sub-item move renumbering edge case: partially accepted (the gap-rename question is a verifiable ambiguity in the proposal, but the detailed NextSubCode analysis is implementation-level detail beyond proposal scope)

## SKIPPED_FINDINGS

- 建议 items that duplicate structural suggestions: investigation step, in-memory filter approach, seed data resolution, transaction constraint, renumbering behavior, status history recording — their substance is incorporated into the structural suggestions above

## Classification Audit

- Factual corrections: 2 (verifiable errors in original text)
- Structural suggestions: 6 (verifiable gaps/ambiguities in proposal structure)
- Subjective preferences: 5 (implementation suggestions from reviewer, substance absorbed where gap exists)

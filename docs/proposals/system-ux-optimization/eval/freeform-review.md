# Freeform Review: System UX Optimization Batch

**Reviewer:** Senior Full-Stack Engineer (Go/Gin/GORM + React/TypeScript, PM SaaS domain)
**Date:** 2026-06-02
**Document:** `docs/proposals/system-ux-optimization/proposal.md`

---

## Section 1: Background Assessment

This proposal addresses 10 UX and functional defects in PM Work Tracker, grouped into two delivery phases. Phase 1 covers eight bug fixes and minor enhancements: improving status-transition error messages, exposing sub-item start-date editing, adding soft-delete for main/sub items, form validation and reset, sub-item list sorting, and fixing a critical member-role permission bug. Phase 2 introduces two new features: sub-item cross-parent move (with renumbering) and filter penetration (multi-status filter + assignee-based child-to-parent inclusion in card and table views).

The codebase uses a layered Go backend (Router -> Handler -> Service -> Repository -> Model) with GORM on SQLite/MySQL, and a React/TypeScript frontend with Zustand stores, TanStack Query, and Radix UI primitives. RBAC is seeded through a migration system (`migration/rbac.go`) that creates three preset roles (superadmin, pm, member) with permission codes. Team-scoped middleware (`TeamScopeMiddleware`) resolves the member's `role_key` to permission codes and injects them into the Gin context. The card view is the Gantt view (`GanttViewPage.tsx`), which uses server-side data via `getGanttViewApi`. The table view (`TableViewPage.tsx`) uses `ViewService.TableView`, which fetches all items in-memory and filters client-side within Go service code.

The proposal correctly identifies that sub-item soft-delete exists at the service layer (`SubItemService.Delete`) but lacks an HTTP endpoint, and that main-item delete is entirely missing. It also correctly notes that the card view's assignee filter does not penetrate into sub-items, and that the table view's filtering is done server-side in Go (not truly "client-side" in the browser), but the filtering logic in `ViewService.TableView` (`matchesFilterMain`/`matchesFilterSub`) does not support parent-child penetration.

---

## Section 2: Key Risks

### Risk 1: Permission Bug Root Cause Is Misidentified

**风险：** The proposal treats #8 as a bug whose root cause is ambiguous, but the code analysis reveals the exact root cause.

The proposal states:

> "#8 权限 bug 根因不明（可能是 seed 数据或中间件逻辑）"

and in Key Risks:

> "先定位根因（检查 member 角色的 RoleKey 是否为 nil）再修复"

The existing test `TestTeamScopeMiddleware_MemberNoRoleID_SetsEmptyPermCodes` in `team_scope_test.go` explicitly proves the code path: when `member.RoleKey` is `nil`, `TeamScopeMiddleware` skips the role lookup entirely and sets `permCodes` to `nil` (not an empty slice -- `nil`). This means `GetPermCodes` returns `nil`, and the `RequirePermission` middleware falls through to the non-team DB query path, which may also fail if the user's role assignment is broken. The root cause is that `TeamScopeMiddleware` line 79 sets `permCodes` even when it is `nil`, rather than defaulting to an empty slice or returning an error.

This matters because the proposed mitigation ("check member role's RoleKey for nil") is close but not precise. The real question is: **why is `RoleKey` nil in production?** Is it a migration issue where `rebuildTeamMembersTable` did not correctly populate `role_key` for existing members? Or is it that newly invited members are created without `role_key`? The proposal should commit to a specific diagnosis step: query the `pmw_team_members` table for records where `role_key IS NULL`, cross-reference with the migration history, and determine whether this is a one-time data fix or an ongoing code path defect in member invitation.

### Risk 2: Filter Penetration Query Strategy Is Underspecified

**风险：** The proposal introduces a fundamentally new query pattern (child-to-parent reverse lookup) without specifying the implementation strategy or performance characteristics.

The proposal states:

> "过滤穿透：负责人筛选时，如果某子事项匹配，则连带展示其主事项。这需要后端支持子→父反向查询，卡片视图从客户端过滤改为服务端过滤"

and in Non-Functional Requirements:

> "过滤穿透查询性能：主事项 + 子事项联合过滤应在 500ms 内返回"

Currently, `ViewService.TableView` fetches **all** non-archived main items and **all** sub-items for the team in two queries, then filters in-memory. For assignee penetration, the logic would need to: (1) find all sub-items where `assignee_key = X`, (2) collect the parent `main_item_key` values, (3) include those parent main items in the result even if the main item's own `assignee_key` does not match. This is doable in the current in-memory model but raises concerns:

- **Data volume:** For a team with 500 main items and 2000 sub-items, the current approach already loads all into memory. Adding penetration does not change this, but the proposal mentions changing the card view to "server-side filtering." If "server-side" means SQL-level pushdown, this is a significant architectural change from the current pattern. If it means the existing Go in-memory filtering with additional logic, that should be stated explicitly.
- **The Gantt view** (`ViewService.GanttView`) also uses the fetch-all-then-filter-in-memory pattern for status filtering. The proposal says "卡片视图和表格视图统一支持" but does not address whether the Gantt view's `GanttFilter` needs the same penetration, or whether penetration only applies to the table view's `TableFilter`.

The ambiguity matters because the implementation approach determines scope, testing strategy, and whether new repository methods or service-layer changes are needed.

### Risk 3: Cascading Delete Lacks Transaction Boundary Specification

**风险：** The proposal describes cascading soft-delete of main items with sub-items but does not address transaction boundaries.

The proposal states:

> "PM 点击删除按钮 → 确认对话框 → 软删除主事项及其所有子事项"

Currently, `MainItemService` has no `Delete` method at all. The `MainItemRepo` interface has no `SoftDelete` method. Adding cascade delete requires:

1. A new `SoftDelete` method on `MainItemRepo`
2. A new `Delete` method on `MainItemService` that soft-deletes the main item and then soft-deletes all its sub-items
3. This must happen in a single database transaction to avoid partial deletion (main item deleted but some sub-items orphaned)

The existing `SubItemService.Delete` performs a single `SoftDelete` and then calls `EvaluateLinkage`. The cascade version would need to: soft-delete all sub-items, then soft-delete the main item, all in one transaction. The linkage evaluation step after deletion is moot (the item is deleted), so it should be skipped or the proposal should clarify this.

Furthermore, the proposal excludes "新增权限码的 seed 数据更新（由 RBAC 提案覆盖）" from scope, but simultaneously requires "删除需新增权限码（main_item:delete, sub_item:delete)" in Constraints & Dependencies. This is a direct contradiction: if the permission codes are needed for the delete endpoints to work, but seed data update is out of scope, then the delete endpoints will require permissions that no role has. Either the seed data must be updated (making it in-scope) or the permission codes must not be added yet (making delete unguarded).

### Risk 4: Sub-Item Move Renumbering Has Hidden Edge Cases

**风险：** The proposal describes sub-item move with renumbering but underestimates the complexity of code management.

The proposal states:

> "子事项移动到其他主事项时自动重新编号，保持编号与父事项的一致性"

and in Key Risks:

> "#9 子事项移动时编号重生成可能与已有编号冲突 -- 使用现有编号服务的序列逻辑，在事务中执行"

The current numbering system (`NextSubCode`) uses a `sub_item_seq` counter on the main item and a `MAX(code)` query to handle concurrent inserts. When moving a sub-item to a new parent:

1. The sub-item's code (e.g., "TEAM-00001-03") must be changed to match the new parent's code prefix (e.g., "TEAM-00002-07").
2. But should the sub-items remaining under the old parent be renumbered to close the gap? The proposal does not address this.
3. If renumbering happens, it changes the codes that users may have referenced in external communications.
4. The `NextSubCode` function increments `sub_item_seq` atomically. If we use it during move, it increments the sequence counter even though this is a move, not a creation. This is acceptable but should be documented as a design choice.

The proposal also does not address what happens to the sub-item's history records. Status history records reference the sub-item by `biz_key`, so they will follow the move naturally. But progress records also reference `sub_item_key` by biz_key, so they will follow as well. The proposal should confirm this is the intended behavior.

### Risk 5: Phase Boundary Hides a Cross-Dependency

**问题：** Phase 1 includes form clear-on-close (#6) but Phase 2 introduces sub-item move (#9), which will require a new dialog/form. This creates an implicit Phase 2 dependency on the Phase 1 pattern.

The proposal states for Phase 1:

> "所有新增/转换表单关闭时清空字段"

This establishes a project-wide convention that all forms must reset on close. Phase 2's sub-item move will introduce a new "Move Sub-Item" dialog. If the Phase 1 convention is implemented as a shared utility or hook (e.g., a `useFormReset` hook), Phase 2 naturally benefits. But if Phase 1 implements the reset ad-hoc per form (which is the simpler approach and more likely given the "1-2 days" timeline), then Phase 2 must remember to apply the same pattern. The proposal should either: (a) extract a shared form-reset utility in Phase 1 that Phase 2 reuses, or (b) acknowledge that Phase 2 must independently implement the reset for the move dialog.

### Risk 6: Excluded Scope Items Are Architectural Prerequisites

**风险：** The proposal excludes "新增权限码的 seed 数据更新" and "删除通知/审计日志" but these are not optional nice-to-haves for certain features.

For delete specifically, the proposal states in Out of Scope:

> "删除通知/审计日志"
> "新增权限码的 seed 数据更新（由 RBAC 提案覆盖）"

And in Constraints:

> "删除需新增权限码（main_item:delete, sub_item:delete）"

The seed data exclusion is contradictory as noted in Risk 3. But beyond that, soft-deleting a main item is a destructive operation that currently has **no undo path**. The existing `Archive` method requires terminal status and sets `archived_at`. Soft-delete sets `deleted_flag = 1`, which causes all `NotDeleted()` scoped queries to exclude it. There is no "restore" endpoint and no way for users to recover from accidental deletion. The confirmation dialog ("将同时删除 N 个子事项") mitigates this, but in a PM tool where items represent committed work, the lack of any recovery path or audit trail is a significant gap.

The proposal should either include a minimal audit trail (e.g., status history record for deletion) or explicitly acknowledge that deleted items are recoverable only through direct database manipulation, and get explicit stakeholder sign-off on this.

---

## Section 3: Improvement Suggestions

### Suggestion 1: Pin Down the Permission Bug Root Cause

**建议：** Before writing the PRD, add a dedicated investigation step that queries the database for `pmw_team_members` records where `role_key IS NULL` and cross-references with the member invitation code path in `TeamService`.

This addresses Risk 1. The investigation should answer two questions: (a) Is this a one-time migration artifact (existing members from before RBAC migration), in which case a data fix script suffices? (b) Or is the `AddMember` code path in the team service not setting `role_key` for newly invited members, which would be an ongoing code bug?

After adopting this, the proposal would include a concrete investigation deliverable in Phase 1 (before the fix), such as: "Step 0: Query pmw_team_members for role_key IS NULL, determine root cause, document findings. Step 1: Apply fix (data migration patch or code fix in AddMember). Step 2: Add integration test that creates a member user, logs in, and verifies permCodes is non-nil."

### Suggestion 2: Specify the Filter Penetration Implementation as In-Memory Enhancement

**建议：** Commit to implementing filter penetration as an enhancement to the existing in-memory filtering in `ViewService.TableView`, not as a new SQL-level query pattern. Add a clear description of the algorithm.

This addresses Risk 2. The algorithm would be:

1. Fetch all non-archived main items and sub-items (existing pattern).
2. When `assigneeKey` filter is set, first collect all matching sub-items.
3. Extract their `main_item_key` values into a set.
4. Include main items where `assignee_key` matches OR `biz_key` is in the parent set.
5. For included main items that are included only due to sub-item match, also include the matching sub-items in the result (not all sub-items of that parent).

This keeps the change contained within `matchesFilterMain`/`matchesFilterSub` and avoids introducing new repository methods. The proposal should also clarify that the Gantt view uses `GanttFilter` which only has a single `Status` string (not multi-select), and that Gantt filter penetration is explicitly Phase 2+ if needed.

After adopting this, the proposal's Phase 2 description would read: "Enhance ViewService.TableView's existing in-memory filtering to support parent inclusion when child matches assignee filter. Add multi-value status filter support to TableFilter (already supported by `Status []string`). Gantt view remains single-status filter in this phase."

### Suggestion 3: Resolve the Permission Code Seed Data Contradiction

**建议：** Either move the seed data update for `main_item:delete` and `sub_item:delete` into scope, or defer delete endpoints to Phase 2 and use existing permissions (`main_item:archive` for main-item delete, `sub_item:update` for sub-item delete) as temporary guards.

This addresses Risk 3's contradiction. The simplest resolution is to add the two new permission codes to the `seedPresetRoles` function's `pmCodes` list and update `VerifyPresetRoleCodes` accordingly, which is a 4-line change. This is too small to justify deferring to a separate "RBAC proposal." If the concern is that adding codes to the seed function triggers a broader RBAC review, then the proposal should note this dependency explicitly and not list delete as a Phase 1 deliverable.

After adopting this, the proposal would include in Phase 1 scope: "Add `main_item:delete` and `sub_item:delete` permission codes to `migration/rbac.go` seedPresetRoles and assign to pm role. Update `permissions` registry."

### Suggestion 4: Require Transaction for Cascade Delete

**建议：** Add an explicit technical constraint: "Main item cascade delete must execute within a single database transaction, soft-deleting all sub-items first, then the main item. Linkage evaluation must be skipped for deleted items."

This addresses Risk 3's transaction boundary concern. The implementation would add a `Delete` method to `MainItemService` that:

1. Begins a transaction.
2. Soft-deletes all sub-items under the main item (bulk update `deleted_flag = 1`).
3. Soft-deletes the main item.
4. Commits the transaction.

No `EvaluateLinkage` call is needed because the item is being removed entirely. The `SubItemRepo` already has `SoftDelete`, but a bulk version (`SoftDeleteByMainItem`) would be more efficient than calling per sub-item.

After adopting this, the proposal's Constraints section would include: "Cascade delete runs in a single DB transaction. Sub-items are soft-deleted first (bulk), then main item. No linkage evaluation post-delete."

### Suggestion 5: Specify Renumbering Behavior on Move

**建议：** State explicitly whether remaining sub-items under the old parent are renumbered to close the gap, and whether the moved sub-item's code is assigned using `NextSubCode` on the new parent.

This addresses Risk 4. The recommended approach is:

- Moved sub-item gets a new code via `NextSubCode` on the target parent (append at end).
- Remaining sub-items on the source parent are **not** renumbered (gap is acceptable, avoids cascading code changes that confuse users who have referenced codes externally).

This keeps the move operation simple and predictable. If renumbering is desired later, it can be a separate "renumber" action.

After adopting this, the proposal would read: "子事项移动到新主事项时，使用 NextSubCode 在目标主事项下生成新编号。源主事项下的剩余子事项不重新编号（保留间隔）。"

### Suggestion 6: Add Minimal Deletion Audit via Status History

**建议：** Record a status history entry when a main item or sub-item is soft-deleted, using the existing `RecordStatusChange` helper with `is_auto = 1` and a remark like "soft-deleted by user X."

This addresses Risk 6. It is a lightweight addition (one extra call within the delete transaction) that provides recoverability information without requiring a full audit log system. Admins can query the status history to see who deleted what and when. This does not require a new table or UI.

After adopting this, the proposal's Out of Scope section would still exclude "删除通知/审计日志" (no notification or dedicated audit UI), but the in-scope delete feature would include: "Record deletion event in status_history for traceability."

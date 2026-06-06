# Design Eval — Iteration 2

**Score**: 910/1000

## Previous Issues Addressed

### Attack 1: Missing data flow arrows in component diagram
**Partially fixed.** The component diagram still uses the same hierarchical listing format without directional arrows. However, the transaction sequences in Interfaces 1-3 (steps 1-6 for delete, 1-6 for move) provide the equivalent traceability for key flows. The missing arrows in the overview diagram remain a minor gap but the transaction sequences compensate.

### Attack 2: schema.sql not executable
**Fixed.** The schema.sql now explicitly states: "Note: Actual seed insertion is handled by Go migration code (SyncPresetRoles with INSERT-IGNORE semantics). The SQL below documents the expected seed data for reference only." The file is intentionally documentation-only and this is now clearly documented.

### Attack 3: StatusHistoryRepo.ListByTeamInRange lacks interface contract
**Partially fixed.** Interface 8 now names `ListByItemKeysInRange` instead of `ListByTeamInRange` and provides a typed signature: `ListByItemKeysInRange(ctx context.Context, itemType string, itemKeys []int64, start, end time.Time) ([]model.StatusHistory, error)`. However, the `filterInactiveTerminal` function in Interface 8 still references the old method name `ListByTeamInRange` in its explanatory text — there is a naming inconsistency between the typed signature and the prose.

### Attack 4: handleError(c, err) does not exist
**Fixed.** The document now correctly uses `apperrors.RespondError(c, err)` in the propagation strategy section: "Handler 层统一通过 `apperrors.RespondError(c, err)` 映射到 HTTP 响应（现有模式）。"

### Attack 5: apperrors.New() is not the actual constructor
**Fixed.** The error definitions now use struct literal construction: `&apperrors.AppError{Code: "BAD_REQUEST", Message: "...", Status: 400}` which matches the codebase's `AppError` struct pattern.

### Attack 6: Repository layer changes not in test plan
**Fixed.** The per-layer test plan now includes a Repository row: "Backend Repository | Unit | Go testing + testify | SubItemRepo ORDER BY, TeamRepo membership filter, StatusHistoryRepo range query | 80%". Key test scenarios section also lists repository-level tests.

### Attack 7: Delete operation missing status_history recording
**Fixed.** Interface 1 transaction sequence step 5 now explicitly states: "为每个已删除事项插入 `status_histories`（`from_status=当前状态, to_status="deleted"`），仅作为审计记录，不修改 `item_status`". Interface 2 step 3 also includes the status_history insert. The note below each interface reinforces this.

### Attack 8: Interface 10 bundles 7 distinct frontend items
**Not fixed.** The frontend items (#1, #2, #4, #6, #7, #13, #14) remain grouped under a single Interface 11 section (renumbered from Interface 10). The PRD Coverage Map correctly references them individually, but the interface section still bundles all 7 items together without explicit task boundaries.

### Attack 9: Cross-team target validation gap in Move
**Fixed.** Interface 3 step 2 now explicitly states: "`FindByBizKey` 获取目标主事项（校验未删除、`team_key == teamBizKey`、非终态、非同一主事项）". The Security section mitigation also now says "Service 层显式校验目标主事项 `team_key == 当前 teamBizKey`，防止跨团队移动" and clarifies "FindByBizKey 仅验证记录存在性，不校验团队归属".

### Attack 10: AssigneeKey type mismatch
**Not fixed.** The Cross-Layer Data Map still shows `*int64` at the Backend Model level for assigneeKey while the Go struct defines `AssigneeKey *string`. The note says "query param 为 string，Service 层用 `pkg.ParseID` 转为 int64" which partially addresses the conversion but the map's "Backend Model" column still shows `*int64` which is misleading since the filter struct uses `*string`.

### Attack 11 (blindspot): PRD Coverage Map mis-indexes Interface 9
**Not fixed.** The PRD Coverage Map row for #8 says "Interface 9 (后端中间件)" but Interface 9 in the document is actually "团队选择器权限过滤 (#15)". The member permission fix (#8) has no dedicated Interface section — it is described only in the Architecture section under Component Diagram > Middleware and as Interface 10. A developer looking up #8 in the coverage map is directed to the wrong interface.

### Attack 12 (blindspot): Weekly progress filter "活跃定义" ambiguity
**Partially fixed.** Interface 8 now provides more detail with `(a) status_history 中存在该主事项的状态变更记录（create_time 在时间范围内）` and `(b) 该主事项下存在 create_time 或 db_update_time 在时间范围内的子事项` and `(c) 该主事项的 db_update_time 在时间范围内且变更类型为进度更新`. However, the PRD User Story #11 defines "活跃" differently from the design's condition (c): the PRD says "该主事项的 updated_at 在时间范围内且变更类型为进度更新" but the design uses "db_update_time" which is not a real column name. The actual column is `updated_at`. This is a minor but real naming mismatch.

## Dimension Scores

### 1. Architecture Clarity: 158/170

**Layer placement explicit (55/60)**:
All 16 items are categorized by architecture layer in the overview table. The layer placement section states: "所有 16 项变更位于现有四层架构（Router→Handler→Service→Repository→Model）内。无新层。" The component diagram maps each component to its layer.

Deduction (-5): The overview table labels #8 as "Backend middleware + API" but the actual change is only middleware (no API change). This mislabel persists from iteration 1.

**Component diagram present (58/60)**:
The ASCII component diagram is detailed and maps all 16 items to their respective frontend and backend components. The transaction sequences in Interfaces 1-3 compensate for the lack of data flow arrows by providing step-by-step operation chains.

Deduction (-2): No directional arrows between layers remain, but the transaction sequences in Interface definitions provide sufficient flow detail for implementers.

**Dependencies listed (45/50)**:
Three dependencies are listed with type, purpose, and version. The note "仅新增 1 个前端依赖" is clear.

Deduction (-5): The new `StatusHistoryRepo.ListByItemKeysInRange` method is a dependency of Interface 8 but is not listed in the dependencies table. While it is a code-level dependency rather than a package dependency, its absence from any tracking list means it could be overlooked during implementation.

### 2. Interface & Model Definitions: 153/170

**Interface signatures typed (38/40)**:
All backend service signatures use Go types. `Delete`, `Move`, `ListByUserMembership`, `ListByItemKeysInRange` all have typed parameters and return types. Frontend TypeScript types are provided.

Deduction (-2): The `filterInactiveTerminal` function in Interface 8 lacks a typed return value and the `WeeklyComparisonGroup` type is not defined inline. A developer would need to look up the existing type to understand what fields are available for filtering decisions.

**Inline models concrete (38/40)**:
`MainItemMatchInfo`, `MoveResult`, `MainItemFilter`, `GanttFilter`, `EditSubItemFormState` all have typed field definitions. The Field Quick Reference table covers key fields.

Deduction (-2): The `MoveResult` struct uses `int64` for `MainItemBizKey` but the API returns it as a string in the JSON example (`"mainItemBizKey": "729384756"`). The struct definition lacks a JSON tag, so it is unclear how the int64-to-string conversion happens at the API boundary. Per convention AB-001, the API boundary should use string bizKey.

**ER diagram complete (28/30)**:
The Mermaid erDiagram includes 7 entities with column details and relationships. The Index Design section references existing indexes and their usage patterns for this feature.

Deduction (-2): `pmw_users` entity in the erDiagram has no column details. While not directly modified, `changed_by` in status_histories references users and the team selector filter (#15) involves user membership lookups.

**SQL DDL directly usable (28/30)**:
The schema.sql now clearly documents its intentional non-executable status and references the Go migration code (`SyncPresetRoles`) that handles the actual seed data insertion. The INSERT statements are provided as documentation with proper explanation.

Deduction (-2): While the file is now intentionally documentation-only, the INSERT statements use `INSERT IGNORE` syntax which is MySQL-specific. SQLite uses `INSERT OR IGNORE`. Since the project supports both databases, this could cause confusion if someone tries to run them on SQLite.

**Cross-layer consistency (21/30)**:
The Cross-Layer Data Map aligns most fields across layers. The `assigneeKey` field conversion from `*string` to `*int64` is now noted: "query param 为 string，Service 层用 `pkg.ParseID` 转为 int64".

Deduction (-9): The Cross-Layer Data Map still shows `*int64` for assigneeKey at the Backend Model level, but the actual Go struct field is `*string`. The "Validation Rule" column says "穿透到子事项匹配" which is a business rule, not a validation rule. The `matchType` and `matchedSubItemIds` are listed as "(计算值)" at the storage layer but have no storage column — the "Storage Layer" column should be empty or N/A rather than implying a computation happens at that layer. The `targetMainItemBizKey` row shows `pmw_main_items.biz_key BIGINT` as the Storage Layer, but this is the target main item's biz_key, not a column of the request — it is a lookup target, not a stored field of the request.

### 3. Error Handling: 118/130

**Error types defined (42/45)**:
Six error codes defined in a table with names, descriptions, and HTTP status codes. Two new business errors use correct struct construction: `&apperrors.AppError{Code: "BAD_REQUEST", Message: "...", Status: 400}`.

Deduction (-3): The error table lists `CONFLICT` (409) and `INVALID_STATUS` (422) but no interface actually uses these codes. The delete interfaces only return 404 and 403. The move interface returns 400 and 404. These error codes are defined but have no mapping to specific operations in the design, making it unclear when they would be triggered.

**Propagation strategy clear (40/45)**:
Backend propagation now correctly references `apperrors.RespondError(c, err)`. Frontend propagation is described with `mutation onError` pattern. Transaction failure handling is documented: "GORM 回滚后返回 ErrInternal，RespondError 自动映射为 500".

Deduction (-5): The design says "沿用现有 `apperrors` 包 + `pkgerrors.MapNotFound` 映射" but `pkgerrors` is still referenced as if it is a separate package. The actual import is `apperrors "pm-work-tracker/backend/internal/pkg/errors"`. There is no `pkgerrors` alias in the codebase. This is a stale reference from iteration 1.

**HTTP status codes mapped (36/40)**:
All error codes have HTTP status mappings. The API handbook provides per-endpoint error response tables.

Deduction (-4): The `ErrTargetClosed` and `ErrSameMainItem` errors are both defined with `BAD_REQUEST` code and `Status: 400`, but the api-handbook.md shows them as separate error rows with the same status code 400 and different descriptions. A client receiving a 400 cannot distinguish which business rule failed from the status code alone — it must parse the `code` field. However, both use `BAD_REQUEST` as the code. The design should either use distinct error codes (e.g., `TARGET_CLOSED`, `SAME_MAIN_ITEM`) or document that the `message` field is the discriminator.

### 4. Testing Strategy: 118/130

**Per-layer test plan (42/45)**:
The test plan now includes all layers: Backend Service, Backend Handler, Backend Middleware, Backend Repository, Frontend Component, Frontend Hook. Repository layer changes are now explicitly listed with specific methods.

Deduction (-3): The Frontend Hook test plan lists `useItemViewPage` but does not mention testing the Gantt view page hook for the new multi-status filter (#12). The gantt filter change affects `GanttViewPage` which likely has its own data fetching logic or hook.

**Coverage target numeric (42/45)**:
Per-layer targets: Service 85%, Handler 80%, Middleware 100%, Repository 80%, Frontend Component 80%, Frontend Hook 80%. Overall: 80%.

Deduction (-3): No tool configuration for measuring coverage is mentioned (e.g., `go test -coverprofile`, `vitest --coverage`). Without specifying how coverage will be measured and enforced, the targets are aspirational rather than verifiable.

**Test tooling named (34/40)**:
Backend: "Go testing + testify" and "Go testing + httptest". Frontend: "Vitest + @testing-library/react".

Deduction (-6): The design still does not specify which testify packages (assert? mock? require?) or which @testing-library packages (react? user-event? jest-dom?). For middleware testing, no mocking strategy is described — how will the `TeamScopeMiddleware` nil RoleKey scenario be set up in a test? The existing test pattern `setupTeamScopeRouter` is not referenced.

### 5. Breakdown-Readiness: 163/180

**Components enumerable (62/65)**:
All components are listed: 3 new API endpoints, 5 enhanced endpoints, 3 new service methods, 2 new repo methods, multiple frontend component changes. The PRD Coverage Map lists all 16 items with design components and interface references.

Deduction (-3): The seed data migration is still not enumerated as a separate component. The document references `SyncPresetRoles` but does not define what code changes are needed to register the new permission codes (`main_item:delete`, `sub_item:delete`) in `internal/pkg/permissions/codes.go` before `SyncPresetRoles` runs. The schema.sql hints at this: "The Go code in internal/pkg/permissions/codes.go must declare the new codes before SyncPresetRoles runs" but this is not tracked as a taskable component.

**Tasks derivable (58/65)**:
Interfaces 1-10 provide clear contracts with typed signatures and transaction sequences. Backend tasks are well-derivable.

Deduction (-7): Interface 11 still bundles 7 distinct frontend items (#1, #2, #4, #6, #7, #13, #14) into one section. The CSS-only change for #14 (macOS scrollbar) is mixed with JavaScript component changes. A task breakdown would need to split these manually. Each item should have explicit task boundaries or the document should enumerate sub-tasks within Interface 11.

**PRD AC coverage (43/50)**:
The PRD Coverage Map maps all 16 PRD items to design components. Key ACs are addressed: delete cascade with status_history (Interfaces 1, 2), move with renumber (Interface 3), filter passthrough with matchType (Interface 4), terminal sort (Interface 6), team selector filter (Interface 9).

Deduction (-7): Two PRD AC gaps remain:
1. PRD User Story #3 AC says "确认对话框（提示'将同时删除 N 个子事项'）" but the design does not specify how the frontend obtains `subItemCount` for the confirmation dialog. Interface 1 returns `{ "message": "ok" }` with no sub-item count. The Integration Specs section mentions "确认对话框需 subItemCount" but no API endpoint or response field provides this count before deletion.
2. PRD User Story #8 AC says "不存在无角色（role_key 为空）的团队成员记录" but the design only fixes the middleware to handle nil RoleKey — it does not address ensuring all team members have a role assigned. The PRD AC implies a data cleanup or migration to assign roles to members with NULL role_key.

### 6. Security Considerations: 70/80

**Threat model present (34/40)**:
Three threats identified: unauthorized deletion (medium), cross-team move (medium), permission code bypass (low). The design now includes explicit cross-team validation in the Move service logic.

Deduction (-6): Missing threats:
1. **Information disclosure via filter passthrough (#10)**: The `matchedSubItemIds` field returns sub-item bizKeys for items the user may not have direct visibility into. While all operations are within `TeamScopeMiddleware`, the design does not discuss whether a user who can see main items should also see sub-items assigned to other users. No threat analysis for this data exposure.
2. **Denial of service via unbounded filter combinations**: The multi-select status filter and assignee passthrough could generate complex in-memory filtering operations. The design states performance target of ≤500ms for 1000+5K records but does not analyze threat scenarios with adversarial filter combinations.

**Mitigations concrete (36/40)**:
Mitigations are paired with threats and now include explicit cross-team validation for Move: "Service 层显式校验目标主事项 `team_key == 当前 teamBizKey`". Team isolation via `TeamScopeMiddleware` is documented. Defensive UI via `PermissionGuard` is noted as non-security-boundary.

Deduction (-4): The mitigation for "越权移动" says "复用现有 `sub_item:update` 权限码" but this means any user with `sub_item:update` can move sub-items between any main items within the same team. A member user with `sub_item:update` could potentially move items they do not own. The design does not discuss whether the move operation should additionally check if the user is the assignee or has a more specific permission.

### 7. Implementation Feasibility: 130/140

**Dependencies available (46/50)**:
All referenced packages exist. `@radix-ui/react-checkbox` is new but Radix is already in use. GORM is the existing ORM. The error package exists with `AppError` struct.

Deduction (-4): The design references `pkg.ParseID` for converting string assigneeKey to int64. This function is not documented as an existing dependency. If it does not exist, it would need to be created. The design assumes it exists without verification.

**Architecture fits project structure (46/50)**:
The proposed architecture fits the existing four-layer structure. New handlers, services, and repositories extend existing patterns. No new layers introduced.

Deduction (-4): The `MainItemService.Delete` method takes `operatorBizKey int64` parameter. The design does not verify that the existing `MainItemService` constructor or dependency injection pattern supports adding methods with different parameter signatures than existing methods. The existing `MainItemService` methods may not accept `operatorBizKey` — this needs verification against the actual codebase.

**Technical claims grounded (38/40)**:
Performance claims reference concrete benchmarks. The atomic counter approach for `sub_item_seq` is grounded in the schema. SQLite/MySQL dual-support is acknowledged.

Deduction (-2): The design claims "并发安全性由原子 UPDATE + 唯一索引兜底保证" for the move operation but does not analyze what happens when two concurrent moves target the same main item. The atomic increment of `sub_item_seq` prevents duplicate codes, but if the read-back of the new sequence number (step 4) happens after another concurrent increment, the code generation could use a stale value. The design should clarify whether the read-back is within the same atomic operation or requires additional synchronization.

## Attack Points

1. **Breakdown-Readiness**: PRD AC gap — delete confirmation dialog needs sub-item count — The PRD User Story #3 AC states: "系统弹出确认对话框（提示'将同时删除 N 个子事项'）" but no API response provides the sub-item count before deletion. The Integration Specs mention "确认对话框需 subItemCount" but there is no data source defined for this count. The design must either add a field to the main item detail response or document that the frontend counts locally from already-loaded sub-item data.

2. **Breakdown-Readiness**: PRD AC gap — role_key NULL data cleanup — PRD User Story #6 AC states "不存在无角色（role_key 为空）的团队成员记录" but Interface 10 only fixes the middleware to handle nil RoleKey at runtime. The design does not include a data migration to assign roles to existing members with NULL role_key, which the PRD AC implies should not exist.

3. **Breakdown-Readiness**: Interface 11 bundles 7 frontend items — Items #1, #2, #4, #6, #7, #13, #14 are grouped under Interface 11 without explicit task boundaries. Quote: "### Interface 11: 前端组件接口变更（#1,#2,#4,#6,#7,#13,#14）". Each item has distinct component targets and should be enumerated separately for task derivation.

4. **Error Handling**: Stale `pkgerrors` reference — The propagation strategy states "沿用现有 `apperrors` 包 + `pkgerrors.MapNotFound` 映射" but `pkgerrors` is not a real package alias in the codebase. This reference from iteration 1 was not cleaned up. Remove the `pkgerrors` reference or clarify the actual import path.

5. **Error Handling**: `CONFLICT` and `INVALID_STATUS` error codes defined but unused — The error table defines `CONFLICT` (409) and `INVALID_STATUS` (422) but no interface in the design uses these codes. If they are intended for future use or for existing status transition flows, this should be stated. If not needed for this batch, they should be removed to avoid confusion.

6. **Cross-Layer**: AssigneeKey type mismatch in Cross-Layer Data Map — The Backend Model column shows `*int64` for assigneeKey but the Go struct `MainItemFilter.AssigneeKey` is `*string`. The design notes "Service 层用 `pkg.ParseID` 转为 int64" but the map should reflect the actual filter struct type (`*string`) at the Backend Model level, with the conversion noted separately.

7. **Interface & Model**: MoveResult JSON serialization gap — The `MoveResult` struct uses `MainItemBizKey int64` but the API JSON example shows `"mainItemBizKey": "729384756"` as a string. Per convention AB-001, the API boundary must use string bizKey. The struct needs a JSON tag and the design should specify how int64-to-string conversion occurs.

8. **Testing**: Frontend Gantt hook test missing — The per-layer test plan lists `useItemViewPage` for filter testing but does not mention testing the Gantt page data fetching logic for the new multi-status filter parameter (#12). The gantt filter change affects `GanttFilter` and the `getGanttViewApi` call — this needs a test scenario.

9. **Security**: No ownership check on sub-item move — The move operation requires `sub_item:update` permission but does not verify the user is the assignee of the sub-item being moved. Quote from mitigations: "移动：复用现有 `sub_item:update` 权限码". This means any team member with `sub_item:update` can move any sub-item, including items assigned to other users. The threat model should acknowledge this or the design should add an ownership check.

10. **Architecture Clarity**: Interface 8 method name inconsistency — The typed signature defines `ListByItemKeysInRange` but the explanatory text in Interface 8 still says "需要新增 `StatusHistoryRepo.ListByTeamInRange` 方法". These are two different method names for what appears to be the same functionality. A developer would not know which name to implement.

11. **[blindspot]** Weekly filter "变更类型为进度更新" condition is unimplementable — Interface 8 condition (c) says "该主事项的 db_update_time 在时间范围内且变更类型为进度更新" but there is no `变更类型` column in the `pmw_main_items` table. The Field Quick Reference lists `item_status`, `completion_pct`, etc. but no change-type field. The design cannot determine whether an `updated_at` change was a "进度更新" versus any other update without a change-type tracking mechanism. This condition is either unimplementable or needs a concrete implementation strategy (e.g., comparing `completion_pct` before/after).

12. **[blindspot]** PRD Coverage Map mis-indexes #8 — The coverage map row for #8 says "Interface 9 (后端中间件)" but Interface 9 is actually "团队选择器权限过滤 (#15)". The member permission fix (#8) is described in Interface 10 in the document body. This index mismatch from iteration 1 was not fixed and would misdirect a developer.

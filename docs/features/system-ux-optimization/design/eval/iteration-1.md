# Design Eval — Iteration 1

**Score**: 770/1000

## Dimension Scores

### 1. Architecture Clarity: 155/170

**Layer placement explicit (55/60)**:
The document clearly states all 16 items belong to the existing four-layer architecture (Router->Handler->Service->Repository->Model) with no new layers. The overview table categorizes each item by layer. The "Architecture > Layer Placement" section states: "所有 16 项变更位于现有四层架构（Router→Handler→Service→Repository→Model）内。无新层。" This is explicit and complete.

Deduction (-5): The layer placement for the Member permission fix (#8) is described as "Backend middleware + API" in the overview table but the actual change is only in middleware (`TeamScopeMiddleware nil RoleKey`), not in an API endpoint. Minor mislabel.

**Component diagram present (55/60)**:
A detailed ASCII component diagram is present showing Frontend components, Backend Router/Handlers/Services/Repositories/Middleware, and Database. Each component is mapped to its corresponding requirement item (#1-#16). The diagram includes interaction flows (e.g., `DELETE /main-items/:itemId (#3) <- main_item:delete`).

Deduction (-5): The diagram does not show data flow arrows between layers (e.g., Handler -> Service -> Repository call chain). It lists components hierarchically but lacks directional arrows showing how a request flows through the system.

**Dependencies listed (45/50)**:
A dependencies table lists three items: `@radix-ui/react-alert-dialog` (existing), `@radix-ui/react-checkbox` (new), and GORM `Expr("sub_item_seq + 1")` (existing). The table includes type, purpose, and version.

Deduction (-5): The design mentions `StatusHistoryRepo.ListByTeamInRange` as a new repository method needed for #16 but does not list it in the dependencies section or explain its interface contract.

### 2. Interface & Model Definitions: 135/170

**Interface signatures typed (35/40)**:
Backend service signatures are typed with Go function signatures (e.g., `Delete(ctx context.Context, teamBizKey, itemBizKey int64, operatorBizKey int64) error`). Frontend TypeScript types are provided for new fields. Repository method signatures are given for new methods (`ListByUserMembership`).

Deduction (-5): The `StatusHistoryRepo.ListByTeamInRange` method referenced in Interface 8 lacks a typed signature. The design says "需要新增 StatusHistoryRepo.ListByTeamInRange 方法获取时间范围内的状态变更记录" but provides no parameter or return types.

**Inline models concrete (35/40)**:
`MainItemMatchInfo`, `MoveResult`, `MainItemFilter`, `GanttFilter`, and `EditSubItemFormState` all have typed field definitions. The Field Quick Reference table covers key fields for MainItem, SubItem, StatusHistory, and pmw_role_permissions.

Deduction (-5): The `filterInactiveTerminal` function signature in Interface 8 shows parameters but the `WeeklyComparisonGroup` type is not defined inline — it references an existing type from the codebase without explaining its structure. A developer would need to look up the existing type to understand what fields are available for filtering.

**ER diagram complete (25/30)**:
The er-diagram.md contains a Mermaid erDiagram with 7 entities (pmw_teams, pmw_main_items, pmw_sub_items, pmw_team_members, pmw_users, pmw_roles, pmw_role_permissions, pmw_status_histories). Relationships and cardinality are specified with proper Mermaid notation.

Deduction (-5): The `pmw_users` entity in the erDiagram has no column details (only the relationship box). While it is an existing table not directly modified, its fields are referenced in the design (e.g., `changed_by BIGINT` in status_histories references users). The Entity Details section also omits pmw_users entirely.

**SQL DDL directly usable (15/30)**:
The schema.sql file is entirely commented out. Every SQL statement is wrapped in `--` comments. The file states "No DDL changes — seed data only" which is correct in intent, but the seed INSERT statements are commented out rather than being executable. A developer cannot run this file as-is to set up the required seed data.

Deduction (-15): Per rubric, "Can schema.sql be executed as-is?" — No, it cannot. All statements are commented out. The file header even says "No DDL changes — seed data only" and the actual INSERTs are commented with `--`. The document acknowledges that `SyncPresetRoles` handles idempotent inserts, but the SQL file should still be directly executable for manual setup or verification.

**Cross-layer consistency (25/30)**:
The Cross-Layer Data Map table aligns storage layer column names with backend model types, API/DTO JSON tags, and frontend types. For example, `assigneeKey` maps through `pmw_main_items.assignee_key BIGINT` -> `*int64` -> `json:"assigneeKey"` -> `string | null`.

Deduction (-5): The `assigneeKey` field in `MainItemFilter` is typed as `*string` in the Go struct but the Cross-Layer Data Map shows it as `*int64` at the Backend Model level. The filter struct definition says `AssigneeKey *string` but the actual column is `assignee_key BIGINT`. This is a type mismatch — the filter should accept string (from form param) but internally needs to parse to int64. The design does not clarify this conversion.

### 3. Error Handling: 110/130

**Error types defined (40/45)**:
Six error codes are defined in a table: NOT_FOUND, FORBIDDEN, BAD_REQUEST, INVALID_STATUS, CONFLICT, INTERNAL. Each has a name, description, and HTTP status. Two new business errors are defined: `ErrTargetClosed` and `ErrSameMainItem`.

Deduction (-5): The design defines new errors using `apperrors.New("BAD_REQUEST", "目标主事项已关闭")` but the actual codebase uses `AppError` struct construction (e.g., `&AppError{Code: "...", Message: "...", Status: ...}`). The `apperrors.New()` function shown in the design does not exist in the codebase's errors package. This is a code-level inaccuracy that would cause confusion during implementation.

**Propagation strategy clear (35/45)**:
The document states: "Handler 层统一通过 `handleError(c, err)` 映射到 HTTP 响应" and "GORM 回滚后返回 ErrInternal，前端展示通用错误提示，数据不变." Frontend propagation is described: mutation `onError` from `err.response.data.message`.

Deduction (-10): The design references `handleError(c, err)` as the handler-layer error mapping function, but this function does not exist in the codebase. The actual function is `apperrors.RespondError(c, err)`. This is a significant inaccuracy — a developer following this design would look for a nonexistent function. Additionally, the design says "沿用现有 `apperrors` 包 + `pkgerrors.MapNotFound` 映射" but `pkgerrors` is not a real package alias in the codebase. The actual import is `apperrors "pm-work-tracker/backend/internal/pkg/errors"`.

**HTTP status codes mapped (35/40)**:
Error codes are mapped to HTTP status codes: NOT_FOUND->404, FORBIDDEN->403, BAD_REQUEST->400, INVALID_STATUS->422, CONFLICT->409, INTERNAL->500. The API handbook also lists status codes per endpoint.

Deduction (-5): The `ErrTargetClosed` and `ErrSameMainItem` errors are both defined with `BAD_REQUEST` code but the design does not show how they map to the HTTP status code 400 in the actual error response. The api-handbook.md shows these correctly, but the tech-design.md error definition uses `apperrors.New("BAD_REQUEST", ...)` which implies the AppError knows to return 400, but the actual AppError struct requires an explicit Status field.

### 4. Testing Strategy: 110/130

**Per-layer test plan (40/45)**:
A comprehensive table lists test types per layer: Backend Service (unit), Backend Handler (integration), Backend Middleware (unit), Frontend Component (unit), Frontend Hook (unit). Each row specifies the tool and what to test.

Deduction (-5): Repository layer changes (SubItemRepo.ListByMainItem ORDER BY change, MainItemRepo.ListNonArchivedByTeam) are not listed in the test plan. While these are minor query changes, they represent testable behavior at the repository layer.

**Coverage target numeric (40/45)**:
Numeric targets are provided: Service 85%, Handler 80%, Middleware 100%, Frontend Component 80%, Frontend Hook 80%. Overall: 80%.

Deduction (-5): The overall coverage target (80%) and per-layer targets are stated, but no rationale is given for why 85% for services vs 80% for handlers. No tool configuration for measuring coverage is mentioned (e.g., `go test -coverprofile`, `vitest --coverage`).

**Test tooling named (30/40)**:
Backend: "Go testing + testify" and "Go testing + httptest". Frontend: "Vitest + @testing-library/react".

Deduction (-10): The design does not specify which testify packages (assert? mock? require?) or which @testing-library packages (react? user-event? jest-dom?). For the middleware test, no mocking strategy is described (how will the TeamScopeMiddleware nil RoleKey scenario be set up?). The existing test files use specific patterns (e.g., `setupTeamScopeRouter` in team_scope_test.go) that should be referenced.

### 5. Breakdown-Readiness: 158/180

**Components enumerable (60/65)**:
All components are listed and countable: 3 new API endpoints (DELETE main-item, DELETE sub-item, PUT move), 5 enhanced existing endpoints, 3 new service methods (Delete, Delete, Move), 1 new repo method (ListByUserMembership), 1 new repo method (ListByTeamInRange), multiple frontend component changes. The PRD Coverage Map lists all 16 items with their design components and interface references.

Deduction (-5): The design does not enumerate the seed data migration as a separate component/task. It mentions `SyncPresetRoles` handles idempotent inserts but does not define where the seed data insertion code lives or how it is triggered.

**Tasks derivable (55/65)**:
Each interface maps to implementation work. Interfaces 1-10 provide clear contracts. The transaction sequences for Delete and Move are detailed enough to derive tasks.

Deduction (-10): Several frontend items are grouped under "Interface 10: 前端组件接口变更" which bundles 7 distinct requirement items (#1, #2, #4, #6, #7, #13, #14) into one interface section. This makes task derivation harder — a task breakdown would need to split these into separate tasks. The design acknowledges these are "前端 only" but the bundling obscures the task boundaries. Additionally, the CSS-only change for #14 (macOS scrollbar) is mixed with JavaScript component changes.

**PRD AC coverage (43/50)**:
The PRD Coverage Map maps all 16 PRD items to design components and interfaces. Key acceptance criteria are addressed:
- #1: StatusTransitionDropdown Alert replaces tooltip
- #3: Delete with confirmation, cascade, status_history
- #9: Move with renumber, status/assignee preservation
- #10: Filter passthrough with matchType
- #15: Team list filtered by membership

Deduction (-7): PRD User Story #3 AC says "status_history 记录删除事件" but the design's transaction sequences for delete (Interfaces 1 and 2) do not include a step for creating a status_history record. Step 5 in Interface 1 says "重新计算相关主事项 completion_pct" but there is no step for inserting into status_history. This is an AC gap.

### 6. Security Considerations: 65/80

**Threat model present (30/40)**:
Three threats are identified: unauthorized deletion (medium), cross-team move (medium), permission code bypass (low).

Deduction (-10): The threat model is thin. Missing threats:
1. **Mass assignment in Move**: The request body contains `targetMainItemBizKey` but no validation is described for ensuring the target belongs to the same team (the design says "team_key 校验" in mitigations but the threat table does not list cross-team move as a separate threat pattern).
2. **Information disclosure via filter passthrough**: The `matchedSubItemIds` field could leak sub-item existence to users who shouldn't know about items in other teams. No threat analysis for this.
3. **IDOR in delete endpoints**: The design uses bizKey in URLs but does not discuss whether sequential or guessable bizKeys could enable enumeration attacks.

**Mitigations concrete (35/40)**:
Mitigations are paired with threats: `RequirePermission` middleware for delete, `sub_item:update` for move, `TeamScopeMiddleware` for team isolation, `PermissionGuard` for defensive UI.

Deduction (-5): The move mitigation says "目标主事项通过 team_key 校验属于同一团队（事务内 FindByBizKey 自动校验团队成员关系）" but `FindByBizKey` looks up by bizKey and does not inherently verify team membership. The actual team membership check is done by `TeamScopeMiddleware` at the router level using the `:teamId` URL parameter. The move endpoint path `/teams/:teamId/sub-items/:subId/move` uses the team scope, but the `targetMainItemBizKey` in the body could reference a main item in a different team. The design should explicitly state the cross-team validation step in the service layer.

### 7. Implementation Feasibility: 137/140

**Dependencies available (48/50)**:
All referenced packages exist in the project. `@radix-ui/react-checkbox` is new but Radix is already in use (8 Radix packages in package.json). GORM is the existing ORM. `apperrors` (aliased as `pm-work-tracker/backend/internal/pkg/errors`) exists.

Deduction (-2): The design references `pkgerrors.MapNotFound` which is not a real package/function in the codebase. The actual function is `apperrors.MapNotFound`. While this is a naming inaccuracy rather than a missing dependency, it could cause implementation confusion.

**Architecture fits project structure (48/50)**:
The proposed architecture fits the existing layer structure perfectly. New handlers follow existing patterns (MainItemHandler, SubItemHandler, ViewHandler). New service methods extend existing services. No new layers or architectural patterns are introduced.

Deduction (-2): The design proposes `MainItemService.Delete` which would need to accept `operatorBizKey` parameter, but the existing `MainItemService` may not have methods that take an operator context. The design does not verify this fits the existing service constructor/injection pattern.

**Technical claims grounded (41/40)**:
Performance claims reference concrete benchmarks ("1000 主事项 + 5000 子事项"). The atomic counter approach for sub_item_seq is grounded in the existing schema. SQLite/MySQL dual-support is acknowledged. The ViewService in-memory filtering approach references the existing `view_service.go`.

## Attack Points

1. **Architecture Clarity**: Missing data flow arrows in component diagram — The component diagram lists components hierarchically but a developer cannot trace how a DELETE request flows from Router -> Handler -> Service -> Repository -> Database. Add directional arrows or a sequence-style notation for key flows.

2. **Interface & Model Definitions**: schema.sql is not executable — Every statement in `design/schema.sql` is commented out with `--`. The file states "No DDL changes — seed data only" but even the seed INSERTs are comments. Per convention, `SyncPresetRoles` handles this, but the SQL file should serve as a verifiable artifact. Either make the INSERTs executable or document that schema.sql is intentionally non-executable and reference the Go migration instead.

3. **Interface & Model Definitions**: `StatusHistoryRepo.ListByTeamInRange` lacks interface contract — Interface 8 says "需要新增 StatusHistoryRepo.ListByTeamInRange 方法获取时间范围内的状态变更记录" but provides no typed signature, parameters, or return types. Every other new repository method has a signature.

4. **Error Handling**: `handleError(c, err)` does not exist — The design states "Handler 层统一通过 handleError(c, err) 映射到 HTTP 响应" but the actual codebase uses `apperrors.RespondError(c, err)`. This is a direct contradiction with the codebase. Replace all references to `handleError` with `apperrors.RespondError`.

5. **Error Handling**: `apperrors.New()` is not the actual constructor — The design shows `ErrTargetClosed = apperrors.New("BAD_REQUEST", "目标主事项已关闭")` but the codebase's errors package uses `AppError` struct with `Code`, `Message`, and `Status` fields. There is no `New()` function. Use the actual struct construction pattern.

6. **Testing Strategy**: Repository layer changes not in test plan — `SubItemRepo.ListByMainItem` ORDER BY change and `MainItemRepo.ListNonArchivatedByTeam` status sort are repository-level behavior changes that should have test scenarios.

7. **Breakdown-Readiness**: Delete operation missing status_history recording — PRD User Story #3 AC states "status_history 记录删除事件" but the transaction sequences in Interfaces 1 and 2 have no step for inserting a status_history record. Step 5 in Interface 1 recalculates completion_pct but does not record the deletion event. This is a direct AC gap.

8. **Breakdown-Readiness**: Interface 10 bundles 7 distinct frontend items — Items #1, #2, #4, #6, #7, #13, #14 are grouped under a single "Interface 10" section. Task breakdown requires splitting these into individual tasks. Each item should have its own interface section or the document should explicitly enumerate task boundaries within Interface 10.

9. **Security**: Cross-team target validation gap in Move — The Move endpoint accepts `targetMainItemBizKey` in the request body. While `TeamScopeMiddleware` validates the `:teamId` from the URL, the body parameter could reference a main item in a different team. The mitigation says "FindByBizKey 自动校验团队成员关系" but `FindByBizKey` does not perform team membership checks — it only looks up the record by bizKey. An explicit cross-team validation step must be added to the Move service logic.

10. **Cross-layer**: AssigneeKey type mismatch — `MainItemFilter.AssigneeKey` is typed as `*string` (from form param) but the Cross-Layer Data Map shows Backend Model as `*int64`. The conversion from string to int64 at the service layer is not documented.

11. **[blindspot]** PRD Coverage Map mis-indexes Interface 9 — The coverage map row for #8 says "Interface 9 (后端中间件)" but Interface 9 in the document is actually "团队选择器权限过滤 (#15)". The middleware fix for #8 is described in the Architecture section under "Component Diagram > Middleware" but has no dedicated Interface section. A developer looking up #8 in the coverage map would be directed to the wrong interface.

12. **[blindspot]** Weekly progress filter (#16) "活跃定义" ambiguity — Interface 8's `filterInactiveTerminal` references "status_history 有变更 / 子事项 created_at 或 updated_at 在范围内 / 主事项 updated_at 且为进度更新" but the PRD spec has a more detailed definition: "status_history 表中存在该主事项的状态变更记录（created_at 在时间范围内）" and "该主事项下存在 created_at 或 updated_at 在时间范围内的子事项". The design does not specify whether "活跃子事项" means the sub-item's own `updated_at` or its `status_history` changes. This ambiguity affects the SQL query implementation.

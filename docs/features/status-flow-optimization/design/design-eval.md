# Design Evaluation: Status Flow Optimization

**Date**: 2026-04-20
**Evaluator**: automated
**Design**: `docs/features/status-flow-optimization/design.md`
**PRD**: `docs/features/status-flow-optimization/prd.md`

---

## 1. Overall Grade: **B**

The design is thorough and directly implementable. All required sections are present, interfaces are typed with concrete method signatures, models are fully defined, and the breakdown-readiness is strong. The two areas holding it back from an A are: (1) no explicit testing strategy section despite the project having a TDD rule, and (2) the linkage warning mechanism has a minor design inconsistency (the LinkageResult struct is introduced in the "Frontend Changes" section rather than where interfaces are defined, and its integration into SubItemService.ChangeStatus return type is described narratively rather than as a concrete signature change).

---

## 2. Dimension-by-Dimension Scores

### Dimension 1: Architecture Clarity -- **A**

- **Layer placement**: Explicit. The Overview names every layer (Router, Handler, Service, Repository, Model) and the file-level breakdown in the Overview and File Change Summary maps each component to its layer.
- **Component diagram**: The Overview contains a text-based package/file tree showing NEW vs MOD annotations. While not an ASCII box-and-arrow diagram, it clearly communicates component placement and data flow.
- **Dependencies**: The `pkg/status` package rationale explicitly explains why it exists as a shared package (avoids circular deps, single source of truth). The service-to-service dependency (SubItemService depends on MainItemService for linkage) is documented.
- **Consistency with project**: Matches existing patterns -- handler constructors with panic-on-nil, repo interface in `repository/` with GORM impl in `repository/gorm/`, VO conversion in handlers, service interfaces with constructors. The Dependencies struct extension in `router.go` follows the existing wiring pattern.

**Justification**: All criteria met. The text-tree diagram is sufficient for this project's complexity.

### Dimension 2: Interface & Model Definitions -- **A**

- **Interface signatures**: All typed. `ChangeStatus(ctx, teamID, callerID, itemID uint, newStatus string) (*model.MainItem, error)`, `EvaluateLinkage(ctx, mainItemID, changedBy uint) error`, `AvailableTransitions(ctx, teamID, callerID, itemID uint) ([]string, error)`, `StatusHistoryRepo.Create(ctx, record) error`, `StatusHistoryRepo.ListByItem(ctx, itemType, itemID, page) (*PageResult, error)`, `StatusHistoryService.Record(ctx, record) error`.
- **Model fields**: `StatusHistory` struct is fully defined with GORM tags, JSON tags, and types. The naming conventions file requires camelCase JSON tags and lets GORM handle snake_case columns -- the model follows this.
- **Completeness**: All major components defined -- pkg/status (two files with types and functions), model, repo (interface + impl), service (interface + impl), handler methods, DTO changes, VO changes, frontend status.ts, API calls.
- **Implementable**: A developer can code directly from these definitions without guessing. The transition maps are concrete Go maps. The method flows are step-by-step numbered lists.

**Minor gap**: The `LinkageResult` struct is defined in the Frontend Changes section (line ~422 of design) rather than alongside the service interfaces. Its impact on `SubItemService.ChangeStatus` return type (currently `error` in the existing code, needs to become `(*model.SubItem, *LinkageResult, error)` or similar) is described narratively rather than as a concrete signature. This is a documentation organization issue, not a design gap.

**Justification**: All interfaces typed, all models concrete. Directly implementable with one minor narrative-vs-signature inconsistency.

### Dimension 3: Error Handling -- **B**

- **Error types**: The design references `ErrInvalidStatus` which already exists in `pkg/errors/errors.go` (defined as `&AppError{Code: "INVALID_STATUS", Status: 422}`). Self-transition returns this same error. The existing `ErrItemNotFound` is used for item lookup failures.
- **Propagation strategy**: Implicit rather than explicit. The design follows the project's existing pattern (service returns AppError, handler calls `apperrors.RespondError`), but does not call this out as a deliberate choice.
- **HTTP mapping**: Relies on the existing `AppError.Status` field (422 for invalid transitions, 404 for not found). The new permission code `main_item:change_status` would produce 403 via existing middleware. This is not explicitly stated but follows from project patterns.
- **Client behavior**: The linkage failure path is well-documented -- `LinkageResult` returned to handler, included in API response as `linkageWarning`, frontend shows toast. Invalid transitions return 400/422 (existing pattern).

**Justification**: Error types exist and are reused. Propagation follows existing patterns but is not explicitly stated. HTTP codes are not listed for the new endpoints. The design would benefit from a brief table mapping the 3 new endpoints to their error responses.

### Dimension 4: Testing Strategy -- **C**

- **Per-layer plan**: Not present as a dedicated section. The design mentions test files at the end (File Change Summary -> Test Files to Update) and calls out that `pkg/status` needs its own tests, new handler tests for 3 endpoints, and service tests for EvaluateLinkage covering all 5 priority levels. But this is a paragraph, not a structured testing strategy.
- **Test types**: Not specified per layer. No mention of unit vs integration distinction for the new components.
- **Coverage target**: No numeric target.
- **Test tooling**: Not named (though the project uses `testify` + `httptest` per `.claude/rules/testing.md`).

**Justification**: The project's TDD convention (`.claude/rules/testing.md`) requires red-green-refactor and co-located test files. The design references test needs but does not provide a structured testing strategy section. This is the weakest dimension.

### Dimension 5: Breakdown-Readiness -- **A**

- **Enumerable components**: The File Change Summary explicitly lists 7 new files and 17+ modified files. Each can become a task.
- **Interface tasks derivable**: Each interface (StatusHistoryRepo, StatusHistoryService, MainItemService.ChangeStatus, MainItemService.EvaluateLinkage, MainItemService.AvailableTransitions, SubItemService.ChangeStatus update, SubItemService.AvailableTransitions) maps to at least one implementation task.
- **Model tasks derivable**: StatusHistory model -> migration task, MainItem/SubItem model changes -> schema update task, pkg/status -> package creation task.
- **No ambiguous ownership**: Every component has a clear layer and file location. The concurrency mechanism (per-MainItem mutex) is owned by MainItemService.
- **PRD traceability**: See matrix below. All 23 ACs are addressed by specific design sections.

**Justification**: A developer can list and count all tasks from this design. The phased delivery in the PRD (Phase 1/2/3) maps cleanly to the design's components.

### Dimension 6: Security Considerations -- **B**

- **Threat model**: The design addresses the PM-only constraint for reviewing transitions (callerID == item.PmID check in service, PM-only filter in AvailableTransitions). The PRD has a clear RBAC requirement (AC-20).
- **Mitigations**: Concrete -- service-level check on `item.PmID`, server-side filtering in AvailableTransitions so non-PM users never see reviewing transitions, new permission code `main_item:change_status` gates the endpoint.
- **Scope-appropriate**: The depth is appropriate. The PRD's security surface is limited to PM-only transitions, which the design covers.

**Justification**: Threats identified (unauthorized reviewing transitions), mitigations concrete (service check + API filter + permission code). No dedicated Security Considerations section, but the mechanism is woven into the relevant service/handler descriptions. Marked B because mitigations are distributed rather than consolidated.

---

## 3. Structure Check

| Section | Required | Present | Status |
|---------|----------|---------|--------|
| Overview | Yes | Yes (with layer placement + tech stack) | Pass |
| Architecture | Yes | Yes (file tree + dependency rationale + execution flow) | Pass |
| Interfaces | Yes | Yes (6+ interfaces with typed method signatures) | Pass |
| Data Models | Yes | Yes (StatusHistory struct, pkg/status types, model changes) | Pass |
| Error Handling | Yes | Yes (references existing AppError types, error paths documented) | Pass |
| Testing Strategy | Yes | **No dedicated section** -- only a paragraph in File Change Summary | **Weak** |
| Security Considerations | Conditional (PRD has RBAC) | No dedicated section, but mechanisms woven into service design | Partial |
| Open Questions | Optional | Not present (decisions are made inline) | N/A |
| Alternatives Considered | Optional | Not present in design (present in PRD) | N/A |

---

## 4. PRD Traceability Matrix

| AC | Description | Design Coverage | Location |
|----|-------------|-----------------|----------|
| AC-1 | Status codes stored as English, API returns code+name | Yes | pkg/status + VO changes (StatusName field) |
| AC-2 | MainItem 10 legal transitions pass, illegal return 400 | Yes | MainItemTransitions map + ChangeStatus flow |
| AC-3 | SubItem 9 legal transitions pass, illegal return 400 | Yes | SubItemTransitions map + updated ChangeStatus |
| AC-4 | Self-transition returns error | Yes | Step 2 of ChangeStatus flow |
| AC-5 | MainItemUpdateReq no Status field | Yes | DTO changes section |
| AC-6 | Terminal sets completion=100 + actual_end_date | Yes | Step 6 of ChangeStatus, Step 4 of SubItem ChangeStatus |
| AC-7 | All subitems completed/closed + at least one completed -> reviewing | Yes | EvaluateLinkage priority rule 1 |
| AC-8 | All 5 linkage priority levels tested | Partial | Mentioned in test paragraph, no structured test plan |
| AC-9 | reviewing + new pending subitem -> progressing | Yes | SubItem Create -> EvaluateLinkage |
| AC-10 | Delete subitem triggers re-evaluation | Yes | SubItem Delete -> EvaluateLinkage |
| AC-11 | No subitems -> no linkage | Yes | EvaluateLinkage step 3 |
| AC-12 | Linkage failure: status unchanged, history records intent, toast shown | Yes | EvaluateLinkage step 6 + LinkageResult + toast |
| AC-13 | RecalcCompletion before linkage on completed | Yes | Execution Flow step 7-8 (fixed order) |
| AC-14 | ChangeStatus records status_histories | Yes | Step 8 of ChangeStatus, Step 6 of SubItem ChangeStatus |
| AC-15 | is_auto=true for linkage, false for manual | Yes | ChangeStatus records non-auto, EvaluateLinkage records auto |
| AC-16 | StatusDropdown calls ChangeStatus API | Yes | Frontend StatusDropdown section |
| AC-17 | StatusDropdown shows only available-transitions | Yes | Frontend StatusDropdown section |
| AC-18 | StatusBadge uses code-to-name map | Yes | Frontend StatusBadge + lib/status.ts |
| AC-19 | Overdue badge for non-terminal past-due items | Yes | lib/status.ts isOverdue + Overdue Badge section |
| AC-20 | reviewing -> completed/progressing PM-only | Yes | Service PM check + AvailableTransitions filter |
| AC-21 | Terminal confirmation dialog | Yes | Frontend StatusDropdown section |
| AC-23 | available-transitions API returns correct sets | Yes | AvailableTransitions methods + routes |

**Uncovered ACs**: None. All 22 ACs are addressed.

---

## 5. Top 3 Issues

### Issue 1: No Testing Strategy Section (Impact: Medium, Dimension 4 -- C)

The project enforces TDD via `.claude/rules/testing.md` (red-green-refactor, co-located test files, table-driven tests, specific tooling). The design mentions test needs in one paragraph but lacks a dedicated Testing Strategy section with per-layer test approaches, test types (unit/integration/e2e), coverage targets, and tooling. This is the only dimension scoring C.

**Fix**: Add a Testing Strategy section covering:
- `pkg/status`: table-driven unit tests for transitions, validation, lookup
- Service layer: mock repos, table-driven tests for all 5 linkage priority levels, ChangeStatus validation, PM-only enforcement
- Handler layer: httptest for 3 new endpoints
- Frontend: component tests for StatusBadge/StatusDropdown with new code-based lookup
- Coverage target (e.g., 90% for new code)

### Issue 2: LinkageResult Integration Not Fully Specified (Impact: Low-Medium)

The `LinkageResult` struct is introduced in the Frontend Changes section rather than with the service interface definitions. More critically, the existing `SubItemService.ChangeStatus` returns `error` (no return value). The design says SubItem's ChangeStatus should "return updated SubItem + LinkageResult" but does not show the updated interface signature. This is a gap between narrative and specification.

**Fix**: Show the updated `SubItemService` interface with the new return type:
```go
type SubItemChangeResult struct {
    SubItem         *model.SubItem
    LinkageResult   *LinkageResult
}
ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) (*SubItemChangeResult, error)
```
Place this in the "Modified Service: SubItemService" section.

### Issue 3: Security Mechanisms Not Consolidated (Impact: Low)

The PM-only enforcement is spread across three sections (ChangeStatus step 4, AvailableTransitions step 3, Frontend PM role control). There is no consolidated statement of the security model. While this is acceptable for a feature with a narrow security surface, a brief Security Considerations section would make the authorization model easier to verify during implementation.

**Fix**: Add a brief Security Considerations section consolidating:
- Permission gating: `main_item:change_status` on the endpoint
- Service-level PM check: `callerID == item.PmID` for reviewing transitions
- API-level filtering: AvailableTransitions removes PM-only options for non-PM callers
- Frontend: PM-only UI visibility (defense in depth, not sole enforcement)

---

## 6. Recommendation

**Proceed to breakdown-tasks.** The design scores B overall with Breakdown-Readiness at A. All ACs are covered, all interfaces are typed, and all models are concrete. The testing strategy gap (Issue 1) should be addressed before or during breakdown -- it affects how tasks are structured but does not block task derivation. Issues 2 and 3 are documentation clarity improvements that can be folded into the breakdown without requiring a design revision cycle.

**Breakdown-Readiness Grade: A** -- All components are enumerable, tasks are clearly derivable, and the PRD is fully covered.

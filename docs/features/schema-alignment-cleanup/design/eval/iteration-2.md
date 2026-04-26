# Tech Design Evaluation Report — Iteration 2

**Document:** `docs/features/schema-alignment-cleanup/design/tech-design.md`
**Scored:** 2026-04-26
**Target:** 90

---

## Previous Attack Point Resolution

| # | Attack (Iteration 1) | Status | Evidence |
|---|----------------------|--------|----------|
| 1 | `RecordStatusChange` silently swallows errors | **FIXED** | Signature now returns `error`; propagation example shown with `fmt.Errorf("record status change: %w", err)`. |
| 2 | Item 8 technical contradiction (`useToast` in non-React module) | **FIXED** | Design now explicitly states `useToast` cannot be used and proposes two viable alternatives (Zustand standalone export or `lib/toast.ts`). |
| 2b | Item 20 offers two alternatives without choosing | **FIXED** | Variadic `userRepo ...repository.UserRepo` chosen, with full code shown. |
| 2c | Item 21 lacks merged function signature | **FIXED** | `buildItemPoolVOs` and `buildProgressRecordVOs` signatures provided with parameter lists. |
| 3 | Security dismissed with "no new attack surface" | **FIXED** | Full threat model added classifying the filter bypass as horizontal privilege escalation, with impact/mitigation/residual risk. |

---

## Scores

| Dimension                          | Score | Max |
|------------------------------------|-------|-----|
| 1. Architecture Clarity            | 17    | 20  |
| 2. Interface & Model Definitions   | 17    | 20  |
| 3. Error Handling                  | 12    | 15  |
| 4. Testing Strategy                | 11    | 15  |
| 5. Breakdown-Readiness (gate)      | 18    | 20  |
| 6. Security Considerations         | 8     | 10  |
| **Total**                          | **83**| **100** |

**Result:** FAIL (83 < 90)

---

## Dimension 1: Architecture Clarity (17/20)

### Layer Placement (6/7)

The "Layer Placement" table names specific file paths for every new helper and justifies each location. Improvements since iteration 1: the `pkg/handler/` sub-package rationale is now explained in the Appendix ("Alternatives Considered" table).

Deduction: `status_history_helper.go` lives in `service/` and references `StatusHistoryRecorder` interface, but the interface definition's package location is not stated. The code block says `type StatusHistoryRecorder interface` with a comment `Record(ctx context.Context, history *model.StatusHistory) error` but the reader must guess whether this interface is defined in `service/`, `repository/`, or `model/`. This matters because it determines the dependency direction.

### Component Diagram (5/7)

The ASCII diagram is present and organized by layers. Issues remaining from iteration 1:

- The Frontend section still has no arrows between `types/index.ts`, `hooks/*.ts`, and `pages/*.tsx`, leaving their dependency relationships ambiguous. A reader cannot tell whether hooks depend on types or vice versa.
- Arrow semantics remain undocumented — no legend explains what `──` vs `│` vs `▼` signify.

### Dependencies (6/6)

External dependencies are now implicitly covered through code snippets (`gorm.DB`, `sql.TxOptions`, `gin.Context`, `strconv`). The Appendix's "Alternatives Considered" section also surfaces dependencies like `dto/` and `handler/` packages. No explicit dependency list section exists, but the document provides enough context for an implementer to infer all required imports.

---

## Dimension 2: Interface & Model Definitions (17/20)

### Interface Signatures (6/7)

Four interfaces are fully typed with Go code. `RecordStatusChange` now returns `error` (iteration 1 attack fixed). Deduction:

- `ResolveBizKey` takes a `lookupFn func(ctx context.Context, bizKey int64) (uint, error)` callback, but the design does not specify what error types the `lookupFn` should return for the not-found case vs. other DB errors. `ResolveBizKey` "responds with appropriate error on parse failure or not-found" but the distinction between 400 (parse) and 404 (not-found) depends on the lookupFn's error type, which is not contractually specified.

### Models Concrete (6/7)

`TeamVO`, `UserVO`, `Role` TableName, and `PermissionData`/`TableRow` are all fully specified with types. Improvements since iteration 1: `TeamVO` RFC3339 formatting is stated; `NewUserVO(u *model.User)` constructor is declared.

Deduction: `NewUserVO` is declared but the field mapping from `model.User` to `UserVO` fields is still not shown. The implementer must guess which `model.User` fields map to which `UserVO` fields. `NewTeamVO` similarly lacks a constructor body — the RFC3339 formatting logic for `CreatedAt`/`UpdatedAt` is described in prose ("Fix createdAt/updatedAt formatting from raw time.Time to RFC3339") but not shown in code.

### Directly Implementable (5/6)

The item-by-item guide provides file names and line numbers for most items. Two items still leave decisions:

- **Item 8** presents two alternative approaches (Zustand standalone export vs. `lib/toast.ts`) with "If the standalone export approach proves too complex..." — the design should pick one approach, not defer the decision to implementation time. Both approaches are valid, but a tech design's purpose is to decide.
- **Item 21** provides the merged function signature but the body is pseudocode comments (`// Batch-resolve submitter names and main-item codes (from itemPoolsToVOs logic)`) rather than concrete implementation. The implementer still needs to figure out the batch resolution logic.

---

## Dimension 3: Error Handling (12/15)

### Error Types (4/5)

The design references existing `apperrors.ErrValidation` and `ErrNotFound`. `RecordStatusChange` now returns `error` (fixed from iteration 1).

Deduction: Item 2's `strconv.ParseInt` failure path now applies `WHERE 1 = 0` instead of silently skipping the filter — this is a significant improvement. However, the design does not define a specific error type for the "invalid assigneeKey format" case. The `WHERE 1=0` is a silent fail-closed that produces no error, no log, and no indication to the caller that input was rejected. The Security section recommends adding an audit log but does not treat it as a required part of the design.

### Propagation Strategy (4/5)

`RecordStatusChange` error propagation is now clearly specified with a code example:

```go
if err := status_history.RecordStatusChange(recorder, ctx, ...); err != nil {
    return fmt.Errorf("record status change: %w", err)
}
```

`ParseBizKeyParam` and `ResolveBizKey` respond directly to Gin context on failure.

Deduction: The propagation strategy for `ResolveBizKey` is incomplete. The doc says it "responds with appropriate error on parse failure or not-found" but does not specify the error wrapping strategy when `lookupFn` returns an unexpected error (e.g., DB connection failure). Should the handler return 500? Should it wrap with `fmt.Errorf`? The design is silent on this path.

### HTTP Status Codes (4/5)

The design now explicitly maps error scenarios: `ParseBizKeyParam` responds with `ErrValidation` (implied 400), and `ResolveBizKey` responds with "appropriate error" for parse failure or not-found. The Security section's threat model clarifies that the filter bypass is an authorization boundary issue.

Deduction: Actual HTTP status codes are still not stated numerically. "ErrValidation" is a Go sentinel error — the reader must trace through the `apperrors` package and the error middleware to find it maps to 400. A simple mapping table (`ErrValidation → 400`, `ErrNotFound → 404`) would eliminate this ambiguity.

---

## Dimension 4: Testing Strategy (11/15)

### Per-Layer Test Plan (4/5)

The testing table covers 7 layers with specific tool names and test function names. Improvements since iteration 1: Test scenarios for Item 2 now include the parse failure case (`WHERE 1 = 0`).

Deduction: The Frontend integration test row still says only "Filter dropdown returns filtered results" — no test setup, mock configuration, or expected behavior detail. The MSW mock setup and specific assertions are unspecified.

### Coverage Target (2/5)

Quote: "Maintain existing coverage. No new coverage targets — this is a cleanup, not a feature addition."

This is still not a numeric coverage target. The rubric explicitly requires a number. The current coverage baseline is not stated, and no minimum threshold is defined. Even for a cleanup, the design should state "current coverage is X%; target: do not drop below X%."

### Test Tooling (5/5)

`go test`, `httptest`, `gorm sqlite`, `vitest`, `MSW`, and `npx tsc --noEmit` are all named. Specific test function names (`TestSubItemService_Assign`, `TestApplyItemFilter_AssigneeKey`, `TestRole_TableName`) are proposed. The test scenarios section provides concrete assertion examples. Full marks.

---

## Dimension 5: Breakdown-Readiness (18/20) — CRITICAL GATE

### Components Enumerable (6/7)

All 24 items are enumerated with file locations. New helper files are named. Improvements since iteration 1: Item 20 now has a single chosen approach (variadic parameter). Item 21 has merged function signatures.

Deduction: Items 16, 21, and 24 still bundle multiple file changes under single item numbers. Item 16 alone covers 22 `String()` removals across 8 files. Item 21 merges 4 functions across 2 handlers. These should be decomposable into sub-tasks for parallel work assignment.

### Tasks Derivable (6/7)

Interfaces map cleanly to implementation items. All iteration 1 blockers resolved: Item 8 has viable approaches, Item 20 has a chosen design, Item 21 has function signatures.

Deduction: Item 8 presents two alternatives without a definitive choice. The implementer must still decide between Zustand standalone export and `lib/toast.ts`. This is a decision the tech design should own.

### PRD AC Coverage (6/6)

All 6 stories map to design items. The PRD Coverage Map is complete with explicit interface/model references. Full marks.

---

## Dimension 6: Security Considerations (8/10)

### Threat Model (4/5)

Major improvement since iteration 1. The threat model now classifies two vulnerabilities:

1. **Item 2 — Assignee filter authorization bypass**: classified as horizontal privilege escalation, HIGH impact, with explicit vulnerability/impact/mitigation/residual-risk structure.
2. **Item 1 — Silent data corruption**: classified as MEDIUM impact.

Deduction: The threat model does not address Item 19's `RENAME TABLE` DDL. Renaming tables on a live database is a potential availability risk — if the migration fails mid-way (e.g., `pmw_roles` renamed but `pmw_role_permissions` not), the application is in a broken state. This was flagged in iteration 1 and remains unaddressed.

### Mitigations (4/5)

The filter bypass mitigation (fail-closed with `WHERE 1=0`) is concrete and well-reasoned. Recommendations section adds audit log and DTO-layer validation as follow-ups.

Deduction: The recommendations are labeled as out-of-scope follow-ups rather than required mitigations. The audit log recommendation is stated as "Consider adding" rather than a hard requirement. For a HIGH-impact authorization bypass, the audit trail should be a required part of the fix, not a nice-to-have. Additionally, `strconv.ParseInt` with base 10 and bitSize 64 has no overflow risk (it returns an error for values outside int64 range), but the design does not explicitly state this analysis.

---

## Top 3 Attack Points

1. **Testing Strategy**: The coverage target is still prose ("Maintain existing coverage. No new coverage targets") rather than a numeric value. The rubric requires a numeric coverage target. The current baseline is not stated. Without a number, there is no pass/fail criterion for test adequacy. Quote: "Maintain existing coverage. No new coverage targets — this is a cleanup, not a feature addition." Must change to: state current coverage percentage (e.g., from `go test -cover`), set a floor (e.g., "must not drop below 75%").

2. **Breakdown-Readiness / Interface & Model**: Item 8 presents two alternative toast implementations without choosing one. Quote: "If the standalone export approach proves too complex, a simpler alternative is to import a `showToast(msg, variant)` function from a new `frontend/src/lib/toast.ts`." A tech design must make the decision, not defer it. Additionally, `NewUserVO` and `NewTeamVO` constructors lack implementation bodies — the field mapping is unspecified. These gaps block clean task decomposition.

3. **Error Handling**: The `ResolveBizKey` function delegates error handling to an unspecified `lookupFn` callback but does not define the error contract for that callback. Quote: "func ResolveBizKey(c *gin.Context, paramName string, lookupFn func(ctx context.Context, bizKey int64) (uint, error)) (uint, bool)" — no specification of what error types `lookupFn` must return for 404 vs 500 scenarios. The HTTP status code mapping remains implicit throughout the document, requiring the reader to trace through `apperrors` to determine the actual status codes.

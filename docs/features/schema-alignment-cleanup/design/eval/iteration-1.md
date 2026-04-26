# Tech Design Evaluation Report — Iteration 1

**Document:** `docs/features/schema-alignment-cleanup/design/tech-design.md`
**Scored:** 2026-04-26
**Target:** 90

---

## Scores

| Dimension                          | Score | Max |
|------------------------------------|-------|-----|
| 1. Architecture Clarity            | 15    | 20  |
| 2. Interface & Model Definitions   | 16    | 20  |
| 3. Error Handling                  | 9     | 15  |
| 4. Testing Strategy                | 10    | 15  |
| 5. Breakdown-Readiness (gate)      | 17    | 20  |
| 6. Security Considerations         | 4     | 10  |
| **Total**                          | **71**| **100** |

**Result:** FAIL (71 < 90)

---

## Dimension 1: Architecture Clarity (15/20)

### Layer Placement (6/7)

The "Layer Placement" table clearly names specific file locations for each new helper. The rationale column explains why each location was chosen. However:

- The design introduces a **new sub-package** `pkg/handler/` without justifying why a new package is needed vs. placing helpers in the existing `handler/` package. The Appendix mentions "mixing concerns" but doesn't define what boundary is being enforced.
- `status_history_helper.go` lives in `service/` but depends on a `StatusHistoryRecorder` interface -- the dependency direction (service -> interface defined where?) is not made explicit.

### Component Diagram (5/7)

An ASCII diagram is present and organized by layers. Issues:

- Arrow semantics are inconsistent: horizontal `──` lines between boxes within the same layer and vertical `│`/`▼` arrows between layers, but the legend is absent.
- The frontend section has no arrows between `types/index.ts`, `hooks/*.ts`, and `pages/*.tsx`, making their relationships ambiguous.
- `filter_helpers.go` appears in the Repository layer box, but the P0 bug fix (Item 2) involves type conversion logic that arguably belongs at a higher level.

### Dependencies (4/6)

Internal modules are named throughout the item-by-item guide. External packages (`gorm.DB`, `sql.TxOptions`, `gin.Context`, `strconv`) appear in code snippets but are never listed explicitly. No dependency section exists. The reader must infer all external dependencies from scattered code fragments.

---

## Dimension 2: Interface & Model Definitions (16/20)

### Interface Signatures (6/7)

Four interfaces are typed with Go code. `DBTransactor`, `ParseBizKeyParam`, `ResolveBizKey`, and `StatusHistoryRecorder` have clear signatures. Deduction for:

- `RecordStatusChange` has **no error return**. For a function that writes to the database, swallowing errors is a design defect not acknowledged. The design says "creates a StatusHistory record if the recorder is non-nil" but never addresses what happens on write failure.

### Models Concrete (5/7)

`TeamVO`, `UserVO`, and `Role` TableName changes are concrete. Deductions:

- `TeamVO.CreatedAt`/`UpdatedAt` are typed as `string` with comment "RFC3339 formatted" but no constructor code is shown. Item 13 says "Fix createdAt/updatedAt formatting from raw time.Time to RFC3339" -- this is prose, not a concrete implementation.
- `TableRow` only shows the one changed field. The full interface context is missing, forcing the implementer to look up the current definition.
- `UserVO.NewUserVO(u *model.User)` constructor is declared but not implemented -- the field mapping from `model.User` to `UserVO` fields is not shown.

### Directly Implementable (5/6)

The item-by-item guide is detailed with file names and line numbers. However, three items leave decisions to the implementer:

- **Item 20**: "Make `NewViewService` accept optional `userRepo` parameter (or use a functional option)" -- two different designs, no decision made.
- **Item 21**: "merge single/batch VO conversion" -- no merged function signature shown.
- **Item 8**: "Import `useToast` or the toast function" -- `client.ts` is a non-React module and cannot use `useToast` (a React hook). The design doesn't resolve this contradiction.

---

## Dimension 3: Error Handling (9/15)

### Error Types (3/5)

The design references existing `apperrors.ErrValidation` and `ErrNotFound`. However:

- `RecordStatusChange` silently swallows all errors. This is not acknowledged as a design choice, let alone justified.
- Item 2's `strconv.ParseInt` conversion silently ignores parse errors with `if err == nil { ... }`. Invalid string input produces no error response and no log -- the filter silently returns unfiltered results, which is the same bug being fixed.

### Propagation Strategy (3/5)

`ParseBizKeyParam` and `ResolveBizKey` respond directly to the Gin context on failure. But:

- The design doesn't describe how errors flow through the new `RecordStatusChange` helper (because they don't -- they're silently dropped).
- No mention of error wrapping or logging strategy for the new shared helpers.

### HTTP Status Codes (3/5)

The design says "Responds with ErrValidation on failure" and "Responds with appropriate error on parse failure or not-found" but never states actual HTTP status codes (400, 404, 500). The reader must trace through `apperrors` package to determine the mapping.

---

## Dimension 4: Testing Strategy (10/15)

### Per-Layer Test Plan (4/5)

A comprehensive table covers 7 layers. Deduction: the Frontend integration test row says only "Filter dropdown returns filtered results" without specifying test setup, mock configuration, or expected behavior detail.

### Coverage Target (2/5)

Quote: "Maintain existing coverage. No new coverage targets -- this is a cleanup, not a feature addition."

This is not a numeric coverage target. The rubric requires a number. No current coverage baseline is stated, nor is a minimum threshold defined.

### Test Tooling (4/5)

`go test`, `httptest`, `gorm sqlite`, `vitest`, `MSW`, and `npx tsc --noEmit` are named. Specific test function names are proposed. Deduction: no Go mocking library is named for service/handler unit tests (hand-rolled structs vs. testify/mock). The design's test code uses assertions but doesn't reference the assertion library.

---

## Dimension 5: Breakdown-Readiness (17/20) -- CRITICAL GATE

### Components Enumerable (6/7)

All 24 items are enumerated with file locations. New helper files are named. Deduction: Items 16, 21, and 24 each bundle multiple file changes under a single item number. Item 16 alone covers 22 changes across 8 files. These should be decomposable into sub-tasks for parallel work.

### Tasks Derivable (5/7)

Interfaces map to implementation items (DBTransactor -> Item 10, ParseBizKeyParam -> Item 12, etc.). Deductions:

- **Item 8** has a technical contradiction: `client.ts` cannot use `useToast` (a React hook) because it's a plain TypeScript module outside a React component tree. The implementer must resolve this.
- **Item 20** offers two alternative implementations without choosing one.
- **Item 21** describes the goal but not the API surface of the merged function.

### PRD AC Coverage (6/6)

All 6 stories map to design items. The PRD Coverage Map is complete. Each PRD acceptance criterion has a corresponding design component with a verification method.

---

## Dimension 6: Security Considerations (4/10)

### Threat Model (2/5)

Quote: "No new attack surface."

This one-liner dismisses security analysis. The design itself identifies that Item 2 was an **authorization bypass** (filter returning all items instead of filtered subset), which is a security vulnerability. A proper threat model should:

1. Classify the filter bypass as an authorization boundary violation.
2. Assess whether the bypass could have been exploited to access data the user shouldn't see.
3. Evaluate whether the `strconv.ParseInt` in the fix introduces new injection or overflow risk.
4. Consider the `RENAME TABLE` DDL in Item 19 as a potential denial-of-service vector if it fails mid-migration.

### Mitigations (2/5)

The only "mitigation" is the bug fix itself. No countermeasures are proposed for:

- **Audit trail gap**: If Item 2's filter bypass was active in production, was sensitive data exposed? No audit recommendation.
- **Silent error swallowing**: `RecordStatusChange` drops errors, which could mask failures in the security audit trail (status history records).
- **ParseInt overflow**: `strconv.ParseInt(*assigneeKey, 10, 64)` with no bounds check or input validation beyond the parse itself.
- **DDL migration risk**: Item 19's `RENAME TABLE` on a live database has no rollback plan.

---

## Top 3 Attack Points

1. **Error Handling**: `RecordStatusChange` silently swallows errors -- the function signature returns void despite performing a database write. Combined with Item 2's filter fix using `if err == nil { ... }` which also silently ignores invalid input, the design has a pattern of dropping errors that contradicts the project's existing error propagation conventions. This must be addressed by adding error returns to `RecordStatusChange` and handling parse failures in Item 2.

2. **Breakdown-Readiness**: Item 8 contains a technical contradiction -- it says to import `useToast` in `client.ts`, but `useToast` is a React hook that can only be used inside a React component. `client.ts` is a plain TypeScript module (the Axios client). The implementer cannot execute this item as written. Item 20 offers two alternative designs without selecting one. Item 21 lacks the merged function signature. These three items block task decomposition.

3. **Security**: The design dismisses security with "No new attack surface" despite documenting a filter bypass that returned all items to unauthorized users. This is an authorization violation that warrants threat classification, impact assessment, and explicit mitigation -- not a one-line dismissal. The `strconv.ParseInt` fix in Item 2 also silently ignores invalid input, which could reproduce the same filter bypass under edge-case inputs.

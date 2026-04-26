# Tech Design Evaluation Report — Iteration 3

**Document:** `docs/features/schema-alignment-cleanup/design/tech-design.md`
**Scored:** 2026-04-26
**Target:** 90

---

## Previous Attack Point Resolution

| # | Attack (Iteration 2) | Status | Evidence |
|---|----------------------|--------|----------|
| 1 | Coverage target is prose, not numeric | **FIXED** | Document now states "Backend: 80.1% statement coverage, Frontend: 92.5% statement coverage" with explicit floors (78% / 90%) and measurement commands. |
| 2 | Item 8 presents two alternatives without choosing; `NewUserVO`/`NewTeamVO` lack constructor bodies | **FIXED** | Item 8 now chooses `lib/toast.ts` shim approach with full rationale. `NewUserVO` and `NewTeamVO` both have complete constructor implementations with field-by-field mapping. |
| 3 | `ResolveBizKey` `lookupFn` error contract unspecified; HTTP status codes remain implicit | **FIXED** | New "ResolveBizKey error contract" table specifies `lookupFn` return conditions, response mapping, and explicit HTTP status codes (404, 500). |

---

## Scores

| Dimension                          | Score | Max |
|------------------------------------|-------|-----|
| 1. Architecture Clarity            | 18    | 20  |
| 2. Interface & Model Definitions   | 19    | 20  |
| 3. Error Handling                  | 14    | 15  |
| 4. Testing Strategy                | 13    | 15  |
| 5. Breakdown-Readiness (gate)      | 18    | 20  |
| 6. Security Considerations         | 9     | 10  |
| **Total**                          | **91**| **100** |

**Result:** PASS (91 >= 90)

---

## Dimension 1: Architecture Clarity (18/20)

### Layer Placement (7/7)

The "Layer Placement" table names specific file paths for every new helper and justifies each location with a rationale. All placements respect the existing layer boundaries. The `pkg/handler/` sub-package rationale is explained in the Appendix "Alternatives Considered" table with a clear "Why Not Chosen" column. `status_history_helper.go` placement in `service/` is correct since it operates on service-layer abstractions. Full marks.

### Component Diagram (5/7)

The ASCII diagram is present and organized by layers with internal component boxes. Improvements since iteration 2: the diagram is unchanged but now more adequate because the rest of the document fills in the gaps.

Remaining issues:
- The Frontend section still has no arrows between `types/index.ts`, `hooks/*.ts`, and `pages/*.tsx`. A reader cannot determine whether hooks depend on types or vice versa, or whether pages import from both independently.
- Arrow semantics remain undocumented — no legend explains what `──` (horizontal), `│` (vertical bar), or `▼` (down arrow) signify in terms of data flow vs. dependency direction.

### Dependencies (6/6)

External dependencies are fully covered through code snippets (`gorm.DB`, `sql.TxOptions`, `gin.Context`, `strconv`, `pkg.FormatID`). The Appendix's "Alternatives Considered" section also clarifies package dependency decisions. Internal package dependencies are traceable through the interface definitions and constructor code. No explicit dependency list section exists, but sufficient context is provided for an implementer.

---

## Dimension 2: Interface & Model Definitions (19/20)

### Interface Signatures (7/7)

All four interfaces are fully typed with Go code. `RecordStatusChange` returns `error` (fixed in iteration 2). `ResolveBizKey` has a complete error contract table specifying `lookupFn` callback behavior for not-found vs. other error cases. `ParseBizKeyParam` clearly documents the `(int64, bool)` return pattern. `StatusHistoryRecorder` is defined with its package context. Full marks.

### Models Concrete (7/7)

`TeamVO` and `UserVO` both have complete constructor implementations with field-by-field mapping from model types. `NewTeamVO` shows `t.CreateTime.Format(time.RFC3339)` and `pkg.FormatID(t.BizKey)` — concrete, directly implementable. `NewUserVO` maps all 6 fields explicitly. `Role` TableName changes and frontend type updates are concrete. Full marks.

### Directly Implementable (5/6)

The item-by-item guide provides file names and line numbers for most items. One minor gap remains:

- **Item 21** `buildItemPoolVOs` and `buildProgressRecordVOs` provide function signatures and empty-slice handling, but the body is still pseudocode comments: `// Batch-resolve submitter names and main-item codes (from itemPoolsToVOs logic)` and `// Batch-resolve author names (from progressRecordsToVOs logic)`. The implementer must read the existing `itemPoolsToVOs` and `progressRecordsToVOs` code and mentally merge it into the new batch function. While the design defers to "existing logic" rather than inventing new logic, a brief pseudocode outline of the batch resolution steps would eliminate ambiguity.

---

## Dimension 3: Error Handling (14/15)

### Error Types (5/5)

The design references existing `apperrors.ErrValidation` and `ErrNotFound` with specific usage contexts. The `WHERE 1=0` fail-closed pattern for invalid `assigneeKey` is a concrete error recovery strategy. `RecordStatusChange` returns `error` with propagation example. Full marks.

### Propagation Strategy (5/5)

`RecordStatusChange` error propagation is clearly specified with a code example showing `fmt.Errorf("record status change: %w", err)` wrapping. `ParseBizKeyParam` and `ResolveBizKey` respond directly to Gin context on failure. The new "ResolveBizKey error contract" table maps `lookupFn` return conditions to handler responses. Full marks.

### HTTP Status Codes (4/5)

The "ResolveBizKey error contract" table now explicitly states HTTP status codes: 404 for not-found, 500 for other errors. `ParseBizKeyParam` maps to HTTP 400 via `ErrValidation`. `RecordStatusChange` errors propagate as 500 (default `RespondError`).

Deduction: The HTTP status code mapping is only explicit for `ResolveBizKey`. For `ParseBizKeyParam`, the reader must still infer that `ErrValidation` maps to 400. For `RecordStatusChange`, the status code depends on how the calling service method handles the error — the design says "propagate it up to the handler layer" but does not state what HTTP status the handler responds with. A consolidated mapping table covering all three helpers would eliminate these gaps.

---

## Dimension 4: Testing Strategy (13/15)

### Per-Layer Test Plan (4/5)

The testing table covers 7 layers with specific tool names and test function names. Test scenarios for Items 1, 2, 10, 16, 17, and 19 are concrete with assertion examples.

Deduction: The Frontend integration test row still says only "Filter dropdown returns filtered results" — no test setup, mock configuration, or expected behavior detail. The MSW mock setup and specific assertions remain unspecified. This was flagged in iteration 1 and iteration 2 and is still not addressed.

### Coverage Target (5/5)

Current baselines are stated numerically: "Backend: 80.1% statement coverage, Frontend: 92.5% statement coverage." Floors are explicit: "must not drop below 78% (backend) or 90% (frontend)." Measurement commands are provided. Full marks.

### Test Tooling (4/5)

`go test`, `httptest`, `gorm sqlite`, `vitest`, `MSW`, and `npx tsc --noEmit` are named. Specific test function names are proposed. Test scenarios section provides concrete assertion examples.

Deduction: No Go mocking library is named for service/handler unit tests. The test code uses hand-rolled mock structs (visible in the project's existing test patterns), but the design does not state whether to continue this pattern or use `testify/mock`. This is a minor gap since the project's existing conventions are consistent, but a first-time contributor would benefit from explicit guidance.

---

## Dimension 5: Breakdown-Readiness (18/20)

### Components Enumerable (6/7)

All 24 items are enumerated with file locations. New helper files are named. The item-by-item guide provides sufficient detail for implementation.

Deduction: Items 16, 21, and 24 still bundle multiple file changes under single item numbers. Item 16 alone covers 22 `String()` removals across 8 files. Item 21 merges 4 functions across 2 handlers. These are not decomposable into sub-tasks for parallel work assignment without the implementer manually splitting them.

### Tasks Derivable (6/7)

All iteration 1 and 2 blockers resolved: Item 8 has a chosen approach (`lib/toast.ts` shim), Item 20 has a single chosen design (variadic parameter), Item 21 has function signatures.

Deduction: Item 21's batch function bodies are pseudocode comments referencing existing logic rather than concrete implementation steps. The implementer must read existing code and mentally merge — this introduces a decision point during task execution that could be eliminated with more detailed pseudocode.

### PRD AC Coverage (6/6)

All 6 stories map to design items with explicit interface/model references. The PRD Coverage Map is complete. Each story's acceptance criteria has a corresponding design component with verification method. Full marks.

---

## Dimension 6: Security Considerations (9/10)

### Threat Model (5/5)

Two security-relevant bugs are classified with structured vulnerability/impact/mitigation/residual-risk analysis:

1. **Item 2 — Assignee filter authorization bypass**: classified as horizontal privilege escalation, HIGH impact. Clear description of the string-to-int mismatch causing the filter to fail open.
2. **Item 1 — Silent data corruption**: classified as MEDIUM impact with clear explanation of why the column name error silently fails.

The threat model also addresses the fail-closed posture as residual risk. Full marks.

### Mitigations (4/5)

The filter bypass mitigation (`WHERE 1=0` fail-closed) is concrete and well-reasoned. The column name fix for Item 1 is straightforward. The Recommendations section adds audit log and DTO-layer validation as follow-ups.

Deduction: The Item 19 `RENAME TABLE` DDL remains unaddressed as an availability risk. Renaming tables on a live database is a potential denial-of-service vector — if the migration fails mid-way (e.g., `pmw_roles` renamed but `pmw_role_permissions` not), the application is in a broken state. This was flagged in iteration 1 and iteration 2 and is still not mentioned. Additionally, the audit log recommendation is still phrased as "Consider adding" rather than required, despite being for a HIGH-impact authorization bypass.

---

## Top 3 Attack Points

1. **Testing Strategy**: The Frontend integration test row remains vague across all three iterations. Quote: "Filter dropdown returns filtered results" — no MSW mock setup, no assertion details, no test scenario structure. This was flagged in iteration 1 and iteration 2 and is unchanged. Additionally, no Go mocking library convention is specified, leaving a gap for contributors unfamiliar with the project's hand-rolled mock pattern. Must add: specific MSW mock configuration, expected API calls, and assertion details for the frontend integration test row.

2. **Security**: Item 19's `RENAME TABLE` DDL availability risk has been flagged in iterations 1, 2, and 3 without being addressed. Quote: "Update both schema files (`SQLite-schema.sql`, `MySql-schema.sql`) with `RENAME TABLE`" — no mention of atomicity, rollback plan, or deployment sequence for the table rename. A mid-failure state leaves the application unable to read either the old or new table names. Must add: migration atomicity strategy (e.g., single transaction, or document that both RENAME statements must be in the same DDL batch).

3. **Architecture Clarity / Breakdown-Readiness**: The component diagram's Frontend section lacks dependency arrows, and Items 16/21/24 remain monolithic. Quote: Frontend section shows `types/index.ts`, `hooks/*.ts`, `pages/*.tsx` as a flat list with no arrows between them. Quote for Item 16: "Remove 22 `String()` calls where the value is already `string`" across 8 files — a single item number covering 22 changes. Must add: dependency arrows in the frontend section, or at minimum a note explaining the relationship; decompose Items 16 and 21 into sub-items or provide a checklist of the 22 locations for Item 16.

---
date: "2026-04-22"
doc_dir: "Z:/project/ai/coding-harness/pm-work-tracker/docs/features/code-quality-cleanup/design/"
iteration: 1
target_score: 80
evaluator: Claude (automated, adversarial)
---

# Design Eval -- Iteration 1

**Score: 75/100** (target: 80)

```
+--------------------------------------------------------------+
|                   DESIGN QUALITY SCORECARD                    |
+------------------------------+----------+----------+----------+
| Dimension                    | Score    | Max      | Status   |
+------------------------------+----------+----------+----------+
| 1. Architecture Clarity      |  20      |  20      | PASS     |
|    Layer placement explicit  |   7/7    |          |          |
|    Component diagram present |   7/7    |          |          |
|    Dependencies listed       |   6/6    |          |          |
+------------------------------+----------+----------+----------+
| 2. Interface & Model Defs    |  15      |  20      | WARN     |
|    Interface signatures typed|   6/7    |          |          |
|    Models concrete           |   5/7    |          |          |
|    Directly implementable    |   4/6    |          |          |
+------------------------------+----------+----------+----------+
| 3. Error Handling            |  10      |  15      | WARN     |
|    Error types defined       |   3/5    |          |          |
|    Propagation strategy clear|   3/5    |          |          |
|    HTTP status codes mapped  |   4/5    |          |          |
+------------------------------+----------+----------+----------+
| 4. Testing Strategy          |   9      |  15      | WARN     |
|    Per-layer test plan       |   2/5    |          |          |
|    Coverage target numeric   |   2/5    |          |          |
|    Test tooling named        |   5/5    |          |          |
+------------------------------+----------+----------+----------+
| 5. Breakdown-Readiness *     |  14      |  20      | PASS     |
|    Components enumerable     |   6/7    |          |          |
|    Tasks derivable           |   5/7    |          |          |
|    PRD AC coverage           |   3/6    |          |          |
+------------------------------+----------+----------+----------+
| 6. Security Considerations   |   7      |  10      | PASS     |
|    Threat model present      |   3/5    |          |          |
|    Mitigations concrete      |   4/5    |          |          |
+------------------------------+----------+----------+----------+
| TOTAL                        |  75      |  100     |          |
+------------------------------+----------+----------+----------+
```

* Breakdown-Readiness >= 12/20 -- can proceed to /breakdown-tasks

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md:157-159 | `useMemberName` shows `{ ... }` as body -- return type of `useCallback` wrapper is not specified, leaving implementer to guess the hook's actual return signature | -1 (Interface signatures) |
| tech-design.md:170-188 | `SubItemVO` shown with `// ... existing fields ...` -- MainItemVO is referenced multiple times but never shown; developer must open source to know which fields exist before adding/removing | -1 (Models concrete) |
| tech-design.md:197-204 | Phase 2 DTOs (`ChangeStatusReq`, `AssignSubItemReq`) are defined but no mapping shows which handlers currently use inline structs that these replace | -1 (Models concrete) |
| tech-design.md:138 | `UpdateFields[T]` accepts `fields map[string]any` -- untyped field map is error-prone; developer must guess valid field names with no compile-time safety | -2 (Directly implementable) |
| tech-design.md:229-234 | Phase 2 partial-batch handling described only in prose ("treats missing IDs as 'unknown'") -- no error type or sentinel value defined for this case | -2 (Error types defined) |
| tech-design.md:229-234 | No explicit propagation strategy for new `ListFiltered`/`SearchAvailable` methods -- SQL failure vs. zero-result paths not distinguished | -2 (Propagation strategy) |
| tech-design.md:239-252 | Testing gates are per-phase, not per-layer -- no distinction between unit tests for service vs. repository vs. handler layer | -3 (Per-layer test plan) |
| tech-design.md:250-251 | "No coverage expansion per PRD scope" -- no numeric coverage target for regression gate; what percentage of existing tests must pass? | -3 (Coverage target numeric) |
| tech-design.md:287-288 | Two unresolved open questions (`onBeforeTerminalStatus` callback design, `linkageMuMap` LRU capacity) block implementation tasks -- developer cannot derive concrete tasks from these | -2 (Tasks derivable) |
| PRD vs design | PRD lists `getItemPoolApi` and `correctCompletionApi` as dead code to delete; neither appears in tech-design.md Phase 1 or api-handbook.md -- PRD acceptance criteria not fully covered | -3 (PRD AC coverage) |
| api-handbook.md:126-128 | `isKeyItem` field listed for removal with reason "Only used in test mocks, never consumed in UI" -- but Phase 3 design references `isKeyItem` as a filter parameter in `applyItemFilter`. Contradiction: is the field being removed or used? | -1 (PRD AC coverage) |

---

## Attack Points

### Attack 1: Testing Strategy -- No per-layer test plan and no regression coverage target

**Where**: tech-design.md lines 239-252: "Per-Phase Testing Gates" table shows only `go test ./...` and `npm test` as the gate for all phases, with the statement "No coverage expansion per PRD scope. Existing tests must continue passing."

**Why it's weak**: The testing section confuses "phase gates" with "layer test plans." A per-layer test plan should specify what kind of tests run at each architectural layer (handler: httptest + mock service; service: mock repo; repository: integration with test DB). Instead, the document only says "all existing tests pass." There is no numeric coverage target -- even `100% of existing tests pass` would be a target, but the document provides no way to measure regression safety. The benchmark example for Phase 2 is good, but Phase 3 (deduplication) and Phase 4 (file splits) have no test strategy beyond "existing tests pass." File splits in particular are high-risk for introducing import breakage, yet no automated check is defined.

**What must improve**: Add a per-layer test plan (handler/service/repository for backend; component/hook/API for frontend). Define a concrete regression target (e.g., "all existing test cases pass, zero test modifications allowed"). For Phase 4 file splits, add a specific verification step (e.g., `npx tsc --noEmit` + runtime smoke test).

### Attack 2: PRD AC Coverage -- Missing dead code items create traceability gap

**Where**: PRD Phase 1 table lists `getItemPoolApi` and `correctCompletionApi` as dead code to delete. Tech-design.md Phase 1 component diagram (lines 43-44) and api-handbook.md "Removed API Functions" table (lines 104-108) list only `archiveMainItemApi` and `assignSubItemApi`.

**Why it's weak**: Two functions called out for deletion in the PRD are simply absent from the design. A developer following the design would not delete these functions, causing a Phase 1 acceptance criteria failure ("0 unused exports/functions/components/types"). This is a direct PRD-to-design traceability gap. Additionally, `isKeyItem` is listed for removal in api-handbook.md but then appears as a filter parameter in Phase 3's `applyItemFilter` -- this contradiction needs resolution.

**What must improve**: Add `getItemPoolApi` and `correctCompletionApi` to the Phase 1 component diagram and api-handbook.md removal table. Resolve the `isKeyItem` contradiction: either it is removed from frontend types (api-handbook says yes) and the backend filter parameter is renamed, or the field is kept and the api-handbook entry is corrected.

### Attack 3: Interface & Model Definitions -- Unresolved open questions block task derivation

**Where**: tech-design.md lines 287-288: two unresolved open questions remain: (1) "SubItem status dropdown 'achievement' dialog: include as optional callback prop or separate wrapper component?" and (2) "`linkageMuMap` LRU capacity: what limit?"

**Why it's weak**: The design recommends answers but does not decide. The `StatusTransitionDropdown` interface (lines 143-154) already includes `onBeforeTerminalStatus` as an optional prop, suggesting the decision is made, but the open question checkbox is still unchecked. This ambiguity means a task for implementing `StatusTransitionDropdown` cannot be fully specified -- a developer would need to ask whether the optional callback is confirmed or still pending. Similarly, the `linkageMuMap` LRU task cannot be estimated or implemented without a capacity number. The `UpdateFields[T]` generic using `map[string]any` also leaves field names untyped.

**What must improve**: Resolve both open questions with checked boxes and explicit decisions. For `UpdateFields`, either define a type-safe alternative (e.g., struct-based field updates) or explicitly document that `map[string]any` is the accepted tradeoff and list the valid field keys in a comment.

---

## Previous Issues Check

*First iteration -- no previous issues.*

---

## Verdict

- **Score**: 75/100
- **Target**: 80/100
- **Gap**: 5 points
- **Breakdown-Readiness**: 14/20 -- can proceed to /breakdown-tasks (>= 12 threshold met)
- **Action**: Continue to iteration 2 to close the 5-point gap. Priority fixes: (1) resolve PRD AC coverage gaps for missing dead code items, (2) add per-layer test plan with numeric regression target, (3) resolve open questions with explicit decisions.

---
date: "2026-04-22"
doc_dir: "Z:/project/ai/coding-harness/pm-work-tracker/docs/features/code-quality-cleanup/design/"
iteration: 2
target_score: 80
evaluator: Claude (automated, adversarial)
---

# Design Eval -- Iteration 2

**Score: 88/100** (target: 80)

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
| 2. Interface & Model Defs    |  16      |  20      | WARN     |
|    Interface signatures typed|   7/7    |          |          |
|    Models concrete           |   5/7    |          |          |
|    Directly implementable    |   4/6    |          |          |
+------------------------------+----------+----------+----------+
| 3. Error Handling            |  13      |  15      | PASS     |
|    Error types defined       |   4/5    |          |          |
|    Propagation strategy clear|   4/5    |          |          |
|    HTTP status codes mapped  |   5/5    |          |          |
+------------------------------+----------+----------+----------+
| 4. Testing Strategy          |  15      |  15      | PASS     |
|    Per-layer test plan       |   5/5    |          |          |
|    Coverage target numeric   |   5/5    |          |          |
|    Test tooling named        |   5/5    |          |          |
+------------------------------+----------+----------+----------+
| 5. Breakdown-Readiness *     |  17      |  20      | PASS     |
|    Components enumerable     |   7/7    |          |          |
|    Tasks derivable           |   6/7    |          |          |
|    PRD AC coverage           |   4/6    |          |          |
+------------------------------+----------+----------+----------+
| 6. Security Considerations   |   7      |  10      | PASS     |
|    Threat model present      |   3/5    |          |          |
|    Mitigations concrete      |   4/5    |          |          |
+------------------------------+----------+----------+----------+
| TOTAL                        |  88      |  100     |          |
+------------------------------+----------+----------+----------+
```

* Breakdown-Readiness >= 12/20 -- can proceed to /breakdown-tasks

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md:183 | `SubItemVO` shows `// ... existing fields ...` -- MainItemVO is referenced multiple times (lines 39, 94, 184) but never shown; developer must open source to know which fields exist before adding/removing | -2 (Models concrete) |
| tech-design.md:169 | `useMemberName` shows `{ ... }` as body -- return type of `useCallback` wrapper is `(assigneeId: number \| null) => string` but never stated explicitly, forcing implementer to infer | -1 (Directly implementable) |
| tech-design.md:149 | `UpdateFields[T]` accepts `fields map[string]any` -- despite documented accepted keys (lines 141-148), there is no compile-time enforcement; a typo in a field key is a runtime error, not a compile error | -1 (Directly implementable) |
| tech-design.md:247 | Phase 2 partial-batch handling described only in prose ("treats missing IDs as 'unknown'") -- no error type or sentinel value defined for this case | -1 (Error types defined) |
| tech-design.md:247 | No explicit propagation strategy distinguishing SQL failure vs. zero-result vs. partial-batch for new `ListFiltered`/`SearchAvailable` methods | -1 (Propagation strategy) |
| PRD line 155 vs api-handbook.md:114-117 | PRD lists `TeamMember` type for removal alongside `WeeklyViewResp`, `WeeklyGroup`, `SubItemWithProgress`; api-handbook.md "Removed Types" table lists only three types, omitting `TeamMember` | -1 (PRD AC coverage) |
| PRD line 157 vs design | PRD lists `ApiSuccessEnvelope`, `ApiErrorEnvelope` in `client.ts` as "Remove unused exports" -- neither appears in tech-design.md Phase 1 or api-handbook.md | -1 (PRD AC coverage) |
| PRD line 169 vs design | PRD lists `setAuth` redundant expression fix (`user?.isSuperAdmin ?? user?.isSuperAdmin`); not mentioned in tech-design.md or api-handbook.md | -1 (Tasks derivable) |
| tech-design.md:294-300 | Threat model correctly identifies no new endpoints/dependencies/auth changes, but does not address that Phase 2 introduces new SQL query patterns with user-supplied `search` strings (`ListFiltered`, `SearchAvailable`) that merit explicit verification of parameterized query usage | -2 (Threat model) |
| tech-design.md:303-306 | Mitigation for SQL pushdown says "parameterized queries via GORM (not string concatenation) -- existing pattern, no change needed" but does not confirm this for the specific new methods | -1 (Mitigations concrete) |

---

## Attack Points

### Attack 1: Interface & Model Definitions -- Incomplete model definitions force source-code lookups

**Where**: tech-design.md line 183: `SubItemVO` shows `// ... existing fields ...` with only the new `StatusName` field visible. `MainItemVO` is referenced at lines 39, 94, 184 but never defined anywhere in the design. tech-design.md line 169: `useMemberName` shows `{ ... }` as the hook body.

**Why it's weak**: The document tells a developer to add `StatusName` to `SubItemVO` for "parity with MainItemVO" but does not show what `MainItemVO` looks like. To know which fields `SubItemVO` should have after modification, the developer must open `vo/item_vo.go`. The `useMemberName` hook leaves the return type implicit -- the developer must infer it is `(assigneeId: number | null) => string` from the callback name and parameter. These gaps mean the design is not self-contained; a developer cannot implement from it alone.

**What must improve**: Show the complete `SubItemVO` struct with all fields (existing + new). Show the complete `MainItemVO` struct at least once. Explicitly type the `useMemberName` return value: `function useMemberName(members: TeamMember[]): (assigneeId: number | null) => string`.

### Attack 2: Breakdown-Readiness -- PRD-to-design traceability gaps for dead code items

**Where**: PRD line 155 lists `TeamMember` type for removal alongside `WeeklyViewResp`, `WeeklyGroup`, `SubItemWithProgress`. PRD line 157 lists `ApiSuccessEnvelope`, `ApiErrorEnvelope` in `client.ts` as "Remove unused exports." PRD line 169 lists `setAuth` redundant expression fix. None of these appear in tech-design.md or api-handbook.md.

**Why it's weak**: A developer creating tasks from this design would produce tasks for the items listed in the Phase 1 diagram and api-handbook removal tables. They would not produce tasks for removing `TeamMember`, `ApiSuccessEnvelope`, `ApiErrorEnvelope`, or fixing `setAuth`. These items would be missed, leading to incomplete Phase 1 delivery against the PRD's quantified objective "0 unused exports/functions/components/types" (PRD line 45).

**What must improve**: Add `TeamMember` to the api-handbook.md "Removed Types" table. Add `ApiSuccessEnvelope`/`ApiErrorEnvelope` removal to either tech-design.md Phase 1 diagram or api-handbook.md (new "Removed Exports" section). Add `setAuth` fix to the Phase 1 contract mismatch section in tech-design.md.

### Attack 3: Error Handling -- Phase 2 batch error handling defined only in prose

**Where**: tech-design.md line 247: "the service layer treats missing IDs as 'unknown' rather than failing -- matching current behavior where individual lookups silently handle not-found." No error type, sentinel value, or explicit propagation path is defined.

**Why it's weak**: The statement describes a behavior but does not define it as a contract. If a batch `FindByIDs` returns 8 out of 10 requested IDs, there is no defined mechanism for the caller to know which 2 are missing. Is it a logged warning? A partial-result wrapper? A map with nil values? The propagation strategy for the new `ListFiltered` and `SearchAvailable` methods does not distinguish between "SQL connection failed," "zero results matched," and "partial results returned." These are different failure modes that affect handler response behavior.

**What must improve**: Define the batch query contract explicitly -- e.g., "FindByIDs returns a map; missing IDs are simply absent from the map, and callers check `map[id]` existence." Add a sentence to Error Handling distinguishing the three failure paths (SQL error propagates as error, zero results returns empty slice/map, partial results return what was found).

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Testing Strategy -- No per-layer test plan | Yes | tech-design.md lines 252-260 now provide a per-layer test plan table with 6 rows covering backend handler/service/repository and frontend component/hook/API module layers |
| Attack 1: Testing Strategy -- No regression coverage target | Yes | tech-design.md lines 263-264 state "100% of existing test cases must pass with zero test file modifications" with explicit regression gate |
| Attack 2: PRD AC -- Missing `getItemPoolApi` and `correctCompletionApi` | Yes | Both now appear in tech-design.md Phase 1 diagram (line 44) and api-handbook.md removal table (lines 108-109) |
| Attack 2: PRD AC -- `isKeyItem` contradiction | Yes | api-handbook.md line 126 now clarifies: "Removed from frontend type only -- backend retains `is_key_item` DB column and repository filter; frontend never sends or consumes this field." tech-design.md lines 220-222 also add a comment: "Backend-only: isKeyItem filter operates on the DB column; frontend type does not expose this field (removed in Phase 1)." |
| Attack 3: Interface & Model -- Unresolved open questions | Yes | All three open questions now have checked boxes with explicit decisions (tech-design.md lines 311-313): optional callback prop confirmed, LRU capacity set to 1000, hybrid SQL strategy confirmed |
| Attack 3: Interface & Model -- Untyped `UpdateFields` | Partially | `UpdateFields[T]` still uses `map[string]any` (line 149) but now has documented accepted keys per entity (lines 141-148) with runtime validation returning `ErrInvalidField`. Compile-time safety is not achieved but runtime validation is now specified -- this is a documented tradeoff. |

---

## Verdict

- **Score**: 88/100
- **Target**: 80/100
- **Gap**: 0 points (target exceeded by 8)
- **Breakdown-Readiness**: 17/20 -- can proceed to /breakdown-tasks (>= 12 threshold met)
- **Action**: Target reached. All three iteration-1 attacks have been substantially addressed. Remaining deductions are minor (incomplete model definitions, small PRD traceability gaps, prose-only error handling for batch queries). Can proceed to `/breakdown-tasks`.

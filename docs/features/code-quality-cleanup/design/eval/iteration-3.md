---
date: "2026-04-22"
doc_dir: "Z:/project/ai/coding-harness/pm-work-tracker/docs/features/code-quality-cleanup/design/"
iteration: 3
target_score: 80
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 3

**Score: 86/100** (target: 80)

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
|    Interface signatures typed|   7/7    |          |          |
|    Models concrete           |   4/7    |          |          |
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
| 5. Breakdown-Readiness *     |  16      |  20      | PASS     |
|    Components enumerable     |   7/7    |          |          |
|    Tasks derivable           |   4/7    |          |          |
|    PRD AC coverage           |   5/6    |          |          |
+------------------------------+----------+----------+----------+
| 6. Security Considerations   |   7      |  10      | PASS     |
|    Threat model present      |   3/5    |          |          |
|    Mitigations concrete      |   4/5    |          |          |
+------------------------------+----------+----------+----------+
| TOTAL                        |  86      |  100     |          |
+------------------------------+----------+----------+----------+
```

* Breakdown-Readiness >= 12/20 -- can proceed to /breakdown-tasks

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md:183 | `SubItemVO` shows `// ... existing fields ...` — only the new `StatusName` field is visible; `MainItemVO` is referenced for "parity" at line 184 but never defined anywhere in either design document; developer cannot know which fields SubItemVO should retain | -2 (Models concrete) |
| tech-design.md:191,197 | `MainItem` and `SubItem` TypeScript interfaces both show `// ... existing fields ...` with only `REMOVED` annotations; developer cannot see what fields survive after cleanup | -1 (Models concrete) |
| tech-design.md:149 | `UpdateFields[T]` accepts `fields map[string]any` — despite documented accepted keys (lines 141-148), there is no compile-time enforcement; a typo in a field key is a runtime error | -1 (Directly implementable) |
| tech-design.md:169 | `useMemberName` shows `{ ... }` as body — return type of `useCallback` wrapper is `(assigneeId: number \| null) => string` but never stated explicitly, forcing implementer to infer | -1 (Directly implementable) |
| tech-design.md:247 | Phase 2 partial-batch handling described only in prose ("treats missing IDs as 'unknown'") — no explicit data structure contract (e.g., "returns a map; missing IDs are absent from the map") | -1 (Error types defined) |
| tech-design.md:247 | No explicit propagation strategy distinguishing SQL failure vs. zero-result vs. partial-batch for new `ListFiltered`/`SearchAvailable` methods | -1 (Propagation strategy) |
| PRD line 155 vs api-handbook.md:114-117 | PRD lists `TeamMember` type for removal alongside `WeeklyViewResp`, `WeeklyGroup`, `SubItemWithProgress`; api-handbook.md "Removed Types" table lists only three types, omitting `TeamMember` | -1 (PRD AC coverage) |
| PRD line 157 vs design | PRD lists `ApiSuccessEnvelope`, `ApiErrorEnvelope` in `client.ts` as "Remove unused exports" — neither appears in tech-design.md Phase 1 or api-handbook.md | -1 (Tasks derivable) |
| PRD line 169 vs design | PRD lists `setAuth` redundant expression fix (`user?.isSuperAdmin ?? user?.isSuperAdmin`); not mentioned in tech-design.md or api-handbook.md | -1 (Tasks derivable) |
| tech-design.md:294-300 | Threat model correctly identifies no new endpoints/dependencies/auth changes but does not address that Phase 2 introduces new SQL query patterns with user-supplied `search` strings (`ListFiltered`, `SearchAvailable`) | -2 (Threat model) |
| tech-design.md:303-306 | Mitigation for SQL pushdown says "parameterized queries via GORM — existing pattern, no change needed" but does not confirm this for the specific new methods (`ListFiltered`, `SearchAvailable`) | -1 (Mitigations concrete) |

---

## Attack Points

### Attack 1: Interface & Model Definitions — Incomplete model definitions persist across all three iterations

**Where**: tech-design.md line 183: `SubItemVO` shows `// ... existing fields ...` with only the new `StatusName` field visible. Lines 191 and 197: `MainItem` and `SubItem` TypeScript interfaces also show `// ... existing fields ...`. Line 184 references `MainItemVO` for parity but never defines it. Line 169: `useMemberName` shows `{ ... }` as the hook body.

**Why it's weak**: This is the same attack raised in iterations 1 and 2. A developer told to "add `StatusName` to `SubItemVO` for parity with `MainItemVO`" must open `vo/item_vo.go` to see what `SubItemVO` currently contains, and must open the same file (or a sibling) to see what `MainItemVO` looks like. The document is not self-contained for these models. The `useMemberName` hook leaves the return type implicit — the developer must infer `(assigneeId: number | null) => string` from the callback name and parameter. These are not edge cases; they are the primary data shapes being modified in Phase 1 and Phase 3.

**What must improve**: Show the complete `SubItemVO` struct with all fields (existing + new). Show the complete `MainItemVO` struct at least once. Explicitly type the `useMemberName` return value: `function useMemberName(members: TeamMember[]): (assigneeId: number | null) => string`.

### Attack 2: Breakdown-Readiness — Three PRD items remain untraceable in the design

**Where**: PRD line 155 lists `TeamMember` type for removal; PRD line 157 lists `ApiSuccessEnvelope`, `ApiErrorEnvelope` removal from `client.ts`; PRD line 169 lists `setAuth` redundant expression fix. None of these appear in tech-design.md or api-handbook.md.

**Why it's weak**: This is the same traceability gap raised as Attack 2 in iteration 2. A developer creating tasks from the design documents would produce tasks for the items listed in the Phase 1 component diagram and api-handbook removal tables. They would NOT produce tasks for removing `TeamMember` (from types/index.ts), removing `ApiSuccessEnvelope`/`ApiErrorEnvelope` (from client.ts), or fixing `setAuth`. Three items would be silently missed, leading to incomplete Phase 1 delivery against the PRD's acceptance criterion "0 unused exports/functions/components/types" (PRD line 45) and the contract mismatch fix for `setAuth` (PRD line 169).

**What must improve**: Add `TeamMember` to the api-handbook.md "Removed Types" table. Add `ApiSuccessEnvelope`/`ApiErrorEnvelope` to either the api-handbook.md (new "Removed Exports" section) or the tech-design.md Phase 1 diagram. Add `setAuth` redundant expression fix to the tech-design.md Phase 1 contract mismatch fixes.

### Attack 3: Error Handling — Phase 2 batch query contract remains prose-only

**Where**: tech-design.md line 247: "the service layer treats missing IDs as 'unknown' rather than failing — matching current behavior where individual lookups silently handle not-found." No error type, sentinel value, or data structure contract is defined.

**Why it's weak**: This is the same attack raised in iterations 1 and 2. The statement describes a behavior but does not define it as a contract. If `FindByIDs` is asked for 10 IDs and 8 exist in the database, there is no defined mechanism for the caller to discover which 2 are missing. The prose says "unknown" but does not specify the mechanism: is it a `map[uint]*T` where missing keys are simply absent? A slice with `nil` entries? A separate `MissingIDs []uint` return value? The `ListFiltered` and `SearchAvailable` methods also lack explicit failure path distinction (SQL connection error vs. zero results vs. partial results). These are different failure modes that would trigger different handler responses.

**What must improve**: Define the batch query return contract explicitly — e.g., "FindByIDs returns `map[uint]*T`; IDs not found in the database are simply absent from the map. Callers check `_, ok := map[id]` to detect missing entries." Add one sentence distinguishing the three failure paths: SQL error propagates as error, zero results returns empty slice/map, partial results return what was found.

---

## Previous Issues Check

| Previous Attack (Iteration 2) | Addressed? | Evidence |
|-------------------------------|------------|----------|
| Attack 1: Interface & Model — Incomplete model definitions with `// ... existing fields ...` placeholders | No | tech-design.md lines 183, 191, 197 still show `// ... existing fields ...`. `MainItemVO` is still referenced but never defined. `useMemberName` still shows `{ ... }` body (line 169). |
| Attack 2: Breakdown-Readiness — Missing traceability for `TeamMember`, `ApiSuccessEnvelope`, `ApiErrorEnvelope`, `setAuth` | No | Grep confirms: `TeamMember` is not in api-handbook.md "Removed Types" table. `ApiSuccessEnvelope`/`ApiErrorEnvelope` do not appear in tech-design.md or api-handbook.md. `setAuth` does not appear in tech-design.md or api-handbook.md. |
| Attack 3: Error Handling — Batch error handling defined only in prose, no explicit error type | No | tech-design.md line 247 still reads "the service layer treats missing IDs as 'unknown' rather than failing" with no data structure contract or explicit propagation path defined. |

---

## Verdict

- **Score**: 86/100
- **Target**: 80/100
- **Gap**: 0 points (target exceeded by 6)
- **Breakdown-Readiness**: 16/20 -- can proceed to /breakdown-tasks (>= 12 threshold met)
- **Action**: Target reached. Score dropped from 88 to 86 due to slightly stricter scoring on Models concrete (the incomplete model definitions are a persistent weakness that a fresh evaluation cannot ignore). All three iteration-2 attacks remain unaddressed, but the document still clears the 80-point target. Can proceed to `/breakdown-tasks`.

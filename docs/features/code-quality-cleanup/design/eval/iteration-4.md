---
date: "2026-04-22"
doc_dir: "Z:/project/ai/coding-harness/pm-work-tracker/docs/features/code-quality-cleanup/design/"
iteration: 4
target_score: 80
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 4

**Score: 93/100** (target: 80)

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
| 2. Interface & Model Defs    |  18      |  20      | PASS     |
|    Interface signatures typed|   7/7    |          |          |
|    Models concrete           |   7/7    |          |          |
|    Directly implementable    |   5/6    |          |          |
+------------------------------+----------+----------+----------+
| 3. Error Handling            |  14      |  15      | PASS     |
|    Error types defined       |   5/5    |          |          |
|    Propagation strategy clear|   5/5    |          |          |
|    HTTP status codes mapped  |   4/5    |          |          |
+------------------------------+----------+----------+----------+
| 4. Testing Strategy          |  15      |  15      | PASS     |
|    Per-layer test plan       |   5/5    |          |          |
|    Coverage target numeric   |   5/5    |          |          |
|    Test tooling named        |   5/5    |          |          |
+------------------------------+----------+----------+----------+
| 5. Breakdown-Readiness *     |  18      |  20      | PASS     |
|    Components enumerable     |   7/7    |          |          |
|    Tasks derivable           |   6/7    |          |          |
|    PRD AC coverage           |   5/6    |          |          |
+------------------------------+----------+----------+----------+
| 6. Security Considerations   |   8      |  10      | PASS     |
|    Threat model present      |   4/5    |          |          |
|    Mitigations concrete      |   4/5    |          |          |
+------------------------------+----------+----------+----------+
| TOTAL                        |  93      |  100     |          |
+------------------------------+----------+----------+----------+
```

* Breakdown-Readiness >= 12/20 -- can proceed to /breakdown-tasks

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md:152 | `UpdateFields[T]` uses `map[string]any` with no compile-time key enforcement; runtime validation via `ErrInvalidField` is documented but a typo in a field key is only caught at runtime, not at compile time | -1 (Directly implementable) |
| tech-design.md:80 | `MemberSelect` component listed in Phase 3 component diagram but never defined with its props interface — implementer must guess the shape | -1 (Directly implementable) |
| tech-design.md:326-329 | Error handling references `apperrors.RespondError` / `RespondOK` patterns but never explicitly maps error types to HTTP status codes (e.g., `MapNotFound` -> 404); relies on existing convention without restating it | -1 (HTTP status codes mapped) |
| tech-design.md / api-handbook.md | PRD line 158 lists `addToast` destructuring removal from ItemViewPage and MainItemDetailPage; this item appears in neither tech-design.md nor api-handbook.md | -1 (Tasks derivable) |
| tech-design.md:396-399 | Threat model says "no new attack surface" but Phase 2 `ListFiltered` and `SearchAvailable` accept user-supplied `search` strings — this is a new input vector that should be called out even if mitigated by parameterized queries | -1 (Threat model present) |
| tech-design.md:405 | Mitigation says "parameterized queries via GORM" for Phase 2 SQL pushdown but does not specify input validation constraints for the `search` parameter (max length, character restrictions) in `ListFiltered` and `SearchAvailable` | -1 (Mitigations concrete) |

---

## Attack Points

### Attack 1: Interface & Model Definitions — `MemberSelect` component listed but never defined

**Where**: tech-design.md line 80: `New: components/shared/MemberSelect.tsx` appears in the Phase 3 component diagram. PRD line 240 states: "Member select repeated in 8+ dialogs" with extraction to `MemberSelect` component. The component is never defined anywhere in either design document — no props interface, no type signatures, no usage example.

**Why it's weak**: This is a shared component that replaces duplicated member selection logic in 8+ dialogs across the codebase. Without a defined props interface, an implementer cannot know what the component receives (member list? team ID for lazy loading? selection callback? multi-select support?). This is the same category of gap that existed for `useMemberName` in earlier iterations. The PRD calls it out explicitly as a deduplication target, and the Phase 3 diagram names it, but the design stops short of defining it.

**What must improve**: Add a `MemberSelectProps` interface definition (similar to how `StatusTransitionDropdownProps` is defined at lines 158-168), including at minimum: `members` array prop, `selectedId` / `onSelect` callback, and any dialog integration props needed.

### Attack 2: Breakdown-Readiness — `addToast` destructuring removal untraceable in design

**Where**: PRD line 158 lists `addToast` destructuring removal from `ItemViewPage` and `MainItemDetailPage` as a dead code item in Phase 1. Neither tech-design.md nor api-handbook.md mentions this item. The Phase 1 component diagram (lines 40-50) does not list it. The api-handbook.md removal tables (lines 103-119) do not include it.

**Why it's weak**: This is a PRD-defined dead code removal item that would be silently missed during task breakdown. A developer creating tasks from the design documents alone would not produce a task for removing unused `addToast` destructuring. The PRD's Phase 1 acceptance criterion "0 unused exports/functions/components/types" (line 45) implicitly requires this item, but the design does not trace it.

**What must improve**: Add `addToast` destructuring removal to the Phase 1 component diagram or api-handbook.md dead code removal section. This is a minor item but it represents a traceability gap between PRD and design.

### Attack 3: Security Considerations — `search` parameter input validation undefined for SQL pushdown methods

**Where**: tech-design.md line 131: `ListFiltered(ctx context.Context, search string, offset, limit int)` and line 133: `SearchAvailable(ctx context.Context, teamID uint, search string, limit int)`. Both accept a `search` string that flows into SQL queries. The threat model (line 396) states "no new attack surface" and the mitigation (line 405) says "parameterized queries via GORM — existing pattern, no change needed."

**Why it's weak**: While GORM parameterization prevents SQL injection, there is no mention of input validation for the `search` parameter. Without constraints, a caller could pass an arbitrarily long string (potentially causing performance issues in `LIKE` queries) or characters that cause unexpected behavior in pattern matching (e.g., `%`, `_` wildcards in LIKE clauses). The threat model should acknowledge the `search` input vector, and the mitigation should include at least a max-length constraint and wildcard escaping or a note that GORM's `Where("field LIKE ?", "%"+search+"%")` handles this safely.

**What must improve**: Add a sentence to the threat model acknowledging the `search` parameter as a new input vector in Phase 2. Add a concrete input validation requirement: e.g., "search strings are capped at 100 characters" or "LIKE wildcards in search input are escaped before query construction."

---

## Previous Issues Check

| Previous Attack (Iteration 3) | Addressed? | Evidence |
|-------------------------------|------------|----------|
| Attack 1: Interface & Model — Incomplete model definitions with `// ... existing fields ...` placeholders | Yes | tech-design.md lines 189-229 now show complete SubItemVO and MainItemVO structs with all fields. TypeScript interfaces (lines 233-276) also fully defined. `useMemberName` (lines 172-178) has explicit return type and complete implementation. |
| Attack 2: Breakdown-Readiness — Missing traceability for `TeamMember`, `ApiSuccessEnvelope`, `ApiErrorEnvelope`, `setAuth` | Yes | Phase 1 diagram (lines 45-49) now includes all three items. api-handbook.md lines 116-119 include `ApiSuccessEnvelope`, `ApiErrorEnvelope` in "Removed Exports" and `TeamMember` in "Removed Types". tech-design.md lines 282-286 define the `setAuth` fix. |
| Attack 3: Error Handling — Batch query contract prose-only, no return type or failure path distinctions | Yes | tech-design.md lines 333-346 now define explicit "Batch Query Contract" with return type `map[uint]*T`, a table covering all 4 conditions (SQL error, empty input, zero results, partial results), and caller pattern `entry, ok := result[id]`. |

---

## Verdict

- **Score**: 93/100
- **Target**: 80/100
- **Gap**: 0 points (target exceeded by 13)
- **Breakdown-Readiness**: 18/20 -- can proceed to /breakdown-tasks (>= 12 threshold met)
- **Action**: Target reached. All three iteration-3 attacks are fully addressed. Remaining deductions are minor: an undefined `MemberSelect` component interface, an untraced `addToast` PRD item, and soft input validation for SQL search parameters. None are blocking. Can proceed to `/breakdown-tasks`.

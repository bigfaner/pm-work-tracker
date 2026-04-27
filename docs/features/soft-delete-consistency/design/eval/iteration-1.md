---
date: "2026-04-27"
doc_dir: "docs/features/soft-delete-consistency/design/"
iteration: 1
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 1

**Score: 68/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  17      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  6/7     │          │            │
│    Dependencies listed       │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼──────────┤
│ 2. Interface & Model Defs    │  14      │  20      │ ⚠️         │
│    Interface signatures typed│  6/7     │          │            │
│    Models concrete           │  3/7     │          │            │
│    Directly implementable    │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┤
│ 3. Error Handling            │  8       │  15      │ ⚠️         │
│    Error types defined       │  3/5     │          │            │
│    Propagation strategy clear│  3/5     │          │            │
│    HTTP status codes mapped  │  2/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  11      │  15      │ ✅         │
│    Per-layer test plan       │  4/5     │          │            │
│    Coverage target numeric   │  4/5     │          │            │
│    Test tooling named        │  3/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  18      │  20      │ ✅         │
│    Components enumerable     │  7/7     │          │            │
│    Tasks derivable           │  6/7     │          │            │
│    PRD AC coverage           │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  0       │  10      │ ❌         │
│    Threat model present      │  0/5     │          │            │
│    Mitigations concrete      │  0/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  68      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness: 18/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Architecture: Dependencies | No internal module paths or external package versions listed; "No new external dependencies" is not the same as listing existing ones | -2 pts |
| Architecture: Diagram | Diagram does not show `role_repo` join-query relationship or the model types that flow through each component | -1 pt |
| Interfaces: Models | "No new models" section has zero field-level detail; the only model shown is a SQL index change, no BaseModel field layout or type definition | -4 pts |
| Interfaces: Signatures | `FindByIDs` body is replaced with a comment `// ... same pattern` instead of full typed code | -1 pt |
| Error: Types defined | No custom error types defined; only references `ErrNotFound` without showing its definition or constructor | -2 pts |
| Error: Propagation | States "Existing error propagation is unchanged" but does not show the actual propagation path (repo -> service -> handler) | -2 pts |
| Error: HTTP status | Claims "Single-layer feature (repository only). Not applicable." but the PRD explicitly mentions "clicking bizKey returns 404" — the design should map how repository ErrNotFound reaches HTTP 404 | -3 pts |
| Testing: Per-layer | Only repository layer is tested; no mention of integration or regression testing at service/handler level even though PRD scope includes "全量回归测试" | -1 pt |
| Testing: Coverage numeric | States "100% of modified methods" which is a method-count target, not a line/branch coverage percentage | -1 pt |
| Testing: Tooling | Names "SQLite in-memory + testify" but does not specify the testify package (assert? require? mock?), the test runner, or the setup helper library | -2 pts |
| Breakdown: Tasks derivable | The "Complete Change List" maps files to methods but does not decompose into discrete, estimable tasks with dependencies | -1 pt |
| Breakdown: PRD AC coverage | PRD Story 1 AC mentions "clicking bizKey returns 404" but the design maps it only to `FindByBizKey() + NotDeleted` without explaining how the 404 HTTP status is produced at the handler layer | -1 pt |
| Security: Threat model | The security section has a single vague sentence: "Deleted roles appearing in permission checks could grant or deny access incorrectly." This is not a structured threat model — no attack vectors, no STRIDE analysis, no data sensitivity classification | -5 pts |
| Security: Mitigations | "This fix IS the mitigation" is circular reasoning, not a concrete countermeasure. No verification steps, no test for the security property, no discussion of what happens if a new repo method is added without NotDeleted | -5 pts |

---

## Attack Points

### Attack 1: Security — threat model is dangerously shallow

**Where**: "Deleted roles appearing in permission checks could grant or deny access incorrectly." and "This fix IS the mitigation — ensuring deleted entities are excluded from all queries, including permission-related join queries."

**Why it's weak**: The PRD explicitly identifies this as a security fix ("本修复提升安全性：已删除的角色不再出现在权限检查中，防止权限泄露"). Yet the design's entire security section is two sentences of circular reasoning. There is no threat model — no enumeration of attack vectors (e.g., stale permission cache, race condition between delete and permission check, direct API call bypass), no classification of data sensitivity, no discussion of what happens if a developer forgets to add NotDeleted to a future repo method. The "mitigation" is the feature itself, which is tautological.

**What must improve**: Add a structured threat model with at minimum: (1) specific threat scenarios (stale session after role deletion, TOCTOU between delete and permission check, new repo method added without NotDeleted), (2) concrete mitigations per threat (e.g., middleware re-checks, lint rule for NotDeleted, integration test), (3) how to verify the security property holds (security-focused test cases).

### Attack 2: Error Handling — no HTTP status mapping despite PRD requiring 404 behavior

**Where**: "Single-layer feature (repository only). Not applicable." in the Cross-Layer Data Map section, and the PRD AC: "该角色不在列表中，点击其 bizKey 返回 404"

**Why it's weak**: The design explicitly claims HTTP status mapping is "Not applicable" because this is a single-layer change. But the PRD's Story 1 AC requires a specific HTTP behavior — "返回 404". The design should explain how `ErrNotFound` from the repository propagates through the existing service/handler stack to produce a 404 response. Without this traceability, a developer implementing this has no confirmation that the existing error propagation actually produces the correct HTTP status. The design also references `ErrNotFound` without defining its type, constructor, or how it differs from a raw GORM `ErrRecordNotFound`.

**What must improve**: (1) Define `ErrNotFound` with its type signature and show the existing propagation path from repo -> service -> handler -> HTTP response code. (2) Explicitly confirm that existing handler error mapping converts `ErrNotFound` to HTTP 404. (3) Remove the "Not applicable" dismissal and show the cross-layer error trace even if no new code is needed.

### Attack 3: Interface & Models — no concrete model definitions, incomplete code

**Where**: "No new models. One schema change:" and `// ... same pattern: if isSoftDeletable[T]() { add NotDeleted }` in the FindByIDs interface section.

**Why it's weak**: The "Data Models" section contains zero field-level type definitions. The `BaseModel` struct (which is central to this entire fix) is never shown — its fields (`deleted_flag`, `deleted_time`) are referenced throughout but never defined with types and constraints. The `FindByIDs` interface shows a comment stub `// ... same pattern` instead of actual typed code. A developer cannot implement `FindByIDs` from this document without reading the existing source code, which defeats the purpose of the design doc. The PRD spec actually has more concrete interface signatures than this design document.

**What must improve**: (1) Add a BaseModel definition section showing all fields with types and GORM tags. (2) Show the full typed implementation of `FindByIDs[T]`, not a comment stub. (3) Consider that the design doc should be at least as detailed as the PRD spec it references — currently the PRD has more complete method tables.

---

## Previous Issues Check

<!-- First iteration — no previous issues -->

---

## Verdict

- **Score**: 68/100
- **Target**: 90/100
- **Gap**: 22 points
- **Breakdown-Readiness**: 18/20 — can proceed to `/breakdown-tasks`
- **Action**: Continue to iteration 2 — biggest gains available in Security (0→10), Error Handling (8→15), and Interface & Models (14→20)

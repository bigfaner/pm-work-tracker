---
date: "2026-04-27"
doc_dir: "docs/features/soft-delete-consistency/design/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 86/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┬──────────┬────────────┤
│ 1. Architecture Clarity      │  19      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  7/7     │          │            │
│    Dependencies listed       │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  17      │  20      │ ⚠️         │
│    Interface signatures typed│  7/7     │          │            │
│    Models concrete           │  6/7     │          │            │
│    Directly implementable    │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  14      │  15      │ ✅         │
│    Error types defined       │  5/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  12      │  15      │ ✅         │
│    Per-layer test plan       │  4/5     │          │            │
│    Coverage target numeric   │  4/5     │          │            │
│    Test tooling named        │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  18      │  20      │ ✅         │
│    Components enumerable     │  7/7     │          │            │
│    Tasks derivable           │  6/7     │          │            │
│    PRD AC coverage           │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┤
│ 6. Security Considerations   │  6       │  10      │ ⚠️         │
│    Threat model present      │  3/5     │          │            │
│    Mitigations concrete      │  3/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  86      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness: 18/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Architecture: Dependencies | External package `gorm.io/gorm` is named but no version pinned; the import path `gormlib` in code samples does not match the stated package name, creating ambiguity about whether this is a renamed import or a different package | -1 pt |
| Interfaces: Models | BaseModel definition is present and complete, but the `isSoftDeletable` blocklist approach uses a negative list (`*model.ProgressRecord, *model.StatusHistory`) — any future non-soft-deletable entity added without updating this list will silently get `NotDeleted` applied and produce SQL errors. The design acknowledges no automatic enforcement but does not discuss this fragility in Decision 1 | -1 pt |
| Interfaces: Directly implementable | Per-repo changes are described as "Each repo method gets `.Scopes(NotDeleted)` added to its query chain" — this is prose instruction, not typed code. The Complete Change List enumerates method names but does not show any of the actual repo method modifications except `RemoveMember` and `SoftDelete`. A developer must read existing source for all other 20+ methods | -2 pt |
| Error: HTTP status mapped | The two cross-layer traces are excellent, but they only cover the success-of-filtering case (deleted record returns 404). There is no trace for the edge case where `RemoveMember` finds no matching row — the code shows `result.RowsAffected == 0` returning `errors.ErrNotFound` (note: `errors` package, not `apperrors`), which may differ from the `apperrors.ErrNotFound` used in the traces. This inconsistency could produce a 500 instead of 404 | -1 pt |
| Testing: Per-layer | Only repository unit tests are planned. The PRD section for Story 4 mentions permission-related joins which span repo + service boundaries. No integration test plan validates the end-to-end "deleted member has no permissions" path across layers | -1 pt |
| Testing: Coverage numeric | "100% of modified methods" is still a method-count metric, not a line/branch coverage percentage. The rubric asks for "numeric coverage target" — while method-level is acceptable, the design does not state the line/branch target explicitly | -1 pt |
| Testing: Tooling | `testify/assert` and `testify/require` are now named, but no test runner is specified (`go test`? `-race`? `-count=1`?). The `setupTestDB` helper is referenced in code but its implementation is not described — is it a project fixture, a shared testutil, or per-test? | -1 pt |
| Breakdown: Tasks derivable | The Complete Change List is a file-level enumeration, not a task decomposition. It does not define task boundaries, dependencies between tasks, or implementation order. E.g., should `helpers.go` changes land before repo-specific changes? Should the migration be first or last? | -1 pt |
| Breakdown: PRD AC coverage | Story 4 AC4 ("multiple deleted members") maps only to "Same as AC3" — but AC3 is about `CountMembersByRoleID` returning 0 for one deleted member. AC4 requires verifying the count is correct with *multiple* deleted members, which implies a distinct test scenario (e.g., 3 members, delete 2, count should be 1, not 0). The design collapses these into one entry | -1 pt |
| Security: Threat model | Three specific threats are now identified with impact/likelihood/mitigation, which is a significant improvement. However, the threat model is incomplete: no threat covers the SubItem re-create race condition (two concurrent requests re-creating the same item_code after soft-delete could hit the unique index differently). No threat covers the migration window — during ALTER TABLE, the database may be briefly inconsistent | -2 pts |
| Security: Mitigations | Mitigations are concrete per threat and include verification tests. However, the "Defense-in-Depth" section explicitly rejects automatic enforcement and relies solely on convention + tests. For a security-critical fix, this is weak: a linter or code review checklist item would provide a stronger guarantee. The stated reason ("would require Go analysis tooling that does not exist") is insufficient — a simple `grep -r` CI check for repo methods missing `NotDeleted` would catch regressions | -2 pts |

---

## Attack Points

### Attack 1: Interface & Models — per-repo changes remain prose, not implementable code

**Where**: "Each repo method gets `.Scopes(NotDeleted)` added to its query chain. For join queries, use `NotDeletedTable(table)`." and the Complete Change List table enumerating method names without code.

**Why it's weak**: The design shows full typed code for `FindByID[T]`, `FindByIDs[T]`, `RemoveMember`, and `SoftDelete` — four code blocks. But the Complete Change List identifies 20+ additional methods across 6 repos that need modification (e.g., `user_repo.FindByUsername`, `main_item_repo.CountByTeam`, `role_repo.GetUserTeamPermissions`). None of these have typed code. For join-query methods like `HasPermission`, the design says "Add `NotDeletedTable("pmw_team_members")`" in a table cell — this is a comment, not a code block showing the actual join query with the scope applied in the correct position. A developer implementing `HasPermission` must open the existing code, figure out where the join is, and insert the scope — the design doc adds no value over the Complete Change List's method names.

**What must improve**: Show at least one representative join-query method (e.g., `HasPermission`) with full typed code including the `NotDeletedTable` scope. For simple `.Scopes(NotDeleted)` additions, a single representative example (e.g., `user_repo.FindByUsername`) with the before/after diff would suffice. The current approach of listing method names in a table is a task list, not a design.

### Attack 2: Security — threat model misses migration window and race conditions

**Where**: "Threat Model" table (T1, T2, T3) and "Defense-in-Depth: Preventing Future Regressions" section stating "No automatic enforcement (e.g., linter rule for `NotDeleted`): rejected because it would require Go analysis tooling that does not exist in this project."

**Why it's weak**: The threat model has three threats but misses two important scenarios: (1) **Migration race condition**: The ALTER TABLE to change the unique index on `pmw_sub_items` runs on a live database. During the brief window between DROP INDEX and ADD INDEX, concurrent inserts could create duplicates — the design does not address whether the migration requires downtime or uses an online DDL strategy. (2) **Concurrent re-create race**: Two requests simultaneously trying to re-create a sub-item with the same `item_code` after its soft-delete could produce unexpected behavior depending on transaction isolation level — the unique index with `deleted_time` should handle this, but the design does not confirm the isolation level. Additionally, the rejection of automated enforcement for a *security* fix is concerning. A simple CI script (`grep -rn 'func.*repo.*Find\|func.*repo.*List' internal/repository/gorm/*.go | grep -v NotDeleted`) would catch most regressions without requiring "Go analysis tooling."

**What must improve**: (1) Add T4 for migration window risk with a mitigation (e.g., "ALTER TABLE runs in a transaction; MySQL supports online DDL for index changes"). (2) Add T5 for concurrent re-create race with a mitigation confirming the unique index + GORM transaction handling prevents duplicates. (3) Reconsider the automatic enforcement rejection — at minimum add a CI grep check to the defense-in-depth section.

### Attack 3: Error Handling — `errors.ErrNotFound` vs `apperrors.ErrNotFound` inconsistency

**Where**: `RemoveMember` code block line: `return errors.ErrNotFound` and the Error Types section: `var ErrNotFound = &AppError{Code: "NOT_FOUND", Status: 404, Message: "resource not found"}` defined in `internal/pkg/errors/errors.go`.

**Why it's weak**: The `RemoveMember` code returns `errors.ErrNotFound` (the `errors` package alias), but the cross-layer error traces show the propagation path using `apperrors.ErrNotFound` and `apperrors.MapNotFound`. If `errors.ErrNotFound` in `RemoveMember` is a different type from `apperrors.ErrNotFound`, the `MapNotFound` helper would not match it, and the error would propagate unhandled to the handler — likely producing a 500 Internal Server Error instead of the expected 404. This is exactly the kind of bug a design document should prevent. The document does not clarify whether `errors` and `apperrors` are the same package with different aliases, or different packages entirely.

**What must improve**: (1) Use consistent package names throughout — either `errors.ErrNotFound` or `apperrors.ErrNotFound`, not both. (2) If `errors` is an alias for `apperrors`, state this explicitly. (3) If `RemoveMember` returns a raw `errors.ErrNotFound` instead of going through `MapNotFound`, the design should add `MapNotFound` to the `RemoveMember` error path or show how the service layer handles the raw `ErrNotFound` from this repo method.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Security — threat model dangerously shallow | Partially | The security section now has a structured threat table with 3 specific threats (T1-T3), each with impact, likelihood, mitigation, and verification test. This is a major improvement from the original two-sentence section. However, the threat model is still incomplete — missing migration window risk and race condition threats. Deduction reduced from -10 to -4. |
| Attack 2: Error Handling — no HTTP status mapping | Yes | The design now includes two detailed cross-layer error traces (SubItem GetByBizKey and Role GetRole) showing the full path from repo -> service -> handler -> HTTP 404. The `ErrNotFound` type is defined with its struct fields. The `MapNotFound` helper is shown with its implementation. This directly addresses the previous attack. The remaining -1 is for the `errors.ErrNotFound` vs `apperrors.ErrNotFound` inconsistency in `RemoveMember`. |
| Attack 3: Interface & Models — incomplete code, no concrete definitions | Mostly Yes | The `FindByIDs[T]` function now has full typed code instead of `// ... same pattern`. The `BaseModel` struct is shown with all fields, types, and GORM tags. The `RemoveMember` and `SoftDelete` methods have full code. However, per-repo changes beyond these are still prose — "Each repo method gets `.Scopes(NotDeleted)`" without showing representative code for any of the 20+ additional methods. |

---

## Verdict

- **Score**: 86/100
- **Target**: 90/100
- **Gap**: 4 points
- **Breakdown-Readiness**: 18/20 — can proceed to `/breakdown-tasks`
- **Action**: Continue to iteration 3 — biggest gains available in Security (6→10), Interface & Models (17→20)

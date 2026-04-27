---
date: "2026-04-27"
doc_dir: "docs/features/soft-delete-consistency/design/"
iteration: 3
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 3

**Score: 93/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┬──────────┬────────────┤
│ 1. Architecture Clarity      │  20      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  7/7     │          │            │
│    Dependencies listed       │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 2. Interface & Model Defs    │  19      │  20      │ ✅         │
│    Interface signatures typed│  7/7     │          │            │
│    Models concrete           │  7/7     │          │            │
│    Directly implementable    │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 3. Error Handling            │  15      │  15      │ ✅         │
│    Error types defined       │  5/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 4. Testing Strategy          │  14      │  15      │ ✅         │
│    Per-layer test plan       │  5/5     │          │            │
│    Coverage target numeric   │  4/5     │          │            │
│    Test tooling named        │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 5. Breakdown-Readiness ★     │  18      │  20      │ ✅         │
│    Components enumerable     │  7/7     │          │            │
│    Tasks derivable           │  6/7     │          │            │
│    PRD AC coverage           │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 6. Security Considerations   │  7       │  10      │ ⚠️         │
│    Threat model present      │  4/5     │          │            │
│    Mitigations concrete      │  3/5     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ TOTAL                        │  93      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness: 18/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Interfaces: Directly implementable | The Complete Change List in the Appendix still enumerates method names per repo file without showing their code. While the two representative examples (simple `user_repo.FindByBizKey` with before/after, and join `role_repo.HasPermission` with before/after) now cover the two patterns, the remaining ~18 methods across `team_repo`, `main_item_repo`, `sub_item_repo`, `item_pool_repo` are still described only as "follow this same one-line addition pattern." A developer must inspect source to confirm the exact `.Scopes()` insertion point for methods like `team_repo.FindTeamsByUserIDs` which uses a join — the representative examples do not cover all join patterns. | -1 pt |
| Testing: Coverage numeric | "100% of modified methods" is a method-count metric. While defensible for this change (each method is a discrete unit), the rubric asks for a numeric coverage target and the document does not state a line or branch coverage percentage. The previous iteration was dinged for the same issue. | -1 pt |
| Breakdown: Tasks derivable | The Complete Change List is a file-level enumeration. Task boundaries, ordering dependencies, and implementation sequence remain implicit. For example, should `helpers.go` changes land before repo-specific changes (so tests can use the new helpers)? Should the migration be first or last? This was a -1 pt deduction in iteration 2 and remains unaddressed. | -1 pt |
| Breakdown: PRD AC coverage | Story 4 AC4 ("multiple deleted members, count returns 2 not 3") maps to "Same as AC3" in the PRD Coverage Map — but AC3 tests one deleted member out of three returning count=2, while AC4 tests three deleted out of five returning count=2. These are distinct test scenarios requiring different seed data. The design still collapses them. Additionally, PRD module 2 lists TeamMember under `nonSoftDeletable` types, but the tech design Decision 2 unifies TeamMember to use soft-delete — the PRD Coverage Map does not address the PRD's own `nonSoftDeletable` classification being overridden. | -1 pt |
| Security: Threat model | The threat model now has 5 entries (T1-T5), a clear improvement. However, T4's mitigation states "Both DDL statements run inside a single transaction (BEGIN; ALTER ... DROP; ALTER ... ADD; COMMIT;)" but the Migration SQL section shows two separate ALTER TABLE statements, not a single transaction-wrapped statement. MySQL's online DDL behavior differs from InnoDB transactional DDL — certain ALTER TABLE operations cause implicit commits, breaking the transaction guarantee. The threat model does not address this MySQL-specific limitation. | -1 pt |
| Security: Mitigations | The CI grep check is a welcome addition (addresses prior attack). However, the grep pattern `grep -rn '\.Find(\|\.First(' backend/internal/repository/gorm/*.go | grep -v 'NotDeleted'` is too narrow: it misses `.Where(` chains that do not end in `.Find()` or `.First()` (e.g., `.Count()`, `.Updates()` on the wrong table). Also, the script syntax uses `\|` ( BRE alternation) with `-rn` flags — on some systems this requires `-E` flag for ERE, and the script does not specify a shell. | -2 pts |

---

## Attack Points

### Attack 1: Security — T4 mitigation contradicts MySQL DDL behavior

**Where**: "Both DDL statements run inside a single transaction (BEGIN; ALTER ... DROP; ALTER ... ADD; COMMIT;). MySQL supports online DDL for index changes — the table remains readable during rebuild."

**Why it's weak**: MySQL's InnoDB engine issues an implicit commit before and after each DDL statement (ALTER TABLE, DROP INDEX, CREATE INDEX). Wrapping two ALTER TABLE statements in `BEGIN; ... COMMIT;` does not make them atomic — the first ALTER commits, the second ALTER is a separate transaction. If the second ALTER fails, the database is left without the unique index. The mitigation's claim of transactional safety is factually incorrect for MySQL. The correct approach would be a single `ALTER TABLE ... DROP INDEX ..., ADD UNIQUE KEY ...` statement or explicit acknowledgment that the migration has a brief inconsistent window with a rollback plan.

**What must improve**: Either (1) combine both DDL operations into a single ALTER TABLE statement (`ALTER TABLE pmw_sub_items DROP INDEX uk_sub_items_main_code, ADD UNIQUE KEY uk_sub_items_main_code (...)`) which MySQL executes atomically, or (2) remove the transaction claim and instead document the brief window risk with a manual rollback SQL statement for operational safety.

### Attack 2: Breakdown-Readiness — task ordering and dependencies remain implicit

**Where**: The Complete Change List table in Appendix listing 9 files with changes but no ordering or dependency information.

**Why it's weak**: The design identifies 9 files to modify but provides no guidance on implementation order. This matters because: (1) `helpers.go` changes (isSoftDeletable + FindByID/FindByIDs modifications) must land before repo-specific tests can rely on the new behavior; (2) the SubItem SoftDelete fix is a prerequisite for the schema migration (the new unique index requires `deleted_flag` and `deleted_time` to be correctly set by SoftDelete); (3) the migration should be applied after code changes are tested, not before. A developer reading the Complete Change List has no indication of these ordering constraints. The design was dinged for this in iteration 2 and the gap persists.

**What must improve**: Add a numbered implementation sequence (e.g., "Phase 1: helpers.go + isSoftDeletable → Phase 2: repo methods → Phase 3: SoftDelete fix → Phase 4: schema migration") with explicit dependencies stated. This can be 3-4 lines of text in the Appendix.

### Attack 3: Interface & Models — PRD/Design inconsistency on TeamMember soft-deletability

**Where**: PRD module 2 lists `nonSoftDeletable: ProgressRecord, StatusHistory, TeamMember` while the tech design Decision 2 states "Replace db.Delete(&model.TeamMember{}) with Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()})" and the `isSoftDeletable` type switch does not list TeamMember in its negative cases.

**Why it's weak**: The PRD classifies TeamMember as non-soft-deletable (no deleted_flag), but the tech design unifies it to soft-delete. The `isSoftDeletable` function's default branch returns `true` — meaning TeamMember would now get `NotDeleted` applied to it if it goes through the generic helpers. However, `team_repo.FindMember` and `team_repo.CountMembers` are listed in the PRD as "no change — TeamMember has no deleted_flag" but the design's `RemoveMember` now sets `deleted_flag=1`. This creates a contradiction: the PRD says these methods need no change, but after Decision 2 they DO need `NotDeleted` filtering. The PRD Coverage Map does not list `FindMember` or `CountMembers` as needing updates. If `FindMember` does not add `NotDeleted`, a soft-deleted member could still be found by `FindMember`, which would allow re-adding to fail (or return stale data). This is a latent correctness bug the design inherits from the PRD but fails to surface or resolve.

**What must improve**: Either (1) update the PRD Coverage Map to include `FindMember` and `CountMembers` as methods requiring `NotDeleted` after the TeamMember soft-delete unification, with corresponding test scenarios, or (2) add a note explaining that these methods are called in contexts where the caller already ensures the member is active (with evidence), or (3) add these to the Complete Change List as additional methods requiring modification.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Interface & Models — per-repo changes remain prose, not implementable code | Mostly Yes | The design now includes two representative examples: `user_repo.FindByBizKey` (simple query, before/after) and `role_repo.HasPermission` (join query, before/after). These cover the two patterns (simple `.Scopes(NotDeleted)` and join `.Scopes(NotDeletedTable(...))`). The remaining methods are described as "follow this same one-line addition pattern" with a one-line summary. This is an improvement from pure prose to pattern-based specification, though the ~18 remaining methods still lack individual code blocks. Deduction reduced from -2 to -1. |
| Attack 2: Security — threat model misses migration window and race conditions | Yes | The threat model now includes T4 (migration window risk) and T5 (concurrent re-create race). T4 addresses the ALTER TABLE window with a mitigation strategy. T5 explains why the unique index with `deleted_time` prevents collision. A CI grep check was added to the defense-in-depth section. However, T4's mitigation claim (transactional DDL) is factually incorrect for MySQL, so new deduction applied. |
| Attack 3: Error Handling — errors.ErrNotFound vs apperrors.ErrNotFound inconsistency | Yes | The design now includes a dedicated "Package alias clarification" paragraph stating: "Repo files import this package as `errors`. Service and handler files import the same package as `apperrors`. They are the same package — `errors.ErrNotFound` and `apperrors.ErrNotFound` are the same value." The cross-layer traces also now consistently use the correct alias per layer. This directly resolves the ambiguity. Full marks restored to Error Handling. |

---

## Verdict

- **Score**: 93/100
- **Target**: 90/100
- **Gap**: 0 points — target exceeded by 3
- **Breakdown-Readiness**: 18/20 — can proceed to `/breakdown-tasks`
- **Action**: Target reached. Design is ready for `/breakdown-tasks`.

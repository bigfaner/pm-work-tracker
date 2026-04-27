---
date: "2026-04-27"
doc_dir: "docs/features/soft-delete-consistency/design/"
iteration: 4
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 4

**Score: 94/100** (target: 90)

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
├──────────────────────────────┼──────────┬──────────┬────────────┤
│ 2. Interface & Model Defs    │  19      │  20      │ ✅         │
│    Interface signatures typed│  7/7     │          │            │
│    Models concrete           │  7/7     │          │            │
│    Directly implementable    │  5/6     │          │            │
├──────────────────────────────┼──────────┬──────────┬────────────┤
│ 3. Error Handling            │  15      │  15      │ ✅         │
│    Error types defined       │  5/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  5/5     │          │            │
├──────────────────────────────┼──────────┬──────────┬────────────┤
│ 4. Testing Strategy          │  14      │  15      │ ✅         │
│    Per-layer test plan       │  5/5     │          │            │
│    Coverage target numeric   │  4/5     │          │            │
│    Test tooling named        │  5/5     │          │            │
├──────────────────────────────┼──────────┬──────────┬────────────┤
│ 5. Breakdown-Readiness ★     │  18      │  20      │ ✅         │
│    Components enumerable     │  7/7     │          │            │
│    Tasks derivable           │  7/7     │          │            │
│    PRD AC coverage           │  4/6     │          │            │
├──────────────────────────────┼──────────┬──────────┬────────────┤
│ 6. Security Considerations   │  8       │  10      │ ⚠️         │
│    Threat model present      │  4/5     │          │            │
│    Mitigations concrete      │  4/5     │          │            │
├──────────────────────────────┼──────────┬──────────┬────────────┤
│ TOTAL                        │  94      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness: 18/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Interfaces: Directly implementable | The two representative examples (user_repo.FindByBizKey simple query, role_repo.HasPermission join query) cover the two patterns well. However, the remaining ~18 methods still rely on "follow this same one-line addition pattern" prose. Methods like `team_repo.FindTeamsByUserIDs` and `team_repo.FindPMMembers` involve multi-table joins with `NotDeletedTable` on different tables — the single join example uses `NotDeletedTable("pmw_team_members")`, but FindPMMembers needs `NotDeletedTable("pmw_users")` and FindTeamsByUserIDs needs `NotDeletedTable("pmw_teams")`. A developer must read source code to confirm the correct table argument. | -1 pt |
| Testing: Coverage numeric | The coverage target is "100% of modified methods (each changed method has at least one test verifying NotDeleted behavior)." This is a method-count metric, not a line or branch coverage percentage. The rubric asks for a numeric coverage target. While defensible for this change's granularity (each method is a discrete unit), a numeric line/branch target (e.g., ">90% line coverage") would be more precise and is standard practice. This gap has persisted across all four iterations. | -1 pt |
| Breakdown: PRD AC coverage | The PRD Coverage Map now includes FindMember and CountMembers, resolving iteration 3's attack. However, the PRD (module 2) classifies TeamMember as `nonSoftDeletable`, while the design's Decision 2 overrides this to soft-deletable. The Coverage Map does not include a reconciliation note explaining this PRD correction. Additionally, the PRD's module 8 schema validation says "deleted_flag=0 时为 NULL" for deleted_time, but the BaseModel definition shows `gorm:"not null;default:'1970-01-01 08:00:00'"` — deleted_time is NOT NULL with a sentinel value, not NULL. This PRD/design mismatch is not surfaced. | -2 pts |
| Security: Threat model | T4 was improved to use a single atomic ALTER TABLE statement. However, T4's description still contains the claim "MySQL ALTER TABLE causes implicit commits — BEGIN; DROP INDEX; ADD INDEX; COMMIT; does NOT wrap atomically" — this describes the OLD approach that was already fixed. The mitigation now correctly uses a single ALTER, but the threat description and mitigation narrative are internally inconsistent: the threat describes the problem with two separate statements, while the mitigation already solved it with one. The description should be streamlined to simply state the risk is eliminated by the single-ALTER approach. | -1 pt |
| Security: Mitigations | The CI grep check script uses `\|` (BRE alternation) with `grep -rn`: `grep -rn '\.Find(\|\.First(' backend/internal/repository/gorm/*.go`. On macOS and some Linux systems, `\|` in BRE requires special handling or `-E` flag for ERE. Also, the pattern misses `.Where(` chains ending in `.Count()` or `.Row()` — these are query methods that also need NotDeleted but would not be caught by this grep. The heuristic is acknowledged as incomplete but the gap is not quantified. | -1 pt |

---

## Attack Points

### Attack 1: Breakdown-Readiness — PRD/design schema mismatch on deleted_time NULLability is not reconciled

**Where**: PRD module 8 schema validation states `deleted_flag=0 时为 NULL` for `deleted_time`, while the design's BaseModel definition shows `DeletedTime time.Time gorm:"not null;default:'1970-01-01 08:00:00'"`.

**Why it's weak**: The PRD says deleted_time is NULL for active records, but the design's BaseModel constraint is `NOT NULL` with a sentinel default of `'1970-01-01 08:00:00'`. This means the unique index `(main_item_key, item_code, deleted_flag, deleted_time)` includes the sentinel timestamp for all active rows. If two different active SubItems with the same `(main_item_key, item_code)` but different IDs existed before the index was applied, the index creation would fail because both rows have `deleted_flag=0, deleted_time='1970-01-01 08:00:00'`. The design does not address this pre-migration data validation step — what happens if duplicate active rows already exist before the ALTER TABLE runs?

**What must improve**: Add a pre-migration data validation query to the Migration SQL section (e.g., `SELECT main_item_key, item_code, COUNT(*) FROM pmw_sub_items WHERE deleted_flag = 0 GROUP BY main_item_key, item_code HAVING COUNT(*) > 1;`) and specify the remediation if duplicates are found (manual deduplication before applying the index).

### Attack 2: Security — CI grep enforcement script has portability and coverage gaps

**Where**: "A lightweight enforcement script runs in CI to catch repo methods that query soft-deletable tables without NotDeleted: `! grep -rn '\.Find(\|\.First(' backend/internal/repository/gorm/*.go | grep -v 'NotDeleted' | grep -v '_test.go'`"

**Why it's weak**: Two issues. First, the `\|` BRE alternation syntax is not portable — on macOS BSD grep, `\|` works in BRE mode, but on some Linux distributions, `grep -rn '\.\Find(\|\.First('` may require `-E` (extended regex) to interpret `|` as alternation. The script should use `-E` and `\|` → `|`, or use `grep -E '\.Find\(|\.First\('`. Second, the grep only checks for `.Find(` and `.First(` calls, but GORM also supports `.Count()`, `.Row()`, `.Rows()`, `.Scan()`, `.Pluck()` as terminal query methods. A method that uses `.Where(...).Count(&n)` without `.Find()` or `.First()` would pass the grep check despite missing `NotDeleted`. The `team_repo.CountMembers` method uses `.Count()` — if a developer adds a new counting method without `NotDeleted`, the grep would miss it.

**What must improve**: Either (1) expand the grep pattern to cover all GORM terminal query methods (`.Find(`, `.First(`, `.Count(`, `.Row()`, `.Rows()`, `.Scan(`, `.Pluck(`), or (2) acknowledge the gap and document that the grep is a heuristic that catches the 80% case, with the test pattern as the primary enforcement mechanism. Also add `-E` flag for portability.

### Attack 3: Interface & Models — isSoftDeletable uses negative list that silently breaks if new non-soft-deletable types are added

**Where**: "func isSoftDeletable[T any]() bool { switch any(new(T)).(type) { case *model.ProgressRecord, *model.StatusHistory: return false; default: return true } }"

**Why it's weak**: The function uses a negative list (excludes specific types) with a `default: return true` branch. If a developer later adds a new model type that does NOT embed BaseModel and does NOT have `deleted_flag`, the function will incorrectly return `true`, causing SQL errors (`column deleted_flag does not exist`) at runtime. This is the exact same class of bug the design is fixing — the lack of `NotDeleted` was caused by an implicit default (no filter), and this function reproduces the same pattern (default = soft-deletable). The design does not discuss this risk or propose a compile-time safeguard (e.g., a `SoftDeletable` interface that types must explicitly implement, which would cause a compile error for new types that don't implement it).

**What must improve**: Either (1) invert the logic to use a positive list (explicitly list types that ARE soft-deletable, with `default: return false`), which makes new types safe by default, or (2) add a comment/defense note acknowledging the negative-list risk and documenting the CI grep as the safety net for new types. Option (1) is strictly safer — a new non-soft-deletable type would simply not be filtered, which is correct (no crash, just no filter until explicitly added).

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (iter 3): T4 false transactional DDL claim | Yes | The Migration SQL section now uses a single atomic ALTER TABLE statement: `ALTER TABLE pmw_sub_items DROP INDEX uk_sub_items_main_code, ADD UNIQUE KEY uk_sub_items_main_code (...)`. The threat model T4 still mentions the old two-statement problem in its description, but the actual mitigation is the single-statement approach. Deduction reduced from prior level. |
| Attack 2 (iter 3): missing phase order / task dependencies | Yes | A new "Implementation Phase Order" section was added to the Appendix with 5 phases (P1-P5), explicit dependency chains (P2 depends on P1, etc.), and rationale per phase. This directly resolves the ordering gap. Full marks restored for "Tasks derivable" (7/7). |
| Attack 3 (iter 3): FindMember/CountMembers not in coverage map | Yes | The PRD Coverage Map now includes explicit rows: "TeamMember: FindMember excludes deleted" and "TeamMember: CountMembers excludes deleted" with their design components and interfaces. The Complete Change List also lists these methods under `gorm/team_repo.go`. |

---

## Verdict

- **Score**: 94/100
- **Target**: 90/100
- **Gap**: 0 points — target exceeded by 4
- **Breakdown-Readiness**: 18/20 — can proceed to `/breakdown-tasks`
- **Action**: Target reached. Design is ready for `/breakdown-tasks`.

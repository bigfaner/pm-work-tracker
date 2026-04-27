---
date: "2026-04-27"
doc_dir: "docs/features/soft-delete-consistency/prd/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 2

**Score: 96/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  20      │  20      │ ✅         │
│    Background three elements │  7/7     │          │            │
│    Goals quantified          │  7/7     │          │            │
│    Logical consistency       │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  19      │  20      │ ✅         │
│    Mermaid diagram exists    │  7/7     │          │            │
│    Main path complete        │  6/7     │          │            │
│    Decision + error branches │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  17      │  20      │ ⚠️         │
│    Tables complete           │  7/7     │          │            │
│    Field descriptions clear  │  7/7     │          │            │
│    Validation rules explicit │  6/6     │          │            │
│    Inconsistency deduction   │  -3      │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  20      │  20      │ ✅         │
│    Coverage per user type    │  7/7     │          │            │
│    Format correct            │  7/7     │          │            │
│    AC per story              │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  20      │  20      │ ✅         │
│    In-scope concrete         │  7/7     │          │            │
│    Out-of-scope explicit     │  7/7     │          │            │
│    Consistent with specs     │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  96      │  100     │ ✅         │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md:109 vs :194-196 | TeamMember deleted_flag inconsistency — line 109 lists pmw_team_members as a table WITHOUT deleted_flag, line 125 classifies TeamMember as nonSoftDeletable, lines 149-150 state "TeamMember 无 deleted_flag", but lines 194-196 apply NotDeletedTable("pmw_team_members") to three Role repo methods, which would generate `WHERE pmw_team_members.deleted_flag = 0` — a SQL error on a column that doesn't exist | -3 pts |
| prd-spec.md:76-89 vs :94 | Flow diagram covers development workflow only; the business data flow (API -> Handler -> Service -> Repo -> SQL) is described only as text at line 94. A Mermaid sequence or data flow diagram would show how deleted records currently leak and how the fix prevents it | -1 pt |

---

## Attack Points

### Attack 1: Functional Specs — TeamMember deleted_flag contradiction across modules

**Where**: prd-spec.md lines 109, 125, 149-150 vs lines 194-196
> Line 109: "对不含 deleted_flag 的表（如 pmw_progress_records、pmw_status_history、**pmw_team_members**）应用 NotDeleted 会导致 SQL 错误"
> Line 125: "**TeamMember** — 不含 deleted_flag 字段" (nonSoftDeletable)
> Line 149: "无变更 — **TeamMember 无 deleted_flag**"
> Line 150: "无变更 — **TeamMember 无 deleted_flag**"
> Line 194: HasPermission adds `.Scopes(NotDeletedTable("pmw_team_members"))`
> Line 195: GetUserTeamPermissions adds `.Scopes(NotDeletedTable("pmw_team_members"))`
> Line 196: CountMembersByRoleID adds `.Scopes(NotDeletedTable("pmw_team_members"))`

**Why it's weak**: The document states in three separate locations that pmw_team_members lacks deleted_flag, yet prescribes applying NotDeletedTable to that table in three Role repo methods. If pmw_team_members truly has no deleted_flag column, the fix at lines 194-196 would cause SQL errors (column not found). If pmw_team_members DOES have deleted_flag, then lines 109, 125, 149, and 150 contain false statements. The PRD cannot be implemented as written — a developer must resolve this contradiction before coding.

**What must improve**: Decide whether TeamMember has deleted_flag or not. If it does, remove pmw_team_members from the nonSoftDeletable list at line 125, update line 109, and change FindMember/CountMembers at lines 149-150 to also apply NotDeleted. If it does not, remove the NotDeletedTable("pmw_team_members") scopes from lines 194-196 and replace with the correct filtering strategy (e.g., JOIN on users table and filter pmw_users.deleted_flag instead).

### Attack 2: Flow Diagrams — development workflow diagram, not business data flow

**Where**: prd-spec.md lines 76-89, the Mermaid flowchart titled "业务流程图"
> The diagram shows: Start -> Tests -> Generic helpers -> Repos -> SubItem -> Schema -> Verify -> Done, which is the developer's implementation sequence, not the system's data flow.

**Why it's weak**: A PRD's flow diagram should illustrate the business problem and solution from a system behavior perspective — how data flows, where deleted records leak, and how the fix closes those gaps. The current diagram is an implementation task sequence (TDD steps), which belongs in a technical design document, not a PRD. The actual data flow is buried in a single prose sentence at line 94: "API 请求 → Handler → Service → Repo（此处添加 NotDeleted 过滤）→ GORM → SQL（WHERE deleted_flag = 0）→ 数据库." A Mermaid sequence diagram showing this flow with decision points (record deleted? -> filter out vs. return) would better serve the PRD's audience.

**What must improve**: Add a second Mermaid diagram (sequence or flowchart) showing the request data flow: API request -> Handler -> Service -> Repo layer (decision: does entity have deleted_flag? -> apply NotDeleted or pass through) -> Database. This shows stakeholders the WHERE and HOW of the fix, not just the development sequence.

### Attack 3: Functional Specs — SoftDelete silent failure on non-existent ID

**Where**: prd-spec.md line 173
> "id 不存在 → Updates 影响行数 0，不报错（当前行为保留）"

**Why it's weak**: The SoftDelete method silently succeeds when given a non-existent ID. The PRD acknowledges this behavior with "（当前行为保留）" but does not document the design decision: is this intentional (idempotent delete) or a known limitation? Other methods like FindByBizKey return ErrNotFound when the record doesn't exist. The inconsistency in error handling across the same module is not justified. If a caller deletes a SubItem that was already deleted (or never existed), they get no feedback — is that acceptable? This should be an explicit design decision, not a parenthetical note.

**What must improve**: Add an explicit statement: either "SoftDelete is idempotent by design — calling it on a non-existent or already-deleted ID returns nil without error" (with rationale), or specify that the fix should check affected rows and return ErrNotFound. This is especially important because the unique index change (line 204) allows multiple soft-deleted records with the same item_code, so the idempotency guarantee matters for correctness.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Functional Specs — missing validation rules and field-level detail | ✅ Yes | Document now has 8 per-module tables with typed parameters, return values, error conditions. Lines 107-110 add explicit NotDeleted scope validation rules (type, value domain, error cases). Line 204 adds unique index validation rules. |
| Attack 2: Functional Specs — table structure does not match rubric requirements | ✅ Yes | Flat change inventory replaced with 8 structured per-module tables. Each method row has: method name, typed parameters, return type, current behavior, fixed behavior, error conditions (6-7 columns). |
| Attack 3: User Stories — ACs lack boundary conditions and edge cases | ✅ Yes | Story 4 expanded from 1 AC to 5 ACs covering HasPermission (AC1), GetUserTeamPermissions (AC2), CountMembersByRoleID (AC3), multiple deleted members boundary (AC4), all-deleted user boundary (AC5). Story 1 first AC rewritten with active voice. |

---

## Verdict

- **Score**: 96/100
- **Target**: 90/100
- **Gap**: 0 points (target exceeded by 6 points)
- **Action**: Target reached. One material inconsistency (TeamMember deleted_flag contradiction) should be resolved before implementation, but overall document quality is strong.

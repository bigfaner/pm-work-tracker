---
date: "2026-05-08"
doc_dir: "docs/proposals/unify-permission-checks/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 2

**Score: 93/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  19      │  20      │ ✅         │
│    Problem clarity           │  7/7     │          │            │
│    Evidence provided         │  7/7     │          │            │
│    Urgency justified         │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Solution Clarity          │  19      │  20      │ ✅         │
│    Approach concrete         │  7/7     │          │            │
│    User-facing behavior      │  6/7     │          │            │
│    Differentiated            │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  15      │  15      │ ✅         │
│    Alternatives listed (≥2)  │  5/5     │          │            │
│    Pros/cons honest          │  5/5     │          │            │
│    Rationale justified       │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  14      │  15      │ ⚠️         │
│    In-scope concrete         │  5/5     │          │            │
│    Out-of-scope explicit     │  5/5     │          │            │
│    Scope bounded             │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Risk Assessment           │  13      │  15      │ ⚠️         │
│    Risks identified (≥3)     │  5/5     │          │            │
│    Likelihood + impact rated │  4/5     │          │            │
│    Mitigations actionable    │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  13      │  15      │ ⚠️         │
│    Measurable                │  5/5     │          │            │
│    Coverage complete         │  4/5     │          │            │
│    Testable                  │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  93      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Problem: Urgency | No quantitative anchor for urgency — how many handlers added per sprint, or how many bypass-related bugs filed? | -1 pt (criterion 1.3) |
| Solution: User Impact table, SuperAdmin row | States "无功能性变化" but also says "前端不再显示 'SuperAdmin' 标签" — that is an observable functional change, making the label contradictory | -1 pt (criterion 2.2) |
| Scope: bounded | Still no time estimate or effort sizing — scope bounded by deliverables only, not by sprint capacity or calendar | -1 pt (criterion 4.3) |
| Risk: likelihood ratings | Rows 1 and 3 lack justification for their likelihood ratings ("Low", "Medium") — rows 2 and 5 added justifications but 1 and 3 did not | -1 pt (criterion 5.2) |
| Risk: row 1 mitigation | "migration 中将 is_super_admin=true 的用户自动绑定到 superadmin 角色；事务保证原子性" — no rollback plan described if migration partially fails | -1 pt (criterion 5.3) |
| Success Criteria: coverage | In-scope item "更新 docs/conventions/permission-codes.md 中的 SuperAdmin Bypass Rule 章节" has no corresponding success criterion | -1 pt (criterion 6.2) |
| Success Criteria: testability | Criterion 6 includes subjective phrase "等同移除前" alongside the checklist — the checklist is testable but the qualifier weakens precision | -1 pt (criterion 6.3) |

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: User-facing behavior completely absent | ✅ Yes | New "User Impact" table added with 4 personas (SuperAdmin, PM, ext-member, member), each with before/after columns describing observable behavior |
| Attack 2: Mitigations non-actionable or circular (risk rows 4 and 5) | ✅ Yes | Row 4 now cites `rbac.go` seed data audit: "当前 3 个预设角色中仅 PM 持有 team:update..." — evidence-based, not circular. Row 5 explicitly accepts behavior as intentional RBAC design with documented escape hatch |
| Attack 3: "所有操作" unmeasurable + team_handler uncovered | ✅ Yes | Criterion 6 replaced "所有操作" with explicit checklist (创建团队、更新团队信息...). Criterion 7 added for team_handler 5 call sites with specific permission codes per operation |

---

## Attack Points

### Attack 1: Scope Definition — Still no effort sizing or time estimate

**Where**: The "Scope" section lists concrete deliverables but contains no time estimate, sprint allocation, or effort sizing of any kind.

**Why it's weak**: This was flagged in iteration 1 and left unchanged. The scope is bounded by what will be delivered but not by when or how much effort it requires. A team reading this proposal cannot assess whether this is a 1-sprint or 3-sprint effort. The alternatives section mentions "~15 个 handler" and "4 个预设角色" as system scale — these same numbers could form a rough effort estimate. Without it, the scope is open-ended by calendar.

**What must improve**: Add an estimated effort (e.g., "estimated 3-5 working days" or "fits within 1 sprint") or list a rough breakdown with effort per workstream (backend migration, handler refactoring, frontend removal, testing).

### Attack 2: Success Criteria — docs update in scope but no criterion covers it

**Where**: In-scope item "更新 `docs/conventions/permission-codes.md` 中的 SuperAdmin Bypass Rule 章节" has no matching success criterion. The 9 success criteria cover code, tests, and behavior but not documentation.

**Why it's weak**: This is a coverage gap per rubric criterion 6.2. A reviewer checking "did we do everything in scope?" would find the docs update is deliverable with no verification gate. This also triggers the rubric's inconsistency penalty consideration — scope items should be traceable to success criteria.

**What must improve**: Add a criterion such as "docs/conventions/permission-codes.md 中 SuperAdmin Bypass Rule 章节已更新，反映权限码驱动的统一校验方式" or fold documentation verification into an existing criterion.

### Attack 3: Risk Assessment — Some likelihood ratings still unjustified

**Where**: Risk row 1: "现有 SuperAdmin 用户迁移失败 — Likelihood: Low". Risk row 3: "前端遗漏 isSuperAdmin 引用导致运行时错误 — Likelihood: Medium". Neither has parenthetical justification unlike rows 2 and 5.

**Why it's weak**: Iteration 1 attacked this for row 2. Rows 2 and 5 were fixed with justifications, but rows 1 and 3 were not. Why is migration failure "Low" likelihood — is it because there are few SuperAdmin users, or because the migration is simple, or because there's a dry-run plan? Why is frontend遗漏 "Medium" — is it because `isSuperAdmin` appears in many files, or because TypeScript won't catch runtime-only references? The reader must guess.

**What must improve**: Add parenthetical justifications for rows 1 and 3 matching the style of rows 2 and 5. E.g., row 1: "Low（仅影响 is_super_admin=true 的少量用户，迁移逻辑为单条 UPDATE 语句）"; row 3: "Medium（isSuperAdmin 在前端 ~N 个文件中引用，TypeScript 编译可捕获大部分但非全部）".

---

## Verdict

- **Score**: 93/100
- **Target**: 90/100
- **Gap**: 0 points (target exceeded by 3)
- **Action**: Target reached

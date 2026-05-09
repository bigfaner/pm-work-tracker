---
date: "2026-05-08"
doc_dir: "docs/proposals/unify-permission-checks/"
iteration: 1
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 1

**Score: 84/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  19      │  20      │ ⚠️         │
│    Problem clarity           │  7/7     │          │            │
│    Evidence provided         │  7/7     │          │            │
│    Urgency justified         │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Solution Clarity          │  16      │  20      │ ⚠️         │
│    Approach concrete         │  7/7     │          │            │
│    User-facing behavior      │  4/7     │          │            │
│    Differentiated            │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  13      │  15      │ ⚠️         │
│    Alternatives listed (≥2)  │  5/5     │          │            │
│    Pros/cons honest          │  4/5     │          │            │
│    Rationale justified       │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  14      │  15      │ ⚠️         │
│    In-scope concrete         │  5/5     │          │            │
│    Out-of-scope explicit     │  5/5     │          │            │
│    Scope bounded             │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Risk Assessment           │  11      │  15      │ ⚠️         │
│    Risks identified (≥3)     │  5/5     │          │            │
│    Likelihood + impact rated │  3/5     │          │            │
│    Mitigations actionable    │  3/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  11      │  15      │ ⚠️         │
│    Measurable                │  4/5     │          │            │
│    Coverage complete         │  4/5     │          │            │
│    Testable                  │  3/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  84      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Solution: item 3 | User-facing behavior not described for any change | -3 pts (criterion 2.2) |
| Solution: item 6 | "替换为权限码驱动" is vague about what the user experiences | -2 pts (criterion 2.2) |
| Risk: row 4 | Mitigation "权限码本身定义了操作权限；如果操作确实只允许 PM，则只给 PM 角色授予该权限码" is circular reasoning, not a mitigation | -2 pts (criterion 5.3) |
| Risk: row 5 | Mitigation punts to "后续迭代" for potential regression | -2 pts (criterion 5.3) |
| Risk: row 2 | Likelihood rated "Medium" without justification — why is it medium? | -2 pts (criterion 5.2) |
| Alternatives: row 3 | "当前系统规模适合一次到位" is an assertion without data on system size | -1 pt (criterion 3.3) |
| Success Criteria: item 5 | "SuperAdmin 角色用户通过权限码正常执行所有操作（等同移除前）" — "所有操作" is unmeasurable | -1 pt (criterion 6.1) |
| Success Criteria | No criterion covers team_handler changes (5 call sites) | -1 pt (criterion 6.2) |
| Success Criteria: items 4, 5, 7 | Grep-based checks are testable, but "正常执行所有操作" is not write-a-test-for verifiable | -2 pts (criterion 6.3) |
| Scope | No time estimate or effort sizing — scope is bounded by deliverables but not by calendar or sprint capacity | -1 pt (criterion 4.3) |

---

## Attack Points

### Attack 1: Solution Clarity — User-facing behavior completely absent

**Where**: The entire "Proposed Solution" section lists 7 internal implementation steps but never describes what an end user (admin, PM, ext-member, SuperAdmin) will observe differently after the change.

**Why it's weak**: A proposal is not just an implementation plan. The rubric asks "What does the end user experience? Not internals — the observable behavior." The 7 items are all internals: delete field, seed permissions, remove function, remove shortcut, inject team identity, replace logic, update frontend. None say: "An ext-member with `sub_item:update` will now see the Edit button and succeed on save" or "A SuperAdmin will see no UI change." The reader must reverse-engineer user impact from code changes.

**What must improve**: Add a "User Impact" subsection with one row per user persona (SuperAdmin, PM, ext-member, member) describing the observable behavior before and after. E.g., "ext-member: before = 403 on sub_item edit; after = edit succeeds if role has `sub_item:update`."

### Attack 2: Risk Assessment — Mitigations are non-actionable or circular

**Where**: Risk row 4 mitigation: "权限码本身定义了操作权限；如果操作确实只允许 PM，则只给 PM 角色授予该权限码". Risk row 5 mitigation: "由 `progress:create` 权限码控制；如果需要限制，可通过更细粒度的权限码拆分（后续迭代）".

**Why it's weak**: Row 4's mitigation is circular — it says "the permission code controls access, and if PM-only is needed, only give the code to PM." That is the existing design, not a mitigation for the risk that non-PM roles with the same permission code will perform PM-only operations. It restates the problem as its own solution. Row 5 defers to a future iteration, which means the risk is acknowledged but not mitigated in this proposal's scope. The rubric requires mitigations someone can act on, not "we'll handle it later."

**What must improve**: Row 4: Audit which permission codes are currently PM-only in practice (not just in name) and list them. State explicitly whether any non-PM role currently holds `team:update` / `team:invite` / `team:remove`. If none do, say so with evidence — that IS the mitigation. Row 5: Either accept the behavior change (non-PM with `progress:create` can add progress to any sub_item) and document it as intentional, or include a fine-grained split in scope.

### Attack 3: Success Criteria — "所有操作" is unmeasurable and team_handler changes are uncovered

**Where**: Success criterion "SuperAdmin 角色用户通过权限码正常执行所有操作（等同移除前）" and the absence of any criterion mentioning team_handler's 5 replaced call sites.

**Why it's weak**: "所有操作" is a rubber criterion — it cannot be objectively verified without an exhaustive list of operations. A test cannot be written for "all operations work." There is also a coverage gap: the scope explicitly lists "替换 team_handler 中 5 处 'SuperAdmin 充当 PM' 逻辑为权限码驱动" but no success criterion checks that these 5 replacements work correctly. This is an inconsistency penalty per the rubric: success criteria don't cover all in-scope items.

**What must improve**: (1) Replace "所有操作" with an explicit checklist: "SuperAdmin can: create team, update team, invite member, remove member, [etc.]" — one testable assertion per operation. (2) Add a criterion: "team_handler 的 5 处 '充当 PM' 逻辑全部替换为权限码校验，非 PM 但有 `team:update` 权限码的角色可执行团队管理操作."

---

## Verdict

- **Score**: 84/100
- **Target**: 90/100
- **Gap**: 6 points
- **Action**: Continue to iteration 2

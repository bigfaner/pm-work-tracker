---
date: "2026-04-27"
doc_dir: "docs/features/soft-delete-consistency/prd/"
iteration: 1
target_score: 90
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 1

**Score: 71/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  17      │  20      │ ⚠️         │
│    Background three elements │  6/7     │          │            │
│    Goals quantified          │  6/7     │          │            │
│    Logical consistency       │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  18      │  20      │ ⚠️         │
│    Mermaid diagram exists    │  7/7     │          │            │
│    Main path complete        │  6/7     │          │            │
│    Decision + error branches │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  13      │  20      │ ❌         │
│    Tables complete           │  5/7     │          │            │
│    Field descriptions clear  │  5/7     │          │            │
│    Validation rules explicit │  3/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  17      │  20      │ ⚠️         │
│    Coverage per user type    │  6/7     │          │            │
│    Format correct            │  6/7     │          │            │
│    AC per story              │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  18      │  20      │ ⚠️         │
│    In-scope concrete         │  7/7     │          │            │
│    Out-of-scope explicit     │  6/7     │          │            │
│    Consistent with specs     │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  71      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md:34 | Goal "防御性覆盖" lacks a numeric quantifier — no %, count, or time metric for the helpers goal | -1 pts |
| prd-spec.md:28 | "开发人员" listed as a user but the problem they face (future work) is aspirational, not a current reported issue | -1 pts |
| prd-spec.md:76-87 | Flow diagram covers only development workflow, not the data/request flow showing how soft-deleted records leak through API calls | -1 pts |
| prd-spec.md:83 | Only one decision diamond in entire diagram; no error branches for schema migration failure or individual repo fix conflicts | -1 pts |
| prd-spec.md:97-106 | Functional description table uses custom columns (序号/涉及模块/功能模块/关联改动点/更改后逻辑说明) instead of the rubric-expected structure with field types, data sources, and validation rules | -2 pts |
| prd-spec.md:100-106 | Field descriptions lack explicit types and data sources — e.g., `NotDeleted` scope is described but its SQL semantics, input parameters, and return types are not specified | -2 pts |
| prd-spec.md:97-106 | No validation rules stated per field — the document says "添加 NotDeleted" but never specifies what constitutes valid/invalid input for the new constraints | -3 pts |
| prd-user-stories.md:14-16 | Story 1 first AC: "Given 角色列表中存在一个已删除的角色" describes a state, not an action trigger — the Given should describe a pre-condition setup (e.g., "Given an admin has deleted a role") | -1 pts |
| prd-user-stories.md:60-67 | Story 4 has only one AC — missing edge case for HasPermission vs GetUserTeamPermissions distinction | -1 pts |
| prd-spec.md:58 | "已在本次会话中修复" — "本次会话" is session-scoped language inappropriate for a PRD; scope should be version/iteration-scoped | -1 pts |
| prd-spec.md:57 | "前端改动（后端修复即可）" listed as out-of-scope but no frontend interaction is described in specs, making this scope item disconnected from the rest of the document | -1 pts |

---

## Attack Points

### Attack 1: [Functional Specs — missing validation rules and field-level detail]

**Where**: prd-spec.md lines 97-106, the entire functional description table:
> "关联改动点" column lists method names like `Scopes(NotDeleted)` but never specifies field types, input/output schemas, or validation rules per field.

**Why it's weak**: The rubric requires "validation rules stated per field/button (not just 'validate input')". The document describes *what* methods change but never specifies *how* the NotDeleted scope validates its inputs, what SQL conditions it generates, what happens when the scope is applied to a table without deleted_flag, or what error conditions exist. This is the biggest gap — the functional specs read like a change log, not a specification.

**What must improve**: Add a structured table per module with columns for: field/method, type, validation rules (e.g., "deleted_flag must be 0 or 1", "deleted_time must be null when deleted_flag=0"), error responses, and edge cases. Specify what `NotDeleted` does in SQL terms (`WHERE deleted_flag = 0 OR deleted_flag IS NULL`). Add explicit validation rules for the unique index change (e.g., "item_code must be unique per main_item_key among non-deleted records").

### Attack 2: [Functional Specs — table structure does not match rubric requirements]

**Where**: prd-spec.md lines 97-106, the functional description section titled "5.4 关联性需求改动":
> The table has columns: 序号 | 涉及模块 | 功能模块 | 关联改动点 | 更改后逻辑说明

**Why it's weak**: The rubric specifies "Tables complete (list page 7 elements, button 4 elements, form 2 elements)". The current table is a flat 8-row change inventory with no differentiation between list views, button actions, and form fields. While this is a backend-only fix with no UI, the rubric's table completeness criteria expects structured breakdowns with field-level granularity. Each repo method should have its own row with: field name, field type, source, description, and validation rule — not a single summary row per repo.

**What must improve**: Restructure the functional description into per-method tables. For each affected method, list: method name, parameters (name, type, validation), return type, soft-delete filter behavior, and error cases. This gives developers an actionable specification rather than a summary.

### Attack 3: [User Stories — ACs lack boundary conditions and edge cases]

**Where**: prd-user-stories.md, Story 4 (lines 60-67):
> "Given 团队成员已被软删除（deleted_flag=1），且其角色包含某权限码 / When 系统检查该用户是否有该权限 / Then 返回 false，不包含该权限"

**Why it's weak**: Story 4 has only one AC covering HasPermission. But the functional spec (prd-spec.md line 105) lists three separate Role repo methods: HasPermission, GetUserTeamPermissions, CountMembersByRoleID. The AC only covers the permission check case. There is no AC for: (1) GetUserTeamPermissions excluding deleted members' permissions, (2) CountMembersByRoleID not counting deleted members, (3) what happens when a member is deleted mid-session. Additionally, Story 1's first AC uses passive state ("存在一个已删除的角色") rather than an active setup action.

**What must improve**: Add at least 2 more ACs to Story 4 covering GetUserTeamPermissions and CountMembersByRoleID. Rewrite Story 1's first AC to use active Given setup. Add boundary-condition ACs: e.g., "Given a member is soft-deleted while their permission check is cached — When the cache expires — Then the permission is revoked."

---

## Previous Issues Check

*Not applicable — this is iteration 1.*

---

## Verdict

- **Score**: 71/100
- **Target**: 90/100
- **Gap**: 19 points
- **Action**: Continue to iteration 2 — focus on (1) restructuring functional specs with per-method validation rules and field types, (2) expanding user story ACs with boundary cases and full method coverage, (3) adding a data-flow or request-flow diagram alongside the development workflow diagram.

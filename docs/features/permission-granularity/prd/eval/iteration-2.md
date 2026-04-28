---
date: "2026-04-28"
doc_dir: "docs/features/permission-granularity/prd/"
iteration: "2"
target_score: ""
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 2

**Score: 90/100** (target: —)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  17      │  20      │ ⚠️          │
│    Background three elements │   7/7    │          │            │
│    Goals quantified          │   7/7    │          │            │
│    Logical consistency       │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  18      │  20      │ ⚠️          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   7/7    │          │            │
│    Decision + error branches │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  17      │  20      │ ⚠️          │
│    Tables complete           │   7/7    │          │            │
│    Field descriptions clear  │   6/7    │          │            │
│    Validation rules explicit │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  20      │  20      │ ✅          │
│    Coverage per user type    │   7/7    │          │            │
│    Format correct            │   7/7    │          │            │
│    AC per story              │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  18      │  20      │ ⚠️          │
│    In-scope concrete         │   6/7    │          │            │
│    Out-of-scope explicit     │   6/7    │          │            │
│    Consistent with specs     │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  90      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md §需求目标 | Goal says "10 条受影响路由" but §5.2 route table lists 14 rows | -3 pts |
| prd-spec.md §业务流程图 | Flow A and Flow B have no error branch for API call failure (network/5xx); only permission-denied (403) branches are shown | -2 pts |
| prd-spec.md §5.2 | `POST /admin/users` is labeled "给用户分配角色" but the conventional semantics of `POST /admin/users` is "create user"; no explanation given. `DELETE /admin/users/:userId` maps to `user:update` with no rationale | -1 pt |
| prd-ui-functions.md §UI Function 3 Validation Rules | `description` field has a 200-char constraint in Form Spec but no validation rule specifying the error message or behavior on overflow | -1 pt |
| prd-ui-functions.md §UI Function 1 Validation Rules | No validation rule for the member search empty-state: what happens when the search query returns zero candidates | -1 pt |
| prd-spec.md §Scope In | `VerifyPresetRoleCodes` test update is listed as a deliverable in §5.5 but absent from the Scope In checklist | -1 pt |
| prd-spec.md §5.3 | "member 角色不变" in Scope In conflicts with §5.3 prose "管理员可按需单独授予 `role:read`" — the distinction between default preset and manual grant is never stated in Scope | -1 pt |

---

## Attack Points

### Attack 1: Background & Goals — Route count in goal contradicts the spec

**Where**: `prd-spec.md §需求目标` — "后端接口权限精确覆盖 | **10 条受影响路由**全部绑定新权限码，0 条残留旧码". `prd-spec.md §5.2` route table has **14 rows**.

**Why it's weak**: The quantified goal is the primary acceptance criterion for this dimension. A reviewer verifying the goal against the spec will immediately find the mismatch: 14 routes are listed (GET/POST/PUT×3/DELETE for users, GET×2/POST/PUT/DELETE for roles, GET /admin/teams, GET /admin/permissions) but the goal says 10. This is not a rounding issue — it's a 40% discrepancy. If the goal is used as a test checklist, 4 routes will be missed. If the spec is wrong, the goal is unverifiable.

**What must improve**: Count the rows in §5.2 and update the goal to "14 条受影响路由". Alternatively, if some routes are intentionally excluded (e.g., the three `user:update`-unchanged routes), explain the counting logic in a footnote.

---

### Attack 2: Flow Diagrams — API error branches absent from Flow A and Flow B

**Where**: `prd-spec.md §业务流程图` — Flow A: after `CheckPerm →|是| FetchRoles[GET /admin/roles]`, there is no branch for when the API call itself fails (5xx, timeout). Flow B: after `CheckList →|是| FetchUsers[GET /admin/users]`, same gap.

**Why it's weak**: The diagrams model only the permission-check failure path (403), not the infrastructure failure path. For a feature whose correctness depends on real-time DB permission loading (explicitly stated: "每次请求从 DB 实时加载权限码"), the failure mode of the permission-loading API is the most operationally relevant error branch. A developer implementing the role dropdown has no spec guidance on what to render when `GET /admin/roles` returns 500 — the States table in UI Function 1 has an `error` state ("显示加载失败，可重试") but the flow diagram never reaches it.

**What must improve**: Add an error branch after each API call node in Flow A and Flow B: `FetchRoles -->|失败| ErrorState[显示加载失败，可重试]` and equivalent for Flow B. This connects the flow diagram to the States table in prd-ui-functions.md.

---

### Attack 3: Functional Specs — Route table has semantic errors; validation rules have gaps

**Where (route table)**: `prd-spec.md §5.2` — `POST /admin/users` is described as "给用户分配角色" with permission `user:assign_role`. Conventionally `POST /admin/users` creates a user. The spec gives no explanation. `DELETE /admin/users/:userId` maps to `user:update` — deleting a resource with an update permission is a design smell that goes unacknowledged.

**Where (validation rules)**: `prd-ui-functions.md §UI Function 3 Validation Rules` — `description` field has constraint "最多 200 个字符" in Form Spec but the Validation Rules section specifies no error message or UI behavior for overflow (contrast with `roleName` which has "输入框下方显示'角色名称不能超过 50 个字符'"). `prd-ui-functions.md §UI Function 1 Validation Rules` — no rule for when member search returns zero results (the `memberId` required rule only covers "未选择成员", not "search returned nothing to select from").

**Why it's weak**: The route table is the primary handoff artifact to backend engineers. An unexplained `POST /admin/users` → assign-role mapping will cause confusion or a wrong implementation. The validation rule gaps mean frontend engineers must invent behavior for description overflow and empty search results — two common edge cases that will produce inconsistent UX across implementations.

**What must improve**: Add a note to the `POST /admin/users` row explaining why this endpoint is used for role assignment (or correct it to the right endpoint). Add a note explaining `DELETE → user:update`. Add a Validation Rule for `description` overflow. Add a Validation Rule for member search returning zero results (e.g., "显示'未找到匹配成员'，提交按钮保持禁用").

---

## Previous Issues Check

| Issue from Iteration 1 | Status |
|------------------------|--------|
| "两条主要业务流程" but listed three flows | ✅ Fixed — now reads "三条主要业务流程" |
| Flow C (data migration) had no Mermaid diagram | ✅ Fixed — full flowchart added |
| No Validation Rules in UI Functions 1 and 2 | ✅ Fixed — both sections now have Validation Rules |
| "添加成员" dialog form fields never specified | ✅ Fixed — Form Spec table added to UI Function 1 |
| Role create/edit form entirely absent | ✅ Fixed — Form Spec added to UI Function 3 |
| In-scope frontend item named no specific components | ✅ Fixed — now names AddMemberDialog, cross-references prd-ui-functions.md |
| Out-of-scope thin (missing caching, notification) | ✅ Fixed — both now explicitly excluded |

All 6 iteration-1 attack points addressed.

---

## Verdict

- **Score**: 90/100
- **Target**: —/100
- **Gap**: —
- **Action**: Significant improvement from iteration 1 (83→90). Remaining gaps: route count mismatch in goals (-3), missing API error branches in flow diagrams (-2), route table semantic issues and validation rule gaps in functional specs (-3), scope checklist missing one deliverable and one ambiguity (-2). Weakest remaining dimension is Functional Specs (17/20) and Background & Goals (17/20).

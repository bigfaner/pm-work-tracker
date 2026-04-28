---
date: "2026-04-28"
doc_dir: "docs/features/permission-granularity/prd/"
iteration: "1"
target_score: ""
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 1

**Score: 83/100** (target: —)

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
│ 2. Flow Diagrams             │  16      │  20      │ ⚠️          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   6/7    │          │            │
│    Decision + error branches │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  13      │  20      │ ❌          │
│    Tables complete           │   5/7    │          │            │
│    Field descriptions clear  │   5/7    │          │            │
│    Validation rules explicit │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  20      │  20      │ ✅          │
│    Coverage per user type    │   7/7    │          │            │
│    Format correct            │   7/7    │          │            │
│    AC per story              │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  17      │  20      │ ⚠️          │
│    In-scope concrete         │   6/7    │          │            │
│    Out-of-scope explicit     │   5/7    │          │            │
│    Consistent with specs     │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  83      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md §流程说明 | "本次变更涉及两条主要业务流程" but then lists three flows (A, B, C) — factual inconsistency | -3 pts |
| prd-spec.md §流程说明 | Flow C (data migration) described as a main flow but has no Mermaid diagram | -1 pt |
| prd-ui-functions.md §UI Function 1, 2 | No Validation Rules section in UI Functions 1 and 2 | -3 pts |
| prd-ui-functions.md (all) | "添加成员" dialog form fields never specified; role create/edit form entirely absent | -2 pts |
| prd-spec.md §Scope In | "内部文档更新：`docs/` 中所有列出权限码的文档" — no specific files named | -1 pt |
| prd-spec.md §Scope Out | Only 4 out-of-scope items; permission caching, notification to affected users, and API versioning strategy not addressed | -2 pts |

---

## Attack Points

### Attack 1: Flow Diagrams — Flow C (data migration) has no diagram despite being called a main flow

**Where**: `prd-spec.md §流程说明` — "本次变更涉及两条主要业务流程" followed by listing flows A, B, **and C**. The Mermaid diagram covers only A and B. Flow C is described in prose only.

**Why it's weak**: The data migration is the highest-risk part of this change — a two-step rollout with atomic writes, a CI grep gate, and a rollback script. It is the one flow where a mistake causes production data corruption. Describing it only in a table with no state diagram means reviewers cannot verify the sequencing logic (e.g., the atomic write of new `user:read` during step one) or spot race conditions. The claim of "两条主要业务流程" while listing three is also a factual error that undermines trust in the document's precision.

**What must improve**: Add a Mermaid flowchart for Flow C covering: pre-migration snapshot → transaction start → `user:manage_role` conversion → `user:read` conversion + atomic new `user:read` grant → commit/rollback → CI grep gate → step two conditional grant. Fix the "两条" claim to "三条".

---

### Attack 2: Functional Specs — Form specs entirely absent; validation rules cover only 1 of 3 UI functions

**Where**: `prd-ui-functions.md` — UI Function 1 (角色下拉选择器) has a Data Requirements table with only `roleId` and `roleName`. The "添加成员" dialog form (member search/select field, role selector, submit button) is never specified. UI Functions 1 and 2 have no Validation Rules section at all.

**Why it's weak**: The rubric requires validation rules per field/button. The "添加成员" dialog is the primary user-facing interaction in this feature (Flow A), yet there is no spec for: what happens if no role is selected on submit, whether the member field is required, what the role selector displays when the list is empty (vs. loading vs. error). UI Function 2 (用户管理页入口) has only a two-row States table — it specifies `visible`/`hidden` but says nothing about what happens during the permission-load latency window (flash of content?). UI Function 3 has validation rules but they are for the delete button only; the edit form for role name/description/permission codes has zero validation rules.

**What must improve**: Add a form spec for the "添加成员" dialog (fields, types, required/optional, validation). Add a form spec for role create/edit (role name constraints, permission code multi-select behavior, empty-state handling). Add a Validation Rules section to UI Functions 1 and 2.

---

### Attack 3: Scope Clarity — Out-of-scope is thin and frontend scope is under-specified

**Where**: `prd-spec.md §Scope Out` — four items listed: UI改版 (×2), 新增API接口, 审计日志. `prd-spec.md §Scope In` item 5: "前端权限判断逻辑更新：`user:manage_role` 替换为 `role:*`，`user:read` 替换为 `user:list`/`user:read`" with no specific component or file names.

**Why it's weak**: The out-of-scope section does not address: (1) whether permission caching is in scope — the spec says "无缓存层" in performance requirements, but this is buried in §其他说明, not in Scope; (2) whether notifying existing users whose permissions change is in scope; (3) whether the CI grep assertion is a new CI job or an addition to an existing one — this is a deliverable but it's not in the In Scope checklist. The frontend in-scope item names no specific files or components, making it impossible to verify completeness at review time. The `prd-ui-functions.md` covers 3 UI functions but the in-scope item doesn't reference them.

**What must improve**: Expand Out of Scope to explicitly exclude: permission caching, user notification on permission change, CI job creation (or move it to In Scope if it's a deliverable). Update the frontend in-scope item to reference the specific components/pages affected (team management page, user management nav, role management page). Cross-reference the UI functions doc.

---

## Previous Issues Check

<!-- Only for iteration > 1 — N/A for iteration 1 -->

---

## Verdict

- **Score**: 83/100
- **Target**: —/100
- **Gap**: —
- **Action**: Iteration 1 complete. Weakest dimension is Functional Specs (13/20) — form specs and validation rules are the primary gap. Flow Diagrams (16/20) need Flow C diagram. Address these before proceeding to tech design.

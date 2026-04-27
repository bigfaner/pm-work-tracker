---
date: "2026-04-27"
doc_dir: "docs/features/user-management-reset-delete/prd/"
iteration: 1
target_score: 90
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 1

**Score: 88/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 1. Background & Goals        │  20      │  20      │ ✅         │
│    Background three elements │  7/7     │          │            │
│    Goals quantified          │  7/7     │          │            │
│    Logical consistency       │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 2. Flow Diagrams             │  19      │  20      │ ✅         │
│    Mermaid diagram exists    │  7/7     │          │            │
│    Main path complete        │  7/7     │          │            │
│    Decision + error branches │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 3. Functional Specs          │  16      │  20      │ ⚠️         │
│    Tables complete           │  5/7     │          │            │
│    Field descriptions clear  │  6/7     │          │            │
│    Validation rules explicit │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 4. User Stories              │  15      │  20      │ ⚠️         │
│    Coverage per user type    │  5/7     │          │            │
│    Format correct            │  6/7     │          │            │
│    AC per story              │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 5. Scope Clarity             │  18      │  20      │ ✅         │
│    In-scope concrete         │  7/7     │          │            │
│    Out-of-scope explicit     │  7/7     │          │            │
│    Consistent with specs     │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ TOTAL                        │  88      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md:5.1 列表字段 | List page table has 6 fields instead of required 7. Missing column — e.g., creation time, role, or last login. | -2 pts |
| prd-spec.md:5.1 列表字段 | "状态" field description says "enabled / disabled" but sample data shows "启用/禁用". Unclear if this is a display mapping or raw API value. | -1 pt |
| prd-spec.md:5.3 校验规则 | Password max length (64 chars) is in form field spec but not enforced in the validation rules table. Only lower bound is validated. | -1 pt |
| prd-user-stories.md:59-77 | Story numbering jumps from 4 to 6. Regular user story is unlabeled and appended to Story 6 without a proper heading, making it appear orphaned. | -2 pts |
| prd-user-stories.md:70-77 | Regular user story not given its own story number or section heading. It reads as an afterthought bolted onto Story 6. | -2 pts |
| prd-user-stories.md: Stories 1,3,6 | Acceptance criteria only cover happy paths. No AC for backend error scenarios (network failure during reset, user already deleted when delete attempted, API returning 500). | -2 pts |
| prd-spec.md: 流程图 + Scope | Scope says "JWT 失效" (implies active invalidation). Flow diagram shows "该用户下次请求时 JWT 被拒绝" (implies passive rejection). These are different implementation approaches and the document doesn't reconcile them. | -1 pt |
| prd-spec.md: Scope In | Scope item "前端用户列表新增「重置密码」「删除」操作入口（仅超级管理员可见）" frames this as UI-only. The actual need includes API-level enforcement (其他说明 mentions it), but scope doesn't surface this. | -1 pt |
| prd-spec.md: 流程图 | Error branches only cover 403. No branch for generic server errors (500), user-not-found, or network timeout in either flow. | -1 pt |

---

## Attack Points

### Attack 1: User Stories — Regular user story is structurally broken and numbering has a gap

**Where**: prd-user-stories.md lines 59-77: Story 6 header is "创建用户后快速复制账号与密码" but lines 70-77 contain an entirely separate story for regular users with its own As a/I want to/So that and AC block, appended without a heading.

**Why it's weak**: The document jumps from Story 4 to Story 6 — Story 5 does not exist. The regular user story (a distinct user type from the background section) is buried at the bottom of Story 6 with no section heading, no story number, and no visual separation. A developer scanning the document could easily miss it. The rubric requires "one story per target user type" — the background defines two types (super admin and regular user), but the regular user's story is structurally invisible.

**What must improve**: Extract the regular user story into its own properly numbered section (Story 5). Give it a clear heading, separate it from Story 6 with a `---` divider, and ensure both user types from the background have top-level story entries.

### Attack 2: Functional Specs — List page table is incomplete and field descriptions have ambiguities

**Where**: prd-spec.md Section 5.1 "列表字段" table has 6 rows (用户名, 显示名, 邮箱, 状态, 所属团队, 操作). The rubric expects 7 elements for the list page.

**Why it's weak**: The table is one field short. Common expected columns for a user management list include creation time, last login time, or role — none of which appear. Additionally, the "状态" field description says "enabled / disabled" (English API values) while the sample data column shows "启用/禁用" (Chinese display text). The document never states whether the frontend maps these values or if the API returns localized strings. This ambiguity forces the implementer to guess.

**What must improve**: Add the missing 7th column to the list page field table (e.g., 创建时间 or 角色). Clarify in the "状态" field description whether "enabled/disabled" is the raw API value and "启用/禁用" is the frontend display mapping, or document the exact transformation.

### Attack 3: User Stories — Acceptance criteria lack error-scenario coverage

**Where**: prd-user-stories.md Stories 1, 3, and 6. For example, Story 1 AC: "Then 弹窗关闭，显示'密码已重置'提示；该用户使用新密码可正常登录" — this is the only AC, covering exclusively the happy path.

**Why it's weak**: Stories 1 (reset password), 3 (soft delete), and 6 (copy credentials) each have only a single AC that describes the success case. No AC addresses what happens when: the backend returns a 500 error, the network times out, the target user was already deleted between list load and action, or the password reset API rejects the request for a reason other than auth. Story 2 is the only story with a failure-path AC (validation errors). The rubric expects thorough acceptance criteria — real-world testing against these stories would leave error handling unverified.

**What must improve**: Add at least one error-scenario AC to Stories 1, 3, and 6. For example, Story 1 should include: "Given 重置密码弹窗已打开并提交 / When 后端返回非 200 响应 / Then 弹窗保持打开，显示后端返回的错误信息，用户可修改后重试". Story 3 should include an AC for attempting to delete a user that was already deleted (stale UI state).

---

## Previous Issues Check

<!-- Only for iteration > 1 — not applicable for iteration 1 -->

---

## Verdict

- **Score**: 88/100
- **Target**: 90/100
- **Gap**: 2 points
- **Action**: Continue to iteration 2. Focus on: (1) restructuring user stories — extract regular user story into proper numbered section, (2) completing list page table to 7 fields and clarifying status field mapping, (3) adding error-scenario ACs to Stories 1, 3, and 6.

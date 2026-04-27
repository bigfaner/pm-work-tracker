---
date: "2026-04-27"
doc_dir: "docs/features/user-management-reset-delete/prd/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 2

**Score: 95/100** (target: 90)

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
│ 3. Functional Specs          │  17      │  20      │ ⚠️         │
│    Tables complete           │  6/7     │          │            │
│    Field descriptions clear  │  6/7     │          │            │
│    Validation rules explicit │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 4. User Stories              │  20      │  20      │ ✅         │
│    Coverage per user type    │  7/7     │          │            │
│    Format correct            │  7/7     │          │            │
│    AC per story              │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 5. Scope Clarity             │  19      │  20      │ ✅         │
│    In-scope concrete         │  7/7     │          │            │
│    Out-of-scope explicit     │  7/7     │          │            │
│    Consistent with specs     │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ TOTAL                        │  95      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md: 流程图 | Error branches cover only 403. No branch for generic server errors (500), network timeout, or user-not-found in either flow. User stories now cover these scenarios but the diagram does not reflect them. | -1 pt |
| prd-spec.md:5.2 按钮操作 | Button spec is distributed across three separate sub-tables (状态条件, 校验规则, 数据处理逻辑) instead of a unified button inventory table with 4 attributes per button (name, type, condition, action). The "delete" button has no explicit validation row in the 校验规则 table. | -1 pt |
| prd-spec.md:5.1 列表字段 | Field descriptions lack a "source" column. The table has 字段名称/类型/说明 but does not state where each field's data comes from (e.g., which API endpoint, which model property). The generic "数据来源：后端用户管理接口" above does not map individual fields. | -1 pt |
| prd-spec.md:5.3 校验规则 | Password max length (64 chars) is stated in the form field spec table but not enforced in the validation rules table. Only the lower bound (≥8) is validated. No validation rule row for "密码长度 ≤ 64 位". | -1 pt |
| prd-spec.md: Scope + 5.5 | Scope item says "JWT 失效" (implies active token blacklisting). Section 5.5 says "deleted_at 不为空时拒绝登录并使 JWT 失效". The flow diagram node D11 says "该用户后续请求 JWT 被拒" (could be passive). Active JWT invalidation vs passive deleted_at check are architecturally different — the document does not reconcile which mechanism is used. | -1 pt |

---

## Attack Points

### Attack 1: Functional Specs — Password upper-bound validation is missing from rules table

**Where**: prd-spec.md Section 5.3 form field spec: "新密码 | 密码输入框 | 是 | 8–64 | 至少8位，含字母和数字". The 校验规则 table immediately below has 3 rows covering empty check, strength (≥8, letters+digits), and confirmation match — but no row for max length (≤64).

**Why it's weak**: The form spec declares a max length of 64 characters but the validation rules table does not enforce it. A developer implementing the validation will follow the rules table and miss the upper bound, allowing users to submit passwords longer than 64 characters. The field spec and validation rules are inconsistent with each other.

**What must improve**: Add a 4th row to the 校验规则 table: "密码长度 ≤ 64 | 提交 | 密码不能超过64位 | 新密码字段下方红色文字". Alternatively, remove the max length from the field spec if it is only enforced by the input control's maxlength attribute rather than a business rule.

### Attack 2: Flow Diagrams — Error branches still incomplete despite user stories being fixed

**Where**: prd-spec.md Mermaid flowchart — both the reset password and delete flows have only a single error branch: "非 super_admin → 返回 403". No nodes exist for server errors (500), network timeout, or user-not-found.

**Why it's weak**: The user stories (Story 1 AC2, Story 3 AC2) now explicitly cover backend error scenarios (500, network timeout, user already deleted). But the flow diagram — the canonical visual reference that developers and QA use — does not reflect these paths. A developer reading only the diagram would miss the error-handling flows that the stories require. The rubric criterion "Decision points + error branches covered" expects the diagram itself to have error/exception branches, not just the prose.

**What must improve**: Add at least one generic error branch to each flow in the diagram. For example, after RP5 "提交到后端" add a diamond "后端响应" with branches for "成功" (→ update hash) and "错误 (500/404/超时)" (→ show error in dialog, keep open). Similarly for the delete flow after D6.

### Attack 3: Scope Clarity — JWT invalidation mechanism is ambiguous across sections

**Where**: prd-spec.md Scope In-Scope item 2: "软删除用户（列表过滤、登录拦截、JWT 失效）". Section 5.5 row 1: "增加软删除状态校验：deleted_at 不为空时拒绝登录并使 JWT 失效". Flow diagram node D11: "该用户后续请求 JWT 被拒".

**Why it's weak**: "JWT 失效" implies active invalidation (token blacklisting, changing a version, or forcing re-auth). "deleted_at 不为空时拒绝" implies passive checking (middleware reads the user record on each request). These are architecturally different: active invalidation requires infrastructure (token blacklist, Redis, etc.); passive checking requires a DB query on every request. The document mentions both without committing to one. A developer reading scope sees "JWT 失效" and might implement active blacklisting, while the flow diagram suggests passive rejection.

**What must improve**: Choose one mechanism and use it consistently. Either: (a) state explicitly that JWT invalidation is achieved by checking deleted_at in the auth middleware on each request (passive), or (b) describe the active invalidation mechanism (e.g., token version bump, blacklist entry). Remove or clarify the contradictory language so all three sections (scope, 5.5, flow diagram) align.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Story 5 missing, regular user story appended to Story 6 | ✅ Fixed | prd-user-stories.md now has Story 5 "普通用户不可见敏感操作入口" with its own heading, As a/I want/So that format, and AC. Story 6 is the separate copy-credentials story. Sequential numbering 1-6. |
| Attack 2: List page missing 7th column, status field ambiguous | ✅ Fixed | prd-spec.md Section 5.1 列表字段 table now has 7 rows including "创建时间". Status field now reads: "后端返回 `enabled` / `disabled`，前端映射为「启用」/「禁用」显示" — explicit mapping. |
| Attack 3: Missing error-scenario acceptance criteria | ✅ Fixed | Story 1 AC2 covers backend 500/timeout. Story 3 AC2 covers stale delete (user already deleted). Story 6 AC2 covers clipboard API failure. All use Given/When/Then format. |

---

## Verdict

- **Score**: 95/100
- **Target**: 90/100
- **Gap**: +5 points above target
- **Action**: Target reached. All three iteration-1 attacks properly addressed. Remaining deductions are minor: password max-length validation gap, incomplete flow diagram error branches, and JWT mechanism ambiguity. Document is ready for design phase.

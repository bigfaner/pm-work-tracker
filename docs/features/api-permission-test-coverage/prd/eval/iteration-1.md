---
date: "2026-04-28"
doc_dir: "docs/features/api-permission-test-coverage/prd/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 1

**Score: 75/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  17      │  20      │ ✅          │
│    Background three elements │   6/7    │          │            │
│    Goals quantified          │   6/7    │          │            │
│    Logical consistency       │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  16      │  20      │ ✅          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   5/7    │          │            │
│    Decision + error branches │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  12      │  20      │ ❌          │
│    Tables complete           │   4/7    │          │            │
│    Field descriptions clear  │   5/7    │          │            │
│    Validation rules explicit │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  16      │  20      │ ✅          │
│    Coverage per user type    │   5/7    │          │            │
│    Format correct            │   6/7    │          │            │
│    AC per story              │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  14      │  20      │ ⚠️          │
│    In-scope concrete         │   5/7    │          │            │
│    Out-of-scope explicit     │   5/7    │          │            │
│    Consistent with specs     │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  75      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| 需求目标 Goal 2 | "代表性端点" used without specifying count or which endpoints | -2 pts (vague language) |
| I-A table / I-C table / 监控需求 | "非403" used as validation criterion — a 500 is also "非403" | -2 pts (vague language) |
| 监控需求 vs Scope | CI assertion deliverable ("验证codes.go中每个权限码在测试矩阵中均有覆盖") appears in 监控需求 but is absent from In Scope list | -3 pts (inconsistency) |
| 人员 section | "开发者" and "代码审查者" stated without any specifics about team context or current pain | -1 pt (vague) |

---

## Attack Points

### Attack 1: Functional Specs — "非403" is not a valid test assertion

**Where**: I-A table: `superadmin | 非 403 | 非 403 | 非 403 | 非 403 | 非 403` and I-C: `superadmin → 调用任意受保护端点 → 非 403（绕过权限检查）`

**Why it's weak**: "非403" as a success criterion is dangerously loose. A 500 Internal Server Error, a 404 Not Found due to missing test fixture data, or a 400 Bad Request all satisfy "非403" — none of them prove the permission check passed. The spec never defines what the actual expected status code IS for a successful authorized request. Integration tests built on this spec will produce false positives whenever test data is missing.

**What must improve**: Replace every "非403" with the concrete expected code. For endpoints that mutate data, the expected code is 200 or 201. For endpoints where test data may not exist, specify the fixture setup required and the exact expected code. At minimum, add a note: "非403 means 200 or 201; a 404 or 500 is a test setup failure, not a pass."

---

### Attack 2: Functional Specs — Tables missing test-execution columns

**Where**: U1 table lists endpoint, required permission, pm/member flags — but stops there. 关联性需求改动 table lists only "涉及项目" and "关联改动点."

**Why it's weak**: The U1 table tells you WHAT to test but not WHERE to write it or HOW to name it. There is no column for: test file path, test function name pattern, or mock setup required. A developer reading this table cannot start writing tests without making decisions the PRD should have made. The 关联性需求改动 table says "在现有文件中追加表驱动测试" but doesn't specify which test function names to use, making the deliverable unverifiable in code review.

**What must improve**: Add columns to the U1 table: `测试文件` (e.g., `middleware/permission_test.go`), `测试函数` (e.g., `TestPermMiddleware_MainItemArchive`), and `Mock 依赖` (what service mocks are needed). Without these, the "24 cases" target in scope cannot be verified against the actual test file.

---

### Attack 3: Scope — CI assertion deliverable is orphaned in 监控需求

**Where**: 监控需求 section: "CI 中新增断言：验证 `codes.go` 中每个权限码在测试矩阵中均有覆盖（防止新增权限码时遗漏测试）"

**Why it's weak**: This is a concrete deliverable — writing a CI assertion that cross-references `codes.go` against the test matrix — but it appears only in 监控需求, not in the In Scope list, not in 功能描述, and has no corresponding User Story or Acceptance Criteria. It is the most forward-looking and highest-value item in the entire PRD (it prevents the exact regression described in 需求背景), yet it has zero traceability. A developer reading only the Scope section will not build it. A code reviewer cannot verify it was done.

**What must improve**: Promote this to a named scope item (e.g., "I-D: 权限码覆盖率 CI 断言") with a functional spec entry describing the assertion logic, and add a corresponding User Story with AC. Alternatively, if it is truly out of scope for this iteration, move it explicitly to Out of Scope with a note that it is deferred.

---

## Previous Issues Check

<!-- Only for iteration > 1 — N/A for iteration 1 -->

---

## Verdict

- **Score**: 75/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Three concrete issues to address before next iteration: (1) replace all "非403" with exact expected codes, (2) add test-execution columns to U1 table, (3) promote CI assertion to a named scope item with AC.

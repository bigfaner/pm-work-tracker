---
date: "2026-04-28"
doc_dir: "docs/proposals/api-permission-test-coverage/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 1

**Score: 72/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  13      │  20      │ ⚠️          │
│    Problem clarity           │   6/7    │          │            │
│    Evidence provided         │   4/7    │          │            │
│    Urgency justified         │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Solution Clarity          │  15      │  20      │ ⚠️          │
│    Approach concrete         │   6/7    │          │            │
│    User-facing behavior      │   4/7    │          │            │
│    Differentiated            │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  11      │  15      │ ⚠️          │
│    Alternatives listed (≥2)  │   4/5    │          │            │
│    Pros/cons honest          │   3/5    │          │            │
│    Rationale justified       │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  13      │  15      │ ✅          │
│    In-scope concrete         │   5/5    │          │            │
│    Out-of-scope explicit     │   5/5    │          │            │
│    Scope bounded             │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Risk Assessment           │   9      │  15      │ ❌          │
│    Risks identified (≥3)     │   4/5    │          │            │
│    Likelihood + impact rated │   1/5    │          │            │
│    Mitigations actionable    │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  11      │  15      │ ⚠️          │
│    Measurable                │   4/5    │          │            │
│    Coverage complete         │   3/5    │          │            │
│    Testable                  │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  72      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Risks section | No likelihood or impact ratings on any of the 3 risks | -4 pts (sub-criterion) |
| Problem section | Problem asserted without data — no test count, no coverage %, no incident | -3 pts (sub-criterion) |
| Success Criteria | Integration test C scenarios (superadmin bypass, empty permission role) described in solution but absent from success criteria | -2 pts (sub-criterion) |
| Scope section | No time or effort estimate — scope is enumerated but not bounded by a timeframe | -2 pts (sub-criterion) |
| Alternatives section | "do nothing" never listed as an explicit alternative; pros/cons are one-liners with no depth | -2 pts (sub-criterion) |

---

## Attack Points

### Attack 1: Risk Assessment — likelihood and impact ratings are entirely absent

**Where**: "1. **测试 DB 初始化成本**... 2. **权限矩阵维护**... 3. **Handler mock 复杂度**"

**Why it's weak**: The rubric requires likelihood + impact ratings for each risk. None of the three risks carry any rating. Without ratings, a reader cannot prioritize mitigations or decide whether the risks are acceptable. Risk 1 (DB init cost) could be low-likelihood if `TestMain` is already used elsewhere in the project — or high-likelihood if it isn't. The proposal gives no signal. Risk 2 (matrix maintenance) is arguably the highest ongoing cost but is treated identically to the others. The section reads as a checklist, not an analysis.

**What must improve**: Add explicit likelihood (Low/Medium/High) and impact (Low/Medium/High) labels to each risk. Justify the ratings with one sentence of reasoning. Example: "Risk 1 — Likelihood: Medium (SQLite is fast but RBAC migration runs ~20 migrations); Impact: Low (CI slowdown, not a correctness issue)."

---

### Attack 2: Problem Definition — assertions without evidence

**Where**: "`permission_test.go` 只验证 mock context 注入，不验证真实角色配置下的 HTTP 响应" and "`rbac_test.go` 只覆盖迁移幂等性和角色 CRUD，没有任何'角色 X 调用端点 Y 得到 200/403'的断言"

**Why it's weak**: These are claims, not evidence. The proposal never shows the current test count, a coverage report, or a single concrete example of a permission bug that slipped through. "任何权限配置变更都可能静默引入回归" is a generic risk statement that applies to any untested code — it doesn't demonstrate that this specific gap has caused or nearly caused a real problem. A skeptical reader could ask: "Has a permission regression actually occurred? Is this gap theoretical or observed?" The proposal cannot answer that.

**What must improve**: Add at least one concrete data point: current test count in `permission_test.go` and `rbac_test.go`, a coverage percentage for the permission middleware, or a specific example of a route binding that is currently unverifiable by tests. Even "we recently added `perm("report:export")` to the export route and had no test to confirm it was wired correctly" would be sufficient.

---

### Attack 3: Success Criteria — incomplete coverage of in-scope scenarios

**Where**: Success criteria list 5 items, but Integration Test C defines three scenarios: "空权限角色：所有受保护端点返回 403", "superadmin：所有端点返回非 403（绕过权限检查）", "无效 token：返回 401（区分认证失败与授权失败）"

**Why it's weak**: The success criteria include "无权限返回 403、未认证返回 401，两者有明确区分的测试 case" — which covers the 401/403 distinction — but there is no criterion for the superadmin bypass scenario or the empty-permission-role scenario. Both are explicitly in scope (Integration Test C), yet neither appears in the success criteria. This means the proposal could be declared "done" without those scenarios ever being implemented, and no one would notice from the criteria alone.

**What must improve**: Add two criteria: (1) "superadmin 角色对所有受保护端点返回非 403，有至少 1 个测试 case 覆盖" and (2) "空权限角色对所有受保护端点返回 403，有至少 1 个测试 case 覆盖". Every in-scope scenario from the solution section must map to at least one verifiable success criterion.

---

## Previous Issues Check

N/A — iteration 1.

---

## Verdict

- **Score**: 72/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Strongest areas are Scope Definition (13/15) and Solution Clarity (15/20). Weakest area is Risk Assessment (9/15), dragged down almost entirely by the missing likelihood/impact ratings. Problem Definition loses points for zero empirical evidence. Success Criteria has a structural gap where in-scope scenarios from the solution are not reflected in any verifiable criterion.

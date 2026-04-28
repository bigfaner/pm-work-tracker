---
date: "2026-04-28"
doc_dir: "docs/proposals/api-permission-test-coverage/"
iteration: "3"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 3

**Score: 90/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  16      │  20      │ ⚠️          │
│    Problem clarity           │   6/7    │          │            │
│    Evidence provided         │   5/7    │          │            │
│    Urgency justified         │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Solution Clarity          │  17      │  20      │ ⚠️          │
│    Approach concrete         │   7/7    │          │            │
│    User-facing behavior      │   5/7    │          │            │
│    Differentiated            │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  13      │  15      │ ⚠️          │
│    Alternatives listed (≥2)  │   5/5    │          │            │
│    Pros/cons honest          │   4/5    │          │            │
│    Rationale justified       │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  15      │  15      │ ✅          │
│    In-scope concrete         │   5/5    │          │            │
│    Out-of-scope explicit     │   5/5    │          │            │
│    Scope bounded             │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Risk Assessment           │  14      │  15      │ ⚠️          │
│    Risks identified (≥3)     │   5/5    │          │            │
│    Likelihood + impact rated │   5/5    │          │            │
│    Mitigations actionable    │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  15      │  15      │ ✅          │
│    Measurable                │   5/5    │          │            │
│    Coverage complete         │   5/5    │          │            │
│    Testable                  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  90      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Problem section | "53 处 perm() 绑定" stated without methodology — no grep command, no file list, no way to verify | -1 pt (evidence) |
| Problem section | No coverage percentage shown — "11 个测试" and "21 个测试" are raw counts, not fractions of what should exist | -1 pt (evidence) |
| Problem section | "预计将引入更多权限码" is speculative — no count of expected new codes from bizkey-unification | -1 pt (urgency) |
| Solution section | Developer experience of running/maintaining tests not described — no CI integration path, no local run command | -2 pts (user-facing behavior) |
| Solution section | Cost comparison for chosen approach is asserted, not shown — "成本最低的有效覆盖方案" without comparing unit-only (~1d) vs. two-layer (~3d) ROI | -1 pt (differentiated) |
| Alternatives section | "do nothing" pros are incomplete — no mention that manual code review currently serves as a partial mitigation, making the trade-off appear more one-sided than it is | -1 pt (pros/cons honest) |
| Alternatives section | Rationale does not explain why integration-only was rejected on cost grounds — "调试成本高" is stated but not quantified against the 2-day integration estimate | -1 pt (rationale justified) |
| Risk section | Risk 1 description says "约需数秒" while Alternatives section says "2–5 秒" — internal inconsistency in the same document | -1 pt (mitigations actionable) |

---

## Attack Points

### Attack 1: Problem Definition — "53 处 perm() 绑定" is an unverifiable assertion with no coverage percentage

**Where**: "路由层共有 53 处 `perm()` 绑定，其中任何一处权限码误写...均无法被现有测试发现"

**Why it's weak**: The number 53 is stated as fact but no methodology is shown — no grep command, no file list, no reference to where it was counted. A skeptical reader cannot verify it. More importantly, the proposal never converts raw counts into a coverage ratio. "11 个测试" and "21 个测试" tell the reader how many tests exist, but not what fraction of the testable surface they cover. If 53 bindings exist and 32 are already covered, the urgency argument weakens considerably. If 0 are covered, that should be stated explicitly. The evidence section gives numbers without the denominator that makes them meaningful.

**What must improve**: Add the grep command or file reference used to arrive at "53 处" (e.g., `grep -r 'perm(' backend/internal/router/ | wc -l`). Then state the coverage ratio: "53 处绑定中，0 处有路由层测试覆盖（0%）" — that single line makes the problem undeniable.

---

### Attack 2: Solution Clarity — developer experience of running and maintaining these tests is absent

**Where**: "在各 handler 的 `_test.go` 文件中，针对权限敏感操作补充...表驱动测试" / "在 `rbac_test.go` 中新增一组测试，使用真实 SQLite DB 和完整路由"

**Why it's weak**: The proposal describes what the tests will verify (200/403 responses) but never describes how a developer interacts with them. There is no local run command for the new integration tests, no description of how they fit into the existing `go test ./...` invocation, and no mention of whether they require any setup (environment variables, test DB path, etc.). The success criteria reference `go test ./tests/integration/... -run TestRBACPermission` but the solution section never establishes that `tests/integration/` is a new directory being created — a reader cannot tell if this path already exists or needs to be scaffolded. For a proposal whose entire deliverable is test infrastructure, the developer experience of using that infrastructure is a first-class concern.

**What must improve**: Add a "How to run" subsection under the solution: the exact commands to run unit tests and integration tests locally, whether any setup is required (e.g., `TEST_DB=:memory:`), and confirm whether `tests/integration/` is a new directory or an existing one. This is one paragraph, not a major rewrite.

---

### Attack 3: Risk Assessment — internal inconsistency on DB init time; Risk 2 mitigation is underspecified

**Where (inconsistency)**: Risk 1 says "SQLite 在 CI 环境下每次冷启动约需数秒" while the Alternatives section says "SQLite 冷启动约需 2–5 秒"

**Where (underspecified)**: Risk 2 mitigation says "CI 中加断言验证覆盖率"

**Why it's weak**: The same document gives two different characterizations of the same metric — "约需数秒" (vague) vs. "2–5 秒" (quantified). A reader who reads both sections will notice the inconsistency and lose confidence in the precision of the analysis. Separately, "CI 中加断言验证覆盖率" in Risk 2's mitigation is the most important mitigation in the document (it prevents the entire class of "new permission added without test" failures) but it is stated in 9 characters with no mechanism. What tool? What assertion? A custom Go test that compares `codes.go` entries against the test matrix? A linter rule? Without specifying the mechanism, this mitigation cannot be acted on.

**What must improve**: Align the DB init time to "2–5 秒" in the risk section (copy from alternatives). For Risk 2, expand the mitigation to specify the mechanism: e.g., "在 `TestMain` 中读取 `codes.go` 的权限码列表，断言每个权限码在测试矩阵常量中均有对应 case，否则 `t.Fatal`".

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Alternatives Analysis — "do nothing" absent; pros/cons have no depth | ✅ | "do nothing" now explicitly listed with pros/cons. All alternatives now carry quantified estimates: "~1 天", "2–5 秒", "约 60 个 case/季度". Substantial improvement. |
| Problem Definition — urgency is theoretical with no concrete trigger | ✅ | Commit `3200bdc` cited as a real, already-occurred blind spot expansion. bizkey-unification cited as upcoming risk multiplier. Urgency is now grounded in evidence. |
| Scope Definition — no time or effort estimate | ✅ | "预计工作量：单元测试 ~1 个工作日，集成测试 ~2 个工作日，合计 ~3 个工作日" added directly in the Scope section. Fully addressed. |

---

## Verdict

- **Score**: 90/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: +13 points from iteration 2. All three previous attacks were fully addressed. Remaining gaps are concentrated in Problem Definition (unverifiable evidence, no coverage %), Solution Clarity (missing developer run experience), and a Risk Assessment internal inconsistency. These are polish-level issues, not structural ones — the proposal is now well-reasoned and actionable.

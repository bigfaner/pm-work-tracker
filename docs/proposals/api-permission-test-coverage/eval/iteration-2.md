---
date: "2026-04-28"
doc_dir: "docs/proposals/api-permission-test-coverage/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 2

**Score: 77/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  14      │  20      │ ⚠️          │
│    Problem clarity           │   6/7    │          │            │
│    Evidence provided         │   5/7    │          │            │
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
│ 5. Risk Assessment           │  12      │  15      │ ⚠️          │
│    Risks identified (≥3)     │   4/5    │          │            │
│    Likelihood + impact rated │   4/5    │          │            │
│    Mitigations actionable    │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  12      │  15      │ ⚠️          │
│    Measurable                │   4/5    │          │            │
│    Coverage complete         │   4/5    │          │            │
│    Testable                  │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  77      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Alternatives section | "do nothing" never listed as an explicit alternative — rubric requires it | -1 pt (sub-criterion) |
| Alternatives section | Pros/cons are one-liners with no depth — "速度快，但无法发现路由绑定错误" is a label, not an analysis | -2 pts (sub-criterion) |
| Problem section | Urgency is a theoretical consequence statement, not a concrete trigger — no upcoming feature, near-miss, or deadline | -3 pts (sub-criterion) |
| Scope section | No time or effort estimate — 24 unit cases + 3 integration scenarios enumerated but not bounded by a timeframe | -2 pts (sub-criterion) |
| Risk section | Risk 1 mitigation vague on timing — "约需数秒" is unquantified; no baseline measurement to validate the claim | -1 pt (sub-criterion) |

---

## Attack Points

### Attack 1: Alternatives Analysis — "do nothing" is absent and pros/cons have no depth

**Where**: "**仅补充单元测试**：速度快，但无法发现路由绑定错误" / "**仅补充集成测试**：覆盖更真实，但每个 case 需要完整 DB 初始化，运行慢" / "**全量路由矩阵测试**：对 40+ 路由 × 3 角色做穷举，维护成本过高"

**Why it's weak**: The rubric explicitly requires "do nothing" as a valid alternative. It is still absent after iteration 1. A reader cannot evaluate the cost of inaction — what is the actual risk of shipping without these tests? Beyond the missing alternative, every listed option gets a single-sentence trade-off. "运行慢" for integration-only has no quantification (how slow? 10s? 5min?). "维护成本过高" for full-matrix has no estimate (how many extra test cases? how many hours per new permission?). These are labels, not analyses. A skeptical reader could reasonably disagree with every trade-off and find no evidence to counter them.

**What must improve**: Add "do nothing" as an explicit alternative with its own pros/cons (e.g., "pro: zero implementation cost; con: any perm() miswrite is undetectable until production"). Expand each existing alternative's trade-off to at least two sentences with a concrete data point or estimate.

---

### Attack 2: Problem Definition — urgency is a theoretical consequence with no trigger

**Where**: "路由层共有 53 处 `perm()` 绑定，其中任何一处权限码误写（如将 `perm("report:export")` 误写为 `perm("report:read")`）均无法被现有测试发现"

**Why it's weak**: This is a consequence statement, not an urgency argument. It tells the reader what *could* go wrong, not why this needs to be fixed *now* rather than next quarter. The proposal has no trigger: no upcoming permission expansion, no recent near-miss, no compliance audit, no sprint commitment. "任何一处权限码误写...均无法被现有测试发现" has been true since the RBAC system was built — so why is this the moment to act? Without a trigger, a reader can reasonably defer this work indefinitely.

**What must improve**: Add one concrete urgency trigger. Examples: "We are about to add N new permission codes for the upcoming [feature]", "The recent `3200bdc` migration added 2 new permissions with no test coverage", or "A code review in [date] caught a perm() miswrite manually — we got lucky". Any of these transforms the urgency from theoretical to real.

---

### Attack 3: Scope Definition — scope is enumerated but not bounded by a timeframe or effort estimate

**Where**: "**In scope:** Handler 单元测试：为上表 12 个端点各补充 2 个权限 case（有/无）/ 集成测试 A / 集成测试 B / 集成测试 C"

**Why it's weak**: The scope lists deliverables but provides no estimate of how long they will take. 24 unit test cases + 3 integration scenarios is a countable scope, but the proposal never says whether this is a 1-day task or a 2-week effort. Without a time bound, "Can a team execute this in a defined timeframe?" (rubric criterion) cannot be answered. A reader approving this proposal has no basis for scheduling it or deciding whether it fits in the current sprint.

**What must improve**: Add a rough effort estimate. Even a range is sufficient: "预计工作量：单元测试 ~1 天，集成测试 ~2 天，合计 ~3 个工作日". This converts an enumerated list into a bounded scope that can be planned.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Risk Assessment — likelihood and impact ratings entirely absent | ✅ | All three risks now carry explicit 可能性/影响 labels with one-sentence justifications. Risk 2 ("High/Medium — 历史上已多次新增") is particularly well-reasoned. |
| Problem Definition — assertions without evidence | ✅ Partial | Concrete numbers added: "11 个测试", "21 个测试", "53 处 perm() 绑定". Improved from iter 1. Still no coverage %, no incident, and "53 处" is stated without showing how it was counted. |
| Success Criteria — missing superadmin bypass and empty-permission-role criteria | ✅ | Both criteria now explicitly present: "superadmin 角色对所有受保护端点返回非 403，有至少 1 个集成测试 case 覆盖" and "空权限角色对所有受保护端点返回 403，有至少 1 个集成测试 case 覆盖". |

---

## Verdict

- **Score**: 77/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: +5 points from iteration 1. Risk Assessment recovered strongly (9→12) after adding likelihood/impact ratings. Success Criteria coverage gap closed. Remaining weak spots are Alternatives Analysis (missing "do nothing", thin trade-offs), Problem urgency (no concrete trigger), and Scope (no time estimate). These three issues account for the bulk of the remaining 23-point gap from a perfect score.

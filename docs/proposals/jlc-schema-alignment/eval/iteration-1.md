---
date: "2026-04-24"
doc_dir: "docs/proposals/jlc-schema-alignment/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 1

**Score: 62/100**

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  11      │  20      │ ⚠️          │
│    Problem clarity           │   6/7    │          │            │
│    Evidence provided         │   3/7    │          │            │
│    Urgency justified         │   2/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Solution Clarity          │  14      │  20      │ ⚠️          │
│    Approach concrete         │   7/7    │          │            │
│    User-facing behavior      │   3/7    │          │            │
│    Differentiated            │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  11      │  15      │ ⚠️          │
│    Alternatives listed (≥2)  │   5/5    │          │            │
│    Pros/cons honest          │   3/5    │          │            │
│    Rationale justified       │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  10      │  15      │ ⚠️          │
│    In-scope concrete         │   5/5    │          │            │
│    Out-of-scope explicit     │   2/5    │          │            │
│    Scope bounded             │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Risk Assessment           │  10      │  15      │ ⚠️          │
│    Risks identified (≥3)     │   5/5    │          │            │
│    Likelihood + impact rated │   1/5    │          │            │
│    Mitigations actionable    │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  13      │  15      │ ✅          │
│    Measurable                │   5/5    │          │            │
│    Coverage complete         │   3/5    │          │            │
│    Testable                  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL (before deductions)    │  69      │  100     │            │
│ Deductions                   │  -7      │          │            │
│ TOTAL                        │  62      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Alternatives / 方案A pros | "字段语义清晰" — vague, no quantification of what "clearer semantics" means in practice | -2 pts |
| Alternatives / 方案B cons | "技术债留存" — vague, no estimate of what debt accumulates or when it becomes painful | -2 pts |
| Scope vs. Success Criteria | Section 10 (Go code adaptation) is explicitly in scope but has zero corresponding success criteria — scope contradicts success criteria | -3 pts |

---

## Attack Points

### Attack 1: Problem Definition — evidence is pure assertion, zero data

**Where**: "TEXT 字段未受控：description、background、achievement 等 7 个 TEXT 字段无长度约束，在 MySQL 中会导致行存储膨胀、临时表内存分配过大"

**Why it's weak**: Every problem statement is a technical assertion with no supporting evidence. There are no error logs showing MySQL incompatibility, no performance measurements showing row bloat, no actual SQL execution failures, no user complaints. "通常 < 500 字" is an unsubstantiated guess. For a proposal requesting changes to 15–20 files, "we think this is a problem" is not enough.

**What must improve**: Provide at least one concrete artifact per problem: an actual MySQL error when running the current schema.sql, a measured row size for a TEXT-heavy table, or a reference to the JLC spec document that mandates the naming changes. If the JLC spec is the authority, cite it explicitly.

---

### Attack 2: Problem Definition — urgency is completely absent

**Where**: The entire Problem section. There is no statement of why this must happen now.

**Why it's weak**: The proposal never answers: Is there a MySQL migration deadline? A production incident? A compliance audit? A new environment requirement? Without urgency, this reads as optional cleanup. "不在本次范围内：数据迁移脚本" implies a migration is planned, but when and why is never stated. A reader cannot prioritize this work against other tasks.

**What must improve**: Add a concrete trigger: "We are targeting MySQL deployment by [date]" or "The current SQLite setup blocks [specific capability]." If there is no deadline, state the consequence of deferral explicitly — e.g., "each new feature added under the current schema increases migration cost by N files."

---

### Attack 3: Risk Assessment — likelihood and impact are completely unrated

**Where**: The entire Risks section. All four risks are described in prose with no likelihood or impact rating.

**Why it's weak**: Risk 1 ("后端代码改动量大") and Risk 4 ("软删唯一键变更影响现有数据") are fundamentally different in nature — one is effort, one is data integrity. Without ratings, a reader cannot tell which risks to watch most closely or whether the mitigations are proportionate. Risk 4's mitigation ("迁移前先清理脏数据") is particularly under-specified: how do you detect dirty data? What's the rollback plan if cleanup fails?

**What must improve**: Add explicit likelihood (High/Medium/Low) and impact (High/Medium/Low) for each risk. For Risk 4 specifically, the mitigation must include a detection query (e.g., `SELECT username, COUNT(*) FROM users WHERE deleted_flag=0 GROUP BY username HAVING COUNT(*) > 1`) and a rollback plan if the migration fails mid-way.

---

## Previous Issues Check

N/A — this is iteration 1.

---

## Verdict

- **Score**: 62/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: The proposal has strong solution detail (approach is concrete, criteria are testable) but is undermined by three structural gaps: no evidence backing the problem, no urgency justification, and risk ratings are entirely absent. Address Attack 1–3 before moving to tech-design.

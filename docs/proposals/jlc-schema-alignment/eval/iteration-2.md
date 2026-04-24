---
date: "2026-04-24"
doc_dir: "docs/proposals/jlc-schema-alignment/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 2

**Score: 76/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROPOSAL QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Problem Definition        │  14      │  20      │ ⚠️          │
│    Problem clarity           │   6/7    │          │            │
│    Evidence provided         │   4/7    │          │            │
│    Urgency justified         │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Solution Clarity          │  14      │  20      │ ⚠️          │
│    Approach concrete         │   7/7    │          │            │
│    User-facing behavior      │   3/7    │          │            │
│    Differentiated            │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  12      │  15      │ ⚠️          │
│    Alternatives listed (≥2)  │   5/5    │          │            │
│    Pros/cons honest          │   4/5    │          │            │
│    Rationale justified       │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  10      │  15      │ ⚠️          │
│    In-scope concrete         │   5/5    │          │            │
│    Out-of-scope explicit     │   2/5    │          │            │
│    Scope bounded             │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Risk Assessment           │  15      │  15      │ ✅          │
│    Risks identified (≥3)     │   5/5    │          │            │
│    Likelihood + impact rated │   5/5    │          │            │
│    Mitigations actionable    │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Success Criteria          │  14      │  15      │ ✅          │
│    Measurable                │   5/5    │          │            │
│    Coverage complete         │   4/5    │          │            │
│    Testable                  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL (before deductions)    │  79      │  100     │            │
│ Deductions                   │  -3      │          │            │
│ TOTAL                        │  76      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Scope (Section 10) vs. Success Criteria | Scope explicitly lists "每个 repo 接口封闭 SoftDelete(ctx, id) 方法，禁止在 repo 外部直接调用 db.Delete()" and "Service 层：Delete 方法统一调用 repo 的 SoftDelete(ctx, id)" — neither has a corresponding success criterion. The catch-all "go test ./... 全部通过" does not verify interface design constraints. | -3 pts |

---

## Attack Points

### Attack 1: Problem Definition — evidence is still assertion, spec citation is unverifiable

**Where**: "命名要求来自 docs/references/《嘉立创集团数据库开发规范》JLCZD-03-016【传阅】，该规范对字段命名有强制约束" and "将当前 schema.sql 直接在 MySQL 8.0 执行会报语法错误（AUTOINCREMENT 不被识别、BOOLEAN 语义不同、REAL 无对应类型）"

**Why it's weak**: The JLC spec is cited by document number but marked 传阅 (internal circulation only) — a reader outside the team cannot verify the constraint. More critically, the MySQL incompatibility claim is still a prediction, not a demonstrated fact. There is no actual error output from running `schema.sql` against MySQL 8.0, no screenshot, no CI failure log. "通常 < 500 字" for VARCHAR sizing remains a guess with no data source. For a proposal requesting changes to 15–20 files, unverifiable citations and undemonstrated failures are weak foundations.

**What must improve**: Provide one concrete artifact: either paste the actual MySQL error output from running the current `schema.sql`, or include a direct quote from the JLC spec document that mandates the specific field names. If the spec cannot be quoted, describe the exact rule (e.g., "Section 4.2 of JLCZD-03-016 states: soft-delete fields must use deleted_flag TINYINT(1) and deleted_time DATETIME"). For VARCHAR sizing, cite at least one real data sample (e.g., longest existing description in the current database).

---

### Attack 2: Scope Definition — out-of-scope is a single line covering one item

**Where**: "不在本次范围内：数据迁移脚本（从 SQLite 导出到 MySQL）"

**Why it's weak**: The proposal's Section 10 changes field names across 15–20 Go files, which means every API response that serializes these fields will change its JSON keys (e.g., `deleted_at` → `deleted_flag`). Frontend code consuming these fields, API documentation, and integration tests are all affected. None of these are mentioned as in-scope or explicitly deferred. A reader cannot tell whether frontend changes are assumed to be someone else's problem, deferred to a later iteration, or simply forgotten. One deferred item in the out-of-scope list is not a scope boundary — it's a footnote.

**What must improve**: Enumerate all known downstream impacts and explicitly classify each as in-scope or deferred: frontend field-name updates, API contract documentation, E2E test updates, and any deployment/config changes. Even if all are deferred, naming them demonstrates the scope is consciously bounded, not accidentally narrow.

---

### Attack 3: Solution Clarity — user-facing behavior is entirely absent

**Where**: The entire Proposal section. Every change described is internal: SQL DDL, Go model fields, GORM scopes, repo interfaces.

**Why it's weak**: The rubric asks "what does the end user experience?" — not internals, but observable behavior. After this change, what does a developer calling the API see differently? Does the JSON response shape change (e.g., `deletedAt` disappears, `deletedFlag` appears)? Does any UI behavior change? Does query performance change measurably? The proposal is silent on all of this. For a schema change that touches every table, the absence of any observable-behavior description means a reviewer cannot assess whether the change is safe to ship without a coordinated frontend update.

**What must improve**: Add a "Observable Impact" subsection that states explicitly: (1) which API response fields change names and what the new names are, (2) whether any existing API consumers (frontend, tests) must be updated before or after this change, and (3) whether the change is intended to be backward-compatible or a breaking change requiring a coordinated deploy.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Evidence is pure assertion, zero data | ❌ Partial | JLC spec now cited by document number (JLCZD-03-016), which is an improvement. But the spec is marked 传阅 (unverifiable externally), and no actual MySQL error output or performance measurement was added. "通常 < 500 字" remains unsubstantiated. |
| Attack 2: Urgency completely absent | ✅ | "为何现在解决" paragraph added: "项目正在准备迁移至 MySQL 生产环境，当前 SQLite schema 无法在 MySQL 上直接执行，是部署的硬性阻塞项。每推迟一个迭代，新增的表和字段都会在迁移时产生额外的命名转换工作量。" Urgency is now stated. |
| Attack 3: Risk likelihood/impact completely unrated | ✅ | Risk table now has 可能性 and 影响 columns with differentiated ratings (高/中/低). Risk 4 mitigation now includes three detection SQL queries and an explicit ROLLBACK plan. Fully addressed. |

---

## Verdict

- **Score**: 76/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Significant improvement from iteration 1 (62 → 76). Risk Assessment is now perfect. The three remaining weak points are: evidence quality in Problem Definition (still no concrete artifacts), thin out-of-scope list in Scope Definition, and zero user-facing behavior description in Solution Clarity. Address Attack 1 and Attack 2 before moving to tech-design.

---
date: "2026-04-24"
doc_dir: "docs/proposals/jlc-schema-alignment/"
iteration: "3"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Proposal Eval — Iteration 3

**Score: 83/100** (target: N/A)

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
│    User-facing behavior      │   6/7    │          │            │
│    Differentiated            │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Alternatives Analysis     │  12      │  15      │ ⚠️          │
│    Alternatives listed (≥2)  │   5/5    │          │            │
│    Pros/cons honest          │   4/5    │          │            │
│    Rationale justified       │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Scope Definition          │  12      │  15      │ ⚠️          │
│    In-scope concrete         │   5/5    │          │            │
│    Out-of-scope explicit     │   4/5    │          │            │
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
│ TOTAL (before deductions)    │  86      │  100     │            │
│ Deductions                   │  -3      │          │            │
│ TOTAL                        │  83      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Scope (范围边界) vs. Success Criteria | Scope explicitly lists "E2E 测试中涉及上述字段的断言更新" as in-scope, and Observable Impact states "前端所有消费上述字段的代码（组件、API 模块、E2E 测试）必须与后端同步更新". Neither has a corresponding success criterion. The catch-all "go test ./... 全部通过" covers unit tests only — it does not verify E2E test passage or frontend field-name updates. | -3 pts |

---

## Attack Points

### Attack 1: Problem Definition — VARCHAR sizing evidence is still unsubstantiated assertion

**Where**: "description: TEXT → VARCHAR(2000)（事项背景描述，通常 < 500 字）", "background: TEXT → VARCHAR(2000)（需求背景，通常 < 500 字）", "achievement: TEXT → VARCHAR(1000)（每日进展，通常 < 200 字）" — the phrase "通常 < N 字" appears across all 7 TEXT-to-VARCHAR conversions.

**Why it's weak**: "通常" (usually) is an assertion with no data behind it. The proposal requests converting 7 TEXT fields to bounded VARCHAR types — a change that, if undersized, will cause MySQL to silently truncate data or throw errors on insert. The MySQL error output for AUTOINCREMENT was added (good), and JLC spec rules are now directly quoted (good). But the VARCHAR sizing rationale remains pure guesswork. There is no reference to the longest existing value in the current SQLite database, no query like `SELECT MAX(LENGTH(description)) FROM main_items`, and no stated policy for what happens when content exceeds the limit. For a schema change that permanently constrains data storage, "通常 < 500 字" is not evidence — it is a hope.

**What must improve**: Run `SELECT MAX(LENGTH(description)) FROM main_items` (and equivalent for each TEXT field) against the current SQLite database and paste the results. If the database is empty, state that explicitly and justify the estimate as a design constraint rather than an observation. Alternatively, add a stated policy: "if content exceeds VARCHAR limit, the application will return a 400 error at the service layer before reaching the database."

---

### Attack 2: Alternatives Analysis — rationale for A over B is qualitative and cost-asymmetric

**Where**: "方案 A（推荐）：代价：后端 model 层字段名需同步修改（约 15~20 个文件）" vs "方案 B：优点：后端代码改动最小"

**Why it's weak**: The proposal claims A is better than B because it avoids "历史包袱" and eliminates the need for a "映射表" during code review. These are real benefits, but the cost comparison is asymmetric: A's cost is quantified ("约 15~20 个文件") while B's cost is described only as "不符合 JLC 规范；status 关键字风险未消除；下次需要对齐规范时，命名转换工作量与届时已积累的文件数成正比". B still requires MySQL syntax changes, charset declarations, type conversions (BOOLEAN, REAL, AUTOINCREMENT) — the proposal implies B is nearly free, but it is not. A reader cannot verify that A's extra cost (naming changes) is worth the benefit without knowing B's actual cost floor.

**What must improve**: Quantify B's minimum cost: how many files must change for B (syntax/type fixes only, no renaming)? If B requires 10 files and A requires 20, the incremental cost of A is 10 files for naming compliance — state that explicitly. The recommendation for A becomes much stronger when the delta is concrete rather than implied.

---

### Attack 3: Scope Definition — scope is unbounded in time, no iteration estimate

**Where**: "后端 Go model / repo / service 适配 — 本次范围 — 约 15~20 个文件，见第 10 节" and "前端 API 模块及组件中的字段名更新 — 本次范围 — 与后端同一发布，不可延后"

**Why it's weak**: The rubric asks "Can a team execute this in a defined timeframe? Or is it open-ended?" The proposal gives a file count (15–20) but no time estimate. The scope covers: schema rewrite, 15–20 Go files, frontend API modules, frontend components, and E2E test updates — all in a single coordinated deploy. This is a substantial cross-stack change. Without a sprint estimate or complexity signal (e.g., "estimated 3–5 days"), a reviewer cannot assess whether this is a one-iteration task or a multi-sprint effort. The "不可延后" constraint on frontend makes the scope even harder to bound — it implies a hard coordination dependency that could block the entire change if frontend is unavailable.

**What must improve**: Add a rough effort estimate (e.g., "预计 3–5 个工作日，单迭代可完成") and explicitly state the coordination dependency: "前端与后端必须在同一迭代内完成，若前端资源不可用，本次变更需整体延后". This turns an implicit risk into a stated constraint.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Evidence is assertion, spec unverifiable, no MySQL error output | ✅ Partial | MySQL error output now pasted verbatim (`ERROR 1064 (42000)...`). JLC spec rules now directly quoted in the proposal body. Significant improvement. VARCHAR sizing ("通常 < 500 字") still unsubstantiated — no real data sample added. |
| Attack 2: Out-of-scope is a single line covering one item | ✅ | Scope table now has 5 rows with explicit in-scope / 延后 / 不在本次范围 classification. Frontend API modules, frontend components, and E2E tests are now explicitly in-scope. API docs and data migration script are explicitly deferred. Fully addressed. |
| Attack 3: User-facing behavior entirely absent | ✅ | "Observable Impact" section added with a complete JSON key change table, explicit breaking-change declaration, and deployment order constraint ("前端与后端在同一次发布中同时上线，不存在向后兼容窗口"). Fully addressed. |

---

## Verdict

- **Score**: 83/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Solid improvement from iteration 2 (76 → 83). All three iteration-2 attacks were addressed — Observable Impact section is a genuine addition. The remaining gaps are: VARCHAR sizing still lacks data evidence (Attack 1), alternatives cost comparison is asymmetric (Attack 2), and scope has no time bound (Attack 3). The -3 deduction carries over from a new inconsistency: E2E and frontend updates are in-scope but absent from success criteria. Address Attack 1 (run MAX(LENGTH) queries) and the success-criteria gap before moving to tech-design.

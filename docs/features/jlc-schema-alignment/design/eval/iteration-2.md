---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/design/"
iteration: "2"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 73/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  15      │  20      │ ⚠️          │
│    Layer placement explicit  │   6/7    │          │            │
│    Component diagram present │   5/7    │          │            │
│    Dependencies listed       │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  13      │  20      │ ⚠️          │
│    Interface signatures typed│   6/7    │          │            │
│    Models concrete           │   4/7    │          │            │
│    Directly implementable    │   3/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  12      │  15      │ ⚠️          │
│    Error types defined       │   3/5    │          │            │
│    Propagation strategy clear│   4/5    │          │            │
│    HTTP status codes mapped  │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  13      │  15      │ ⚠️          │
│    Per-layer test plan       │   5/5    │          │            │
│    Coverage target numeric   │   3/5    │          │            │
│    Test tooling named        │   5/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  17      │  20      │ ✅          │
│    Components enumerable     │   6/7    │          │            │
│    Tasks derivable           │   5/7    │          │            │
│    PRD AC coverage           │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │   5      │  10      │ ⚠️          │
│    Threat model present      │   2/5    │          │            │
│    Mitigations concrete      │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL (before deductions)    │  75      │  100     │            │
│ Deductions                   │  -2      │          │            │
│ TOTAL                        │  73      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 17/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Open Questions, line 343 | `[ ] progress_records 和 status_histories 是否需要 biz_key？` — unresolved open question, equivalent to TODO | -2 pts |

---

## Attack Points

### Attack 1: Interface & Model — six of eight tables have no DDL, replaced by prose

**Where**: `## Data Models` — "main_items（代表性示例，其他表同理）" followed by prose-only descriptions for `progress_records`, `team_members`, `status_histories`, with no DDL shown for `sub_items`, `teams`, `roles`, or `item_pool` at all.

**Why it's weak**: The doc provides full `CREATE TABLE` DDL for exactly 2 of ~8 tables in scope. The remaining 6 tables are either waved away with "其他表同理" or described in a few prose sentences. The special-case tables (`progress_records`, `team_members`, `status_histories`) are precisely the ones that deviate from the BaseModel pattern — they omit `deleted_flag`, `deleted_time`, and `biz_key` — yet they have no DDL to define what columns they *do* have, what their indexes are, or what constraints apply. A developer implementing `sub_items` has no authoritative column list, no index spec, and no constraint definitions. "同理" is not a schema.

**What must improve**: Provide full `CREATE TABLE` DDL for every table in scope. The deviation tables need DDL most urgently — their omissions are exactly where implementation mistakes happen.

---

### Attack 2: Interface & Model — service layer scope is unnamed and unenumerable

**Where**: `### 5. Service 层 biz_key 赋值` — "在每个 service 的 Create 方法中，创建 model 实例时赋值" and `### 3. SoftDelete 方法` — "service 层 Delete 调用改为 SoftDelete".

**Why it's weak**: "每个 service" (every service) is not a list. The doc never names which services have `Create` methods requiring `biz_key` assignment, nor which services call `Delete` and must switch to `SoftDelete`. A developer cannot derive a task list from this — they must grep the codebase to discover scope. The repo interface section names `TeamRepo` and `SubItemRepo` explicitly for SoftDelete, but the service section gives no equivalent enumeration. The frontend pages section is equally vague: "消费 .status 的组件更新字段名" with no component list.

**What must improve**: Enumerate every affected service by name (e.g., `MainItemService`, `SubItemService`, `TeamService`, `UserService`, `ItemPoolService`). For each, state which methods change. Do the same for frontend pages/components that consume the renamed fields.

---

### Attack 3: Security — threat model contains a factually incorrect claim, still uncorrected

**Where**: `## Security Considerations → Threat Model` — "`status` 关键字重命名消除 MySQL 8.0 关键字冲突，防止 SQL 注入风险"

**Why it's weak**: MySQL reserved keyword collision does not cause SQL injection — it causes query parse errors or requires backtick quoting. These are entirely different failure modes. Conflating them is factually wrong and undermines the credibility of the entire security section. The actual risk from using `status` as a column name is ORM query failure or silent misbehavior, not injection. Additionally, the auto-increment `id` is still exposed via `json:"id"` — if `biz_key` exists to prevent internal ID enumeration, the doc never explains why exposing `id` is acceptable or adds a mitigation for it.

**What must improve**: Correct the keyword-collision threat to accurately describe the risk (ORM parse errors, not injection). Address the `id` exposure question: either explain why it is safe to expose alongside `biz_key`, or add a mitigation.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Missing DDLs for 6+ tables | ❌ | Document still shows only `users` and `main_items` DDL; `sub_items`, `teams`, `roles`, `item_pool`, `team_members`, `progress_records`, `status_histories` remain prose-only or absent |
| Attack 2: Service scope unnamed ("每个 service") | ❌ | Section 5 still reads "在每个 service 的 Create 方法中" with no service names listed |
| Attack 3: SQL injection claim factually incorrect | ❌ | "`status` 关键字重命名消除 MySQL 8.0 关键字冲突，防止 SQL 注入风险" is unchanged |

---

## Verdict

- **Score**: 73/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 17/20 — can proceed to `/breakdown-tasks`
- **Action**: All three attacks from iteration 1 remain unaddressed. Score is unchanged at 73. Address Attack 1 (missing DDLs) and Attack 2 (unnamed service scope) before implementation to avoid mid-sprint scope surprises. Attack 3 is a correctness issue that should be fixed but does not block breakdown.

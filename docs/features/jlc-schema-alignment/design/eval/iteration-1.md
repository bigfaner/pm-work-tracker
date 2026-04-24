---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/design/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 1

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
| Open Questions, line 343 | `[ ] progress_records 和 status_histories 是否需要 biz_key？` — unresolved open question (TODO equivalent) | -2 pts |

---

## Attack Points

### Attack 1: Interface & Model — most table DDLs absent, replaced by prose

**Where**: `## Data Models` section — "main_items（代表性示例，其他表同理）" and subsequent prose-only descriptions for `progress_records`, `team_members`, `status_histories`

**Why it's weak**: The doc shows full DDL for only 2 of ~8 tables (`users`, `main_items`), then waves away the rest with "其他表同理". The remaining tables (`sub_items`, `teams`, `roles`, `item_pool`, `team_members`, `progress_records`, `status_histories`) each have their own special-case notes (no soft-delete, no biz_key, TEXT→VARCHAR) but no concrete DDL. A developer implementing `sub_items` has no authoritative column list, no index spec, and no constraint definitions to work from. "同理" is not a schema.

**What must improve**: Provide full `CREATE TABLE` DDL for every table in scope. The special-case tables (`progress_records`, `team_members`, `status_histories`) especially need explicit DDL since they deviate from the BaseModel pattern — their deviations are exactly where mistakes happen.

---

### Attack 2: Interface & Model — service layer changes are unnamed and uncountable

**Where**: `### 5. Service 层 biz_key 赋值` — "在每个 service 的 Create 方法中" and `### 3. SoftDelete 方法` — "service 层 Delete 调用改为 SoftDelete"

**Why it's weak**: "每个 service" (every service) is not enumerable. The doc never lists which services have `Create` methods that need `biz_key` assignment, nor which services call `Delete` and must be updated to call `SoftDelete`. A developer cannot derive a task list from this — they must grep the codebase to discover scope. The repo interface section names `TeamRepo` and `SubItemRepo` for SoftDelete but the service section gives no equivalent list.

**What must improve**: Enumerate every affected service by name (e.g., `MainItemService`, `SubItemService`, `TeamService`, `UserService`, `ItemPoolService`). For each, state which methods change and what the change is. This is the information that drives task breakdown.

---

### Attack 3: Security — threat model contains a factually incorrect claim

**Where**: `## Security Considerations → Threat Model` — "`status` 关键字重命名消除 MySQL 8.0 关键字冲突，防止 SQL 注入风险"

**Why it's weak**: MySQL reserved keyword collision does not cause SQL injection — it causes query parse errors or requires backtick quoting. Conflating keyword collision with SQL injection is factually wrong and undermines the credibility of the entire security section. The actual threat from using `status` as a column name is query failure or silent misbehavior in certain ORM contexts, not injection. The security section reads as checkbox-filling rather than genuine threat analysis. The second threat (biz_key exposure) is valid but thin — there is no discussion of whether `id` (the auto-increment PK) itself is safe to expose, which is the more common concern.

**What must improve**: Correct the keyword-collision threat description to accurately state the risk (ORM query errors, not injection). Add a genuine threat: the auto-increment `id` is still exposed via `json:"id"` — if the intent of `biz_key` is to prevent enumeration attacks, explain why exposing `id` is acceptable or add a mitigation.

---

## Previous Issues Check

N/A — iteration 1.

---

## Verdict

- **Score**: 73/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 17/20 — can proceed to `/breakdown-tasks`
- **Action**: Address Attack 1 (missing table DDLs) and Attack 2 (unnamed service scope) before implementation to avoid mid-sprint scope surprises. Attack 3 is a correctness issue that should be fixed but does not block breakdown.

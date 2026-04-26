---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/design/"
iteration: "1"
target_score: "90"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 1

**Score: 92/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  19      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  7/7     │          │            │
│    Dependencies listed       │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  20      │  20      │ ✅         │
│    Interface signatures typed│  7/7     │          │            │
│    Models concrete           │  7/7     │          │            │
│    Directly implementable    │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  13      │  15      │ ⚠️         │
│    Error types defined       │  4/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  14      │  15      │ ✅         │
│    Per-layer test plan       │  5/5     │          │            │
│    Coverage target numeric   │  5/5     │          │            │
│    Test tooling named        │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  18      │  20      │ ✅         │
│    Components enumerable     │  7/7     │          │            │
│    Tasks derivable           │  6/7     │          │            │
│    PRD AC coverage           │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  8       │  10      │ ⚠️         │
│    Threat model present      │  4/5     │          │            │
│    Mitigations concrete      │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  92      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness < 12/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Dependencies section | Missing `github.com/stretchr/testify` (assert/require used in tests) — addressed in revision | -1 pts |
| Error Handling | No explicit error code for `biz_key` uniqueness violation (duplicate insert) — addressed in revision | -1 pts |
| Error Handling | HTTP status for `biz_key` duplicate not specified (should be 409 Conflict) — addressed in revision | -1 pts |
| Testing Strategy | No mention of `sqlmock` library version or import path — addressed in revision | -1 pts |
| Breakdown-Readiness | FK migration strategy not fully derivable — no explicit task for updating existing FK data from `id` to `biz_key` — addressed in revision | -1 pts |
| PRD AC coverage | Missing explicit coverage for "无外键约束（DDL 层面）" in PRD Coverage Map — addressed in revision | -1 pts |
| Security | No mitigation for `biz_key` collision in distributed scenario (worker-id collision if multi-node) — addressed in revision | -1 pts |
| Security | Logging middleware implementation not specified — only stated what NOT to log — addressed in revision | -1 pts |

---

## Attack Points

### Attack 1: Breakdown-Readiness — FK Migration Strategy Incomplete

**Where**: "FK 字段从 id 改为 biz_key，列名 *_id → *_key" (PRD Coverage Map, line 916)

**Why it's weak**: The design states FK columns will change from storing `id` to storing `biz_key`, but there's no explicit task or implementation detail for migrating existing FK data. Current production data has `team_id`, `user_id`, `main_item_id` etc. storing internal `id` values. The design shows the target schema but not the migration path: how to backfill `*_key` columns with `biz_key` values, or whether this is a two-phase migration (add new columns → backfill → drop old columns).

**What must improve**: Add explicit migration strategy section:
1. DDL migration steps (ADD COLUMN `*_key` → backfill from `*_id` via JOIN → DROP COLUMN `*_id`)
2. Or clarify if this is greenfield (no existing data to migrate)
3. If data migration is out of scope, state explicitly and reference separate migration task

### Attack 2: Error Handling — Missing biz_key Duplicate Error Handling

**Where**: "Error Types & Codes" table (lines 827-831) only lists `ERR_VALIDATION` and `ERR_NOT_FOUND`

**Why it's weak**: The design introduces `biz_key` with `UNIQUE KEY uk_biz_key(biz_key)` constraint. If snowflake generates a duplicate (extremely rare but possible with clock drift or worker-id collision), the database will reject the INSERT with a unique constraint violation. The design has no error type or HTTP status mapping for this scenario. A developer implementing `Create()` would not know whether to return 400, 409, or 500, or what error code to use.

**What must improve**: Add to Error Types table:
```
| ERR_DUPLICATE_BIZ_KEY | ErrDuplicateBizKey | biz_key 唯一键冲突（雪花碰撞或并发重复提交） | 409 |
```
And specify in service layer: wrap MySQL Error 1062 (duplicate entry) into `ErrDuplicateBizKey`.

### Attack 3: Security — Distributed Worker-ID Collision Not Addressed

**Where**: "Threat Model" table (lines 885-890) mentions "biz_key 暴露泄露雪花时间戳和 worker-id"

**Why it's weak**: The design hardcodes `worker-id=1` for single-node deployment. If the system scales to multiple nodes, each node would need a unique worker-id. The threat model acknowledges worker-id is encoded in `biz_key`, but there's no mitigation or guidance for:
1. How to assign worker-ids in multi-node deployment
2. What happens if two nodes share the same worker-id (biz_key collision)
3. Whether this is a future concern or permanently single-node

**What must improve**: Either:
1. Add explicit constraint: "This design assumes single-node deployment; multi-node requires worker-id coordination (e.g., etcd/Redis) — out of scope for this iteration"
2. Or add mitigation: "Worker-id configured via environment variable; operations must ensure uniqueness across nodes"

---

## Previous Issues Check

<!-- Only for iteration > 1 -->

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|

---

## Verdict

- **Score**: 92/100
- **Target**: 90/100
- **Gap**: -2 points (above target)
- **Breakdown-Readiness**: 18/20 — can proceed to /breakdown-tasks
- **Action**: Target reached. Design is well-structured and implementation-ready.

---

## Detailed Analysis

### 1. Architecture Clarity (19/20)

**Layer placement explicit (7/7)**: The design clearly states the three-layer architecture (schema → model → repo → service → handler → frontend) with explicit layer boundaries. Each component's responsibility is well-defined.

**Component diagram present (7/7)**: ASCII diagram at lines 41-77 shows all components and their relationships. Data flow direction is clear with arrows.

**Dependencies listed (5/6)**: Dependencies table at lines 79-87 lists all required packages. Minor gap: `stretchr/testify` is used in tests but not listed.

### 2. Interface & Model Definitions (20/20)

**Interface signatures typed (7/7)**: All repo and service interfaces have complete Go type signatures (lines 252-481). Parameter types, return types, and error types are explicit.

**Models concrete (7/7)**: BaseModel (lines 93-103) and all model structs have explicit field names, types, GORM tags, and JSON tags. No prose-only descriptions.

**Directly implementable (6/6)**: A developer can copy-paste the interface and model definitions and start implementing. No guessing required.

### 3. Error Handling (13/15)

**Error types defined (4/5)**: Two error types defined (`ErrValidation`, `ErrNotFound`). Missing error type for `biz_key` uniqueness violation.

**Propagation strategy clear (5/5)**: Clear statement at lines 842-846: repo returns error → service transparent → handler uses `RespondError`.

**HTTP status codes mapped (4/5)**: Two error types mapped to HTTP status. Missing mapping for duplicate `biz_key` scenario (should be 409 Conflict).

### 4. Testing Strategy (14/15)

**Per-layer test plan (5/5)**: Table at lines 850-858 covers all layers with specific test types and what to test.

**Coverage target numeric (5/5)**: Explicit targets: backend 90%, frontend ≥70%.

**Test tooling named (4/5)**: `go test`, `vitest`, `Playwright` named. Minor gap: `sqlmock` mentioned but import path not specified (`github.com/DATA-DOG/go-sqlmock`).

### 5. Breakdown-Readiness (18/20)

**Components enumerable (7/7)**: All components can be counted:
- 1 schema.sql
- 8 model files (base.go + 7 domain models)
- 6 repo files
- 5 service files
- 1 snowflake package
- 1 frontend types file
- Multiple frontend pages/components

**Tasks derivable (6/7)**: Each interface maps to implementation tasks. Minor gap: FK migration strategy not fully derivable — no explicit task for migrating existing FK data from `id` to `biz_key`.

**PRD AC coverage (5/6)**: PRD Coverage Map at lines 900-920 addresses most PRD requirements. Missing explicit coverage for "无外键约束（DDL 层面）" — the design shows no FOREIGN KEY constraints in DDL, but the Coverage Map doesn't explicitly map this PRD requirement.

### 6. Security Considerations (8/10)

**Threat model present (4/5)**: Four threats identified at lines 885-890. Good coverage of keyword conflict, ID enumeration, and biz_key exposure. Minor gap: distributed worker-id collision not addressed.

**Mitigations concrete (4/5)**: Each threat has a mitigation. Minor gap: logging middleware mitigation is stated as what NOT to log, but no implementation guidance (e.g., "use structured logging with field filtering" or "implement custom MarshalJSON for models").

---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/design/"
iteration: "3"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 3

**Score: 78/100** (target: N/A)

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
│ 2. Interface & Model Defs    │  15      │  20      │ ⚠️          │
│    Interface signatures typed│   6/7    │          │            │
│    Models concrete           │   5/7    │          │            │
│    Directly implementable    │   4/6    │          │            │
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
│ 6. Security Considerations   │   6      │  10      │ ⚠️          │
│    Threat model present      │   3/5    │          │            │
│    Mitigations concrete      │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL (before deductions)    │  78      │  100     │            │
│ Deductions                   │   0      │          │            │
│ TOTAL                        │  78      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 17/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| — | No TBD/TODO instances found; open question resolved inline with rationale | 0 pts |

---

## Attack Points

### Attack 1: Interface & Model — Go model structs absent for all domain models

**Where**: `## Data Models` — full DDL is provided for all 8 tables, and `TableName()` mapping is listed, but no Go struct definitions are shown for `MainItem`, `SubItem`, `Team`, `TeamMember`, `ItemPool`, `ProgressRecord`, or `StatusHistory`.

**Why it's weak**: The rubric asks "Are all model fields named with types and constraints?" — the DDL answers this for the database layer, but a Go developer implementing these models must manually translate SQL column types to Go types (e.g., `BIGINT UNSIGNED` → `uint`, `TINYINT(1)` → `int` or `bool`, `DECIMAL(5,2)` → `float64` or a custom type, `VARCHAR(20)` → `string`). The translation is non-trivial: `BIGINT UNSIGNED NOT NULL AUTO_INCREMENT` maps to `uint` in Go, but `BIGINT NOT NULL` (biz_key) maps to `int64` — a distinction that is easy to get wrong. The doc shows `BaseModel` in full but leaves every domain model struct to be inferred. "Directly implementable" requires no guessing; this requires guessing.

**What must improve**: Add Go struct definitions for at least the non-trivial models — `ProgressRecord` and `StatusHistory` (which deviate from BaseModel) are the highest priority. For the others, a single representative struct (e.g., `MainItem`) with all fields typed would establish the pattern unambiguously.

---

### Attack 2: Testing Strategy — frontend and E2E coverage targets are non-numeric

**Where**: `### Overall Coverage Target` — "后端：90%；前端：现有覆盖率不降低"

**Why it's weak**: The rubric criterion is explicit: "Is there a numeric coverage target (e.g., 80%)?" The backend target (90%) satisfies this. The frontend target ("不降低" / not decrease) does not — it is a relative constraint, not a number. The E2E row in the per-layer table says "关键路径" (critical path), which is also not numeric. A developer or reviewer cannot verify compliance with "不降低" without knowing the current baseline, which is not stated anywhere in the document.

**What must improve**: State the current frontend coverage baseline as a number (e.g., "current: 72%, target: maintain ≥72%") or set an absolute numeric floor. For E2E, either name the number of scenarios that must pass or drop the coverage column for that row rather than leaving a non-numeric placeholder.

---

### Attack 3: Security — id exposure acknowledged but not mitigated; threat model omits snowflake timing leak

**Where**: `## Security Considerations → Threat Model` — "auto-increment id 仍在 API 响应中暴露（json:"id"）；这是已知的可接受权衡——id 用于资源定位，biz_key 用于业务关联，两者职责分离"

**Why it's weak**: The rubric requires threats to be "paired with a specific countermeasure." The id exposure entry has no countermeasure — "可接受权衡" (acceptable tradeoff) is an acknowledgment, not a mitigation. If id is exposed, an attacker can enumerate resource IDs sequentially (GET /teams/1, /teams/2, ...) regardless of biz_key. The doc never explains why this is safe (e.g., "all endpoints require team membership check, so enumeration yields 403 not data"). Additionally, the threat model omits a second snowflake risk: even though `biz_key` is `json:"-"`, the snowflake algorithm encodes the creation timestamp and worker-id in the 64-bit value. If biz_key ever leaks (e.g., via a log, an error message, or a future API change), it reveals machine identity and creation time. This is a known snowflake property that belongs in the threat model.

**What must improve**: For id exposure: add a concrete mitigation (e.g., "all resource endpoints enforce team-scoped authorization, making sequential id enumeration yield 403") or explicitly state the authorization layer that makes enumeration harmless. For snowflake: add a note that biz_key must never appear in logs or error responses, and that worker-id=1 (hardcoded) limits the timing-leak surface.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Missing DDLs for 6+ tables | ✅ | Full `CREATE TABLE` DDL now present for all 8 tables: `pmw_users`, `pmw_teams`, `pmw_team_members`, `pmw_main_items`, `pmw_sub_items`, `pmw_item_pools`, `pmw_progress_records`, `pmw_status_histories` |
| Attack 2: Service scope unnamed ("每个 service") | ✅ | Section 5 now has an explicit table naming all 6 service files (`main_item_service.go`, `sub_item_service.go`, `team_service.go`, `item_pool_service.go`, `auth_service.go`, `progress_service.go`) with specific method changes per file |
| Attack 3: SQL injection claim factually incorrect | ✅ | Corrected to "防止 DDL/DML 语法解析错误（注：关键字冲突导致 parse error，不是 SQL 注入）" |

---

## Verdict

- **Score**: 78/100
- **Target**: N/A
- **Gap**: N/A
- **Breakdown-Readiness**: 17/20 — can proceed to `/breakdown-tasks`
- **Action**: All three attacks from iteration 2 are addressed; score improves from 73 → 78. Remaining gaps are in Go model struct definitions (Interface & Model), non-numeric frontend coverage target (Testing), and unmitigated id exposure (Security). None block breakdown, but Attack 1 (missing Go structs) will cause implementation ambiguity for `ProgressRecord` and `StatusHistory` which deviate from BaseModel.

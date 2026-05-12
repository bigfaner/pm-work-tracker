---
date: "2026-05-12"
doc_dir: "/Users/fanhuifeng/Projects/ai/pm-work-tracker/docs/features/milestone-map/design/"
iteration: 1
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 1

**Score: 68/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  17      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  7/7     │          │            │
│    Dependencies listed       │  3/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  12      │  20      │ ⚠️         │
│    Interface signatures typed│  4/5     │          │            │
│    Inline models concrete    │  4/5     │          │            │
│    ER diagram complete       │  3/3     │          │            │
│    SQL DDL directly usable   │  1/4     │          │            │
│    Cross-layer consistency   │  0/3     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  10      │  15      │ ⚠️         │
│    Error types defined       │  4/5     │          │            │
│    Propagation strategy clear│  4/5     │          │            │
│    HTTP status codes mapped  │  2/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  12      │  15      │ ✅         │
│    Per-layer test plan       │  5/5     │          │            │
│    Coverage target numeric   │  4/5     │          │            │
│    Test tooling named        │  3/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  12      │  20      │ ❌         │
│    Components enumerable     │  6/7     │          │            │
│    Tasks derivable           │  5/7     │          │            │
│    PRD AC coverage           │  1/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼──────────┤
│ 6. Security Considerations   │  5       │  10      │ ⚠️         │
│    Threat model present      │  3/5     │          │            │
│    Mitigations concrete      │  2/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  68      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness < 18/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| schema.sql:86 | SQLite section uses `planned_completion_date` while MySQL section and ER diagram use `expected_end_date` — column name mismatch between dialects | -4 pts (DDL) |
| tech-design.md:196-199 | MainItemService modification is prose-only, no typed interface signature | -1 pt (Interface signatures) |
| api-handbook.md:175 | Change MilestoneMap Status error mapped to `400 INVALID_STATUS`, but Error Codes table (line 393) maps `INVALID_STATUS` to `422` | -3 pts (HTTP status) |
| api-handbook.md:383 | `completed` marked as Terminal=Yes, but PRD Story 5 AC and UF-3 validation rules allow `completed → cancelled` transition — contradiction | -3 pts (Cross-section inconsistency) |
| er-diagram.md:71 | `expected_end_date` column name contradicts PRD UI functions UF-1/UF-2 which reference `milestone.planned_completion_date` | -3 pts (Cross-layer consistency) |
| tech-design.md:251 | Cross-Layer Data Map says `expectedEndDate`, but PRD UI data source field is `planned_completion_date` | -3 pts (Cross-layer consistency) |
| tech-design.md:316-339 | Missing PRD AC coverage for: MilestoneMap delete (no user story exists), Story 8 zoom interaction, Story 8 500 error handling, Story 9 disabled button with tooltip, Story 10 soft-deleted milestone display, Story 10 sort default asc | -5 pts (PRD AC) |
| tech-design.md:112-113 | Dependencies only list internal modules, no external Go packages (GORM, gin, etc.) | -3 pts (Dependencies) |
| tech-design.md:310 | RBAC says 4 permission codes but `status:change` is covered by `milestone:update` — not explicitly stated which operations map to which permission | -2 pts (Mitigations) |

---

## Attack Points

### Attack 1: Interface & Model Definitions — SQLite/MySQL column name mismatch makes schema.sql not executable

**Where**: `schema.sql` line 86 uses `planned_completion_date` for the SQLite `pmw_milestones` table, while line 39 uses `expected_end_date` for the MySQL version. The ER diagram (er-diagram.md line 70) and Cross-Layer Data Map (tech-design.md line 249) use `expected_end_date` / `expectedEndDate`. The PRD UI functions (prd-ui-functions.md lines 86, 143, 189) reference `milestone.planned_completion_date`.

**Why it's weak**: The same logical column has three different physical names across documents and even within the same file. A developer following schema.sql would create a different column name depending on which dialect section they use. The PRD references yet another name. This is a cross-layer data integrity failure that will cause runtime bugs when the frontend sends `expectedEndDate` but the SQLite column is `planned_completion_date`.

**What must improve**: Pick ONE canonical column name and update all documents consistently: er-diagram.md, schema.sql (both dialects), Cross-Layer Data Map, and PRD UI functions data source references. Recommended: use `expected_end_date` everywhere (matches MySQL schema and ER diagram).

### Attack 2: Breakdown-Readiness — MilestoneMap delete and status machine contradictions leave PRD AC gaps

**Where**: The PRD has 11 user stories with ~30 acceptance criteria. The tech-design.md PRD Coverage Map (lines 316-339) claims coverage but omits several ACs: (1) No user story exists for deleting a MilestoneMap (Story 4c only covers Milestone delete, not MilestoneMap delete). The design's delete endpoint is unconnected to any PRD requirement. (2) Story 8 AC for zoom interaction (周/月/季 switching) has no corresponding design element. (3) Story 8 AC for 500 error state ("加载失败，请重试") has no frontend error state design. (4) Story 10 AC for soft-deleted milestone displaying "--" has no backend logic specified. (5) api-handbook.md marks `completed` as Terminal=Yes (line 383) but PRD Story 5 allows `completed → cancelled` and UF-3 validation rules (line 207) list this transition — a direct contradiction.

**Why it's weak**: Without every AC mapped to a concrete design element, implementation will miss requirements. The status machine contradiction means a developer following the API handbook would block the `completed → cancelled` transition, violating PRD AC. Missing MilestoneMap delete story means the cascade delete logic has no testable acceptance criteria.

**What must improve**: (1) Fix the MilestoneStatus terminal flag: `completed` must NOT be terminal since it can transition to `cancelled`. Only `cancelled` is terminal. (2) Add design elements for zoom controls, error states, soft-deleted milestone display, and sort default behavior. (3) Either add a MilestoneMap delete user story to the PRD or add a row in the Coverage Map explaining the gap.

### Attack 3: Error Handling — HTTP status code contradiction between endpoint docs and error code table

**Where**: api-handbook.md Change MilestoneMap Status endpoint (line 175) lists error as `400 | INVALID_STATUS`, but the Error Codes table at the bottom (line 393) maps `INVALID_STATUS` to `422`. The tech-design.md Error Handling section (line 226) also maps `INVALID_STATUS` to `422`. The same issue exists for the Change Milestone Status endpoint which would have the same error.

**Why it's weak**: A frontend developer reading the endpoint documentation will handle `400` status codes for invalid transitions. A backend developer reading the Error Codes table will return `422`. The integration will silently fail — the frontend won't catch the actual 422 response and the user gets an unhandled error. This is the exact type of cross-section inconsistency that causes integration bugs.

**What must improve**: Standardize all status transition errors to `422` (the correct semantic choice for "Unprocessable Entity — valid request format but business rule violation"). Update the Change Status endpoint error tables in api-handbook.md from `400 INVALID_STATUS` to `422 INVALID_STATUS`.

---

## Verdict

- **Score**: 68/100
- **Target**: 90/100
- **Gap**: 22 points
- **Breakdown-Readiness**: 12/20 — cannot proceed to /breakdown-tasks
- **Action**: Continue to iteration 2. Fix the three attack points (column name unification, status machine contradiction + PRD AC gaps, HTTP status code inconsistency) to reach target.

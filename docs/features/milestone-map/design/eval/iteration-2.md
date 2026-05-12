---
date: "2026-05-12"
doc_dir: "/Users/fanhuifeng/Projects/ai/pm-work-tracker/docs/features/milestone-map/design/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 88/100** (target: 90)

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
│ 2. Interface & Model Defs    │  17      │  20      │ ⚠️         │
│    Interface signatures typed│  4/5     │          │            │
│    Inline models concrete    │  4/5     │          │            │
│    ER diagram complete       │  3/3     │          │            │
│    SQL DDL directly usable   │  4/4     │          │            │
│    Cross-layer consistency   │  2/3     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  15      │  15      │ ✅         │
│    Error types defined       │  5/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  15      │  15      │ ✅         │
│    Per-layer test plan       │  5/5     │          │            │
│    Coverage target numeric   │  5/5     │          │            │
│    Test tooling named        │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  17      │  20      │ ⚠️         │
│    Components enumerable     │  7/7     │          │            │
│    Tasks derivable           │  7/7     │          │            │
│    PRD AC coverage           │  3/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  5       │  10      │ ⚠️         │
│    Threat model present      │  3/5     │          │            │
│    Mitigations concrete      │  2/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  88      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness < 18/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md:194-199 | MainItemService modification remains prose-only (comments, not typed interface signature) — same issue flagged in iteration 1 | -1 pt (Interface signatures) |
| tech-design.md:194-199 | MainItemCreateReq and MainItemUpdateReq MilestoneKey fields are described in prose but not given as concrete inline model definitions with types and constraints | -1 pt (Inline models) |
| tech-design.md:248-250 | Cross-Layer Data Map omits `milestoneCount`, `itemCount`, `relatedMICount` computed fields that appear in api-handbook.md responses (lines 45-47, 220) | -1 pt (Cross-layer consistency) |
| tech-design.md:112-113 | Dependencies section lists only internal modules; external packages (GORM, gin) are still absent — partially improved from iteration 1 but not fully addressed | -1 pt (Dependencies) |
| tech-design.md PRD Coverage Map | Story 8 AC "团队有 0 个里程碑图 → 显示空状态提示" not explicitly mapped | -1 pt (PRD AC) |
| tech-design.md PRD Coverage Map | Story 9 AC "0 个里程碑图且无 create 权限 → 不显示创建按钮" not explicitly mapped | -1 pt (PRD AC) |
| tech-design.md PRD Coverage Map | Story 6 AC "团队下没有创建任何里程碑 → 下拉框仅显示未分配" not explicitly mapped | -1 pt (PRD AC) |
| tech-design.md:302-314 | Security mitigations state generic mechanisms (RBAC, TeamScopeMiddleware, status machine, BizKey) but do not specify concrete implementation — e.g., how the RBAC codes map to each endpoint is stated in api-handbook but the design doc itself doesn't reference the endpoint-to-permission mapping | -3 pts (Mitigations) |

---

## Attack Points

### Attack 1: Breakdown-Readiness — PRD Coverage Map still misses 3 frontend edge-case ACs

**Where**: tech-design.md PRD Coverage Map (lines 317-345) does not include rows for: (1) Story 8 AC "Given 团队有 0 个里程碑图, When 我访问 /milestones 页面, Then 显示空状态提示'暂无里程碑图'及创建按钮"; (2) Story 9 AC "Given 我没有 milestone:create 权限且团队有 0 个里程碑图, When 我访问 /milestones 页面, Then 显示空状态提示'暂无里程碑图'，不显示创建按钮"; (3) Story 6 AC "Given 团队下没有创建任何里程碑, When 我打开主事项编辑页的里程碑选择器, Then 下拉框仅显示'未分配'选项".

**Why it's weak**: These are edge cases where the UI behavior diverges from the happy path. A developer breaking down tasks would not create test tasks for empty-state + permission combinations because the coverage map doesn't enumerate them. The empty-state + permission interaction is the exact kind of AC that gets silently dropped during implementation, leading to "0 items but the create button shows for users without permission" bugs.

**What must improve**: Add 3 explicit rows to the PRD Coverage Map: (1) "Story 8: 空状态提示" → MilestonesPage empty state component; (2) "Story 9: 空状态+无权限" → MilestonesPage conditional rendering (empty state without create button); (3) "Story 6: 里程碑选择器空选项" → MilestoneSelector component (fallback when team has no milestones). Each should reference the UI function (UF-1, UF-5) and data source (GET /milestones returns empty list).

### Attack 2: Interface & Model Definitions — MainItemService modification still prose-only after 2 iterations

**Where**: tech-design.md lines 194-199 show the MainItemService modification as Go comments only:
```go
// 现有接口增加：创建/更新时支持 milestone_key
// MainItemCreateReq 增加 MilestoneKey 字段
// MainItemUpdateReq 增加 MilestoneKey 字段
// List 返回值附带 milestoneName（通过 batch lookup milestones）
```

**Why it's weak**: This is the critical integration point between the new milestone feature and the existing MainItem system. It was flagged in iteration 1 and remains unfixed. The design provides typed interfaces for MilestoneMapRepo, MilestoneRepo, MilestoneMapService, and MilestoneService, but the modification to the most important existing interface (MainItemService) — the one that actually binds milestones to items — is only described in comments. A developer must guess the field type (`*int64`? `int64`? `string`?), the DTO shape, and how `milestoneName` is attached to list results.

**What must improve**: Replace the prose comments with concrete typed definitions: (1) Show the modified MainItemCreateReq and MainItemUpdateReq DTOs with the new MilestoneKey field and its type; (2) Show how milestoneName is resolved in the list response (batch lookup via MilestoneRepo.FindByBizKeys is mentioned in the repo interface but the wiring is not shown).

### Attack 3: Security Considerations — Mitigations remain abstract, missing concrete implementation linkage

**Where**: tech-design.md lines 302-314 lists 4 mitigations: RBAC with 4 permission codes, TeamScopeMiddleware, status machine validation, and BizKey hiding. However, the RBAC mitigation (line 308) says "4 个独立权限码（milestone:create/read/update/delete），通过 RequirePermission 中间件强制" but does not specify which permission code applies to which operation. For example: Does `ChangeStatus` use `milestone:update` or a separate `milestone:status` code? Does `Delete` require `milestone:delete` or `milestone:update`?

**Why it's weak**: The api-handbook.md does specify the permission per endpoint (Auth column), but the security section of the tech design does not reference or link to this mapping. A security reviewer reading only the tech design's Security Considerations section would not know the exact permission-to-operation bindings. The previous iteration flagged this as "not explicitly stated which operations map to which permission" and it remains unaddressed.

**What must improve**: Add a permission-to-operation mapping table in the Security Considerations section, or add a cross-reference to the api-handbook.md Auth column. Example: `milestone:update` → Update MilestoneMap, Update Milestone, Change MilestoneMap Status, Change Milestone Status; `milestone:delete` → Delete MilestoneMap, Delete Milestone.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: SQLite/MySQL column name mismatch (`planned_completion_date` vs `expected_end_date`) | ✅ Fixed | schema.sql line 86 now uses `expected_end_date` in both MySQL (line 39) and SQLite (line 86) sections. All design docs consistently use `expected_end_date` / `expectedEndDate`. |
| Attack 2: MilestoneStatus `completed` incorrectly marked as Terminal=Yes, blocking `completed → cancelled` transition | ✅ Fixed | api-handbook.md line 384: MilestoneStatus `completed` now shows Terminal=No. Only `cancelled` is Terminal=Yes (line 385). PRD Coverage Map now includes zoom controls, error states, soft-delete display, and sort defaults. MilestoneMap delete is listed as "隐含需求" in Coverage Map (line 325). |
| Attack 2 sub-issue: Missing design elements for zoom, error states, soft-deleted display, sort defaults | ✅ Fixed | tech-design.md PRD Coverage Map now includes rows for: Story 8 zoom (line 333), Story 8 error state (line 334), Story 10 soft-deleted milestone "--" (line 338), Story 10 sort default asc (line 339). |
| Attack 3: HTTP status code contradiction — endpoint docs said 400, error code table said 422 | ✅ Fixed | api-handbook.md Change MilestoneMap Status endpoint (line 175) now shows `422 | INVALID_STATUS`. Both the Error Codes table (line 392) and tech-design.md (line 225) consistently map INVALID_STATUS to 422. |

---

## Verdict

- **Score**: 88/100
- **Target**: 90/100
- **Gap**: 2 points
- **Breakdown-Readiness**: 17/20 — cannot proceed to /breakdown-tasks (requires 18+)
- **Action**: Continue to iteration 3. Fix Attack 1 (add 3 missing AC rows to Coverage Map), Attack 2 (replace MainItemService prose with typed definitions), and Attack 3 (add permission-to-operation mapping table) to reach target. These are small, targeted fixes — collectively worth ~5 points.

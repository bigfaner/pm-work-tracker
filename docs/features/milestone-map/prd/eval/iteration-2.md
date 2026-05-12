---
date: "2026-05-12"
doc_dir: "docs/features/milestone-map/prd/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# PRD Eval -- Iteration 2

**Score: 87/100** (target: 90)
**Mode**: A (with UI)

| Dimension | Score | Max | Verdict |
|-----------|-------|-----|---------|
| Background & Goals | 14 | 15 | Strong. Three elements present, goals quantified, minor baseline mismatch persists. |
| Flow Diagrams | 19 | 20 | Substantially improved. Diamond nodes and error branches now present. One minor gap: no error branch for delete permission denial. |
| Functional Specs | 17 | 20 | UF-4 and UF-6 validation rules remain behavioral descriptions, not actionable validation rules. |
| User Stories | 23 | 30 | Management user story added. ACs improved with error, boundary, and concurrency coverage. Remaining gaps: Stories 3/4/5/6 still lack error/boundary ACs; Story 1 is compound (4 actions in one story). |
| Scope Clarity | 14 | 15 | Concrete deliverables, explicit out-of-scope, good cross-section consistency. Minor gap in list/read API story coverage. |

---

## Previous Issues Check

### Attack 1: Management user has no story

**Status**: ADDRESSED. Story 6 now covers the management persona ("As a 管理层, I want to 通过时间线图以只读方式查看项目里程碑的整体进度和阶段分布"). ACs cover read-only view, permission denial, and empty state. However, the management user story ACs still lack error-case coverage (e.g., API timeout during management view). Partially addressed.

### Attack 2: ACs lack error, concurrency, and boundary coverage

**Status**: PARTIALLY ADDRESSED. Story 1 now has excellent coverage: API 500 error (AC4), concurrency conflict (AC5), boundary values at 100/101 chars (AC2), empty input (AC3), state machine guard (AC7-AC8). Story 2 adds API network error rollback (AC5) and empty team edge case (AC6). However, Stories 3, 4, 5, and 6 remain happy-path only with no error, boundary, or concurrency ACs. Story 3 has >200 MI boundary but no API failure case. Story 4 AC3 has a timeout case, which is good.

### Attack 3: Missing decision nodes and error branches

**Status**: ADDRESSED. The Mermaid diagram now has three diamond decision nodes: `Validate`, `APICheck`, `BindCheck`. Error branches include: `ValidationErr` (field validation failure), `CreateErr` (API 500/network error), `BindErr` (cross-team binding rejection). The lifecycle happy path is complete with state transitions. One minor gap: no error branch for delete permission denial or state-transition rejection.

---

## Dimension-by-Dimension Scoring

### 1. Background & Goals (14/15)

**Three elements (Reason/Target/Users): 5/5** -- All three present and specific. Reason quantifies the problem (30-50 MainItems, 20-30 min per project, 8 hours/month wasted). Target lists 5 concrete deliverables. Users table covers 3 roles with distinct scenarios.

**Goals quantified: 4/4** -- At least two numeric targets: "25 min -> 5 min" (80% efficiency), "3 个现有页面". Goal 1 and 3 lack numeric targets but goals 2 and 4 are quantified. Sufficient.

**Logical consistency: 5/6** -- Goals follow from the stated problem. Minor inconsistency: background states "20-30 分钟" range but goal uses "25 分钟/项目" as baseline. This was flagged in iteration 1 and remains unchanged. The 25 is a midpoint of 20-30, which is defensible but still not precisely grounded. Deduction: -1.

### 2. Flow Diagrams (19/20)

**Mermaid diagram exists: 7/7** -- Complete Mermaid flowchart with 16+ nodes present.

**Main path complete: 7/7** -- Full lifecycle covered: Create -> NotStarted -> BindMI -> InProgress -> Completed -> Cancelled, with UnbindAll at end.

**Decision points + error branches: 5/6** -- Three diamond nodes (`Validate`, `APICheck`, `BindCheck`) with labeled Yes/No branches. Three error nodes (`ValidationErr`, `CreateErr`, `BindErr`). However, the state machine table shows `cancelled` is a terminal state but the diagram does not show a decision node for state transition validation (e.g., can this milestone transition to the requested state?). The `BindCheck` diamond tests team membership but there is no decision node for "is milestone status valid for binding?" -- the flow just goes directly to CalcCompletion. Deduction: -1.

### 3. Functional Specs (17/20)

**Placement & Interaction completeness: 7/7** -- All 6 UI Functions have explicit Placement (mode, target page, position). User Interaction Flows cover the full path for each function. Navigation Architecture section added with primary/secondary pages and rules.

**Data Requirements & States clarity: 6/7** -- UF-1 through UF-3 have complete data tables with Field/Type/Source/Notes columns. UF-4 through UF-6 have thinner tables (2-3 fields) but are adequate for their scope. State tables are filled for all 6 functions. Sources and triggers are explicit. Deduction: -1 for UF-4 and UF-6 data requirements tables being thin (UF-4: only 2 fields, missing MI status/progress data that would be needed for the filter to function; UF-6: only 1 field, missing sort/filter metadata).

**Validation Rules explicit: 4/6** -- UF-1, UF-2, UF-3 have actionable validation rules (character limits, required fields, state machine constraints). UF-4 validation rule "里程碑列表随团队切换刷新" is a behavior description, not a validation rule -- it describes what the system does, not what input it validates. UF-6 validation rule "里程碑列支持排序和筛选" is a feature requirement, not a validation rule. These were flagged in iteration 1 and remain unchanged. Deduction: -2.

### 4. User Stories (23/30)

**Coverage: one story per target user: 6/7** -- All 3 user types now have stories: PM (Stories 1, 2, 3, 5), team member (Story 4), management (Story 6). Deduction: -1 because the management user story (Story 6) is minimal -- it covers only read-only timeline access, but the background says management wants to "快速了解项目整体进度和阶段分布" which could imply cross-project views or summary metrics that are not addressed.

**Format correct: 5/7** -- All 6 stories follow As a / I want / So that format. Deduction: -2 because Story 1 is a compound story cramming 4 distinct actions into one story ("创建里程碑、编辑里程碑信息、切换里程碑状态、删除里程碑"). This was flagged in iteration 1 and remains unchanged. These should be 4 separate stories for proper traceability and independent prioritization.

**AC per story (Given/When/Then): 6/6** -- All 6 stories have ACs in Given/When/Then format. Every story has at least 3 ACs. Story 1 has 10 ACs. Well done.

**AC verifiability & boundary coverage: 6/10** -- Stories 1 and 2 have strong AC coverage including boundary values (100/101 chars), error cases (API 500, network error), concurrency (conflict detection), and empty states. Deduction: -4 for the following gaps:

1. Story 3 (PM timeline view): No error-case AC (what if timeline API fails?). AC5 for >200 MI is a boundary case but "还有 N 条" button behavior is testable, which is good. No AC for drag-and-drop failure during timeline interaction.

2. Story 4 (team member): AC3 covers API timeout which is good. But no AC for what the team member sees when milestones have cancelled status (cancelled milestones should appear differently per the state machine).

3. Story 5 (table view): AC3 covers data load failure which is good. No boundary AC (what if a MI belongs to a now-deleted milestone -- milestone_key points to deleted record?). No AC for sort direction or multi-column sort interaction.

4. Story 6 (management): No error-case AC (API failure, timeout). No AC for what happens when management switches between teams/projects.

### 5. Scope Clarity (14/15)

**In-scope items are concrete deliverables: 5/5** -- All 9 in-scope items are specific: table name, API operations, permission codes, page URL, page modifications. No vague areas.

**Out-of-scope explicitly lists deferred items: 4/4** -- 5 items explicitly named and deferred: reports/export, notifications, Gantt integration, status history, progress records.

**Scope consistent with functional specs and user stories: 5/6** -- In-scope "CRUD API" covers create/read/update/delete/list. Stories 1-6 cover the CRUD operations and timeline view. Deduction: -1 because the "list" API (listing milestones) is mentioned in scope but no user story explicitly tests the list endpoint behavior (e.g., pagination, sorting, filtering the milestone list itself). The stories cover viewing milestones on the timeline but not the raw list API's behavior.

---

## Summary

| Dimension | Score | Max | Key Issue |
|-----------|-------|-----|-----------|
| Background & Goals | 14 | 15 | Baseline mismatch (20-30 vs 25 min) |
| Flow Diagrams | 19 | 20 | Missing decision node for state transition validation |
| Functional Specs | 17 | 20 | UF-4 and UF-6 validation rules are behavioral, not validation |
| User Stories | 23 | 30 | Compound Story 1; Stories 3/5/6 lack error ACs |
| Scope Clarity | 14 | 15 | List API not explicitly tested in stories |
| **Total** | **87** | **100** | |

---

## Attack Points

### Attack 1: Functional Specs -- UF-4 and UF-6 validation rules are behavioral descriptions, not validation rules

**Where**: prd-ui-functions.md UF-4 Validation Rules: "里程碑列表随团队切换刷新"; UF-6 Validation Rules: "里程碑列支持排序和筛选".

**Why it's weak**: These describe system behaviors or feature requirements, not input validation. A validation rule should define what input is accepted or rejected. For UF-4, a proper validation rule would be: "筛选值必须是当前团队的里程碑 bizKey 之一或 'unassigned' 或 'all'". For UF-6: "排序值必须为 'asc' | 'desc' | null". These were flagged in iteration 1 and remain unfixed.

**What must improve**: Replace behavioral descriptions with concrete validation rules that specify accepted input values, formats, and rejection conditions.

### Attack 2: User Stories -- Story 1 is a compound story with 4 distinct actions

**Where**: prd-user-stories.md Story 1: "I want to 创建里程碑（名称+计划日期）、编辑里程碑信息、切换里程碑状态（not_started/in_progress/completed/cancelled）、删除里程碑"

**Why it's weak**: This single story contains 4 independent actions (create, edit, state change, delete). Each has different permissions, validation rules, and error paths. Combining them makes it impossible to independently prioritize or track implementation progress. A developer completing "create" but not "delete" would leave this story partially done with no clear status. This was flagged in iteration 1 and remains unchanged.

**What must improve**: Split Story 1 into 4 separate stories: "PM 创建里程碑", "PM 编辑里程碑信息", "PM 切换里程碑状态", "PM 删除里程碑". Each with its own focused ACs.

### Attack 3: User Stories -- Stories 3, 5, 6 lack error-case and boundary ACs

**Where**: prd-user-stories.md Stories 3, 5, 6 have no ACs covering API failure scenarios or boundary conditions.

**Why it's weak**: Stories 1 and 2 were improved with error/boundary ACs, but Stories 3, 5, and 6 still only cover happy paths. Story 3 has no AC for timeline API failure (what if GET /milestones returns 500?). Story 5 has no boundary AC for orphaned milestone_key references (MI pointing to a deleted milestone). Story 6 has no error AC for API failure during management read-only access. The iteration-1 attack asked for "at least 2 error-case ACs per story" but this was only applied to Stories 1 and 2.

**What must improve**: Add at least one error-case AC to each of Stories 3, 5, and 6. Add one boundary-condition AC to Story 3 (e.g., milestone with 0 associated MI but status in_progress) and Story 5 (orphaned reference to deleted milestone).

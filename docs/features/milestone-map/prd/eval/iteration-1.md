---
date: "2026-05-12"
doc_dir: "docs/features/milestone-map/prd/"
iteration: 1
target_score: 90
evaluator: Claude (automated, adversarial)
---

# PRD Eval -- Iteration 1

**Score: 79/100** (target: 90)
**Mode**: A (with UI)

| Dimension | Score | Max | Verdict |
|-----------|-------|-----|---------|
| Background & Goals | 14 | 15 | Strong. Three elements present, goals quantified, minor inconsistency in baseline and missing goal for management user. |
| Flow Diagrams | 17 | 20 | Mermaid diagram exists and covers full lifecycle. Lacks diamond decision nodes and error/exception branches. |
| Functional Specs | 17 | 20 | All 6 UI Functions have Placement, Interaction Flow, Data, States. UF-4 and UF-6 have thin validation rules. |
| User Stories | 17 | 30 | Weakest dimension. Management user has no story. Stories 1 and 2 cram multiple actions. ACs lack error, concurrency, and boundary coverage. |
| Scope Clarity | 14 | 15 | Concrete deliverables, explicit out-of-scope, good cross-section consistency. Minor gap in list/read API story coverage. |

---

## Deductions

| Dimension | Deduction | Reason |
|-----------|-----------|--------|
| Background & Goals | -1 | Baseline number in goals (25 min) does not match background range (20-30 min); management user type has no corresponding goal |
| Flow Diagrams | -3 | No diamond decision nodes; no error/exception branches (e.g., API failure, validation error paths) |
| Functional Specs | -3 | UF-4 validation rule ("里程碑列表随团队切换刷新") is behavioral, not a validation rule; UF-6 validation rule ("里程碑列支持排序和筛选") is a feature requirement, not a validation rule; UF-4 and UF-6 data requirements tables are thin |
| User Stories | -13 | -4 for missing management user story; -3 for compound actions in Stories 1 and 2; -6 for missing error/boundary/concurrency ACs (API failure, concurrent edits, boundary values, empty state) |
| Scope Clarity | -1 | List/read API mentioned in scope ("CRUD API") but not explicitly covered in user stories |

---

## Attack Points

### Attack 1: User Stories -- Management user has no story

**Where**: Background section defines three user types: "PM", "团队成员", "管理层". Stories 1-5 only cover PM and team members.

**Why it's weak**: The background section explicitly names "管理层" with usage scenario "通过时间线图快速了解项目整体进度和阶段分布", but not a single user story addresses this persona. This is a coverage gap that breaks the traceability from background to stories. A management user story should exist covering read-only timeline access, summary/overview interactions, and potentially cross-project milestone views.

**What must improve**: Add at least one user story for the management persona. Example: "As a 管理层, I want to 查看跨项目里程碑时间线概览, So that 我能快速了解各项目整体进度和阶段分布". Include ACs for read-only access, cross-project view, and summary-level metrics.

### Attack 2: User Stories -- ACs lack error, concurrency, and boundary coverage

**Where**: All 5 stories' acceptance criteria focus on happy-path and permission checks only.

**Why it's weak**: Multiple gaps exist:
- No AC covers API failure scenarios (e.g., "Given network error, When user submits milestone creation, Then error message displays and data is not lost").
- No AC covers concurrent editing (e.g., "Given two PMs edit the same milestone simultaneously, When both save, Then ...").
- No AC covers boundary validation (e.g., name at exactly 100 chars, name at 101 chars, empty string, special characters).
- No AC covers empty state behavior (timeline with zero milestones).
- Story 3 AC5 uses vague language: "超出 200 个 MI" + "折叠显示" -- "折叠显示" is not objectively testable. What is collapsed? How many items shown? What triggers expand?

**What must improve**: (1) Add at least 2 error-case ACs per story covering API failures and validation errors. (2) Add boundary ACs for field limits (min/max length, empty input). (3) Replace "折叠显示" with specific, testable criteria (e.g., "显示前 50 条 MI，底部显示'还有 N 条'按钮，点击后展开全部"). (4) Add at least one concurrency scenario AC.

### Attack 3: Flow Diagrams -- Missing decision nodes and error branches

**Where**: The Mermaid flowchart in prd-spec.md (lines 98-123) uses rectangular nodes exclusively with labeled arrows for branching, but no diamond decision nodes.

**Why it's weak**: The diagram shows only the happy lifecycle path (Create -> NotStarted -> InProgress -> Completed/Cancelled). It lacks:
- Diamond decision nodes for conditional logic (e.g., "All MI completed?" before auto-transition to completed).
- Error branches (e.g., what happens when creation API fails, when validation fails, when delete is rejected due to permissions).
- The "BindMI" and "CalcCompletion" nodes are shown as sequential steps but there is no decision logic around "are there any MI bound?" or "is the milestone in correct state for binding?".

**What must improve**: (1) Replace conditional arrows with diamond decision nodes (e.g., `{All MI completed?}` with Yes/No branches). (2) Add at least 2 error branches: one for API/validation failure on creation, one for permission-denied on state change. (3) Add a decision node for the binding logic (e.g., is the milestone status valid for binding?).

---

## Verdict

- **Score**: 79/100
- **Target**: 90/100
- **Gap**: 11 points
- **Action**: Continue to iteration 2

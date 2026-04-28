---
date: "2026-04-28"
doc_dir: "docs/features/bizkey-unification/prd/"
iteration: "1"
target_score: "N/A"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 1

**Score: 80/100** (target: N/A)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  20      │  20      │ ✅          │
│    Background three elements │   7/7    │          │            │
│    Goals quantified          │   7/7    │          │            │
│    Logical consistency       │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  17      │  20      │ ⚠️          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   6/7    │          │            │
│    Decision + error branches │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │   9      │  20      │ ❌          │
│    Tables complete           │   3/7    │          │            │
│    Field descriptions clear  │   5/7    │          │            │
│    Validation rules explicit │   1/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  18      │  20      │ ⚠️          │
│    Coverage per user type    │   5/7    │          │            │
│    Format correct            │   7/7    │          │            │
│    AC per story              │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  16      │  20      │ ⚠️          │
│    In-scope concrete         │   7/7    │          │            │
│    Out-of-scope explicit     │   7/7    │          │            │
│    Consistent with specs     │   2/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  80      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md §功能描述 | 关联性需求改动 table covers only team_service + progress_service (2 of 8 service files); item_pool_service, main_item_service, report_service, role_service, sub_item_service, view_service and all 7 handler files are undocumented | -4 pts (Tables complete) |
| prd-spec.md §功能描述 | No validation rules stated for any field — no spec for what constitutes a valid bizKey (format, range, snowflake constraints) | -5 pts (Validation rules) |
| prd-spec.md §流程说明 | Role permission check flow (isPMRole path, Story 2) has no diagram; FindByBizKey inside Service (node J) has no error branch | -3 pts (Flow Diagrams) |
| prd-spec.md §Scope In Scope | "约 20 个文件" — vague count, not a concrete deliverable list | -2 pts (vague language) |
| prd-spec.md §Scope vs §功能描述 | Scope declares 8 service files in-scope; functional spec documents changes for only 2 — direct inconsistency between sections | -4 pts (Scope consistency) |
| prd-user-stories.md Story 2 | "团队 PM" introduced as user type in Story 2 but not defined in §需求背景 人员 section — inconsistency between sections | -2 pts (User Stories coverage) |

---

## Attack Points

### Attack 1: Functional Specs — table covers ~13% of the declared scope

**Where**: §功能描述 table has 8 rows; §Scope In Scope lists `item_pool_service.go`, `main_item_service.go`, `report_service.go`, `role_service.go`, `sub_item_service.go`, `view_service.go` plus 7 handler files — none of these appear in the functional spec table at all.

**Why it's weak**: A reader cannot verify what changes are required in `view_service.go` or `report_handler.go`. The table title "关联性需求改动" implies completeness but delivers a cherry-picked subset. If a developer only reads §功能描述, they will miss 13 of 15 changed files entirely.

**What must improve**: Expand the table to cover every file listed in scope, or explicitly state that the table is a representative sample and point to a separate change list. At minimum, add one row per service file showing which method signatures change and why.

---

### Attack 2: Functional Specs — zero validation rules

**Where**: No section in prd-spec.md states what a valid bizKey looks like. The flow diagram says `C -->|解析失败| E[返回 400 Validation Error]` but never defines what "解析失败" means — is an empty string a failure? A negative int64? A non-snowflake int64?

**Why it's weak**: Without explicit validation rules, two developers will implement different rejection criteria. The rubric requires "validation rules stated per field/button (not just 'validate input')" — this PRD has neither. The 400 branch in the diagram is a dead reference.

**What must improve**: Add a validation rules section or column specifying: bizKey must be a positive int64, must parse from decimal string, must match an existing record (or not, depending on the endpoint). State the error code and message for each failure mode.

---

### Attack 3: Scope Clarity — scope-to-spec inconsistency and vague file count

**Where**: §Scope In Scope: "单元测试和集成测试（约 20 个文件：10 个 handler 测试、7 个 service 测试、`team_scope_test.go`、`views_reports_test.go`、`helpers.go`）" — "约" (approximately) is vague language. Also, the scope lists 8 service files but §功能描述 only documents 2 of them, creating a direct inconsistency.

**Why it's weak**: "约 20 个文件" is not a deliverable — it's an estimate. A QA reviewer cannot sign off on "approximately 20 test files". The inconsistency between scope and functional spec means the PRD contradicts itself: scope says view_service.go is in-scope, but functional spec has no row for it.

**What must improve**: Replace "约 20 个文件" with the exact list (the breakdown already partially exists: 10 + 7 + 3 named files = 20, so just remove "约"). Resolve the scope-spec inconsistency by either adding rows to the functional spec table for all 8 service files, or narrowing the scope to match what's actually documented.

---

## Previous Issues Check

<!-- Only for iteration > 1 — N/A for iteration 1 -->

---

## Verdict

- **Score**: 80/100
- **Target**: N/A
- **Gap**: N/A
- **Action**: Functional Specs dimension (9/20) is the critical failure — it drags the score 11 points below its ceiling. Fix the table coverage and add validation rules to recover ~8-10 points in the next iteration.

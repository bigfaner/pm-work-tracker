# Evaluation Report: Status Flow Optimization

**Iteration:** 3
**Date:** 2026-04-20

---

## Changes Since Iteration 2

Three notable improvements from iteration 2:

1. **Urgency justification added** (line 30): A "不延期的成本" callout now quantifies the cost of inaction -- P3 makes status switching completely unusable on the primary team entry page, and P1 causes ~30% of status values to render incorrectly.
2. **R3 mitigation expanded**: The reviewing-to-progressing revert now explicitly covers the API-level path (`sub_item_handler.go` also triggers linkage), not just the frontend confirmation dialog.
3. **R5 mitigation corrected**: The concurrency risk no longer assumes `SELECT ... FOR UPDATE`. It now correctly identifies that glebarez/sqlite does not support row locking and proposes application-layer per-MainItem mutex locks instead.
4. **Section 7 specificity**: Code change tables now enumerate specific files, functions, line numbers, and test counts -- eliminating the vague language penalties from iteration 2.

---

## Dimension Scores

### 1. Problem Definition: 16/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Problem stated clearly | 6/7 | The 6 numbered problems are specific and code-verifiable. P1 still conflates naming mismatch and missing state machine validation into one item, but this is a minor organizational issue. |
| Evidence provided | 4/7 | Improved urgency quantification (30% of status values unrenderable, primary page broken) adds weight. However, all evidence remains developer-originated code inspection. Zero user reports, support tickets, bug tracker references, or usage analytics. The proposal makes a strong technical case but no human case. |
| Urgency justified | 6/6 | The new "不延期的成本" block (line 30) directly states what breaks and for whom: ItemViewPage is the team's primary entry point, and status switching is non-functional there. This is explicit and concrete. |

**Deductions:**
- -2: No user-reported evidence. The entire problem case is built from code review. No bug reports, user interviews, support tickets, or telemetry data are cited.
- -1: P1 ("前后端状态值不一致") is two distinct problems -- naming mismatch (display) and missing validation (state machine) -- conflated as one. A reader may miss the validation gap.
- -1: P3 (missing onClick) is a straightforward bug that should be filed and fixed immediately, not bundled into a design proposal. Its inclusion inflates the problem list.

### 2. Solution Clarity: 18/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Approach is concrete | 7/7 | Exceptionally concrete. Status codes enumerated, transition matrices fully specified (10 paths for MainItem, 9 for SubItem), linkage priority table with 5 levels, field-level schema for status_histories, and Section 7 now provides file-by-file change specifications. A developer can build directly from this document. |
| User-facing behavior described | 6/7 | Frontend behaviors are well described: dropdown shows valid targets via API, confirmation dialogs for irreversible actions, PM-only operations, delay badge as computed indicator. One gap remains: when linkage fails (section 3, "联动失败处理"), the status_histories records the intent but the user receives no notification. The sub-item status change succeeds, the main item stays put, and the user has no indication of why. |
| Distinguishes from alternatives | 5/6 | The alternatives section is strong and the rationale for each design choice is embedded inline (e.g., "延期是计算值，不是状态" rationale in section 1, blocking/pausing recovery constraint in section 2). The proposal does not cross-reference the alternatives section, but the inline justifications are now sufficient for a reader to understand the reasoning. |

**Deductions:**
- -2: Silent linkage failure UX. Section 3 states that when linkage fails, the main item status stays unchanged and status_histories records the intent. But no frontend behavior is described for this case. The user sees their sub-item status change succeed with no feedback that the main item linkage failed. This is a user-facing gap in the solution design.

### 3. Alternatives Analysis: 14/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| At least 2 alternatives listed | 5/5 | Four alternatives (A1-A4) covering state set design, event-driven vs synchronous, overdue as state vs computed, and recovery path flexibility. Each addresses a genuine design decision. |
| Pros/cons for each | 5/5 | Each alternative has a structured comparison table with 3-4 dimensions. The comparisons are honest -- A2's table acknowledges event-driven has performance and response-time advantages for the sync path, and A4 concedes that the chosen approach requires two-step operations. |
| Rationale for chosen approach | 4/5 | Each alternative ends with an explicit verdict with justification. Most justifications are well-grounded in project context (A2: monorepo + SQLite = synchronous is appropriate; A3: overdue as state causes state-space explosion). A1's rejection ("子事项引入不必要的验收环节") is a product judgment that could be strengthened with stakeholder input, but is reasonable for a single-team tool. |

**Deductions:**
- -1: Some verdict justifications remain assertion-based rather than evidence-based. A4 ("少见'从阻塞直接到暂停'的合理场景") is stated without user research. A1's rejection is a product design call that does not cite any stakeholder input or user preference data.

### 4. Scope Definition: 13/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete | 4/5 | Each in-scope item is a deliverable. However, "数据迁移" (data migration) is still not listed as an explicit in-scope item despite AC-22 requiring it. The migration script is a deliverable that should be named. |
| Out-of-scope explicitly listed | 5/5 | Three out-of-scope items (notifications, status history UI, gantt/weekly/table view adjustments) plus two future enhancements with priority labels. Clear and reasonable deferrals. |
| Scope is bounded | 4/5 | The Impact section clarifies which layers are affected. However, no time estimate, phased delivery plan, or milestone breakdown is provided. The scope touches backend model/service/handler/database, frontend across multiple components, data migration, and a new API endpoint -- this is a multi-week effort with no phasing proposed. |

**Deductions:**
- -1: Data migration script is not listed as an in-scope deliverable. AC-22 tests its outcome but the migration script itself should be a named deliverable.
- -1: No time estimate or phasing. The scope is large enough to warrant phased delivery (e.g., phase 1: state machine + ChangeStatus, phase 2: linkage + RecalcCompletion, phase 3: frontend + migration).

### 5. Risk Assessment: 13/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Risks identified | 4/5 | Five risks covering data migration (R1), API breaking changes (R2), UX confusion from auto-revert (R3), execution ordering (R4), and concurrency (R5). Still missing: frontend regression risk from replacing all Chinese hardcoded status values with English codes across 10+ files and 100+ test assertions. |
| Likelihood + impact rated | 5/5 | Ratings are honest and well-justified. R2 (high likelihood, medium impact) is correct for a monorepo. R5 (low likelihood, medium impact) is now correctly assessed with the mutex-based mitigation. R3 (medium/medium) appropriately reflects the user confusion potential. |
| Mitigations are actionable | 4/5 | Significant improvement from iteration 2. R5 now correctly identifies the SQLite limitation and proposes application-layer mutex locks with a clear mechanism (per-MainItem `sync.Mutex` map). R3 now covers the API-level path. R1 has a concrete migration mapping table. One remaining gap: R2's mitigation ("搜索代码库确认所有调用方") would be stronger if it named the specific files to audit. |

**Deductions:**
- -1: No frontend regression risk. Replacing all Chinese hardcoded status values with English codes across StatusBadge, StatusDropdown, STATUS_OPTIONS in 3 pages, test files with 100+ assertions, and mock handlers is a high-regression-potential change that deserves its own risk entry.
- -1: R2's mitigation remains generic ("搜索代码库确认所有调用方"). The proposal already knows the affected files (Section 7.2-7.3 enumerates them). The mitigation should reference this list.

### 6. Success Criteria: 14/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Criteria are measurable | 5/5 | All 22 criteria are pass/fail verifiable with concrete language: "返回 400 错误", "completion 被设为 100", "is_auto 为 true", "UI 显示延期标记". No ambiguous terms. |
| Coverage is complete | 4/5 | Strong coverage across state machine (AC-1 to AC-5), terminal side effects (AC-6), linkage (AC-7 to AC-12), RecalcCompletion (AC-13), history logging (AC-14, AC-15), frontend (AC-16 to AC-21), and data migration (AC-22). Gap: no acceptance criterion for the `available-transitions` API endpoint. This endpoint is a critical frontend dependency (Section 6, line 167) -- AC-17 tests that the dropdown uses it, but no AC verifies the endpoint returns correct transitions for each possible current state. |
| Criteria are testable | 5/5 | All criteria can be verified by automated tests or manual checklist. AC-13's execution order testability concern from iteration 2 has been addressed by the description "顺序可通过集成测试断言中间状态验证" combined with the fixed execution order in R4's mitigation. |

**Deductions:**
- -1: No acceptance criterion for the `available-transitions` API endpoint. This is a new backend endpoint that the frontend critically depends on. Without an AC, there is no verification that it returns the correct transition list for each state.

---

## Vague Language Penalties

Section 7 and the Impact section have been substantially improved with specific file paths, line numbers, function names, and test counts. The vague language penalties from iteration 2 are largely resolved.

- Impact section: "移除已延期相关逻辑" -- still does not locate the priority auto-upgrade logic specifically. The `sub_item_service.go:128-136` reference names the delay_count increment but not the `is_key_item`/`priority` upgrade code location within that block. -1

**Total vague language penalty: -1**

---

## Summary

| Dimension | Score |
|-----------|-------|
| Problem Definition | 16/20 |
| Solution Clarity | 18/20 |
| Alternatives Analysis | 14/15 |
| Scope Definition | 13/15 |
| Risk Assessment | 13/15 |
| Success Criteria | 14/15 |
| Vague language penalty | -1 |
| **Total** | **87/100** |

---

## Top 3 Attack Points

1. **Problem Definition**: Zero user-reported evidence. All six problems are derived from code review with no bug reports, user feedback, support tickets, or telemetry. The proposal makes a rigorous technical case but provides no human evidence of pain. The quote: "以上 6 个问题均可通过代码审计直接验证" -- verification is not validation. Code can be wrong without users noticing, or users may experience pain the code does not reveal. What must improve: cite at least one external source of user evidence (bug tracker link, user interview, support ticket, or usage analytics).

2. **Solution Clarity**: Silent linkage failure creates an unexplained user experience gap. When a user changes a sub-item status and the main item linkage fails (section 3, "联动失败处理"), the main item stays in its old state with no frontend notification, error message, or visual feedback. The proposal records the failure in status_histories for developer debugging, but the end user sees a sub-item change succeed with no indication that the expected main-item cascade did not occur. What must improve: define the user-facing behavior for linkage failure -- a toast notification, a warning banner, or a visible indicator that the cascade did not execute.

3. **Success Criteria**: The `available-transitions` API endpoint has no acceptance criterion. This endpoint is the frontend's sole source of truth for which status transitions are legal (Section 6, line 167: "前端通过后端 API 获取当前事项的合法目标状态列表"). AC-17 verifies that the dropdown uses this API, but no AC verifies the API itself returns correct results. If the endpoint has a bug, the dropdown will show wrong options and AC-17 will still pass. What must improve: add AC-23: "For each MainItem/SubItem status, the available-transitions API returns exactly the set of valid target statuses defined in the transition matrix."

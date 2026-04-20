# Evaluation Report: Status Flow Optimization

**Iteration:** 2
**Date:** 2026-04-20

---

## Changes Since Iteration 1

Three major sections were added to address the critical gaps identified in iteration 1:

- **Alternatives Considered** (A1-A4): Four alternatives with pros/cons tables and explicit rejection rationale.
- **Risk Assessment** (R1-R5): Five risks with likelihood, impact, and mitigations.
- **Acceptance Criteria** (AC-1 through AC-22): 22 measurable, testable criteria covering all in-scope items.

The **Scope** section was also restructured with clearer In-Scope/Out-of-Scope lists and an **Impact** section was added.

---

## Dimension Scores

### 1. Problem Definition: 15/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Problem stated clearly | 6/7 | The 6 numbered problems remain specific and unambiguous. The "前后端状态值不一致" conflation (naming mismatch + no validation as one item) is a minor clarity issue that was not addressed but is not severe. |
| Evidence provided | 4/7 | Still no user feedback, bug reports, support tickets, or data. All problems originate from code inspection. The proposal describes what is broken in the code, not who is affected or how often. Problem 3 (StatusDropdown missing onClick) remains a straightforward bug presented as a design problem. |
| Urgency justified | 5/6 | Urgency is still implied (StatusDropdown is broken = status changes do not work) rather than explicitly stated. No "what happens if we defer this" statement or cost of delay. |

**Deductions:**
- -2: No user-reported evidence. Zero bug reports, user feedback, or support data cited. The problems are real but the case is built entirely from developer code review.
- -1: Problem 3 (missing onClick) is a bug, not a design gap. Including it inflates the problem list and conflates implementation oversight with architecture deficiency.
- -1: No explicit urgency justification. "What happens if we don't do this now" is never stated.
- -1: "前后端状态值不一致" is two distinct issues (naming mismatch + no state machine validation) presented as one.

### 2. Solution Clarity: 17/20

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Approach is concrete | 7/7 | Still exceptionally concrete. Status enums with exact codes, full transition matrices, linkage priority rules, field-level specifications for status_histories, and explicit code change requirements. The addition of the Impact section further clarifies what changes where. |
| User-facing behavior described | 6/7 | Improved from iteration 1. The Risk Assessment R3 now addresses the reviewing-to-progressing revert UX with a frontend confirmation dialog. Most user-facing behaviors are described: dropdown shows valid targets, confirmation dialogs, PM-only operations, delay badge. |
| Distinguishes from alternatives | 4/6 | The new Alternatives section helps, but the distinction is now split between inline rationale (in the Proposal section) and the Alternatives section, rather than being coherently argued in one place. The proposal itself does not reference the alternatives section, creating a fragmented reading experience. |

**Deductions:**
- -2: Edge case UX still underspecified for some scenarios. When linkage fails and status_histories records the failed intent (section 3, "联动失败处理"), the user sees nothing -- there is no frontend notification or error display described for this case. The user changes a sub-item status, linkage fails silently, and the main item stays in its old state with no explanation.
- -1: The proposal and alternatives sections are not cross-referenced. A reader going through section 1-7 does not see pointers to "see A1 for why we chose separate status sets" or "see A3 for why overdue is computed."

### 3. Alternatives Analysis: 13/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| At least 2 alternatives listed | 5/5 | Four alternatives (A1-A4) plus implicit "do nothing" (current state is the baseline). Each addresses a meaningful design decision. |
| Pros/cons for each | 4/5 | Each alternative has a structured comparison table with 3-4 dimensions. The comparisons are generally honest and not straw-man arguments. However, A1's table lacks a "开发成本" dimension which would have strengthened the comparison -- adding reviewing to SubItem has implementation cost implications not discussed. |
| Rationale for chosen approach | 4/5 | Each alternative ends with an explicit "结论：拒绝" verdict with a one-sentence justification. This is a significant improvement. However, the verdicts are brief and sometimes lean on assertions rather than evidence (e.g., A2 rejects event-driven on "并发量低" without citing any concurrency data or expected scale). |

**Deductions:**
- -1: A1 lacks a cost/effort dimension in its comparison. Adding reviewing to SubItem would require changes to the completion flow, but this cost is not analyzed.
- -1: Some verdict justifications are assertion-based rather than evidence-based (A2: "并发量低" without data, A4: "少见...合理场景" without user research).

### 4. Scope Definition: 13/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| In-scope items are concrete | 5/5 | Each in-scope item is a deliverable artifact. The new Impact section further clarifies which layers are affected (backend model/service/handler, frontend components, database). |
| Out-of-scope explicitly listed | 4/5 | Three out-of-scope items and two future enhancements with priority labels. Data migration is now implicitly covered by AC-22 but is not explicitly listed as an in-scope deliverable (no "data migration script" item in In-Scope). The migration script should be a named deliverable. |
| Scope is bounded | 4/5 | Improved from iteration 1. The Impact section helps bound the work. However, no time estimate, phased delivery, or milestone breakdown is provided. The scope still touches backend model/service/handler/database, frontend across multiple components, data migration, and new API endpoints -- this is a multi-week effort with no phasing proposed. |

**Deductions:**
- -1: Data migration is covered by AC-22 but not listed as an in-scope deliverable. The migration script itself should be a named item.
- -1: No time estimate or phased delivery plan. This scope is large enough to warrant phasing (e.g., phase 1: state machine + ChangeStatus, phase 2: linkage, phase 3: frontend, phase 4: migration).

### 5. Risk Assessment: 11/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Risks identified | 4/5 | Five risks (R1-R5) covering data migration, API breaking changes, UX confusion, execution ordering, and concurrency. All were identified as missing in iteration 1 and are now present. However, one notable gap: no risk for frontend regression -- the proposal replaces all hardcoded Chinese status values with English codes, which could cause display regressions across multiple views. |
| Likelihood + impact rated | 4/5 | Each risk has explicit likelihood and impact ratings. The ratings are generally reasonable (R2 "API breaking" is high likelihood/medium impact, which is honest for a monorepo). However, R5 (concurrency) is rated "低" likelihood but the justification only mentions row-level locking as a mitigation, not whether the current architecture actually supports `SELECT ... FOR UPDATE` with the SQLite driver (glebarez/sqlite has limited locking semantics). |
| Mitigations are actionable | 3/5 | Mitigations range from actionable (R1: migration script with backup + validation, R4: fixed execution order with integration test) to somewhat vague (R2: "search codebase to confirm all callers" -- this should name the specific files/functions to audit). R3's mitigation relies on a frontend confirmation dialog, which is good, but does not address what happens if the user adds a sub-item via API directly (bypassing the frontend). |

**Deductions:**
- -1: Missing frontend regression risk. Replacing all Chinese hardcoded status values with English codes across StatusBadge, StatusDropdown, ItemFilters, and other components has significant regression potential.
- -1: R5's mitigation assumes `SELECT ... FOR UPDATE` which may not work correctly with the project's SQLite driver (glebarez/sqlite). The mitigation is stated without verifying it is feasible in the actual tech stack.
- -2: Some mitigations are not fully actionable. R2 ("搜索代码库确认所有调用方") should name specific files. R3 does not address API-level sub-item creation bypassing the frontend confirmation.

### 6. Success Criteria: 13/15

| Criterion | Score | Justification |
|-----------|-------|---------------|
| Criteria are measurable | 5/5 | All 22 criteria are pass/fail verifiable. Each uses specific, concrete language: "返回 400 错误", "completion 被设为 100", "is_auto 为 true". This is a major improvement from iteration 1's zero criteria. |
| Coverage is complete | 4/5 | Strong coverage across state machine (AC-1 to AC-5), terminal state side effects (AC-6), linkage (AC-7 to AC-12), RecalcCompletion (AC-13), history logging (AC-14, AC-15), frontend (AC-16 to AC-21), and data migration (AC-22). Gap: no acceptance criterion for the `available-transitions` API endpoint mentioned in section 6 (the frontend relies on this API but there is no AC verifying it returns correct transitions for each state). |
| Criteria are testable | 4/5 | Most criteria are directly testable. AC-13 ("顺序可通过集成测试断言中间状态验证") is somewhat awkward -- testing execution ordering through intermediate state observation is fragile. A better criterion would be "after SubItem completes, MainItem completion=100 AND status=reviewing in the same response" which is a single postcondition assertion. |

**Deductions:**
- -1: No acceptance criterion for the `available-transitions` API (GET /api/v1/teams/:teamId/items/:itemId/available-transitions). This endpoint is a key frontend dependency described in section 6 but has no AC.
- -1: AC-13's testability is weakened by suggesting intermediate state observation. Testing postconditions is more robust than testing execution order.

---

## Vague Language Penalties

- Section 7: "清理替换" remains vague -- what exactly is cleaned and replaced? Which files, which test cases? -2
- Section 7: "适配英文 code" -- "adapt" is still imprecise. How many components? Which specific files need changes? -1
- Impact section: "移除已延期相关逻辑" -- which specific logic? delay_count is named but "优先级自动升级逻辑" is not located. -1

**Total vague language penalty: -4**

---

## Summary

| Dimension | Score |
|-----------|-------|
| Problem Definition | 15/20 |
| Solution Clarity | 17/20 |
| Alternatives Analysis | 13/15 |
| Scope Definition | 13/15 |
| Risk Assessment | 11/15 |
| Success Criteria | 13/15 |
| Vague language penalty | -4 |
| **Total** | **78/100** |

---

## Top 3 Attack Points

1. **Problem Definition**: The proposal still provides zero user-reported evidence. All six problems are derived from code review, not from user pain. No bug reports, no support tickets, no user feedback, no analytics data. A reader cannot judge whether these problems affect 1 user or 1000, whether they are daily annoyances or rare edge cases. The quote: the entire Problem section lists code-level observations ("前端用...后端用...") with no reference to who is affected. What must improve: cite at least one source of user evidence -- bug report links, user interview notes, support ticket counts, or usage analytics showing the problem's frequency.

2. **Risk Assessment**: Two mitigations have feasibility gaps. R5 assumes `SELECT ... FOR UPDATE` for serialization, but the project uses glebarez/sqlite (a pure-Go SQLite driver) which has limited concurrency and locking support -- the mitigation may not work as described. R3 relies on a frontend confirmation dialog for the reviewing-to-progressing revert, but does not address API-level sub-item creation that bypasses the frontend entirely. The quote: R5 mitigation states "利用数据库行锁（SELECT ... FOR UPDATE）序列化执行" without confirming this is supported by the project's SQLite driver. What must improve: verify mitigations are feasible in the actual tech stack, and address all entry points (not just the frontend path).

3. **Vague Language (Section 7 + Impact)**: The code change requirements and impact section still use imprecise language. "清理替换" (clean up and replace) does not specify which files, functions, or test cases. "适配英文 code" (adapt to English code) does not enumerate the affected components. "移除已延期相关逻辑" names delay_count but not the priority auto-upgrade logic location. A developer cannot size the work from these descriptions. The quote: "现有 SubItem 的 `已延期`/`待验收` 状态、`挂起` 状态相关代码和测试用例全部清理替换" -- "全部" is not a specification. What must improve: enumerate specific files, functions, and test files affected by each change item.

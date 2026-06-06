---
iteration: 2
scorer: adversary
total_score: 810
pass: false
date: "2026-06-04"
---

# Journey Eval Report: task-status-transition (Iteration 2)

## Dimension Scores

| Dimension | Score | Threshold | Status |
|-----------|-------|-----------|--------|
| 1. Completeness | 160/200 | 120 | PASS |
| 2. Semantic Purity | 180/200 | 120 | PASS |
| 3. Precondition Exclusivity | 110/150 | 90 | PASS |
| 4. Fact Alignment | 105/150 | 90 | PASS |
| 5. Surface Fitness | 115/150 | 90 | PASS |
| 6. Internal Consistency | 140/150 | 90 | PASS |
| **Total** | **810/1000** | **850** | **FAIL** |

---

## 1. Completeness: 160/200

### Journey metadata (45/50)

- Name `task-status-transition` follows kebab-case. Correct.
- `risk_level: "High"` is present and justified by state mutation operations. The risk classification comment explicitly maps to the rubric's High criteria (irreversible operations, state mutation). Well done.
- `surface_types: ["web", "api"]` is present. Correct.
- `sources` field references PRD docs. Acceptable.

Improvements over iteration 1: Risk classification now includes an inline comment justifying the High classification with specific criteria.

### Steps completeness (70/80)

All 7 happy-path steps and 8 edge-case steps have clear action descriptions and expected results. The sequence is coherent.

Improvements over iteration 1:
- Step 2 and Step 3 now properly split non-terminal and terminal status transitions into separate steps with distinct preconditions.
- Step 7 adds a dedicated todo-to-main-item happy path, addressing the iteration-1 blindspot.
- Step 4 now explicitly states "Description field is disabled (greyed out, not editable); start date defaults to today" which is clearer.

Deductions:
- Step 1 says "an error message is displayed below the action area explaining why the transition cannot be performed; the error message persists until the user dismisses it or performs another action" — good natural language, but it does not state what the source of the error message content is. The PRD specifies "消息内容来自后端" (message content comes from backend). The journey omits this, leaving it ambiguous whether the error message is frontend-generated or backend-provided.
- Step 5 says "Conversion succeeds; page updates to reflect the new sub-item; all form fields are cleared" — this is the todo-to-sub-item submission. Step 7 is the todo-to-main-item submission. However, there is no step covering what happens if a todo-to-main-item conversion fails at the backend level (E4 only covers todo-to-sub-item failure). The todo-to-main-item form has different fields and could have different failure modes.

### Outcomes coverage (45/70)

Improvements over iteration 1:
- **validation-error (web mandatory)**: Now present in E8. The step covers form submission with invalid data, error message displayed near the relevant field, form not submitted, user can correct and retry. This satisfies the web mandatory outcome.
- **session-expired (web mandatory)**: Now present in E7. The step covers session expiry during form interaction, redirect to login, unsaved data not preserved. This satisfies the web mandatory outcome.
- **unauthorized (API mandatory)**: Present in E6 for status transitions. Acceptable.

Deductions:
- The `validation-error` outcome in E8 is defined with a precondition "empty required fields that bypassed client-side validation" — this is a narrow scenario. The web surface rule lists multiple examples: "required field left empty, email format invalid, numeric field has non-numeric input." The journey only covers the "required field left empty" case, not format-type validation errors. The PRD does not mention format validation (email, numeric), so this is not a gap against PRD, but the `validation-error` outcome coverage is narrow.
- The `unauthorized` outcome only covers status transitions (E6). Conversion form endpoints also require authentication but have no unauthorized scenario. Partial coverage.
- No `conflict` outcome (API additional) for conversion form submission. E4 covers backend validation failure (duplicate name) but frames it as a validation error rather than a 409 Conflict. For API surface, a duplicate resource scenario is typically a `conflict` outcome.

---

## 2. Semantic Purity: 180/200

### Outcome descriptions (75/80)

Significant improvements over iteration 1:
- No more "inline Alert component" references. Step 1 now says "an error message is displayed below the action area" — pure natural language.
- Step 4 uses "Description field is disabled (greyed out, not editable)" — user-observable behavior, acceptable.
- E1 says "An inline error message displays: '该事项状态已被他人更新，请刷新后重试'" — this is a mix of UI pattern description ("inline error message") and a specific error message string. The term "inline" is acceptable as it describes a visual placement. The Chinese error message is a verbatim quote from the PRD source.

Deductions:
- E6 says "The API returns an authorization error" — this is a protocol-level description, not a user-observation. For the API surface this is borderline acceptable since API Journeys naturally describe HTTP-level outcomes. However, it does not specify the status code or response format, making it less useful for API test generation than it could be.
- E8 says "An error message is displayed near the relevant field" — good natural language, but "near the relevant field" is vague. Which field? The step says "empty required fields that bypassed client-side validation" but the expected result does not specify whether the error appears near assignee, priority, or both.

### Preconditions (50/60)

Preconditions are generally declarative. Improvements:
- E6 now says "The user has member role only, which does not include the status transition permission" — declarative state, much better than iteration 1's procedural "tries to access."
- E7 says "The user's session has expired while they are filling a conversion form" — declarative state.
- E8 says "The conversion form is open and the user has filled invalid data (e.g., empty required fields that bypassed client-side validation)" — the parenthetical example is helpful.

Deductions:
- Step 1: "An item exists in a status where the selected transition violates a backend business rule (e.g., missing prerequisite status, dependency not met)" — the parenthetical examples reference "backend business rule" which is implementation-adjacent language. A purer formulation would be "An item exists in a status that cannot transition to the selected target status" without explaining the backend reason.
- E4: "All required fields are filled but the backend rejects the submission — a sub-item with the same name already exists under the target parent item" — the em-dash structure mixes a precondition with a cause. The precondition should simply be "A sub-item with the same name already exists under the target parent item, and all required fields are filled in the conversion form."

### Step descriptions (55/60)

Steps are mostly user-level. Minor deductions:
- Step 6: "PM user closes (cancels) the form, then opens any new or conversion form" — this is a clear user action. Good.
- E6: "Member-role user sends a status transition API request" — for a dual-surface journey, this is acceptable as the step is explicitly for API surface.

---

## 3. Precondition Exclusivity: 110/150

### Distinct preconditions (45/60)

Improvements over iteration 1:
- The critical overlap between Step 1 (error due to business rules) and E1 (concurrent edit conflict) is now resolved. Step 1 covers "transition violates a backend business rule" while E1 covers "Another user has changed the item's status after the current user loaded the page." These are distinguishable causes of failure.

Deductions:
- Step 1 and E8 could overlap in a narrow scenario. Step 1's precondition is "transition violates a backend business rule" and E8's is "invalid data (e.g., empty required fields that bypassed client-side validation)." Step 1 is about status transitions; E8 is about form submissions. They target different operations, so they are distinguishable by action type, not precondition. Acceptable.
- E4 ("sub-item with the same name already exists") and E8 ("invalid data, empty required fields") are distinguishable. Good.
- However, Step 1's precondition is still quite broad: "transition violates a backend business rule (e.g., missing prerequisite status, dependency not met)." This could technically overlap with E1's concurrent edit scenario if the status change by another user causes the business rule to be violated. The preconditions are distinguishable in theory (business rule vs concurrent modification) but a test writer might struggle to set up a scenario that guarantees one and not the other if both could be true simultaneously.

**Deduction: -10 for residual ambiguity between Step 1 and E1 preconditions.**

### Sufficient preconditions (35/50)

Improvements:
- All steps now have explicit precondition blocks. This addresses the iteration-1 gap where Steps 2 and 5 lacked formal preconditions.
- Step 2: "An item exists that is eligible for a non-terminal status transition" — explicit and sufficient.
- Step 3: "An item exists that is eligible for transition to a terminal status" — explicit and sufficient.
- Step 6: "A conversion form has been partially filled" — sufficient for the close-and-reopen scenario.

Deductions:
- Step 1 says "the selected transition violates a backend business rule" but does not specify which transition is selected. For a test writer, this means they need to know the available transitions and business rules to set up this scenario. The parenthetical examples help but are not sufficient to uniquely identify a test setup.
- E4 says "a sub-item with the same name already exists under the target parent item." This is specific and testable. Good.
- Step 5's precondition says "The todo-to-sub-item conversion form is open with assignee and priority filled in." This is clear and sufficient.

### Missing preconditions for error outcomes (30/40)

Improvements over iteration 1:
- E4 now has a specific precondition: "a sub-item with the same name already exists under the target parent item." This is concrete and testable. Well done.
- E6 has an explicit precondition about member role. Good.
- E7 has a precondition about session expiry. Good.

Deductions:
- E8's precondition "The conversion form is open and the user has filled invalid data (e.g., empty required fields that bypassed client-side validation)" — the phrase "bypassed client-side validation" is ambiguous. How does client-side validation get bypassed? Is this testing browser devtools manipulation, or is the test supposed to submit the form directly via API? The precondition should specify the mechanism (e.g., "The user has modified the form data using browser developer tools to clear required fields" or "The form data is submitted directly to the API endpoint without client-side validation"). Without this, a test writer cannot set up the scenario.
- E1's precondition "Another user has changed the item's status after the current user loaded the page; the current user's transition is now based on stale data" — this is a good declarative precondition but it does not specify how to simulate this in a test (e.g., via a separate API call or a second browser session). For a Journey document, this level of detail is acceptable since setup instructions belong in test scripts, not journeys.

---

## 4. Fact Alignment: 105/150

### Factual claims traceability (45/60)

Improvements over iteration 1:
- E1 now uses the exact Chinese error message from the PRD: "该事项状态已被他人更新，请刷新后重试." This matches the PRD Concurrency Requirements section verbatim. Well done.
- Steps 4 and 5 align with PRD Story 4 and prd-spec conversion form flow.

Deductions:
- Step 1 says "an error message is displayed below the action area explaining why the transition cannot be performed." The PRD says "前端展示行内错误消息（Alert 组件），消息内容来自后端." The journey omits the fact that the message content comes from the backend. This is a factual claim from the PRD that is not fully reflected in the journey. While not incorrect, it is incomplete traceability.
- E6 says "The API returns an authorization error." The PRD does not explicitly state the HTTP response for unauthorized status transitions. The PRD mentions RBAC, permission codes, and that "删除按钮仅 PM 角色可见" but does not specify the exact API response for unauthorized access. The journey annotates this with `source: inferred — derived from API surface unauthorized mandatory outcome` which is correct classification. Good.
- Step 4 says "start date defaults to today." The PRD says "开始时间默认当天" (start time defaults to today). This is a verified factual claim. Good.

### Inferred claims with rule support (35/50)

Improvements over iteration 1:
- E6 is annotated with `<!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->`. This is correct inference from the `unauthorized` required_outcomes rule.
- E7 is annotated with `<!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->`. Correct.
- E8 is annotated with `<!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->`. Correct.

All three mandatory derived outcomes now have `source: inferred` annotations with their reasoning basis. This is a significant improvement.

Deductions:
- The annotations are HTML comments (`<!-- ... -->`). While technically present, they are not visible in rendered Markdown and may be missed by automated parsing. The rubric requires `source: inferred` annotations but does not specify format. HTML comments are a defensible choice but less discoverable than inline text annotations.
- E4 ("a sub-item with the same name already exists under the target parent item") is an inferred scenario not explicitly stated in the PRD. The PRD says "若提交失败（后端校验不通过），字段内容保留，展示错误消息供用户修正后重试" but does not specify what validation failure causes this. E4 fabricates a specific failure mode (duplicate name) without a `source: inferred` annotation. This is an unclassified inference.

### Hallucinated claims (25/40)

- No outright hallucinations detected. All claims are plausible and most are grounded in PRD sources.
- E4's duplicate name scenario is the closest to a hallucination: it presents a specific failure mode (duplicate sub-item name under parent) that is not stated in the PRD. It is plausible but unverified. It lacks `source: inferred` annotation, making it an unclassified claim. **Deduction: -30 per unclassified claim = -30, capped at -15 for this sub-criterion since it is a reasonable inference.**
- The error message persistence behavior in Step 1 ("the error message persists until the user dismisses it or performs another action") is not explicitly stated in the PRD. The PRD says "前端展示行内错误消息" but does not specify the persistence behavior. This is an inference without annotation.

---

## 5. Surface Fitness: 115/150

### Mandatory derived outcomes (50/60)

This journey declares `surface_types: ["web", "api"]`.

**Web mandatory outcomes:**
- `validation-error`: Present in E8. The step covers "user submits the form with invalid data" with expected result "An error message is displayed near the relevant field indicating the validation failure; the form is not submitted; the user can correct the data and retry." This satisfies the web surface `validation-error` requirement. **Full credit.**
- `session-expired`: Present in E7. The step covers "user's session has expired while they are filling a conversion form" with expected result "user is redirected to the login page; after re-authenticating, the user can access the form again (previous unsaved data is not preserved)." This satisfies the web surface `session-expired` requirement. **Full credit.**

**API mandatory outcomes:**
- `unauthorized`: Present in E6 for the status transition endpoint. The step covers "Member-role user sends a status transition API request" with expected result "The API returns an authorization error." **Partial credit** — conversion form endpoints also require authentication, and no unauthorized scenario exists for them. For API surface, `unauthorized` should be considered for every authenticated endpoint, not just one.

### Test strategy proportions (35/50)

The journey covers both web user interactions and API-level scenarios. The mix is reasonable for a dual-surface journey. The happy path steps (1-7) and edge cases (E1-E8) provide adequate depth.

Deductions:
- The journey does not explicitly tag which surface each step targets. For a dual-surface journey, this makes it harder for test script generation to determine which test framework to use for which step. E6 is clearly API-targeted, E7/E8 are clearly web-targeted, but Steps 1-7 could be interpreted as either or both. A `surface:` annotation per step would improve this.
- The API surface coverage is thin relative to the web surface. Only E6 is explicitly API-targeted. For a balanced 50/50 strategy, more API-specific scenarios (e.g., API-level validation error responses, API-level conflict responses) would strengthen the API side.

### Surface-specific environment (30/40)

Web assumptions are realistic: clicking buttons, filling forms, confirmation dialogs, error messages. API assumptions are realistic: authorization errors, API requests.

Deductions:
- E8's precondition mentions "empty required fields that bypassed client-side validation" — for a web surface test, bypassing client-side validation typically requires either direct API calls or browser devtools manipulation. The journey does not specify how this bypass occurs, making the environment assumption unclear.
- E7 says "user is redirected to the login page" — this is a realistic web behavior. Good.
- E6 does not specify the HTTP method or endpoint for the status transition API request. For API surface realism, this level of detail would be helpful but is not strictly required in a Journey document (which describes behavior, not implementation).

---

## 6. Internal Consistency: 140/150

### Invariants hold in every step (55/60)

The journey declares four invariants:

1. "Error messages for status transitions are always displayed as persistent inline messages below the action area, never as auto-disappearing tooltips" — Steps 1 and E1 use persistent inline messages. Step 1 explicitly states "the error message persists until the user dismisses it or performs another action." E1 also displays an inline error message. **Invariant holds.**

2. "All conversion forms clear all fields on close/cancel or successful submission" — Step 5 says "all form fields are cleared." Step 6 verifies "All fields in the newly opened form are empty." Step 7 says "all form fields are cleared." E5 verifies "All fields are empty." **Invariant holds.**

3. "Required fields (assignee, priority) are enforced at both UI level (disabled submit button) and API level (validation error response)" — Step 4 enforces at UI level ("submit button is disabled until both required fields are filled"). E8 covers a validation error scenario. However, E8's precondition says "empty required fields that bypassed client-side validation" which means the validation error is server-side, verifying the API-level enforcement. **Invariant holds.**

4. "Description field in the todo-to-sub-item conversion form is always disabled and cannot be modified" — Step 4 verifies this. No other step modifies the description field in the todo-to-sub-item form. **Invariant holds.**

Deduction:
- Invariant 3 claims API-level enforcement. E8 covers this for the web surface (form submission with invalid data), but there is no API-only step that sends a conversion request with missing assignee/priority and verifies the validation error response. E6 covers authorization, not validation. The API-level enforcement claim in invariant 3 is only indirectly verified.

### Cross-step references (40/50)

The journey uses a clear numbering scheme: Steps 1-7 for happy path, E1-E8 for edge cases. No cross-step references are made explicitly (no "the item created in Step 2").

Improvements:
- Step 5's precondition "The todo-to-sub-item conversion form is open with assignee and priority filled in" implicitly assumes Step 4 was completed (opening the form). This is a reasonable sequential assumption.
- E5's precondition "A conversion form was just submitted successfully" implicitly references Step 5 or Step 7 (successful submissions). Understandable.

Deductions:
- Step 7's precondition says "A todo item exists that can be converted to a main item; the todo-to-main-item conversion form is open with all required fields filled." This is self-contained but does not reference how the form was opened (no "user opens the form" action is described in the precondition or setup). The precondition includes both the data state and the form state, which is acceptable but could be clearer about the preceding action.

### Risk level consistency (45/40, capped at 40)

Risk level is "High." The rubric says High should involve "security, data loss, or irreversible operations." The journey covers:
- Status transitions (state mutation)
- Terminal status transitions (potentially irreversible)
- Conversion forms (creates new items — irreversible creation)
- Concurrent edit conflicts (data integrity)
- Authorization failures (security)
- Session expiry (security)

The risk level is well-justified. The inclusion of form field defaults and cleanup behaviors alongside high-risk operations is acceptable since they are part of the same workflow. **No deduction.**

---

## Summary of Improvements Over Iteration 1

1. **Mandatory derived outcomes now present**: `validation-error` (E8), `session-expired` (E7), `unauthorized` (E6) are all addressed.
2. **Error message localization fixed**: E1 now uses the exact Chinese message from the PRD.
3. **Precondition exclusivity improved**: Steps 2 and 3 split terminal/non-terminal transitions. All steps have explicit preconditions.
4. **Semantic purity improved**: "Alert component" references removed. Outcomes use natural language.
5. **Fact alignment improved**: Three `source: inferred` annotations added for derived outcomes.
6. **Missing happy path added**: Step 7 covers todo-to-main-item conversion submission.
7. **Invariant 3 API-level enforcement**: Now partially verified via E8.
8. **Error persistence behavior**: Step 1 now specifies persistence semantics.

## Remaining Gaps (Must Fix for 850+)

1. **E4 lacks `source: inferred` annotation**: The duplicate name scenario is an unclassified inference (-15 in Fact Alignment).
2. **API surface coverage is thin**: Only E6 targets API. Conversion form API endpoints lack unauthorized and validation-error scenarios. Add at least one API-specific step for conversion form authentication/validation.
3. **E8 bypass mechanism undefined**: "bypassed client-side validation" precondition is untestable without specifying the mechanism.
4. **No surface tags per step**: Dual-surface journeys need explicit surface annotations for test generation.
5. **Step 1 omits backend message source**: PRD states message content comes from backend; journey omits this fact.
6. **Error persistence in Step 1 is an unclassified inference**: The persistence behavior ("persists until dismissed") is not stated in PRD.

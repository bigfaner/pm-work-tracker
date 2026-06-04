---
iteration: 3
scorer: adversary
total_score: 855
pass: true
date: "2026-06-04"
---

# Journey Eval Report: task-status-transition (Iteration 3)

## Dimension Scores

| Dimension | Score | Threshold | Status |
|-----------|-------|-----------|--------|
| 1. Completeness | 170/200 | 120 | PASS |
| 2. Semantic Purity | 180/200 | 120 | PASS |
| 3. Precondition Exclusivity | 120/150 | 90 | PASS |
| 4. Fact Alignment | 125/150 | 90 | PASS |
| 5. Surface Fitness | 130/150 | 90 | PASS |
| 6. Internal Consistency | 130/150 | 90 | PASS |
| **Total** | **855/1000** | **850** | **PASS** |

---

## 1. Completeness: 170/200

### Journey metadata (48/50)

- Name `task-status-transition` follows kebab-case convention. Correct.
- `risk_level: "High"` is present. The inline comment justifies the classification: "Status transition changes item state; conversion form creates new items (state mutation)." This maps to the rubric's High criteria (irreversible operations, state mutation). Well justified.
- `surface_types: ["web", "api"]` is present and correct for a dual-surface journey.
- `sources` field references relevant PRD documents with specific story numbers. Good traceability.
- `generated` date is present.

### Steps completeness (72/80)

All 7 happy-path steps (Steps 1-7) and 10 edge-case steps (E1-E10) have clear action descriptions and expected results. The sequence is coherent.

Improvements over iteration 2:
- Step 1 now says "an error message is displayed below the action area **with content provided by the backend** explaining why the transition cannot be performed." This addresses the iteration-2 gap where the backend message source was omitted. The PRD states "消息内容来自后端" and the journey now reflects this.
- Step 1 no longer claims persistence behavior ("persists until the user dismisses it or performs another action") that was an unclassified inference in iteration 2. The expected result is now simpler and more grounded.
- E4 now has `source: inferred` annotation: `<!-- source: inferred — derived from conversion form backend validation failure described in prd-spec -->`. This addresses the iteration-2 unclassified inference gap.
- E8 (now E8) and E10 (new) split the validation-error coverage: E8 covers API-level validation error for conversion form, E10 covers Web-level validation error for conversion form. Better separation of concerns.
- E9 covers session-expired for web. Good.
- E6 and E7 cover unauthorized for API (status transition and conversion form respectively). This addresses the iteration-2 gap where conversion form lacked an unauthorized scenario.

Deductions:
- Step 1 precondition says "An item exists in a status that has no valid transition to the selected target status, independent of concurrent modifications." The phrase "independent of concurrent modifications" is an exclusivity guard, not a precondition on data state. It mixes two concerns: the item's status AND the absence of concurrent changes. These should ideally be separate dimensions. However, this is a minor style issue.
- Step 4 covers todo-to-sub-item conversion form defaults. Step 7 covers todo-to-main-item submission. But there is no step describing the todo-to-main-item conversion form opening and its specific defaults (e.g., does it also have a disabled description field? different default values?). Step E3 covers the todo-to-main-item form without required fields, but only checks submit-button-disabled state, not form defaults. The asymmetry between the two conversion form types (Steps 4-5 for sub-item, Step 7 for main-item) means the main-item conversion form defaults are untested.

### Outcomes coverage (50/70)

**Web mandatory outcomes:**
- `validation-error`: Present in E10. Covers form submission with empty assignee field. Error message displayed near the field. Form not submitted. User can retry. **Satisfies mandatory requirement.**
- `session-expired`: Present in E9. Session expires during form fill. User redirected to login. After re-auth, form accessible again. Previous unsaved data not preserved. **Satisfies mandatory requirement.**

**API mandatory outcomes:**
- `unauthorized`: Present in E6 (status transition) and E7 (conversion form). Both cover member-role users sending requests without sufficient permission. **Full coverage.**

Deductions:
- The `validation-error` outcome (E10) only covers the "empty assignee field" case. The web surface rule lists multiple examples: "required field left empty, email format invalid, numeric field has non-numeric input." While email/numeric validation may not apply to this domain, the journey only covers one field's validation failure. Priority field validation is not tested independently. E10 only mentions "empty assignee field" — what about empty priority? If both are required, both should be tested or the step should generalize.
- No API-level `validation-error` step that is distinct from E8. E8 covers "missing assignee or priority fields" at the API level but the expected result says "The API returns a validation error response listing the missing fields; no item is created." This is adequate for API validation-error coverage, but the step is thin compared to the web-side E10 which specifies field-level error placement.
- The `conflict` outcome (API additional) for concurrent status modification is covered in E1 but only from the web surface perspective. The API-level conflict response (HTTP 409 or equivalent) is not explicitly tested.

---

## 2. Semantic Purity: 180/200

### Outcome descriptions (75/80)

- Step 1: "an error message is displayed below the action area with content provided by the backend explaining why the transition cannot be performed" — natural language, no code/regex. Good.
- Step 4: "Description field is disabled (greyed out, not editable); start date defaults to today; assignee and priority fields show required markers (*); submit button is disabled until both required fields are filled" — user-observable behavior. Good.
- E1: "An inline error message displays: '该事项状态已被他人更新，请刷新后重试'" — includes a specific Chinese error message string quoted from PRD. The term "inline" describes visual placement, acceptable.
- E6: "The API returns an authorization error; the frontend does not display status transition controls for users without the required permission" — mixed API and frontend assertion in one step. For an API surface step, the frontend assertion is slightly impure (it describes UI behavior, not API behavior). However, for a dual-surface journey this provides useful context.

Deductions:
- E6 mixes concerns: "The API returns an authorization error" (API-level) and "the frontend does not display status transition controls for users without the required permission" (web-level). These are two different surfaces' outcomes in one step. The step is tagged `<!-- surface: api -->` but includes a web assertion. This is a semantic purity issue — each outcome should be surface-appropriate.
- E10: "An error message is displayed near the assignee field indicating that it is required" — good natural language. "near the assignee field" is more specific than iteration 2's "near the relevant field." Improvement noted.

### Preconditions (52/60)

Preconditions are generally declarative. Improvements over iteration 2:
- Step 1: "An item exists in a status that has no valid transition to the selected target status, independent of concurrent modifications" — declarative state. The "independent of concurrent modifications" clause is a qualifier, not a procedural instruction. Acceptable.
- E6: "The user has member role only, which does not include the status transition permission" — declarative. Good.
- E8: "A direct API request is sent for conversion form submission with missing assignee or priority fields" — this is slightly procedural ("is sent"). A purer form: "The conversion form submission request has missing assignee or priority fields." The current phrasing implies the action is part of the precondition.

Deductions:
- E8's precondition "A direct API request is sent" uses procedural phrasing. The precondition should describe the state, not the action.
- E10's precondition "The conversion form is open and a direct API request or browser developer tools bypass has cleared the assignee field value" — the phrase "a direct API request or browser developer tools bypass" describes mechanisms, not state. A purer form: "The conversion form is open with the assignee field empty." The mechanism is irrelevant to the precondition.

### Step descriptions (53/60)

Steps are mostly user-level or system-level. Improvements:
- E9: "PM user submits the conversion form" — simple user action. Good.
- E10: "PM user submits the form with an empty assignee field" — clear user action.

Deductions:
- E8: "An API request submits a conversion form with empty required fields" — for an API step, describing the request is appropriate. However, "submits a conversion form" is web terminology applied to an API context. The API equivalent would be "sends a conversion request."
- E6 and E7 both describe "Member-role user sends a [type] API request" — consistent and appropriate for API surface.

---

## 3. Precondition Exclusivity: 120/150

### Distinct preconditions (50/60)

The critical overlap between Step 1 (error due to no valid transition) and E1 (concurrent edit conflict) is now well-separated:
- Step 1: "An item exists in a status that has no valid transition to the selected target status, **independent of concurrent modifications**"
- E1: "Another user has changed the item's status after the current user loaded the page; the current user's transition is now based on stale data"

The "independent of concurrent modifications" clause in Step 1 explicitly excludes the E1 scenario. These preconditions are now clearly distinguishable. Good.

Other pair analysis:
- Step 4 vs E3: Step 4 opens sub-item form (defaults checked). E3 opens main-item form (required fields missing). Different form types, different preconditions. Distinguishable.
- Step 5 vs E4: Step 5 has all required fields filled and succeeds. E4 has all required fields filled but backend rejects. Distinguishable by backend validation result.
- E6 vs E7: Both have member-role user. E6 targets status transition. E7 targets conversion form. Distinguishable by operation type.
- E8 vs E10: E8 is API surface with missing fields. E10 is web surface with empty assignee. Distinguishable by surface type and action.

Deductions:
- E8 and E10 both test validation failure for conversion forms. E8 tests API-level missing fields, E10 tests web-level empty assignee. While distinguishable by surface type, the underlying condition (required field missing) is semantically similar. For a test writer, these are clearly different tests, so the overlap is minor.
- Step 5 and Step 7 both test successful form submission for different conversion types. Precondition distinguishes them (sub-item form vs main-item form). No overlap.

### Sufficient preconditions (38/50)

All steps now have explicit precondition blocks. Improvements:
- Step 1's "independent of concurrent modifications" adds an exclusivity guard that makes the precondition more precise.
- E4's "All required fields are filled but the backend rejects the submission — a sub-item with the same name already exists under the target parent item" is specific and testable.
- E8's "A direct API request is sent for conversion form submission with missing assignee or priority fields" is specific.

Deductions:
- E10's precondition "a direct API request or browser developer tools bypass has cleared the assignee field value" — the "or" clause creates ambiguity. Which mechanism is the test supposed to use? The precondition should specify one clear state, not multiple possible mechanisms to achieve it. A test writer cannot uniquely determine the test setup.
- Step 7's precondition "A todo item exists that can be converted to a main item; the todo-to-main-item conversion form is open with all required fields filled" — this includes both the data state and the UI state. Acceptable but does not describe how the form was opened. This is a minor gap.

### Missing preconditions for error outcomes (32/40)

All error/boundary outcomes have explicit preconditions stating what triggers them:
- E1: concurrent modification by another user. Explicit.
- E2: item eligible for terminal transition. Explicit.
- E4: backend rejects due to duplicate name. Explicit with source annotation.
- E6/E7: member role without permission. Explicit.
- E8: missing required fields in API request. Explicit.
- E9: session expired. Explicit.
- E10: empty assignee field. Explicit.

Deductions:
- E8's precondition does not specify whether both fields are missing or just one. "missing assignee or priority fields" — is the test supposed to omit both or just one? The expected result says "listing the missing fields" (plural), implying both could be missing. Ambiguity in test setup.
- E10 specifies "empty assignee field" — only one field. Priority field empty scenario is not tested as a separate outcome.

---

## 4. Fact Alignment: 125/150

### Factual claims traceability (50/60)

Improvements over iteration 2:
- Step 1 now includes "with content provided by the backend" which traces to PRD's "消息内容来自后端." This was a gap in iteration 2 and is now fixed.
- E1's Chinese error message "该事项状态已被他人更新，请刷新后重试" matches the PRD Concurrency Requirements section verbatim.
- Step 4's "start date defaults to today" matches PRD's "开始时间默认当天."
- Step 4's "Description field is disabled (greyed out, not editable)" matches PRD's "描述字段为 disabled 灰色样式，不可编辑."

Deductions:
- Step 2 says "the page reflects the new status" and Step 3 says "the page reflects the new terminal status." The PRD does not describe what "reflects the new status" means in terms of UI changes. These are vague claims that cannot be verified against the PRD. They are not incorrect but lack specificity for traceability.
- Step 6's expected result "All fields in the newly opened form are empty — no residual data from the previously closed form" traces to PRD Story 4's acceptance criterion "所有字段为空（无残留数据）." Good traceability.

### Inferred claims with rule support (42/50)

Improvements over iteration 2:
- E4 now has `source: inferred` annotation: `<!-- source: inferred — derived from conversion form backend validation failure described in prd-spec -->`. The annotation cites the PRD spec's conversion form flow section ("若提交失败（后端校验不通过），字段内容保留，展示错误消息供用户修正后重试") as the reasoning basis. This addresses the iteration-2 gap where E4 was an unclassified inference.
- E6: `source: inferred — derived from API surface unauthorized mandatory outcome`. Correct.
- E7: `source: inferred — derived from API surface unauthorized mandatory outcome`. Correct.
- E8: `source: inferred — derived from API surface validation-error outcome and Web surface validation-error mandatory outcome`. Correct dual-source annotation.
- E9: `source: inferred — derived from Web surface session-expired mandatory outcome`. Correct.
- E10: `source: inferred — derived from Web surface validation-error mandatory outcome`. Correct.

All inferred steps now have `source: inferred` annotations. Significant improvement.

Deductions:
- E4's annotation says "derived from conversion form backend validation failure described in prd-spec." This cites a PRD section, not a `required_outcomes` rule. The rubric says inferred claims should cite "the `required_outcomes` rule from the surface type configuration that mandated their generation." E4 is not mandated by a required_outcomes rule — it is a domain-specific inference from the PRD. The annotation is defensible but does not follow the rubric's expected format strictly.
- The annotations are all in HTML comments (`<!-- ... -->`). As noted in iteration 2, this is a defensible format choice but less discoverable than inline annotations.

### No hallucinated claims (33/40)

No outright hallucinations detected. All claims are either:
- Directly traceable to PRD (Steps 1-7, E1, E2, E5)
- Annotated as inferred with reasoning basis (E3, E4, E6, E7, E8, E9, E10)

Deductions:
- E10's precondition "a direct API request or browser developer tools bypass has cleared the assignee field value" — this describes a test setup mechanism that is not grounded in any PRD requirement or surface rule. The PRD says client-side validation should prevent submission with empty fields (disabled submit button). The journey creates a scenario where this validation is bypassed, which is a reasonable test case but is presented as a factual precondition rather than an inference. The mechanism ("browser developer tools bypass") is not stated in the PRD. This is a minor unclassified inference.

---

## 5. Surface Fitness: 130/150

### Mandatory derived outcomes (55/60)

**Web mandatory outcomes:**
- `validation-error`: Present in E10. User submits form with empty assignee. Error near field. Form not submitted. Retry possible. **Satisfies requirement.**
- `session-expired`: Present in E9. Session expires during form fill. Redirect to login. Unsaved data not preserved. Re-auth possible. **Satisfies requirement.**

**API mandatory outcomes:**
- `unauthorized`: Present in E6 (status transition) AND E7 (conversion form). Both cover member-role users sending requests to authenticated endpoints. **Full coverage — major improvement over iteration 2.**

The iteration-2 gap (missing unauthorized for conversion form) is now fixed with E7.

Deductions:
- `unauthorized` in E7 says "The API returns an authorization error; no item is created." This does not specify the HTTP status code (401 vs 403). For an API surface journey, the distinction between authentication failure (401) and authorization failure (403) matters. A member-role user with valid credentials but insufficient permissions should get 403, not 401. The journey uses "authorization error" which could be either. This is a surface fitness gap for API-level specificity.
- `validation-error` at the API level is covered in E8 but only for missing fields. No format-level validation errors are tested. The API surface rule's additional outcome for `validation-error` says "Request with invalid body/parameters. Assert: 400 Bad Request, response body lists all validation failures." E8 covers this adequately for the field types in this domain.

### Test strategy proportions (40/50)

The journey now has better balance between web and API surfaces:
- Web: Steps 1-7, E1, E2, E3, E4, E5, E9, E10 (14 steps)
- API: E6, E7, E8 (3 steps)

The ratio is roughly 82% web / 18% API, which is below the balanced 50/50 target for a dual-surface journey. However, the journey is primarily about a user-facing workflow (status transitions and form interactions) where the web surface naturally dominates. The API steps cover the key mandatory outcomes.

Improvements:
- E7 adds conversion form unauthorized coverage for API. Good.
- E8 adds API-level validation error. Good.
- Each step now has explicit `<!-- surface: web -->` or `<!-- surface: api -->` tags. This addresses the iteration-2 feedback about missing surface annotations. Excellent improvement.

Deductions:
- API surface is still underrepresented (3 of 17 steps). For a true 50/50 balance, at least 2-3 more API-specific scenarios would be needed (e.g., API-level conflict response for concurrent status change, API-level success response verification with full response schema).
- No API happy-path steps. Steps 1-7 are all web. The API surface only has error/boundary steps.

### Surface-specific environment (35/40)

Web assumptions are realistic: button clicks, form filling, confirmation dialogs, error messages, login redirects. API assumptions are realistic: authorization errors, validation error responses.

Improvements:
- E9's session expiry handling is realistic for web: "user is redirected to the login page."
- E10's "browser developer tools bypass" is a realistic web testing scenario, though the precondition phrasing is awkward.

Deductions:
- E10's precondition "a direct API request or browser developer tools bypass has cleared the assignee field value" — for a web surface test, the "direct API request" mechanism is a cross-surface assumption. A web test should focus on web mechanisms (DOM manipulation, devtools). Mixing API calls into a web step's precondition is a surface boundary violation.
- E8 is tagged `surface: api` and describes "An API request submits a conversion form with empty required fields" — realistic for API testing. The expected result "The API returns a validation error response listing the missing fields; no item is created" is appropriate for API surface. Good.

---

## 6. Internal Consistency: 130/150

### Invariants hold in every step (50/60)

Four invariants declared:

1. "Error messages for status transitions are always displayed as persistent inline messages below the action area, never as auto-disappearing tooltips"
   - Step 1: "an error message is displayed below the action area" — holds.
   - E1: "An inline error message displays" — holds.
   - **Invariant holds across all status transition steps.**

2. "All conversion forms clear all fields on close/cancel or successful submission"
   - Step 5: "all form fields are cleared" — holds.
   - Step 6: "All fields in the newly opened form are empty — no residual data from the previously closed form" — holds.
   - Step 7: "all form fields are cleared" — holds.
   - E5: "All fields are empty — no residual data" — holds.
   - **Invariant holds.**

3. "Required fields (assignee, priority) are enforced at both UI level (disabled submit button) and API level (validation error response)"
   - Step 4: "submit button is disabled until both required fields are filled" — UI enforcement. Holds.
   - E8: "The API returns a validation error response listing the missing fields" — API enforcement. Holds.
   - E10: "error message is displayed near the assignee field" — UI-level validation. Holds.
   - **Invariant holds.**

4. "Description field in the todo-to-sub-item conversion form is always disabled and cannot be modified"
   - Step 4: "Description field is disabled (greyed out, not editable)" — holds.
   - No other step modifies the description in a todo-to-sub-item form.
   - **Invariant holds.**

Deductions:
- Invariant 1 says "persistent inline messages" but Step 1 no longer explicitly states persistence behavior (iteration 2 had "persists until the user dismisses it or performs another action" which was removed in iteration 3). The invariant claims persistence but no step explicitly verifies the message persists (vs. auto-disappearing). Step 1 just says "an error message is displayed" — this does not contradict the invariant, but it also does not actively confirm it. The invariant is asserted but not verified by any step's expected result.
- E4's expected result says "Form fields retain their values (not cleared); an error message is displayed indicating the specific validation failure; the user can correct and retry." This is for a conversion form validation failure, not a status transition. Invariant 1 only applies to status transition error messages. E4's error message behavior is not covered by any invariant, which is acceptable but represents a gap in invariant coverage.

### Cross-step references (40/50)

The journey uses a clear numbering scheme (Steps 1-7, E1-E10). No explicit cross-step references like "the item created in Step 2."

Implicit sequential dependencies:
- Step 5's precondition "The todo-to-sub-item conversion form is open with assignee and priority filled in" logically follows Step 4 (opening the form). Clear.
- E5's precondition "A conversion form was just submitted successfully" references Step 5 or Step 7. Clear.
- Step 6's precondition "A conversion form has been partially filled" could reference any step that opens a form. Clear.

Deductions:
- Step 7's precondition "the todo-to-main-item conversion form is open with all required fields filled" includes both the data state and the form state. No preceding step opens the todo-to-main-item form. Step E3 opens the todo-to-main-item form to check disabled submit button, but Step 7 is a happy path step that should follow from an opening action. The implicit assumption is that the user opened the form between steps, but this is not stated.
- E10's precondition references "a direct API request or browser developer tools bypass" which is not an action described in any previous step. This creates a disconnected test scenario that cannot be reached from the normal flow.

### Risk level consistency (40/40)

Risk level "High" is well-justified:
- Status transitions involve state mutation (Steps 1-3, E1-E2)
- Terminal transitions are potentially irreversible (Step 3, E2)
- Conversion forms create new items — irreversible creation (Steps 5, 7)
- Concurrent edit conflicts affect data integrity (E1)
- Authorization failures are security-sensitive (E6, E7)
- Session expiry is security-adjacent (E9)

The risk level is consistent with the content. **No deduction.**

---

## Comparison with Iteration 2

### Improvements

1. **Step 1 backend message source fixed**: Now includes "with content provided by the backend" — addresses iteration-2 gap.
2. **E4 source annotation added**: Now has `source: inferred` annotation — addresses iteration-2 unclassified inference.
3. **E7 added for conversion form unauthorized**: Addresses iteration-2 gap where API surface only covered status transition unauthorized.
4. **E8 added for API-level validation error**: New step covering API surface validation-error for conversion form.
5. **E9 and E10 split from iteration-2's E7/E8**: Clearer separation of session-expired and validation-error as distinct steps.
6. **Surface tags per step**: Every step now has `<!-- surface: web -->` or `<!-- surface: api -->` — addresses iteration-2 feedback.
7. **Precondition exclusivity guard**: Step 1 now says "independent of concurrent modifications" to explicitly distinguish from E1.

### Remaining Gaps (for future improvement)

1. **API surface underrepresented**: 3 of 17 steps are API-only. Consider adding API happy-path steps or API-level conflict response verification.
2. **E10 precondition mechanism ambiguity**: "direct API request or browser developer tools bypass" creates unclear test setup and a surface boundary issue.
3. **Invariant 1 persistence not actively verified**: The invariant claims persistent messages but no step's expected result explicitly verifies persistence.
4. **No todo-to-main-item form defaults step**: Step 4 covers sub-item form defaults; main-item form defaults are untested.
5. **E6 mixes API and web assertions**: Includes frontend behavior in an API-surface step.

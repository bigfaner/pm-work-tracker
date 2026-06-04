---
iteration: 1
scorer: adversary
total_score: 680
pass: false
date: "2026-06-04"
---

# Journey Eval Report: task-status-transition (Iteration 1)

## Dimension Scores

| Dimension | Score | Threshold | Status |
|-----------|-------|-----------|--------|
| 1. Completeness | 145/200 | 120 | PASS |
| 2. Semantic Purity | 160/200 | 120 | PASS |
| 3. Precondition Exclusivity | 85/150 | 90 | FAIL |
| 4. Fact Alignment | 85/150 | 90 | FAIL |
| 5. Surface Fitness | 110/150 | 90 | PASS |
| 6. Internal Consistency | 95/150 | 90 | PASS |
| **Total** | **680/1000** | **850** | **FAIL** |

---

## 1. Completeness: 145/200

### Journey metadata (40/50)

- Name `task-status-transition` follows kebab-case. Correct.
- `risk_level: "High"` is present and justified by state mutation operations (status transitions, conversion forms). The risk classification comment in the document body explicitly maps to the rubric's High criteria (irreversible operations, state mutation). However, the `sources` field references PRD docs but the journey lacks a `fact_table` or any formal fact-tracing mechanism. Minor deduction for missing `fact_ids` traceability in metadata.

### Steps completeness (65/80)

All 5 happy-path steps and 6 edge-case steps have clear action descriptions and expected results. The sequence is coherent: error case -> success -> form open -> form submit -> form cleanup, then edge cases for each. Deductions:

- Step 1 ("Trigger status transition with error") has a weak expected result. It says "frontend displays an inline Alert component below the action area with the backend-provided error message (not a tooltip that disappears after 2 seconds)" -- this describes the error display mechanism but does not specify what the user observes about the item's status after the failed transition. Does the item remain unchanged? This is implied but not stated.
- Step 2 mentions "if transitioning to a terminal status, a confirmation dialog appears first" but conflates two sub-scenarios (terminal vs non-terminal) into one step without splitting them into separate outcomes with distinct preconditions.
- No step covers the sub-item conversion form (only todo-to-sub-item and todo-to-main-item are covered). The PRD mentions "sub-item conversion" in the journey overview but no step exercises it.

### Outcomes coverage (40/70)

Deductions:

- **Missing `validation-error` outcome for web**: While Step 3b covers disabled submit button (client-side validation), there is no scenario where a user submits invalid data that reaches the backend and receives a validation error response displayed on the form. Step 4b mentions "backend validation" but frames it as a conflict/duplicate, not as a field-level validation error. The web surface requires `validation-error` as a mandatory derived outcome.
- **Missing `session-expired` outcome for web**: No step covers what happens when a user's session expires during a workflow (e.g., while filling a conversion form). This is a mandatory web outcome.
- **`unauthorized` outcome is partially covered**: Step 6b covers unauthorized status transition but only for API surface (member-role user). It does not cover unauthorized access for conversion form endpoints or what happens on the web surface when an unauthorized user navigates to a restricted action.
- No outcome covers the network-error scenario (web surface additional common outcome).
- No outcome covers what happens when a conversion form is opened for a todo item that no longer exists (race condition between listing and opening).

---

## 2. Semantic Purity: 160/200

### Outcome descriptions (65/80)

Most outcomes use natural language. However:

- Step 1 references "inline Alert component" -- this is a framework-specific component name, not a user-visible description. A user does not see "an Alert component"; they see "an error message displayed below the action area." Component names are implementation details.
- Step 2 references "confirmation dialog" -- this is acceptable as it describes a UI pattern, not an implementation.
- Step 3 references "disabled (greyed out, not editable)" and "required markers" -- these are user-observable behaviors, acceptable.
- Step 6b references "API returns 403 Forbidden" -- this is HTTP status code implementation detail, not a user observation. For the API surface this is borderline acceptable, but for the web surface portion it leaks implementation.

### Preconditions (45/60)

Preconditions are generally declarative. However:

- Step 6b: "A member-role user tries to access the status transition endpoint" -- this is borderline procedural ("tries to access") rather than a state declaration ("the user has member role and the item exists").
- Step 4b: "PM user has filled all required fields but backend rejects the submission (e.g., duplicate name, business rule violation)" -- this is a good declarative precondition but the parenthetical examples are somewhat vague. "business rule violation" is catch-all and does not help distinguish this outcome from other error outcomes.

### Step descriptions (50/60)

Steps are mostly user-level. Deductions:

- Step 1: "Backend returns a specific error reason" in the Expected Result leaks backend implementation. Should be "an error message is displayed explaining why the transition cannot be performed."
- Step 6b: "Member user sends a status transition request via API" -- this is an API-level action description, not a user-level action. For a dual-surface journey this is acceptable but should be clearer about which surface this step belongs to.

---

## 3. Precondition Exclusivity: 85/150 (FAILS THRESHOLD)

### Distinct preconditions (35/60)

**Critical issue**: Steps 1 and 1b have overlapping preconditions.

- Step 1: "PM user clicks a status transition button for an item that cannot be transitioned due to backend business rules"
- Step 1b: "PM user clicks a status transition button" with precondition "Another user has changed the item's status between the time the current user loaded the page and attempted the transition"

Both steps involve clicking a status transition button and getting an error. The distinguishing factor is the *cause* of the error (business rules vs concurrent edit). However, from the user's perspective, the precondition for Step 1 could also be true during a concurrent edit -- the business rules could still reject the transition regardless. The preconditions are not truly mutually exclusive because "cannot be transitioned due to backend business rules" is a superset that could include "status was changed by another user." These should be structured as separate outcomes within the same step, with explicit mutually exclusive preconditions.

**Deduction: -20 for ambiguous pair (Steps 1 vs 1b)**.

### Sufficient preconditions (25/50)

- Step 2 has no explicit precondition stated (it is in the Happy Path section but has no "Precondition:" line). Given that Step 1 requires an item that cannot be transitioned and Step 2 requires an item that is eligible, these are implicitly exclusive. But implicit exclusivity is not sufficient -- the preconditions should be explicitly stated.
- Step 5 and Step 5b both involve opening forms, but their preconditions differ (one is "close a partially filled form" the other is "after successful submission"). These are distinguishable but neither Step 5 nor Step 2 has a formal precondition block.
- Steps 3b and 3 have the same action ("open conversion form") but for different form types (todo-to-main-item vs todo-to-sub-item). This is distinguishable but the precondition in Step 3b only says "A todo item exists that can be converted to a main item" -- it does not state that the user opens the *todo-to-main-item* form specifically vs the *todo-to-sub-item* form. The action description clarifies this, but the precondition alone does not.

### Missing preconditions for error outcomes (25/40)

- Step 4b's precondition is vague: "backend rejects the submission (e.g., duplicate name, business rule violation)." What specific input triggers this? A test writer cannot set up this scenario without knowing what data to provide. The precondition should specify something concrete, e.g., "a sub-item with the same name already exists under the target parent."
- Step 1b's precondition is well-specified (concurrent edit scenario). Good.
- Step 6b's precondition is adequate ("member-role user" is specific).

---

## 4. Fact Alignment: 85/150 (FAILS THRESHOLD)

### Factual claims traceability (35/60)

The journey references `docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 1, Story 4)` and `prd-spec.md` as sources, but:

- No individual outcome references a specific `fact_id` or even a PRD section number. The mapping is implicit.
- Verification against PRD sources:
  - Step 1 (status transition error with inline Alert): Aligns with Story 1 and prd-spec #1. **Verified.**
  - Step 3 (description field disabled, start time defaults to today): Aligns with Story 4 and prd-spec #4. **Verified.**
  - Step 5/5b (form cleanup on close): Aligns with Story 4 and prd-spec #6. **Verified.**
  - Step 6b (unauthorized/403): PRD mentions member permissions (#8) but does not explicitly state 403 for status transitions. The PRD says "delete button only visible for PM role" and "team selector only shows authorized teams." The 403 for status transitions is an **inferred** behavior not directly stated in the sources. It should be marked `source: inferred`.
  - Step 1b (concurrent edit / conflict error): prd-spec Concurrency Requirements section confirms "状态流转：基于乐观锁机制...后端返回冲突错误，前端展示'该事项状态已被他人更新，请刷新后重试'". **Verified.** However, the journey's exact error message "This item status has been updated by another user, please refresh and retry" does not match the PRD's Chinese message "该事项状态已被他人更新，请刷新后重试". This is a discrepancy that should be noted.

- **Unclassified claims**: The error message in Step 1b is stated as a specific string but is not traced to a fact. The PRD provides the Chinese version. The English translation in the journey is an inference.

### Inferred claims with rule support (25/50)

- `validation-error` (web mandatory): Step 3b partially covers client-side validation (disabled submit button), but no outcome covers a server-side validation error returned for the conversion form. This mandatory derived outcome is missing as a server-side scenario. No `source: inferred` annotation exists for any outcome.
- `session-expired` (web mandatory): Completely absent. No inferred claim exists for this mandatory outcome.
- `unauthorized` (API mandatory): Step 6b covers this for status transitions. However, conversion form endpoints also require authentication, and no unauthorized scenario exists for them. Partially covered.

**No outcome in this journey is annotated with `source: inferred`.** This means all derived boundary outcomes are unclassified in terms of their reasoning basis.

### Hallucinated claims (25/40)

- Step 1b's error message "This item status has been updated by another user, please refresh and retry" -- the PRD specifies this message in Chinese. The journey uses an English translation without noting the translation or citing the source. This is not a hallucination per se but an unclassified transformation of source material.
- Step 6b: "API returns 403 Forbidden; frontend does not show the transition button for unauthorized users." The 403 status code for status transitions is not explicitly stated in the PRD sources reviewed. The PRD mentions RBAC and permission codes but does not specify the exact HTTP response for unauthorized status transitions. This is an inference that should be marked as such.
- No outright hallucinations detected (all claims are plausible inferences from the source material), but the lack of classification is a systematic failure.

---

## 5. Surface Fitness: 110/150

### Mandatory derived outcomes (40/60)

This journey declares `surface_types: ["web", "api"]`.

**Web mandatory outcomes:**
- `validation-error`: Partially present. Step 3b covers client-side validation (disabled submit button) but this is not a true validation-error outcome where a form submission is rejected. Step 4b covers backend rejection but frames it as "duplicate name, business rule violation" rather than field-level validation. The web surface rule says validation-error should cover "required field left empty, email format invalid, numeric field has non-numeric input." Step 3b's disabled submit button is the closest match but it prevents submission rather than showing an error after submission. **Partial credit.**
- `session-expired`: **Completely absent.** No scenario covers what happens when a user's session expires during a status transition or form interaction. This is a mandatory web outcome. Score 0 for this sub-criterion.

**API mandatory outcomes:**
- `unauthorized`: Present in Step 6b for the status transition endpoint. However, conversion form endpoints also require authentication, and no unauthorized scenario exists for them. **Partial credit.**

### Test strategy proportions (40/50)

The journey covers both web user interactions and API-level scenarios. The mix is reasonable for a dual-surface journey. The web steps (1-5, 1b-5b) focus on user-facing interactions appropriate for E2E testing. Step 6b is an API-only step. The balance is acceptable.

However, the journey does not explicitly separate web steps from API steps or indicate which surface each step targets. For a dual-surface journey, this tagging is important for test script generation.

### Surface-specific environment (30/40)

Web assumptions are realistic: clicking buttons, filling forms, confirmation dialogs, inline error messages. API assumptions are realistic: HTTP status codes, endpoint access.

Deduction: The journey mixes web and API perspectives without clear separation. Step 6b switches from web user actions ("PM user clicks") to API-level actions ("Member user sends a status transition request via API"). This mixing could confuse test generation about which test framework to use for which step.

---

## 6. Internal Consistency: 95/150

### Invariants hold in every step (45/60)

The journey declares four invariants:

1. "Inline error messages (Alert component) are always used for status transition errors, never auto-disappearing tooltips" -- Steps 1 and 1b use inline error messages. Step 2 (successful transition) has no error. Step 4b uses "an error message is displayed." **Invariant holds.**
2. "All conversion forms clear all fields on close/cancel or successful submission" -- Steps 5 and 5b verify this. Step 4 (successful submission) says "all form fields are cleared." **Invariant holds.**
3. "Required fields (assignee, priority) are enforced at both UI level (disabled submit) and API level (validation error response)" -- Step 3 enforces at UI level. However, no step explicitly verifies the API-level enforcement for the conversion form. Step 4b mentions backend rejection but for "duplicate name, business rule violation" not specifically for missing assignee/priority. **Invariant partially unverified.**
4. "Description field in todo-to-sub-item conversion form is always disabled and cannot be modified" -- Step 3 verifies this for the todo-to-sub-item form. No other step modifies the description field. **Invariant holds.**

Deduction: Invariant 3 claims API-level validation is enforced, but no outcome demonstrates this for missing assignee/priority specifically.

### Cross-step references (30/50)

- The journey uses flat numbering (Step 1-5 for happy path, Step 1b-6b for edge cases). There is no cross-reference between them. Step 1b is labeled as an edge case of Step 1 (hence "1b"), which is a logical connection.
- However, Step 5b references "a conversion form was just submitted successfully" as a precondition. This implicitly references Step 4 (successful submission). The reference is understandable but not explicit (it does not say "the form submitted in Step 4").
- No dangling references detected.

### Risk level consistency (20/40)

Risk level is "High." The rubric says High should involve "security, data loss, or irreversible operations." The journey covers:
- Status transitions (state mutation -- reversible in some cases)
- Conversion forms (creates new items -- irreversible creation)
- Terminal status transitions (could be considered irreversible in practice)
- Concurrent edit conflicts (data integrity)

The risk level is justified. However, the journey also covers form field clearing and disabled submit buttons, which are low-risk operations. The mix of High-risk operations (concurrent edits, authorization failures) with trivial operations (form field defaults) within a single High-risk journey is inconsistent. These could arguably be split into separate journeys at different risk levels.

---

## Blindspots (Outside Rubric Dimensions)

1. **No todo-to-main-item conversion happy path**: Step 3 covers todo-to-sub-item form, Step 3b covers todo-to-main-item but only the validation case. There is no happy-path step for submitting a todo-to-main-item conversion. The PRD (Story 4) covers both conversion types equally.

2. **No sub-item conversion form**: The journey overview mentions "sub-item conversion" as a form type, but no step exercises it. The PRD references this form type in the conversion forms section.

3. **No step covers the todo-to-sub-item conversion form submission happy path specifically**: Step 4 is generic ("submit the conversion form") but Step 3 specifically opened the todo-to-sub-item form. Is Step 4 continuing from Step 3, or is it a new form? The relationship is unclear.

4. **Missing concurrent edit for conversion forms**: Step 1b covers concurrent edit for status transitions, but conversion form submissions could also conflict (e.g., two users converting the same todo item). This scenario is not covered.

5. **No loading/async state coverage**: For a web surface, no step mentions what the user sees while a transition or submission is in progress (loading indicators, disabled buttons during API calls).

6. **Step numbering inconsistency**: Happy path uses "Step 1-5", edge cases use "Step 1b-6b". Step 6b has no corresponding happy-path Step 6. This asymmetry suggests the unauthorized scenario was added as an afterthought.

7. **Mixed language error messages**: Step 1b uses English error message while the PRD source uses Chinese. The journey should either use the PRD's exact message or note that it is a translation.

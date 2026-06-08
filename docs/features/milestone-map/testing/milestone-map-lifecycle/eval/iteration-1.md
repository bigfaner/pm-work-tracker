---
iteration: 1
score: 810
verdict: fail
date: "2026-06-08"
---

# Eval Report: milestone-map-lifecycle Journey (Iteration 1)

**Total Score: 810 / 1000** | Threshold: 850

---

## Dimension Scores

| Dimension | Score | Threshold | Verdict |
|-----------|-------|-----------|---------|
| 1. Completeness | 165 / 200 | 120 | PASS |
| 2. Semantic Purity | 165 / 200 | 120 | PASS |
| 3. Precondition Exclusivity | 125 / 150 | 90 | PASS |
| 4. Fact Alignment | 120 / 150 | 90 | PASS |
| 5. Surface Fitness | 115 / 150 | 90 | PASS |
| 6. Internal Consistency | 120 / 150 | 90 | PASS |

---

## 1. Completeness (165 / 200)

### Journey metadata complete (45 / 50)

- Name `milestone-map-lifecycle` follows kebab-case. Present.
- Risk level is `High`, justified by state mutation and irreversible operations. Appropriate.
- `surface_types` lists both `web` and `api`, with corresponding `surface_keys: ["frontend", "backend"]`. Sources reference both prd-user-stories and prd-spec.
- Deduction: `generated` timestamp is present but no `version` field. Minor metadata gap. (-5)

### Steps complete with required fields (70 / 80)

- 10 Happy Path steps (1-8, 6d, 7) and 20 Edge Case steps (1b-1i, 2b-2d, 3b, 4b-4d, 6b, 6c, 8b-8d, E1-E4) all have step names, user actions, and expected results.
- Steps form a coherent ordered sequence covering MilestoneMap CRUD lifecycle.
- Deduction: Steps E2-E4 (API surface) are isolated "appendix" steps rather than being integrated into the main flow. A downstream agent cannot execute Step E2 ("unauthenticated request to any MilestoneMap endpoint") without knowing which specific endpoint to target first. The API steps lack the operational specificity of the web steps. (-10)

### Outcomes cover happy path plus required derived scenarios (50 / 70)

- Happy path covers create, edit, status transitions (forward and rollback), cancel cascade, delete.
- Validation edge cases: empty name, name length, missing owner, invalid date range, no-change save.
- Error paths: server error on create (1f), server error on status transition (3b), concurrent edit conflict (2c).
- API derived outcomes: unauthorized (E2), not-found (E3), validation-error (E4).
- Web derived outcomes: validation-error (1b-1e, 2d), session-expired (E1), loading-state (1g), concurrent-edit (2c).
- Deductions:
  - **Missing: RBAC enforcement at the API level.** Step 8c tests missing `milestone:delete` on the web surface only. No API step tests a 403 response for any endpoint (create without `milestone:create`, update without `milestone:update`, etc.). The PRD security requirements (prd-spec: "无权限操作返回 403") explicitly require this. (-10)
  - **Missing: API happy-path steps.** The journey declares both `web` and `api` surfaces but provides zero API happy-path steps. There is no API step for creating a milestone map (POST /api/milestone-maps with valid data and asserting 201), listing maps (GET with pagination), or updating via PUT/PATCH. The API surface exists only as error-case appendices. (-10)

---

## 2. Semantic Purity (165 / 200)

### Outcome descriptions use natural language, not code/regex (60 / 80)

- Outcomes are generally in natural language: "Status changes from 'planning' to 'reviewed', the badge updates visually, and no error occurs", "Form displays a validation error near the name field, the form is not submitted, and the dialog remains open".
- Deduction: Step E3 specifies the HTTP endpoint path `GET /api/milestone-maps/{non-existent-id}` with the placeholder `{non-existent-id}`. For the API surface this is appropriate, but it introduces a structural pattern (URL templates) not present in any other step. The inconsistency means downstream agents must handle two different description paradigms. (-5)
- Deduction: Step E4 specifies `POST /api/milestone-maps with an empty name field` -- this mixes natural language with an HTTP method in a way that is borderline procedural. A purer form: "An authenticated create request is submitted with an empty name field." (-5)
- Deduction: Step 1g asserts "The submit button shows a loading state and the form prevents further interaction until the request completes." The phrase "the form prevents further interaction" is a behavioral assertion which is good, but "submit button shows a loading state" is a UI implementation detail. (-10)

### Preconditions are declarative statements (55 / 60)

- Preconditions are declarative: "The milestone map is in 'executing' status and has milestones with associated MainItems", "The team has 0 milestone maps and the empty state page is shown", "PM does not have milestone:delete permission".
- Deduction: Step E2 precondition says "An API request is sent without valid credentials" -- this uses passive voice describing an action (sent), not a state. A declarative form: "No valid authentication credentials are provided." (-5)

### No implementation coupling in Step descriptions (50 / 60)

- Steps describe semantic user actions: "PM opens the create dialog from the list page", "PM navigates to the milestone map detail page, opens the edit function".
- The web-surface steps appropriately reference UI elements (dialogs, pages, buttons) without coupling to specific component implementations.
- Deduction: Step 6d precondition says "has milestones with associated MainItems" -- this references entity relationships (milestones having MainItems) which is domain knowledge, not implementation coupling. Acceptable for the domain. No deduction.
- Deduction: Steps reference display labels like "'Reviewed' status option", "'Ready for Implementation' status option", "'In Progress' status option". These are UI display strings that may change. A semantic alternative would reference the status enum value (reviewed, ready, executing) rather than the display label. (-5)
- Deduction: Step 1 mentions "name (1-100 chars)" in the user action. The character constraint is a business rule, not an implementation detail, but embedding the exact range in the action description couples it to the constraint value. (-5)

---

## 3. Precondition Exclusivity (125 / 150)

### Preconditions are distinct across Outcomes within each Step (50 / 60)

- Happy Path steps each have a single outcome, so distinctness is trivially satisfied.
- Step 6 (complete with all milestones terminal) vs Step 6b (complete with non-terminal milestones) have distinct preconditions.
- Step 8 (delete in planning) vs Step 8b (delete in executing/completed) have distinct preconditions.
- Deduction: Step 8b lists "executing" and "completed" as non-deletable statuses in a single outcome. However, `cancelled` is also non-deletable (BIZ-milestone-004: only planning/reviewed/ready are deletable). The precondition does not mention `cancelled`, leaving it ambiguous whether the behavior for `cancelled` is the same. (-5)
- Deduction: Step 1f (server error during create) and Step 1g (loading state during create) share the same precondition "Create dialog is open with valid data entered" (for 1f: "the backend is unavailable"). These test different behaviors but their preconditions partially overlap -- both assume valid data and a submission in flight. The distinction is the backend state (available vs unavailable), which is stated in 1f but implicit in 1g. (-5)

### Preconditions sufficient to uniquely select an Outcome (40 / 50)

- Most steps have a single outcome, making uniqueness trivial.
- Deduction: Step 1e (invalid date range) has precondition "Create dialog is open with both plan start and plan end dates available". But the outcome describes the plan end date being earlier than plan start. What if only one date is filled? The precondition does not fully specify the state needed to trigger this outcome. (-5)
- Deduction: Step 6b preconditions state "has at least one milestone that is not in a terminal state" -- but this is a runtime condition that depends on prior steps creating milestones and adding items. The journey has no steps that create milestones or bind items, so the precondition is unreachable within the journey's own steps. This makes the outcome selection ambiguous. (-5)

### No missing Preconditions for error/boundary Outcomes (35 / 40)

- Most error outcomes state their trigger clearly.
- Deduction: Step 2c (concurrent modification conflict) says "Another PM has edited the same milestone map while the edit dialog is open." This is a valid precondition, but it assumes a conflict detection mechanism (optimistic locking, ETag, version field). If the backend does not implement this, the precondition describes an impossible state. The precondition is declaratively valid but may be factually unsound. (-5)

---

## 4. Fact Alignment (120 / 150)

### Factual claims traceable to fact_id or marked UNKNOWN (50 / 60)

- Factual claims use HTML comments for traceability: `<!-- fact: prd-spec -- initial status is planning -->`, `<!-- fact: prd-spec MilestoneMap state machine -- planning -> reviewed -->`, `<!-- fact: prd-spec Story 3 -- soft delete with cascade -->`.
- This is a reasonable inline annotation system, though it does not use formal `fact_id` references.
- Deduction: Step 6d's fact annotation says `<!-- fact: prd-spec -- cancelled is terminal, cascade behaviour -->` which is vague. The cascade behavior should cite a specific business rule (BIZ-milestone-006). The annotation is a summary, not a traceable reference. (-5)
- Deduction: Step 4b (filter by status), Step 4c (filter by owner), Step 4d (search by name) all have `<!-- fact: prd-spec -- list page supports status/owner/name filter -->` annotations. These are brief but adequate. No deduction.

### Inferred claims have required_outcomes rule support and source: inferred (45 / 50)

- Inferred outcomes are annotated: `<!-- source: inferred -- derived from Web surface validation-error mandatory outcome -->` (Steps 1b, 1c, 1d, 1e, 2d), `<!-- source: inferred -- derived from Web surface session-expired mandatory outcome -->` (Step E1), `<!-- source: inferred -- derived from API surface unauthorized mandatory outcome -->` (Step E2).
- This satisfies the annotation requirement.
- Deduction: Step 1g (loading state) has no `source: inferred` annotation, nor does it reference a surface `required_outcomes` rule. The loading-state boundary outcome is listed in the Web surface's "additional boundary Outcomes". The annotation should note this derivation. (-5)

### No hallucinated claims without classification (25 / 40)

- Most claims are well-grounded in PRD source material.
- Deduction: Step 2c (concurrent modification conflict) tests a scenario where "A conflict notification appears. No silent overwrite occurs." The PRD user story 2 explicitly states this acceptance criterion: "两个 PM 同时编辑同一里程碑图...收到冲突提示'数据已被其他人修改，请刷新后重试'，不会静默覆盖". So the claim IS in the PRD. However, the injected context's business rules make no mention of conflict detection, versioning, or optimistic locking for MilestoneMap updates. The factual status of this claim is ambiguous -- it is in the PRD but may not be implemented. The journey does not classify this ambiguity. (-10)
- Deduction: Step 1f (server error) has precondition "the backend is unavailable" and expected result "Page displays a retryable error message. The dialog stays open with all entered data preserved. No duplicate submission occurs." The "No duplicate submission occurs" assertion is not traceable to any PRD acceptance criterion. It is a reasonable assumption but unclassified as inferred. (-5)

---

## 5. Surface Fitness (115 / 150)

### Mandatory derived Outcomes from surface required_outcomes are present (45 / 60)

**Web surface**:
- `validation-error`: Present in Steps 1b, 1c, 1d, 1e, 2d. Well covered.
- `session-expired`: Present in Step E1. Covered.
- Web mandatory outcomes: COMPLETE. 30/30.

**API surface**:
- `unauthorized`: Present in Step E2 ("The API returns an authentication error"). Covered.
- `validation-error`: Present in Step E4 ("The API returns a validation error response"). Covered.
- API mandatory outcomes: COMPLETE. 15/15.

- Deduction: Step E2 tests unauthorized access generically ("An unauthenticated request is sent to any MilestoneMap endpoint") rather than testing specific endpoints. For comprehensive API coverage, each distinct endpoint should have its own unauthorized test. The current single step provides weak coverage. (-0 -- this is covered under test strategy proportions below, not double-counted)
- Deduction: No API step tests a `403 Forbidden` response (authenticated but insufficient permissions). The PRD security requirements state "无权限操作返回 403". This is an API-level boundary outcome not covered. (-15)

### Test strategy proportions match surface guidance (35 / 50)

- The journey has approximately 30 steps total. Of these, 3 (E2, E3, E4) are API surface steps and the remaining 27 are web surface steps.
- For a dual-surface journey (web + api), the 50/50 balanced strategy requires substantially more API coverage.
- The API steps are exclusively error-case appendices (unauthorized, not-found, validation-error). There are zero API happy-path steps: no POST to create, no GET to list, no PUT/PATCH to update, no DELETE to remove. An API journey should mirror the CRUD lifecycle at the HTTP level.
- Deduction: API surface is ~10% of the journey content. The 50/50 balance is severely off. (-15)

### Surface-specific environment and execution assumptions are realistic (35 / 40)

- Web surface assumptions are realistic: dialog interactions, form submissions, loading states, list filtering, page navigation.
- API surface assumptions are realistic: HTTP methods, status codes, endpoint paths.
- Deduction: Step E2's execution assumption "An unauthenticated request is sent to any MilestoneMap endpoint" is too vague for a downstream agent. Which endpoint? What HTTP method? What request body? The agent would need to guess. A concrete example (e.g., "GET /api/milestone-maps without Authorization header") would be more actionable. (-5)

---

## 6. Internal Consistency (120 / 150)

### Invariants hold in every Step (50 / 60)

- Invariant #1 (delete only in planning/reviewed/ready): Consistent with Steps 8 (planning delete), 8b (executing/completed non-deletable), 8c (no permission). Note: `cancelled` is also non-deletable but is not mentioned in Step 8b or Invariant #1.
- Invariant #2 (state machine transitions): Consistent with Steps 3-6, 6d, 7.
- Invariant #3 (cascade soft delete): Consistent with Step 8.
- Invariant #4 (completed requires terminal milestones): Consistent with Steps 6, 6b.
- Invariant #5 (any non-terminal to cancelled): Consistent with Step 6d.
- Invariant #6 (RBAC): Consistent with Setup and Step 8c.
- Deduction: Invariant #1 states "A milestone map can only be deleted when it is in 'planning', 'reviewed', or 'ready' status" but does not mention `cancelled` as non-deletable. Step 8b lists "executing" and "completed" as non-deletable but omits `cancelled`. Both should explicitly list `cancelled` as non-deletable for completeness. (-5)
- Deduction: Invariant #5 states "Any non-terminal state can transition to 'cancelled'" but Step 6d is the only step that tests this transition (from `executing` only). The invariant claims 4 non-terminal states (planning, reviewed, ready, executing) can cancel, but only 1 is tested. The invariant overpromises relative to step coverage. (-5)

### Cross-Step references are consistent (35 / 50)

- Steps 1-8 form a sequential flow that is internally consistent.
- Step 6d references "has milestones with associated MainItems" but no prior step creates milestones. The Setup mentions "At least one MainItem exists in the team (for milestone binding in later steps)" but does not mention milestones being created or items being bound to them. The cross-step reference to milestones is dangling.
- Step 6b references "at least one milestone that is not in a terminal state" which is unreachable without prior milestone creation steps.
- Step 8 references "all its milestones are soft-deleted" but no milestones exist in the journey flow.
- Deduction: Three steps (6, 6b, 6d) reference milestones that are never created in any step. The Setup section mentions MainItems but not milestones. This is a structural consistency gap. (-15)

### Risk level consistent with Journey content (35 / 40)

- Risk level is `High` and the journey involves state mutations, cascading deletes, status machine transitions, and irreversible operations (soft delete). This is well-justified.
- Deduction: The risk justification comment mentions "state mutation, data loss risk, or irreversible operations" but the most dangerous operation -- cancellation cascade (BIZ-milestone-006: cascades to cancel all milestones and unbind all MainItems within a transaction) -- is tested by only one step (6d) from only one starting state. The risk level is appropriate but the coverage of the highest-risk operation is thin. (-5)

---

## Pre-Score Anchors (Phase 1 Independent Observations)

1. **Cancelled cascade is undertested.** Step 6d tests cancellation from `executing` state only. BIZ-milestone-006 specifies cascade from ANY non-terminal state. The journey does not verify that the cascade behavior is the same regardless of starting state. If the backend has a bug where cancellation from `planning` does not cascade, this journey will not catch it.

2. **Milestone entity gap remains.** The journey covers MilestoneMap lifecycle but references milestones in Steps 6, 6b, 6d, and 8 without ever creating them. The Setup mentions MainItems but not milestones. A downstream agent cannot set up the required state for Steps 6, 6b, 6d without additional context about how to create milestones and bind items.

3. **API surface is structurally deficient.** The journey declares dual surfaces but provides API coverage only as error-case appendices. There are no API happy-path steps (create via POST, list via GET, update via PUT, delete via DELETE, status transition via PATCH). This means API contract tests cannot be generated from this journey.

4. **Step numbering inconsistency.** The journey has Steps 1-8 (happy path), then 6d (inserted between 6 and 7), then Step 7 (rollback). Step 6d is placed before Step 7 but logically could be placed after Step 8 (as a separate cancellation scenario). The numbering suggests cancellation is part of the main flow (between completing and rolling back), but Step 7 rolls back from `reviewed` to `planning`, which means the main flow has already progressed past `reviewed` by Step 6d. This ordering creates a logical inconsistency: the main flow reaches `executing` (Step 6), then Step 6d cancels from `executing`, but Step 7 rolls back from `reviewed` (which was 3 steps earlier).

---

## Phase 3: Blindspot Hunt

### [blindspot-1] Step ordering creates an unreachable main-path flow

The Happy Path order is: Step 1 (create) -> Step 2 (edit) -> Step 3 (planning->reviewed) -> Step 4 (reviewed->ready) -> Step 5 (ready->executing) -> Step 6 (executing->completed, terminal). But then Step 6d transitions to cancelled from executing -- but Step 6 already moved the map to `completed`. And Step 7 rolls back from `reviewed` to `planning` -- but the map is now either `completed` (Step 6) or `cancelled` (Step 6d), both terminal states from which no rollback is possible. The step ordering implies these are independent scenarios, not a sequential flow, but the document structure (Happy Path section with numbered steps) strongly suggests sequential execution. A downstream agent following steps in order will hit a contradiction after Step 6.

### [blindspot-2] No step tests the `reviewed` or `ready` delete paths

Step 8 tests deletion in `planning` status. Step 8b tests that `executing` and `completed` are non-deletable. But `reviewed` and `ready` are also deletable per BIZ-milestone-004, and no step tests this. The journey will not generate tests verifying that `reviewed` and `ready` maps CAN be deleted. This is a gap between the invariant (which correctly states planning/reviewed/ready are deletable) and the test coverage (which only tests planning).

### [blindspot-3] No step tests the `cancelled` status as non-deletable

BIZ-milestone-004 states that only planning/reviewed/ready are deletable. `cancelled` is not in this list. Step 8b lists "executing" and "completed" as non-deletable but omits `cancelled`. If a developer allows deletion of cancelled maps (a reasonable mistake since cancelled is a terminal state like completed), this journey will not catch it.

### [blindspot-4] Step E1 session-expired test does not specify which form

Step E1 says "PM submits a create or edit form" but does not specify which form. The expected result ("redirected to the login page; no data is modified") is generic. A downstream agent needs to know which specific form to test. Is it the create dialog (Step 1), the edit dialog (Step 2), or both? The vague "create or edit form" forces the agent to guess, and may result in testing only one form or an inconsistent choice.

### [blindspot-5] Missing navigation guard for unsaved changes

The Web surface lists `navigation-guard` as an additional boundary outcome. No step tests what happens when a user navigates away from an open create or edit dialog with unsaved changes. The PRD does not explicitly cover this, but the create dialog (Step 1) and edit dialog (Step 2) both involve form state that could be lost. The journey does not address this web-surface-relevant scenario.

### [blindspot-6] Filter/search steps (4b-4d) have no counterpart in the Happy Path

Steps 4b, 4c, 4d test filtering and searching on the list page, but these are placed in the "Edge Cases" section. Filtering and searching are core list-page behaviors, not edge cases. More importantly, there is no Happy Path step that exercises the list page at all (Step 1 creates from the list page but does not verify list display). The journey jumps from create (Step 1) directly to edit (Step 2) on the detail page, skipping any list-page interaction. A downstream agent testing the happy path will not verify that the created milestone map actually appears in the list with correct data.

---

## Summary of Required Fixes

1. **Reorder steps for logical consistency**: Steps 6d and 7 cannot follow Step 6 in a sequential flow because Step 6 terminates the map. Either: (a) restructure as independent scenarios rather than a sequential flow, or (b) move 6d and 7 into a separate "Branch Scenarios" section with explicit "starting from state X" preconditions.

2. **Add API happy-path steps**: Add steps for POST /api/milestone-maps (create), GET /api/milestone-maps (list), PUT /api/milestone-maps/:id (update), PATCH /api/milestone-maps/:id/status (transition), DELETE /api/milestone-maps/:id (delete). Include proper request/response assertions.

3. **Add milestone setup**: Either add a Setup item for "at least one milestone exists under the map with at least one MainItem bound" or add steps that create milestones and bind items before Step 6.

4. **Add reviewed/ready delete tests**: Add steps verifying that maps in `reviewed` and `ready` status can be deleted (positive tests), not just that `executing`/`completed` cannot (negative test).

5. **Add cancelled non-deletable test**: Explicitly test that `cancelled` maps cannot be deleted in Step 8b or a new step.

6. **Add API 403 test**: Add a step testing that authenticated requests without the required permission return 403 Forbidden.

7. **Clarify Step E1**: Specify whether session-expired applies to the create form, edit form, or both, and provide a concrete scenario rather than "create or edit form".

8. **Add Step 1g source annotation**: Annotate the loading-state outcome with its surface derivation.

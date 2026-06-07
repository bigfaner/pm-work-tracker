---
iteration: 1
score: 620
verdict: fail
date: "2026-06-08"
---

# Eval Report: milestone-map-lifecycle Journey (Iteration 1)

**Total Score: 620 / 1000** | Threshold: 850

---

## Dimension Scores

| Dimension | Score | Threshold | Verdict |
|-----------|-------|-----------|---------|
| 1. Completeness | 135 / 200 | 120 | PASS |
| 2. Semantic Purity | 145 / 200 | 120 | PASS |
| 3. Precondition Exclusivity | 85 / 150 | 90 | FAIL |
| 4. Fact Alignment | 65 / 150 | 90 | FAIL |
| 5. Surface Fitness | 100 / 150 | 90 | PASS |
| 6. Internal Consistency | 90 / 150 | 90 | PASS |

---

## 1. Completeness (135 / 200)

### Journey metadata complete (35 / 50)

- Name `milestone-map-lifecycle` follows kebab-case. Present.
- Risk level is `High`, justified by state mutation and irreversible operations (delete). Appropriate.
- Deduction: The `surface_types` field lists both `web` and `api` but the journey steps are written entirely from a web/UI perspective. There is no indication of which steps apply to which surface, nor are there any API-level steps. The dual surface claim is structurally present but substantively empty for the API surface. (-15)

### Steps complete with required fields (65 / 80)

- All 8 Happy Path steps and 16 Edge Case steps have clear user actions and expected results.
- Steps form a coherent ordered sequence for MilestoneMap CRUD.
- Deduction: Steps are exclusively web-surface (UI clicks, dialogs, badges). For a journey declaring both `web` and `api` surface types, the absence of any API-level steps (HTTP method, status code, request/response) is a completeness gap. The API surface exists only as a metadata tag. (-15)

### Outcomes cover happy path plus required derived scenarios (35 / 70)

- Happy path for MilestoneMap CRUD is covered: create, edit, status transitions, delete.
- Validation edge cases are present: empty name, name too long, missing owner, invalid date range.
- Error paths partially covered: server error on create, concurrent edit conflict, server error on status transition.
- Deductions:
  - **Missing: `cancelled` status path.** The state machine has 6 states including `cancelled` (a terminal state). No step tests transition to `cancelled` or verifies cancelled-state behavior. The PRD explicitly lists `cancelled` transitions from every non-terminal state. This is a major gap. (-20)
  - **Missing: API `unauthorized` outcome.** For the API surface, every authenticated endpoint should have an `unauthorized` derived outcome. None are present. (-10)
  - **Missing: `not-found` outcome for API surface.** No step tests accessing a non-existent milestone map via API. (-5)

---

## 2. Semantic Purity (145 / 200)

### Outcome descriptions use natural language, not code/regex (70 / 80)

- Outcomes are generally in natural language: "Status changes from 'planning' to 'reviewed'", "Form displays 'Name cannot be empty' error near the name field".
- Deduction: Step 4d references "debounce 300ms" which is an implementation detail of the search input, not a user-observable behavior. A user cannot observe debounce timing. (-5)
- Deduction: Step 1g asserts "The create button shows a loading state, all input fields are disabled" -- "all input fields are disabled" is a UI implementation assertion. A purer description would be "the form prevents further interaction". (-5)

### Preconditions are declarative statements (50 / 60)

- Preconditions are generally declarative: "Create dialog is open", "Another PM has edited the same milestone map while the edit dialog is open".
- Deduction: Step 3b precondition "PM clicks the status Badge and selects a new status" is procedural, not declarative. It should describe the state ("A status transition request has been sent to the backend"). (-10)

### No implementation coupling in Step descriptions (25 / 60)

- Steps heavily reference UI-specific implementation: "clicks the status Badge", "clicks the edit button in the title area", "the delete button on the detail page", "clicks the '+ Create Milestone Map' button on the list page".
- While these are user-level actions for the web surface, they couple to a specific UI layout (Badge, title area, list page) rather than describing the semantic action ("changes milestone map status", "edits milestone map information").
- Deduction: Steps are tightly coupled to a specific UI implementation. For a journey declaring dual surfaces (web + API), the coupling to UI widgets makes the steps non-portable. A semantic description like "PM transitions the milestone map from planning to reviewed" would be surface-agnostic. (-25)
- Deduction: Steps reference specific UI labels ("Create Milestone Map", "Ready for Implementation", "In Progress") that are display names, not semantic identifiers. (-10)

---

## 3. Precondition Exclusivity (85 / 150)

### Preconditions are distinct across Outcomes within each Step (35 / 60)

- For the Happy Path steps, each step has a single outcome (no branching), so distinctness is trivially satisfied within those steps.
- Deduction: Step 8b ("Delete non-planning milestone map") lists "reviewed", "executing", "completed" as example statuses, but these are not separate outcomes with distinct preconditions -- they are collapsed into a single outcome. The behavior might differ: for "reviewed" the delete button is hidden, but what about "cancelled"? The precondition lumps multiple states without distinguishing whether the behavior differs. (-10)
- Deduction: Step 1b/1c/1d/1e each test a single validation failure, but no step tests *multiple simultaneous* validation failures (e.g., empty name AND empty owner). The preconditions are distinct but incomplete -- they only cover one-at-a-time failures. (-15)

### Preconditions sufficient to uniquely select an Outcome (25 / 50)

- Most steps have a single happy-path outcome, making uniqueness trivial.
- Deduction: Step 1e (invalid date range) has precondition "Create dialog is open with both plan start and plan end dates available". But the outcome describes the plan end date being earlier than plan start. What if only one date is filled? The precondition does not fully specify the state needed to trigger this outcome. (-10)
- Deduction: Step 6b preconditions state "has at least one milestone that is not in a terminal state" -- but this is a runtime condition that depends on prior steps creating milestones and adding items. The journey has no steps that create milestones or bind items, so the precondition is unreachable within the journey's own steps. This makes the outcome selection ambiguous. (-15)

### No missing Preconditions for error/boundary Outcomes (25 / 40)

- Most error outcomes state their trigger: "PM does not have milestone:delete permission", "Another PM has edited the same milestone map".
- Deduction: Step 1f (server error) has precondition "Create dialog is open with valid data entered" but does not specify what makes the server return 500. The precondition describes the UI state, not the condition that triggers the error path (e.g., "backend service is unavailable"). (-5)
- Deduction: Step 3b (status transition server error) precondition is "PM clicks the status Badge and selects a new status" -- this is the action, not the precondition that triggers the error. What causes the backend error? Network failure? Validation failure? The precondition is missing. (-10)

---

## 4. Fact Alignment (65 / 150)

### Factual claims traceable to fact_id or marked UNKNOWN (20 / 60)

- The journey contains no fact_id references or UNKNOWN markings at all. This is a structural gap.
- Specific factual errors:
  - **Step 8b claims "Only milestone maps in 'planning' status can be deleted (BR-4)"** but the API handbook (Delete MilestoneMap error response) says: "only planning/reviewed/ready deletable". The tech design BR-4 section explicitly states: "仅 planning（规划中）、reviewed（已评审）、ready（待实施）状态的里程碑图允许删除". This is a **direct factual contradiction**. The journey understates the allowed delete statuses. (-30)
  - **Journey Invariant #1 repeats the same error**: "A milestone map can only be deleted when it is in 'planning' status". This compounds the factual error.
  - **Step 6b references "(BR-2)"** but the injected context uses `BIZ-milestone-002` naming. The BR numbering does not align with any standard reference. No source document is cited for the BR-2 claim. (-10)

### Inferred claims have required_outcomes rule support and source: inferred (15 / 50)

- Validation error outcomes (Steps 1b-1e, 2d) are reasonable inferences from the web surface's `required_outcomes: validation-error` rule.
- Deduction: No inferred outcomes are annotated with `source: inferred` or cite the `required_outcomes` rule that triggered their generation. This is a structural requirement per the rubric. (-20)
- Deduction: The `session-expired` mandatory outcome for web surface is completely absent. No step tests what happens when a user's session expires during form submission or page navigation. (-15)

### No hallucinated claims without classification (30 / 40)

- Most claims are reasonable and align with PRD content.
- Deduction: Step 2c (concurrent modification conflict) describes "Data has been modified by someone else, please refresh and retry" but the tech design does not mention optimistic locking or conflict detection for MilestoneMap updates. The API handbook Update response has no 409 status code. This claim appears fabricated -- the backend has no mechanism for this. (-10)

---

## 5. Surface Fitness (100 / 150)

### Mandatory derived Outcomes from surface required_outcomes (35 / 60)

**Web surface**:
- `validation-error`: Present in Steps 1b, 1c, 1d, 1e, 2d. Well covered. (+)
- `session-expired`: Completely absent. No step tests session expiration. This is a mandatory outcome for web surface. (-15)

**API surface**:
- `unauthorized`: Completely absent. No step tests unauthenticated access to any endpoint. This is mandatory for API surface. (-10)

### Test strategy proportions match surface guidance (35 / 50)

- The journey is written exclusively from a web/UI perspective (browser interactions, dialogs, buttons).
- For a dual-surface journey (web + API), the API surface is entirely missing. There are no Contract-level steps testing individual API endpoints (HTTP methods, status codes, response schemas).
- The 50/50 balanced strategy for both web and API is not achieved. The document is ~100% web journey, 0% API. (-15)

### Surface-specific environment and execution assumptions are realistic (30 / 40)

- Web surface assumptions are realistic: dialog interactions, form submissions, loading states, list filtering.
- No unrealistic assumptions detected for the web surface.
- Deduction: No API execution assumptions exist at all (no HTTP method references, no status code assertions, no request/response examples). (-10)

---

## 6. Internal Consistency (90 / 150)

### Invariants hold in every Step (40 / 60)

- Invariant #1 ("only deleted in planning") is incorrect per the source material, but within the journey's own logic it is consistently applied: Steps 8, 8b, 8c all reference planning-only deletion. However, Step 7 rolls back to planning, and Step 8 deletes in planning -- the flow is self-consistent even though the underlying fact is wrong.
- Deduction: Invariant #2 claims "rollback allowed from any non-terminal state to its predecessor". But the PRD state machine shows rollback from `reviewed` to `planning`, `ready` to `reviewed`, `executing` to `ready` -- only to *immediate* predecessor, not "any" non-terminal state. The invariant overstates the rollback capability. (-10)
- Deduction: Invariant #2 also says "no transitions from terminal states (completed)" but ignores `cancelled` as a terminal state. The PRD defines both `completed` and `cancelled` as terminal. This is an inconsistency. (-10)

### Cross-Step references are consistent (30 / 50)

- Steps 1-8 form a sequential flow that is internally consistent.
- Deduction: Step 6b references "at least one milestone that is not in a terminal state" but no prior step creates a milestone. The journey is about MilestoneMap lifecycle, but testing the "completed" transition requires milestones which are never created in any step. The cross-step reference to "milestones under this map" is dangling -- those milestones don't exist in the journey. (-15)
- Deduction: Step 8 references "all its milestones are soft-deleted" but no step created any milestones. The expected result references entities that the journey never set up. (-5)

### Risk level consistent with Journey content (20 / 40)

- Risk level is `High` and the journey involves state mutations and deletion. This is reasonable.
- Deduction: The justification comment says "Workflow involves state mutation, data loss risk, or irreversible operations" but Step 2c tests concurrent modification which (as noted in Fact Alignment) is not actually implemented in the backend. The risk level is partially justified by a feature that doesn't exist. (-10)
- Deduction: The `High` risk classification is appropriate but the journey lacks steps testing the highest-risk operations: cancellation cascade (which per BR-6 cascades to cancel all milestones and unbind all MainItems) and the terminal state constraints. These are the most dangerous operations and they are missing. (-10)

---

## Pre-Score Anchors (Phase 1 Independent Observations)

These observations were formed before rubric application and informed scoring:

1. **Missing entity coverage**: The journey covers MilestoneMap but not Milestone. The feature is called "milestone-map" but includes both entities. A lifecycle journey that ignores the child entity leaves a major behavioral gap.

2. **Unreachable preconditions**: Step 6b and Step 8 reference milestones that are never created in any step. This makes the journey non-executable as written -- a downstream agent would not know how to set up the required state.

3. **Factual error on delete constraint**: Step 8b and Invariant #1 both claim only `planning` status is deletable. The API handbook and tech design confirm `planning`, `reviewed`, and `ready` are all deletable. This will cause a downstream test agent to write incorrect negative tests.

4. **Cancelled status completely absent**: The 6-state machine has `cancelled` as a terminal state with specific cascade behavior (BR-6). No step tests cancellation from any status, nor the cascade effects on milestones and MainItems.

---

## Phase 3: Blindspot Hunt

### [blindspot-1] No step tests the `cancelled` status cascade

The PRD and tech design specify that cancelling a MilestoneMap cascades: all non-terminal milestones are cancelled, and all MainItems are unbound (BR-6). This is the most complex business rule in the feature, involving multi-entity cascading deletes and unbinds within a transaction. The journey entirely omits this path, which means downstream test generation will miss the most dangerous operation.

### [blindspot-2] Journey is non-executable due to dangling entity references

Steps 6 and 8 reference "milestones under this map" but no step in the journey creates milestones or binds MainItems. A downstream agent executing this journey would need to either: (a) infer that milestones must be created in a setup step, or (b) fail at runtime when no milestones exist. The Setup section does not mention milestone creation. This is a structural flaw that makes the journey non-self-contained.

### [blindspot-3] Missing `cancelled` as a terminal state in invariants

Invariant #2 says "no transitions from terminal states (completed)" but `cancelled` is also terminal per the PRD. This omission will lead downstream test generation to miss that `cancelled` maps cannot be transitioned, edited, or have milestones created under them.

### [blindspot-4] No step tests the `reviewed` and `ready` delete paths

Since the journey incorrectly claims only `planning` is deletable, it misses testing that `reviewed` and `ready` maps can also be deleted. The API handbook explicitly lists these as valid delete statuses. The journey will generate tests that incorrectly assert deletion is blocked for `reviewed` maps.

### [blindspot-5] No step tests RBAC at the API level

The Setup mentions permissions (`milestone:create`, `milestone:update`, `milestone:delete`) but only Step 8c tests missing `milestone:delete` permission. No step tests: missing `milestone:create` during creation, missing `milestone:update` during edit/status change, or missing `milestone:read` during listing. For a dual-surface journey, API-level 403 responses should be tested.

### [blindspot-6] Concurrent modification conflict is unimplemented

Step 2c tests concurrent modification conflict, but neither the API handbook nor the tech design mention optimistic locking, version fields, or 409 conflict responses for MilestoneMap updates. This step will generate tests for a feature that does not exist, wasting implementation effort or producing false failures.

---

## Summary of Required Fixes

1. **Fix delete constraint**: Change Step 8b precondition to "The milestone map is in executing, completed, or cancelled status" and expected result to "The delete button is not displayed. Only milestone maps in planning, reviewed, or ready status can be deleted (BR-4)." Update Invariant #1 accordingly.

2. **Add cancelled status steps**: Add steps for transitioning to cancelled from each non-terminal state, and verify the cascade behavior (milestones cancelled, MainItems unbound).

3. **Add milestone setup**: Either add a Setup item for "at least one milestone exists under the map with at least one MainItem bound" or add steps that create milestones and bind items before Step 6.

4. **Add API surface steps**: For the declared API surface, add steps testing HTTP methods, status codes, and error responses. At minimum, add `unauthorized` and `not-found` outcomes.

5. **Add session-expired outcome**: For the web surface, add a step testing session expiration during form submission.

6. **Remove or mark Step 2c**: Either remove the concurrent modification conflict step (unimplemented feature) or mark it as UNKNOWN/future scope.

7. **Fix Invariant #2**: Correct to "rollback allowed from non-terminal states to their immediate predecessor only" and include `cancelled` as a terminal state.

8. **Add fact traceability**: Add `source: inferred` annotations to derived outcomes and reference the `required_outcomes` rules.

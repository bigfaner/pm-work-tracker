# Eval Report: milestone-lifecycle Journey (Iteration 1)

**Date**: 2026-06-08
**Score**: 827/1000
**Threshold**: 850
**Result**: FAIL

---

## Dimension Scores

| Dimension | Score | Min | Status |
|-----------|-------|-----|--------|
| 1. Completeness | 167/200 | 120 | PASS |
| 2. Semantic Purity | 175/200 | 120 | PASS |
| 3. Precondition Exclusivity | 122/150 | 90 | PASS |
| 4. Fact Alignment | 118/150 | 90 | PASS |
| 5. Surface Fitness | 110/150 | 90 | PASS |
| 6. Internal Consistency | 135/150 | 90 | PASS |
| **Total** | **827/1000** | **850** | **FAIL** |

---

## Dimension 1: Completeness -- 167/200

### Journey Metadata (45/50)

- feature, journey name, risk_level, surface_types, surface_keys, sources, generated date all present.
- Risk classification criteria comment included.
- **-5**: The risk classification comment is generic ("Workflow involves state mutation, data loss risk, or irreversible operations") rather than a narrative justification tied to this journey's specific content (e.g., "cancelled is terminal and cascades unbind MainItems irreversibly").

### Steps Complete (72/80)

- Happy path Steps 1-6 cover the core lifecycle: create, edit, three forward transitions (not_started->in_progress, in_progress->completed, completed->cancelled), and delete. Each step has User Action, Expected Result, surface annotation, and fact annotations.
- Edge cases 1b-1g, 2b-2d, 3b-3c, 4b-4c, 5b-5d, 6b-6e provide comprehensive boundary coverage including validation, loading, cancel, conflict, server error, terminal state, permission, and cascade scenarios.
- Cross-cutting Steps E1-E5 address session-expired, unauthorized, not-found, conflict, and terminal-parent guard.
- **-5**: Step 1 says "timeline refreshes showing the new node at the correct date position" -- "correct date position" is ambiguous for a downstream agent. Position is relative to what? Other milestones? The plan completion date?
- **-3**: Step E2 says "any Milestone endpoint" -- this is too vague for a downstream agent to generate specific test cases. Which endpoints? All 7?

### Outcomes Coverage (50/70)

- Happy path outcomes present for all 6 steps.
- Validation errors (1b, 1c, 1d), server errors (1e, 3c), loading state (1f), dialog cancel (1g, 2d, 6e), no-change edit (2b), concurrent conflict (2c), incomplete MI rejection (3b), cancel cascade from not_started (4b), cancel cascade from in_progress (4c), terminal state behavior (5b), binding rejection (5c), cancelled appearance (5d), delete cancelled (6b), delete hidden for in_progress/completed (6c), delete hidden without permission (6d).
- Session-expired (E1), unauthorized (E2), not-found (E3), conflict (E4), terminal parent (E5).
- **-10**: Missing `completed -> in_progress` reopen transition. The PRD state machine explicitly includes this (completed -> in_progress: "PM 重新开启"). The journey's invariants declare it but no step tests it. A downstream agent would produce tests that never verify this business-critical path.
- **-5**: Missing empty-milestone completion edge case. Step 4 always assumes associated MainItems exist. The PRD states completion = MI average, empty = 0. An empty milestone has zero non-terminal MIs, so the completed guard is vacuously satisfied, but this is never tested.
- **-5**: Missing `navigation-guard` outcome. The web surface lists navigation-guard as an additional outcome. When a PM edits a milestone (Step 2) and navigates away without saving, no step tests what happens.

---

## Dimension 2: Semantic Purity -- 175/200

### Natural Language Outcomes (75/80)

- All outcomes use natural language: "Form displays a validation error", "Status changes to cancelled", "Dialog closes".
- No regex patterns, CSS selectors, XPath, or assertion calls anywhere.
- **-5**: Step E5's expected result says "The API returns an error" -- for an API-surface step, this is borderline acceptable. However Step 1e says "the backend returns an error" which is fine as it describes the user-observable behavior (error message displayed). No significant deduction here, the -5 is for Step E2 saying "returns an authentication error" without specifying what the user/system observes beyond the return value.

### Declarative Preconditions (48/60)

- Preconditions like "Create dialog is open" (1b), "Milestone is in cancelled status" (5b), "Milestone is in not_started status and has associated MainItems" (4b) are clear declarative statements.
- **-7**: Step 4 embeds "All associated MainItems are in terminal states (completed/closed)" inside the User Action block rather than as a separate Precondition. This mixes state requirement with action description, making it harder for a downstream agent to identify the precondition independently.
- **-5**: Steps 1e and 3c have preconditions like "the backend is unavailable" which is not a declarable test state -- a test agent cannot set "backend is unavailable" as a precondition.

### No Implementation Coupling (52/60)

- Steps describe user-level actions (clicking buttons, selecting from dropdowns, filling forms). No database queries, function calls, or code-level references.
- **-4**: Step 6 says "The milestone is soft-deleted" -- "soft-deleted" is an implementation detail. The user-observable behavior is "the milestone disappears from the timeline and panel closes." Whether it is soft-deleted or hard-deleted is an implementation choice.
- **-4**: Step 5 says "All associated MainItems are auto-unbound in the same transaction" -- "in the same transaction" is an implementation/consistency guarantee, not user-observable behavior. The user sees "associated items list becomes empty."

---

## Dimension 3: Precondition Exclusivity -- 122/150

### Distinct Preconditions (50/60)

- Each step has a single outcome, so no ambiguity between competing outcomes within a step.
- Steps 4b and 4c test cancellation from different source states (not_started vs in_progress) with distinct preconditions.
- Steps 1b, 1c, 1d test different validation failures with distinct triggers.
- **-10**: Steps 4b and 4c have nearly identical expected results ("Same cascade behavior as Step 4b"). While the preconditions differ (source status), the outcomes are semantically identical. A downstream agent might collapse these into one test, losing coverage of the different source states.

### Sufficient to Uniquely Select Outcome (42/50)

- Each step's precondition + action uniquely determines one outcome.
- **-5**: Step 1f (loading state) has precondition "Create dialog is open with valid data" which overlaps with Step 1's precondition. The distinguishing factor is "request is in flight" -- this is a transient runtime state, not a precondition that can be independently set up.
- **-3**: Steps 2b (no-change edit) and 2d (cancel edit) have similar preconditions ("Edit dialog is open") with different actions. The distinction is clear but the preconditions could be more precise.

### Missing Preconditions for Error/Boundary (30/40)

- Most error/boundary outcomes state their triggers clearly with Precondition blocks.
- Steps E1-E5 have well-formed preconditions.
- **-10**: Step 3c has no explicit Precondition block -- the precondition "The backend is unavailable when a status transition is attempted" is stated as a Precondition but it describes a failure injection condition, not a settable state.

---

## Dimension 4: Fact Alignment -- 118/150

### Factual Claims Traceable (45/60)

- The journey uses `<!-- fact: ... -->` annotations extensively, tracing to prd-spec stories, state machine rules, and business rules. Examples: `<!-- fact: prd-spec Story 5 -->`, `<!-- fact: prd-spec Milestone state machine -->`, `<!-- fact: prd-spec Story 7 -- soft delete -->`.
- **-10**: Annotations use informal text references ("prd-spec Story 5") rather than formal fact_id identifiers (e.g., "BIZ-milestone-001"). A downstream verification agent cannot programmatically resolve these references.
- **-5**: Step E4 claims "milestone name uniqueness within map" with `<!-- fact: prd-spec -- milestone name uniqueness within map -->` but the PRD does not explicitly state name uniqueness for milestones within a map. The PRD's Story 5 acceptance criteria do not mention this constraint. This may be an inferred or assumed rule.

### Inferred Claims with Rule Support (38/50)

- Steps E1-E5 use `<!-- source: inferred -- derived from [surface] mandatory outcome -->` annotations correctly. Example: `<!-- source: inferred -- derived from API surface unauthorized mandatory outcome -->`.
- Steps 1b-1d use `<!-- source: inferred -- derived from Web surface validation-error mandatory outcome -->`.
- Steps 1e and 3c reference "server-error boundary outcome".
- **-7**: Steps 2c (concurrent modification conflict) and Step E4 (duplicate name) do not have source annotations indicating whether these are facts from the PRD or inferred from surface requirements.
- **-5**: The journey does not cite which specific required_outcomes rule from surface-web.md or surface-api.md triggered each derived outcome. The annotations reference the outcome type but not the rule document.

### No Hallucinated Claims (35/40)

- Most claims are grounded in the PRD. Error messages are described generically ("validation error", "conflict notification") without fabricating exact strings.
- **-5**: Step 4 says "Completion percentage reflects the associated MI average" with fact annotation `<!-- fact: prd-spec -- completed sets completion to MI average -->`. The PRD says completion = MI average for all states (calculated on GET), not specifically set on completed transition. The phrasing implies completion is set during transition rather than always calculated.

---

## Dimension 5: Surface Fitness -- 110/150

### Mandatory Derived Outcomes Present (50/60)

- **Web surface**: `validation-error` present (Steps 1b, 1c, 1d). `session-expired` present (Step E1). Both mandatory outcomes covered.
- **API surface**: `unauthorized` present (Step E2). `not-found` present (Step E3). `conflict` present (Step E4). Additional `validation-error` at API level absent. Mandatory outcome covered.
- **-5**: API `validation-error` (400) is not explicitly tested. Steps E2-E5 cover 401, 404, 409 but no step tests sending invalid request bodies to API endpoints (e.g., missing required fields, type mismatches).
- **-5**: Step E1 is the only session-expired scenario and covers form submission only. Session expiry during status transition, delete confirmation, or long idle on timeline is untested.

### Test Strategy Proportions Match (30/50)

- **Web**: Approximately 25 journey-style steps + 6 contract-level validation steps. Ratio roughly 80/20 journey/contract -- heavier on journey than the 50/50 target but acceptable for a lifecycle journey.
- **API**: 4 steps (E2-E5) covering contract-level API testing. Zero journey-style API steps. The API surface is exclusively contract-tested with no journey flow.
- **-20**: API coverage is thin -- 4 steps covering 5 scenarios for a full CRUD lifecycle. No API journey testing (e.g., create milestone via API, then update, then transition status, then delete).

### Realistic Surface Assumptions (30/40)

- Web assumptions are realistic: browser-based interaction, dialogs, panels, badge dropdowns, loading indicators, timeline rendering.
- API steps mention HTTP methods (GET, POST), endpoint paths (`/api/milestones/{id}`), status codes (404, 409).
- **-10**: No API environment setup assumptions. No mention of base URL, authentication headers, content-type, or request format. A downstream agent generating API tests would need to infer these.

---

## Dimension 6: Internal Consistency -- 135/150

### Invariants Hold (52/60)

- Invariant "Cancellation of a milestone...automatically unbinds all associated MainItems within the same transaction" -- consistent with Steps 4b, 4c, and 5.
- Invariant "A milestone can only be marked as completed when all its associated MainItems are in terminal states" -- consistent with Step 4 (action) and Step 3b (rejection).
- Invariant "Cancelled milestones cannot receive new MainItem bindings" -- consistent with Step 5c.
- Invariant "Delete is only available for not_started or cancelled" -- consistent with Steps 6, 6b, 6c.
- Invariant "cancelled is a terminal state" -- consistent with Step 5b.
- Invariant "Status machine: ... completed -> in_progress (reopen)" -- declared but no step tests this.
- **-8**: The invariant declares completed -> in_progress as a valid transition but no step exercises it. This is a gap between declared invariant and tested behavior.

### Cross-Step References Consistent (45/50)

- Step 4c references "Same cascade behavior as Step 4b" -- Step 4b exists and describes cascade behavior. Valid cross-reference.
- Steps are numbered sequentially (1-6 for happy path, letter suffixes for variants, E-prefix for cross-cutting). No dangling references.
- **-5**: Step 5d says "Delete button is visible" for cancelled milestones, and Step 6b covers deleting cancelled milestones. However Step 5d references no fact about delete permission, while Step 6b includes the permission precondition. These two steps describe overlapping scenarios without cross-referencing each other.

### Risk Level Consistent (38/40)

- `High` risk is justified: the journey involves irreversible state transitions (cancelled is terminal with cascade unbinding), state mutation operations, and data loss potential.
- The journey covers auth (E2), permissions (6d), cascade unbinding (4b, 4c, 5), and terminal state constraints (5b).
- **-2**: The journey does not test permission escalation (e.g., user with milestone:read trying milestone:delete) which is typically part of High-risk security testing. Step 6d tests absence of delete button without permission but not a direct unauthorized API call.

---

## Blindspot Hunt

### [blindspot] Missing completed -> in_progress reopen path

The PRD state machine explicitly includes `completed -> in_progress` ("PM 重新开启"). The journey's invariants declare this transition but no step tests it. A PM who accidentally completes a milestone needs to reopen it. This is a distinct behavior from cancellation that affects completion percentage recalculation and MI binding state.

> Quote from journey invariants: "Status machine: not_started -> in_progress -> completed (with rollback to cancelled from any non-terminal state); completed -> cancelled (manual cancel); completed -> in_progress (reopen); cancelled is terminal." -- The reopen transition is declared but untested.

### [blindspot] Empty milestone completion edge case

Step 4 always assumes associated MainItems exist. The PRD states completion = MI average, empty = 0. An empty milestone (zero associated MIs) has no non-terminal MIs, so the completion guard is vacuously satisfied. This edge case is never tested. Downstream tests would require at least one MI to pass the completed transition.

> Quote from Step 4: "All associated MainItems are in terminal states (completed/closed)" -- this is stated as a precondition, not tested as a variable (0 MIs vs all-terminal MIs).

### [blindspot] API surface lacks journey-style flow

Steps E2-E5 are isolated contract-level tests. No step exercises a multi-step API flow (e.g., create via POST, read via GET, update via PUT, transition status, delete). The web surface has a 6-step journey flow but the API surface has only disconnected one-shot tests. A downstream agent generating API tests would produce integration tests that never verify cross-request state persistence.

> Quote from Step E2: "An unauthenticated request is sent to any Milestone endpoint" -- no step chains API calls together.

### [blindspot] Name uniqueness tested only on API surface

Step E4 tests duplicate name via API POST but the web-surface create flow (Steps 1, 1b-1g) never tests what happens when a duplicate name is submitted through the web form. If name uniqueness is enforced at the API level, the web form should display an appropriate error, but this is untested.

> Quote from Step 1: "fills in name (1-100 chars)" -- no mention of uniqueness constraint. Step E4 tests it only on API surface.

### [blindspot] Step E2 "any Milestone endpoint" is untestable

Step E2 says "An unauthenticated request is sent to any Milestone endpoint" -- "any" is not actionable. A downstream agent cannot generate a test for "any endpoint." It must be either: (a) a parameterized test iterating over all endpoints, or (b) a representative endpoint. The step should specify which endpoints or declare it as parameterized.

> Quote from Step E2: "An unauthenticated request is sent to any Milestone endpoint." -- "any" is ambiguous for test generation.

---

## Summary of Required Revisions

1. **Add completed -> in_progress reopen step**: Add a step (e.g., Step 5a or Step 4d) testing the reopen transition from completed to in_progress, verifying status change and completion recalculation.
2. **Add empty milestone completion step**: Add a step testing successful completion of a milestone with zero associated MainItems.
3. **Add web-surface duplicate name test**: Add a step testing the web form's response to submitting a duplicate milestone name within the same map.
4. **Specify Step E2 endpoints**: Replace "any Milestone endpoint" with either a list of specific endpoints or a declared parameterization strategy.
5. **Add API journey flow**: Consider adding at least one multi-step API flow (create -> read -> update -> transition -> delete) to complement the contract-level tests.
6. **Add navigation-guard step**: Add a step testing browser navigation behavior when unsaved changes exist in the edit dialog.
7. **Separate preconditions from actions in Step 4**: Move "All associated MainItems are in terminal states" from the User Action block to a distinct Precondition block.
8. **Replace "soft-deleted" and "same transaction" with user-observable behavior**: In Steps 5, 6, 6b -- describe what the user sees, not the implementation mechanism.

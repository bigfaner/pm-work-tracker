---
journey: "milestone-lifecycle"
eval_type: "contract-eval"
iteration: 1
date: "2026-06-08"
evaluator: "QA-Adversary"
score: 665
---

# Contract Evaluation: milestone-lifecycle (Iteration 1)

**Score: 665 / 1000**

---

## Dimension Scores

| Dimension | Score | Threshold | Status |
|-----------|-------|-----------|--------|
| 1. Completeness | 140/200 | 120 | PASS |
| 2. Semantic Purity | 160/200 | 120 | PASS |
| 3. Precondition Exclusivity | 95/150 | 90 | PASS (barely) |
| 4. Fact Alignment | 90/150 | 90 | PASS (barely) |
| 5. Surface Fitness | 95/150 | 90 | PASS |
| 6. Internal Consistency | 85/150 | 90 | FAIL |

---

## 1. Completeness — 140/200

### Every Outcome has all four mandatory dimensions (Preconditions, Input, Output, State) — 55/70

All 29 outcomes across 6 contracts include Preconditions, Input, Output, and State dimensions. Every outcome also has a Side-effect field (consistently "none" for non-mutating outcomes). This is structurally complete.

Deductions:
- Step 1 "success" Output mentions "completion 0" but State does not mention the initial `completion` computed field. The State says "status not_started, completion 0" which is correct. Minor gap: no mention of `relatedMICount` defaulting to 0. (-5)
- Step 5 "panel-muted-appearance" Output mentions "delete button visible" but there is no separate Outcome testing what happens when that delete button is clicked from the cancelled state -- that is covered in Step 6 "delete-cancelled", but the cross-reference is implicit. (-5)
- Step 1 "session-expired" State says "after re-authentication, milestone retains original values" which is nonsensical for a create operation -- there are no "original values" since the milestone was never created. This is a semantic error in the State dimension. (-5)

### Journey Invariants section present with at least one entry in every Contract — 50/60

All 6 contracts have Journey Invariants sections with entries. Step 3 has only 2 invariants, and Step 5 has only 3, which is thinner than the others but still passes the minimum.

Deduction:
- Step 3 invariants do not mention the BR-5 guard (parent map terminal blocks transitions) despite the step dealing with status transitions. (-5)
- Step 5 invariants say "cancelled is a terminal state with no status transitions available" but the step action is "Cancelled state interactions" -- it does not test the cancel transition itself, only post-cancel behavior. The invariant about terminal state is correct but the contract confuses the journey flow: Step 4 handles cancel, Step 5 handles post-cancel. (-5)

### Outcomes cover happy path + required derived scenarios — 35/70

Happy paths are covered. Several journey edge cases are covered. However, significant gaps:

Missing outcomes (derived from journey):
1. **Step 1**: No "server-error" outcome. The journey has Step 1e (server error) but the contract omits it. (-5)
2. **Step 1**: No "loading-state" outcome. The journey has Step 1f (loading state) but the contract omits it. (-5)
3. **Step 1**: No "cancel-dialog" outcome. The journey has Step 1g (cancel create dialog) but the contract omits it. (-5)
4. **Step 2**: No "concurrent-modification" outcome. The journey has Step 2c but the contract omits it. (-5)
5. **Step 2**: No "cancel-dialog" outcome. The journey has Step 2d but the contract omits it. (-5)
6. **Step 3**: No "server-error" outcome. The journey has Step 3c but the contract omits it. (-5)
7. **Step 6**: No "cancel-delete-confirmation" outcome. The journey has Step 6e but the contract omits it. (-5)
8. **Step 4**: No "not_started -> cancelled" outcome. The journey has Step 4b which is a distinct precondition from Step 4c "in_progress -> cancelled". The "cancel-cascade" outcome mentions "not_started or in_progress" as a combined precondition, but the journey treats them as separate edge cases. (-5)

---

## 2. Semantic Purity — 160/200

### Dimension values use natural language, not code/regex — 65/80

Most outcomes use clear natural language. Issues:
- Step 1 "unauthorized-api" Input: "Unauthenticated POST to /api/v1/teams/:teamId/milestone-maps/:mapId/milestones" -- uses API route with path parameters. This is implementation-coupled. (-5)
- Step 3 "unauthorized-api" Input: "Unauthenticated PUT to /api/v1/teams/:teamId/milestones/:milestoneId/status" -- same issue. (-5)
- Step 6 "api-not-found" Input: "Authenticated request to GET /api/v1/teams/:teamId/milestones/{non-existent-id}" -- uses curly-brace placeholder mixed with route format. (-5)

### Preconditions are declarative state descriptions — 50/60

Most preconditions are clean declarative statements. Issues:
- Step 1 "terminal-parent-map" Preconditions: "Parent MilestoneMap is in terminal state (completed or cancelled)" -- good, but parenthetical enumeration of specific states adds implementation detail. Minor. (-3)
- Step 2 "session-expired" Preconditions: "User session has expired while edit dialog is open" -- this is a valid precondition describing world state. Acceptable. (no deduction)
- Step 5 "cannot-receive-bindings" Preconditions: "Milestone is in cancelled status" -- declarative. Good. (no deduction)

Deduction for Step 1 "unauthorized-api": "API request sent without valid credentials" -- this is technically declarative but "valid credentials" is vague; does it mean expired token, missing header, or invalid token? (-7)

### No implementation coupling — 45/60

The contracts reference HTTP methods and routes in several "unauthorized-api" outcomes:
- Step 1: "Unauthenticated POST to /api/v1/teams/:teamId/milestone-maps/:mapId/milestones" (-3)
- Step 2: "Unauthenticated PUT to /api/v1/teams/:teamId/milestones/:milestoneId" (-3)
- Step 3: "Unauthenticated PUT to /api/v1/teams/:teamId/milestones/:milestoneId/status" (-3)
- Step 4: same pattern (-3)
- Step 6: "Unauthenticated DELETE to /api/v1/teams/:teamId/milestones/:milestoneId" (-3)

The contracts also reference HTTP status codes (401, 409, 404) in Outputs which is implementation coupling to the API surface. While acceptable for API-surface outcomes, these appear in contracts that are supposed to be surface-agnostic journey contracts. (-0 -- these are surface-tagged outcomes, so partial allowance)

---

## 3. Precondition Exclusivity — 95/150

### Preconditions distinct across Outcomes — 35/60

Most outcomes have distinct preconditions. Issues:

- **Step 4 "cancel-cascade"**: Preconditions combine "not_started or in_progress status" into a single outcome. The journey (Step 4b and Step 4c) treats these as separate edge cases because the user experience and the cascade behavior context differ. The combined precondition makes it ambiguous which specific scenario is being tested. (-10)
- **Step 5 "panel-muted-appearance"** vs **Step 5 "cannot-receive-bindings"**: Both have the same precondition "Milestone is in cancelled status". They are distinguishable by Input (viewing vs. binding), but the preconditions are not distinct. (-5)
- **Step 6 "non-deletable-status"** vs **Step 6 "no-permission"**: Both produce the same output ("Delete action is not displayed") with different preconditions. These are properly distinct. (no deduction)

### Preconditions sufficient to uniquely select Outcome — 35/50

Most outcomes can be uniquely selected by their preconditions + input combination. Issues:

- **Step 1 "validation-error-missing-name"** vs **Step 1 "validation-error-missing-date"**: These have nearly identical preconditions ("Create dialog is open; [field] is empty/missing"). They are only distinguishable by which field is problematic, which is captured in the Input. This is acceptable but borderline -- a test automation agent would need to parse both Preconditions and Input to disambiguate. (-5)
- **Step 4 "cancel-cascade"**: The precondition "not_started or in_progress" covers two distinct states. An agent cannot determine from the precondition alone which starting state is being tested. (-5)
- **Step 1 "terminal-parent-map"** vs **Step 1 "success"**: The success precondition includes "parent map is not terminal" while terminal-parent-map says "Parent MilestoneMap is in terminal state". These are properly mutually exclusive. (no deduction)

### Error/boundary Outcomes state triggering conditions — 25/40

Error outcomes generally state their triggering conditions. Issues:

- **Step 1 "duplicate-name"**: Preconditions say "A milestone with the same name already exists in the same map" -- this is a clear triggering condition. Good.
- **Step 1 "terminal-parent-map"**: Preconditions say "Parent MilestoneMap is in terminal state (completed or cancelled)" -- clear. Good.
- **Step 4 "incomplete-items"**: Preconditions say "at least one associated MainItem is not in a terminal state" -- clear triggering condition.
- **Step 5 "cannot-receive-bindings"**: Does not specify what "tries to bind" means at the UI level. Is it via drag-and-drop? Via an edit dialog? Via an API call? The Input says "PM tries to bind a MainItem to this cancelled milestone" but does not specify the interaction mechanism. (-5)
- **Step 3**: No error outcome for invalid status transition (e.g., trying to jump directly from not_started to completed). The transition table says this is not allowed but there is no contract outcome testing it. (-5)
- **Step 6 "api-not-found"**: Input is a GET request but the step is about deletion. This is inconsistent -- the API-not-found test should relate to the DELETE endpoint, not GET. (-5)

---

## 4. Fact Alignment — 90/150

### Factual claims traceable to fact_id or marked UNKNOWN — 40/60

The contracts reference facts via `<!-- fact: ... -->` annotations in the journey file. The contracts themselves do not include fact_id annotations on individual outcomes. Some outcomes have `<!-- source: inferred -->` annotations.

Issues:
- Step 1 "duplicate-name" references "(milestone_service.go:66-72)" in its reasoning comment, which is a code file reference, not a PRD fact. This is an implementation reference, not a source-of-truth document. (-5)
- Step 1 "terminal-parent-map" references "(milestone_service.go:61-63)" -- same issue. (-5)
- Step 4 "cancel-cascade" claims "all associated MainItems auto-unbound in same transaction" but does not cite a PRD fact. The journey Step 4b cites "<!-- fact: prd-spec -- cancel cascade unbinds all MIs -->" but the contract omits the fact reference. (-5)
- Step 6 "success" claims "Milestone soft-deleted; MI milestone_keys cleared; all in single transaction" -- no fact_id. The journey Step 6 cites a fact. (-5)

### Inferred claims have required_outcomes rule support — 25/50

Inferred outcomes are marked with `<!-- source: inferred -->`. Issues:

- Step 1 "session-expired" is marked inferred with reasoning "Web surface mandatory session-expired outcome for form submission steps" -- this is a surface-level rule, not a business rule inference. Acceptable but thin. (-0)
- Step 1 "terminal-parent-map" is inferred from code references (milestone_service.go:61-63), not from PRD business rules. The PRD spec BR-5 does support this, but the contract's reasoning cites code instead of the PRD. (-5)
- Step 1 "duplicate-name" is inferred from code (milestone_service.go:66-72). The PRD does not explicitly state a duplicate name rule -- it mentions "milestone name uniqueness within map" only in the API handbook (DUPLICATE_NAME error code). This is a gap: the contract references code, not a requirement document. (-5)
- Step 4 "cancelled-is-terminal" is marked as an outcome in Step 4, but the journey has it in Step 5b. This is a structural mismatch in inference. (-5)
- Step 4 "cancel-cascade" combines not_started and in_progress preconditions into one outcome. The journey treats them as separate steps (4b and 4c). The inference collapses two distinct scenarios. (-5)
- No outcomes are marked UNKNOWN. The contracts appear to assume complete knowledge. (-5)

### No hallucinated claims — 25/40

Most claims are grounded. Issues:

- **Step 4 "success"** Output: "completion percentage reflects associated MI average" -- the tech design confirms completion is computed as MI completion average. However, the PRD says completion for `completed` milestones is the MI average. This is accurate. (no deduction)
- **Step 1 "session-expired"** State: "No new milestone created; after re-authentication, milestone retains original values" -- for a CREATE operation, "milestone retains original values" is nonsensical because no milestone was ever created. This is a hallucinated State description that was copied from an edit pattern. (-5)
- **Step 4 "cancelled-is-terminal"** is placed in Step 4 contract (transition to completed) but logically belongs in Step 5 (cancelled state interactions). The journey places it at Step 5b. This is a misplacement, not a hallucination, but it creates confusion. (-5)
- **Step 3** claims "Status transitions follow defined state machine: not_started -> in_progress -> completed" as a journey invariant, but this omits the cancelled path and the completed -> in_progress (reopen) path. The actual state machine in the PRD is more complex. This is a factual simplification that could mislead test generation. (-5)

---

## 5. Surface Fitness — 95/150

### Mandatory derived outcomes present — 40/60

For the **web** surface: validation-error outcomes are present in Step 1 (3 validation outcomes). session-expired is present in Steps 1, 2, 3. Missing:
- No server-error outcomes in any contract (journey has Steps 1e, 3c). (-5)
- No loading-state outcomes in any contract (journey has Step 1f). (-5)
- No cancel-dialog outcomes in any contract (journey has Steps 1g, 2d, 6e). (-5)

For the **api** surface: unauthorized (401) outcomes are present in all 6 contracts. not-found (404) is present in Step 6. duplicate-name (409) is present in Step 1. terminal-parent (400 MAP_IS_TERMINAL) is present in Step 1.

### Surface-appropriate language — 40/50

Web outcomes use user-facing language ("PM fills name", "dialog closes", "timeline refreshes"). API outcomes use technical language ("HTTP 401", "Unauthenticated POST"). This is appropriate for the dual surface.

Issues:
- Step 4 "cancel-cascade" mixes web language ("PM selects the Cancelled status option") with API-level detail ("all MI milestone_keys cleared in transaction"). The Side-effect field contains transaction-level detail that is not surface-appropriate for web. (-5)
- Step 6 "success" Side-effect: "Unbind all associated MainItems within delete transaction" -- again, transaction-level detail in a web-surface step. (-5)

### TUI timeout check (skip for non-TUI) — 15/40

This journey does not have a TUI surface. The rubric says "skip for non-TUI" but awards 0-40 points. Since there is no TUI surface, I am awarding partial credit for the clear absence of inappropriate TUI timeout claims. The surface types are correctly declared as ["web", "api"]. However, the rubric explicitly allocates 40 points for this sub-dimension, and without a TUI surface, the full points cannot be justified. Awarding 15 as baseline acknowledgment that no TUI timeout issue exists.

---

## 6. Internal Consistency — 85/150 (FAILS THRESHOLD)

### Invariants hold in every Contract — 30/60

Issues:
- **Step 3 invariant**: "Status transitions follow defined state machine: not_started -> in_progress -> completed." This is an oversimplified state machine. The actual PRD state machine includes cancelled and completed->in_progress (reopen) transitions. The invariant as stated is factually incomplete. (-10)
- **Step 1 invariant**: "Cancelled milestones cannot receive new MainItem bindings." This invariant is about cancelled milestones, but Step 1 is about CREATE. The invariant is copied from the journey-level invariants and is irrelevant to the create step. It should be in Step 5. (-5)
- **Step 1 invariant**: "A milestone can only be marked completed when all its associated MainItems are in terminal states." Again, irrelevant to the create step. (-5)
- **Step 2 invariant**: "Parent map must not be terminal for status-relevant updates (name, date)." But Step 2 is about editing information, not status transitions. The API handbook shows Update Milestone returns MAP_IS_TERMINAL error, so this is relevant. However, Step 2 has no outcome testing the terminal-parent-map scenario for edits. The invariant is stated but not covered by any outcome. (-5)
- **Step 5 invariant**: "Panel for cancelled milestones shows muted tone, empty MI list, no add control, but visible delete." This is an invariant that describes UI appearance, not a behavioral invariant. It cannot be verified by a test script in a meaningful way without coupling to CSS/rendering. (-5)

### Cross-Contract state references consistent — 25/50

Issues:
- **Step 4 "cancelled-is-terminal"** outcome conflicts with Step 5 "cancelled is terminal" invariant. The same behavioral claim appears in two contracts, creating maintenance risk. If one is updated, the other may not be. (-5)
- **Step 4 "cancel-cascade"** State says "Milestone status set to cancelled (terminal); all MI milestone_keys cleared in transaction". Step 5 "panel-muted-appearance" Preconditions: "Milestone is in cancelled status". The state produced by Step 4 is consumed by Step 5. However, Step 5 also says "associated MI list is empty" which is consistent with the unbind. This is correct. (no deduction)
- **Step 3** produces state "Milestone status updated to in_progress". Step 4 "success" Preconditions: "Milestone is in in_progress status". This is consistent with the journey flow (Step 3 feeds Step 4). However, Step 4 also has an outcome "cancel-cascade" with precondition "Milestone is in not_started or in_progress status". If the journey flows Step 3 -> Step 4, the milestone is already in_progress, so "not_started" in the combined precondition would never be reached through this flow. This means the contract mixes a sequential outcome (from the happy path) with an alternative-path outcome (from a branch), which is correct for contract coverage but creates confusion about the sequential flow. (-5)
- **Step 6 "delete-cancelled"** Preconditions: "Milestone is in cancelled status". This outcome can only be reached after Step 4 or Step 5, which produce cancelled state. Consistent. But Step 6 also has "success" with precondition "Milestone is in not_started status" which is the state from Step 1 (if we skip Steps 2-5). The contract set assumes the full journey flow but also includes branch outcomes. This is acceptable. (-0)

Larger issue:
- **Step 2** has no outcome for the terminal-parent-map scenario. Step 1 has "terminal-parent-map". Step 3 and Step 4 preconditions include "parent map is not terminal". But Step 2 (edit) has the invariant "Parent map must not be terminal for status-relevant updates" but no corresponding outcome. The API handbook confirms Update Milestone can return MAP_IS_TERMINAL. This is a cross-contract gap. (-10)
- **Step 3** has no outcome for the terminal-parent-map scenario either, despite its precondition including "parent map is not terminal". The BR-5 rule applies to status transitions too (as shown in tech-design BR-5: "UpdateStatus -- guard"). (-5)

### Preconditions consistent with preceding Steps' State changes — 30/40

Issues:
- **Step 5 "cannot-receive-bindings"**: Preconditions say "Milestone is in cancelled status". The preceding Step 4 "cancel-cascade" produces state "Milestone status set to cancelled (terminal)". Consistent. But Step 5 also includes "unauthorized-api" which has no precondition about milestone status. This is a generic surface outcome, acceptable. (-0)
- **Step 6 "success"**: Preconditions "Milestone is in not_started status". In the happy path flow (Step 1 -> 2 -> 3 -> 4 -> 5 -> 6), the milestone is cancelled by Step 4/5. The "success" outcome for delete assumes a not_started milestone, which means the journey flow must branch -- a different milestone or skipping steps. The contract does not clarify this. (-5)
- **Step 6 "delete-cancelled"**: Preconditions "Milestone is in cancelled status". This is consistent with the state produced by Step 4/5 in the main flow. However, the journey's Step 6 is labeled "Delete milestone in not_started status" (the happy path), while Step 6b is "Delete cancelled milestone" (the edge case). The contract's "success" outcome (not_started) and "delete-cancelled" outcome cover both, but the "success" outcome is labeled as the primary outcome despite being unreachable from the sequential flow. (-5)

---

## Blindspot Hunt

### [blindspot] 1: Missing server-error outcomes across all contracts

The journey defines server-error edge cases at Steps 1e, 2c (concurrent modification), and 3c. None of the 6 contracts include a server-error outcome. For a web+API surface journey, server errors are a critical derived scenario.

**Quote**: Journey Step 1e: "User Action: PM submits and the backend returns an error. Expected Result: Page displays a retryable error message. Dialog stays open with data preserved."

**Impact**: No contract tests server-error handling. A test generation agent would produce no tests for this scenario, leaving a gap in coverage.

### [blindspot] 2: Missing cancel-dialog outcomes

Journey Steps 1g, 2d, and 6e define cancel-dialog scenarios. None of the contracts include a "dialog-cancelled" outcome for their respective steps.

**Quote**: Journey Step 1g: "User Action: PM cancels or closes the dialog. Expected Result: Dialog closes, no operation is performed."

**Impact**: No test coverage for dialog cancellation, a fundamental web surface interaction pattern.

### [blindspot] 3: Step 1 session-expired State contains nonsensical claim

**Quote**: Step 1 contract, Outcome "session-expired", State: "No new milestone created; after re-authentication, milestone retains original values"

For a CREATE operation, the phrase "milestone retains original values" is logically impossible since the milestone was never created. This appears to be a copy-paste error from an edit/update pattern.

### [blindspot] 4: Step 4 combines two distinct journey edge cases into one outcome

**Quote**: Step 4 contract, Outcome "cancel-cascade", Preconditions: "Milestone is in not_started or in_progress status and has associated MainItems"

Journey Step 4b tests "not_started -> cancelled" and Step 4c tests "in_progress -> cancelled" as separate edge cases with different preconditions. The contract merges them, losing the distinction. A test agent may only generate one test instead of two.

### [blindspot] 5: Step 3 invariant oversimplifies the state machine

**Quote**: Step 3 contract, Journey Invariants: "Status transitions follow defined state machine: not_started -> in_progress -> completed."

The actual state machine (per PRD) includes: not_started -> in_progress, in_progress -> completed, completed -> in_progress (reopen), any non-terminal -> cancelled, completed -> cancelled. The invariant as written omits the cancelled and reopen paths entirely, which could lead an agent to generate an incomplete state machine test.

### [blindspot] 6: Step 2 has no outcome testing terminal-parent-map for edit operations

**Quote**: Step 2 contract, Journey Invariants: "Parent map must not be terminal for status-relevant updates (name, date)."

The invariant is stated but no outcome tests this condition. The API handbook confirms Update Milestone returns MAP_IS_TERMINAL. A test agent would not generate a test for this business rule in Step 2.

### [blindspot] 7: Step 6 api-not-found outcome uses GET instead of DELETE

**Quote**: Step 6 contract, Outcome "api-not-found", Input: "Authenticated request to GET /api/v1/teams/:teamId/milestones/{non-existent-id}"

The step action is "Delete milestone" and the success outcome describes a delete operation. The api-not-found outcome tests a GET request, which is inconsistent with the step context. The API handbook defines DELETE /milestones/:milestoneId, so the not-found test should use DELETE.

### [blindspot] 8: Step 5 "panel-muted-appearance" has no actionable test assertion

**Quote**: Step 5 contract, Outcome "panel-muted-appearance", Output: "Panel displays with muted visual tone; associated MI list is empty; add button not shown; delete button visible"

"muted visual tone" is a subjective visual description that cannot be asserted by an automated test. The outcome mixes testable assertions (empty list, button visibility) with untestable ones (visual tone). A test agent would not know how to verify "muted visual tone".

### [blindspot] 9: No outcome tests the completed -> in_progress (reopen) transition

The PRD state machine explicitly defines: "completed -> in_progress: PM reopens". The journey does not include this as a step or edge case, and no contract has an outcome testing it. This is a valid state transition per the PRD that is completely uncovered.

**Quote**: PRD spec, milestone state machine: "completed | in_progress | PM reopens"

### [blindspot] 10: Step 1 irrelevant invariants copied from journey level

**Quote**: Step 1 contract, Journey Invariants:
- "Cancellation of a milestone automatically unbinds all associated MainItems within the same transaction."
- "A milestone can only be marked completed when all its associated MainItems are in terminal states."
- "Cancelled milestones cannot receive new MainItem bindings."

None of these invariants are relevant to the CREATE operation in Step 1. They are copied from the journey-level invariants without filtering for step relevance. This could mislead a test agent into generating irrelevant assertions for Step 1 tests.

### [blindspot] 11: Missing validation-error for edit (Step 2)

Step 2 "edit milestone" has no validation-error outcomes. If a user edits the name to exceed 100 characters or clears the name entirely during edit, there should be validation outcomes. The journey does not define these edge cases for Step 2, but the API handbook shows Update Milestone returns INVALID_PARAMS for name validation failures. This is a gap in both the journey and the contracts.

### [blindspot] 12: No permission-denied outcome for web surface

All contracts include "unauthorized-api" for API surface (401 responses), but no contract includes a "forbidden" outcome for authenticated users without the required permission. Step 6 has "no-permission" where the delete action is hidden, but Steps 1-5 have no equivalent. The API handbook shows FORBIDDEN (403) error responses, but no contract tests authenticated-but-unauthorized access.

---

## Summary of Required Improvements

1. Add server-error outcomes to Steps 1, 2, and 3
2. Add cancel-dialog outcomes to Steps 1, 2, and 6
3. Fix Step 1 session-expired State to remove "retains original values" (create has no original values)
4. Split Step 4 "cancel-cascade" into separate outcomes for not_started and in_progress
5. Fix Step 3 state machine invariant to include cancelled and reopen paths
6. Add terminal-parent-map outcome to Step 2 and Step 3
7. Fix Step 6 api-not-found Input to use DELETE instead of GET
8. Make Step 5 "panel-muted-appearance" Output testable (remove "muted visual tone")
9. Add reopen (completed -> in_progress) outcome to Step 4
10. Remove irrelevant invariants from Step 1 (keep only create-relevant invariants)
11. Add validation-error outcomes to Step 2 for name constraints
12. Add permission-denied/forbidden outcomes for authenticated users without permission

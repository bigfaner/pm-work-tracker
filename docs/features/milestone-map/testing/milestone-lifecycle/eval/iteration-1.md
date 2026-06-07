# Eval Report: milestone-lifecycle Journey (Iteration 1)

**Date**: 2026-06-08
**Score**: 685/1000
**Threshold**: 850
**Result**: FAIL

---

## Dimension Scores

| Dimension | Score | Min | Status |
|-----------|-------|-----|--------|
| 1. Completeness | 155/200 | 120 | PASS |
| 2. Semantic Purity | 180/200 | 120 | PASS |
| 3. Precondition Exclusivity | 125/150 | 90 | PASS |
| 4. Fact Alignment | 75/150 | 90 | **FAIL** |
| 5. Surface Fitness | 50/150 | 90 | **FAIL** |
| 6. Internal Consistency | 100/150 | 90 | PASS |
| **Total** | **685/1000** | **850** | **FAIL** |

---

## Dimension 1: Completeness — 155/200

### Journey Metadata (45/50)

- Name `milestone-lifecycle` follows kebab-case. Risk level `High` is present with a classification comment. Surface types `["web", "api"]` and sources are listed. Generated date present.
- **-5**: No explicit narrative justification for `High` risk level within the journey body; the classification criteria comment is generic, not specific to this journey's content.

### Steps Complete (70/80)

- Steps 1-6 cover the core happy path (create, edit, status transitions, delete) with clear actions and expected results.
- Edge cases 1b-1g, 2b-2d, 3b-3c, 4b-4c, 5b-5d, 6b-6e provide comprehensive boundary coverage.
- **-10**: Some expected results lack specificity about what the user observes. For example, Step 1 says "timeline refreshes showing the new node at the correct date position" but does not define what "correct date position" means for a downstream agent.

### Outcomes Coverage (40/70)

- Happy path outcomes are present for all 6 steps.
- Validation errors (1b, 1c, 1d), server errors (1e, 3c), loading states (1f), and concurrency (2c) are covered.
- **-15**: Missing `session-expired` outcome (mandatory for web surface per `surface-web.md`).
- **-15**: Zero API-level outcomes despite declaring `api` surface type. No `unauthorized`, `not-found`, or `conflict` outcomes.

---

## Dimension 2: Semantic Purity — 180/200

### Outcome Descriptions Use Natural Language (75/80)

- All outcomes are in natural language: "Form displays error", "Status changes to cancelled", "Dialog closes".
- No regex patterns, CSS selectors, XPath, or assertion calls anywhere.
- **-5**: Step 1e says "the backend returns 500" -- this leaks a backend implementation detail (HTTP status code) into a web-surface user action description. Should be "server encounters an error."

### Preconditions Are Declarative (50/60)

- Preconditions like "Create dialog is open" and "Milestone is in cancelled status" are declarative statements of required state.
- **-10**: Some preconditions are implicit. Step 4 ("All associated MainItems are in terminal states") embeds a precondition inside the User Action block rather than separating it. This makes it harder for a downstream agent to identify the precondition independently.

### No Implementation Coupling in Steps (55/60)

- Steps describe user-level actions (clicking buttons, selecting from dropdowns, filling forms). No API endpoints, database queries, or function calls.
- **-5**: Step 6 mentions "Associated MainItems have their milestone_key set to null within the same transaction" -- this is an implementation detail (column name, transaction scope) in an expected result.

---

## Dimension 3: Precondition Exclusivity — 125/150

### Preconditions Distinct Across Outcomes (50/60)

- Each step has a single outcome, so there is no ambiguity between competing outcomes within a step.
- Steps 4b and 4c test cancellation from different source states (not_started vs in_progress) with distinct preconditions.
- **-10**: Steps 4b and 4c have overlapping action descriptions ("PM clicks the status Badge and selects Cancelled") and their preconditions differ only in the starting status. The expected results are nearly identical ("Same cascade behavior as Step 4b"). This creates a semantic near-duplicate that a downstream agent might collapse.

### Preconditions Sufficient to Uniquely Select Outcome (45/50)

- Each step's precondition + action uniquely determines one outcome. No step has multiple competing outcomes.
- **-5**: Step 1f (loading state) has precondition "Create dialog is open with valid data" which overlaps with Step 1's precondition. The distinguishing factor is "request is in flight" which is a transient state, not a precondition that can be set up.

### Missing Preconditions for Error/Boundary Outcomes (30/40)

- Most error/boundary outcomes state their triggers clearly.
- **-10**: Step 3c (server error during transition) has precondition "PM selects a new status from the Badge dropdown" -- this is an action, not a precondition describing the required state. The actual precondition (server will return an error) is not declarable by the test agent.

---

## Dimension 4: Fact Alignment — 75/150

### Factual Claims Traceable to fact_id or UNKNOWN (30/60)

- The journey references BR-1, BR-3, BR-4 informally but uses its own numbering scheme rather than the BIZ-milestone-XXX identifiers from the business rules context.
- Error messages in expected results (e.g., "All associated items must be completed before marking as complete") are English translations that do not match the tech design's actual error messages (e.g., "里程碑下存在未完成的事项"). These are unverified claims without `UNKNOWN` marking.
- **-30**: No `fact_id` traceability. No claims are marked `UNKNOWN`. All factual assertions are presented as verified facts without evidence linkage.

### Inferred Claims Have Required Outcomes Rule Support (15/50)

- Derived outcomes like validation-error (Steps 1b-1d) and server-error (Steps 1e, 3c) are present but lack `source: inferred` annotations.
- No citation of which `required_outcomes` rule from `surface-web.md` or `surface-api.md` triggered each derived outcome.
- **-35**: Zero inferred-outcome annotations in the entire document.

### No Hallucinated Unclassified Claims (30/40)

- Most claims are grounded in the tech design. UI-level error messages are reasonable inferences for a journey document.
- The status machine description in the invariants is inaccurate (see Internal Consistency), which is a factual error rather than a hallucination.
- **-10**: Step 2c claims "Conflict notification: 'Data has been modified by someone else, please refresh and retry'" -- this specific error message does not appear in any source document. It is an unverified claim presented as fact.

---

## Dimension 5: Surface Fitness — 50/150

### Mandatory Derived Outcomes Present (10/60)

- **Web surface**: `validation-error` is present (Steps 1b, 1c, 1d). `session-expired` is **completely absent**.
- **API surface**: `unauthorized` is **completely absent**. No step tests unauthenticated or token-expired access to any of the 7 authenticated Milestone endpoints.
- API-required additional outcomes also absent: `not-found` (milestone does not exist), `conflict` (duplicate milestone name, defined in API handbook as `DUPLICATE_NAME 409`), `validation-error` at API level.
- **-50**: The document declares `surface_types: ["web", "api"]` but provides zero API-surface content. This is a fundamental surface type mismatch.

### Test Strategy Proportions Match Surface Guidance (15/50)

- **Web**: The journey is predominantly journey-style (user workflows) with some contract-level validation (form field errors). Approximately 30/70 journey/contract -- slightly off from the 50/50 target but acceptable.
- **API**: Zero API coverage. The 50/50 balance for API is completely missed.
- **-35**: Complete absence of API-level test strategy.

### Surface-Specific Environment Assumptions Realistic (25/40)

- Web assumptions are realistic: browser-based interaction, dialog boxes, panels, badge dropdowns, loading indicators.
- **-15**: No API environment assumptions at all (HTTP client setup, authentication headers, base URL, etc.). The document pretends API is a surface but provides no API execution context.

---

## Dimension 6: Internal Consistency — 100/150

### Invariants Hold in Every Step (20/60)

- **Invariant violation**: The declared status machine in the invariants says: "not_started -> in_progress -> completed (with rollback to cancelled from any non-terminal state); completed -> cancelled (manual cancel); cancelled is terminal."
  - This contradicts the tech design's `MilestoneTransitions` which shows `completed: {"cancelled", "in_progress"}`. The journey's invariant omits the `completed -> in_progress` rollback.
  - The phrase "rollback to cancelled from any non-terminal state" is self-contradictory because `completed` is a terminal state, yet Step 5 explicitly transitions from `completed` to `cancelled`.
- Step 5 says "This is a terminal state with no further transitions" after transitioning to cancelled -- consistent with the invariant, but the status machine description is inaccurate.
- **-40 (invariant violation deduction)**

### Cross-Step References Consistent (45/50)

- Step 4c references "Same cascade behavior as Step 4b" -- Step 4b exists and describes cascade behavior. Valid.
- No dangling references detected. Steps are numbered sequentially with letter suffixes for variants.
- **-5**: Step 5d says "Delete button is visible (BR-4 allows deleting cancelled milestones)" but this is the only step that cross-references BR-4 for the cancelled-delete case. Step 6b also covers this but does not reference Step 5d.

### Risk Level Consistent (35/40)

- `High` risk is justified: the journey involves irreversible state transitions (cancelled is terminal), cascade unbinding of MainItems (data loss potential), and state mutation operations.
- **-5**: The journey does not involve security-sensitive operations (no auth testing, no permission escalation testing) which typically characterizes `High` risk journeys. `High` is still justified by the data loss risk.

---

## Blindspot Hunt

### [blindspot] Missing parent-context terminal guard (BR-5)

The tech design defines BR-5: "Terminal Map blocks child Milestone create/update." The API returns `MAP_IS_TERMINAL` error when trying to create or update a milestone under a terminal MilestoneMap. This journey never tests what happens when the parent MilestoneMap is in a terminal state. Steps 1-6 all assume a non-terminal parent Map implicitly. A downstream agent would produce tests that never verify this guard.

> Quote from journey Setup: "A milestone map exists and is accessible via its timeline view" -- no mention of status constraint on the parent.

### [blindspot] Duplicate milestone name scenario missing

The API handbook defines `DUPLICATE_NAME` (409 Conflict) for creating a milestone with the same name within the same map. The tech design has `ErrDuplicateMilestoneName` and the service layer checks `ExistsByNameAndMap`. The journey's create steps never test this constraint.

> Quote from journey Step 1: "fills in name (1-100 chars)" -- no mention of uniqueness constraint or duplicate rejection.

### [blindspot] completed -> in_progress rollback untested

The tech design's `MilestoneTransitions` explicitly includes `completed: {"cancelled", "in_progress"}`. The journey tests completed -> cancelled (Step 5) but never tests the completed -> in_progress rollback path. A PM might need to reopen a completed milestone.

> Quote from journey invariants: "Status machine: not_started -> in_progress -> completed (with rollback to cancelled from any non-terminal state)" -- the rollback to in_progress from completed is omitted entirely.

### [blindspot] Empty milestone completion edge case

Tech design BR-1 note 3: "Empty milestones can directly be marked as completed." The journey's Step 4 always assumes the milestone has associated MainItems in terminal states. Step 3b tests the negative case (non-terminal MIs) but never tests the positive case of completing a milestone with zero associated items.

> Quote from Step 4: "All associated MainItems are in terminal states (completed/closed)" -- this is stated as a precondition, not tested as a variable.

### [blindspot] No session-expired handling for long-running operations

The web surface requires `session-expired` as a mandatory outcome. The journey includes long-form interactions (editing, filling forms) where session expiry during the workflow is a realistic scenario. This is completely absent.

> No quote exists -- this is a gap where content should be.

---

## Summary of Required Revisions

1. **Add API surface steps**: Either add API-level steps covering HTTP methods, auth headers, status codes, and error responses, OR remove `api` from `surface_types` if this journey is web-only.
2. **Add `unauthorized` outcome**: For every authenticated Milestone endpoint, add a step or outcome testing missing/invalid credentials.
3. **Add `session-expired` outcome**: Add at least one step where a user's session expires during an active workflow.
4. **Fix status machine invariant**: Add the `completed -> in_progress` transition and clarify terminal vs non-terminal rollback rules.
5. **Add `not-found` outcome**: Test operations on a non-existent milestone.
6. **Add `conflict` outcome**: Test duplicate milestone name creation.
7. **Add BR-5 parent-terminal guard**: Test creating/updating a milestone when the parent MilestoneMap is terminal.
8. **Add `source: inferred` annotations**: Annotate all derived outcomes with their originating required_outcomes rule.
9. **Add fact traceability**: Either reference specific fact_ids or mark claims as `UNKNOWN`.
10. **Test empty-milestone completion**: Add a step for completing a milestone with no associated MainItems.

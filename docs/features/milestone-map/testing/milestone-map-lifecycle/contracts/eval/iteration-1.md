---
iteration: 1
evaluated: "2026-06-08"
scorer: "adversarial-qa"
score: 665
---

# Contract Evaluation Report — Iteration 1

**Target**: docs/features/milestone-map/testing/milestone-map-lifecycle/contracts/ (step-1 through step-8)

---

## Dimension Scores

### 1. Completeness — 130/200

**Every Outcome has all four mandatory dimensions (Preconditions, Input, Output, State) — 50/70**

All outcomes across all eight contracts include Preconditions, Input, Output, and State. Most also include Side-effect. However:
- Step 4 has no `server-error` outcome, unlike Step 3 which does. The journey has Step 3b (server error for status transition) but no corresponding edge case for steps 4 or 5. The pattern is inconsistent.
- Step 5 has no `server-error` outcome either, despite being the same type of operation (status transition).
- Step 1 is missing the `server-error` outcome despite journey step 1f explicitly covering it ("Create during server error").
- Step 1 is also missing the `loading-state` outcome despite journey step 1g covering it and the invariant saying "Create forms display loading state and prevent further interaction during submission."
- Step 2 is missing the `concurrent-modification` outcome despite journey step 2c explicitly covering it.

**Journey Invariants section present with at least one entry in every Contract — 55/60**

All eight contracts have a Journey Invariants section. Step 1 has four entries; steps 3, 4, 5, 6, 7, 8 have 2-4 entries each. Step 2 has three entries. However, the invariants are unevenly distributed -- some contracts repeat the same generic invariants ("All mutation operations require their respective RBAC permissions") rather than step-specific invariants.

**Outcomes cover happy path + required derived scenarios — 25/70**

Significant gaps between journey edge cases and contract outcomes:
- Journey step 1f (server error) has no contract outcome in step-1.
- Journey step 1g (loading state) has no contract outcome in step-1.
- Journey step 1h (create from empty state) has no contract outcome.
- Journey step 1i (create from grid dashed card) has no contract outcome.
- Journey step 2c (concurrent modification) has no contract outcome in step-2.
- Journey steps 4b, 4c, 4d (filter by status/owner/search) have no corresponding contract outcomes.
- Journey step 8d (cancel delete confirmation) has no contract outcome in step-8.
- Journey step E3 (API not-found for GET) is partially covered (only DELETE has api-not-found in step-8, not GET).
- Journey step E4 (API validation-error on create) has no contract outcome in step-1.

This is a major gap -- the contracts do not cover roughly half of the journey's edge cases.

---

### 2. Semantic Purity — 160/200

**Dimension values use natural language, not code/regex — 65/80**

Most dimension values are in natural language. Minor issues:
- Step 1 Input: "PM fills name (1-100 chars), owner (assigneeBizKey)" -- "assigneeBizKey" is implementation jargon, not user-facing language.
- Step 8 Input: "PM triggers delete action and confirms" is appropriately declarative.

**Preconditions are declarative state descriptions — 50/60**

Most preconditions are declarative. However:
- Step 1 success precondition: "User has milestone:create permission; create dialog is open with valid data" -- "valid data" is vague and should specify what valid data consists of.
- Step 2 success precondition: "Milestone map exists; edit dialog is open with pre-filled current values; user has milestone:update permission" -- does not specify which status the milestone map is in. Can you edit in any status?

**No implementation coupling — 45/60**

Most contracts avoid implementation coupling. However:
- API endpoint paths appear in Input fields (e.g., "Unauthenticated POST to /api/v1/teams/:teamId/milestone-maps") which is acceptable for API-surface outcomes.
- Step 6 `success-cancelled` State: "all MI milestone_keys cleared" -- uses implementation field name "milestone_keys" instead of natural language.
- Step 6 `success-cancelled` Side-effect: "Cascade: cancel non-terminal milestones + unbind all MIs in single transaction" -- uses abbreviation "MI" and exposes transaction-level implementation detail.

---

### 3. Precondition Exclusivity — 100/150

**Preconditions distinct across Outcomes — 40/60**

Within individual contracts, preconditions are generally distinct. However:
- Step 1 `validation-error-missing-name` and `validation-error-missing-owner` have overlapping structure: both have "Create dialog is open" as the only state precondition, differing only in which field is empty. This is acceptable for distinct outcomes but makes automated selection ambiguous -- what if BOTH are empty?
- Step 8 `non-deletable-status` and `no-permission` are distinct but could theoretically overlap (a user without permission viewing a non-deletable status map).

**Preconditions sufficient to uniquely select Outcome — 35/50**

Issues:
- Step 1 `validation-error-missing-name`: If both name AND owner are empty, which outcome fires? The contract does not specify combined validation errors.
- Step 2 `no-changes`: The precondition is "Edit dialog is open with pre-filled current values" which is identical to the success precondition minus the permission check. The differentiator is in the Input ("saves without making any modifications") not the Preconditions, which is acceptable but borderline.
- Step 6 `success-completed` vs `completed-with-incomplete-milestones`: Preconditions differ in milestone terminal state -- this is well done.

**Error/boundary Outcomes state triggering conditions — 25/40**

- Step 3 `server-error` precondition: "Backend is unavailable when status transition is attempted" -- this is adequate.
- Missing boundary: What happens if status transition is attempted on a map in the wrong state (e.g., trying planning->reviewed on an already-reviewed map)? No outcome covers this.
- Step 7 (rollback) does not cover the case where rollback is attempted from a state that doesn't allow it (e.g., from executing back to planning).
- Step 5 has no error outcomes at all (no server-error, no wrong-status rejection).

---

### 4. Fact Alignment — 105/150

**Factual claims traceable to fact_id or marked UNKNOWN — 45/60**

Journey file has good fact annotations referencing prd-spec. The contracts inherit these via sources. However:
- Step 6 `success-cancelled` says "all milestones cancelled in cascade" but there is no explicit fact_id annotation in the contract itself -- only the journey has the fact reference.
- Step 7 rollback from reviewed to planning: the journey references "prd-spec MilestoneMap state machine -- reviewed -> planning rollback" but the contract doesn't annotate this fact.
- Contracts lack explicit fact annotations entirely (unlike the journey which has `<!-- fact: prd-spec ... -->` annotations). The contracts only have `<!-- source: inferred -->` or `<!-- source: surface-required -->` annotations.

**Inferred claims have required_outcomes rule support — 35/50**

- Step 2 `no-changes`: The claim "Dialog closes as a no-op, equivalent to Cancel" is inferred but has no supporting rule or fact reference. Is this specified in the PRD?
- Step 6 `success-cancelled` Side-effect: "Cascade: cancel non-terminal milestones + unbind all MIs in single transaction" -- the "single transaction" claim is an implementation detail, not a behavioral requirement. Should be marked as inferred or removed.
- Step 8 `success` State: "all child milestones soft-deleted; all linked MI milestone_keys cleared; all in single transaction" -- again "soft-deleted" and "single transaction" are implementation details.

**No hallucinated claims — 25/40**

- Step 1 success Output: "list refreshes showing new entry with correct name, status badge, and owner info" -- the "status badge" and "owner info" display claims are not explicitly sourced from any fact.
- Step 2 no-changes: "equivalent to Cancel" -- this equivalence claim has no source. What if the backend is still called with no changes? The contract assumes a no-op but this may not be accurate.
- Step 6 `success-cancelled`: The cascade behavior "all milestones cancelled" and "all associated MainItems unbound" is sourced from BIZ-milestone-002 and BIZ-milestone-006, but the contract doesn't reference these business rules.

---

### 5. Surface Fitness — 95/150

**Mandatory derived outcomes present — 35/60**

For web+api surfaces, the following are mandatory:
- **unauthorized (API)**: Present in all 8 contracts. Good.
- **session-expired (web)**: Present in steps 1, 2, 3, 4, 5, 7, 8. **Missing from step 6**. Step 6 is particularly concerning as it involves terminal state transitions with cascade effects.
- **validation-error (web)**: Present in steps 1, 2. Steps 3-8 are status transitions which may not have form validation, so this is acceptable.
- **server-error (web)**: Present only in step 3. **Missing from steps 1, 2, 4, 5, 6, 7, 8** despite the journey having server-error edge cases for step 1 (step 1f) and step 3 (step 3b).
- **api-not-found (API)**: Present only in step 8. **Missing from steps 1-7** despite the journey having E3 (API not-found for GET) and the API surface requiring this outcome.

**Surface-appropriate language — 35/50**

- Web outcomes correctly use user-action language ("PM selects...", "PM submits...").
- API outcomes correctly use request language ("Unauthenticated POST/PUT/DELETE to /api/v1/...").
- However, some web outcomes mix API language: Step 1 `unauthorized-api` uses "Unauthenticated POST to /api/v1/teams/:teamId/milestone-maps" which is API language in a web-focused contract. This is acceptable as a surface-required outcome.
- Step 6 mixes levels: "Cascade: cancel non-terminal milestones + unbind all MIs in single transaction" reads like backend implementation, not surface behavior.

**TUI timeout check — 25/40**

N/A -- no TUI surface. Points awarded as this dimension does not apply. Awarded partial credit as the surfaces are correctly identified as web+api.

---

### 6. Internal Consistency — 75/150

**Invariants hold in every Contract — 25/60**

- Step 2 invariants say "A milestone map can only be deleted when it is in planning, reviewed, or ready status." This is a delete invariant, not an edit invariant. It's irrelevant to step 2 and appears to be copy-pasted from step 8.
- Step 2 invariants say "Edit forms display loading state and prevent further interaction during submission." But step 2 has no loading-state outcome to verify this invariant holds.
- Step 1 invariants say "Create forms display loading state and prevent further interaction during submission." But step 1 has no loading-state outcome.
- Step 5 invariants: "Status transitions follow the defined state machine." and "All mutation operations require their respective RBAC permissions." -- these are generic and don't add step-specific value.
- Step 6 missing the invariant about cancellable statuses -- the journey says "Any non-terminal state can transition to cancelled" but step 6 only covers cancelled from executing.

**Cross-Contract state references consistent — 25/50**

- Step 3 changes status from planning to reviewed. Step 7 rolls back from reviewed to planning. This is consistent.
- Step 8 success precondition: "Milestone map is in planning status" -- but the journey also says milestone maps in reviewed or ready status can be deleted. The contract only covers planning-status deletion, not reviewed or ready.
- Step 6 `success-cancelled` preconditions: "Milestone map is in executing status" -- but the journey invariant says "any non-terminal state can transition to cancelled." The contract is more restrictive than the journey invariant. This is inconsistent.
- Step 8 `non-deletable-status` precondition says "executing or completed status" but doesn't mention cancelled. Can a cancelled milestone map be deleted? The journey invariant says only "planning, reviewed, or ready" are deletable, so cancelled should also be non-deletable, but it's not listed.

**Preconditions consistent with preceding Steps' State changes — 25/40**

- Step 5 preconditions require "ready status" which is the state set by step 4. Consistent.
- Step 6 preconditions require "executing status" which is the state set by step 5. Consistent.
- Step 7 preconditions require "reviewed status" -- but in the happy path flow, by step 7 the map would be in executing status (steps 3->4->5->6). Step 7 is described as a rollback "from reviewed back to planning" but it's placed after step 6. The ordering implies the map would need to be reset to reviewed before step 7 can execute, but no step resets it. This is a journey-level ordering issue, but the contract doesn't acknowledge this context.

---

## Blindspot Findings

1. **[blindspot] Missing session-expired in step 6**: Step 6 is a web-surface mutation operation involving terminal-state transitions with cascade side effects, but it lacks the mandatory `session-expired` outcome. Quote from step-6 contract: only outcomes listed are "success-completed", "success-cancelled", "completed-with-incomplete-milestones", "completed-is-terminal", and "unauthorized-api". No session-expired outcome exists. Steps 4 and 7 correctly include session-expired.

2. **[blindspot] Missing server-error outcomes**: Step 1 has no server-error outcome despite journey step 1f explicitly describing this scenario. Quote from step-1 contract: outcomes are "success", "validation-error-missing-name", "validation-error-name-too-long", "validation-error-missing-owner", "validation-error-invalid-date-range", "session-expired", "unauthorized-api". No server-error. Steps 2, 4, 5, 7, 8 also lack server-error outcomes.

3. **[blindspot] Missing API validation-error outcome for step-1**: Journey step E4 explicitly covers "API validation-error on create" but step-1 contract has no API-surface validation-error outcome. The contract only has web-surface validation errors.

4. **[blindspot] Delete only covers planning status**: Step 8 success precondition is "Milestone map is in planning status" but journey invariant says "planning, reviewed, or ready" are all deletable. Quote from step-8 success: `Preconditions: "Milestone map is in planning status; user has milestone:delete permission"`. Should also cover reviewed and ready status deletions.

5. **[blindspot] Cancel only covers executing status**: Step 6 `success-cancelled` precondition: "Milestone map is in executing status and has milestones with associated MainItems" but the journey invariant says "Any non-terminal state can transition to cancelled." The contract restricts cancel to only executing status, which contradicts the journey.

6. **[blindspot] No combined validation outcome**: Step 1 has separate outcomes for missing-name, missing-owner, name-too-long, and invalid-date-range, but no outcome for when multiple validation errors occur simultaneously (e.g., both name and owner empty). A downstream agent would not know how to test the combined case.

7. **[blindspot] Missing wrong-status rejection for transitions**: Steps 3, 4, 5, 7 do not cover what happens when a status transition is attempted on a map in the wrong state (e.g., trying to transition from planning to ready, skipping reviewed). The state machine has a defined flow but no negative test for invalid transitions.

8. **[blindspot] Step 6 cancel precondition is overly specific**: Quote: "Milestone map is in executing status and has milestones with associated MainItems". The "has milestones with associated MainItems" part makes this precondition non-general. What if the executing map has no milestones? What if milestones have no MainItems? Cancel should still work per the business rules but this precondition implies otherwise.

9. **[blindspot] No cancelled-is-terminal outcome**: Step 6 has `completed-is-terminal` but no corresponding `cancelled-is-terminal` outcome. Both are terminal states per the journey invariant but only completed is explicitly verified as terminal.

10. **[blindspot] Step 2 missing server-error and concurrent-modification**: Journey step 2c explicitly covers concurrent modification and step 2d is the only edit-specific edge case covered. No server-error or concurrent-modification outcome exists in step-2 contract.

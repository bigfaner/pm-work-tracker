---
iteration: 1
evaluator: QA-Adversary
date: "2026-06-08"
score: 680
verdict: FAIL — below threshold on 2 dimensions
---

# Journey Eval: item-milestone-binding — Iteration 1

## Reasoning Audit

The journey covers the core item-milestone binding workflow through the MainItem edit dialog. It traces bind, rebind, and unbind flows with reasonable edge cases (terminal state, cross-team, cancelled milestone, server error, session expiry). The primary structural weakness is that the document declares only `surface_types: ["web"]` while the system has a backend API surface that enforces RBAC, validation, and state constraints — no `unauthorized` outcome is tested. Additionally, several preconditions are vague or overlap, making them unexecutable by a downstream test agent. Fact alignment is generally strong with direct quotes from prd-spec Story 11, but two steps introduce behaviors not traceable to any source document.

## Dimension Scores

### 1. Completeness — 140/200

**Metadata (35/50):**
- `feature`, `journey`, `risk_level`, `surface_types`, `surface_keys`, `sources`, `generated` are present.
- `surface_types` is `["web"]` only — backend API is excluded despite the PRD defining RBAC permissions, API validation, and state-machine enforcement that are backend responsibilities. This is a material omission.
- No `surface_type: api` declared, so the `unauthorized` mandatory outcome is absent.

**Steps (60/80):**
- Happy path covers open, bind, rebind, unbind, no-change-save — 5 steps. Good breadth for the core workflow.
- Edge cases cover 9 additional scenarios: empty dropdown, cancelled milestones hidden, terminal MI rejected, cross-team rejected, cancelled milestone rejected, rebind completion recalc, unbind completion recalc, server validation error, session expired.
- Missing: no step for a user who lacks `milestone:update` permission attempting to change the milestone selector. The PRD Story 11 acceptance criteria mention permissions but the journey has no step verifying the permission gate.
- Missing: no step for binding when the milestone itself is in a terminal state (completed milestone — BIZ-milestone-001 says milestone completion requires all items terminal; but what about binding TO a completed milestone?). Step 2d covers cancelled, but not completed.

**Outcomes (45/70):**
- Most outcomes are declarative and verifiable.
- Step 5b outcome ("Filter resets to show all milestones without producing an error") — this filter behavior is not mentioned anywhere in the PRD Story 11 or prd-spec. The PRD does not describe milestone filtering within the edit dialog. This appears to be a fabricated scenario.
- Step 2e outcome mentions "the selected milestone is no longer available" — this is a specific error message not found in the PRD. The PRD only says backend returns errors; the exact message is unspecified.
- Step E1 outcome is adequate: redirect to login, no data modified, original assignment retained after re-auth.

### 2. Semantic Purity — 155/200

**Natural Language (65/80):**
- Steps are generally written in natural, user-centric language: "PM opens the MainItem edit dialog", "PM selects a milestone from the dropdown and saves".
- Step 5b ("PM applies a filter combination that matches no milestones") — "filter combination" is vague. What filter? The edit dialog has a dropdown, not a multi-field filter panel.
- Step 2e precondition references implementation detail ("the backend rejects the save") in the precondition itself.

**Declarative Preconditions (45/60):**
- Preconditions are mostly declarative: "MainItem is not assigned to any milestone", "MainItem is in a terminal state".
- Step 2d precondition is self-contradictory: "A milestone in cancelled status exists (though it should not appear in the dropdown)". If it doesn't appear in the dropdown, how does the PM select it? This precondition is unexecutable — a test agent cannot satisfy it through the UI.
- Step 2c precondition ("The selected milestone belongs to a different team than the MainItem") — how does the PM select a cross-team milestone if the dropdown only shows same-team milestones? The PRD says the dropdown shows milestones in the same team. This step tests an impossible UI action unless the dropdown is unfiltered, which contradicts Step 1c/2d's assertion that cancelled milestones are filtered.

**No Impl Coupling (45/60):**
- Steps reference "milestone_key" in fact annotations — acceptable as annotations, not step text.
- Step 2e references "the backend rejects the save" in the precondition — couples the precondition to backend behavior rather than observable state.
- Step E1 references "session has expired" — acceptable for web surface.
- Step 2d references cancelled status not appearing in the dropdown as a precondition element, which is a UI behavior constraint used as a test precondition.

### 3. Precondition Exclusivity — 95/150

**Distinct (40/60):**
- Steps 2b, 2c, 2d each have unique preconditions (terminal MI, cross-team milestone, cancelled milestone). Good.
- Steps 3b and 4b overlap with Steps 3 and 4 respectively — they add "last non-terminal MI" / "contributes to completion" but the base action is identical. This is acceptable as they test specific recalculations.
- Step 5 and Step 5b share "MainItem is assigned to a milestone" precondition — no distinction in the base state.

**Sufficient (35/50):**
- Missing precondition: no step explicitly states the user's permission level for each step. Step 2c, 2d, 2b should specify what permission the user holds. The Setup section mentions `milestone:update` but doesn't clarify if this applies globally.
- Missing precondition for Step 2e: "the milestone was deleted by another user while the dialog was open" — this is a race condition, not an observable precondition. A test agent cannot reliably set up this state.

**Missing for Errors (20/40):**
- No error path for: user without `milestone:update` permission (only mentioned in Setup, no error step).
- No error path for: binding to a completed milestone (only cancelled is tested).
- No error path for: milestone map being in a terminal state that should block child operations (BIZ-milestone-005: parent-terminal blocks child operations — no step tests what happens when the MilestoneMap is completed/cancelled and user tries to rebind an MI).
- Step 2c and 2d test error paths but their preconditions are arguably unexecutable.

### 4. Fact Alignment — 115/150

**Traceable (50/60):**
- Most steps have `<!-- fact: prd-spec Story 11 -->` annotations that accurately reference prd-spec content.
- Steps 1, 2, 3, 4, 1b, 1c, 2b, 4b trace correctly to prd-spec.
- Step 2b fact annotation references "terminal state MI cannot change milestone assignment" — matches prd-spec Story 11 acceptance criteria.

**Inferred with Rules (35/50):**
- Step 2e is marked `<!-- source: inferred -- derived from Web surface validation-error mandatory outcome -->`. The scenario (milestone deleted by another user) is plausible but not grounded in any PRD statement about concurrent access for the edit dialog.
- Step E1 is marked `<!-- source: inferred -- derived from Web surface session-expired mandatory outcome -->`. Acceptable inference.
- Step 5b ("Milestone filter with no matching results") has NO fact annotation. It describes a filter feature not mentioned in prd-spec Story 11 or anywhere in the PRD. This is a hallucinated scenario.

**No Hallucinated (30/40):**
- Step 5b is a hallucinated scenario — the edit dialog does not have a filter panel in the PRD. Story 11 describes a simple dropdown selector, not a filterable list.
- Step 2e's exact error message ("the selected milestone is no longer available") is fabricated — no PRD source specifies this message.
- Step 2d tests binding to a cancelled milestone that "should not appear in the dropdown" — this is testing a bypass scenario (e.g., direct API call) but the journey is web-surface-only, making it a hallucinated UI scenario.

### 5. Surface Fitness — 110/150

**Mandatory Outcomes (40/60):**
- `validation-error` present: Step 2e covers server validation error. Adequate.
- `session-expired` present: Step E1 covers session expired. Adequate.
- `unauthorized` MISSING: No step tests an unauthenticated or unauthorized user attempting milestone binding. This is a mandatory outcome for the web surface when the system has RBAC permissions. The PRD explicitly defines `milestone:update` permission, but no step tests what happens when the user lacks it.
- The document declares only `surface_types: ["web"]` but the system has both backend API and web surfaces. The `unauthorized` outcome should be covered either via a web step (permission-denied UI state) or by adding an API surface.

**Strategy Proportions (40/50):**
- 14 steps total, all web surface. Reasonable for a web-only journey.
- Happy path: 5 steps. Edge/error: 9 steps. The ratio skews heavily toward edge cases, but for a High-risk journey this is acceptable.

**Realistic Assumptions (30/40):**
- Step 2c assumes a user can select a cross-team milestone from the dropdown — the PRD implies the dropdown filters to same-team milestones, making this step unrealistic for web surface.
- Step 2d assumes a user can select a cancelled milestone from the dropdown — contradicts Step 1c which asserts cancelled milestones are excluded.
- Step 5b assumes a filter mechanism in the edit dialog — no PRD support.

### 6. Internal Consistency — 65/150

**Invariants (30/60):**
- 6 journey invariants declared. Most are well-formed and testable.
- Invariant 4 ("Every bind/rebind/unbind operation triggers completion recalculation on the affected milestone(s)") — good.
- Invariant 5 ("The milestone dropdown never shows cancelled milestones as selectable options") — contradicts Step 2d which tests binding to a cancelled milestone through the dropdown.
- Missing invariant: no invariant about permission gating (user must have `milestone:update` to bind/unbind).
- Missing invariant: no invariant about parent MilestoneMap state blocking child operations (BIZ-milestone-005).

**Cross-Step Refs (20/50):**
- No cross-step references exist. Steps do not reference other steps' outcomes or state.
- Steps 3b and 4b conceptually extend Steps 3 and 4 but are not linked.
- Step 5 (no-change save) should reference Step 2's result state but doesn't.

**Risk Level (15/40):**
- Risk level is "High" which is appropriate for state-mutation operations.
- However, the journey does not test the highest-risk scenarios: concurrent edits (two PMs editing same MI's milestone simultaneously), or cascading effects from MilestoneMap cancellation (BIZ-milestone-006: MilestoneMap cancel cascades to all children).
- BIZ-milestone-005 (parent-terminal blocks child operations) is completely untested — what happens when the MilestoneMap is in a terminal state?

## Blindspot Hunt

1. [blindspot] **No permission-denied step**: The Setup mentions `milestone:update` permission but no step tests the behavior when the user lacks this permission. The PRD Story 11 implicitly assumes the user has permission. A user without `milestone:update` should not be able to change the milestone dropdown or should see it disabled. No step verifies this.

2. [blindspot] **Parent-terminal constraint untested**: BIZ-milestone-005 states "Parent-terminal blocks child operations." If the MilestoneMap is completed or cancelled, binding/rebinding/unbinding MIs within its milestones should be blocked. No step tests this constraint.

3. [blindspot] **MilestoneMap cancellation cascade**: BIZ-milestone-006 states "MilestoneMap cancel cascades to all children." When a MilestoneMap is cancelled, all its milestones are cancelled and MIs are unbound. The journey should verify that after such a cascade, the MI's milestone assignment is correctly cleared and the edit dialog reflects this.

4. [blindspot] **Step 2d is unexecutable on web surface**: The precondition says "A milestone in cancelled status exists (though it should not appear in the dropdown)" but the user action is "PM attempts to select and bind to a cancelled milestone." If it doesn't appear in the dropdown, the PM cannot select it through the UI. This step requires either: (a) changing the surface to api to test a direct API call, or (b) removing the step as untestable on web.

5. [blindspot] **Step 2c may be unexecutable on web surface**: Similar to 2d — if the dropdown only shows same-team milestones, the user cannot select a cross-team milestone through the UI. The PRD Story 11 says "MI with target milestone belong to different team" triggers rejection, but doesn't specify how the cross-team milestone would be presented in the dropdown.

6. [blindspot] **No step for binding to a completed milestone**: Step 2d tests cancelled milestones but BIZ-milestone-001 states "Milestone completion requires all related items terminal." What happens when a user tries to bind a new MI to an already-completed milestone? This should be blocked but is untested.

7. [blindspot] **Step 5b is fabricated**: "Milestone filter with no matching results" references a filter mechanism that does not exist in prd-spec Story 11. The edit dialog has a simple dropdown selector, not a filter panel with filter criteria.

8. [blindspot] **No concurrent/race condition test**: The journey has a server validation error step (2e) but does not test the more common race condition where two users edit the same MI's milestone simultaneously. The PRD mentions conflict detection for milestone editing (Story 6) but not for item-milestone binding.

9. [blindspot] **Missing invariant for team consistency at journey level**: While Step 2c tests cross-team rejection, there is no invariant stating that the dropdown is scoped to the MI's team. Invariant 5 mentions cancelled milestones are excluded but nothing about team scoping.

10. [blindspot] **Completion recalculation not verified numerically**: Steps 2, 3, 3b, 4, 4b all mention completion recalculation but none specify expected values or thresholds. A test agent cannot verify "recalculated" without knowing what the expected value should be.

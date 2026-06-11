---
iteration: 2
score: 812
target: 850
verdict: FAIL
dimensions:
  completeness: 148
  semantic_purity: 180
  precondition_exclusivity: 124
  fact_alignment: 118
  surface_fitness: 124
  internal_consistency: 118
---

# Journey Eval: milestone-item-management (Iteration 2)

**Verdict: FAIL** -- 812/1000 (target: 850)

## Dimension Breakdown

| Dimension | Score | Min | Status |
|-----------|-------|-----|--------|
| Completeness | 148/200 | 120 | PASS |
| Semantic Purity | 180/200 | 120 | PASS |
| Precondition Exclusivity | 124/150 | 90 | PASS |
| Fact Alignment | 118/150 | 90 | PASS |
| Surface Fitness | 124/150 | 90 | PASS |
| Internal Consistency | 118/150 | 90 | FAIL |

## Scoring Rationale

### Completeness: 148/200

**Metadata (42/50):** Name "milestone-item-management" is valid kebab-case. Risk_level "High" is justified by state mutation operations (unbind, rebind, cancel auto-unbind). surface_types lists "web" correctly. Sources list two PRD files. Gap: sources omits prd-ui-functions.md, which is the primary reference for the detail panel UI spec referenced in Story 10. Setup section lists permissions but does not state the specific team/team-name required for cross-team tests in Step 6b, making it ambiguous whether the default setup covers cross-team scenarios.

**Steps (62/80):** 22 steps total (7 happy path + 15 edge cases). Each step has a name, user action, and expected result. Ordered sequence is coherent. Gaps:
- No network-error/server-error boundary steps for any mutation operation (unbind Step 3, quick-add Step 4, drag-rebind Step 6). The PRD Story 10 acceptance criteria explicitly cover 500 errors for quick-add (implied by Story 5's pattern and Story 9's error handling), and the previous iteration flagged this same gap. Still unaddressed.
- Step 5 (Navigate to MI detail) is a single-sentence step with no edge cases at all: no coverage for MI deleted between panel load and click, no 404 scenario, no loading state for navigation.
- No step covers undo expiry: Step 3b tests undo within window, but no step tests what happens when the undo window expires -- does the toast disappear? Can the user still see the MI elsewhere?
- No step covers drag-and-drop cancellation (user starts dragging but drops outside any valid target).

**Outcomes (44/70):** validation-error covered (Steps 4b-4d for quick-add form). session-expired covered (Step E1). loading-state covered (Steps 1b, 4e). Missing:
- network-error/server-error for mutation APIs (3 missing: unbind, create, rebind). This was flagged in iteration 1 and remains unfixed.
- conflict-error (concurrent edit): PRD Story 6 mentions concurrent editing conflict ("data has been modified by another person, please refresh"), but no step tests this for the milestone edit within the panel.
- No permission-denied server response (403) boundary: Step 1f covers UI hiding of controls, but if permissions change server-side during the session, there is no step testing a 403 response to an unbind/add/rebind API call.

### Semantic Purity: 180/200

**Natural language (76/80):** All expected results describe user-observable behavior in natural language. No regex, CSS selectors, XPath, or framework assertions. Minor deduction: Step 4f uses "cannot be modified" which is acceptable. "Danger zone" in Step 1 and 7b is a user-visible section label per PRD Story 10 ("hazardous operations area"). Clean overall.

**Declarative preconditions (55/60):** All preconditions describe states, not setup procedures. Step 3b's "PM has just unbound a MainItem and the undo option is visible" is borderline -- it references a prior action result rather than a pure state description, but it is the resulting state. Step E1 "session has expired while the panel is open" is acceptable as a state description. Deduction: Step 4b-4d share identical preconditions ("Quick-add dialog is open") and only differ by action -- the precondition alone does not uniquely distinguish them; the action is the differentiator.

**No implementation coupling (49/60):** Steps use user-level verbs (click, hover, drag, press). However, HTML comments contain implementation-level references: Step 3 "<!-- fact: prd-spec -- unbind sets milestone_key to null -->" references the database field milestone_key. Step 6 "<!-- fact: prd-spec -- drag-drop updates milestone_key -->" does the same. While these are in comments and not step narrative, they embed implementation knowledge. Step 5b's expected result "MIs auto-unbound on cancel" references a business rule rather than user-observable behavior -- the user sees an empty list, not "auto-unbinding." This is a minor coupling of cause (auto-unbind mechanism) into the expected result.

### Precondition Exclusivity: 124/150

**Distinct preconditions (50/60):** Edge cases are mostly properly separated. Steps 4b, 4c, 4d share the exact same precondition text ("Quick-add dialog is open") and differ only by the user action. While this is technically valid (same state, different action), it means the precondition alone does not uniquely identify the step -- the action is required for disambiguation. Steps 1c and 1d both have "Panel is open" but differ by action (hover vs press).

**Sufficient for unique selection (42/50):** Most steps are uniquely identifiable. However:
- Steps 4b, 4c, 4d: precondition "Quick-add dialog is open" is identical for all three. The action is the differentiator, but the precondition dimension itself is not distinct.
- Step 1b ("Panel opens but data is still loading") and Step 1 ("PM clicks a milestone node") -- Step 1b's precondition "Panel opens but data is still loading" implies Step 1's action has already been taken, creating an implicit sequential dependency that is not declared as a precondition referencing Step 1.

**Missing preconditions for error paths (32/40):** Error paths generally state their triggers. Gaps:
- Step 6b (cross-team drag): Precondition states "target milestone belongs to a different team" but does not specify that the MI belongs to team A. It should state "MI belongs to Team A, target milestone belongs to Team B (different team)."
- Step 6c (terminal MI): Precondition states "MI is in terminal state" but does not specify the MI's current milestone binding state (is it bound to a milestone? not bound?).
- Step 6d (cancelled target milestone): Precondition states "target milestone is cancelled" but does not specify the source milestone state or MI state.
- Step E1: Precondition states "session has expired while the panel is open" but does not specify which mutation action is attempted. It says "attempts an action (unbind, add, drag-rebind)" -- listing three possible actions makes it ambiguous which one triggers the session expiry response.

### Fact Alignment: 118/150

**Factual claims traceable (45/60):** Key steps have fact annotations: Steps 1, 3, 4, 5b, 6, 6b, 6c, 6d, 7b have `<!-- fact: -->` annotations. Gaps:
- Steps 1b, 1c, 1d, 1e, 1f, 2, 4e, 4f, 5 lack fact annotations despite being directly derived from PRD Story 10 acceptance criteria. This was flagged in iteration 1 and remains partially unfixed.
- Step 2 (hover unbind): PRD Story 10 states "When I hover over a certain MI, a x unbind button appears on the right side of the row." The step has no fact annotation.
- Step 5 (navigate to MI detail): PRD Story 10 states "When I click the number or title of a certain MI, jump to the /items/:mainItemId main item detail page." No fact annotation.
- Step 7 (close via overlay): PRD Story 10 states "When I click the overlay area outside the panel, the panel closes." No fact annotation despite being directly from Story 10.

**Inferred claims with rules (40/50):** Steps 4b-4d have proper `source: inferred` annotations citing validation-error mandatory outcome. Step E1 has `source: inferred` for session-expired. However:
- Step E1's expected result claims "After re-authenticating, the panel retains its original state." This is not a standard session-expired behavior and is not derived from any source document. The session-expired rule says "unsaved data is either preserved or user is warned about data loss" -- this is about data, not panel state. Panel state retention after re-auth is speculative and should be marked UNKNOWN.
- Step 3's undo behavior ("undo option appears briefly") -- the PRD Story 10 says "undo toast (5s)" but the journey does not cite this source and says "briefly" instead of the specific 5-second window from the PRD.

**No hallucinated claims (33/40):** Most claims are grounded. Concerns:
- Step E1's "panel retains its original state" after re-authentication is an unverified claim that goes beyond the session-expired rule.
- Step 3 says "undo option appears briefly" -- the PRD specifies 5 seconds. "Briefly" is vaguer but not incorrect.
- Step 6 says "Visual feedback is shown during drag" -- this is generic enough to be safe but does not match the PRD's specificity ("drag shows opacity-50, target milestone highlights").

### Surface Fitness: 124/150

**Mandatory outcomes (50/60):** validation-error is present (Steps 4b-4d for quick-add form). session-expired is present (Step E1). Both are mandatory for Web surface. Gap: validation only covers the quick-add dialog. The journey does not cover validation for the milestone edit dialog within the panel (editing name/description/dates), which is another form that should have validation-error coverage. The PRD Story 6 covers milestone editing and has validation for name length (1-100 chars) and date constraints.

**Test strategy proportions (38/50):** Approximately 60/40 contract/journey split, overweight on contract. Steps 4b, 4c, 4d are three separate steps testing the same form's required-field validation from different angles. Steps 1d and 1e test two close methods that are functionally identical (keyboard close vs button close). The Web strategy recommends 50/50 balance. The journey-level depth is thin: drag-and-drop has 4 edge cases but they are all rejection scenarios; no positive journey for drag-drop with undo, no drag-drop across visible milestones, no drag-drop where source milestone recalculation is verified by the user.

**Realistic assumptions (36/40):** All browser interactions (click, hover, drag, keyboard) are Playwright-compatible. No DOM structure or CSS class assertions. Focus management assertions are realistic. Minor concern: Step 6's drag-and-drop between milestone nodes on a timeline is complex to automate but is a valid Playwright scenario. Step 3's undo within "brief window" is time-sensitive and may be flaky in automation, though this is a testability concern, not a journey validity issue.

### Internal Consistency: 118/150

**Invariants hold (34/60):** The journey defines 5 invariants. Verification status:
- Invariant 1 ("Unbind operations show undo option"): Verified by Steps 3 and 3b. PASS.
- Invariant 2 ("Quick-add pre-fills and locks milestone field"): Verified by Step 4 and 4f. PASS.
- Invariant 3 ("Completion percentage recalculated"): Claimed in Steps 3, 4, and 6 expected results. However, no step actually VERIFIES the recalculated value -- the expected result just says "recalculated" without stating what the new value should be. This is a weak verification: the invariant is stated but never concretely checked.
- Invariant 4 ("Cancelled milestones display muted, empty list, no add, visible delete"): Verified by Step 5b. PASS.
- Invariant 5 ("All mutation operations require milestone:update permission; controls hidden"): PARTIALLY VERIFIED. Step 1f checks edit control and status badge. But the invariant claims "unbind, add, drag-rebind" controls are hidden -- these are NEVER tested under no-permission precondition. Steps 2 (unbind), 4 (add), and 6 (drag-rebind) all assume permission exists. This is a significant gap: the invariant makes a sweeping claim that is only verified for 2 of 5 controls (edit button, status badge).

**Cross-step references (44/50):** Steps reference each other implicitly through preconditions (e.g., Step 3b references Step 3's outcome). No dangling references. Minor issue: Step 5b's precondition "Milestone is in cancelled status" does not reference the cancellation action (from another journey or business rule), leaving the setup ambiguous for the test executor.

**Risk level consistent (40/40):** "High" is justified by irreversible operations (cancel auto-unbinds all MIs, delete removes milestones), state mutations, and data loss potential. Consistent with the step content.

## Attacks

1. **[Internal Consistency] Invariant 5 grossly under-verified:** Quote: "All mutation operations (unbind, add, drag-rebind) require milestone:update permission; edit/delete controls are hidden without permission." Step 1f only checks "Edit control is not displayed. Status badge is not interactive." The unbind button (Step 2), add button (Step 4), drag-rebind interaction (Step 6), and delete button (Step 7b) are NEVER tested under the no-permission precondition. The invariant makes a claim about 5 controls but only 2 are verified. Must add no-permission edge cases for Steps 2, 4, 6, and 7b, or expand Step 1f to cover all mutation controls.

2. **[Completeness] No network/server error boundaries for any mutation:** Steps 3 (unbind), 4 (quick-add), and 6 (drag-rebind) all make API calls. None have error-boundary variants. PRD Story 10 and Story 5 both specify 500-error handling ("display 'creation failed, please retry', dialog does not close, form retains data"). The journey assumes all API calls succeed. Must add steps for: unbind API failure (Step 3 error variant), create MI API failure (Step 4 error variant), drag-rebind API failure (Step 6 error variant).

3. **[Fact Alignment] Step E1 session recovery claim is hallucinated:** Quote: "After re-authenticating, the panel retains its original state." No source document in PRD describes panel state retention after session expiry and re-authentication. The session-expired surface rule says "unsaved data is either preserved or user is warned about data loss." Panel state is not unsaved data. This claim must be marked UNKNOWN or removed.

4. **[Completeness] Invariant 3 never concretely verified:** Quote from Invariant 3: "Completion percentage is recalculated whenever MIs are bound, unbound, or rebound." Steps 3, 4, 6 all say "recalculated" in expected results but none state the expected numerical value. A test executor cannot verify "recalculated" without knowing what the new percentage should be. Example fix: "Completion percentage changes from X% to Y%."

5. **[Fact Alignment] Missing fact annotations on 9 steps derived from PRD Story 10:** Steps 1b, 1c, 1d, 1e, 1f, 2, 4e, 4f, 5 all describe behavior directly from PRD Story 10 acceptance criteria but lack `<!-- fact: prd-spec Story 10 -->` annotations. This was flagged in iteration 1 and remains unfixed. Quote from Step 2: "PM hovers over a MainItem row in the associated items list." -- directly from Story 10 "When I hover over a certain MI, a x unbind button appears on the right side."

6. **[Precondition Exclusivity] Steps 4b/4c/4d share identical preconditions:** All three have precondition "Quick-add dialog is open" with only the user action varying. The precondition dimension does not uniquely identify these steps. Best practice: each step should have a precondition that, combined with the step identity, is distinct. Consider combining into a single step with multiple validation outcomes, or enriching preconditions (e.g., "Quick-add dialog is open, title field is empty" for 4b).

7. **[Completeness] Step 5 lacks any edge cases:** Quote: "**Expected Result**: Route navigates to the main item detail page." Single sentence, no edge cases. Missing: MI deleted between panel load and click, MI detail page returning 404, navigation loading state, back-navigation behavior (does the panel re-open?).

8. **[Precondition Exclusivity] Step E1 ambiguous trigger:** Quote: "PM attempts an action (unbind, add, drag-rebind) in the panel." Three different actions are listed. Each would trigger a different API call and potentially a different error response. The step should either pick one representative action or be split into three steps (one per mutation type) for precision.

9. **[Fact Alignment] Step 6 visual feedback is vague vs PRD:** Quote: "Visual feedback is shown during drag." The PRD Story 9 acceptance criteria specifies: "drag shows opacity-50, target milestone highlights." The journey's "Visual feedback is shown" is technically correct but loses the specificity of the PRD. If the fact annotation is meant for traceability, the expected result should match the PRD's specificity level.

10. **[Completeness] No undo-expiry edge case:** Step 3b tests undo within the window, but no step tests what happens when the undo window expires. Does the toast disappear? Is the MI permanently unbound? What if the user tries to interact with the expired toast? The PRD specifies 5 seconds for the undo toast, so expiry behavior should be covered.

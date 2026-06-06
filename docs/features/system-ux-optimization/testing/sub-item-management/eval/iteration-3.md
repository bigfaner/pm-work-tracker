---
iteration: 3
scorer: adversary
total_score: 885
pass: true
date: "2026-06-04"
---

# Journey Eval: sub-item-management (Iteration 3)

## Dimension Scores

| Dimension | Score | Threshold | Pass |
|-----------|-------|-----------|------|
| 1. Completeness | 175/200 | 120 | PASS |
| 2. Semantic Purity | 175/200 | 120 | PASS |
| 3. Precondition Exclusivity | 130/150 | 90 | PASS |
| 4. Fact Alignment | 135/150 | 90 | PASS |
| 5. Surface Fitness | 135/150 | 90 | PASS |
| 6. Internal Consistency | 135/150 | 90 | PASS |

**Total: 885/1000** | Target: 850 | **PASS**

---

## Dimension 1: Completeness (175/200)

### Journey Metadata (45/50)

- Name `sub-item-management` is valid kebab-case.
- Risk level "Medium" with explicit justification comment: "Workflow involves data mutation (start time edit) without destructive operations. Sorting is read-only. No security-sensitive operations." Justification is clear and consistent with the content.
- Surface types `["web", "api"]` are declared and match the body content (web: Steps 1-4, E1-E4, E9; API: Step 5, E5-E8).
- Sources cite three documents: prd-user-stories.md (Story 2, Story 5), prd-spec.md (#2, #5), api-handbook.md (Enhanced List Main Items). All relevant to the journey scope.

**Deduction (-5)**: Sources still cite api-handbook.md for "Enhanced List Main Items" while the journey scope is intentionally limited to Stories 2 and 5. The citation is valid for Step 5 but slightly over-broad as a source reference. Minor.

### Steps Complete (75/80)

- Happy path Steps 1-4 cover Story 2 (sub-item start time edit: open dialog, modify, save, verify list position) and Story 5 (sub-item list sorted by creation time descending). Step 3 verifies the invariant that editing start time does not change list position.
- Step 5 adds API-surface coverage for list sub-items sorted by creation time.
- Edge cases now cover 9 scenarios: invalid date format (E1), start time after end time (E1b), single item (E2), tied timestamps (E3), session expired (E4), unauthorized (E5), unauthenticated (E6), malformed request (E7), not-found API (E8), concurrent deletion (E9).
- E1b addresses the PRD-defined "start time must not be later than end time" validation rule from prd-ui-functions.md line 115. This directly resolves iteration 2 critical issue #1.
- E9 addresses the web-surface not-found boundary case (concurrent deletion during edit). This resolves iteration 2 critical issue #4.
- The Overview note now explicitly states: "sub-item start time update has no standalone API endpoint; the mutation is handled via the web form only." This resolves iteration 2 critical issue #2.
- E7 now specifies "non-numeric value for a team ID parameter (e.g., `abc`)" as concrete malformed input. This resolves iteration 2 critical issue #3.

**Deduction (-5)**: No step covers the PRD-defined positive boundary case where start time is set to a past date (prd-ui-functions.md line 116: "start time allows past dates"). This is a lower-priority gap since it confirms allowed behavior rather than testing a failure mode, but it is still a PRD-grounded scenario.

### Outcomes Coverage (55/70)

- Mandatory derived outcomes for web: E1 covers `validation-error`, E4 covers `session-expired`. Both present.
- Mandatory derived outcomes for API: E5 covers `unauthorized` (authenticated without permission), E6 covers `unauthorized` (unauthenticated). Both present.
- Boundary outcomes: E1 (web validation-error), E1b (web validation-error for business rule), E7 (API validation-error), E8 (API not-found), E9 (web not-found).
- E9 now covers the web-surface not-found boundary outcome that was missing in iteration 2.

**Deduction (-15)**: No explicit boundary outcome for empty-state behavior (zero sub-items on a main item). The journey tests single sub-item (E2) but never tests the zero-sub-item case. Also, no API boundary outcome for an empty response (main item with no sub-items). These are lower-priority gaps.

---

## Dimension 2: Semantic Purity (175/200)

### Natural Language Outcomes (70/80)

- Most expected results use user-observable language well:
  - "Edit dialog renders with all existing fields populated, including the '开始时间' field showing the current start time value" (Step 1) -- good.
  - "Start time is updated; save succeeds; dialog closes; sub-item detail reflects the new start time" (Step 2) -- good.
  - "A validation error message is displayed indicating that start time must not be later than end time; the change is not saved" (E1b) -- good, user-observable.
  - "An error message is displayed indicating the sub-item no longer exists; the edit dialog closes; no data is modified" (E9) -- good, user-observable.
- API outcomes appropriately use technical language: "The API returns a validation error response listing the invalid parameter; no data is returned" (E7) -- good.
- E3 "Sub-items with identical creation times are displayed in a stable, deterministic order" -- good, user-observable.

**Deduction (-10)**: E7 "The API returns a validation error response listing the invalid parameter" is slightly implementation-leaning ("listing the invalid parameter" describes the response structure). Could be rephrased as "The API rejects the request and indicates the team ID is invalid." Also, some outcomes mix user-observable with system-level descriptions (e.g., "no data is modified" in E9 is a system-level claim rather than a user observation).

### Declarative Preconditions (55/60)

- Most preconditions are declarative state declarations:
  - "A sub-item exists with a start time value set" (Step 1) -- declarative.
  - "A main item has multiple sub-items created on different dates" (Step 4) -- declarative.
  - "A sub-item has an end time set to 2026-05-01" (E1b) -- declarative, concrete.
  - "Another user has deleted the sub-item while the current user has the edit dialog open" (E9) -- declarative state description.
- E1 precondition "The sub-item edit dialog is open" is slightly procedural (implies a prior action) but is a common pattern in journey preconditions.
- Setup is declarative: "A team exists with a main item that has multiple sub-items created at different times" and "PM user is logged in with sub_item:update and main_item:read permissions."

**Deduction (-5)**: A few preconditions (E1, E4) are slightly procedural. E9's "while the current user has the edit dialog open" is borderline but acceptable as a concurrent state description.

### No Implementation Coupling (50/60)

- No HTTP status codes in web steps -- good.
- No CSS/XPath selectors -- good.
- No component framework names leaked -- good.
- API outcomes use semantic error descriptions: "authorization error" (E5), "authentication error" (E6), "validation error response" (E7), "not found error" (E8) -- all appropriate for API surface without being raw HTTP codes.
- E1b "A validation error message is displayed indicating that start time must not be later than end time" -- good, no implementation coupling.
- E9 "An error message is displayed indicating the sub-item no longer exists" -- good.

**Deduction (-10)**: E7 outcome "The API returns a validation error response listing the invalid parameter" describes response structure in a way that is borderline implementation-coupled. For API surface, referencing error types is acceptable, but "listing the invalid parameter" presumes a specific response format. Also, Step 5 outcome "The response includes sub-items under each main item, ordered by creation time descending" presumes a nested response structure, which couples to the API design rather than describing observable behavior.

---

## Dimension 3: Precondition Exclusivity (130/150)

### Preconditions Distinct Across Outcomes (55/60)

- Happy path preconditions are distinct:
  - Step 1: sub-item with start time set
  - Step 2: edit dialog open with start time field
  - Step 3: start time just edited in Step 2
  - Step 4: main item with multiple sub-items on different dates
  - Step 5: authenticated API request for main item with multiple sub-items
- Edge case preconditions are distinct:
  - E1: edit dialog open (for invalid date format)
  - E1b: sub-item with end time set to 2026-05-01 (for start-after-end validation)
  - E2: main item with exactly one sub-item
  - E3: multiple sub-items created simultaneously
  - E4: previously authenticated, session expired
  - E5: authenticated user without main_item:read permission
  - E6: no valid credentials
  - E7: non-numeric team ID parameter
  - E8: non-existent main item ID
  - E9: sub-item deleted by another user while edit dialog open
- E1 vs E1b: E1 tests invalid date format; E1b tests start-after-end business rule. These are distinct validation scenarios. Good.
- E9 vs E8: E8 tests API not-found (non-existent main item); E9 tests web not-found (concurrently deleted sub-item). Distinct surfaces and triggers. Good.

**Deduction (-5)**: Step 1 and Step 2 preconditions are sequential (Step 2 depends on Step 1's outcome), which is inherent to journey flow but creates overlap if treated as independent outcomes.

### Preconditions Sufficient to Uniquely Select Outcome (45/50)

- All preconditions uniquely determine their outcomes:
  - E1b: "sub-item has end time set to 2026-05-01" + "user sets start time to 2026-06-01" uniquely leads to start-after-end validation error. Concrete and unambiguous.
  - E7: "non-numeric value for team ID parameter (e.g., `abc`)" uniquely leads to validation error. Concrete. Improved from iteration 2.
  - E9: "Another user has deleted the sub-item while the current user has the edit dialog open" uniquely leads to not-found error on save. Concrete.
- Setup now includes "sub_item:update and main_item:read permissions" which resolves the iteration 2 issue where Step 5's permission was unclear.

**Deduction (-5)**: Step 5 precondition "An authenticated API request is sent for a main item that has multiple sub-items" does not explicitly state the user has main_item:read permission (though setup now declares it). The step-level precondition could be more self-contained.

### No Missing Preconditions for Error/Boundary (30/40)

- E1b precondition is concrete: "A sub-item has an end time set to 2026-05-01" and user sets start time to 2026-06-01. All necessary state is declared. Good.
- E7 precondition is concrete: "An API request uses a non-numeric value for a team ID parameter." Improved from iteration 2.
- E9 precondition is concrete: "Another user has deleted the sub-item while the current user has the edit dialog open." Full state described.
- Setup now includes "At least one sub-item has both a start time and an end time set" which establishes the necessary state for E1b. Good.

**Deduction (-10)**: E4 (session expired) does not specify how the session expires or what constitutes "expired" in this context (token timeout, server-side session invalidation). E5 (unauthorized) does not specify what permissions the user does have -- only what they lack. Both are functional but could be more complete.

---

## Dimension 4: Fact Alignment (135/150)

### Factual Claims Traceable (50/60)

- Fact annotations are present and traceable:
  - Step 1: `<!-- fact: prd-spec #2, Story 2 AC1 -->` -- traceable and correct.
  - Step 2: `<!-- fact: Story 2 AC1 -->` -- traceable and correct.
  - Step 3: `<!-- fact: prd-spec #5 -- sort order is by creation time, not start time -->` -- traceable and correct.
  - Step 4: `<!-- fact: Story 5 AC1 -->` -- traceable and correct.
  - Step 5: `<!-- fact: api-handbook Enhanced List Main Items -->` -- traceable. The api-handbook does document the main items endpoint with sub-items in the response, though sub-item sort order within the response is more precisely documented in prd-ui-functions.md UF-5 ("ORDER BY id DESC").
  - E1b: `<!-- fact: prd-ui-functions.md line 115 -- start time must not be later than end time -->` -- traceable and correct. Line 115 of prd-ui-functions.md states exactly this rule. Good.

**Deduction (-10)**: Step 5's fact annotation points to api-handbook but the sort order claim is more precisely grounded in prd-ui-functions.md UF-5. The api-handbook "Enhanced List Main Items" section documents the endpoint but does not explicitly state sub-item sort order. This is a minor misattribution carried from iteration 2.

### Inferred Claims Have Source (45/50)

- All inferred edge cases are explicitly marked with `source: inferred` and include reasoning:
  - E1: `derived from Web surface validation-error mandatory outcome` -- properly marked.
  - E1b: This is NOT marked as inferred -- it has a `fact:` annotation pointing to prd-ui-functions.md line 115. This is correct because the validation rule is explicitly defined in the PRD.
  - E4: `derived from Web surface session-expired mandatory outcome` -- properly marked.
  - E5/E6: `derived from API surface unauthorized mandatory outcome` -- properly marked.
  - E7: `derived from API surface validation-error outcome` -- properly marked.
  - E8: `derived from API surface not-found common boundary outcome` -- properly marked.
  - E9: `derived from Web surface not-found boundary outcome` -- properly marked.
  - E3: `inferred -- tiebreaker behavior when creation times match` -- properly marked.
  - E2: `fact: Story 5 -- sorting applies regardless of item count` -- traceable.

**Deduction (-5)**: E3's inference reasoning ("tiebreaker behavior when creation times match") is sound but the PRD does not mention tiebreaker behavior at all. This is a pure inference based on a gap in the specification. The annotation is honest but the inference could note that no PRD source addresses this case.

### No Hallucinated Claims (40/40)

- No hallucinated content detected.
- E1b's validation rule is directly from prd-ui-functions.md line 115. Not hallucinated.
- E9's concurrent deletion scenario is a standard web boundary case. Not hallucinated.
- E7's non-numeric team ID is a reasonable concrete example of malformed input. Not hallucinated.
- E3's "stable, deterministic order" is a safe, non-committal description that does not fabricate specific behavior. Not hallucinated.
- The Overview note about no standalone API endpoint is an accurate architectural observation. Not hallucinated.

**No deduction.** The journey is free of hallucinated claims.

---

## Dimension 5: Surface Fitness (135/150)

### Mandatory Derived Outcomes Present (60/60)

- Web surface:
  - `validation-error`: E1 (invalid date format) and E1b (start after end time) cover this. Present.
  - `session-expired`: E4 covers this. Present.
- API surface:
  - `unauthorized`: E5 (authenticated without permission) and E6 (unauthenticated) cover this. Present.
- All mandatory derived outcomes are now covered for both declared surface types.

**No deduction.** Full coverage of mandatory outcomes.

### Test Strategy Proportions (40/50)

- Web steps: Steps 1-4, E1, E1b, E2, E3, E4, E9 = 10 steps
- API steps: Step 5, E5, E6, E7, E8 = 5 steps
- Ratio: approximately 67% web / 33% API.
- The Overview now explicitly acknowledges that sub-item start time update has no standalone API endpoint, explaining why the write path is web-only. This addresses iteration 2's critical issue #2.
- The imbalance is justified by the architectural constraint (no API mutation endpoint for sub-item updates). The read-path API coverage (Step 5, E5-E8) is appropriate.

**Deduction (-10)**: The API coverage is read-only and could include one more read-path scenario (e.g., API request for a main item with zero sub-items, or API request with pagination parameters). The 67/33 split is justified but tilts toward web. A 60/40 split would be more balanced.

### Surface-Specific Environment (35/40)

- Web environment: realistic assumptions about browser, dialog, list page, date picker, concurrent editing. Good.
- API environment: Step 5 references "main items endpoint for a team" which is concrete. E5-E8 reference authentication states and request parameters.
- E7 now specifies "non-numeric team ID (e.g., `abc`)" which gives a concrete test input. Good.
- Setup includes "sub_item:update and main_item:read permissions" which establishes the API auth context. Good.

**Deduction (-5)**: The API environment still does not specify the authentication mechanism (e.g., JWT Bearer token) or request format. Step 5 says "An authenticated API request is sent to the main items endpoint for a team" without specifying GET method or query parameters. These are minor but would strengthen the environment specification.

---

## Dimension 6: Internal Consistency (135/150)

### Invariants Hold in Every Step (55/60)

Four invariants declared:

1. "The sub-item edit dialog always includes a '开始时间' field that can be modified and saved" -- Steps 1, 2, E1, E1b, E9 all operate on the edit dialog with a start time field. E4 also involves the edit dialog. Consistent.

2. "Sub-item lists on main item detail pages are always sorted by creation time in descending order (newest first)" -- Steps 3, 4 verify this. E2 and E3 also operate on sorted lists. Step 5 verifies API equivalent. Consistent.

3. "Sub-item creation time determines sort order, not start time or update time" -- Step 3 explicitly verifies this. Consistent.

4. "Editing start time does not affect the sub-item's position in the sorted list" -- Step 3 verifies this. Consistent.

E1b introduces a validation failure (start after end), so the save does not succeed. This does not violate any invariant because the invariant says the field "can be modified and saved" only for valid inputs. E1b tests an invalid input. Consistent.

E9 introduces a concurrent deletion. The edit dialog was open before deletion, so the invariant about the dialog including a start time field held at the time the dialog was opened. The deletion happens externally. Consistent.

**Deduction (-5)**: E4 says "after re-authenticating, the sub-item retains its original values" which is a claim not directly verified by any step and not covered by the declared invariants. This is a minor consistency gap -- the claim is reasonable but unsupported.

### Cross-Step References Consistent (45/50)

- Step 3 references "Step 2" in its precondition: "A sub-item's start time was just edited in Step 2" -- clear and correct.
- Setup establishes the context (team, main item, sub-items, PM user with permissions, sub-item with end time) and all steps reference entities within this context.
- E1b's precondition ("A sub-item has an end time set to 2026-05-01") aligns with the setup ("At least one sub-item has both a start time and an end time set"). Consistent.
- E9 references "the current user" and "another user" which implies a multi-user context consistent with the setup.
- No dangling references detected.

**Deduction (-5)**: E1b uses a specific date (2026-05-01) that is not established in setup. Setup says "At least one sub-item has both a start time and an end time set" but does not specify the end time value. The specific date in E1b's precondition is self-contained, which is fine, but there is a slight disconnect between the general setup and the specific precondition value.

### Risk Level Consistent with Content (35/40)

- "Medium" risk is appropriate: data mutation (start time edit), read operations (sorting), no destructive operations, standard auth checks.
- The dual-surface coverage (web + API) adds moderate complexity, consistent with Medium.
- The 14 steps (5 happy path + 9 edge cases) represent moderate complexity. Medium is appropriate.
- E9 adds a concurrent modification scenario which slightly increases complexity, but not enough to warrant High.

**Deduction (-5)**: The concurrent deletion scenario (E9) and the multi-user context it implies could be noted in the risk justification. The current justification says "No security-sensitive operations" which is correct, but concurrent operations add a dimension of complexity not reflected in the justification.

---

## Critical Issues

No critical issues remain. All four critical issues from iteration 2 have been addressed:

1. **RESOLVED**: E1b now covers the PRD-defined "start time after end time" validation rule (prd-ui-functions.md line 115).
2. **RESOLVED**: Overview note explains that sub-item start time update has no standalone API endpoint.
3. **RESOLVED**: E7 now specifies "non-numeric team ID (e.g., `abc`)" as concrete malformed input.
4. **RESOLVED**: E9 covers concurrent deletion during edit for the web surface.

### Minor Issues (Non-blocking)

1. Step 5's fact annotation could point to prd-ui-functions.md UF-5 for sort order rather than solely to api-handbook.
2. E4's re-authentication claim ("sub-item retains its original values") is unsupported by any step or invariant.
3. E1b's specific date (2026-05-01) is not echoed in the setup section.
4. No test for the PRD-defined "start time allows past dates" positive boundary case.

## Pass/Fail Status

| Check | Result |
|-------|--------|
| Dimension 1 >= 120 | PASS (175) |
| Dimension 2 >= 120 | PASS (175) |
| Dimension 3 >= 90 | PASS (130) |
| Dimension 4 >= 90 | PASS (135) |
| Dimension 5 >= 90 | PASS (135) |
| Dimension 6 >= 90 | PASS (135) |
| Total >= 850 | PASS (885) |
| **Overall** | **PASS** |

All 6 dimensions pass their individual thresholds. Total score 885/1000 exceeds the 850 target by 35 points. The journey has improved from 810 (iteration 2) to 885 (iteration 3), a +75 point gain driven primarily by the addition of E1b (+15 Completeness, +10 Fact Alignment), E9 (+10 Completeness, +10 Surface Fitness), the Overview architectural note (+5 Surface Fitness), and E7 concretization (+5 Precondition Exclusivity). All 4 critical issues from iteration 2 are resolved.

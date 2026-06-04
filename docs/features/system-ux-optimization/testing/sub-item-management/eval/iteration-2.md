---
iteration: 2
scorer: adversary
total_score: 810
pass: false
date: "2026-06-04"
---

# Journey Eval: sub-item-management (Iteration 2)

## Dimension Scores

| Dimension | Score | Threshold | Pass |
|-----------|-------|-----------|------|
| 1. Completeness | 155/200 | 120 | PASS |
| 2. Semantic Purity | 170/200 | 120 | PASS |
| 3. Precondition Exclusivity | 115/150 | 90 | PASS |
| 4. Fact Alignment | 120/150 | 90 | PASS |
| 5. Surface Fitness | 110/150 | 90 | PASS |
| 6. Internal Consistency | 140/150 | 90 | PASS |

**Total: 810/1000** | Target: 850 | **FAIL**

---

## Dimension 1: Completeness (155/200)

### Journey Metadata (40/50)

- Name `sub-item-management` is valid kebab-case.
- Risk level "Medium" with explicit justification comment: "Workflow involves data mutation (start time edit) without destructive operations. Sorting is read-only." This is clearer than iteration 1 and properly justified.
- Surface types `["web", "api"]` are declared and the journey body now contains both web and API steps (Steps 1-4 web, Step 5 API, E1-E4 web, E5-E8 API). The declaration matches the content.
- Sources cite three documents: prd-user-stories.md (Story 2, Story 5), prd-spec.md (#2, #5), api-handbook.md (Enhanced List Main Items). All are relevant.

**Deduction (-10)**: Sources cite the api-handbook for "Enhanced List Main Items" but the journey does not reference the move/delete/other endpoints defined in that handbook. The scope is appropriate for Stories 2 and 5, but the source reference is slightly over-broad.

### Steps Complete (60/80)

- Happy path Steps 1-4 cover Story 2 (sub-item start time edit: open dialog, modify, save, verify list position) and Story 5 (sub-item list sorted by creation time descending).
- Step 5 adds API-surface coverage for list sub-items sorted by creation time.
- Step 3 explicitly verifies the invariant that editing start time does not change list position -- this directly addresses the iteration 1 critical issue.

- Edge cases cover: validation error (E1), single item (E2), tied timestamps (E3), session expired (E4), unauthorized (E5), unauthenticated (E6), malformed request (E7), not-found (E8).

- Missing steps:
  - No step covers the PRD-defined validation rule: "start time must not be later than end time" (prd-ui-functions.md line 115). E1 covers invalid date format but not the business rule about start-vs-end time ordering. This is a PRD-grounded scenario that should be tested.
  - No step covers start time set to a past date (prd-ui-functions.md line 116: "start time allows past dates"). This is a positive boundary case that confirms allowed behavior.

**Deduction (-20)**: Missing PRD-defined start-vs-end time validation scenario and past-date boundary case.

### Outcomes Coverage (55/70)

- Mandatory derived outcomes for web: E1 covers `validation-error`, E4 covers `session-expired`. Both are present.
- Mandatory derived outcomes for API: E5 covers `unauthorized` (authenticated user without permission), E6 covers `unauthorized` (unauthenticated). Both are present.
- Boundary outcomes: E8 covers `not-found`, E7 covers `validation-error` (API), E1 covers `validation-error` (web). Good coverage.
- Missing: No `not-found` outcome for web surface (e.g., editing a sub-item that was concurrently deleted by another user). This is a common boundary case for edit operations.

**Deduction (-15)**: Missing web-surface not-found/concurrent-deletion boundary outcome.

---

## Dimension 2: Semantic Purity (170/200)

### Natural Language Outcomes (70/80)

- Most expected results use user-observable language well:
  - "Edit dialog renders with all existing fields populated, including the 'start time' field showing the current start time value" (Step 1) -- good.
  - "Start time is updated; save succeeds; dialog closes; sub-item detail reflects the new start time" (Step 2) -- good.
  - "The single sub-item is displayed; no sorting issues occur; empty state is not triggered" (Step E2) -- good.
- API outcomes use appropriately technical language: "The API returns an authorization error; no sub-item data is returned" (E5) -- appropriate for API surface.
- Minor issue: E7 "The API returns a validation error response describing the invalid fields" is slightly vague about what the user/developer observes. "Describing the invalid fields" is acceptable but could be more specific.
- E3 "Sub-items with identical creation times are displayed in a stable, deterministic order" -- this is a good user-observable description, much improved from iteration 1's "code number as tiebreaker".

**Deduction (-10)**: Some outcomes could be more specific (E7); a few API outcomes are slightly generic.

### Declarative Preconditions (55/60)

- Most preconditions are declarative state declarations:
  - "A sub-item exists with a start time value set" (Step 1) -- declarative.
  - "A main item has multiple sub-items created on different dates" (Step 4) -- declarative.
  - "An authenticated API request is sent for a main item that has multiple sub-items" (Step 5) -- declarative.
  - "An API request is sent without valid credentials" (E6) -- declarative.
- E1 precondition "The sub-item edit dialog is open" is borderline procedural (describes a UI state that implies a prior action). Same issue as iteration 1.
- E4 precondition "The user was previously authenticated and the session has since expired while the edit dialog is open" is slightly procedural ("has since expired") but acceptable as a state description.

**Deduction (-5)**: A few preconditions are slightly procedural rather than pure state declarations.

### No Implementation Coupling (45/60)

- No HTTP status codes in web steps -- good, iteration 1 issue resolved.
- No CSS/XPath selectors -- good.
- No component names leaked (e.g., no "Alert component", "DatePicker" references) -- good.
- E5/E6 outcomes use "authorization error" / "authentication error" rather than HTTP status codes -- appropriate for API surface.
- E8 "The API returns a 'not found' error" -- this is semantically acceptable for API surface (not a raw HTTP code like "404"), but it does reference a specific error type. Acceptable.
- E7 "The API returns a validation error response describing the invalid fields; no data is modified" -- acceptable.

**Deduction (-15)**: API outcomes are generally clean but a few could be more behavior-focused rather than error-type-focused. The line between acceptable API terminology and implementation coupling is thin here; the deduction is for borderline cases.

---

## Dimension 3: Precondition Exclusivity (115/150)

### Preconditions Distinct Across Outcomes (50/60)

- Happy path preconditions are distinct:
  - Step 1: sub-item with start time set
  - Step 2: edit dialog open with start time field
  - Step 3: start time just edited in Step 2
  - Step 4: main item with multiple sub-items on different dates
  - Step 5: authenticated API request for main item with multiple sub-items
- Edge case preconditions are distinct:
  - E1: edit dialog open (for validation)
  - E2: main item with exactly one sub-item
  - E3: multiple sub-items created simultaneously
  - E4: previously authenticated, session expired
  - E5: authenticated user without main_item:read permission
  - E6: no valid credentials
  - E7: malformed parameters
  - E8: non-existent main item ID
- Overlap concern: Step 1 precondition ("A sub-item exists with a start time value set") and Step 2 precondition ("The sub-item edit dialog is open with a start time field") are sequential rather than exclusive. This is acceptable for a journey (sequential flow), but if treated as independent outcomes, they overlap.
- E5 and E6 have distinct preconditions (authenticated-without-permission vs unauthenticated) -- good.

**Deduction (-10)**: Some sequential precondition pairs overlap, though this is inherent to journey flow.

### Preconditions Sufficient to Uniquely Select Outcome (40/50)

- All preconditions in iteration 2 now uniquely determine their outcomes:
  - E3: "Multiple sub-items created simultaneously" uniquely leads to "stable, deterministic order". No ambiguity. Fixed from iteration 1.
  - E1: "Edit dialog open" + "invalid date format" uniquely leads to validation error. Clear.
  - E5/E6: Distinct auth states lead to distinct error types. Clear.
- Minor gap: Step 5 precondition says "authenticated API request for a main item that has multiple sub-items" but does not specify the permission level. Setup says "PM user is logged in with sub_item:update permission" but the API endpoint requires `main_item:read`. The setup declares `sub_item:update` but Step 5 exercises `main_item:read`. This is a slight mismatch.

**Deduction (-10)**: Step 5 precondition does not explicitly establish the user has `main_item:read` permission; setup only declares `sub_item:update`.

### No Missing Preconditions for Error/Boundary (25/40)

- E5 precondition specifies "without main_item:read permission" -- explicit. Good.
- E6 precondition specifies "without valid credentials" -- explicit. Good.
- E7 precondition "malformed parameters" is slightly vague -- what constitutes malformed? No concrete setup for what the malformed input looks like.
- E8 precondition "main item ID does not exist in database" -- explicit. Good.
- Missing: No precondition for the start-time-after-end-time business rule validation. This error scenario (from prd-ui-functions.md line 115) has no corresponding edge case with concrete setup.

**Deduction (-15)**: E7 precondition is vague; missing preconditions for PRD-defined business rule (start time vs end time).

---

## Dimension 4: Fact Alignment (120/150)

### Factual Claims Traceable (45/60)

- Fact annotations are present and improved from iteration 1:
  - Step 1: `<!-- fact: prd-spec #2, Story 2 AC1 -->` -- traceable.
  - Step 2: `<!-- fact: Story 2 AC1 -->` -- traceable.
  - Step 3: `<!-- fact: prd-spec #5 -- sort order is by creation time, not start time -->` -- traceable and correct.
  - Step 4: `<!-- fact: Story 5 AC1 -->` -- traceable.
  - Step 5: `<!-- fact: api-handbook Enhanced List Main Items -->` -- traceable.
- Edge cases:
  - E1: `<!-- source: inferred -- derived from Web surface validation-error mandatory outcome -->` -- properly marked as inferred.
  - E4: `<!-- source: inferred -- derived from Web surface session-expired mandatory outcome -->` -- properly marked.
  - E5/E6: `<!-- source: inferred -- derived from API surface unauthorized mandatory outcome -->` -- properly marked.
  - E7: `<!-- source: inferred -- derived from API surface validation-error outcome -->` -- properly marked.
  - E8: `<!-- source: inferred -- derived from API surface not-found common boundary outcome -->` -- properly marked.

**Deduction (-15)**: Some fact annotations could be more precise. Step 5 claims sorting behavior per api-handbook, but the api-handbook "Enhanced List Main Items" endpoint does not explicitly document sub-item sort order. The sort order is documented in prd-ui-functions.md (UF-5: "ORDER BY id DESC"), not in the api-handbook. This is a minor misattribution.

### Inferred Claims Have Source (45/50)

- All inferred edge cases are explicitly marked with `source: inferred` and include reasoning:
  - E1, E4, E5, E6, E7, E8 all have `source: inferred -- derived from ...` annotations.
  - E2 has `fact: Story 5 -- sorting applies regardless of item count` -- traceable.
  - E3 has `source: inferred -- tiebreaker behavior when creation times match` -- properly marked.

**Deduction (-5)**: E3's inference reasoning could be more specific about why tiebreaker behavior matters (it is not explicitly discussed in any source material, making it a true inference rather than a derived requirement).

### No Hallucinated Claims (30/40)

- The hallucinated Step 1b from iteration 1 (start time before parent constraint) has been removed. Good.
- The unverified tiebreaker claim from iteration 1 (code number as tiebreaker) has been replaced with "stable, deterministic order" in E3. This is a safe, non-committal description. Good.
- One concern: E2 claims "no sorting issues occur; empty state is not triggered" for a single sub-item. The "empty state is not triggered" part is reasonable but not explicitly stated in any PRD source. It is an inferred expectation. Minor.
- E7 "API request contains malformed parameters" is generic. The specific nature of the malformed request is not defined. This is not hallucinated but is underspecified.

**Deduction (-10)**: E7 is underspecified (what malformed input?); E2 makes a claim about empty state not triggered that is reasonable but not sourced.

---

## Dimension 5: Surface Fitness (110/150)

### Mandatory Derived Outcomes Present (50/60)

- Web surface:
  - `validation-error`: E1 covers this. Present.
  - `session-expired`: E4 covers this. Present.
- API surface:
  - `unauthorized`: E5 (authenticated without permission) and E6 (unauthenticated) cover this. Present.

**Deduction (-10)**: Web surface is missing a `not-found` boundary outcome (e.g., editing a sub-item that was deleted). While not in the "mandatory" category, it is a common boundary outcome that strengthens the surface coverage.

### Test Strategy Proportions (35/50)

- Web steps: Steps 1-4, E1-E4 = 8 steps
- API steps: Step 5, E5-E8 = 5 steps
- Ratio: approximately 62% web / 38% API.
- This is a significant improvement from iteration 1 (100%/0%) and shows reasonable balance.
- However, the API coverage is somewhat shallow: Step 5 covers listing, and edge cases cover auth/validation/not-found. There is no API step for updating a sub-item's start time (the primary data mutation). The PRD does not define a standalone sub-item update endpoint (per the API Handbook note), but the journey could acknowledge this limitation.

**Deduction (-15)**: API coverage lacks the primary mutation operation (start time update). The API Handbook notes there is no standalone sub-item update endpoint, so this is an architectural constraint, but the journey does not acknowledge it.

### Surface-Specific Environment (25/40)

- Web environment: reasonable assumptions about browser, dialog, list page, date picker. Good.
- API environment: Step 5 and E5-E8 reference "API request", "authenticated API request", "valid authentication token". These are adequate but generic.
- Missing: No explicit mention of the API authentication mechanism (JWT, session token) or request format (JSON body, query parameters). Step 5's action says "authenticated API request is sent to the main items endpoint for a team" which is fine but could specify GET method.
- Setup declares "PM user is logged in with sub_item:update permission" but does not describe the API authentication context.

**Deduction (-15)**: API environment assumptions are generic; no explicit authentication mechanism or request format specified.

---

## Dimension 6: Internal Consistency (140/150)

### Invariants Hold in Every Step (55/60)

Four invariants declared:

1. "The sub-item edit dialog always includes a 'start time' field" -- Steps 1 and 2 verify this. E1 also operates on the dialog with a start time field. Consistent.

2. "Sub-item lists are always sorted by creation time descending" -- Steps 3, 4, and 5 verify this. E2 and E3 also operate on sorted lists. Consistent.

3. "Sub-item creation time determines sort order, not start time or update time" -- Step 3 explicitly verifies this by checking that editing start time does not change list position. Consistent.

4. "Editing start time does not affect the sub-item's position in the sorted list" -- Step 3 is specifically designed to verify this invariant. Consistent.

All four invariants are verified in steps. This directly addresses the iteration 1 critical issue.

**Deduction (-5)**: E4 (session expired) says "no data is modified" which is consistent with invariants, but the "after re-authenticating, the sub-item retains its original values" claim is not verified by any step.

### Cross-Step References Consistent (50/50)

- Step 2 references "Step 2" in Step 3's precondition: "A sub-item's start time was just edited in Step 2" -- clear and correct.
- Setup establishes the context (team, main item, sub-items, PM user with permissions) and all steps reference entities within this context.
- No dangling references.
- No contradictions between steps.

**No deduction.**

### Risk Level Consistent with Content (35/40)

- "Medium" risk is appropriate: the journey involves data mutation (start time edit) and read operations (sorting). No destructive operations (delete is out of scope for this journey). No security-sensitive operations beyond standard auth checks.
- The dual-surface coverage (web + API) adds moderate complexity, consistent with Medium.
- The 13 steps (5 happy path + 8 edge cases) represent moderate complexity. Medium is appropriate.

**Deduction (-5)**: The journey covers auth/session edge cases (E4, E5, E6) which add security-related complexity. "Medium" is still appropriate but the security surface could be noted in the risk justification.

---

## Critical Issues

1. **Missing PRD-defined validation scenario**: prd-ui-functions.md line 115 explicitly states "start time must not be later than end time (if end time is set)". No edge case tests this business rule. This is a PRD-grounded scenario, not an inferred one, and its absence means a concrete product requirement is untested.

2. **No API step for the primary mutation (start time update)**: The API Handbook notes there is no standalone sub-item update endpoint, but the journey does not acknowledge this. Step 5 covers only the read path (listing). The write path (start time update) is covered only on the web surface. If the update happens via the web form and not a dedicated API, this should be explicitly noted.

3. **E7 underspecification**: "An API request contains malformed parameters" / "An API request is sent with invalid parameter values" -- what specific parameters? What invalid values? The precondition does not define concrete malformed input, making the edge case difficult to translate into a test.

## Pass/Fail Status

| Check | Result |
|-------|--------|
| Dimension 1 >= 120 | PASS (155) |
| Dimension 2 >= 120 | PASS (170) |
| Dimension 3 >= 90 | PASS (115) |
| Dimension 4 >= 90 | PASS (120) |
| Dimension 5 >= 90 | PASS (110) |
| Dimension 6 >= 90 | PASS (140) |
| Total >= 850 | FAIL (810) |
| **Overall** | **FAIL** |

All 6 dimensions now pass their individual thresholds. Total score 810/1000 is 40 points below the 850 target. The journey shows substantial improvement from iteration 1 (555 -> 810, +255 points). All 5 critical issues from iteration 1 have been addressed.

### Recommended Fixes for Iteration 3 (Priority Order)

1. Add an edge case for "start time after end time" validation (prd-ui-functions.md line 115). This is the most significant gap -- a PRD-defined rule with no corresponding test step. (+15-20 points across Completeness and Fact Alignment)
2. Specify concrete malformed parameters in E7 (e.g., "start_date field contains non-date string 'abc'"). (+5-10 points across Precondition Exclusivity and Completeness)
3. Add a note acknowledging that sub-item start time update has no standalone API endpoint and is handled via the web form. (+5-10 points across Surface Fitness)
4. Add a web-surface not-found edge case (editing a concurrently-deleted sub-item). (+5-10 points across Completeness and Surface Fitness)

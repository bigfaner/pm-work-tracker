---
iteration: 1
scorer: adversary
total_score: 555
pass: false
date: "2026-06-04"
---

# Journey Eval: sub-item-management (Iteration 1)

## Dimension Scores

| Dimension | Score | Threshold | Pass |
|-----------|-------|-----------|------|
| 1. Completeness | 110/200 | 120 | FAIL |
| 2. Semantic Purity | 155/200 | 120 | PASS |
| 3. Precondition Exclusivity | 100/150 | 90 | PASS |
| 4. Fact Alignment | 50/150 | 90 | FAIL |
| 5. Surface Fitness | 60/150 | 90 | FAIL |
| 6. Internal Consistency | 80/150 | 90 | FAIL |

**Total: 555/1000** | Target: 850 | **FAIL**

---

## Dimension 1: Completeness (110/200)

### Journey Metadata (30/50)

- Name `sub-item-management` is valid kebab-case.
- Risk level "Medium" is stated but justification is thin. The comment mentions "multi-step interaction without irreversible side effects" which is reasonable but generic.
- Surface types `["web", "api"]` are declared, but the journey body contains **no API-specific steps**. All steps are described from a web UI perspective only. The surface_types declaration is inaccurate relative to actual content.
- Sources are cited and traceable.

**Deduction (-20)**: Surface types claim `api` but no API journey steps exist. Risk justification is minimal.

### Steps Complete (50/80)

- Happy path covers Story 2 (sub-item start time edit) and Story 5 (sub-item list sorted by creation time) adequately.
- Edge cases cover: start time before parent, invalid date, single sub-item, tied creation times.
- Missing steps:
  - No API-surface steps despite declaring `api` in surface_types. An API journey should include: PUT/PATCH sub-item update request, GET sub-item list sorted by creation time.
  - No step covers the "start time not later than end time" validation rule (prd-ui-functions.md line 115), which is an actual PRD-defined constraint.
  - No step covers the "start time allows past dates" behavior (prd-ui-functions.md line 116).

**Deduction (-30)**: Missing API steps; missing PRD-defined validation scenarios.

### Outcomes Coverage (30/70)

- No derived outcomes for mandatory surface types:
  - Web: missing `validation-error` outcome (Step 2b partially covers this but is not formally declared as a derived outcome).
  - Web: missing `session-expired` outcome entirely.
  - API: missing `unauthorized` outcome entirely.
- Common boundary outcomes partially covered: `not-found` is absent (e.g., editing a deleted sub-item), `validation-error` is implicit but not formally declared.

**Deduction (-40)**: Mandatory derived outcomes are absent for both surface types.

---

## Dimension 2: Semantic Purity (155/200)

### Natural Language Outcomes (65/80)

- Most expected results use user-observable language: "Edit dialog renders with all existing fields populated", "Start time is updated; save succeeds; dialog closes".
- Minor issue: "Start Time" field name in quotes is acceptable as it refers to a user-visible label.
- Step 4b uses "sub-item code number as tiebreaker" which is a data model concept rather than user-observable behavior. A user would observe "items appear in a consistent order" without knowing about code numbers.

**Deduction (-15)**: Step 4b leaks a data model concept into user-observable outcome.

### Declarative Preconditions (55/60)

- Most preconditions are declarative state declarations: "The main item has a start time of 2026-06-01", "A main item has exactly one sub-item".
- Step 2b precondition "Sub-item edit dialog is open" is borderline procedural (describes a UI state that implies a prior action).

**Deduction (-5)**: One precondition is slightly procedural.

### No Implementation Coupling (35/60)

- No HTTP status codes in web steps -- good.
- No CSS/XPath selectors -- good.
- However, Step 4b mentions "sub-item code number" which is an implementation detail leaking through. The PRD does use this term in design docs but for a journey-level description this is coupling to the data model.

**Deduction (-25)**: Step 4b references code number (data model coupling). Step 1b references "backend validation rules" which is implementation leakage.

---

## Dimension 3: Precondition Exclusivity (100/150)

### Preconditions Distinct Across Outcomes (45/60)

- The four edge cases have distinct preconditions: different dates, invalid input, single item, tied timestamps.
- However, Step 1b's precondition ("main item has start time 2026-06-01") and Step 2b's precondition ("sub-item edit dialog is open") have different scopes but the distinction is not always clean.

**Deduction (-15)**: Some precondition pairs could be more clearly differentiated.

### Preconditions Sufficient to Uniquely Select Outcome (30/50)

- Step 1b precondition specifies a concrete date but the expected result is ambiguous: "System behavior depends on backend validation rules; if disallowed... if allowed...". This means the precondition does NOT uniquely determine the outcome. This is a significant issue.
- Step 2b is clearer: invalid date format leads to validation error.
- Steps 3b and 4b are clear: single item and tied timestamps have deterministic outcomes.

**Deduction (-20)**: Step 1b precondition does not uniquely determine outcome.

### No Missing Preconditions for Error/Boundary (25/40)

- Step 1b has preconditions but the outcome is ambiguous (see above).
- Missing error/boundary preconditions:
  - What happens when editing a sub-item that has been deleted by another user concurrently?
  - What happens when the sub-item has an end time set and the new start time is after it?
  - Session expiry during edit is not covered.

**Deduction (-15)**: Several error/boundary scenarios lack preconditions.

---

## Dimension 4: Fact Alignment (50/150)

### Factual Claims Traceable (15/60)

- The journey makes no use of fact annotations (no `[fact:PRD-...]` style markers). This is a structural requirement.
- The sort order claim (creation time descending) is correct per PRD Story 5.
- The code number tiebreaker claim (Step 4b) is partially traceable: the tech design mentions `ORDER BY id DESC` but the PRD itself does not mention code number as a tiebreaker. The PRD UI functions spec says `ORDER BY id DESC` which is an implementation detail, and id DESC is not the same as "creation time descending with code number tiebreaker".

**Deduction (-45)**: No fact annotations; tiebreaker claim is not in the PRD.

### Inferred Claims Have Source (20/50)

- Step 1b ("start time before parent") is presented without any `inferred` or `source:inferred` marking. More critically, this claim **contradicts the PRD**.
- Step 4b (code number tiebreaker) is presented without inference marking.
- The PRD UI functions doc (line 117) explicitly states: "开始时间无需与父主事项时间范围对齐（子事项时间可独立于主事项）" -- sub-item start time does NOT need to align with parent main item time range. Step 1b is a hallucinated constraint.

**Deduction (-30)**: Step 1b contradicts PRD; no inference markings on derived claims.

### No Hallucinated Claims (15/40)

- **Step 1b is a hallucinated constraint.** The PRD explicitly states start time is independent of parent main item. The journey fabricates a validation rule that does not exist in the source material.
- Step 4b's "code number as tiebreaker" claim is not found in the PRD user stories or spec. The tech design says `ORDER BY id DESC`, which means database ID order, not "creation time descending with code number tiebreaker". This is a misrepresentation.

**Deduction (-25)**: One hallucinated constraint (Step 1b); one unverified claim (Step 4b tiebreaker).

---

## Dimension 5: Surface Fitness (60/150)

### Mandatory Derived Outcomes Present (20/60)

Despite declaring `["web", "api"]` surface types:

- **Web missing**: No formally declared `validation-error` derived outcome, no `session-expired` derived outcome.
- **API missing**: No `unauthorized` derived outcome. In fact, no API outcomes at all.
- The journey reads entirely as a web UI journey with no API-surface content.

**Deduction (-40)**: Critical missing outcomes for both declared surface types.

### Test Strategy Proportions (20/50)

- 100% web steps, 0% API steps for a dual-surface journey.
- No balance between web and API test strategies.

**Deduction (-30)**: Completely unbalanced surface coverage.

### Surface-Specific Environment (20/40)

- Web environment assumptions are reasonable (browser, dialog, list page).
- No API environment assumptions at all (no mention of API authentication, request/response patterns).

**Deduction (-20)**: API environment entirely absent.

---

## Dimension 6: Internal Consistency (80/150)

### Invariants Hold in Every Step (30/60)

- Invariant: "sub-item edit dialog always includes a Start Time field" -- holds in all edit steps.
- Invariant: "sub-item lists sorted by creation time descending" -- holds in list steps.
- Invariant: "Editing start time does not affect sorted list position" -- Step 2 mentions the edit but does not verify list position preservation. The journey should have a verification step showing the edited sub-item remains in its original list position.
- Invariant: "Sub-item creation time determines sort order, not start time" -- consistent but not explicitly verified in any step.

**Deduction (-30)**: Invariants are declared but not verified in steps; Step 2 does not confirm list position preservation.

### Cross-Step References Consistent (30/50)

- Steps reference "sub-item" and "main item" consistently.
- No dangling references.
- However, Step 1b references "the main item has a start time of 2026-06-01" but no prior step established this value. Setup section mentions "Sub-items have various start times assigned" but does not mention main item start time.

**Deduction (-20)**: Step 1b introduces a main item start time not established in setup.

### Risk Level Consistent with Content (20/40)

- "Medium" risk is reasonable for editing and sorting.
- However, the journey covers two distinct stories (edit + sort) and includes edge cases that span different concerns. The complexity is moderate, consistent with Medium.
- The lack of API coverage despite declaring it suggests the risk assessment did not account for dual-surface testing needs.

**Deduction (-20)**: Risk level does not account for the gap between declared and actual surface coverage.

---

## Critical Issues

1. **Hallucinated constraint (Step 1b)**: The journey asserts that start time before parent main item may be disallowed. The PRD explicitly contradicts this (prd-ui-functions.md line 117: "开始时间无需与父主事项时间范围对齐"). This edge case must be removed or replaced.

2. **No API surface content**: The journey declares `surface_types: ["web", "api"]` but contains zero API-specific steps. Either remove `api` from surface_types or add API journey steps covering PUT sub-item update and GET sub-item list.

3. **Missing mandatory derived outcomes**: No `validation-error`, `session-expired` (web), or `unauthorized` (api) outcomes are declared.

4. **Unverified tiebreaker claim (Step 4b)**: "Code number as tiebreaker" is not documented in the PRD. The tech design says `ORDER BY id DESC`, which means id-based ordering. Either remove this claim or trace it to a source.

5. **Invariants not verified in steps**: The invariant "editing start time does not affect list position" is declared but no step actually verifies this by editing start time and then checking the list order.

## Pass/Fail Status

| Check | Result |
|-------|--------|
| Dimension 1 >= 120 | FAIL (110) |
| Dimension 2 >= 120 | PASS (155) |
| Dimension 3 >= 90 | PASS (100) |
| Dimension 4 >= 90 | FAIL (50) |
| Dimension 5 >= 90 | FAIL (60) |
| Dimension 6 >= 90 | FAIL (80) |
| Total >= 850 | FAIL (555) |
| **Overall** | **FAIL** |

4 of 6 dimensions below threshold. Total score 555/1000 is 295 points below the 850 target.

### Recommended Fixes (Priority Order)

1. Remove Step 1b (hallucinated constraint contradicts PRD). Replace with a PRD-grounded edge case: start time after end time (prd-ui-functions.md line 115: "开始时间不得晚于结束时间").
2. Either add API-surface journey steps or remove `api` from surface_types.
3. Declare mandatory derived outcomes: `validation-error`, `session-expired` for web; `unauthorized` for API (if kept).
4. Add a verification step after Step 2 that confirms the edited sub-item retains its original position in the sorted list (validates the invariant).
5. Verify or remove the "code number tiebreaker" claim in Step 4b against the tech design's `ORDER BY id DESC`.

---
date: "2026-05-13"
doc_dir: "docs/features/milestone-map/testing/"
iteration: "1"
target_score: "85"
evaluator: Claude (automated, adversarial)
---

# Test Cases Eval -- Iteration 1

**Score: 62/100** (target: 85)

```
+------------------------------------------------------------------+
|                  TEST CASES QUALITY SCORECARD                     |
+------------------------------------------------------------------+
| Dimension                    | Score    | Max      | Status       |
|------------------------------|----------|----------|--------------|
| 1. PRD Traceability          |  16      |  25      | WARNING      |
|    TC-to-AC mapping          |  7/9     |          |              |
|    Traceability table        |  7/8     |          |              |
|    Reverse coverage          |  2/8     |          |              |
|------------------------------|----------|----------|--------------|
| 2. Step Actionability        |  18      |  25      | WARNING      |
|    Steps concrete            |  5/9     |          |              |
|    Expected results          |  7/9     |          |              |
|    Preconditions explicit    |  6/7     |          |              |
|------------------------------|----------|----------|--------------|
| 3. Route & Element Accuracy  |  12      |  20      | WARNING      |
|    Routes valid              |  6/7     |          |              |
|    Elements identifiable     |  3/7     |          |              |
|    Consistency               |  3/6     |          |              |
|------------------------------|----------|----------|--------------|
| 4. Completeness              |  11      |  20      | WARNING      |
|    Type coverage             |  6/7     |          |              |
|    Boundary cases            |  5/7     |          |              |
|    Integration scenarios     |  0/6     |          |              |
|------------------------------|----------|----------|--------------|
| 5. Structure & ID Integrity  |  5       |  10      | WARNING      |
|    IDs sequential/unique     |  4/4     |          |              |
|    Classification correct    |  1/3     |          |              |
|    Summary matches actual    |  0/3     |          |              |
|------------------------------|----------|----------|--------------|
| TOTAL                        |  62      |  100     |              |
+------------------------------------------------------------------+
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| TC-044..TC-047 | Integration TCs are classified as Type "UI" instead of "Integration" -- cross-section inconsistency | -3 pts (Structure) |
| Summary table | Says "UI: 47" but 47 includes 4 Integration TCs listed as Type UI. Summary also lists "Integration: 4" separately. Double-counting: 47 UI + 4 Integration = 51 but actual UI section has 47 entries, 4 of which are integration. Summary math is inconsistent | -3 pts (Structure) |
| TC-001..TC-026, TC-035..TC-037, TC-047 | All milestone page TCs use `Element: sitemap-missing` -- no identifiable selector | -4 pts (Route & Element) |
| TC-001 step 1 | "Click the create milestone map button" -- no data-testid or aria-label to identify the button | -2 pts (Step Actionability) |
| TC-005 step 1 | "Click the edit button on a milestone map card" -- which card? which button? No element identifier | -2 pts (Step Actionability) |
| TC-010 step 2 | "Observe the list view" -- vague observation step, not a concrete action | -2 pts (Step Actionability) |
| TC-044..TC-047 | Integration TCs use Source "UF-4/UF-5/UF-6/UF-1" which points to UI functions, not PRD acceptance criteria. Source should trace to PRD user stories or spec | -2 pts (Traceability) |
| Story 8/AC-2 | PRD AC says "filter by status shows only in-progress cards" but no TC maps Source to Story 8/AC-2 directly. TC-009 maps to Story 3/AC-4 | -2 pts (Reverse coverage) |
| Story 11/AC-1, AC-3, AC-4 | No TC maps Source to Story 11/AC-1, AC-3, or AC-4. While functionally covered by TC-035 (Story 9) and TC-036 (Story 9), the Source does not reference Story 11, breaking traceability | -4 pts (Reverse coverage) |
| TC-044..TC-047 | Integration TCs are listed under "UI Test Cases" section header but are integration tests. The document has no "Integration Test Cases" section | -2 pts (Classification) |
| TC-007 expected | "Dropdown shows 'pending-implementation' and 'completed' options" but PRD Story 3/AC-2 says the map is in "in-progress" and options should be "pending-implementation" (回退) and "completed". The PRD state machine uses "待实施" not "pending-implementation" -- term mismatch with PRD Chinese status names | -2 pts (Expected results) |

---

## Attack Points

### Attack 1: Route & Element Accuracy -- 26 TCs have no identifiable element selectors

**Where**: TC-001 through TC-026, TC-035 through TC-037, TC-047 all have `Element: sitemap-missing`. This is 29 out of 65 TCs (45%) with zero element specificity.

**Why it's weak**: The rubric requires "Every Element field uses a selector strategy: data-testid, aria-label, or semantic locator." Having `sitemap-missing` in 45% of TCs means nearly half the test suite cannot be automated as written. A test script generator would have to guess or invent selectors for every milestone page interaction. The WARNING at the top of the document acknowledges this but does not mitigate the scoring impact -- the sitemap should be generated or elements manually specified before test cases are finalized.

**What must improve**: Either (a) run `/gen-sitemap` to populate sitemap.json with `/milestones` page data and re-generate element IDs, or (b) manually add `data-testid` or `aria-label` selectors for each milestone page element. Every TC on the `/milestones` route needs a real element identifier.

### Attack 2: Reverse Coverage -- Story 11 acceptance criteria are untraceable

**Where**: No TC has `Source: Story 11 / AC-1`, `Source: Story 11 / AC-3`, or `Source: Story 11 / AC-4`. Story 11 defines management-level read-only access behaviors (4 ACs). AC-2 is covered by TC-037, but AC-1, AC-3, and AC-4 are orphaned.

**Why it's weak**: While TC-035 and TC-036 functionally test similar scenarios (Story 9), the PRD intentionally separates Story 9 (team member view) from Story 11 (management view) because they represent different user personas with potentially different UX requirements. The test cases do not distinguish between these personas. Additionally, Story 11/AC-4 (API timeout/500 for management-level user) has no coverage at all -- TC-012 covers Story 8/AC-6 which is the same error state but for a different user persona.

**What must improve**: Add TCs with Source mapping to Story 11/AC-1, AC-3, AC-4. At minimum, update TC-035 and TC-036 Source fields to include "Story 11 / AC-1" and "Story 11 / AC-3" alongside their existing Story 9 references. Add a new TC for Story 11/AC-4 (management user seeing retry on API failure).

### Attack 3: Step Actionability -- Steps use vague action verbs without element identification

**Where**: Multiple TCs use non-actionable step language:
- TC-001 step 2: "Fill in name (valid, 1-100 chars) and optional description" -- which input field? How to identify it?
- TC-005 step 1: "Click the edit button on a milestone map card" -- which card? which button specifically?
- TC-010 step 2: "Observe the list view" -- this is not an action, it's a passive observation
- TC-013 step 2: "Observe the timeline view" -- same passive observation problem
- TC-014 steps 2,4,6: "Observe axis labels" -- what labels? What values to verify?
- TC-020 step 1: "Open milestone detail panel by clicking a milestone node" -- which node?

**Why it's weak**: The rubric requires "Each step describes a single, unambiguous user action" with specificity like "Click the Submit button" not "Submit the form". Steps like "Observe the list view" and "Observe axis labels" are not verifiable actions -- they describe nothing a test script can execute. Combined with the `sitemap-missing` element problem, this means the milestone page TCs are closer to requirement restatements than executable test procedures.

**What must improve**: Replace all "Observe X" steps with concrete verification actions (e.g., "Verify element [data-testid='milestone-card-0'] is visible"). Add element identifiers to click/fill actions. Each step must be a single, automatable action with a clear target element.

---

## Previous Issues Check

<!-- First iteration -- no previous issues -->

---

## Verdict

- **Score**: 62/100
- **Target**: 85/100
- **Gap**: 23 points
- **Step Actionability**: 18/25 (not blocking, but borderline)
- **Action**: Continue to iteration 2. Priority fixes: (1) generate sitemap and add real element IDs to all `sitemap-missing` TCs, (2) add TCs for Story 11 orphaned ACs, (3) rewrite vague "Observe" steps as concrete verify actions with element targets.

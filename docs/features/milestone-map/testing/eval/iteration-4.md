---
date: "2026-05-13"
doc_dir: "docs/features/milestone-map/testing/"
iteration: "4"
target_score: "85"
evaluator: Claude (automated, adversarial)
---

# Test Cases Eval -- Iteration 4

**Score: 78/100** (target: 85)

```
┌─────────────────────────────────────────────────────────────────┐
│                  TEST CASES QUALITY SCORECARD                     │
├──────────────────────────────────────────────────────────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. PRD Traceability          │  22.5    │  25      │ ⚠️         │
│    TC-to-AC mapping          │  6.5/9   │          │            │
│    Traceability table        │  8/8     │          │            │
│    Reverse coverage          │  8/8     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Step Actionability        │  17.5    │  25      │ ⚠️         │
│    Steps concrete            │  6/9     │          │            │
│    Expected results          │  7/9     │          │            │
│    Preconditions explicit    │  4.5/7   │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Route & Element Accuracy  │  14      │  20      │ ⚠️         │
│    Routes valid              │  7/7     │          │            │
│    Elements identifiable     │  4/7     │          │            │
│    Consistency               │  3/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Completeness              │  19      │  20      │ ⚠️         │
│    Type coverage             │  7/7     │          │            │
│    Boundary cases            │  7/7     │          │            │
│    Integration scenarios     │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┤            │
│ 5. Structure & ID Integrity  │  4.5     │  10      │ ⚠️         │
│    IDs sequential/unique     │  3.5/4   │          │            │
│    Classification correct    │  1/3     │          │            │
│    Summary matches actual    │  0/3     │          │            │
├──────────────────────────────┼──────────┼──────────┤            │
│ TOTAL                        │  77.5    │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

> Rounded total: 78/100

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| TC-047 Source | Source is "PRD UI Function 'UF-1' Navigation Architecture" -- references a UI function document, not a PRD user story acceptance criterion. Persisted since iteration 2 | -1 pt (TC-to-AC mapping) |
| TC-056, TC-058, TC-063, TC-065 Sources | Source is "PRD Spec Related Changes #N" -- a spec section reference describing implementation changes, not a user story AC. 4 TCs affected | -1 pt (TC-to-AC mapping) |
| TC-070 Source | Source is "PRD Spec Security Requirements" -- vague; should specify which permission codes and which entity | -0.5 pt (TC-to-AC mapping) |
| TC-010 step 3 | "milestone count, item count, and progress percentage" referenced without selectors -- no data-testid for these sub-elements | -0.5 pt (Steps concrete) |
| TC-013 step 3 | "each `[data-testid='milestone-node']` displays name, planned completion date, status, and completion rate" -- no selectors for these sub-elements | -0.5 pt (Steps concrete) |
| TC-038 step 3 | "Verify `[data-testid='columnheader-milestone']` is positioned between the 'title' (E-089) and 'priority' (E-090) columns" -- spatial assertion mixing data-testid with sitemap IDs, not programmatically verifiable | -0.5 pt (Steps concrete) |
| TC-044 step 2 | "Verify `[data-testid='filter-milestone']` is visible in the filter bar area, positioned to the right of the 'assignee' filter (E-011)" -- spatial assertion mixing selector systems | -0.5 pt (Steps concrete) |
| TC-045 step 3 | "Verify `[data-testid='select-milestone']` is visible below the 'Assignee' combobox (E-037)" -- spatial assertion mixing selector systems | -0.5 pt (Steps concrete) |
| TC-046 step 2 | "Verify `[data-testid='columnheader-milestone']` is visible between the 'Title' (E-089) and 'Priority' (E-090) column headers" -- spatial assertion mixing selector systems | -0.5 pt (Steps concrete) |
| TC-010 expected | "milestone count, item count, and overall progress visible" -- no specific expected values to verify against | -0.5 pt (Expected results) |
| TC-013 expected | "each node shows name, date, status, and completion rate" -- no specific expected values | -0.5 pt (Expected results) |
| TC-023 expected | "the `[data-testid='milestone-node']` visual style updates to in-progress appearance" -- "visual style" is subjective, not objectively verifiable without a CSS class or attribute | -0.5 pt (Expected results) |
| TC-035 expected | "no create/edit/delete buttons are functional or visible" -- conflates "not functional" and "not visible"; step 3 checks btn-create-map is disabled but step 4 checks btn-edit-map/btn-delete are "not present" -- two different verification strategies conflated in one expected result | -0.5 pt (Expected results) |
| TC-021 preconditions | "Two PMs are editing the same milestone simultaneously" -- does not explain how to establish two concurrent sessions or what data state is required. Persisted since iteration 2 | -1 pt (Preconditions) |
| TC-044 preconditions | "Milestone filter component build complete, integration task complete" -- build-process precondition, not test data state | -0.5 pt (Preconditions) |
| TC-045 preconditions | Same build-process precondition pattern as TC-044 | -0.5 pt (Preconditions) |
| TC-046 preconditions | Same build-process precondition pattern as TC-044 | -0.5 pt (Preconditions) |
| TC-010 step 3 | "milestone count, item count, and progress percentage" -- no data-testid selectors for these sub-fields | -0.5 pt (Elements identifiable) |
| TC-013 step 3 | "name, planned completion date, status, and completion rate" -- no data-testid selectors for these sub-fields | -0.5 pt (Elements identifiable) |
| TC-038 step 3 | Mixes E-089/E-090 sitemap IDs with data-testid selectors in same step | -0.5 pt (Elements identifiable) |
| TC-044 step 2 | Uses E-011 sitemap ID mixed with data-testid | -0.5 pt (Elements identifiable) |
| TC-045 step 3 | Uses E-037 sitemap ID mixed with data-testid | -0.5 pt (Elements identifiable) |
| TC-046 step 2 | Uses E-089/E-090 sitemap IDs mixed with data-testid | -0.5 pt (Elements identifiable) |
| TC-048 section vs type | Body: Type: UI. Traceability: Type: UI. Section: "### Cross-Interface Integration Tests". TC classified as UI but placed under Integration section. Cross-section inconsistency: -3 pts per rubric rule | -3 pts (Consistency) |
| No UI->API integration | All integration TCs (TC-049..TC-052) go API->UI direction only. No TC verifies UI action resulting in correct API state | -1 pt (Integration scenarios) |
| TC-043 -> TC-047 -> TC-044 | IDs not sequential in document order: TC-047 (line 791) appears before TC-044 (line 808), TC-045, TC-046. All 70 IDs exist but out of sequence | -0.5 pt (IDs sequential) |
| TC-048 classification | Type: UI placed under "Cross-Interface Integration Tests" section -- section misclassification | -1 pt (Classification correct) |
| TC-044/045/046 classification | Classified as "Integration" but only verify component visibility on existing pages, not cross-interface behavior | -1 pt (Classification correct) |
| Summary table UI count | Summary says UI=46, Integration=6. Actual body/traceability counts: UI=45, Integration=7. UI off by +1, Integration off by -1 | -3 pts (Summary matches actual) |

---

## Attack Points

### Attack 1: Structure & ID Integrity -- Summary table does not match actual counts

**Where**: Summary table (line 57-62): `| UI | 46 |`, `| Integration | 6 |`. Actual body/traceability Type counts: UI=45, Integration=7 (verified by grep).

**Why it's weak**: The summary table is a required section that directly communicates the document's test distribution to reviewers and test script generators. A mismatch of +1 UI / -1 Integration means anyone relying on the summary has incorrect data. The rubric allocates 3 points specifically for "Summary table matches actual" and awards 0 for a mismatch. This is a trivially fixable issue -- the count is off by exactly 1 for both UI and Integration -- but it has persisted since the summary was written. Combined with TC-048 being a UI test placed under an Integration section header, the count error suggests the author may have counted TC-048 as Integration by section placement rather than by its declared Type field.

**What must improve**: Change summary table to `UI | 45` and `Integration | 7` to match the body and traceability counts. Optionally, also move TC-048 out of the "Cross-Interface Integration Tests" section since it is Type: UI, which would resolve the section-vs-type inconsistency simultaneously.

### Attack 2: Route & Element Accuracy -- TC-048 is a UI test placed under Integration section (cross-section inconsistency)

**Where**: TC-048 body (line 859): `**Type**: UI`. TC-048 section header (line 855): `### Cross-Interface Integration Tests`. TC-048 traceability table (line 1216): `TC-048 | Story 11 / AC-4 | UI | ui/milestones | P1`.

**Why it's weak**: The rubric penalizes cross-section inconsistencies at -3 pts per conflict. TC-048 is declared as Type: UI everywhere (body, traceability) but lives under a section titled "Cross-Interface Integration Tests". This creates ambiguity for test script generators: should TC-048 be placed in the UI test suite or the Integration test suite? Additionally, TC-048 tests the same behavior as TC-012 (error state on API failure for /milestones page) but from a different persona (management user). It is functionally a UI test and should be in a UI section.

**What must improve**: Either (a) move TC-048 to a UI section (e.g., under "Permission-Based Views" or "Milestone Map Page") and keep Type: UI, or (b) if it should be Integration, change Type to Integration and add cross-interface verification steps. Option (a) is the correct fix since TC-048 has no cross-interface behavior.

### Attack 3: Step Actionability -- 6 TCs mix sitemap element IDs (E-XXX) with data-testid selectors in spatial assertions

**Where**: TC-038 step 3 (line 709): "Verify `[data-testid='columnheader-milestone']` is positioned between the 'title' (E-089) and 'priority' (E-090) columns". TC-044 step 2 (line 818): "positioned to the right of the 'assignee' filter (E-011)". TC-045 step 3 (line 834): "visible below the 'Assignee' combobox (E-037)". TC-046 step 2 (line 849): "visible between the 'Title' (E-089) and 'Priority' (E-090) column headers".

**Why it's weak**: These steps use two different element identification systems in a single assertion: provisional data-testid selectors for the new milestone components and sitemap E-XXX IDs for existing page elements. A test script generator must resolve both systems. More critically, the assertions are spatial ("positioned between", "to the right of", "below") which have no programmatic verification method in Playwright -- CSS grid/flexbox ordering cannot be reliably verified by position relative to sibling elements. The rubric requires steps to describe "a single, unambiguous user action" with specific element identification. These spatial assertions are verifiable only by human visual inspection, not by automated scripts.

**What must improve**: Replace spatial assertions with structural ones. For example, instead of "positioned between title and priority columns", verify that the milestone column header appears at the expected DOM index within the header row: `expect(headerRow.locator('th').nth(2)).toHaveTestId('columnheader-milestone')`. For the filter bar and dialog assertions, verify the element exists in the correct container rather than its position relative to another element.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (Iter 3): TC-047 three-way classification conflict (UI body / Integration traceability / Integration section) | ✅ | TC-047 body Type: UI (line 794), traceability Type: UI (line 1215), section: "### Navigation" under UI Test Cases (line 789). Now consistent across all three. The TC was moved out of the Integration section |
| Attack 2 (Iter 3): Story 8 / AC-2 remains orphaned for 3 iterations | ✅ | TC-009 Source (line 209): "Story 3 / AC-4, Story 8 / AC-2". Traceability table (line 1177): "Story 3 / AC-4, Story 8 / AC-2". Now correctly traced |
| Attack 3 (Iter 3): Edit button clicks in 4 TCs lack element selectors | ✅ | TC-031 Element field (line 595): includes `[data-testid='btn-edit-item']`. TC-032 (line 611), TC-034 (line 641) also include it. Step 2 in TC-031 (line 598): "Click `[data-testid='btn-edit-item']` to open the edit dialog". Provisional element map (line 48) defines it |

---

## Verdict

- **Score**: 78/100
- **Target**: 85/100
- **Gap**: 7 points
- **Step Actionability**: 17.5/25 (below blocking threshold of 20 -- ⚠️ BLOCKING)
- **Action**: Continue to iteration 5. Priority fixes: (1) correct summary table counts to UI=45, Integration=7 (+3 pts), (2) move TC-048 out of "Cross-Interface Integration Tests" section or reclassify it (+3 pts), (3) replace spatial assertions with structural DOM-index checks in TC-038/044/045/046 (+3 pts), (4) add concrete data-testid selectors for sub-elements referenced in TC-010/013 steps (+1 pt). These four fixes would recover approximately 10 points, bringing the score to ~88/100.

---
date: "2026-05-13"
doc_dir: "docs/features/milestone-map/testing/"
iteration: "3"
target_score: "85"
evaluator: Claude (automated, adversarial)
---

# Test Cases Eval -- Iteration 3

**Score: 80/100** (target: 85)

```
┌─────────────────────────────────────────────────────────────────┐
│                  TEST CASES QUALITY SCORECARD                     │
├──────────────────────────────────────────────────────────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. PRD Traceability          │  20      │  25      │ ⚠️         │
│    TC-to-AC mapping          │  7/9     │          │            │
│    Traceability table        │  7/8     │          │            │
│    Reverse coverage          │  6/8     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Step Actionability        │  19      │  25      │ ⚠️         │
│    Steps concrete            │  6/9     │          │            │
│    Expected results          │  8/9     │          │            │
│    Preconditions explicit    │  5/7     │          │            │
├──────────────────────────────┼──────────┼──────────┤            │
│ 3. Route & Element Accuracy  │  15      │  20      │ ⚠️         │
│    Routes valid              │  7/7     │          │            │
│    Elements identifiable     │  6/7     │          │            │
│    Consistency               │  2/6     │          │            │
├──────────────────────────────┼──────────┼──────────┤            │
│ 4. Completeness              │  18      │  20      │ ⚠️         │
│    Type coverage             │  7/7     │          │            │
│    Boundary cases            │  7/7     │          │            │
│    Integration scenarios     │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┤            │
│ 5. Structure & ID Integrity  │  8.5     │  10      │ ⚠️         │
│    IDs sequential/unique     │  4/4     │          │            │
│    Classification correct    │  1.5/3   │          │            │
│    Summary matches actual    │  3/3     │          │            │
├──────────────────────────────┼──────────┼──────────┤            │
│ TOTAL                        │  80.5    │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

> Rounded total: 80/100

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| TC-047 Source | Source is "PRD UI Function 'UF-1' Navigation Architecture" which references a UI function document, not a PRD user story acceptance criterion. Persisted from iteration 2 | -1 pt (TC-to-AC mapping) |
| TC-056, TC-058, TC-063, TC-065 Sources | Source is "PRD Spec Related Changes #N" -- a spec section reference, not a user story AC. Less specific than the "Story N / AC-N" standard used by most TCs | -1 pt (TC-to-AC mapping) |
| TC-070 Source | Source is "PRD Security Requirements" -- vague. Should specify which permission codes and which entity (milestone-map vs milestone) | -0.5 pt (TC-to-AC mapping) |
| Traceability TC-047 | Source column says "UF-1 Navigation Architecture" while TC body says "PRD UI Function 'UF-1' Navigation Architecture" -- abbreviation mismatch | -0.5 pt (Traceability table) |
| TC-047 Type mismatch | TC body says Type: UI; traceability table says Type: Integration; section header says "Integration Tests". Three-way classification conflict | -1 pt (Traceability table) |
| Story 8 / AC-2 reverse coverage | PRD says "When I filter by status selecting 'in-progress', then only in-progress milestone map cards are shown" (Story 8 / AC-2). No TC's Source field references Story 8 / AC-2. TC-009 covers the same behavior but traces to Story 3 / AC-4. Persisted from iteration 2 | -1 pt (Reverse coverage) |
| PRD Security Requirements detail | TC-070 tests permission denial but the Source "PRD Security Requirements" is vague -- the PRD specifies 4 separate permission codes (milestone:create/update/delete/read) applied to both milestone-map and milestone entities. The TC does not distinguish these in its Source | -1 pt (Reverse coverage) |
| TC-031 step 2 | "Click the 'Edit' button to open the edit dialog" -- no data-testid or element identifier for the Edit button. The Element field lists `[data-testid='select-milestone']` and `[data-testid='btn-save']` but no edit trigger | -1 pt (Steps concrete) |
| TC-032 step 2, TC-034 step 2 | Same "Click the 'Edit' button" pattern with no selector | -0.5 pt each (Steps concrete) |
| TC-045 step 2 | Same "Click the 'Edit' button" pattern with no selector | -0.5 pt (Steps concrete) |
| TC-047 step 1 | "Verify `[data-testid='sidebar-link-milestones']` is visible in the sidebar, positioned between the 'Items' and 'Gantt' links" -- "positioned between" is a spatial assertion with no programmatic verification method | -0.5 pt (Steps concrete) |
| TC-038 step 3 | Mixes E-089/E-090 sitemap IDs with data-testid selectors in the same step: "Verify `[data-testid='columnheader-milestone']` is positioned between the 'title' (E-089) and 'priority' (E-090) columns" | -0.5 pt (Steps concrete) |
| TC-010 expected | "milestone count, item count, and progress percentage" -- no specific expected values to verify against. What counts? What percentage? | -0.5 pt (Expected results) |
| TC-013 expected | "each node shows name, date, status, and completion rate" -- no specific expected values | -0.5 pt (Expected results) |
| TC-023 expected | "the `[data-testid='milestone-node']` visual style updates to in-progress appearance" -- "visual style" is subjective and not objectively verifiable without a specific CSS class or attribute check | -0.5 pt (Expected results) |
| TC-021 preconditions | "Two PMs are editing the same milestone simultaneously" -- persists from iteration 2. Does not explain how to establish two concurrent sessions or what data state is required | -1 pt (Preconditions) |
| TC-044, TC-045, TC-046 preconditions | "Milestone [filter/selector/column] component build complete, integration task complete" -- these are build-process preconditions, not test data state. A tester cannot satisfy these preconditions | -0.5 pt each (Preconditions) |
| TC-038 step 3, TC-045 step 3 | Mix sitemap element IDs (E-089, E-090, E-037) with provisional data-testid selectors in element references | -0.5 pt each (Element accuracy) |
| TC-047 Type vs traceability | Body: Type: UI. Traceability table: Type: Integration. Cross-section inconsistency per rubric rule: -3 pts per conflict | -3 pts (Consistency) |
| TC-048 section vs type | Under "### Cross-Interface Integration Tests" section but Type: UI. Section header implies Integration classification | -1 pt (Consistency) |
| TC-044 through TC-046 | These verify component visibility only, not cross-interface behavior. TC-049..TC-052 address this but the visibility-only TCs are still classified as "Integration" | -1 pt (Integration scenarios) |
| Missing UI->API integration | All integration TCs (TC-049..TC-052) go API->UI direction only. No TC verifies that a UI action (e.g., creating a milestone via the UI form) results in correct API state | -1 pt (Integration scenarios) |
| TC-047 classification | Body says Type: UI, section says Integration, traceability says Integration. Three-way conflict | -1 pt (Classification) |
| TC-048 section placement | Placed under "Cross-Interface Integration Tests" header but is Type: UI. Misleading section organization | -0.5 pt (Classification) |

---

## Attack Points

### Attack 1: Route & Element Consistency -- TC-047 has a three-way classification conflict (UI vs Integration)

**Where**: TC-047 body (line 838): `**Type**: UI`. TC-047 in traceability table (line 1212): `| TC-047 | UF-1 Navigation Architecture | Integration | ui/milestones | P0 |`. Section header (line 788): `### Integration Tests (Existing Page Modifications)`.

**Why it's weak**: The same TC is classified as UI in its body, Integration in the traceability table, and lives under an Integration section. The rubric penalizes cross-section inconsistencies at -3 pts per conflict. This single TC creates ambiguity for the test script generator: should it be placed in the UI test suite or the Integration test suite? The Test ID (`ui/milestones/integration-navigation-link`) uses the `ui/` prefix, contradicting the traceability table's Integration classification.

**What must improve**: Pick one classification. If TC-047 is Integration (it tests navigation link existence, which is a cross-cutting concern), change the body Type to Integration and the Test ID prefix to `integration/`. If it is UI, move it out of the Integration section and update the traceability table.

### Attack 2: PRD Traceability -- Story 8 / AC-2 remains orphaned for 3 iterations

**Where**: PRD prd-user-stories.md, Story 8 / AC-2: "Given 列表视图已渲染，When 我按状态筛选选择'实施中'，Then 仅显示实施中状态的里程碑图卡片". No TC in test-cases.md has Source: "Story 8 / AC-2". TC-009 covers the identical filtering behavior but its Source says "Story 3 / AC-4".

**Why it's weak**: This was flagged in iteration 2 and remains unfixed. The traceability chain is broken: a reviewer auditing PRD coverage cannot find which TC validates Story 8 / AC-2 because no TC's Source field names it. The behavior IS tested (TC-009), but the documentation link is missing. This is a documentation integrity failure, not a coverage failure. It undermines the entire purpose of the Source field if ACs can be functionally covered but traceably orphaned.

**What must improve**: Add "Story 8 / AC-2" to TC-009's Source field, making it "Story 3 / AC-4, Story 8 / AC-2". This is a one-line fix that has been missed for two iterations.

### Attack 3: Step Actionability -- "Edit" button clicks in 4 TCs lack element selectors

**Where**: TC-031 step 2 (line 597): "Click the 'Edit' button to open the edit dialog". TC-032 step 2 (line 612): same. TC-034 step 2 (line 643): same. TC-045 step 2 (line 814): same. None of these steps include a data-testid, aria-label, or any selector for the Edit button. The Element fields in these TCs do not include an edit trigger element.

**Why it's weak**: The rubric requires steps to describe "a single, unambiguous user action" with specific element identification. "Click the 'Edit' button" is ambiguous: which Edit button? On the items detail page, there may be multiple buttons. A test script generator cannot automate this step because no selector is provided. The existing-page TCs (TC-031..TC-043) were the focus of iteration 2's Attack 1, and the Element fields were updated to use provisional data-testid selectors, but the Edit button selector was overlooked in the step descriptions.

**What must improve**: Define a provisional data-testid for the edit button on the items detail page (e.g., `[data-testid='btn-edit-item']`) in the document header element map, add it to the Element field of TC-031/032/034/045, and update the step text to use the selector.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (Iter 2): Step Actionability -- TC-031 to TC-043 have wrong element references (E-010, E-011, E-035, E-036, E-037, E-089 pointed to non-milestone elements) | ✅ | All existing-page TCs now use provisional data-testid selectors: `[data-testid='select-milestone']`, `[data-testid='filter-milestone']`, `[data-testid='columnheader-milestone']`, `[data-testid='cell-milestone']`. Defined in document header element map |
| Attack 2 (Iter 2): Completeness -- Integration scenarios only verify component visibility, not cross-interface behavior | ✅ (partial) | TC-049 through TC-052 added: real API->UI round-trip integration TCs (API create/bind/delete/cancel -> UI verification). However TC-044..TC-046 still only verify visibility. No UI->API direction tested |
| Attack 3 (Iter 2): Structure -- TC-038a breaks sequential ID pattern | ✅ | TC-038a eliminated. All 70 TCs are now TC-001 through TC-070, sequential with no gaps or non-numeric suffixes |

---

## Verdict

- **Score**: 80/100
- **Target**: 85/100
- **Gap**: 5 points
- **Step Actionability**: 19/25 (not blocking, but borderline)
- **Action**: Continue to iteration 4. Priority fixes: (1) resolve TC-047 three-way classification conflict (pick UI or Integration, apply consistently across body + traceability table + section), (2) add "Story 8 / AC-2" to TC-009 Source field, (3) define provisional `[data-testid='btn-edit-item']` and update TC-031/032/034/045 steps. These three fixes would recover approximately 5-7 points.

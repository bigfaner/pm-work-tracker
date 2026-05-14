---
date: "2026-05-13"
doc_dir: "docs/features/milestone-map/testing/"
iteration: "5"
target_score: "85"
evaluator: Claude (automated, adversarial)
---

# Test Cases Eval -- Iteration 5

**Score: 91/100** (target: 85)

```
+------------------------------------------------------------------+
|                  TEST CASES QUALITY SCORECARD                     |
+------------------------------------------------------------------+
| Dimension                    | Score    | Max      | Status       |
+------------------------------+----------+----------+--------------+
| 1. PRD Traceability          |  22.5    |  25      | OK           |
|    TC-to-AC mapping          |  6.5/9   |          |              |
|    Traceability table        |  8/8     |          |              |
|    Reverse coverage          |  8/8     |          |              |
+------------------------------+----------+----------+--------------+
| 2. Step Actionability        |  20      |  25      | OK           |
|    Steps concrete            |  8/9     |          |              |
|    Expected results          |  7.5/9   |          |              |
|    Preconditions explicit    |  4.5/7   |          |              |
+------------------------------+----------+----------+--------------+
| 3. Route & Element Accuracy  |  19      |  20      | OK           |
|    Routes valid              |  7/7     |          |              |
|    Elements identifiable     |  6/7     |          |              |
|    Consistency               |  6/6     |          |              |
+------------------------------+----------+----------+--------------+
| 4. Completeness              |  19.5    |  20      | OK           |
|    Type coverage             |  7/7     |          |              |
|    Boundary cases            |  7/7     |          |              |
|    Integration scenarios     |  5.5/6   |          |              |
+------------------------------+----------+----------+--------------+
| 5. Structure & ID Integrity  |  9.5     |  10      | OK           |
|    IDs sequential/unique     |  3.5/4   |          |              |
|    Classification correct    |  3/3     |          |              |
|    Summary matches actual    |  3/3     |          |              |
+------------------------------+----------+----------+--------------+
| TOTAL                        |  90.5    |  100     |              |
+------------------------------+----------+----------+--------------+
```

> Rounded total: 91/100

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| TC-047 Source | Source is "PRD UI Function 'UF-1' Navigation Architecture" -- references a UI function spec section, not a PRD user story acceptance criterion. Persisted since iteration 2 | -1 pt (TC-to-AC mapping) |
| TC-056, TC-058, TC-063, TC-065 Sources | Source is "PRD Spec Related Changes #N" -- a spec section reference describing implementation changes, not a user story AC. 4 TCs affected | -1 pt (TC-to-AC mapping) |
| TC-070 Source | Source is "PRD Spec Security Requirements" -- vague; should specify which permission codes and which entity | -0.5 pt (TC-to-AC mapping) |
| TC-010 step 3 | "milestone count, item count, and progress percentage" referenced without data-testid selectors -- no selectors for these sub-fields in provisional element map | -0.5 pt (Steps concrete) |
| TC-013 step 3 | "each `[data-testid='milestone-node']` displays name, planned completion date, status, and completion rate" -- no selectors for these sub-elements | -0.5 pt (Steps concrete) |
| TC-010 expected | "milestone count, item count, and overall progress visible" -- no specific expected values to verify against | -0.5 pt (Expected results) |
| TC-013 expected | "each node shows name, date, status, and completion rate" -- no specific expected values | -0.5 pt (Expected results) |
| TC-023 expected | "the `[data-testid='milestone-node']` visual style updates to in-progress appearance" -- "visual style" is subjective, not objectively verifiable without a CSS class or aria attribute | -0.5 pt (Expected results) |
| TC-035 expected | "no create/edit/delete buttons are functional or visible" -- conflates "not functional" (btn-create-map is disabled) and "not present" (btn-edit-map/btn-delete are not in DOM); two different verification strategies merged in one expected result | -0.5 pt (Expected results) |
| TC-021 preconditions | "Two PMs are editing the same milestone simultaneously" -- does not explain how to establish two concurrent sessions or what data state is required. Persisted since iteration 2 | -1 pt (Preconditions) |
| TC-044/045/046 preconditions | "component build complete, integration task complete" -- build-process precondition, not test data state | -1.5 pts (Preconditions, 0.5 each) |
| TC-010 step 3 | "milestone count, item count, and progress percentage" -- no data-testid selectors for these sub-fields | -0.5 pt (Elements identifiable) |
| TC-013 step 3 | "name, planned completion date, status, and completion rate" -- no data-testid selectors for these sub-fields | -0.5 pt (Elements identifiable) |
| Cross-Interface Integration section | All integration TCs (TC-049..TC-052) go API-to-UI direction only; no dedicated integration TC verifying UI action resulting in correct API state within the integration section. (TC-031/032 cover this as UI TCs but the dedicated integration section lacks bidirectional coverage) | -0.5 pt (Integration scenarios) |
| TC-048/047/044 ordering | IDs not sequential in document order: TC-048 (line 696) appears before TC-038 (line 713), TC-047 (line 806) appears before TC-044 (line 823). All 70 IDs exist but out of document sequence | -0.5 pt (IDs sequential) |

---

## Attack Points

### Attack 1: Step Actionability -- TC-021 preconditions still lack concurrent session setup instructions (persisted since iteration 2)

**Where**: TC-021 preconditions (line 421): "Two PMs are editing the same milestone simultaneously; user has `milestone:update` permission". Steps (lines 424-427): "In browser tab A, open detail-panel and edit" / "In browser tab B, open detail-panel for the same milestone and save first" / "In tab A, click btn-save".

**Why it's weak**: The preconditions say "Two PMs are editing the same milestone simultaneously" but do not explain how to establish this state. A test script generator needs to know: (a) how to create two authenticated sessions, (b) what API calls or UI actions establish the "editing" state in both tabs before step 1 begins, (c) whether two different users or one user with two tabs is required. The AC (Story 4b AC-2) says "two PMs simultaneously edit" which implies two distinct users, but the steps say "browser tab A" and "browser tab B" which implies one user. This ambiguity has persisted since iteration 2 and makes the TC non-actionable for automated test generation.

**What must improve**: Rewrite preconditions to specify: (a) two distinct user accounts with milestone:update permission, (b) a milestone exists with known initial data, (c) both users have navigated to the detail panel and loaded the same milestone. Alternatively, specify a single user with two Playwright browser contexts. Add a step 0 for each tab that loads the milestone data before the edit begins.

### Attack 2: Step Actionability -- 3 Integration TCs have build-process preconditions instead of test data state

**Where**: TC-044 preconditions (line 828): "Milestone filter component build complete, integration task complete". TC-045 preconditions (line 843): "Milestone selector component build complete, integration task complete". TC-046 preconditions (line 859): "Milestone column component build complete, integration task complete".

**Why it's weak**: These preconditions describe build/integration task completion -- a process concern, not test data state. A test script cannot verify "build complete" at runtime. Preconditions should specify the required data state: e.g., "Team has at least 3 milestones created via API", "At least one MainItem exists with milestone_key set", etc. The current preconditions are tautological for a CI environment (if the code isn't deployed, no test runs) and unhelpful for a test data setup phase.

**What must improve**: Replace build-process preconditions with concrete test data requirements. For TC-044: "Team has 2+ milestones; /items page is accessible". For TC-045: "A MainItem exists; a milestone exists in the team". For TC-046: "Table view has MIs; at least one MI is bound to a milestone via API".

### Attack 3: PRD Traceability -- 6 TCs have non-AC Source references

**Where**: TC-047 Source (line 807): "PRD UI Function 'UF-1' Navigation Architecture". TC-056 Source (line 977): "PRD Spec Related Changes #1". TC-058 Source (line 999): "PRD Spec Related Changes #1". TC-063 Source (line 1060): "PRD Spec Related Changes #2". TC-065 Source (line 1082): "PRD Spec Related Changes #2". TC-070 Source (line 1144): "PRD Spec Security Requirements".

**Why it's weak**: The rubric requires Source to point to "a specific PRD acceptance criterion" at "AC-3.1 level specificity." These 6 TCs reference spec sections, UI function identifiers, or security requirement categories rather than user story ACs. While the referenced content exists in the PRD documents, the traceability chain is broken -- a reviewer cannot navigate from the TC back to a specific Given/When/Then acceptance criterion. TC-047 tests navigation which is covered by UF-1 but has no user story AC. TC-056/058/063/065 test Get-by-ID and Delete operations implied by the data model but not explicitly stated as ACs. TC-070 tests permission denial implied by security requirements but without mapping to specific permission-entity combinations.

**What must improve**: For TC-047, either create an explicit AC in the user stories for navigation or accept the non-AC source with documentation. For TC-056/058/063/065, map to the nearest applicable AC (e.g., TC-058 could reference Story 4c AC-1 which describes deletion behavior, or Story 1/2 context). For TC-070, specify which permission codes are tested: "PRD Spec Security Requirements (milestone:create/update/delete/read)".

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (Iter 4): Summary table counts wrong (UI=46, Integration=6 vs actual UI=45, Integration=7) | YES | Summary table (lines 58-62) now reads UI=45, Integration=7, API=18, Total=70. Body and traceability both confirm these counts |
| Attack 2 (Iter 4): TC-048 is UI test placed under "Cross-Interface Integration Tests" section | YES | TC-048 (line 696) is now under "### Permission-Based Views" within "## UI Test Cases". Type: UI in body, traceability, and section all agree |
| Attack 3 (Iter 4): 6 TCs mix sitemap element IDs (E-XXX) with data-testid selectors in spatial assertions | YES | TC-038 step 3 (line 724): now uses DOM-index check "headerRow.locator('th').nth(2)". TC-044 step 2 (line 833): uses parent container "within the filter bar container". TC-045 step 3 (line 849): uses "within the edit dialog form container". TC-046 step 2 (line 864): uses "within the `<thead>` row at the expected DOM index". No more spatial assertions with E-XXX IDs |

---

## Verdict

- **Score**: 91/100
- **Target**: 85/100
- **Gap**: target exceeded by 6 points
- **Step Actionability**: 20/25 (meets blocking threshold of 20)
- **Action**: Target reached. Remaining issues (TC-021 preconditions, TC-044/045/046 preconditions, non-AC source references, TC-010/013 sub-element selectors) are minor and do not block downstream test script generation.

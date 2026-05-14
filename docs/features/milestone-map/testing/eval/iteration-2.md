---
date: "2026-05-13"
doc_dir: "docs/features/milestone-map/testing/"
iteration: "2"
target_score: "85"
evaluator: Claude (automated, adversarial)
---

# Test Cases Eval -- Iteration 2

**Score: 73/100** (target: 85)

```
┌─────────────────────────────────────────────────────────────────┐
│                  TEST CASES QUALITY SCORECARD                     │
├──────────────────────────────────────────────────────────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. PRD Traceability          │  20      │  25      │ ⚠️         │
│    TC-to-AC mapping          │  8/9     │          │            │
│    Traceability table        │  7/8     │          │            │
│    Reverse coverage          │  5/8     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Step Actionability        │  18      │  25      │ ⚠️         │
│    Steps concrete            │  6/9     │          │            │
│    Expected results          │  8/9     │          │            │
│    Preconditions explicit    │  4/7     │          │            │
├──────────────────────────────┼──────────┼──────────┤            │
│ 3. Route & Element Accuracy  │  15      │  20      │ ⚠️         │
│    Routes valid              │  6/7     │          │            │
│    Elements identifiable     │  5/7     │          │            │
│    Consistency               │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┤            │
│ 4. Completeness              │  13      │  20      │ ⚠️         │
│    Type coverage             │  6/7     │          │            │
│    Boundary cases            │  6/7     │          │            │
│    Integration scenarios     │  1/6     │          │            │
├──────────────────────────────┼──────────┼──────────┤            │
│ 5. Structure & ID Integrity  │  7       │  10      │ ⚠️         │
│    IDs sequential/unique     │  3/4     │          │            │
│    Classification correct    │  2/3     │          │            │
│    Summary matches actual    │  2/3     │          │            │
├──────────────────────────────┼──────────┼──────────┤            │
│ TOTAL                        │  73      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| TC-047 Source | Source is "UF-1 Navigation Architecture" which is a UI function reference, not a PRD acceptance criterion. Source should trace to a PRD user story or spec section | -1 pt (Traceability) |
| Traceability table TC-047 | Source column says "UF-1 Navigation Architecture" -- not a standard PRD AC reference format, inconsistent with other entries | -1 pt (Traceability) |
| Story 8 / AC-2 | PRD says "filter by status shows only in-progress cards" on the milestone map list. TC-009 covers Story 3/AC-4 which is the same filtering behavior. But Story 8/AC-2 is not directly mapped by any TC Source. Story 8/AC-2 reads: "When I filter by status selecting 'in-progress', then only in-progress milestone map cards are shown" -- functionally identical to TC-009 but traceability is broken because no TC's Source field says "Story 8 / AC-2" | -1 pt (Reverse coverage) |
| PRD Spec Security Requirements | TC-065 covers API permission tests but only tests milestone-level permissions. The PRD also specifies milestone-map-level RBAC ("milestone:create/update/delete/read" applies to both maps and milestones). TC-065 Source says "PRD Security Requirements" which is vague -- should specify which permission code and which entity | -1 pt (Reverse coverage) |
| TC-031 steps | Step 1: "Open main item edit dialog" -- no element identifier. Step 2: "Select a milestone from the milestone dropdown" -- which dropdown? Element field says "E-035, E-036, E-037" but none of these map to a milestone selector (E-037 is "assignee" per sitemap.json, not milestone). Step 3: "Click save" -- no element identifier | -3 pts (Steps concrete) |
| TC-032 steps | Same vagueness: "Open main item edit dialog", "Clear the milestone field (select 'unassigned')", "Click save" -- none reference actual selectors | -1 pt (Steps concrete, overlapping with TC-031 pattern) |
| TC-033 steps | "Select a specific milestone from the milestone filter dropdown" -- E-010 is "status" and E-011 is "assignee" per sitemap.json. Neither is a milestone filter. The Element field references wrong sitemap elements | -1 pt (Steps concrete) |
| TC-038 to TC-043 steps | Steps like "Verify the table header row contains a 'milestone' column positioned between 'title' and 'priority'" and "Use milestone column filter to select a specific milestone" -- no data-testid or selector. Element field says "E-089" but sitemap.json maps E-089 to "columnheader name='标题'" (title column), not milestone column. Wrong element reference | -1 pt (Steps concrete) |
| TC-022 step 3 | "Click 'Confirm' in the confirmation dialog" -- no data-testid or element identifier for the confirmation dialog or its Confirm button | -1 pt (Steps concrete) |
| TC-033 Element | E-010 (status filter) and E-011 (assignee filter) are listed as elements but the test is about milestone filtering. No milestone filter element exists in sitemap.json | -2 pts (Element accuracy) |
| TC-031 Element | E-035 is "编辑主事项" heading, E-036 is "优先级" combobox, E-037 is "负责人" combobox per sitemap.json. None of these are the milestone selector. The Element field references elements that do not match the test action | -1 pt (Element accuracy) |
| TC-038 Element | E-089 is the "标题" columnheader, not a milestone column. The milestone column does not yet exist in sitemap.json | -1 pt (Element accuracy) |
| TC-031, TC-032, TC-034 preconditions | No precondition specifies that the test requires an existing MainItem with a specific state (e.g., "an unassigned MainItem exists"). TC-031 says "A MainItem is not assigned to any milestone" but does not specify how to create or ensure this state | -1 pt (Preconditions) |
| TC-021 preconditions | "Two PMs are editing the same milestone simultaneously" -- this is a complex multi-session setup that is not described in sufficient detail. How are two sessions established? What data state is required? | -1 pt (Preconditions) |
| TC-038 preconditions | "Milestone feature is available" -- vague. What does "available" mean? API deployed? Feature flag enabled? | -1 pt (Preconditions) |
| TC-038a | ID "TC-038a" breaks the TC-NNN sequential pattern. It introduces a non-numeric suffix that makes programmatic parsing unreliable | -1 pt (IDs sequential) |
| Integration TCs under UI section | TC-044 through TC-047 have Type: Integration but are placed under "## UI Test Cases" section. No separate "## Integration Test Cases" top-level section exists. The subsection "### Integration Tests (Existing Page Modifications)" is nested under UI | -1 pt (Classification) |
| Integration scenarios | TC-044 through TC-047 verify that components are *visible* on existing pages. They do not test cross-interface scenarios (e.g., UI action triggers API call, API state change reflects in UI). No TC verifies that creating a milestone via API causes it to appear in the items page filter, or that deleting a milestone via API removes it from the table view | -5 pts (Integration scenarios) |
| TC-007 | Expected result says dropdown shows "pending-implementation" and "completed". PRD Story 3/AC-2 uses Chinese status names (待实施/已完成). The TC uses English translations that are inconsistent with the PRD's original terminology. While the UI may use English internally, the TC should document which status values appear to avoid ambiguity | -1 pt (Expected results) |

---

## Attack Points

### Attack 1: Step Actionability -- Existing-page TCs (TC-031 to TC-043) have wrong or missing element references

**Where**: TC-031 Element field: `E-035, E-036, E-037`. TC-033 Element field: `E-010, E-011`. TC-038 Element field: `E-089`. Per sitemap.json, E-010="状态" (status filter), E-011="负责人" (assignee filter), E-035="编辑主事项" (edit heading), E-036="优先级" (priority combobox), E-037="负责人" (assignee combobox), E-089="标题" (title columnheader). None of these are milestone-related elements. The milestone filter, milestone selector, and milestone column do not exist in sitemap.json yet because they are new additions.

**Why it's weak**: 13 TCs (TC-031 through TC-043) reference sitemap element IDs that point to existing (non-milestone) elements. A test script generator would target the wrong UI elements. The Element field is supposed to identify the actual interaction target, but here it identifies adjacent elements that have nothing to do with milestone functionality. Steps like "Select a milestone from the milestone dropdown" cannot be automated because no milestone dropdown element ID exists.

**What must improve**: Either (a) update sitemap.json with the new milestone-related elements (milestone filter dropdown, milestone selector in edit dialog, milestone column in table) and reference those IDs, or (b) use provisional `data-testid` selectors as done for the /milestones page TCs. The current approach of referencing unrelated existing elements is worse than no element at all because it points automation at the wrong target.

### Attack 2: Completeness -- Integration scenarios only verify component visibility, not cross-interface behavior

**Where**: TC-044 through TC-047 are the only "Integration" TCs. TC-044: "Verify milestone filter dropdown is visible in the filter bar area". TC-045: "Verify 'Milestone' selector is visible below the 'Assignee' field". TC-046: "Verify 'Milestone' column header is visible". These verify *rendering*, not *integration behavior*.

**Why it's weak**: The rubric defines integration scenarios as "TCs cover cross-feature or cross-interface scenarios (e.g., UI action triggers API call, CLI command affects UI state)". The current "integration" TCs are glorified smoke tests -- they check that a component appears on the page. No TC verifies: (1) creating a milestone via API then confirming it appears in the items page filter, (2) binding an MI to a milestone via API then checking the table view updates, (3) deleting a milestone via API and verifying the items page shows the MI as unassigned, (4) cancelling a milestone and verifying the table view shows "--" for affected MIs. These are the actual cross-interface integration scenarios the rubric demands.

**What must improve**: Add integration TCs that exercise the full UI-to-API-to-UI round trip. Each integration TC should: perform an action through one interface, then verify the result through another. For example: "POST /api/v1/teams/:teamId/milestones/:id/status with {status: 'cancelled'} then navigate to /table and verify affected MIs show '--' in milestone column."

### Attack 3: Structure -- TC-038a breaks sequential ID pattern, creating ambiguity in traceability

**Where**: TC-038a appears between TC-037 and TC-038. The ID "TC-038a" uses a non-numeric suffix, deviating from the TC-NNN pattern used by all other 65 TCs.

**Why it's weak**: The rubric requires "IDs follow the pattern (e.g., TC-001, TC-002...). No gaps, no duplicates, no re-used IDs." TC-038a breaks the sequential integer pattern. It creates ambiguity: does "TC-038..TC-043" in the route validation table include TC-038a? The range notation on line 1099 says "TC-038..TC-043" but TC-038a (which targets /milestones, not /table) would be missed by a naive range parser. Additionally, TC-038a appears at line 624 before TC-038 at line 641, so document order contradicts ID order. This undermines the document's machine-parseability.

**What must improve**: Renumber TC-038a to a proper sequential ID. The simplest fix: TC-038a should become TC-044, and all subsequent TCs should shift by one (old TC-044 becomes TC-045, etc., up to TC-066). Alternatively, TC-038a could be moved to its logical position after TC-065 as TC-066. Either way, the "a" suffix must be eliminated.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1 (Iter 1): Route & Element Accuracy -- 29 TCs have `sitemap-missing` elements | ✅ | All /milestones page TCs (TC-001 to TC-030, TC-035 to TC-038a, TC-047) now have provisional `data-testid` selectors defined in the document header element map |
| Attack 2 (Iter 1): Reverse Coverage -- Story 11 ACs are orphaned | ✅ | TC-035 now maps to "Story 9 / AC-1, AC-2, Story 11 / AC-1". TC-036 maps to "Story 9 / AC-3, Story 11 / AC-3". TC-037 maps to "Story 11 / AC-2". New TC-038a maps to "Story 11 / AC-4" |
| Attack 3 (Iter 1): Step Actionability -- vague "Observe" steps | ✅ | All "Observe" steps have been replaced with concrete "Verify" steps that reference specific elements and expected values |

---

## Verdict

- **Score**: 73/100
- **Target**: 85/100
- **Gap**: 12 points
- **Step Actionability**: 18/25 (not blocking, but borderline)
- **Action**: Continue to iteration 3. Priority fixes: (1) fix wrong element references in TC-031 to TC-043 (use provisional data-testid selectors matching the milestone page pattern), (2) add real cross-interface integration TCs (API action -> UI verification round-trips), (3) renumber TC-038a to eliminate the non-sequential ID.

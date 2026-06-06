---
iteration: 2
scorer: adversary
total_score: 830
pass: false
date: "2026-06-04"
---

# Journey Eval Report: list-filtering-and-sorting (Iteration 2)

**Date**: 2026-06-04
**Journey**: `list-filtering-and-sorting`
**Surface Types**: web, api

---

## Dimension Scores

| Dimension | Score | Min | Status |
|-----------|-------|-----|--------|
| 1. Completeness | 150/200 | 120 | PASS |
| 2. Semantic Purity | 165/200 | 120 | PASS |
| 3. Precondition Exclusivity | 125/150 | 90 | PASS |
| 4. Fact Alignment | 115/150 | 90 | PASS |
| 5. Surface Fitness | 115/150 | 90 | PASS |
| 6. Internal Consistency | 60/150 | 90 | **FAIL** |
| **Total** | **830/1000** | **850** | **FAIL** |

---

## 1. Completeness -- 150/200 (min 120) -- PASS

### Journey Metadata: 42/50

- Name `list-filtering-and-sorting` is properly kebab-case.
- Risk level "Medium" is explicitly justified in the HTML comment block ("multi-step interaction without irreversible side effects"). Well reasoned.
- Surface types `["web", "api"]` are correctly declared.
- Sources section lists three source documents with specific sections cited (Story 8, Story 9, PRD spec #10/#11/#12, API Handbook).
- **Deduction (-8)**: The generated date is present but no `author` or `version` field. Minor metadata gap.

### Steps Complete: 65/80

- 7 happy-path steps + 9 edge-case steps = 16 steps total. Each step has a User Action and Expected Result.
- Steps cover a coherent sequence: assignee filter penetration (Steps 1, 6), terminal sorting (Step 2), progress page default (Step 3), clearing filters (Step 4), empty state (Step 5), API assignee filter (Step 6), API multi-status (Step 7).
- Edge cases cover: no filters (E1), assignee without sub-items (E2), additional status (E3), unauthorized (E4), unauthenticated (E5), validation error API (E6), session expired web (E7), validation error web (E8), non-existent team (E9).
- PRD Story 8 AC1-AC3 all covered. PRD Story 9 AC1-AC3 all covered.
- **Deduction (-10)**: Step 4 ("Clear all status filters") and E1 ("No filters selected shows all items") test conceptually similar behavior. E1 tests initial state with no filters on the item list page, while Step 4 tests deselecting all filters on the progress page. The distinction is valid (different pages), but the journey does not explicitly test deselecting all filters on the *item list page* vs *progress page*. Step 4 is progress-page-specific, and E1 is list-page-specific. The "no filters = all items" invariant is covered but only for initial load, not for the explicit deselection flow on the item list page.
- **Deduction (-5)**: No step tests the Gantt view's enhanced status filter (API Handbook documents `GET /teams/:teamId/views/gantt` with enhanced `status` param). While this may be scoped to a different journey, the journey's title is "list-filtering-and-sorting" and the API Handbook entry is listed as a source. If the Gantt view is intentionally excluded, a note in the journey stating this exclusion would strengthen completeness.

### Outcomes Coverage: 43/70

- Mandatory derived outcomes for web: `validation-error` present (E8), `session-expired` present (E7). Both covered.
- Mandatory derived outcomes for api: `unauthorized` present (E4, E5). Covered.
- Common boundary outcomes: `not-found` present (E9 for non-existent team). `validation-error` for API present (E6).
- Empty state outcome present (Step 5) with correct Chinese text "没有符合条件的事项" and "清除过滤条件" button -- matching PRD spec exactly.
- **Deduction (-15)**: No outcome covers what happens when an assignee filter is applied on the *progress page* (only status filter is described for progress page). The PRD spec #12 says progress page supports status filter, but Story 8 AC1 mentions assignee filter on the "事项清单页面" (item list page). The journey correctly scopes assignee filter to the item list page, but the progress page's interaction with assignee filter is untested. This is a minor coverage gap.
- **Deduction (-12)**: No performance boundary outcome. PRD spec mentions "filter response <= 500ms (1000 main + 5000 sub-items benchmark)" but no step or edge case tests this. For a "Medium" risk journey this may be acceptable, but the journey sources cite the PRD spec which includes this requirement. A note acknowledging the exclusion of non-functional requirements would mitigate this.

---

## 2. Semantic Purity -- 165/200 (min 120) -- PASS

### Natural Language Outcomes: 70/80

- Outcomes describe user-observable states well: "results include main items where A is the direct assignee AND main items where A is the assignee of a sub-item" (Step 1). "Empty state message is displayed; a action button is available to reset all filters" (Step 5).
- No CSS selectors, XPath, or framework-specific assertions in the expected results.
- **Deduction (-5)**: Step 6 and Step 7 use API field names (`matchType`, `matchedSubItemIds`) in expected results. While these are surface-api steps and API field names are the user-observable output of an API (unlike internal component names), using raw JSON field names still leans toward implementation language rather than semantic description. For API steps this is borderline acceptable but technically impure -- a semantic description would say "each item indicates whether the match was direct or indirect and lists the matching sub-item identifiers." Minor deduction.
- **Deduction (-5)**: Step 6 references a URL pattern `GET /teams/:teamId/main-items?assigneeKey=<A_bizKey>` in the user action. For API surface steps, the URL is the "action" so this is somewhat necessary, but the expected result references `matchType: "direct"` with code-style formatting.

### Declarative Preconditions: 55/60

- Preconditions are predominantly declarative state descriptions: "Assignee A is the direct assignee of some main items AND the assignee of sub-items under other main items" (Step 1). "Filter criteria are applied that match no items in the system" (Step 5). "No status or assignee filters are selected" (E1).
- The Setup section uses declarative state: "Multiple main items exist with various statuses", "Assignee A is responsible for some main items directly and some sub-items under other main items."
- **Deduction (-5)**: Step 3 precondition "PM user has never opened the progress page before in this session" is borderline procedural -- it describes a sequence of actions the user has NOT taken rather than a system state. A more declarative version would be "The progress page has no stored filter preferences for the current session."

### No Implementation Coupling: 40/60

- No HTTP status codes in web steps. Good.
- No CSS/XPath selectors anywhere. Good.
- No framework or component names in web steps. Good.
- **Deduction (-20)**: API steps (Steps 6, 7, E4, E5, E6, E9) necessarily include HTTP methods, URL patterns, and response field names. While this is inherent to the api surface type, the degree of coupling is high. Step E6 describes "A list API request contains invalid filter values (e.g., non-existent status value, malformed assigneeKey)" -- mentioning specific parameter names like `assigneeKey` is implementation-level detail. A less coupled version would say "A request contains filter values that are not recognized by the system." Similarly E9 mentions `GET /teams/<non-existent-teamId>/main-items` with specific URL structure. This is a structural tension in the rubric for dual-surface journeys -- API steps will inherently be more coupled. However, the rubric asks for no implementation coupling, and the API steps do embed implementation details. The deduction is mitigated by the web steps being clean.

---

## 3. Precondition Exclusivity -- 125/150 (min 90) -- PASS

### Preconditions Distinct Across Outcomes: 50/60

- Each step has a single precondition and a single expected outcome, making exclusivity straightforward.
- Steps that test similar features have distinct preconditions: Step 1 (assignee A with sub-items) vs E2 (assignee B without sub-items). Step 3 (first time opening progress page) vs E3 (progress page loaded with default filter, then adding more).
- **Deduction (-10)**: E4 ("A user without main_item:read permission sends a list API request") and E5 ("A list API request is sent without valid credentials") test different authentication/authorization failure modes. However, the preconditions overlap semantically: both describe "a request that should not be allowed." They are distinguished by the specific failure mode (no permission vs no credentials), which is adequate but could be more explicit. E4 should clarify that the user IS authenticated but lacks the specific permission; E5 should clarify that the user has NO valid authentication token. The current wording of E4 ("A user without main_item:read permission") does not explicitly state the user is authenticated, creating ambiguity with E5.

### Preconditions Sufficient to Uniquely Select Outcome: 45/50

- Most preconditions are specific enough to determine a unique outcome. Step 1's precondition fully specifies the trigger state (assignee A has both direct and indirect assignments). Step 5's precondition ("Filter criteria are applied that match no items in the system") fully specifies the empty state trigger.
- **Deduction (-5)**: Step E8's precondition ("The filter controls allow free-text input for assignee search") describes a UI capability rather than a system state that leads to the error outcome. The actual trigger for the validation error is the user entering invalid input, which is the action, not the precondition. The precondition should describe what makes the system accept and then reject the input -- e.g., "The assignee filter field accepts text input but validates against known assignee identifiers."

### No Missing Preconditions for Error/Boundary: 30/40

- Error outcomes (E4, E5, E6, E7, E8, E9) all have explicit preconditions.
- Boundary outcomes (E1 no filters, E2 assignee without sub-items, E3 additional status) have preconditions.
- **Deduction (-10)**: E7 ("Session expired during filter interaction") has a precondition that says "The user's session has expired while the filter controls are displayed" -- but does not specify what constitutes a session expiry in this context (timeout duration, server-side invalidation). More critically, it does not specify that the user was previously authenticated and performing an operation, which is the key trigger state for the session-expired outcome. A reader cannot set up this test case without knowing what "session expired" means in practice.

---

## 4. Fact Alignment -- 115/150 (min 90) -- PASS

### Factual Claims Traceable: 50/60

- Fact annotations are present throughout the journey using HTML comments: `<!-- fact: prd-spec #10, Story 8 AC1 -->`, `<!-- fact: prd-spec #11, Story 9 AC1 -->`, `<!-- fact: prd-spec #12, Story 9 AC2 -->`, `<!-- fact: Story 9 AC3 -->`, `<!-- fact: prd-spec empty state handling -->`, `<!-- fact: api-handbook Enhanced List Main Items -->`, `<!-- fact: api-handbook Enhanced Query Parameters -->`.
- All major claims are traced to PRD stories, PRD spec sections, or API handbook entries.
- **Deduction (-10)**: Some fact annotations are imprecise. "prd-spec #10" refers to the PRD spec item #10, but Step 1's claim about "因子事项匹配" visual indicator is actually from Story 8 AC2, not #10. #10 is about the status multi-select + assignee filter penetration. The indicator is part of #10's description but the story reference should also cite AC2 specifically. Step 2 cites "prd-spec #11, Story 9 AC1" which is correct and precise. Step 7 cites "api-handbook Enhanced Query Parameters" which is a section reference but not a specific claim -- the handbook does not explicitly say "terminal status main items are sorted to the bottom of the response list." That claim comes from PRD spec #11/Story 9 AC1, not the API handbook. The fact annotation is misattributed.

### Inferred Claims Have Source: Inferred: 40/50

- Steps E4, E5, E6, E7, E8, E9 are marked with `<!-- source: inferred -- derived from API surface unauthorized mandatory outcome -->` or similar annotations. These correctly identify why the outcome exists.
- **Deduction (-10)**: The inferred annotations vary in quality. E4 and E5 both say "derived from API surface `unauthorized` mandatory outcome" but do not explain the reasoning chain (e.g., "the API handbook requires main_item:read permission, so a user without this permission must receive an authorization error"). E7 says "derived from Web surface `session-expired` mandatory outcome" but does not explain why session expiry is plausible in this journey (the filtering workflow is interactive and could span time). The annotations state *what* rule generated the outcome but not *why* it applies to this specific journey context.

### No Hallucinated Claims: 25/40

- Most claims are well-grounded. The empty state message "没有符合条件的事项" matches the PRD spec exactly. The "清除过滤条件" action matches the PRD spec. The "因子事项匹配" indicator matches Story 8 AC2. Terminal status definition (closed, completed) matches the PRD spec terminal status definition section.
- **Deduction (-15)**: Step 1's expected result says "only the matching sub-items are shown under those main items." This claim -- that non-matching sub-items are hidden when a main item is shown due to indirect match -- is not explicitly stated in the PRD stories or spec. Story 8 AC1 says "展示 A 负责的主事项 + 含 A 负责子事项的主事项（连同该子事项一起展示）" which says "along with that sub-item" but does not say *only* the matching sub-items are shown. The API handbook says `matchedSubItemIds` is returned for indirect items, which implies filtering, but this is a design-level claim, not a PRD requirement. The journey states "only the matching sub-items are shown" as a definitive fact without marking it as inferred or derived from the API handbook. This is an unclassified inference.
- **Deduction**: Step 3 says terminal status items are "filtered out" by the default "进行中" filter. The PRD Story 9 AC2 says "仅展示进行中状态的主事项" -- it says only in-progress items are shown, which inherently excludes terminal items, but the journey adds the explicit statement "terminal status items are filtered out" which is a correct inference from the default but is stated as fact rather than inference.

---

## 5. Surface Fitness -- 115/150 (min 90) -- PASS

### Mandatory Derived Outcomes Present: 50/60

- Web `validation-error`: E8 covers validation error on filter parameters. Present.
- Web `session-expired`: E7 covers session expired during filter interaction. Present.
- API `unauthorized`: E4 (no permission) and E5 (no credentials) both cover unauthorized. Present.
- **Deduction (-10)**: The `validation-error` outcome for web (E8) is somewhat weak. The precondition says "The filter controls allow free-text input for assignee search" -- this assumes the UI has a free-text input for assignee filtering. The PRD stories describe assignee filtering as a selection ("选中负责人 A 进行过滤") not a free-text search. If the actual UI uses a dropdown or autocomplete with no free-text validation, E8's premise is incorrect. The outcome is present but its precondition may not match the real UI. This is a surface fitness concern because the web validation-error outcome must be realistic for the declared surface type.

### Test Strategy Proportions: 35/50

- The journey has 9 web steps (Steps 1-5, E1-E3, E7-E8) and 5 API steps (Steps 6-7, E4-E6, E9), plus 1 mixed (E3 references both). Wait -- let me recount:
  - Web: Steps 1, 2, 3, 4, 5, E1, E2, E3, E7, E8 = 10 web steps
  - API: Steps 6, 7, E4, E5, E6, E9 = 6 API steps
- Ratio is approximately 63% web / 37% API. For a dual-surface journey, this is within acceptable bounds.
- **Deduction (-15)**: The API steps are thinner than the web steps. Steps 6 and 7 cover the two main API behaviors (assignee filter, multi-status filter) but do not test combined filters (assignee + status together), which is a common API use case. The web steps test combined behaviors (Step 4 combines clearing filters with terminal sorting). The API coverage is adequate but not as thorough as the web coverage.

### Surface-Specific Environment: 30/40

- Web assumptions are realistic: filter selection, checkbox interaction, page navigation, visual indicators. The "因子事项匹配" indicator is described as a visual element, appropriate for web.
- API assumptions include authenticated requests, team scoping, and query parameters. Appropriate for API surface.
- **Deduction (-10)**: Step E8 (web validation error) assumes the assignee filter allows free-text input that can contain special characters or extremely long strings. The PRD describes assignee selection as choosing from a list ("选中负责人"), not free-text entry. If the real implementation uses a dropdown/autocomplete, E8's scenario is implausible. The surface-specific environment should match the actual UI interaction model.

---

## 6. Internal Consistency -- 60/150 (min 90) -- FAIL

### Invariants Hold in Every Step: 25/60

Five invariants are declared. Checking each against every step:

1. **"Assignee filter always penetrates to sub-item level"** -- Holds in Step 1, Step 6, E1 (no filter, not applicable), E2 (assignee B has no sub-items, so penetration occurs but finds nothing -- the invariant says "always penetrates" which is consistent). **OK.**

2. **"Terminal status main items always sort to the bottom of item lists, regardless of active filters"** -- Checking:
   - Step 2: Terminal items sort to bottom. OK.
   - Step 3: Progress page defaults to "进行中", terminal items filtered out. **POTENTIAL VIOLATION**: The invariant says terminal items "always sort to the bottom of item lists." Step 3 says terminal items are "filtered out" -- they are not shown at all, not sorted to the bottom. The invariant should distinguish between "when terminal items are visible, they sort to bottom" vs "terminal items are always shown but sorted." Step 3's behavior (filtering out terminal items) is consistent with PRD Story 9 AC2 but **inconsistent with the invariant as written**.
   - Step 4: All items displayed, terminal items "still sort to the bottom." OK, explicitly confirmed.
   - Step 7: API response, terminal items sorted to bottom. OK.
   - E3: Both in-progress and completed shown, terminal sort to bottom. OK.

   **Deduction**: Invariant 2 as written says "always sort to the bottom" but Step 3 hides terminal items entirely. The invariant and Step 3's behavior are contradictory. -10.

3. **"When no filters are selected, all items are displayed"** -- E1 confirms this for the item list page. Step 4 confirms for progress page when all checkboxes are deselected. **OK.**

4. **"Empty filter results always show [message] with [action]"** -- Step 5 confirms. **OK.**

5. **"Sub-item match indicator is displayed whenever a main item is shown due to sub-item filter match"** -- Step 1 confirms. E2 confirms absence when no sub-item match. **OK.**

**Deduction (-15)**: Invariant 2 contradicts Step 3. The invariant needs to be rewritten to scope the sorting behavior to "when terminal items are visible in the list" or Step 3 needs to acknowledge that the default progress page filter hides terminal items before sorting applies.

### Cross-Step References Consistent: 20/50

- The Setup section defines: multiple main items with various statuses, Assignee A (direct + sub-item), Assignee B (direct only), at least one terminal main item, a user without permission.
- Steps reference these consistently: Step 1 uses Assignee A, Step 6 uses Assignee A, E2 uses Assignee B, E4 uses the user without permission.
- **Deduction (-15)**: Steps 6 and 7 reference API endpoints with `:teamId` as a path parameter but the Setup section does not define a specific team. The setup says "Multiple main items exist" but does not establish a team context for the API steps. API steps implicitly assume a team exists but this is not declared in setup.
- **Deduction (-10)**: Step 3 references "PM user has never opened the progress page before in this session" -- this is a temporal precondition about the user's history within a session. No other step references session state, and the Setup section does not establish session-related state. This creates an isolated reference that is not grounded in the setup data.
- **Deduction (-5)**: Step E3 references "completed main items exist" in its precondition but does not specify whether these are the same terminal items defined in Setup or different ones. While likely the same, the reference is ambiguous.

### Risk Level Consistent with Content: 15/40

- Risk level "Medium" is justified as "multi-step interaction without irreversible side effects." This is reasonable.
- **Deduction (-15)**: The journey covers 16 steps across 2 surface types with 9 edge cases including authentication, authorization, session management, and input validation scenarios. The breadth of error scenarios (unauthorized, unauthenticated, session-expired, validation error, not-found) suggests a higher complexity than a simple "Medium" classification implies. A journey with this many error paths across two surface types could reasonably be classified as "Medium-High" or have a more detailed risk justification that addresses the error scenario coverage.
- **Deduction (-10)**: The risk justification only mentions "multi-step interaction without irreversible side effects" -- it does not address the dual-surface complexity or the authentication/authorization scenarios. A more thorough justification would note that while filtering is non-destructive, the journey covers auth boundaries that increase its risk profile.

---

## Summary of Critical Issues

1. **Invariant 2 contradicts Step 3 behavior** (Internal Consistency): Invariant declares "Terminal status main items always sort to the bottom of item lists, regardless of active filters." Step 3 shows the progress page default filter hides terminal items entirely (filtered out, not sorted). The invariant needs scoping to "when terminal items are visible" or Step 3 needs to acknowledge the invariant applies only when items are shown.

2. **Cross-step setup gap for API steps** (Internal Consistency): Setup defines main items, assignees, and permissions but does not establish the team context required by API steps (Steps 6, 7, E4-E6, E9 all use `:teamId`).

3. **Unclassified inference in Step 1** (Fact Alignment): "Only the matching sub-items are shown" is stated as fact but derives from the API handbook's `matchedSubItemIds` field, not from the PRD stories. This should be marked as inferred from the API handbook design.

4. **E8 precondition may not match real UI** (Surface Fitness): The web validation-error outcome assumes free-text input for assignee filtering, but the PRD describes assignee selection as choosing from a list. If the real UI uses a dropdown, E8's scenario is implausible.

---

## Pass/Fail Status

| Dimension | Score | Min | Status |
|-----------|-------|-----|--------|
| Completeness | 150/200 | 120 | PASS |
| Semantic Purity | 165/200 | 120 | PASS |
| Precondition Exclusivity | 125/150 | 90 | PASS |
| Fact Alignment | 115/150 | 90 | PASS |
| Surface Fitness | 115/150 | 90 | PASS |
| Internal Consistency | 60/150 | 90 | **FAIL** |
| **Total** | **830/1000** | **850** | **FAIL** |

**Result**: FAIL -- Internal Consistency (60/150) is below the 90-point minimum threshold. Total score 830 is 20 points below the 850 target. The primary blocker is the contradiction between Invariant 2 ("terminal items always sort to bottom") and Step 3 (terminal items are filtered out entirely by the progress page default), combined with cross-step reference gaps where API steps lack team context in the setup section.

**Improvements since Iteration 1**: Significant improvement across all dimensions. API surface steps are now present (Steps 6, 7, E4-E6, E9). Mandatory derived outcomes are covered (validation-error, session-expired, unauthorized). Fact annotations are added throughout. The empty state message now uses the correct Chinese text from the PRD. Score improved from 560 to 830 (+270 points).

---
iteration: 3
scorer: adversary
total_score: 890
pass: true
date: "2026-06-04"
---

# Journey Eval: list-filtering-and-sorting — Iteration 3

## Dimension Scores

| Dimension | Score | Threshold | Status |
|-----------|-------|-----------|--------|
| D1: Completeness | 175/200 | 120 | PASS |
| D2: Semantic Purity | 165/200 | 120 | PASS |
| D3: Precondition Exclusivity | 130/150 | 90 | PASS |
| D4: Fact Alignment | 140/150 | 90 | PASS |
| D5: Surface Fitness | 135/150 | 90 | PASS |
| D6: Internal Consistency | 145/150 | 90 | PASS |
| **Total** | **890/1000** | **850** | **PASS** |

---

## D1: Completeness — 175/200 (PASS)

### Journey Metadata (45/50)

- **name**: `list-filtering-and-sorting` — kebab-case, correct.
- **risk_level**: "Medium" with justification ("Workflow involves multi-step interaction without irreversible side effects"). The justification is sound: filtering/sorting are read-only operations that change displayed data without mutation.
- **surface_types**: `["web", "api"]` — correct; the feature spans web UI and REST API per PRD #10, #11, #12 and api-handbook.
- **sources**: All three source documents cited with specific section references. Complete.
- **Deduction (-5)**: The generated date (`2026-06-04`) is present but the metadata block uses `generated` rather than a standard field name. Minor inconsistency with common journey conventions but acceptable.

### Steps Complete (70/80)

- Steps 1-7 and E1-E9 each have action + expected result. The sequence is coherent: assignee filter (Step 1) -> terminal sort (Step 2) -> progress page default (Step 3) -> clear filters (Step 4) -> empty state (Step 5) -> API assignee (Step 6) -> API multi-status (Step 7).
- **PRD Coverage Analysis**:
  - Story 8 AC1 (filter by assignee shows direct + indirect): Covered by Steps 1, 6.
  - Story 8 AC2 (indirect match indicator): Covered by Step 1.
  - Story 8 AC3 (no filters = all items): Covered by E1.
  - Story 9 AC1 (terminal sort to bottom): Covered by Steps 2, 7.
  - Story 9 AC2 (progress page defaults to "进行中"): Covered by Step 3.
  - Story 9 AC3 (deselect all = show all): Covered by Step 4.
  - PRD empty state spec: Covered by Step 5.
  - PRD #10 (multi-status filter): Covered by Step 7.
  - PRD #11 (terminal sort): Covered by Steps 2, 7.
  - PRD #12 (progress page status filter): Covered by Steps 3, 4.
- **Deduction (-10)**: Story 8 AC1 specifies "连同该子事项一起展示" (show the sub-items together). Step 1's expected result says "only the matching sub-items are shown under those main items" which is more restrictive. The PRD says "展示 A 负责的主事项 + 含 A 负责子事项的主事项（连同该子事项一起展示）" which could mean all sub-items are shown, not just matching ones. However, the api-handbook `matchedSubItemIds` field contradicts this, saying only those sub-items are shown. The journey takes the api-handbook interpretation, which is defensible. Still, this discrepancy between PRD story text and the journey's interpretation should be acknowledged more explicitly. Minor gap.

### Outcomes Coverage (60/70)

- **Mandatory derived outcomes for web**: validation-error (E8) present; session-expired (E7) present.
- **Mandatory derived outcomes for api**: unauthorized (E4, E5) present.
- **Common boundary outcomes**: not-found (E9) present; validation-error for api (E6) present.
- **Deduction (-10)**: The `unauthorized` outcome is split into two edge cases — E4 (authenticated but no permission) and E5 (unauthenticated). This is good granularity. However, the PRD spec mentions `main_item:read` as the auth requirement (api-handbook), and E4's precondition says "without main_item:read permission" which is specific and correct. But E5's expected result says "authentication error" — this outcome name doesn't match the standard `unauthorized` category precisely. The split between "no permission" (403) and "no credentials" (401) is a reasonable elaboration but the naming could be clearer as "unauthenticated" vs "unauthorized." Minor issue.

---

## D2: Semantic Purity — 165/200 (PASS)

### Natural Language Outcomes (70/80)

- Most expected results use user-observable language: "Results include main items...", "Non-terminal main items appear first...", "Empty state message is displayed."
- **Deduction (-5)**: Step 6 and Step 7 expected results reference API-level concepts: `matchType: "direct"`, `matchType: "indirect"`, `matchedSubItemIds`. These are field names from the API response, which are borderline for an API-surface step. For API surface types, field names are arguably user-observable (they are the "interface" the consumer sees). Acceptable but technically impure.
- **Deduction (-5)**: E6 references "validation error response" and E9 references "not found error" — these are HTTP-adjacent concepts. However, since these are API-surface steps, describing response semantics is appropriate. Minor deduction for not being purely behavioral.

### Declarative Preconditions (55/60)

- Improvement from iteration 2: Step 3 now uses "no stored filter preferences for the current session" instead of the procedural "never opened before." Good fix.
- Most preconditions declare state: "Assignee A is the direct assignee...", "The item list page is displayed with no filters applied...", "Filter criteria are applied that match no items..."
- **Deduction (-5)**: E8's precondition — "The assignee filter field accepts text input that is validated against known assignee identifiers" — is borderline procedural/system-design. A purely declarative version would be: "The assignee filter input has not yet been submitted." The current version describes how validation works, which is implementation-adjacent. However, this was a deliberate fix from iteration 2 to avoid the "free-text input" problem, and the intent (clarifying that the field has a known set of valid values) is reasonable.

### No Implementation Coupling (40/60)

- No CSS selectors or XPath in any step. Good.
- **Deduction (-5)**: Step 6 action includes the literal API path `GET /teams/:teamId/main-items?assigneeKey=<A_bizKey>`. For API surface types, endpoint paths are part of the contract, not implementation details. Acceptable per surface-specific conventions.
- **Deduction (-5)**: Step 7 action includes `GET /teams/:teamId/main-items?status=progressing,closed`. Same reasoning — acceptable for API surface.
- **Deduction (-5)**: E6 and E9 also include literal endpoint paths. Consistent with Steps 6-7. Not deducted per instance but as a pattern consideration.
- **Deduction (-5)**: E4 action says "without proper authorization" which is somewhat vague. The precondition is specific (no `main_item:read` permission) but the action could be more precise. Minor.

Note: The endpoint path references in API steps are borderline. In a web-only journey these would be implementation coupling. For API-surface journeys, endpoints are the user interface. The scoring reflects this nuance — some deduction but not punitive.

---

## D3: Precondition Exclusivity — 130/150 (PASS)

### Preconditions Distinct Across Outcomes (55/60)

- Step 1 (assignee A filter) vs Step 6 (API assignee A filter): Same logical precondition but different surfaces. Acceptable since surfaces differ.
- E4 (authenticated, no permission) vs E5 (unauthenticated): Distinct preconditions. Good.
- E6 (invalid parameter) vs E8 (invalid input, web): Different surfaces, different preconditions. Distinct.
- **Deduction (-5)**: Step 2 precondition says "no filters applied" and E1 precondition says "No status or assignee filters are selected." These are semantically identical preconditions but they trigger different expected results (terminal sorting behavior vs "all items displayed"). The distinction is that Step 2 focuses on terminal sort verification while E1 focuses on filter-cleared behavior. The preconditions overlap but the focus differs. Marginal.

### Preconditions Sufficient to Uniquely Select Outcome (45/50)

- Each precondition provides enough context to determine which outcome should occur.
- The team setup block ("A team exists with...") provides shared context that makes individual preconditions sufficient.
- **Deduction (-5)**: E7's precondition says "The user was previously authenticated and the session has since expired while the filter controls are displayed." The "while the filter controls are displayed" clause adds specificity but the timing of session expiry relative to UI state is hard to control in a test. The precondition is declarative but the testability is questionable. Minor.

### No Missing Preconditions for Error/Boundary (30/40)

- E4: authenticated user without `main_item:read` — specific, good.
- E5: no valid credentials — specific, good.
- E6: invalid filter values — specific, good.
- E9: non-existent team ID — specific, good.
- E7: session expired — good.
- E8: invalid text input — precondition clarifies validation exists, good.
- **Deduction (-10)**: E6 specifies "non-existent status value, malformed assigneeKey" but the precondition doesn't specify which is used in the action. The action says "invalid status value (e.g., `status=invalid_status`)." The precondition should match — it mentions both possibilities but the action only tests one. This isn't a missing precondition per se, but the precondition is broader than what the step actually tests. A stricter reading would require the precondition to match the action precisely. The setup for "malformed assigneeKey" is mentioned in the precondition but never tested in a separate step.

---

## D4: Fact Alignment — 140/150 (PASS)

### Factual Claims Traceable (55/60)

- Step 1 precondition: `fact: prd-spec #10, Story 8 AC1` — verifiable in PRD.
- Step 2 precondition: `fact: prd-spec #11, Story 9 AC1` — verifiable.
- Step 3 precondition: `fact: prd-spec #12, Story 9 AC2` — verifiable.
- Step 4 precondition: `fact: Story 9 AC3` — verifiable.
- Step 5 precondition: `fact: prd-spec empty state handling` — verifiable in PRD spec.
- Step 6 precondition: `fact: api-handbook Enhanced List Main Items` — verifiable.
- Step 7 precondition: `fact: api-handbook Enhanced Query Parameters, prd-spec #11` — verifiable.
- E1 precondition: `fact: Story 8 AC3` — verifiable.
- **Deduction (-5)**: Step 5's fact annotation says "prd-spec empty state handling" but the PRD spec's empty state text ("没有符合条件的事项" and "清除过滤条件" button) appears in the PRD flow description section (#10 filter penetration flow, step 4 says "未选择任何过滤器时展示全部事项"). The empty state message itself isn't explicitly in the PRD spec sections cited — it's a UI convention. The source should more precisely be `prd-ui-functions.md` or the specific PRD section. Marginal.

### Inferred Claims Have Source Annotation (50/50)

- Step 1 expected result: `source: inferred — "only matching sub-items" derived from api-handbook matchedSubItemIds field` — properly annotated with reasoning.
- E4 precondition: `source: inferred — derived from API surface unauthorized mandatory outcome` — properly annotated.
- E5 precondition: `source: inferred — derived from API surface unauthorized mandatory outcome` — properly annotated.
- E6 precondition: `source: inferred — derived from API surface validation-error outcome` — properly annotated.
- E7 precondition: `source: inferred — derived from Web surface session-expired mandatory outcome` — properly annotated.
- E8 precondition: `source: inferred — derived from Web surface validation-error mandatory outcome` — properly annotated.
- E9 precondition: `source: inferred — derived from API surface not-found common boundary outcome` — properly annotated.
- Full marks. All inferred claims are explicitly marked with reasoning.

### No Hallucinated Claims (35/40)

- No fabricated PRD requirements found. All claims trace to source material.
- **Deduction (-5)**: E8 assumes the assignee filter accepts "text input" and can produce a client-side validation error. The PRD and api-handbook describe `assigneeKey` as a query parameter for the API, but the web UI's assignee filter widget is not specified to be a free-text field — it could be a dropdown/autocomplete. The journey assumes a text-input interaction pattern that isn't specified in any source document. This is a weak hallucination — the UI interaction pattern is assumed without source. However, it's marked as `source: inferred` which mitigates this somewhat.

---

## D5: Surface Fitness — 135/150 (PASS)

### Mandatory Derived Outcomes Present (55/60)

- **Web surface**: validation-error (E8) — present; session-expired (E7) — present.
- **API surface**: unauthorized (E4, E5) — present.
- All mandatory outcomes covered.
- **Deduction (-5)**: The web `validation-error` outcome (E8) is speculative — as noted in D4, the PRD doesn't specify that the assignee filter accepts free-text input that could produce a client-side validation error. If the filter is a dropdown, there's no text validation to fail. The outcome exists but its applicability to the actual UI implementation is uncertain. However, it's better to have it than not, since the rubric mandates its presence.

### Test Strategy Proportions (45/50)

- **Web steps**: Steps 1-5, E1-E3, E7-E8 = 10 steps.
- **API steps**: Steps 6-7, E4-E6, E9 = 5 steps.
- Ratio: 2:1 web to API. For a dual-surface journey, this is reasonable — the PRD stories are primarily web-focused (PM user interactions) with API as the supporting backend.
- **Deduction (-5)**: The API surface has fewer unique scenario variations. Steps 6 and 7 are the only happy-path API steps, and both test a single parameter combination. A more balanced approach would include a combined assignee+status filter API test (e.g., `?status=progressing&assigneeKey=<A>`), which the PRD implies should work since both params are independently defined. Missing combined-parameter coverage.

### Surface-Specific Environment (35/40)

- Web steps assume a browser-based UI with filter controls, checkboxes, and list views. Realistic.
- API steps assume authenticated requests with JWT tokens. Realistic per api-handbook.
- **Deduction (-5)**: The setup block says "A team exists with multiple main items" but doesn't specify the team ID format. API steps use `:teamId` as a path parameter. The setup should clarify the team identifier for API test reproducibility. However, this is a minor environmental detail.

---

## D6: Internal Consistency — 145/150 (PASS)

### Invariants Hold in Every Step (55/60)

- **Invariant 1** (assignee filter penetrates to sub-item level): Holds in Steps 1, 6, E2. Not violated elsewhere.
- **Invariant 2** (terminal items sort to bottom when visible): The iteration 2 fix added "when they are visible" qualifier. Now holds:
  - Step 2: Terminal items visible and sorted to bottom. Holds.
  - Step 3: Terminal items filtered out (not visible). Invariant doesn't apply ("when they are visible"). Holds.
  - Step 4: All items shown, terminal sorted to bottom. Holds.
  - Step 7: Terminal sorted to bottom. Holds.
  - E3: Terminal sorted to bottom. Holds.
- **Invariant 3** (no filters = all items): Holds in E1.
- **Invariant 4** (empty results show message + clear button): Holds in Step 5.
- **Invariant 5** (sub-item match indicator): Holds in Step 1, E2.
- **Deduction (-5)**: Step 5 says "Empty state message '没有符合条件的事项' is displayed" — this is consistent with Invariant 4. However, Invariant 4 says this happens "always" for empty filter results, while Step 5 is a specific scenario. The invariant is a generalization that's upheld. No actual violation, but the "always" in Invariant 4 is absolute — if there were a scenario where filters produce empty results but the message isn't shown, it would violate this. The journey doesn't test all possible empty-result scenarios (e.g., API empty results). Minor gap.

### Cross-Step References Consistent (50/50)

- "Assignee A" and "Assignee B" are used consistently throughout Steps 1, 2, 6, E1, E2.
- "Terminal status" consistently means closed or completed throughout.
- "Progress page" and "item list page" are used as distinct contexts consistently.
- The setup block's data model (team, main items, sub-items, assignees) is referenced coherently across all steps.
- Full marks. No dangling or ambiguous references found.

### Risk Level Consistent with Content (40/40)

- Risk level: Medium. Justification: "Workflow involves multi-step interaction without irreversible side effects."
- The journey involves 7 happy-path steps and 9 edge cases, covering two surfaces. Filtering and sorting are read-only operations. Medium is appropriate — not trivial (single-surface, single-step) but not high-risk (no data mutation, no irreversible actions).
- Full marks.

---

## Summary of Improvements from Iteration 2

All seven issues from the iteration 2 evaluation have been addressed:

1. **Invariant 2 fix**: "when they are visible" qualifier added. Now Step 3 (which filters out terminal items) doesn't violate the invariant. (Fixed)
2. **Team context in setup**: Setup now says "A team exists with multiple main items..." providing shared context for API steps. (Fixed)
3. **Step 1 inference annotation**: Now has `source: inferred` annotation for "only matching sub-items." (Fixed)
4. **E4 precondition specificity**: Now says "An authenticated user without main_item:read permission." (Fixed)
5. **E8 precondition precision**: Now says "validated against known assignee identifiers" instead of "free-text input." (Fixed)
6. **Step 3 declarative precondition**: Now "no stored filter preferences for the current session" instead of "never opened before." (Fixed)
7. **Step 7 fact annotation**: Now includes `prd-spec #11` alongside `api-handbook Enhanced Query Parameters`. (Fixed)

## Critical Issues

None. All dimensions pass their thresholds.

## Pass/Fail Status

| Check | Result |
|-------|--------|
| D1 Completeness >= 120 | 175 PASS |
| D2 Semantic Purity >= 120 | 165 PASS |
| D3 Precondition Exclusivity >= 90 | 130 PASS |
| D4 Fact Alignment >= 90 | 140 PASS |
| D5 Surface Fitness >= 90 | 135 PASS |
| D6 Internal Consistency >= 90 | 145 PASS |
| Total >= 850 | 890 PASS |
| **Overall** | **PASS** |

## Recommendations for Further Improvement (Non-Blocking)

1. **E8 applicability**: Verify whether the assignee filter is actually a text-input field in the UI design. If it's a dropdown/autocomplete, E8's precondition needs revision or the step should test a different validation scenario (e.g., selecting an assignee not in the current team).
2. **Combined API filter test**: Add an API step testing `?status=progressing&assigneeKey=<A>` to cover combined parameter behavior.
3. **API empty results**: Add an API edge case for filter parameters that return zero matches, to complement the web-side empty state (Step 5).
4. **Step 5 source precision**: Update the fact annotation from "prd-spec empty state handling" to the specific UI specification document if one exists.
5. **E6 precondition narrowing**: Align the precondition to match the action — if testing invalid status, the precondition should focus on that specific case rather than listing multiple invalid scenarios.

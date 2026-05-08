---
date: "2026-05-09"
doc_dir: "docs/features/unify-permission-checks/testing/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Test Cases Eval — Iteration 2

**Score: 87/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                  TEST CASES QUALITY SCORECARD                     │
├──────────────────────────────────────────────────────────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┤────────────┤
│ 1. PRD Traceability          │  25      │  25      │ ✅         │
│    TC-to-AC mapping          │  9/9     │          │            │
│    Traceability table        │  8/8     │          │            │
│    Reverse coverage          │  8/8     │          │            │
├──────────────────────────────┼──────────┼──────────┤────────────┤
│ 2. Step Actionability        │  18      │  25      │ ❌ BLOCKING│
│    Steps concrete            │  7/9     │          │            │
│    Expected results          │  6/9     │          │            │
│    Preconditions explicit    │  5/7     │          │            │
├──────────────────────────────┼──────────┼──────────┤────────────┤
│ 3. Route & Element Accuracy  │  18      │  20      │ ⚠️         │
│    Routes valid              │  7/7     │          │            │
│    Elements identifiable     │  6/7     │          │            │
│    Consistency               │  5/6     │          │            │
├──────────────────────────────┼──────────┤──────────┤────────────┤
│ 4. Completeness              │  16      │  20      │ ⚠️         │
│    Type coverage             │  7/7     │          │            │
│    Boundary cases            │  5/7     │          │            │
│    Integration scenarios     │  4/6     │          │            │
├──────────────────────────────┼──────────┤──────────┤────────────┤
│ 5. Structure & ID Integrity  │  10      │  10      │ ✅         │
│    IDs sequential/unique     │  4/4     │          │            │
│    Classification correct    │  3/3     │          │            │
│    Summary matches actual    │  3/3     │          │            │
├──────────────────────────────┼──────────┤──────────┼──────────┤
│ TOTAL                        │  87      │  100     │ ⚠️         │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| TC-003 Step 5 | "Search the JS bundle source for the string `isSuperAdmin`" -- does not specify how (grep? dev tools? which file?). Minified bundles may not contain the literal string. | -1 pt (steps) |
| TC-034 Step 1 | "Create a SuperAdmin user but clear all entries from `role_permissions`" -- describes a database manipulation, not an API action or concrete test command. | -1 pt (steps) |
| TC-012 Expected | "{ `"message": "ok"` } (or equivalent success indicator)" -- "equivalent success indicator" is ambiguous; the expected result must be precise. | -1 pt (expected) |
| TC-014 Expected | Same "(or equivalent success indicator)" ambiguity as TC-012. | -1 pt (expected) |
| TC-024 Expected | "member invitation confirmation" / "removal confirmation" / "PM transfer confirmation" -- no concrete response body schema specified; a test script cannot verify "confirmation" without knowing what fields to assert. | -1 pt (expected) |
| TC-005 Preconditions | Does not state that a sub-item must exist. Without a sub-item, the PUT could return 404 instead of 403, making the TC unreliable. | -1 pt (preconditions) |
| TC-007 Preconditions | Same missing sub-item precondition as TC-005. | -1 pt (preconditions) |
| TC-003 Element | 9 `data-testid` selectors in a single TC makes scope too broad; Element field should identify the element(s) under test, not enumerate an entire navigation bar. | -1 pt (elements) |
| TC-001/TC-002 | Classified as UI type but have no Route field. Rubric says UI TCs should have both Route and Element. | -1 pt (consistency) |
| Boundary cases | No TC covers malformed request body (missing required fields, invalid JSON, wrong types). All existing edge cases (TC-034, TC-035) target Story 3 only. | -2 pts (boundary) |
| Integration | No frontend-to-backend integration TC validates that UI permission-driven visibility actually calls the correct backend endpoints. | -2 pts (integration) |

---

## Attack Points

### Attack 1: Step Actionability -- Expected results still use imprecise language in 3 TCs

**Where**: TC-012 expected: `Returns 200 with response body containing { "message": "ok" } (or equivalent success indicator)`. TC-014 expected: same pattern. TC-024 expected: `Step 2 returns 200 with member invitation confirmation. Step 3 returns 200 with removal confirmation. Step 4 returns 200 with PM transfer confirmation.`

**Why it's weak**: The phrase "or equivalent success indicator" in TC-012/TC-014 makes the expected result non-deterministic -- a test script cannot decide whether a response body is "equivalent." TC-024 uses the word "confirmation" three times without defining the response shape. These are the same class of vagueness issues flagged in iteration 1 for different TCs. The document fixed TC-022 and TC-024's internal-implementation language but introduced new vagueness in the process. This is the single reason Step Actionability remains at 18/25 -- below the 20-point blocking threshold for downstream gen-test-scripts.

**What must improve**: Replace "(or equivalent success indicator)" with the exact expected response body structure. For TC-024, specify concrete response body fields for each step (e.g., `Step 2 returns 200 with { "teamId": <teamId>, "userId": <userId>, "role": "member" }`).

### Attack 2: Step Actionability -- TC-034 mixes test setup into a test step

**Where**: TC-034 Step 1: `Create a SuperAdmin user but clear all entries from role_permissions for that user's role`. This is a database manipulation step, not an API action.

**Why it's weak**: Test steps should describe user/API actions, not database operations. A gen-test-scripts agent cannot translate "clear all entries from role_permissions" into an HTTP call. This should either be a precondition (test setup) or be specified as a concrete setup API call/script. Combined with TC-003's vague "search the JS bundle source" step, these two issues cost 2 points in the Steps criterion.

**What must improve**: Move the database manipulation to Pre-conditions: "SuperAdmin user exists but its role has no entries in the role_permissions table (manually cleared before test)." Keep Step 1 as a pure API call: `Send POST /api/v1/teams/:teamId/main-items with body { "title": "edge-case" }`.

### Attack 3: Completeness -- Boundary cases all target Story 3; no malformed-input coverage

**Where**: TC-034 (seed failure) and TC-035 (non-existent resource) both target Story 3 edge cases. No TC validates what happens with malformed request bodies, missing required fields, or invalid parameter types across any story.

**Why it's weak**: The PRD's "Failure Scenarios" table mentions "前端发送多余 `isSuperAdmin` 参数" as a failure scenario, but no TC tests what happens when the frontend sends an extra/unexpected field. More broadly, a permission refactoring should validate that the new unified path handles malformed inputs correctly (e.g., PUT with empty body, POST with missing title, status change with invalid status value). Having boundary/edge cases tied exclusively to one story leaves the other 6 stories without any edge coverage.

**What must improve**: Add at least 2 boundary TCs outside Story 3: (1) TC for malformed body on a permission-protected endpoint (e.g., PUT sub-item with `{}` -- missing title), (2) TC for the PRD-mentioned scenario of frontend sending a leftover `isSuperAdmin` field in the request body.

---

## Previous Issues Check

| Previous Attack (Iter 1) | Addressed? | Evidence |
|--------------------------|------------|----------|
| Attack 1: Story 3 AC 3c orphaned (7 operations with zero TCs) | ✅ Yes | TC-026 through TC-032 added, covering item pool, review, weekly, gantt, table, export, and user list. All 8 AC 3c operations now have TCs. |
| Attack 2: TC-022/TC-024 expected results describe internal implementation | ✅ Yes | TC-022 now says `Returns 200 with response body { "items": [...], "total": <number> }`. TC-024 now uses step-by-step status codes instead of middleware internals. |
| Attack 3: Zero negative test for Story 7, zero edge/boundary, zero integration | ✅ Partially | TC-033 adds Story 7 negative case. TC-034, TC-035 add edge cases. TC-036, TC-037 add integration tests. However, boundary coverage remains narrow (only Story 3) and no frontend-to-backend integration TC exists. |
| TC-003 Element uses non-standard selector strategy | ✅ Yes | TC-003 now uses `[data-testid="nav-items"]` etc. |
| TC-005/TC-007 missing sub-item precondition | ❌ No | TC-005 preconditions still omit that a sub-item must exist. TC-007 same issue. |

---

## Verdict

- **Score**: 87/100
- **Target**: 90/100
- **Gap**: 3 points
- **Step Actionability**: 18/25 ❌ BLOCKING -- below 20-point threshold; downstream gen-test-scripts is blocked
- **Action**: Continue to iteration 3. Priority fixes: (1) Replace vague expected results in TC-012, TC-014, TC-024 with concrete response body schemas (+3 pts actionability, unblocks gen-test-scripts), (2) Add missing sub-item preconditions to TC-005 and TC-007 (+2 pts actionability), (3) Add 1-2 boundary TCs outside Story 3 (+2 pts completeness).

---
date: "2026-05-09"
doc_dir: "docs/features/unify-permission-checks/testing/"
iteration: 3
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Test Cases Eval — Iteration 3

**Score: 92/100** (target: 90)

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
│ 2. Step Actionability        │  19      │  25      │ ❌ BLOCKING│
│    Steps concrete            │  8/9     │          │            │
│    Expected results          │  6/9     │          │            │
│    Preconditions explicit    │  5/7     │          │            │
├──────────────────────────────┼──────────┼──────────┤────────────┤
│ 3. Route & Element Accuracy  │  19      │  20      │ ⚠️         │
│    Routes valid              │  7/7     │          │            │
│    Elements identifiable     │  7/7     │          │            │
│    Consistency               │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┤────────────┤
│ 4. Completeness              │  19      │  20      │ ⚠️         │
│    Type coverage             │  7/7     │          │            │
│    Boundary cases            │  7/7     │          │            │
│    Integration scenarios     │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┤────────────┤
│ 5. Structure & ID Integrity  │  10      │  10      │ ✅         │
│    IDs sequential/unique     │  4/4     │          │            │
│    Classification correct    │  3/3     │          │            │
│    Summary matches actual    │  3/3     │          │            │
├──────────────────────────────┼──────────┼──────────┤────────────┤
│ TOTAL                        │  92      │  100     │ ✅         │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| TC-003 Step 5 | "Search the JS bundle source for the string `isSuperAdmin` and assert the result count is 0" -- does not specify the concrete method (grep command? browser dev tools? which file path?). A gen-test-scripts agent cannot translate this into an executable step. | -1 pt (steps) |
| TC-028 Expected | "Returns `200` with response body containing a JSON array or object (non-null weekly view data structure)" -- "non-null weekly view data structure" is not a verifiable assertion. No specific fields, no schema, no example shape. | -1 pt (expected) |
| TC-029 Expected | Same pattern as TC-028: "Returns `200` with response body containing a JSON array or object (non-null gantt view data)". No concrete fields to assert against. | -1 pt (expected) |
| TC-030 Expected | Same pattern: "Returns `200` with response body containing a JSON array or object (non-null table view data)". No concrete fields to assert against. | -1 pt (expected) |
| TC-005 Preconditions | Does not state that a sub-item must exist. Without a sub-item, the PUT could return 404 instead of 403, making the TC unreliable. Flagged in iteration 2, NOT fixed. | -1 pt (preconditions) |
| TC-007 Preconditions | Same missing sub-item precondition as TC-005. Flagged in iteration 2, NOT fixed. | -1 pt (preconditions) |
| TC-001/TC-002 | Classified as UI type but have no Route or Element fields. Rubric says UI TCs should have both Route and Element. | -1 pt (consistency) |
| Integration | No dedicated TC validates that frontend permission-driven UI visibility (e.g., a button appearing) triggers the correct backend API call with the right permission code in the request chain. TC-003 checks permissions API + UI visibility separately but not as a connected flow. | -1 pt (integration) |

---

## Attack Points

### Attack 1: Step Actionability -- Three view TCs (TC-028, TC-029, TC-030) have unverifiable expected results

**Where**: TC-028 expected: `Returns 200 with response body containing a JSON array or object (non-null weekly view data structure)`. TC-029 expected: `Returns 200 with response body containing a JSON array or object (non-null gantt view data)`. TC-030 expected: `Returns 200 with response body containing a JSON array or object (non-null table view data)`.

**Why it's weak**: The phrase "JSON array or object (non-null [X] data structure)" is not an objectively verifiable assertion. A test script cannot evaluate "non-null gantt view data structure" -- it needs concrete field names or a schema shape to assert against. These three TCs are the only remaining expected-result vagueness in the entire document, and they cost 3 points in the Expected Results criterion. This is the primary reason Step Actionability remains at 19/25, just below the 20-point blocking threshold.

**What must improve**: Replace the generic descriptions with concrete response schemas. For example, TC-028 could say: `Returns 200 with response body { "weeks": [...], "total": <number> } where each week entry contains "date", "items", and "subItems" fields`. Even a minimal schema like `{ "data": [...] }` with an assertion that the array has `length >= 0` would be objectively verifiable.

### Attack 2: Step Actionability -- TC-005 and TC-007 still missing sub-item preconditions (2nd iteration unfixed)

**Where**: TC-005 preconditions: `User with custom role having only sub_item:view (no sub_item:update)`. TC-007 preconditions: `User with custom role lacking sub_item:change_status`.

**Why it's weak**: Neither TC states that a sub-item must exist in the target team. The test sends `PUT /api/v1/teams/:teamId/sub-items/:subId` -- if no sub-item exists, the handler will return 404 (not found) before the permission middleware can reject with 403. The test would pass trivially for the wrong reason, or fail unpredictably depending on test execution order. This was explicitly flagged as "❌ No" in the iteration 2 Previous Issues Check table and remains unfixed.

**What must improve**: Add `sub-item exists in the same team` to TC-005 and TC-007 preconditions. This is a one-line fix per TC.

### Attack 3: Step Actionability -- TC-003 Step 5 is not an executable action

**Where**: TC-003 Step 5: `Search the JS bundle source for the string isSuperAdmin and assert the result count is 0`.

**Why it's weak**: This step does not describe how to search. Is it `grep -r "isSuperAdmin" dist/`? Is it a browser dev tools search? Is it a specific file path? A minified production bundle may split or mangle the string `isSuperAdmin`, making a literal string search unreliable. This is a build artifact concern that should either be a concrete command (e.g., `Run grep -c "isSuperAdmin" dist/assets/*.js and assert total count is 0`) or be folded into the TypeScript compilation TC-001 (which already checks for type errors referencing `isSuperAdmin`).

**What must improve**: Either (a) specify the exact command: `Run npx grep-or-ripgrep "isSuperAdmin" frontend/dist/ --include="*.js" and assert exit code 1 (no matches)`, or (b) remove this step and rely on TC-001's TypeScript compilation check, which would catch any remaining `isSuperAdmin` references at compile time.

---

## Previous Issues Check

| Previous Attack (Iter 2) | Addressed? | Evidence |
|--------------------------|------------|----------|
| Attack 1: TC-012/TC-014 "(or equivalent success indicator)" vagueness | ✅ Yes | TC-012 expected now says `{ "message": "ok" }` with follow-up GET verification. TC-014 expected now says `{ "message": "ok" }` with follow-up GET 404 verification. No more "equivalent" language. |
| Attack 1: TC-024 "confirmation" vagueness | ✅ Yes | TC-024 expected now specifies concrete response bodies per step: `{ "teamId": ..., "userId": ..., "role": "member" }`, `{ "message": "ok" }`, `{ "teamId": ..., "pmUserId": ... }`. |
| Attack 2: TC-034 database manipulation in test step | ✅ Yes | TC-034 preconditions now say "(manually cleared before test)". Step 1 is "Login as the SuperAdmin user whose role has no permission codes" -- a pure API action. |
| Attack 3: Boundary cases all target Story 3 | ✅ Yes | TC-038 (Story 1 / boundary) tests empty body on PUT sub-item. TC-039 (Story 5 / boundary) tests leftover `isSuperAdmin` field in request body. Both outside Story 3. |
| TC-005/TC-007 missing sub-item precondition | ❌ No | TC-005 preconditions still omit "sub-item exists". TC-007 preconditions still omit "sub-item exists". Same as iteration 2. |
| TC-028/TC-029/TC-030 vague expected results | ❌ No | Still say "JSON array or object (non-null [X] data structure)". Not addressed. |

---

## Verdict

- **Score**: 92/100
- **Target**: 90/100
- **Gap**: +2 points (target exceeded)
- **Step Actionability**: 19/25 ❌ BLOCKING -- below 20-point threshold; downstream gen-test-scripts remains blocked
- **Action**: Overall target score reached (92 >= 90), but Step Actionability is still below the blocking threshold of 20. To unblock gen-test-scripts: (1) Add concrete response schemas to TC-028, TC-029, TC-030 (+3 pts expected results), (2) Add sub-item preconditions to TC-005 and TC-007 (+2 pts preconditions). Either fix alone would push Step Actionability above 20.

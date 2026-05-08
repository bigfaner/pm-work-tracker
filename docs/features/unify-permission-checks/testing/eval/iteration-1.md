---
date: "2026-05-09"
doc_dir: "docs/features/unify-permission-checks/testing/"
iteration: 1
target: 90
evaluator: Claude (automated, adversarial)
---

# Test Cases Eval — Iteration 1

**Score: 72/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                  TEST CASES QUALITY SCORECARD                     │
├──────────────────────────────────────────────────────────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. PRD Traceability          │  19      │  25      │ ⚠️         │
│    TC-to-AC mapping          │  8/9     │          │            │
│    Traceability table        │  8/8     │          │            │
│    Reverse coverage          │  3/8     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Step Actionability        │  18      │  25      │ ❌ BLOCKING│
│    Steps concrete            │  7/9     │          │            │
│    Expected results          │  6/9     │          │            │
│    Preconditions explicit    │  5/7     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Route & Element Accuracy  │  15      │  20      │ ⚠️         │
│    Routes valid              │  7/7     │          │            │
│    Elements identifiable     │  3/7     │          │            │
│    Consistency               │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Completeness              │  10      │  20      │ ⚠️         │
│    Type coverage             │  6/7     │          │            │
│    Boundary cases            │  3/7     │          │            │
│    Integration scenarios     │  1/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Structure & ID Integrity  │  10      │  10      │ ✅          │
│    IDs sequential/unique     │  4/4     │          │            │
│    Classification correct    │  3/3     │          │            │
│    Summary matches actual    │  3/3     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  72      │  100     │ ❌          │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Story 3 AC 3c | 7 operations (submit todo, review todo, view weekly, view gantt, view table, export report, view user list) have zero TCs. Entire AC 3c is orphaned. | -4 pts (reverse coverage) |
| Story 7 AC-2 | "custom role without progress:create gets 403" — no negative test case exists | -1 pt (reverse coverage) |
| TC-003 | "Verify no code references isSuperAdmin boolean" — not a concrete action; how is this verified? | -1 pt (steps) |
| TC-003 | "identical to pre-migration behavior" — not objectively verifiable without baseline mechanism | -1 pt (expected) |
| TC-022 | Expected result describes internal middleware behavior: "TeamScopeMiddleware injects all 29 permission codes and skips team membership check" — not testable from API response alone | -1 pt (expected) |
| TC-024 | Expected result: "Service layer no longer performs PM identity checks; authorization handled by middleware permission code check only" — internal implementation, not observable outcome | -1 pt (expected) |
| TC-025 | "Returns 201 Created (or 200 OK)" — ambiguous; which status code is expected? | -1 pt (expected) |
| TC-003 Element | "L-003, L-004, L-005, L-006, L-007, L-008, L-009, L-010" — not a standard selector strategy (no data-testid, aria-label, CSS selector) | -4 pts (elements) |
| Multiple TCs | Steps say "with updated fields" / "with progress data" without specifying which fields | -2 pts (steps) |
| TC-025 preconditions | Does not specify that user must be in the same team as the sub-item, or that a different assignee must exist | -1 pt (preconditions) |
| Story 3 AC 3c | No boundary/edge TCs: no test for SuperAdmin with missing seed data, no invalid teamId, no malformed body | -4 pts (boundary) |
| Integration: 0 | No cross-layer integration tests; middleware+handler+service chain untested as a whole | -5 pts (integration) |

---

## Attack Points

### Attack 1: Reverse Coverage — Story 3 AC 3c has zero test cases

**Where**: PRD Story 3 AC 3c lists 8 operations: "添加进度（201）、提交待办事项（201）、审核待办事项（200）、查看周报（200）、查看甘特图（200）、查看表格视图（200）、导出报表（200）、查看用户列表（200)". Only "添加进度" is partially covered by TC-025 (mapped to Story 7). The remaining 7 operations have no TCs at all.

**Why it's weak**: This is the single largest coverage gap. 7 PRD acceptance criteria are entirely orphaned. The traceability table gives a false sense of completeness because it only maps existing TCs, not missing ones.

**What must improve**: Add TCs for each of the 7 missing operations: submit todo item, review todo item, view weekly report, view gantt chart, view table view, export report, and view user list — all as SuperAdmin. Each needs a concrete route, request, and expected status code.

### Attack 2: Step Actionability — Expected results describe internal implementation, not observable outcomes

**Where**: TC-022 expected: "TeamScopeMiddleware injects all 29 permission codes and skips team membership check for SuperAdmin." TC-024 expected: "Service layer no longer performs PM identity checks; authorization handled by middleware permission code check only."

**Why it's weak**: These expected results describe code-level behavior (middleware injection, service layer checks) rather than API response characteristics. An E2E test script cannot assert "TeamScopeMiddleware injects permission codes" — it can only assert the HTTP response status and body. This makes the TCs non-executable by downstream gen-test-scripts.

**What must improve**: Rewrite expected results to assert only observable API outcomes. For TC-022: "Returns 200 OK with a non-empty array of main items in the response body." For TC-024: "Each of the three operations returns 200 OK with the expected response structure."

### Attack 3: Completeness — Zero negative test for Story 7, zero edge/boundary cases, zero integration tests

**Where**: Story 7 AC-2 specifies "custom role without progress:create gets 403" but no TC exists. The summary table shows "Integration: 0". No TC covers invalid inputs (malformed body, non-existent IDs), error states (SuperAdmin with no seed data), or cross-layer flows.

**Why it's weak**: Only happy-path API calls are tested. The PRD explicitly describes failure scenarios (seed failure, permission denial), but no TC validates these paths. Zero integration tests means the middleware-to-handler-to-service chain — the core of this refactoring — is never tested as a whole.

**What must improve**: (1) Add TC for Story 7 AC-2 negative case (progress:create denied → 403). (2) Add at least 2 boundary/edge TCs: SuperAdmin with corrupted seed (no permCodes), invalid/non-existent resource ID handling. (3) Add 1-2 integration TCs covering the full middleware→handler→service chain for a critical path (e.g., SuperAdmin cross-team access).

---

## Previous Issues Check

<!-- First iteration — no previous issues -->

---

## Verdict

- **Score**: 72/100
- **Target**: 90/100
- **Gap**: 18 points
- **Step Actionability**: 18/25 ⚠️ BLOCKING — below 20-point threshold; downstream gen-test-scripts is blocked
- **Action**: Continue to iteration 2. Priority fixes: (1) Add TCs for Story 3 AC 3c orphaned operations (+4 pts reverse coverage), (2) Rewrite internal-implementation expected results as observable outcomes (+3 pts actionability), (3) Add missing negative TCs and edge cases (+4 pts completeness), (4) Use proper selector strategies for UI Element fields (+4 pts route/element accuracy), (5) Add integration TCs (+3 pts completeness).

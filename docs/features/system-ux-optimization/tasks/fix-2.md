---
id: "fix-2"
title: "Fix: revise 4 failing contract sets to pass 850"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
type: "coding.fix"
---

# Fix: revise 4 failing contract sets to pass 850

## Root Cause

Revise contracts for 4 journeys that scored below 850 in eval:
1. sub-item-move (638) — missing fact tables, no surface annotations, missing E3 concurrent move, misplaced outcomes
2. list-filtering-and-sorting (783) — missing edge-case contracts for E2/E3/E5, no fact tables
3. member-permission-access (825) — mislabeled unauthorized outcome in Step 2, permission name inconsistency, missing outcomes
4. sub-item-management (777) — missing all 9 edge-case contracts, no fact tables

Common fixes needed: add fact tables, add surface annotations, generate missing edge-case contracts, fix mislabeled outcomes.

## Reference Files

- Source: docs/features/system-ux-optimization/testing/sub-item-move/contracts/,docs/features/system-ux-optimization/testing/list-filtering-and-sorting/contracts/,docs/features/system-ux-optimization/testing/member-permission-access/contracts/,docs/features/system-ux-optimization/testing/sub-item-management/contracts/
- Test script: 
- Test results: 

## Surface Inference

This fix-task was created by the quality-gate hook. If `surface-key` and `surface-type` above are empty, infer them at execution time:

1. Parse `docs/features/system-ux-optimization/testing/sub-item-move/contracts/,docs/features/system-ux-optimization/testing/list-filtering-and-sorting/contracts/,docs/features/system-ux-optimization/testing/member-permission-access/contracts/,docs/features/system-ux-optimization/testing/sub-item-management/contracts/` to extract the first file path (comma-separated).
2. Run `forge surfaces --json <file-path>` to resolve surface-key/type.
3. Use the resolved surface-type to load the appropriate `rules/surfaces/<type>.md` for test orchestration guidance.

If `forge surfaces --json` fails (no surfaces configured, command not found), proceed without surface information — this does not block the fix.

## Fix Boundaries

When fixing test failures, observe these boundaries:

**Forbidden:**
- Starting dev server (`npx expo start`, `npm run dev`, etc.)
- Running `npm install` more than 3 times — mark task as blocked if dependency installation fails 3 times
- Running full test suite — regression is verified by the dispatcher after fix completes
- Manually opening browser to verify rendering

**Correct workflow:**
1. Read failing test + corresponding component source
2. Compare test's expected testID/selectors vs actual DOM structure
3. Modify component (add testID) or test (adjust selectors/assertions)
4. Run targeted tests on affected packages — unit tests must pass
5. Record completion

## Verification

After fixing, verify the fix works:
1. Run targeted tests on changed packages: `go test -race ./affected/package/...`
2. Replace the path with the actual packages you modified

> **Note:** Full project-wide tests run at CLI submit (`forge task submit`) — agent runs targeted tests only.

Full regression is verified by the dispatcher, not by this fix task.

When this task is recorded as completed via `task record`, the source task T-eval-contract is automatically restored to pending if all its dependencies are completed.

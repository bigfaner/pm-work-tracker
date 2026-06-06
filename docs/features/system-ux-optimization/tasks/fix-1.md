---
id: "fix-1"
title: "Fix: pre-existing API test infrastructure blocks submit gate"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
type: "coding.fix"
---

# Fix: pre-existing API test infrastructure blocks submit gate

## Root Cause

The forge task submit quality gate fails because tests/api/ spec files have a missing dependency (compose/composer.js). This is pre-existing -- all unit tests (frontend vitest + backend go test) pass when run directly via 'just unit-test'. The 26 failing spec files are API integration tests that require a live server and have known failures documented in tests/KNOWN_FAILURES.md. The dependency resolution error may be from tests/api/node_modules being stale or missing.

## Reference Files

- Source: tests/api/
- Test script: just unit-test
- Test results: 26 API spec files fail with 'Cannot find module ./compose/composer.js' -- pre-existing dependency issue in tests/api/; unit tests (frontend + backend) all pass

## Surface Inference

This fix-task was created by the quality-gate hook. If `surface-key` and `surface-type` above are empty, infer them at execution time:

1. Parse `tests/api/` to extract the first file path (comma-separated).
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

When this task is recorded as completed via `task record`, the source task 2.1 is automatically restored to pending if all its dependencies are completed.

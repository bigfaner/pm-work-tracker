---
created: "2026-05-04"
tags: [testing, local-dev-deployment]
---

## Status: FIXED

Fix applied on 2026-05-05. Decision-log e2e tests were moved from `tests/e2e/decision-log/` to `tests/e2e/items/` with renamed spec files (`decision-log-api.spec.ts`, `decision-log-ui.spec.ts`).

---

# E2E beforeAll Failure Cascades and Misdiagnosis

## Problem

Decision-log e2e tests (28 total) failed with 3 passed, 19 failed, 6 skipped. The task-executor agent diagnosed the root cause as "Windows Playwright beforeAll fixture libuv handle-closing error" — but this was a **misdiagnosis**. The real issue was never properly identified despite 3 retry attempts.

Key symptoms:
- `teamId` and `mainItemId` variables were `undefined` in tests
- All API calls went to `/v1/teams/undefined/main-items/undefined/decision-logs` → 404
- Only tests that didn't depend on `beforeAll` fixtures passed (TC-025: unauthenticated 401 check)

## Root Cause

Causal chain (4 levels deep):

1. **Symptom**: 19/28 tests fail with 404 on undefined IDs
2. **Direct cause**: `beforeAll` throws, module-level `let` variables (`teamId`, `mainItemId`) stay `undefined`
3. **Secondary error**: The "Windows libuv handle-closing" assertion is a **side effect** of Node.js async cleanup failure, NOT the root cause — it appears when an async operation throws during `beforeAll` and the error propagates through libuv's handle cleanup
4. **Root cause** (unconfirmed but most likely): The agent either (a) didn't properly restart the backend after code changes (disc-1/disc-2 fixes), or (b) the backend wasn't ready when `beforeAll` ran, or (c) rate limiting (429) from repeated test runs caused `createTestTeam`/`createTestMainItem` to fail — which then triggered the libuv assertion as a secondary crash

**Evidence that "Windows libuv" was NOT the root cause:**
- 551 other e2e tests in the same project use `beforeAll` + `fetch` and **pass** on Windows
- `item-pool/pool-api.spec.ts` uses identical `getApiToken` + `createAuthCurl` pattern and passes
- The task-executor's "manual API verification" confirmed endpoints work — meaning the backend WAS running

**Why the agent misdiagnosed:** The libuv assertion error is the **first error message visible** in Playwright output. The agent stopped at this surface symptom without investigating what actually caused `beforeAll` to fail.

## Solution

### Investigation approach that should have been used

1. **Add logging to beforeAll**: Temporarily add `console.log` statements in `beforeAll` to see which step fails
2. **Reproduce manually**: Run the exact `beforeAll` code outside Playwright to isolate the issue
3. **Check backend logs**: Verify the backend received and responded to the `createTestTeam`/`createTestMainItem` requests
4. **Run with `--debug`**: Use `npx playwright test decision-log/api.spec.ts --debug` to step through `beforeAll`

### Pattern comparison

Passing tests use one of two patterns:
- `test.describe.serial` with `playwright.request.newContext()` (items/item-list.spec.ts)
- `beforeAll` with `createAuthCurl` but **reuse existing team** (item-pool/pool-api.spec.ts)

Failing test pattern:
- `beforeAll` with `createAuthCurl` + **creates new team + main item** (items/decision-log-api.spec.ts)

The extra create operations introduce more failure modes.

## Reusable Pattern

### 1. Agent misdiagnosis is common — verify surface-level explanations

When an agent reports "Windows libuv error" or "environment issue", treat it as unconfirmed. These are often secondary errors masking the real root cause. Always ask: "Why does this test fail when 500 other tests with the same pattern pass?"

### 2. beforeAll must not leave module variables undefined

Use defensive patterns:
```typescript
// BAD — variables stay undefined if beforeAll throws
let teamId: string;
test.beforeAll(async () => {
  teamId = await createTestTeam(token, name);
});

// GOOD — fail explicitly with diagnostic info
test.beforeAll(async () => {
  try {
    teamId = await createTestTeam(token, name);
  } catch (e) {
    console.error('beforeAll failed at createTestTeam:', e);
    throw e; // re-throw so Playwright marks beforeAll as failed
  }
  if (!teamId) throw new Error('teamId is undefined after createTestTeam');
});
```

### 3. Task file `just test-e2e` syntax was wrong

The task definition file said `just test-e2e --feature decision-log` but the justfile uses positional args: `just test-e2e decision-log`. Task files should reference the correct command syntax.

### 4. Backend lifecycle matters for e2e

After code changes (disc-1, disc-2 fixes), the backend binary must be rebuilt AND restarted before e2e re-runs. The task-executor may not have reliably done this.

## Example

```bash
# Correct justfile syntax (positional arg, no flag)
just test-e2e decision-log

# Debug a specific failing test
cd tests/e2e && npx playwright test items/decision-log-api.spec.ts --debug

# Check if backend is running and responsive
curl http://localhost:8080/health
```

## Related Files

- `tests/e2e/items/decision-log-api.spec.ts` — failing test (beforeAll pattern)
- `tests/e2e/items/decision-log-ui.spec.ts` — failing test (beforeAll pattern)
- `tests/e2e/item-pool/pool-api.spec.ts` — passing test (similar pattern, reuses team)
- `tests/e2e/items/item-list.spec.ts` — passing test (serial + request context)
- `tests/e2e/helpers.ts` — shared helpers (curl, createTestTeam, createTestMainItem)
- `docs/features/decision-log/tasks/run-e2e-tests.md` — task definition (wrong just test-e2e syntax)
- `justfile` — test-e2e recipe definition

## References

- Node.js libuv handle-closing assertion: known Windows issue in Node.js < 22, usually a secondary effect
- Playwright `test.describe.serial`: ensures tests run in order with shared state

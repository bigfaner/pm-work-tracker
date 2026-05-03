# External CLI tool in E2E tests causes cascade timeout inflation

## Problem

Full E2E test suite (373 tests) ran for 50+ minutes without producing output. After investigation, 133 browser-based tests were timing out at 60+ seconds each because they called an external CLI tool (`agent-browser`) that was not installed on the machine. This inflated total runtime by 133 × 60s ≈ 133 minutes of pure timeout waste.

## Root Cause

Causal chain (3 levels deep):

1. **Symptom**: Test suite takes 50+ minutes, no visible output
2. **Direct cause**: `execSync('npx agent-browser ...')` in test helpers hangs for 30-60s per call, then throws. Each test calls this 3-5 times → ~3 min per test × 133 tests
3. **Root cause**: Test scripts were generated to use `agent-browser` (an external browser automation CLI) instead of Playwright's built-in `page` fixture. When `agent-browser` is unavailable, `execSync` blocks the Node.js event loop waiting for a process that eventually fails
4. **Trigger condition**: Running E2E tests in an environment without `agent-browser` installed — which is the default for any standard CI or developer machine

## Solution

Rewrite all `agent-browser` based tests to use standard Playwright browser automation:

| agent-browser | Playwright |
|---|---|
| `ab('open URL')` | `await page.goto(URL)` |
| `ab('wait --load networkidle')` | `await page.waitForLoadState('networkidle')` |
| `ab('click REF')` | `await locator.click()` |
| `ab('fill REF VALUE')` | `await locator.fill(VALUE)` |
| `snapshotContains(text)` | `await page.getByText(text).isVisible()` |
| `findElement(role, name)` | `page.getByRole(role, { name })` |
| `ab('close')` | Remove (Playwright manages lifecycle) |
| `browserLogin(u, p)` | `await page.goto('/login'); await fill; await click` |

## Key Takeaway

**Never use an external CLI tool for browser automation in Playwright tests.** Playwright already provides the `page` fixture in every test — use it directly. External CLI tools add a dependency, block the event loop on `execSync`, and produce opaque timeout failures when unavailable.

When generating E2E test scripts, ensure they use only `@playwright/test` APIs: `page.goto()`, `page.getByRole()`, `page.getByText()`, `locator.click()`, `locator.fill()`, `expect().toBeVisible()`, etc.

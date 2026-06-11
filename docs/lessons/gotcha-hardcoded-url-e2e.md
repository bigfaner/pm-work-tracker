---
created: "2026-06-11"
tags: [testing, e2e, anti-pattern]
---

# E2E Tests: Hardcoding IP/Port in page.goto Instead of Using baseUrl

## Problem

When E2E tests fail with `page.goto` errors like "invalid URL" (caused by passing a relative path instead of an absolute URL), the tempting quick fix is to hardcode `http://127.0.0.1:5173` as the URL prefix. This "works locally" but breaks port flexibility, multi-environment support, and CI pipelines.

## Root Cause

**L1 — Symptom**: `page.goto('/some-path')` throws "invalid URL" because Playwright requires absolute URLs.

**L2 — Quick-fix trap**: The developer hardcodes `http://127.0.0.1:5173/some-path` to make it work immediately, instead of using the existing `baseUrl` export from the test helpers.

**L3 — Existing infrastructure ignored**: The helpers file already exports `baseUrl` (read from `config.yaml` with a sensible fallback). Hardcoding bypasses this infrastructure entirely.

## Solution

Import and use `baseUrl` from the test helpers:

```javascript
import { baseUrl } from './helpers';

// Correct
await page.goto(`${baseUrl}/some-path`);

// Wrong — never do this
await page.goto('http://127.0.0.1:5173/some-path');
```

The helpers file constructs `baseUrl` from `config.yaml`, so the same test suite works across dev, CI, and any custom port configuration.

## Reusable Pattern

- **Always use the project's `baseUrl` helper for navigation URLs in E2E tests.** Never hardcode host, IP, or port.
- **When a test helper already exports a value, use it.** Search existing helpers before inventing ad-hoc values.
- **Verify the helpers module before writing navigation code.** A `baseUrl` export with config-file fallback is a standard pattern in this project.

## Related Lessons

- [[gotcha-e2e-script-generation]] — broader E2E script generation anti-patterns (path/port guessing during generation)

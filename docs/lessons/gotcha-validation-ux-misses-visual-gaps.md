---
created: "2026-06-08"
tags: [testing, architecture, interface]
---

# validation-UX Task Misses Visual/Interaction Gaps

## Problem

After running the complete pipeline (prd → design → tasks → code → eval → test → validate-code → validate-ux) for the milestone-map feature, manual QA found 6 issues that validation-ux did not catch:

1. **Status filter style inconsistency**: List page uses a different filter component style than detail page
2. **Refresh button wrong position**: Should be right of reset button per design spec, but placed elsewhere
3. **Refresh button no feedback**: Clicking refresh refreshes data but shows no toast/notification
4. **Date picker broken**: Calendar popup doesn't appear when clicking date fields across all forms
5. **API 404 on available-transitions**: Wrong URL pattern — uses `/main-items/:id/available-transitions` instead of `/milestone-maps/:id/available-transitions`
6. **Visual regression undetected**: validation-ux reported "all 7 UI functions match design spec" with only minor notes

## Root Cause

Three levels deep:

**L1 — validation-ux task is structural, not visual**: The task checks component existence and code-level design token usage (reads source code, matches component names), but does NOT verify visual layout, positioning, or interaction behavior. It validates the code matches the spec on paper, not the rendered output.

**L2 — No screenshot/visual comparison step**: The pipeline lacks a step that renders the actual UI and compares against design mockups. validation-ux relies on code reading + test passing, which cannot catch: button positioning, style consistency across pages, calendar popup rendering, or missing toast notifications.

**L3 — E2E tests verify structure, not behavior**: Generated E2E tests click buttons and check API responses but don't verify: visual consistency across pages, toast appearance, calendar popup rendering, or correct component variant usage.

## Solution

Immediate fixes needed for the 6 reported issues. Long-term: validation-ux should include a manual visual QA checklist or screenshot comparison step, not just code-level verification.

## Reusable Pattern

- **validation-ux ≠ visual QA**: Never trust validation-ux to catch layout, positioning, style consistency, or interaction issues. It validates code structure, not rendered output.
- **Design spec compliance requires visual check**: After all code tasks complete, manually verify: button positions, component variants (same filter type across pages), interactive elements (calendars, dropdowns, toasts), and API URL patterns.
- **E2E tests should verify real content**: Tests that pass with zero milestones in every map are not meaningful. Seed data must match realistic scenarios.
- **API URL patterns must be checked**: Wrong path segments (main-items vs milestone-maps) cause 404s that E2E tests may not exercise.

## Related Files

- `docs/features/milestone-map/tasks/validate-ux.md`
- `docs/features/milestone-map/ui/ui-design.md`

## References

- Related: `gotcha-ac-self-report-without-verification.md` — agents self-report success without verification
- Related: `gotcha-e2e-script-generation.md` — generated E2E scripts have structural gaps

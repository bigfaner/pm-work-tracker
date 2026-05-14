---
name: Browser autofill breaks React controlled inputs
description: Browser autofill on native <input type="date"> sets DOM value without firing React onChange, causing state/display mismatch. E2E tests miss this because fill() properly dispatches events.
type: gotcha
tags: testing, interface
---

# Browser Autofill Breaks React Controlled Inputs

## Problem

A form dialog shows a date value in the input field ("2026/05/14"), but clicking submit triggers the validation error "请选择计划完成时间" (please select planned completion date). The displayed value and the validated state are out of sync.

User-visible symptom: filled-in field that the form claims is empty.

## Root Cause

**Causal chain (3 levels deep):**

1. **Symptom**: DateInput displays "2026/05/14" but validation says `!endDate` is true
2. **Direct cause**: React state `endDate` is `''` while the DOM `<input>` value is "2026-05-14"
3. **Root cause**: Chrome's autofill feature sets the DOM value on `<input type="date">` without firing `input`/`change` events that React's synthetic event system listens to. Since `onChange` never fires, `setEndDate()` is never called, and React state remains empty.
4. **Trigger condition**: The DateInput component has no `autoComplete` attribute, so Chrome autofills today's date into any recognized date input.

The DateInput component (`components/ui/date-input.tsx`) is a thin wrapper around native `<input type="date">` with controlled `value` + `onChange`. It works correctly when users interact via the date picker, but browser autofill bypasses React's event system entirely.

## Solution

Add `autoComplete="off"` to all DateInput instances (or add it as a default in the DateInput component itself):

```tsx
// Option A: Fix in DateInput component (preferred — single point fix)
<input
  type="date"
  autoComplete="off"  // <-- add this
  ref={forwardedRef}
  className="..."
  {...props}
/>

// Option B: Pass via props at each usage site
<DateInput
  autoComplete="off"
  value={endDate}
  onChange={...}
/>
```

Option A is preferred because it prevents the issue for every usage without each consumer needing to remember.

## Reusable Pattern

**When using native form inputs (`<input type="date">`, `<input type="email">`, etc.) with React controlled state, always add `autoComplete="off"`.** Browser autofill sets DOM values without triggering React's synthetic onChange, creating a state/display mismatch that validation catches but users find confusing.

**Testing gap**: E2E tests using Playwright's `fill()` will NEVER catch this bug. `fill()` dispatches proper input/change events that React listens to. But real users encounter browser autofill which bypasses React's event system. This is a systematic blind spot in browser automation testing.

**How to apply**: When reviewing code that uses native `<input>` elements with React controlled state, check for `autoComplete` attributes. If missing, add `"off"` to prevent autofill. For shared components (like DateInput), set it as a default in the component itself.

## Example

```tsx
// BAD: Browser can autofill, breaking React state
<input type="date" value={date} onChange={handleChange} />

// GOOD: Prevent autofill from interfering with React state
<input type="date" autoComplete="off" value={date} onChange={handleChange} />
```

## Related Files

- `frontend/src/components/ui/date-input.tsx` — DateInput component (needs `autoComplete="off"`)
- `frontend/src/pages/milestones/CreateEditMilestoneDialog.tsx` — Dialog where bug was observed
- `tests/e2e/features/milestone-map/milestones-page.spec.ts` — E2E tests that passed despite the bug

## References

- [React controlled components docs](https://react.dev/reference/react-dom/components/input#controlling-an-input-with-a-state-variable)
- [Chrome autofill behavior](https://bugs.chromium.org/p/chromium/issues/detail?id=468153) — autofill doesn't fire standard events

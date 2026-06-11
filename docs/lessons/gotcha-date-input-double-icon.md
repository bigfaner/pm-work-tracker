---
created: "2026-06-08"
tags: [interface, testing]
---

# Date Input Fix Created Double Calendar Icon

## Problem

After fixing the hidden calendar picker by removing `[&::-webkit-calendar-picker-indicator]:hidden`, the date input now shows TWO calendar icons: the native browser picker indicator AND the decorative Lucide Calendar icon. Clicking the decorative icon (rightmost) does nothing because it has `pointer-events-none`.

## Root Cause

**L1 — Incomplete fix**: The original bug was "calendar doesn't appear." The fix removed the CSS hide rule, which restored the native picker. But the decorative Calendar icon (Lucide) was added as a workaround when the native picker was hidden. Now both exist.

**L2 — No visual verification after fix**: The fix was committed based on code reading and TypeScript compilation, without visually verifying the rendered component in a browser.

## Solution

Remove the decorative Lucide Calendar icon since the native picker indicator is now visible. The native indicator provides the same visual cue and is clickable.

## Reusable Pattern

- **Verify UI fixes visually**: After fixing CSS/component changes that affect visual rendering, verify in browser. TypeScript compilation passing does not mean the UI looks correct.
- **Remove workarounds when fixing root cause**: When a fix addresses the root cause (un-hiding native picker), also remove any workarounds added for the original symptom (decorative icon).

## Related Files

- `frontend/src/components/ui/date-input.tsx`

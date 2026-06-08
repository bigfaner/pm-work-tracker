---
created: "2026-06-08"
tags: [interface, architecture]
---

# Radix Component Used Without Required Provider

## Problem

Editing a milestone and saving causes a blank page crash: `Uncaught Error: Tooltip must be used within TooltipProvider`. The milestone detail panel uses Tooltip components (from Radix UI) but the component tree does not wrap them in `<TooltipProvider>`.

## Root Cause

**L1 — Component added without provider**: A developer added `<Tooltip>` usage inside MilestoneDetailPanel or its children, but did not add `<TooltipProvider>` to the component tree. Radix UI requires provider wrappers for Tooltip, Dialog, etc.

**L2 — No compile-time or test-time check**: TypeScript compilation passes because Tooltip is a valid JSX element. Unit tests mock or shallow-render components, skipping the provider requirement. E2E tests don't exercise the save-and-rerender path that triggers Tooltip rendering.

**L3 — validation-ux missed it (again)**: validation-ux reads code but does not render pages and interact. The TooltipProvider absence is invisible until runtime user interaction.

## Solution

Wrap the component tree with `<TooltipProvider>` at the appropriate level (layout or page component). Verify by running the dev server and exercising the full interaction.

## Reusable Pattern

- **Radix provider checklist**: When using Radix UI components (Tooltip, Dialog, Popover, DropdownMenu), ensure the corresponding Provider is in the component tree. Common missing providers: `TooltipProvider`, `ToastProvider`.
- **Test the save path**: E2E tests should exercise create → save → verify-render cycle, not just create → assert-response.
- **Provider placement**: Prefer adding providers at the layout level (AppLayout) so all pages inherit them, rather than per-component.

## Related Files

- `frontend/src/pages/milestones/MilestoneDetailPanel.tsx`
- `frontend/src/App.tsx` or `frontend/src/components/layout/AppLayout.tsx`

## References

- Related: `gotcha-validation-ux-misses-visual-gaps.md` — validation-ux cannot catch runtime errors from missing providers

---
created: "2026-06-08"
tags: [interface]
---

# Milestone map card layout deviates from UI design

## Problem

The milestone map list page cards do not follow the UI design specification. The most visible deviation is Row 4 (milestone node thumbnail): the spec requires dot-and-line thumbnail with first/last milestone names, but the implementation uses pill-shaped badges with per-milestone name and percentage.

Secondary deviations:
- Row 2 merges milestone count and item count with a middle-dot separator instead of showing three separate spans
- `in_progress` status color in Row 4 uses `bg-warning` (yellow) instead of the spec's blue/info variant

## Root Cause

1. **Symptom**: Card visual output doesn't match the prototype/screenshot
2. **Direct cause**: Developer implemented an alternative visual approach (pill badges) instead of the specified dot-and-line thumbnail without flagging the deviation
3. **Process gap**: No visual verification step against the UI design spec or HTML prototype (`ui/prototype/milestones.html`) was performed after implementation
4. **Underlying pattern**: When a visual element is complex (dot-and-line thumbnail), developers tend to substitute a simpler alternative that they're more comfortable implementing, without raising the deviation for discussion

## Solution

Re-implement Row 4 to match the dot-and-line thumbnail specification:
- Horizontal line of status-colored dots connected by lines
- First milestone name on the left, last milestone name on the right
- No per-milestone name/percentage in the thumbnail (those details belong in detail views)

Fix Row 2 to show three separate spans matching the prototype, and correct `in_progress` color mapping.

## Reusable Pattern

When implementing UI components from design specs, especially visual elements with no standard library equivalent (custom thumbnails, diagrams, timelines):
1. **Check the prototype first** — `ui/prototype/*.html` files contain the visual ground truth
2. **Implement exactly what the spec says** — if it's too hard, raise it for discussion rather than silently substituting
3. **Visual verification** — after implementation, compare the rendered output side-by-side with the prototype/screenshot
4. **Design deviations are decisions** — if you choose to deviate, document it as a decision with rationale

## Related Files

- `docs/features/milestone-map/ui/ui-design.md` — card layout specification
- `docs/features/milestone-map/ui/prototype/milestones.html` — visual prototype
- `frontend/src/pages/milestones/MilestoneMapCard.tsx` — card component implementation

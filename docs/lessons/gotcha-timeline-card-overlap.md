---
created: "2026-06-08"
tags: [architecture, interface]
---

# Timeline Card Overlap — No Collision Detection for Date-Proximate Nodes

## Problem

On the milestone map detail page, when two milestones have close or identical `expectedEndDate` values, their timeline cards (160px wide) visually overlap. All cards render at `absolute top-0` with only a horizontal `left` offset derived from date-to-pixel mapping, producing a single-row layout with no collision handling.

## Root Cause

1. **Surface**: Two milestone cards overlap on the timeline — unreadable and unclickable.
2. **Direct cause**: `calculateNodePositions` maps dates to X via `x = days / totalDays * containerWidth`, all nodes render `absolute top-0`. Same date → same X → overlap.
3. **Design spec is self-contradictory**: The UI design (`ui-design.md`) simultaneously requires (a) date-proportional positioning and (b) `minWidth = nodeCount * 184` to guarantee no overlap. These are incompatible for clustered dates — `minWidth` only prevents overlap if nodes are evenly spaced, but date-proportional positioning clusters nodes when dates cluster. The spec declares "不会出现节点重叠" without specifying a collision-resolution fallback.
4. **Executor followed spec literally but not logically**: Task 3.6 implemented both formulas exactly as written. Spec-Code Scan is textual matching ("does code contain formula X?") not logical analysis ("do formulas X + Y together achieve the stated goal?"). The executor is instructed to treat Reference Files as authority (Spec Authority Enforcement), with no obligation to detect internal contradictions.
5. **validate-ux is structural, not visual**: The pipeline's final quality gate checks code-to-spec text alignment, not rendered output. It reported "All 7 UI functions match design spec" without ever rendering the page. No pipeline step performs visual QA.

## Solution

Add collision detection after initial date-based X calculation:
1. Sort nodes by X position.
2. Iterate left-to-right, pushing any node whose left edge overlaps the previous node's right edge down to the next row (or rightward to minimum gap).
3. Track the row count to set the container's `minHeight` dynamically.

## Reusable Pattern

When implementing timeline/calendar/Gantt layouts with absolute positioning:
- Always compute positions in two passes: first the ideal (date-based) positions, then a collision-resolution pass that staggers overlapping items into multiple rows or enforces minimum gaps.
- Test with edge cases: same date, adjacent dates, many items in a narrow container.
- Verify the rendered output visually, not just the code structure — unit tests and type checks cannot catch layout overlap.
- When a design spec specifies multiple positioning constraints, check that they are logically compatible — "date-proportional" + "no overlap guarantee" can contradict for clustered dates. If they conflict, flag it and add a collision-detection fallback.

## Related Files

- `frontend/src/pages/milestones/MilestoneTimeline.tsx` — `calculateNodePositions` (line 97) and node rendering (line 648)
- `frontend/src/pages/milestones/MilestoneNode.tsx` — card component, `w-40` (160px wide)

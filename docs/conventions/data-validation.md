---
scope: global
source: feature/improve-ui BIZ-004, BIZ-005, BIZ-008
---

# Data Validation Rules

## BIZ-validation-001: Item Validation

For all items (main items and sub items):
- **Title**: required
- **Date range**: end date must be ≥ start date
- **Progress percentage**: must be ≥ last recorded value

## BIZ-validation-002: Weekly View Time Constraint

Weekly progress view does not accept future week dates. Requesting a future week returns `FUTURE_WEEK_NOT_ALLOWED` (422).

Progress data only exists for past and current weeks.

## BIZ-validation-003: Item Sorting Convention

Items are sorted by priority ascending (P1 → P2 → P3). Items with the same priority are sorted by deadline ascending (earliest first).

Applied consistently across: item list, table view, weekly view groups.

## BIZ-validation-004: Multi-Select Status Filter

Status filter parameters accept comma-separated values (e.g., `?status=progressing,blocking`), parsed as `[]string` via the `form:"status"` binding tag. Single-value queries continue to work as one-element slices. Applies to:
- MainItemFilter (item list view)
- GanttFilter (gantt view)

**Source**: feature/system-ux-optimization TECH-006

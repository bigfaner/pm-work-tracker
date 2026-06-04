---
title: "Item Lifecycle Rules"
domains: [terminal, sorting, status, move, lifecycle, item]
---

# Item Lifecycle Rules

Business rules governing item status transitions, terminal state behavior, and structural operations (delete, move).

Source: feature/system-ux-optimization

## Terminal Status

### BIZ-lifecycle-001: Terminal Status Definition for Main Items

Main items with `item_status` = `completed` or `closed` are terminal. Terminal status determines sorting, filtering, and visibility behavior across all views.

**Cross-view implications**:
- Item list: terminal items sort to bottom
- Progress page: terminal items sort to bottom, filterable by status
- Gantt view: terminal items filterable by status, excluded from default view
- Weekly view: terminal items hidden when inactive (no recent activity)

**Source**: feature/system-ux-optimization BIZ-001

### BIZ-lifecycle-002: Terminal Items Sort to Bottom

In item list and progress views, terminal main items are sorted after all non-terminal items, preserving relative order within each group. Implemented via `sort.SliceStable` with `status.IsMainTerminal()` check appended after existing sort criteria.

**Source**: feature/system-ux-optimization BIZ-002

## Sub-Item Move

### BIZ-lifecycle-003: Move Preserves Status and Assignee

When a sub-item is moved to a different main item, its `item_status` and `assignee_key` remain unchanged. Only `main_item_key` and `item_code` are updated. Move is a re-organization operation, not a status change.

**Source**: feature/system-ux-optimization BIZ-006

### BIZ-lifecycle-004: Move Target Validation

Sub-item move targets must satisfy all of:
1. Target main item is not terminal (`item_status` not `completed` or `closed`)
2. Target is not the same main item the sub-item currently belongs to
3. Target belongs to the same team (`team_key == current teamBizKey`)

Violations return `BAD_REQUEST` (400) with specific error messages.

**Source**: feature/system-ux-optimization BIZ-007

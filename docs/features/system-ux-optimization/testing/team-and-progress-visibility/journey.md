---
feature: "system-ux-optimization"
journey: "team-and-progress-visibility"
risk_level: "Low"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 10, Story 11)
  - docs/features/system-ux-optimization/prd/prd-spec.md
generated: "2026-06-04"
---

# Journey: team-and-progress-visibility

**Risk Level**: Low

<!-- Risk Classification Criteria:
  Low = Workflow is read-only or purely observational
  Team selector filtering and weekly progress page filtering are read-only display operations.
-->

## Overview

User sees only teams they have permission to in the team selector, and the weekly progress page hides terminal main items with no activity in the current or previous week.

## Setup

- User belongs to Team A and Team B, but not Team C
- Team C exists in the system and other users are members
- Multiple main items exist across teams with various statuses and activity levels

## Happy Path

### Step 1: View team selector with permission filtering

**User Action**: User opens the team dropdown selector after login

**Expected Result**: Only teams the user has permission to access are displayed (Team A and Team B); Team C is not shown in the dropdown

### Step 2: View weekly progress page with active items

**User Action**: User opens the weekly progress page; non-terminal main items exist with activity this week

**Expected Result**: All non-terminal main items are displayed; terminal main items with activity (status change or sub-item update) this week or last week are displayed; terminal main items with no activity this week or last week are hidden

### Step 3: View weekly progress page with no active terminal items

**User Action**: User opens the weekly progress page; all terminal items had no activity this week or last week

**Expected Result**: Only non-terminal main items are displayed; terminal items are hidden; page layout is clean without gaps

## Edge Cases

### Step 1b: User with access to only one team

**Precondition**: User has permission to exactly one team

**User Action**: User opens the team selector

**Expected Result**: Only that one team is shown; the selector may still function as a dropdown for consistency but contains a single entry

### Step 2b: Terminal item with activity in current week only

**Precondition**: A completed main item had a sub-item updated this week but no activity last week

**User Action**: User views the weekly progress page

**Expected Result**: The completed main item is displayed because it has activity in the current week (sub-item update falls within the active definition)

### Step 3b: Terminal item with activity in last week only

**Precondition**: A closed main item had a status change last week but no activity this week

**User Action**: User views the weekly progress page

**Expected Result**: The closed main item is displayed because it had activity last week (status_history record within range)

## Journey Invariants

- Team selector always filters to only teams the current user has permission to access
- Non-terminal main items are always displayed on the weekly progress page regardless of activity
- Terminal main items are hidden only when they have zero activity in both the current week (Mon-Sun) and the previous week
- Activity definition for terminal items includes: status_history changes, sub-item creation/update, and main item progress updates within the time range
- The weekly progress page uses natural week boundaries (Monday through Sunday) for activity calculation

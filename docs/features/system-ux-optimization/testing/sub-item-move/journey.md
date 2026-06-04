---
feature: "system-ux-optimization"
journey: "sub-item-move"
risk_level: "High"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 7)
  - docs/features/system-ux-optimization/prd/prd-spec.md
generated: "2026-06-04"
---

# Journey: sub-item-move

**Risk Level**: High

<!-- Risk Classification Criteria:
  High = Workflow involves state mutation, data loss risk, or irreversible operations
  Moving a sub-item changes its parent relationship and reassigns its code number.
-->

## Overview

PM user moves a sub-item from one main item to another, with automatic re-numbering under the target main item while preserving status and assignee.

## Setup

- PM user is logged in with appropriate permissions
- A sub-item exists under main item A
- Another main item B exists in a non-terminal (non-closed, non-completed) status

## Happy Path

### Step 1: Initiate sub-item move

**User Action**: PM user selects "Move to another main item" from the sub-item detail view

**Expected Result**: A target main item selector is displayed, listing available (non-closed, non-completed) main items excluding the current parent

### Step 2: Select valid target and confirm

**User Action**: PM user selects main item B as the target and confirms the move

**Expected Result**: Sub-item is moved to main item B; code number is auto-regenerated using target's NextSubCode; status and assignee remain unchanged; move executes in a single transaction

### Step 3: Verify sub-item in new location

**User Action**: PM user navigates to main item B's detail page

**Expected Result**: The moved sub-item appears in the sub-item list under main item B with its new code number; the sub-item no longer appears under main item A

## Edge Cases

### Step 1b: Move to a closed main item

**Precondition**: A main item C is in terminal status (closed or completed)

**User Action**: PM user attempts to select main item C as the move target

**Expected Result**: The operation is rejected; a message indicates that the target main item is closed and cannot receive sub-items; the closed main item may be shown as disabled/non-selectable in the selector

### Step 2b: Move to the same main item

**Precondition**: Sub-item currently belongs to main item A

**User Action**: PM user attempts to move the sub-item to main item A (same parent)

**Expected Result**: The operation is rejected; a message indicates the sub-item is already under this main item; no change occurs

### Step 3b: Concurrent move of same sub-item

**Precondition**: Two PM users simultaneously attempt to move the same sub-item to different targets

**User Action**: Both users confirm the move at nearly the same time

**Expected Result**: The first transaction succeeds; the second transaction either succeeds with the correct re-numbering under its target (since NextSubCode increments per request) or fails gracefully if the sub-item's source has been modified; no data corruption occurs

### Step 4b: Move sub-item from a deleted main item

**Precondition**: Main item A has been soft-deleted by another user while the current user has the move dialog open

**User Action**: PM user confirms the move

**Expected Result**: Backend returns "source main item does not exist" error; frontend displays the error message; no move is performed

### Step 5b: Unauthorized move attempt via API

**Precondition**: A member-role user sends a move request via API

**User Action**: Member user sends POST /api/sub-items/:id/move

**Expected Result**: API returns 403 Forbidden; no data is modified

## Journey Invariants

- Sub-item code number is always auto-regenerated using the target main item's NextSubCode mechanism
- Status and assignee of the moved sub-item are never changed by the move operation
- Move and code re-generation always execute within a single database transaction
- Closed/completed main items are never valid targets for sub-item moves
- Moving to the same parent main item is always rejected

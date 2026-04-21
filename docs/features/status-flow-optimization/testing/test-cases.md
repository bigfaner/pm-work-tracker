---
feature: "status-flow-optimization"
sources:
  - prd/prd-user-stories.md
  - prd/prd.md
generated: "2026-04-21"
---

# Test Cases: Status Flow Optimization

## Summary

| Type | Count |
|------|-------|
| UI   | 22    |
| API  | 35    |
| CLI  | 0     |
| **Total** | **57** |

---

## UI Test Cases

## TC-001: Status badge displays correct Chinese name
- **Source**: US-1 / AC-1
- **Type**: UI
- **Pre-conditions**: MainItem or SubItem exists with any status code
- **Steps**:
  1. Navigate to any list page, detail page, or weekly view
  2. Observe the status badge for an item with status `progressing`
- **Expected**: Badge shows "进行中", not a fallback/default style
- **Priority**: P0

## TC-002: All status codes render correct Chinese names
- **Source**: US-1 / AC-1; Spec R1
- **Type**: UI
- **Pre-conditions**: Items exist with each of the 7 MainItem status codes
- **Steps**:
  1. View items with each status: `pending`, `progressing`, `blocking`, `pausing`, `reviewing`, `completed`, `closed`
  2. Observe the status badge for each
- **Expected**: Each badge shows the correct Chinese name: 待开始, 进行中, 阻塞中, 已暂停, 待验收, 已完成, 已关闭. No fallback/default styling appears.
- **Priority**: P0

## TC-003: StatusDropdown calls ChangeStatus API on selection
- **Source**: US-13 / AC-1; Spec AC-16
- **Type**: UI
- **Pre-conditions**: User is on an item detail page; item is in `pending` state
- **Steps**:
  1. Open the status dropdown
  2. Select `progressing`
  3. Observe the displayed status
- **Expected**: ChangeStatus API is called; displayed status updates to "进行中"
- **Priority**: P0

## TC-004: StatusDropdown shows only valid transitions for current status
- **Source**: US-14 / AC-1; Spec AC-17
- **Type**: UI
- **Pre-conditions**: MainItem is in `blocking` state
- **Steps**:
  1. Open the status dropdown on the item detail page
- **Expected**: Only `progressing` (进行中) is shown as an option
- **Priority**: P0

## TC-005: StatusDropdown for reviewing state — PM role
- **Source**: US-14 / AC-1; US-5 / AC-3
- **Type**: UI
- **Pre-conditions**: MainItem is in `reviewing` state; logged-in user is a PM
- **Steps**:
  1. Open the status dropdown
- **Expected**: Only `completed` (已完成) and `progressing` (进行中) are shown
- **Priority**: P0

## TC-006: StatusDropdown for reviewing state — non-PM role
- **Source**: US-14 / AC-2; US-5 / AC-3
- **Type**: UI
- **Pre-conditions**: MainItem is in `reviewing` state; logged-in user is NOT a PM
- **Steps**:
  1. Open the status dropdown
- **Expected**: No options are shown; dropdown is empty or disabled
- **Priority**: P0

## TC-007: StatusDropdown disabled for terminal states
- **Source**: US-14 / AC-4
- **Type**: UI
- **Pre-conditions**: MainItem is in `completed` or `closed` state
- **Steps**:
  1. Open the status dropdown
- **Expected**: No options are shown; dropdown is disabled
- **Priority**: P0

## TC-008: Overdue indicator shown for non-terminal overdue item
- **Source**: US-15 / AC-1; Spec AC-19
- **Type**: UI
- **Pre-conditions**: Item has `expected_end_date` set to yesterday; status is `progressing`
- **Steps**:
  1. View the item in a list or detail view
- **Expected**: An overdue badge/indicator is displayed alongside the status
- **Priority**: P1

## TC-009: No overdue indicator for terminal state item
- **Source**: US-15 / AC-2; Spec AC-19
- **Type**: UI
- **Pre-conditions**: Item has `expected_end_date` in the past; status is `completed` or `closed`
- **Steps**:
  1. View the item in a list or detail view
- **Expected**: No overdue indicator is shown
- **Priority**: P1

## TC-010: No overdue indicator when expected_end_date is today or future
- **Source**: US-15 / AC-3; Spec AC-19
- **Type**: UI
- **Pre-conditions**: Item has `expected_end_date` set to today or tomorrow; status is `progressing`
- **Steps**:
  1. View the item in a list or detail view
- **Expected**: No overdue indicator is shown
- **Priority**: P1

## TC-011: Confirmation dialog appears before completing or closing
- **Source**: US-16 / AC-1; Spec AC-21
- **Type**: UI
- **Pre-conditions**: Item is in a non-terminal state; user selects `completed` or `closed` from dropdown
- **Steps**:
  1. Open the status dropdown
  2. Select `completed`
- **Expected**: A confirmation dialog appears explaining the action is irreversible; no API call is made yet
- **Priority**: P0

## TC-012: Cancel on confirmation dialog aborts status change
- **Source**: US-16 / AC-2; Spec AC-21
- **Type**: UI
- **Pre-conditions**: Confirmation dialog is shown for `completed` transition
- **Steps**:
  1. Click "Cancel" in the confirmation dialog
- **Expected**: Status change is not executed; item remains in its current state
- **Priority**: P0

## TC-013: Confirm on confirmation dialog executes status change
- **Source**: US-16 / AC-3; Spec AC-21
- **Type**: UI
- **Pre-conditions**: Confirmation dialog is shown for `completed` transition
- **Steps**:
  1. Click "Confirm" in the confirmation dialog
- **Expected**: ChangeStatus API is called; status changes to `completed`
- **Priority**: P0

## TC-014: Toast warning shown when linkage fails
- **Source**: US-9 / AC-1; Spec AC-12
- **Type**: UI
- **Pre-conditions**: MainItem is in `completed` state (terminal); a SubItem status change triggers linkage evaluation targeting a transition from `completed`
- **Steps**:
  1. Change a SubItem status to trigger linkage
  2. Observe the UI
- **Expected**: A toast warning is displayed notifying the user that automatic linkage could not be applied
- **Priority**: P1

## TC-015: StatusBadge uses code-to-name mapping, not hardcoded Chinese
- **Source**: Spec AC-18
- **Type**: UI
- **Pre-conditions**: Application is running
- **Steps**:
  1. View any item with a status badge
  2. Inspect the rendering logic (or verify via code review that no hardcoded Chinese strings are used)
- **Expected**: StatusBadge derives display text from a code-to-name mapping; no hardcoded Chinese values in component
- **Priority**: P0

## TC-016: Reviewing → progressing/completed options hidden for non-PM in UI
- **Source**: US-5 / AC-3; Spec AC-20
- **Type**: UI
- **Pre-conditions**: MainItem is in `reviewing` state; logged-in user is non-PM
- **Steps**:
  1. Open the status dropdown
- **Expected**: `completed` and `progressing` options are not visible
- **Priority**: P0

---

## API Test Cases

## TC-017: MainItem valid transition — pending → progressing
- **Source**: US-2 / AC-1; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `pending` state
- **Steps**:
  1. Call `POST /api/v1/teams/:teamId/items/:itemId/change-status` with `{"status": "progressing"}`
- **Expected**: HTTP 200; item status is now `progressing`
- **Priority**: P0

## TC-018: MainItem valid transition — pending → closed
- **Source**: US-2 / AC-1; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `pending` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "closed"}`
- **Expected**: HTTP 200; item status is now `closed`
- **Priority**: P0

## TC-019: MainItem invalid transition — pending → reviewing (skip)
- **Source**: US-2 / AC-2; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `pending` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "reviewing"}`
- **Expected**: HTTP 400; item status remains `pending`
- **Priority**: P0

## TC-020: MainItem invalid transition — pending → completed (skip)
- **Source**: US-2 / AC-2; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `pending` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "completed"}`
- **Expected**: HTTP 400; item status remains `pending`
- **Priority**: P0

## TC-021: MainItem valid transitions from progressing
- **Source**: US-2 / AC-3; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `progressing` state
- **Steps**:
  1. Call ChangeStatus with each of: `blocking`, `pausing`, `reviewing`, `closed`
- **Expected**: Each call returns HTTP 200; status updates accordingly
- **Priority**: P0

## TC-022: MainItem invalid transitions from progressing
- **Source**: US-2 / AC-4; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `progressing` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "pending"}`
  2. Call ChangeStatus with `{"status": "completed"}`
- **Expected**: Both return HTTP 400; status remains `progressing`
- **Priority**: P0

## TC-023: MainItem valid transition — blocking → progressing
- **Source**: US-2 / AC-5; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `blocking` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "progressing"}`
- **Expected**: HTTP 200; status is now `progressing`
- **Priority**: P0

## TC-024: MainItem invalid transitions from blocking
- **Source**: US-2 / AC-6; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `blocking` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "pausing"}`
  2. Call ChangeStatus with `{"status": "closed"}`
- **Expected**: Both return HTTP 400; status remains `blocking`
- **Priority**: P0

## TC-025: MainItem valid transitions from pausing
- **Source**: US-2 / AC-7; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `pausing` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "progressing"}`
  2. Call ChangeStatus with `{"status": "closed"}`
- **Expected**: Both return HTTP 200
- **Priority**: P0

## TC-026: MainItem invalid transitions from pausing
- **Source**: US-2 / AC-8; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `pausing` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "blocking"}`, `{"status": "reviewing"}`, `{"status": "completed"}`
- **Expected**: All return HTTP 400; status remains `pausing`
- **Priority**: P0

## TC-027: MainItem terminal states are irreversible
- **Source**: US-2 / AC-9; Spec AC-2
- **Type**: API
- **Pre-conditions**: MainItem exists in `completed` state
- **Steps**:
  1. Call ChangeStatus with any status value
- **Expected**: HTTP 400; status remains `completed`
- **Priority**: P0

## TC-028: Self-transition returns error
- **Source**: US-2 / AC-10; Spec AC-4
- **Type**: API
- **Pre-conditions**: MainItem exists in `progressing` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "progressing"}`
- **Expected**: HTTP 400; status remains `progressing`
- **Priority**: P0

## TC-029: SubItem valid transition — pending → progressing
- **Source**: US-3 / AC-1; Spec AC-3
- **Type**: API
- **Pre-conditions**: SubItem exists in `pending` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "progressing"}`
- **Expected**: HTTP 200; status is now `progressing`
- **Priority**: P0

## TC-030: SubItem invalid transitions from pending
- **Source**: US-3 / AC-2; Spec AC-3
- **Type**: API
- **Pre-conditions**: SubItem exists in `pending` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "blocking"}`, `{"status": "pausing"}`, `{"status": "completed"}`
- **Expected**: All return HTTP 400; status remains `pending`
- **Priority**: P0

## TC-031: SubItem valid transitions from progressing
- **Source**: US-3 / AC-3; Spec AC-3
- **Type**: API
- **Pre-conditions**: SubItem exists in `progressing` state
- **Steps**:
  1. Call ChangeStatus with each of: `blocking`, `pausing`, `completed`, `closed`
- **Expected**: Each returns HTTP 200
- **Priority**: P0

## TC-032: SubItem blocking/pausing → progressing only
- **Source**: US-3 / AC-4; Spec AC-3
- **Type**: API
- **Pre-conditions**: SubItem exists in `blocking` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "progressing"}`
  2. Call ChangeStatus with `{"status": "pausing"}`
- **Expected**: First call returns HTTP 200; second returns HTTP 400
- **Priority**: P0

## TC-033: SubItem terminal states are irreversible
- **Source**: US-3 / AC-6; Spec AC-3
- **Type**: API
- **Pre-conditions**: SubItem exists in `completed` state
- **Steps**:
  1. Call ChangeStatus with any status value
- **Expected**: HTTP 400; status remains `completed`
- **Priority**: P0

## TC-034: Update API ignores status field — MainItem
- **Source**: US-4 / AC-1; Spec AC-5
- **Type**: API
- **Pre-conditions**: MainItem exists in `pending` state
- **Steps**:
  1. Call `PUT /api/v1/teams/:teamId/items/:itemId` with `{"status": "completed", "title": "updated"}`
- **Expected**: HTTP 200; title is updated; status remains `pending`
- **Priority**: P0

## TC-035: Update API ignores status field — any item any state
- **Source**: US-4 / AC-2; Spec AC-5
- **Type**: API
- **Pre-conditions**: Any item in any state
- **Steps**:
  1. Call Update API with any `status` value in the body
- **Expected**: Status field has no effect; only ChangeStatus can alter status
- **Priority**: P0

## TC-036: PM can transition reviewing → completed
- **Source**: US-5 / AC-1
- **Type**: API
- **Pre-conditions**: MainItem in `reviewing` state; caller is a PM
- **Steps**:
  1. Call ChangeStatus with `{"status": "completed"}` as PM
- **Expected**: HTTP 200; status is now `completed`
- **Priority**: P0

## TC-037: Non-PM cannot transition reviewing → completed
- **Source**: US-5 / AC-2
- **Type**: API
- **Pre-conditions**: MainItem in `reviewing` state; caller is NOT a PM
- **Steps**:
  1. Call ChangeStatus with `{"status": "completed"}` as non-PM
- **Expected**: HTTP 403 (permission error); status remains `reviewing`
- **Priority**: P0

## TC-038: Linkage — all SubItems completed/closed with at least one completed → reviewing
- **Source**: US-6 / AC-1; Spec AC-7
- **Type**: API
- **Pre-conditions**: MainItem in `progressing` state with 2 SubItems; one is `completed`, one is `closed`
- **Steps**:
  1. Verify both SubItems are in terminal states with at least one `completed`
  2. Observe MainItem status
- **Expected**: MainItem automatically transitions to `reviewing`
- **Priority**: P0

## TC-039: Linkage — all SubItems closed (none completed) → closed
- **Source**: US-6 / AC-2; Spec AC-8
- **Type**: API
- **Pre-conditions**: MainItem in `progressing` state; all SubItems are `closed`
- **Steps**:
  1. Transition last SubItem to `closed`
  2. Observe MainItem status
- **Expected**: MainItem automatically transitions to `closed`
- **Priority**: P0

## TC-040: Linkage — all SubItems pausing → pausing
- **Source**: US-6 / AC-3; Spec AC-8
- **Type**: API
- **Pre-conditions**: MainItem in `progressing` state; all SubItems are `pausing`
- **Steps**:
  1. Transition last SubItem to `pausing`
  2. Observe MainItem status
- **Expected**: MainItem automatically transitions to `pausing`
- **Priority**: P1

## TC-041: Linkage — any SubItem blocking → MainItem blocking
- **Source**: US-6 / AC-4; Spec AC-8
- **Type**: API
- **Pre-conditions**: MainItem in `progressing` state; one SubItem transitions to `blocking`
- **Steps**:
  1. Call ChangeStatus on a SubItem with `{"status": "blocking"}`
  2. Observe MainItem status
- **Expected**: MainItem automatically transitions to `blocking`
- **Priority**: P1

## TC-042: Linkage — SubItem progressing triggers MainItem pending → progressing
- **Source**: US-6 / AC-5; Spec AC-8
- **Type**: API
- **Pre-conditions**: MainItem in `pending` state; SubItem transitions to `progressing`
- **Steps**:
  1. Call ChangeStatus on a SubItem with `{"status": "progressing"}`
  2. Observe MainItem status
- **Expected**: MainItem automatically transitions to `progressing`
- **Priority**: P1

## TC-043: Adding pending SubItem to reviewing MainItem reverts to progressing
- **Source**: US-8 / AC-1; Spec AC-9
- **Type**: API
- **Pre-conditions**: MainItem in `reviewing` state
- **Steps**:
  1. Add a new SubItem (default status `pending`)
  2. Observe MainItem status
- **Expected**: MainItem automatically reverts to `progressing`
- **Priority**: P0

## TC-044: Deleting SubItem triggers linkage re-evaluation
- **Source**: US-8 / AC-2; Spec AC-10
- **Type**: API
- **Pre-conditions**: MainItem in `progressing` state with 2 SubItems; one is `completed`, one is `progressing`
- **Steps**:
  1. Delete the `progressing` SubItem
  2. Observe MainItem status
- **Expected**: Linkage re-evaluates against remaining SubItems; MainItem transitions to `reviewing` (only completed SubItem remains)
- **Priority**: P1

## TC-045: Terminal side effects — completion=100 and actual_end_date set
- **Source**: US-11 / AC-1; Spec AC-6
- **Type**: API
- **Pre-conditions**: MainItem in `progressing` state with `completion < 100`
- **Steps**:
  1. Call ChangeStatus with `{"status": "closed"}`
  2. Fetch the item
- **Expected**: `completion` is 100; `actual_end_date` is set to current timestamp
- **Priority**: P0

## TC-046: Status history recorded for every successful change
- **Source**: US-12 / AC-1; Spec AC-14
- **Type**: API
- **Pre-conditions**: MainItem in `pending` state
- **Steps**:
  1. Call ChangeStatus with `{"status": "progressing"}`
  2. Query `status_histories` for this item
- **Expected**: A new record exists with `from_status=pending`, `to_status=progressing`, `changed_by` set to caller ID, `is_auto=false`, `created_at` set
- **Priority**: P0

## TC-047: Linkage-triggered history has is_auto=true
- **Source**: US-12 / AC-2; Spec AC-15
- **Type**: API
- **Pre-conditions**: MainItem auto-transitions via linkage
- **Steps**:
  1. Trigger a linkage-based MainItem status change
  2. Query `status_histories` for the MainItem
- **Expected**: The linkage-triggered record has `is_auto=true`
- **Priority**: P0

## TC-048: Manual change history has is_auto=false
- **Source**: US-12 / AC-3; Spec AC-15
- **Type**: API
- **Pre-conditions**: Any item
- **Steps**:
  1. Manually call ChangeStatus
  2. Query `status_histories`
- **Expected**: The record has `is_auto=false`
- **Priority**: P0

## TC-049: available-transitions returns correct options per status
- **Source**: Spec AC-23
- **Type**: API
- **Pre-conditions**: MainItem exists in `blocking` state
- **Steps**:
  1. Call `GET /api/v1/teams/:teamId/items/:itemId/available-transitions`
- **Expected**: Response contains only `["progressing"]`
- **Priority**: P0

## TC-050: RecalcCompletion runs before linkage when SubItem completes
- **Source**: US-10 / AC-1; Spec AC-13
- **Type**: API
- **Pre-conditions**: MainItem with multiple SubItems; all SubItems about to be completed
- **Steps**:
  1. Transition the last SubItem to `completed`
  2. Fetch the MainItem
- **Expected**: `completion` percentage is updated (recalculated) AND status is `reviewing`; both changes visible in the same response
- **Priority**: P1

## TC-051: Status dropdown options refresh after status change
- **Source**: UI cache invalidation
- **Type**: UI
- **Pre-conditions**: MainItem exists in `pending` state with valid transitions
- **Steps**:
  1. Open the status dropdown for the item
  2. Select `progressing` to change status
  3. Wait for the status badge to update to "进行中"
  4. Click the status badge again to open the dropdown
- **Expected**: The dropdown shows transitions valid for `progressing` (e.g. 阻塞中, 已暂停, 待验收, 已关闭), NOT the transitions for `pending`
- **Priority**: P0

## TC-052: Update MainItem returns 422 when status is terminal
- **Source**: Terminal main item edit guard
- **Type**: API
- **Pre-conditions**: MainItem exists in `closed` state
- **Steps**:
  1. Call `PUT /api/v1/teams/:teamId/main-items/:itemId` with `{"title": "updated"}`
- **Expected**: HTTP 422; response code is `TERMINAL_MAIN_ITEM`; item title unchanged
- **Priority**: P0

## TC-053: Create SubItem returns 422 when MainItem status is terminal
- **Source**: Terminal main item sub-item creation guard
- **Type**: API
- **Pre-conditions**: MainItem exists in `closed` state
- **Steps**:
  1. Call `POST /api/v1/teams/:teamId/main-items/:itemId/sub-items` with valid sub-item body
- **Expected**: HTTP 422; response code is `TERMINAL_MAIN_ITEM`; no sub-item created
- **Priority**: P0

## TC-054: ChangeStatus to terminal returns 422 when sub-items are not all terminal
- **Source**: Sub-items not terminal guard
- **Type**: API
- **Pre-conditions**: MainItem in `reviewing` state; has one SubItem in `progressing` state
- **Steps**:
  1. Call `PUT /api/v1/teams/:teamId/main-items/:itemId/status` with `{"status": "completed"}` as PM
- **Expected**: HTTP 422; response code is `SUB_ITEMS_NOT_TERMINAL`; main item status unchanged
- **Priority**: P0

## TC-055: ChangeStatus to terminal succeeds when all sub-items are terminal
- **Source**: Sub-items not terminal guard
- **Type**: API
- **Pre-conditions**: MainItem in `reviewing` state; all SubItems are `completed` or `closed`
- **Steps**:
  1. Call `PUT /api/v1/teams/:teamId/main-items/:itemId/status` with `{"status": "completed"}` as PM
- **Expected**: HTTP 200; main item status is `completed`
- **Priority**: P0

## TC-056: Edit button disabled when sub-item is in terminal state
- **Source**: Terminal sub-item edit guard
- **Type**: UI
- **Pre-conditions**: SubItem exists in `completed` or `closed` state; user has `main_item:update` permission
- **Steps**:
  1. Navigate to the SubItem detail page
  2. Observe the Edit button
- **Expected**: Edit button is disabled (not clickable); clicking it does not open the edit dialog
- **Priority**: P0

## TC-057: Append progress button disabled when sub-item is in terminal state
- **Source**: Terminal sub-item progress guard
- **Type**: UI
- **Pre-conditions**: SubItem exists in `completed` or `closed` state; user has `progress:update` permission
- **Steps**:
  1. Navigate to the SubItem detail page
  2. Observe the "追加进度" button
- **Expected**: Button is disabled; clicking it does not open the append progress dialog
- **Priority**: P0

## TC-058: Achievement dialog appears when switching sub-item to completed
- **Source**: Completed status achievement dialog
- **Type**: UI
- **Pre-conditions**: SubItem is in a non-terminal state with `completed` as a valid transition
- **Steps**:
  1. Open the status dropdown on the SubItem detail page
  2. Select `已完成`
- **Expected**: An achievement dialog appears (not the generic confirm dialog); it contains a textarea for 成果 and no fields for 完成度 or 卡点
- **Priority**: P0

## TC-059: Cancel on achievement dialog aborts completed status change
- **Source**: Completed status achievement dialog
- **Type**: UI
- **Pre-conditions**: Achievement dialog is open for `completed` transition
- **Steps**:
  1. Click "取消" in the achievement dialog
- **Expected**: Status change is not executed; sub-item remains in its current state; no progress record is created
- **Priority**: P0

## TC-060: Confirm on achievement dialog executes status change and appends progress record
- **Source**: Completed status achievement dialog
- **Type**: UI
- **Pre-conditions**: Achievement dialog is open; user enters achievement text
- **Steps**:
  1. Enter achievement text in the textarea
  2. Click "确认完成"
- **Expected**: Sub-item status changes to `已完成`; completion is set to 100%; a progress record is created with the entered achievement text
- **Priority**: P0

## TC-061: SubItem closed transition forces completion=100 and sets actual_end_date
- **Source**: Terminal side effects for closed
- **Type**: API
- **Pre-conditions**: SubItem in `progressing` state with `completion < 100`
- **Steps**:
  1. Call `PUT /api/v1/teams/:teamId/sub-items/:subId/status` with `{"status": "closed"}`
  2. Fetch the sub-item
- **Expected**: `completion` is 100; `actual_end_date` is set to current timestamp
- **Priority**: P0

---

## Traceability

| TC ID | Source | Type | Priority |
|-------|--------|------|----------|
| TC-001 | US-1 / AC-1 | UI | P0 |
| TC-002 | US-1 / AC-1; Spec R1 | UI | P0 |
| TC-003 | US-13 / AC-1; Spec AC-16 | UI | P0 |
| TC-004 | US-14 / AC-1; Spec AC-17 | UI | P0 |
| TC-005 | US-14 / AC-1; US-5 / AC-3 | UI | P0 |
| TC-006 | US-14 / AC-2; US-5 / AC-3 | UI | P0 |
| TC-007 | US-14 / AC-4 | UI | P0 |
| TC-008 | US-15 / AC-1; Spec AC-19 | UI | P1 |
| TC-009 | US-15 / AC-2; Spec AC-19 | UI | P1 |
| TC-010 | US-15 / AC-3; Spec AC-19 | UI | P1 |
| TC-011 | US-16 / AC-1; Spec AC-21 | UI | P0 |
| TC-012 | US-16 / AC-2; Spec AC-21 | UI | P0 |
| TC-013 | US-16 / AC-3; Spec AC-21 | UI | P0 |
| TC-014 | US-9 / AC-1; Spec AC-12 | UI | P1 |
| TC-015 | Spec AC-18 | UI | P0 |
| TC-016 | US-5 / AC-3; Spec AC-20 | UI | P0 |
| TC-017 | US-2 / AC-1; Spec AC-2 | API | P0 |
| TC-018 | US-2 / AC-1; Spec AC-2 | API | P0 |
| TC-019 | US-2 / AC-2; Spec AC-2 | API | P0 |
| TC-020 | US-2 / AC-2; Spec AC-2 | API | P0 |
| TC-021 | US-2 / AC-3; Spec AC-2 | API | P0 |
| TC-022 | US-2 / AC-4; Spec AC-2 | API | P0 |
| TC-023 | US-2 / AC-5; Spec AC-2 | API | P0 |
| TC-024 | US-2 / AC-6; Spec AC-2 | API | P0 |
| TC-025 | US-2 / AC-7; Spec AC-2 | API | P0 |
| TC-026 | US-2 / AC-8; Spec AC-2 | API | P0 |
| TC-027 | US-2 / AC-9; Spec AC-2 | API | P0 |
| TC-028 | US-2 / AC-10; Spec AC-4 | API | P0 |
| TC-029 | US-3 / AC-1; Spec AC-3 | API | P0 |
| TC-030 | US-3 / AC-2; Spec AC-3 | API | P0 |
| TC-031 | US-3 / AC-3; Spec AC-3 | API | P0 |
| TC-032 | US-3 / AC-4; Spec AC-3 | API | P0 |
| TC-033 | US-3 / AC-6; Spec AC-3 | API | P0 |
| TC-034 | US-4 / AC-1; Spec AC-5 | API | P0 |
| TC-035 | US-4 / AC-2; Spec AC-5 | API | P0 |
| TC-036 | US-5 / AC-1 | API | P0 |
| TC-037 | US-5 / AC-2 | API | P0 |
| TC-038 | US-6 / AC-1; Spec AC-7 | API | P0 |
| TC-039 | US-6 / AC-2; Spec AC-8 | API | P0 |
| TC-040 | US-6 / AC-3; Spec AC-8 | API | P1 |
| TC-041 | US-6 / AC-4; Spec AC-8 | API | P1 |
| TC-042 | US-6 / AC-5; Spec AC-8 | API | P1 |
| TC-043 | US-8 / AC-1; Spec AC-9 | API | P0 |
| TC-044 | US-8 / AC-2; Spec AC-10 | API | P1 |
| TC-045 | US-11 / AC-1; Spec AC-6 | API | P0 |
| TC-046 | US-12 / AC-1; Spec AC-14 | API | P0 |
| TC-047 | US-12 / AC-2; Spec AC-15 | API | P0 |
| TC-048 | US-12 / AC-3; Spec AC-15 | API | P0 |
| TC-049 | Spec AC-23 | API | P0 |
| TC-050 | US-10 / AC-1; Spec AC-13 | API | P1 |
| TC-051 | UI cache invalidation | UI | P0 |
| TC-052 | Terminal main item edit guard | API | P0 |
| TC-053 | Terminal main item sub-item creation guard | API | P0 |
| TC-054 | Sub-items not terminal guard | API | P0 |
| TC-055 | Sub-items not terminal guard | API | P0 |
| TC-056 | Terminal sub-item edit guard | UI | P0 |
| TC-057 | Terminal sub-item progress guard | UI | P0 |
| TC-058 | Completed status achievement dialog | UI | P0 |
| TC-059 | Completed status achievement dialog | UI | P0 |
| TC-060 | Completed status achievement dialog | UI | P0 |
| TC-061 | Terminal side effects for closed sub-item | API | P0 |

# User Stories: Status Flow Optimization

## User Roles

| Role | Description |
|------|-------------|
| PM | Product Manager — manages MainItems, reviews and verifies completion |
| Executor | Team member — manages SubItems, marks own work as complete |

---

## US-1: Consistent Status Display

**As a** PM,
**I want** status values to display consistently in Chinese across all views,
**So that** I can quickly understand item states without confusion caused by mismatched labels.

### Acceptance Criteria

**Given** a MainItem or SubItem with any status,
**When** I view it in any page (list, detail, weekly view),
**Then** the status label always shows the correct Chinese name (e.g., `progressing` → "进行中") with no fallback/default styling.

---

## US-2: Valid MainItem Status Transitions

**As a** PM,
**I want** to transition a MainItem's status only through defined valid paths,
**So that** workflow integrity is maintained and invalid state changes are prevented.

### Acceptance Criteria

**Given** a MainItem in `pending` state,
**When** I change it to `progressing` or `closed`,
**Then** the transition succeeds and the new status is saved.

**Given** a MainItem in `pending` state,
**When** I attempt to change it to `reviewing` or `completed` (skipping steps),
**Then** the API returns an error and the status remains unchanged.

**Given** a MainItem in `progressing` state,
**When** I change it to `blocking`, `pausing`, `reviewing`, or `closed`,
**Then** the transition succeeds.

**Given** a MainItem in `progressing` state,
**When** I attempt to change it to `pending` or `completed` (skipping reviewing),
**Then** the API returns an error and the status remains unchanged.

**Given** a MainItem in `blocking` state,
**When** I change it to `progressing`,
**Then** the transition succeeds.

**Given** a MainItem in `blocking` state,
**When** I attempt to change it to any state other than `progressing`,
**Then** the API returns an error and the status remains unchanged.

**Given** a MainItem in `pausing` state,
**When** I change it to `progressing` or `closed`,
**Then** the transition succeeds.

**Given** a MainItem in `pausing` state,
**When** I attempt to change it to `blocking`, `reviewing`, or `completed`,
**Then** the API returns an error and the status remains unchanged.

**Given** a MainItem in `completed` or `closed` state,
**When** I attempt any status change,
**Then** the API returns an error — terminal states are irreversible.

**Given** a MainItem in any state,
**When** I attempt to change it to its current status (self-transition),
**Then** the API returns an error and the status remains unchanged.

---

## US-3: Valid SubItem Status Transitions

**As an** Executor,
**I want** to transition a SubItem's status only through defined valid paths,
**So that** my work progress is accurately tracked without invalid state jumps.

### Acceptance Criteria

**Given** a SubItem in `pending` state,
**When** I change it to `progressing` or `closed`,
**Then** the transition succeeds.

**Given** a SubItem in `pending` state,
**When** I attempt to change it to `blocking`, `pausing`, or `completed`,
**Then** the API returns an error and the status remains unchanged.

**Given** a SubItem in `progressing` state,
**When** I change it to `blocking`, `pausing`, `completed`, or `closed`,
**Then** the transition succeeds.

**Given** a SubItem in `blocking` or `pausing` state,
**When** I change it to `progressing`,
**Then** the transition succeeds.

**Given** a SubItem in `blocking` or `pausing` state,
**When** I attempt to change it to any state other than `progressing`,
**Then** the API returns an error and the status remains unchanged.

**Given** a SubItem in `completed` or `closed` state,
**When** I attempt any status change,
**Then** the API returns an error — terminal states are irreversible.

**Given** a SubItem in any state,
**When** I attempt to change it to its current status (self-transition),
**Then** the API returns an error and the status remains unchanged.

---

## US-4: Status Cannot Be Changed via Update API

**As a** system integrator,
**I want** the item Update endpoint to ignore any `status` field in the request body,
**So that** all status changes are forced through the state machine and cannot be bypassed.

### Acceptance Criteria

**Given** a MainItem in `pending` state,
**When** I call the Update API with `status: "completed"` in the request body,
**Then** the API returns success (other fields may update), but the item's status remains `pending` — the status field is silently ignored.

**Given** any item in any state,
**When** I call the Update API with any `status` value,
**Then** the status field has no effect; only ChangeStatus can alter status.

---

## US-5: PM-Only Verification

**As a** PM,
**I want** to be the only one who can mark a MainItem as `completed` from `reviewing`,
**So that** quality control is enforced and executors cannot self-approve their own work.

### Acceptance Criteria

**Given** a MainItem in `reviewing` state and I am a PM,
**When** I change it to `completed` or back to `progressing`,
**Then** the transition succeeds.

**Given** a MainItem in `reviewing` state and I am a non-PM user,
**When** I attempt to change it to `completed` or `progressing` via the API,
**Then** the API returns a permission error and the status remains `reviewing`.

**Given** a MainItem in `reviewing` state and I am a non-PM user,
**When** I open the status dropdown in the UI,
**Then** `completed` and `progressing` options are not visible — the dropdown shows no available transitions.

---

## US-6: Automatic MainItem Linkage

**As a** PM,
**I want** the MainItem status to automatically update when SubItems reach certain states,
**So that** I don't need to manually track and update the parent item's status.

### Acceptance Criteria

**Given** a MainItem with multiple SubItems,
**When** all SubItems are in `completed` or `closed` state with at least one `completed`,
**Then** the MainItem automatically transitions to `reviewing`.

**Given** a MainItem with multiple SubItems,
**When** all SubItems are `closed` (none `completed`),
**Then** the MainItem automatically transitions to `closed`.

**Given** a MainItem with multiple SubItems,
**When** all SubItems are `pausing`, or a mix of `pausing` and `closed` (no other states),
**Then** the MainItem automatically transitions to `pausing`.

**Given** a MainItem currently in `pending` or `progressing` state with multiple SubItems,
**When** any SubItem is `blocking` and not all SubItems are in terminal states,
**Then** the MainItem automatically transitions to `blocking`.

**Given** a MainItem currently in `pending` state with multiple SubItems,
**When** any SubItem becomes `progressing`,
**Then** the MainItem automatically transitions to `progressing`.

---

## US-7: Manual Reviewing for MainItem Without SubItems

**As a** PM,
**I want** to manually move a MainItem to `reviewing` when it has no SubItems,
**So that** I can still use the verification workflow for items I manage directly without breaking them into sub-tasks.

### Acceptance Criteria

**Given** a MainItem in `progressing` state with no SubItems,
**When** I (as PM) change it to `reviewing`,
**Then** the transition succeeds.

**Given** a MainItem in `progressing` state with at least one SubItem,
**When** I attempt to manually call ChangeStatus to `reviewing`,
**Then** the API returns an error — when SubItems exist, `reviewing` is only reachable via automatic linkage, not manual transition.

---

## US-8: Linkage on SubItem Add/Remove

**As a** PM,
**I want** the MainItem status to re-evaluate when I add or remove a SubItem,
**So that** the parent status always reflects the current set of child items.

### Acceptance Criteria

**Given** a MainItem in `reviewing` state,
**When** I add a new `pending` SubItem,
**Then** the MainItem automatically reverts to `progressing`.

**Given** a MainItem with SubItems,
**When** I delete a SubItem,
**Then** linkage is re-evaluated against the remaining SubItems and the MainItem status updates accordingly.

**Given** a MainItem with no SubItems,
**When** any SubItem operation occurs,
**Then** no linkage evaluation is triggered and the MainItem status is unchanged.

---

## US-9: Linkage Failure Notification

**As a** PM,
**I want** to be notified when automatic linkage cannot apply,
**So that** I can manually intervene and keep the status accurate.

### Acceptance Criteria

**Given** a SubItem status change triggers linkage evaluation,
**When** the MainItem's current state does not allow the computed target transition,
**Then** the MainItem status remains unchanged, the failed attempt is recorded in `status_histories` with the reason, and a toast warning is shown in the UI.

---

## US-10: RecalcCompletion and Linkage Coordination

**As a** PM,
**I want** the MainItem's completion percentage and status to both update correctly when a SubItem is completed,
**So that** I see an accurate and consistent picture of progress in a single action.

### Acceptance Criteria

**Given** a MainItem with SubItems where all SubItems are about to be completed,
**When** the last SubItem transitions to `completed`,
**Then** the MainItem's `completion` percentage is recalculated first, and then linkage evaluation runs — resulting in both an updated completion% and a status of `reviewing`.

**Given** a MainItem with SubItems,
**When** a SubItem transitions to `completed` but other SubItems are still active,
**Then** the MainItem's `completion` percentage is recalculated and the MainItem status is updated by linkage if a rule matches; both changes are visible in the same response.

---

## US-11: Terminal State Side Effects

**As a** PM,
**I want** completion percentage and actual end date to be automatically set when an item reaches a terminal state,
**So that** I don't need to manually update these fields after closing or completing an item.

### Acceptance Criteria

**Given** a MainItem or SubItem in any non-terminal state,
**When** it transitions to `completed` or `closed`,
**Then** `completion` is set to 100 and `actual_end_date` is set to the current timestamp.

---

## US-12: Status Change History

**As a** PM,
**I want** every status change to be recorded with the actor and timestamp,
**So that** I can audit the history of an item and understand how it progressed.

### Acceptance Criteria

**Given** any successful status change (manual or automatic),
**When** the change is applied,
**Then** a new record is created in `status_histories` with `from_status`, `to_status`, `changed_by`, `is_auto`, and `created_at`.

**Given** a linkage-triggered status change,
**When** the change is applied,
**Then** `is_auto` is `true` in the history record.

**Given** a manually triggered status change,
**When** the change is applied,
**Then** `is_auto` is `false` in the history record.

**Given** a linkage evaluation that fails (target transition not allowed),
**When** the failure occurs,
**Then** a record is still created in `status_histories` with the intended `to_status` and a `remark` explaining the failure reason.

---

## US-13: Working Status Dropdown

**As a** PM or Executor,
**I want** the status dropdown in the item detail page to actually change the status when I select a new value,
**So that** I can manage status directly from the detail view without workarounds.

### Acceptance Criteria

**Given** I am on an item detail page,
**When** I select a new status from the dropdown,
**Then** the ChangeStatus API is called and the displayed status updates to reflect the change.

---

## US-14: Filtered Status Options

**As a** PM or Executor,
**I want** the status dropdown to only show valid next states for the current status and my role,
**So that** I cannot accidentally attempt an invalid or unauthorized transition.

### Acceptance Criteria

**Given** a MainItem in `reviewing` state and I am a PM,
**When** I open the status dropdown,
**Then** only `completed` and `progressing` are shown.

**Given** a MainItem in `reviewing` state and I am a non-PM user,
**When** I open the status dropdown,
**Then** no options are shown (dropdown is empty or disabled).

**Given** a MainItem in `blocking` state,
**When** I open the status dropdown,
**Then** only `progressing` is shown.

**Given** a MainItem in `completed` or `closed` state,
**When** I open the status dropdown,
**Then** no options are shown and the dropdown is disabled.

---

## US-15: Overdue Indicator

**As a** PM,
**I want** to see an overdue indicator on items past their expected end date,
**So that** I can quickly identify delayed work at a glance.

### Acceptance Criteria

**Given** an item with `expected_end_date` in the past and a non-terminal status,
**When** I view the item in any list or detail view,
**Then** an overdue badge/indicator is displayed alongside the status.

**Given** an item with `expected_end_date` in the past but status is `completed` or `closed`,
**When** I view the item,
**Then** no overdue indicator is shown.

**Given** an item with `expected_end_date` today or in the future,
**When** I view the item,
**Then** no overdue indicator is shown regardless of status.

---

## US-16: Irreversible Action Confirmation

**As a** PM or Executor,
**I want** a confirmation dialog before marking an item as `completed` or `closed`,
**So that** I don't accidentally make irreversible changes.

### Acceptance Criteria

**Given** I select `completed` or `closed` from the status dropdown,
**When** the selection is made,
**Then** a confirmation dialog appears explaining the action is irreversible before any API call is made.

**Given** the confirmation dialog is shown,
**When** I click "Cancel",
**Then** the status change is not executed and the item remains in its current state.

**Given** the confirmation dialog is shown,
**When** I click "Confirm",
**Then** the ChangeStatus API is called and the status changes.

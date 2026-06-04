---
journey: "team-and-progress-visibility"
step: 2
step-action: "View weekly progress page with mixed activity"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/team-and-progress-visibility/journey.md
---

# Contract: team-and-progress-visibility / Step 2: View weekly progress page with mixed activity

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "Non-terminal main items exist; a terminal main item has activity (status change or sub-item update) this week"
- Input: "User opens the weekly progress page"
- Output: "All non-terminal main items are displayed; terminal main items with activity this week or last week are displayed; terminal main items with no activity in either week are hidden"
- State: "No state change; display only"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user was previously authenticated and the session has since expired"
- Input: "User attempts to view the weekly progress page after session expiry"
- Output: "The user is redirected to the login page; after re-authenticating, the weekly progress page loads with default filters"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Non-terminal main items are always displayed on the weekly progress page regardless of activity
- Terminal main items are hidden only when they have zero activity in both the current week and the previous week
- The weekly progress page uses natural week boundaries (Monday through Sunday) for activity calculation

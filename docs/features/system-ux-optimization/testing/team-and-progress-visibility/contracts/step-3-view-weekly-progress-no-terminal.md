---
journey: "team-and-progress-visibility"
step: 3
step-action: "View weekly progress page with no active terminal items"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/team-and-progress-visibility/journey.md
---

# Contract: team-and-progress-visibility / Step 3: View weekly progress page with no active terminal items

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "All terminal main items have had no activity in the current or previous week"
- Input: "User opens the weekly progress page"
- Output: "Only non-terminal main items are displayed; terminal items are hidden"
- State: "No state change; display only"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired"
- Input: "User opens the weekly progress page after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Non-terminal main items are always displayed on the weekly progress page regardless of activity
- Terminal main items are hidden only when they have zero activity in both the current week and the previous week

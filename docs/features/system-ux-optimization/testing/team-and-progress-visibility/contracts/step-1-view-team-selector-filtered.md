---
journey: "team-and-progress-visibility"
step: 1
step-action: "View team selector with permission filtering"
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/team-and-progress-visibility/journey.md
---

# Contract: team-and-progress-visibility / Step 1: View team selector with permission filtering

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Outcome "success"
- Preconditions: "User has permission to access Team A and Team B only"
- Input: "User opens the team selector after login"
- Output: "Only teams the user has permission to access are displayed (Team A and Team B); Team C is not shown"
- State: "No state change; team selector displays filtered list"
- Side-effect: "none"

## Outcome "session-expired"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired"
- Input: "User attempts to open the team selector after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Outcome "validation-error"
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface validation-error mandatory outcome; invalid team identifier in URL -->
- Preconditions: "The team selector URL is manipulated to contain an invalid team identifier"
- Input: "User navigates to a URL containing an invalid team identifier"
- Output: "An error message is displayed indicating the team identifier is invalid; no data for the invalid team is shown"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Team selector always filters to only teams the current user has permission to access
- Non-terminal main items are always displayed on the weekly progress page regardless of activity

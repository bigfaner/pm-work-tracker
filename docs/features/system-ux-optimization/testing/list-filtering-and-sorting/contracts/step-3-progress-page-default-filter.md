---
journey: "list-filtering-and-sorting"
step: 3
step-action: "Progress page default filter"
surface_types: ["web"]
generated: "2026-06-04"
sources:
  - docs/features/system-ux-optimization/testing/list-filtering-and-sorting/journey.md
  - docs/features/system-ux-optimization/prd/prd-spec.md (#12)
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 9)
---

# Contract: list-filtering-and-sorting / Step 3: Progress page default filter

<!-- gen-contracts: do not edit manually. Regenerate via /gen-contracts. -->

## Fact Table

| Claim | Source | Type |
|---|---|---|
| Progress page defaults to in-progress filter | prd-spec #12, Story 9 AC2 | explicit |
| Terminal items hidden by default filter | prd-spec #12 | inferred |

## Outcome "success"
<!-- surface: web -->
- Preconditions: "The progress page has no stored filter preferences for the current session"
- Input: "PM user opens the overall progress page for the first time"
- Output: "The in-progress status checkbox is selected by default; only in-progress main items are displayed; terminal status items are filtered out"
- State: "No state change; default filter is applied for display"
- Side-effect: "none"

## Outcome "session-expired"
<!-- surface: web -->
<!-- source: inferred -->
<!-- reasoning: Derived from Web surface session-expired mandatory outcome -->
- Preconditions: "The user's session has expired before opening the progress page"
- Input: "PM user attempts to open the progress page after session expiry"
- Output: "The user is redirected to the login page"
- State: "No state change"
- Side-effect: "none"

## Journey Invariants

- Terminal status main items are hidden only when filtered out, not by default on the main item list page
- Empty filter results always show a message with a clear filter action

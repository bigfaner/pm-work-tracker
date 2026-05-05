---
scope: feature
source: feature/improve-ui BIZ-006
---

# Weekly View Conventions

## Progress Delta Markers

Weekly view displays progress deltas with specific visual markers:

| Marker | Color | Meaning |
|--------|-------|---------|
| `+N%` | Green (success) | Progress increased by N% this week |
| `已完成` | Green (success) | Item completed this week |
| `NEW` | Amber (warning) | Item added this week |

## Collapsed Items

Completed items with no change this week are collapsed by default. Users can expand to view details.

## Comparison Structure

Each main item card shows three groups:
- **lastWeek**: previous week's snapshot
- **thisWeek**: current week snapshot with delta markers
- **completedNoChange**: completed items with no change (collapsed)

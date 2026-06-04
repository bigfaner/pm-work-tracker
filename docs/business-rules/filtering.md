---
title: "Filtering Rules"
domains: [filter, penetration, assignee, status, multi-select, match]
---

# Filtering Rules

Business rules for data filtering across views, including filter penetration and multi-select patterns.

Source: feature/system-ux-optimization

## Filter Penetration

### BIZ-filter-001: Assignee Filter Penetration

When filtering by assignee, results include both:
- **Direct matches**: Main items directly assigned to the filtered assignee (`matchType=direct`)
- **Indirect matches**: Main items whose sub-items are assigned to the filtered assignee (`matchType=indirect`)

When status filter is also applied, an AND logic applies: main item status must match AND (assignee directly matches OR sub-item assignee matches).

**No filter applied**: All items returned, no `matchType` field present.

**Source**: feature/system-ux-optimization BIZ-010

---
scope: global
source: feature/code-conventions BIZ-003
---

# Code Deduplication Principle

When a code pattern appears 3+ times across files, extract it into a shared helper function or component. Each distinct helper/component should exist as exactly one copy.

## Threshold

- **3 copies** → mandatory extraction
- **1-2 copies** → acceptable, monitor for growth

## Examples from This Project

| Pattern | Before | After |
|---------|--------|-------|
| mapNotFound | 5 per-domain functions | 1 generic `MapNotFound` |
| Pagination defaults | 6 inline blocks | 1 `ApplyPaginationDefaults` |
| Date parsing | 11 inline `time.Parse` | 1 `ParseDate` |
| Textarea styling | 14 copy-pasted styles | 1 `Textarea` component |
| PrioritySelect options | 21 inline option groups | 1 `PrioritySelectItems` component |

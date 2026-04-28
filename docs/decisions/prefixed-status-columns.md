# Decision: Prefixed Status Columns

**Date**: 2026-04-28
**Category**: database
**Status**: Active

## Context

MySQL reserves the word `status` as a keyword. Using `status` as a column name requires backtick quoting in raw SQL and can cause subtle issues with different MySQL versions or ORM configurations.

## Decision

All status-like columns use a prefixed name following the pattern `<entity-prefix>_status`. The Go struct field uses PascalCase (e.g., `ItemStatus`) and JSON tag uses camelCase (e.g., `itemStatus`). GORM auto-maps camelCase JSON tags to snake_case columns.

## Applied To

| Model | Column | Struct Field | JSON Tag |
|-------|--------|-------------|----------|
| MainItem | `item_status` | `ItemStatus` | `itemStatus` |
| SubItem | `item_status` | `ItemStatus` | `itemStatus` |
| ItemPool | `pool_status` | `PoolStatus` | `poolStatus` |
| DecisionLog | `log_status` | `LogStatus` | `logStatus` |

## Rule

When adding a new model with a status-like field, prefix it with a short entity identifier. Never use bare `status` as a column name.

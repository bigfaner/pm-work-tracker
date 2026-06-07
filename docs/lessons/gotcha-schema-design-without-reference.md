---
created: "2026-06-07"
tags: [data-model, architecture]
---

# Schema designed without referencing existing conventions

## Problem

Milestone-map `schema.sql` was written from scratch using ad-hoc naming and indexing choices. After review, it violated at least 6 established conventions:

1. Unique key naming: `uk_milestone_maps_biz_key` instead of `uk_biz_key`
2. No business unique constraints with `(deleted_flag, deleted_time)` — soft-deleted records could violate uniqueness
3. Standalone single-column indexes instead of composite indexes
4. `DATE` type for date columns instead of `DATETIME` (existing tables use DATETIME)
5. Explicit `DEFAULT NULL` on nullable columns (existing tables omit it)
6. Missing business-level unique constraints (map_name per team, milestone_name per map)

Required a full rewrite of schema.sql plus cascading updates to er-diagram.md and tech-design.md.

## Root Cause

**L1 — Schema was designed in isolation.** The author wrote the DDL without opening the existing `backend/migrations/MySql-schema.sql` or `SQLite-schema.sql` files to check naming patterns, index conventions, or soft-delete unique constraint patterns.

**L2 — No "convention check" step in the design workflow.** The design phase went directly from PRD to schema authoring without an explicit step to audit existing conventions. The project has `docs/conventions/` but no rule requiring a schema-author to read existing migration files first.

**L3 — The schema.sql is a design-time artifact (in `docs/features/`), not a runtime migration file.** This separation creates a gap: developers know to check `backend/migrations/` for runtime consistency, but the design-time schema in feature docs is treated as a fresh sketch rather than an extension of existing conventions.

## Solution

1. Read `backend/migrations/MySql-schema.sql` and `SQLite-schema.sql` before writing any new table DDL
2. Applied the full convention checklist:
   - `uk_biz_key` for biz_key UK (MySQL), `uk_<table>_biz_key` (SQLite)
   - `uk_<desc>_deleted (..., deleted_flag, deleted_time)` for business-unique constraints
   - `DATETIME` for date fields, no explicit `DEFAULT NULL`
   - Composite indexes over standalone where possible
   - `idx_<table>_<desc>` naming for non-unique indexes

## Reusable Pattern

**Before writing any new table DDL, read the existing schema files first.** Specifically:

1. Open `backend/migrations/MySql-schema.sql` and `SQLite-schema.sql`
2. Check: UK naming (`uk_biz_key` vs `uk_<table>_biz_key`), business UK pattern with `(deleted_flag, deleted_time)`, date column types, index naming
3. Apply the same patterns to new tables
4. Verify the design-time schema in feature docs matches the runtime conventions exactly

This applies to any feature that adds or modifies database tables.

## Related Files

- `docs/features/milestone-map/design/schema.sql`
- `backend/migrations/MySql-schema.sql`
- `backend/migrations/SQLite-schema.sql`

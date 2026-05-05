---
feature: "improve-ui"
generated: "2026-05-04"
status: draft
---

# Business Rules: Improve UI

## Authorization

### BIZ-001: Page-Level Access Control

**Rule**: Access control follows a strict role hierarchy: SuperAdmin pages (user management) → PM pages (team detail, member management) → Team Member pages (read-only views). Non-authorized users see the navigation entry hidden, not a 403 page.

**Context**: The original design mixed SuperAdmin and PM views in one page, causing confusion about what actions are available to which role.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section 4.1

| Role | Can Access |
|------|-----------|
| SuperAdmin | User management, all teams, team detail |
| PM | Own team detail, member management |
| Team Member | Read-only views |

## User Management

### BIZ-002: Auto-Generated Initial Password

**Rule**: When creating a user, the backend generates a random 12-character password (mixed case + digits). The password is returned in the API response via `initialPassword` field exactly once. It is never stored in retrievable form — only the bcrypt hash is persisted.

**Context**: Avoids transmitting passwords out-of-band. The one-time display ensures the admin who created the user can relay the initial credentials.

**Scope**: [CROSS]

**Source**: design/api-handbook.md "创建用户"

### BIZ-003: User Status Lifecycle

**Rule**: Users have `enabled`/`disabled` status. Disabled users cannot login (return `USER_DISABLED` error). SuperAdmin cannot disable themselves (`CANNOT_DISABLE_SELF`).

**Context**: Disabling is preferred over deletion for audit trails. Self-disable prevention avoids accidental lockout.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section 5.1, design/api-handbook.md

## Data Validation

### BIZ-004: Item Validation Rules

**Rule**: For all items (main items and sub items): title is required. For items with date ranges: end date must be ≥ start date. Progress percentage must be ≥ last recorded value.

**Context**: Prevents data integrity issues from incomplete or inconsistent input.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section 5.3

### BIZ-005: Weekly View Time Constraint

**Rule**: Weekly progress view does not accept future week dates. Requesting a future week returns `FUTURE_WEEK_NOT_ALLOWED` error.

**Context**: Progress data only exists for past and current weeks. Allowing future weeks would show empty or misleading data.

**Scope**: [CROSS]

**Source**: design/tech-design.md Error Handling

## Display Conventions

### BIZ-006: Progress Delta Markers

**Rule**: Weekly view displays progress deltas with specific markers: `+N%` (green) for progress increase, `已完成` (green) for items completed this week, `NEW` (amber) for items added this week. Completed items with no change this week are collapsed by default.

**Context**: Helps PMs quickly identify what changed between weeks without scanning all items.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section 5.5

### BIZ-007: Empty State Convention

**Rule**: All list/dashboard pages must display an empty state message with guidance (e.g., "暂无事项，引导创建") when no data matches the current filter. Never show a blank page.

**Context**: Blank pages confuse users about whether the app is broken or simply empty.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Sections 4.2, 4.3

### BIZ-008: Item Sorting Convention

**Rule**: Items are sorted by priority ascending (P1 → P2 → P3). Items with the same priority are sorted by deadline ascending (earliest first).

**Context**: Ensures highest-priority and most urgent items appear at the top across all views.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Sections 5.3, 5.4

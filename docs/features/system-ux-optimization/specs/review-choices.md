---
feature: "system-ux-optimization"
reviewed: "2026-06-05"
---

# Review Choices

Auto-integrated in non-interactive mode. All CROSS items approved for integration.

## Approved for Integration

### Business Rules -> docs/business-rules/

- BIZ-001 -> docs/business-rules/item-lifecycle.md (terminal status definition, overlaps with status-machine.md terminal column)
- BIZ-002 -> docs/business-rules/item-lifecycle.md (terminal sorting)
- BIZ-003 -> docs/conventions/soft-delete.md (cascade soft-delete, extends SD-003)
- BIZ-004 -> docs/conventions/soft-delete.md (deletion audit entry, new SD-006)
- BIZ-005 -> docs/conventions/permission-codes.md (new permission codes)
- BIZ-006 -> docs/business-rules/item-lifecycle.md (move preserves status/assignee)
- BIZ-007 -> docs/business-rules/item-lifecycle.md (move target validation)
- BIZ-008 -> docs/conventions/item-codes.md (atomic counter for move, extends BIZ-code-004)
- BIZ-009 -> docs/conventions/data-model.md (completion recalculation on move, extends DM-008)
- BIZ-010 -> docs/business-rules/filtering.md (filter penetration logic)
- BIZ-011 -> docs/conventions/security.md (team selector membership filter)
- BIZ-012 -> docs/conventions/authorization.md (member nil RoleKey fallback)
- BIZ-014 -> docs/conventions/frontend-ux.md (filter empty state with reset)

### Technical Specs -> docs/conventions/

- TECH-001 -> docs/conventions/api-boundary.md (delete main item endpoint)
- TECH-002 -> docs/conventions/api-boundary.md (delete sub item endpoint)
- TECH-003 -> docs/conventions/api-boundary.md (move sub item endpoint)
- TECH-004 -> docs/conventions/error-codes.md (new error codes)
- TECH-005 -> docs/conventions/permission-codes.md (new permission codes)
- TECH-006 -> docs/conventions/data-validation.md (multi-select status filter)
- TECH-008 -> docs/conventions/data-model.md (completion recalculation on structural changes)
- TECH-010 -> docs/conventions/security.md (permission-guarded UI buttons)
- TECH-013 -> docs/conventions/performance-targets.md (filter penetration response time)

## Skipped

(none -- all CROSS items auto-approved)

## Related Existing Entries

- conventions/status-machine.md: Terminal states already defined in status table -> BIZ-001 adds cross-view sorting/filtering rules (keep both)
- conventions/soft-delete.md SD-003: SoftDelete method pattern -> BIZ-003/004 extend with cascade and audit patterns (append)
- conventions/item-codes.md BIZ-code-004: Atomic counter for creation -> BIZ-008 extends to move operation (append)
- conventions/data-model.md DM-008: Completion calculation -> BIZ-009/TECH-008 extend with move trigger (append)
- conventions/error-codes.md: Error code registry -> TECH-004 adds new codes (append to table)
- conventions/permission-codes.md: Permission code registry -> BIZ-005/TECH-005 add new codes (append to table)
- conventions/frontend-ux.md: Empty state convention -> BIZ-014 extends with filter-specific behavior (append)
- conventions/performance-targets.md: Performance targets -> TECH-013 adds filter penetration target (append)

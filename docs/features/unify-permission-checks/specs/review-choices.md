---
feature: "unify-permission-checks"
reviewed: "2026-05-09"
status: pending-review
---

# Review Choices

> Auto-generated proposal. User review required before integration.
> Run `/consolidate-specs` interactively to confirm or adjust choices.

## Proposed for Integration

- BIZ-001 → docs/conventions/permission-codes.md (replace "SuperAdmin Bypass Rule" section)
- BIZ-002 → docs/conventions/permission-codes.md (add new section)
- TECH-001 → docs/conventions/permission-codes.md (add to existing file)
- TECH-002 → docs/conventions/permission-codes.md (add to existing file)
- TECH-004 → docs/conventions/authorization.md (append "Service Layer Authorization" section)
- TECH-006 → docs/conventions/permission-codes.md (add to existing file)
- TECH-007 → docs/conventions/security.md (append "IsSuperAdmin Field Visibility" section)
- TECH-009 → docs/conventions/security.md (update "Authorization" section)

## Skipped (LOCAL — stays in feature)

- BIZ-003: Custom roles must not be blocked by bypass logic
- BIZ-004: IsSuperAdmin removed from API responses
- BIZ-005: Seed data: SuperAdmin gets all 29 codes
- BIZ-006: IsSuperAdmin column retained for loading only
- TECH-003: GetUserPermissions SuperAdmin path
- TECH-005: ProgressService skipRegressionCheck param
- TECH-008: Authorization error path migration

## Related Existing Entries (Overlaps)

- docs/conventions/permission-codes.md "SuperAdmin Bypass Rule" → **replace** with BIZ-001 + BIZ-002 (PRD Section 5.7 item 2 explicitly calls for this update)
- docs/conventions/security.md "Authorization" section → **update** with TECH-009 (middleware chain description should mention RequirePermission, not RequireRole)

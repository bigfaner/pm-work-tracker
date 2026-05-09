---
feature: "unify-permission-checks"
generated: "2026-05-09"
status: draft
---

# Business Rules: Unify Permission Checks

## Authorization — Unified Permission Code Enforcement

### BIZ-001: Permission Codes as Single Authorization Path

**Rule**: All authorization checks (middleware, handler, service, frontend) must go through the same permCodes path. No bypass mechanisms (SuperAdmin short-circuit, handler-level `isPMOrSuperAdmin`, PM BizKey substitution) are permitted.
**Context**: Previously, SuperAdmin bypassed `RequirePermission` via a tier-1 short-circuit, and handlers had separate `isPMOrSuperAdmin()` checks. Custom roles with valid permission codes were blocked by these bypass layers.
**Scope**: [CROSS]
**Source**: prd/prd-spec.md Background, Goals, Scope

This rule applies to all future features that add authorization checks. Any new endpoint must use `RequirePermission` middleware exclusively -- no handler-level or service-level identity checks for authorization.

### BIZ-002: SuperAdmin Loads All Permission Codes

**Rule**: When a user has `IsSuperAdmin=true`, the system loads all 29 permission codes (via `permissions.AllCodeStrings()`). SuperAdmin does NOT bypass permission checks; instead, the loaded codes include every defined code, so all `RequirePermission` checks pass.
**Context**: The `is_super_admin` DB column is retained as a flag to determine which permission codes to load -- not as a bypass mechanism. This ensures SuperAdmin is subject to the same authorization flow as all other users.
**Scope**: [CROSS]
**Source**: prd/prd-spec.md Scope "Intentionally Kept", design/tech-design.md Overview

This affects `TeamScopeMiddleware` (injects all 29 codes for SuperAdmin) and `GetUserPermissions` (returns all 29 codes for all teams).

### BIZ-003: Custom Roles Must Not Be Blocked by Bypass Logic

**Rule**: A custom role possessing a permission code (e.g., `sub_item:update`) must be able to perform the corresponding action on any resource within scope, without being blocked by assignee ownership checks or PM identity checks.
**Context**: This is the core fix. Previously, custom roles with `sub_item:update` were blocked by handler-level assignee checks (`isPMOrSuperAdmin`). The fix removes these checks entirely, relying solely on middleware permission code enforcement.
**Scope**: [LOCAL]
**Source**: prd/prd-spec.md Goals, design/tech-design.md Error Path Migration

### BIZ-004: IsSuperAdmin Removed from API Responses

**Rule**: The `isSuperAdmin` boolean field must not appear in any API response DTO. Frontend authorization uses only permission codes via `hasPermission()`.
**Context**: Exposing SuperAdmin status to the frontend created a parallel authorization path. Removing it forces all frontend authorization through the permission code system.
**Scope**: [LOCAL]
**Source**: prd/prd-spec.md Section 5.5, design/api-handbook.md Breaking Changes

### BIZ-005: Seed Data — SuperAdmin Role Gets All 29 Permission Codes

**Rule**: The superadmin preset role is seeded with all 29 permission codes at startup via `seedPresetRoles`. The `seedRole` function is additive and idempotent.
**Context**: Previously, the superadmin role had zero permission codes (relying on bypass). After unification, it needs all codes for the permCodes path to work.
**Scope**: [LOCAL]
**Source**: prd/prd-spec.md Section 5.1, design/tech-design.md Interface 6

## Data Ownership

### BIZ-006: IsSuperAdmin Column Retained for PermCodes Loading Only

**Rule**: The `pmw_users.is_super_admin` column and `User.IsSuperAdmin` model field are retained. Their sole purpose is to signal `TeamScopeMiddleware` and `GetUserPermissions` to load all 29 permission codes. They must not be used for any bypass or direct authorization check.
**Context**: Column kept to avoid schema migration and bootstrapping problems. Its semantics changed from "bypass all checks" to "load all permission codes".
**Scope**: [LOCAL]
**Source**: prd/prd-spec.md Scope "Intentionally Kept", design/tech-design.md Scope Alignment

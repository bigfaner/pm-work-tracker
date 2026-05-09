---
feature: "unify-permission-checks"
generated: "2026-05-09"
status: draft
---

# Technical Specifications: Unify Permission Checks

## Middleware Architecture

### TECH-001: RequirePermission Two-Tier Structure (No SuperAdmin Bypass)

**Requirement**: `RequirePermission` middleware has exactly 2 tiers: (1) permCodes context check, (2) non-team DB query fallback. No SuperAdmin short-circuit tier is permitted.
**Scope**: [CROSS]
**Source**: design/tech-design.md Interface 1

All users, including SuperAdmin, pass through the permCodes check. SuperAdmin passes because `TeamScopeMiddleware` injects all 29 codes. This pattern must be followed for any future middleware additions.

### TECH-002: TeamScopeMiddleware SuperAdmin Code Injection

**Requirement**: When `IsSuperAdmin` context is true, `TeamScopeMiddleware` must inject all 29 permission codes via `permissions.AllCodeStrings()` into the `permCodes` context key. SuperAdmin still skips team membership check.
**Scope**: [CROSS]
**Source**: design/tech-design.md Interface 2

```go
if IsSuperAdmin(c) {
    c.Set("callerTeamRole", "superadmin")
    c.Set("permCodes", permissions.AllCodeStrings())
    c.Next()
    return
}
```

### TECH-003: GetUserPermissions SuperAdmin Path

**Requirement**: `GetUserPermissions` for SuperAdmin users returns all 29 codes for all teams. The response struct (`UserPermissions`) must not contain an `IsSuperAdmin` field.
**Scope**: [LOCAL]
**Source**: design/tech-design.md Interface 3

Requires `teamRepo.ListTeamBizKeys(ctx)` to enumerate all teams. When no teams exist, returns an empty `TeamPermissions` map (correct behavior for fresh systems).

## Service Layer Simplification

### TECH-004: Service Methods Must Not Perform Authorization Checks

**Requirement**: Service-layer methods (team_service, etc.) must not perform PM identity checks or role-based authorization. Authorization is enforced exclusively by middleware. Service method signatures must not include `pmBizKey` parameters used for identity verification.
**Scope**: [CROSS]
**Source**: design/tech-design.md Interface 4

This applies to `UpdateTeam`, `RemoveMember`, `TransferPM`, `DisbandTeam`, `UpdateMemberRole`, `ListTeams`. Authorization for these operations is enforced by `RequirePermission` middleware with the appropriate code (`team:update`, `team:delete`, `team:remove`, `team:transfer`, `team:update_role`).

### TECH-005: ProgressService Append — skipRegressionCheck Parameter

**Requirement**: The `Append` method parameter `isPM bool` is renamed to `skipRegressionCheck bool`. The value is sourced from the `sub_item:assign` permCode in the handler, not from an identity check.
**Scope**: [LOCAL]
**Source**: design/tech-design.md Interface 5

Users with `sub_item:assign` permission skip the regression check (completion decreasing validation). All others get the regression check enforced.

## Seed Data

### TECH-006: seedPresetRoles — SuperAdmin Role Seeding Pattern

**Requirement**: The superadmin preset role is seeded with `permissions.AllCodeStrings()` (all 29 codes). The `seedRole` function is additive (INSERT-IGNORE semantics), so re-running is idempotent.
**Scope**: [CROSS]
**Source**: design/tech-design.md Interface 6

```go
seedRole(tx, "superadmin", "系统超级管理员", true, permissions.AllCodeStrings())
```

This pattern applies to any future preset role that needs full permissions.

## Cross-Layer Data Flow

### TECH-007: IsSuperAdmin Field Visibility Rules

**Requirement**: The `IsSuperAdmin` field is visible only within backend internal code (model, middleware context). It must not appear in: API response JSON, VO/DTO structs, frontend TypeScript types, or mock data.
**Scope**: [CROSS]
**Source**: design/tech-design.md Cross-Layer Data Map

| Layer | Visibility |
|-------|-----------|
| DB column `is_super_admin` | KEPT |
| Go model `User.IsSuperAdmin` | KEPT |
| AuthMiddleware context `isSuperAdmin` | KEPT |
| TeamScopeMiddleware (reads context) | KEPT |
| GetUserPermissions (reads model) | KEPT |
| JSON response DTOs | REMOVED |
| Frontend TypeScript types | REMOVED |
| Frontend auth store | REMOVED |

## Error Handling

### TECH-008: Authorization Error Path Migration

**Requirement**: When handler-level or service-level authorization checks are removed, the error path must shift to middleware-level `RequirePermission`. The HTTP status code (403) and error code (`FORBIDDEN`) remain the same, but the enforcement layer changes.
**Scope**: [LOCAL]
**Source**: design/tech-design.md Error Path Migration

No new error codes. Existing `ErrForbidden` (403) and `ErrNotTeamMember` (403) are sufficient.

## Security

### TECH-009: No Handler-Level SuperAdmin Special Paths

**Requirement**: Handler code must not contain any SuperAdmin-specific branches (no `if IsSuperAdmin`, no `isPMOrSuperAdmin()`, no PM BizKey substitution). All authorization is enforced by middleware before the handler executes.
**Scope**: [CROSS]
**Source**: design/tech-design.md Security Considerations, prd/prd-spec.md Security Requirements

This is a security invariant: any future handler must rely exclusively on middleware `RequirePermission` for authorization.

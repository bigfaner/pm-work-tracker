---
created: 2026-05-08
related: design/tech-design.md
---

# API Handbook: 统一权限码校验，移除 bypass 模式

## API Overview

No new endpoints. This feature modifies response shapes, middleware authorization behavior, and handler-level logic. SuperAdmin now passes through permission codes (all 29) instead of middleware bypass.

## Breaking Changes

### Response Field Removal: `isSuperAdmin`

The `isSuperAdmin` boolean field is removed from all response DTOs. The backend retains the field internally but no longer exposes it.

#### Affected Endpoints

| Endpoint | DTO | Field Removed |
|----------|-----|---------------|
| `POST /v1/auth/login` | AdminUserDTO | `isSuperAdmin` |
| `GET /v1/admin/users/:userId` | AdminUserDTO | `isSuperAdmin` |
| `GET /v1/admin/users` | AdminUserDTO[] | `isSuperAdmin` |
| `POST /v1/admin/users` | AdminUserDTO | `isSuperAdmin` |
| `PUT /v1/admin/users/:userId` | AdminUserDTO | `isSuperAdmin` |
| `GET /v1/me/permissions` | UserPermissions | `isSuperAdmin` |

#### Before

```json
{
  "code": 0,
  "data": {
    "isSuperAdmin": true,
    "permissions": { "1": [] }
  }
}
```

#### After (SuperAdmin)

```json
{
  "code": 0,
  "data": {
    "permissions": {
      "1": ["team:create", "team:read", "...all 29 codes..."]
    }
  }
}
```

### Authorization Behavior Change: Unified Permission Codes

The `RequirePermission` middleware no longer has a SuperAdmin short-circuit. SuperAdmin users receive all 29 permission codes from `TeamScopeMiddleware`, passing the same permCodes check as all other users.

The `isPMOrSuperAdmin()` handler-level check is removed from:
- `PUT /v1/teams/:teamId/sub-items/:subId` (sub_item Update)
- `PUT /v1/teams/:teamId/sub-items/:subId/status` (sub_item ChangeStatus)
- `POST /v1/teams/:teamId/sub-items/:subId/progress` (progress Append)

**Before**: PM/SuperAdmin could edit any sub-item; other members only their assigned ones.
**After**: Any user with `sub_item:update` can edit any sub-item. Assignee ownership check removed.

The PM BizKey substitution in team_handler is removed for:
- `PUT /v1/teams/:teamId` (Update)
- `DELETE /v1/teams/:teamId` (Disband)
- `DELETE /v1/teams/:teamId/members/:userId` (RemoveMember)
- `PUT /v1/teams/:teamId/members/:userId/role` (UpdateMemberRole)
- `PUT /v1/teams/:teamId/pm` (TransferPM)

**Before**: SuperAdmin handler substituted PM's BizKey to pass service-layer PM identity check.
**After**: Service layer no longer checks PM identity; relies on middleware `RequirePermission`.

## Non-Breaking Changes

### `GET /v1/me/permissions` — SuperAdmin Response

For SuperAdmin users, the response now includes all 29 permission codes for every team. Previously returned `isSuperAdmin: true` with empty permission arrays.

This enables the frontend to use `hasPermission()` uniformly without a SuperAdmin bypass.

### `GET /v1/teams` — ListTeams Simplified

The `isSuperAdmin` parameter is removed from the service interface. The parameter was already ignored (using `_` blank identifier). No behavior change.

## Error Codes

No new error codes.

| Code | HTTP Status | Description |
|------|-------------|-------------|
| (existing) | 403 | Permission denied — now also applies to handler-level checks (removed assignee bypass) |

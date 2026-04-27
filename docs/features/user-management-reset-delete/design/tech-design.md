---
created: 2026-04-27
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: 用户管理增强 — 重置密码 & 删除用户

## Overview

Add two admin operations to the existing user management module:
1. **Reset Password** — super admin sets a new password for any user via a dialog with client-side validation.
2. **Soft Delete User** — super admin marks a user as deleted; user disappears from list, login is blocked, existing JWT is rejected on next request.

Plus one frontend-only enhancement: **Copy credentials button** on the create-user result dialog.

All changes follow existing patterns: handler → service → repo, interface-driven repositories, DTO/VO layers, Radix Dialog components, React Query mutations.

### Key Facts

- **No new database columns** — `BaseModel` already has `DeletedFlag` (int, 0=active) and `DeletedTime`.
- **No new dependencies** — reuses bcrypt, GORM scopes, Radix Dialog, Clipboard API.
- **JWT invalidation is passive** — auth middleware already loads user from DB on every request; a `DeletedFlag` check is added there.

## Architecture

### Layer Placement

```
Backend:
  Router (router.go) → Handler (admin_handler.go) → Service (admin_service.go) → Repository (user_repo.go)
  Auth Middleware (auth.go) — adds DeletedFlag check
  Auth Service (auth_service.go) — adds DeletedFlag check on login

Frontend:
  UserManagementPage.tsx — adds reset-password dialog, delete confirmation dialog, copy button
  api/admin.ts — adds resetPasswordApi, deleteUserApi
  types/index.ts — adds request/response types
```

### Component Diagram

```
                          ┌──────────────┐
                          │   Router      │
                          │ (router.go)   │
                          └──────┬───────┘
                                 │ registers routes
                    ┌────────────┼────────────┐
                    ▼            ▼            ▼
           PUT .../password  DELETE .../:id   (existing routes)
                    │            │
                    ▼            ▼
              ┌──────────────────────────┐
              │    AdminHandler          │
              │  (admin_handler.go)      │
              │  + ResetPassword()       │
              │  + DeleteUser()          │
              └──────────┬───────────────┘
                         │ calls
                         ▼
              ┌──────────────────────────┐
              │    AdminService          │
              │  (admin_service.go)      │
              │  + ResetPassword()       │
              │  + SoftDeleteUser()      │
              └──────────┬───────────────┘
                         │ calls
                         ▼
              ┌──────────────────────────┐
              │    UserRepo              │
              │  + SoftDelete()          │
              │  (ListFiltered, FindByBizKey gain NotDeleted scope)  │
              └──────────────────────────┘

  ┌─────────────────────────────────────────────┐
  │  Cross-cutting: Auth Middleware + Auth Service │
  │  Both add DeletedFlag != 0 check              │
  └─────────────────────────────────────────────┘
```

### Dependencies

| Dependency | Purpose | Already in project? |
|------------|---------|---------------------|
| `golang.org/x/crypto/bcrypt` | Password hashing | Yes |
| `Radix UI Dialog` | Modal dialogs | Yes |
| Clipboard API (browser) | Copy text to clipboard | Native browser API |
| React Query | Data fetching / mutations | Yes |

## Interfaces

### Backend: AdminService ( additions )

```go
// ResetPassword sets a new password for the target user.
// newPassword is already validated by handler (length, strength).
// Returns a lean DTO with identifying fields for confirmation display.
ResetPassword(ctx context.Context, targetBizKey int64, newPassword string) (*dto.ResetPasswordResp, error)

// SoftDeleteUser marks the target user as deleted.
// callerID is the authenticated admin's internal ID (for self-delete prevention).
// Sets DeletedFlag=1, DeletedTime=now. Does NOT remove team memberships.
SoftDeleteUser(ctx context.Context, callerID uint, targetBizKey int64) error
```

### Backend: UserRepo ( additions )

```go
// SoftDelete sets DeletedFlag=1 and DeletedTime=now for the given user.
SoftDelete(ctx context.Context, user *model.User) error
```

### Backend: Handler methods

```go
// ResetPassword handles PUT /api/v1/admin/users/:userId/password
func (h *AdminHandler) ResetPassword(c *gin.Context)

// DeleteUser handles DELETE /api/v1/admin/users/:userId
func (h *AdminHandler) DeleteUser(c *gin.Context)
```

### Frontend: API functions

```ts
// api/admin.ts additions
function resetPasswordApi(userId: string, req: ResetPasswordReq): Promise<ResetPasswordResp>
function deleteUserApi(userId: string): Promise<void>
```

## Data Models

### Backend: ResetPasswordReq DTO

```go
// In dto/auth.go
type ResetPasswordReq struct {
    NewPassword string `json:"newPassword" binding:"required,min=8,max=64"`
}
```

### Backend: ResetPasswordResp DTO

```go
// In dto/auth.go
type ResetPasswordResp struct {
    BizKey      string `json:"bizKey"`
    Username    string `json:"username"`
    DisplayName string `json:"displayName"`
}
```

No new database model — reuses existing `User` and `BaseModel.DeletedFlag`/`DeletedTime`.

### Frontend: TypeScript types

```ts
// In types/index.ts additions
interface ResetPasswordReq {
  newPassword: string;
}

interface ResetPasswordResp {
  bizKey: string;
  username: string;
  displayName: string;
}
```

> **Note**: The PRD requires "两次输入不一致" validation (Story 2). `confirmPassword` is a **frontend-only field** — it is validated client-side and never sent to the backend. The `ResetPasswordReq` DTO intentionally omits it; only `newPassword` reaches the API. The frontend form holds both fields locally and checks equality before calling the API.

### Frontend Dialog States

#### Reset Password Dialog

| State | UI Display | Trigger |
|-------|------------|---------|
| `idle` | Two empty inputs (newPassword, confirmPassword), "确认" button enabled | Dialog opens |
| `validationError` | Red error text below the failing field; dialog stays open | "确认" clicked, client-side validation fails (empty, <8 chars, no letters+digits, or mismatch) |
| `submitting` | "确认" button shows loading spinner; both inputs disabled | API request in flight |
| `success` | Dialog closes; Toast "密码已重置" | API returns 200 |
| `error` | Error message displayed inside dialog; inputs remain editable; dialog stays open | API returns non-200 (network error, 400, 404, etc.) |

#### Delete Confirmation Dialog

| State | UI Display | Trigger |
|-------|------------|---------|
| `confirming` | Text "确认删除用户 {username}？此操作不可通过界面撤销。" + "确认删除" / "取消" buttons enabled | Dialog opens |
| `submitting` | "确认删除" button loading spinner; both buttons disabled | API request in flight |
| `success` | Dialog closes; row removed from list; Toast "用户已删除" | API returns 200 |
| `error` | Error message displayed inside dialog; buttons re-enabled | API returns non-200 (404 already-deleted, etc.) |

## Error Handling

### Error Types & Codes

| Error Code | Name | Description | HTTP Status |
|------------|------|-------------|-------------|
| `CANNOT_DELETE_SELF` | ErrCannotDeleteSelf | Admin attempted to delete their own account | 422 |
| `USER_DELETED` | ErrUserDeleted | Login attempted by a soft-deleted user | 403 |
| `USER_NOT_FOUND` | ErrUserNotFound | Target user not found (or already deleted) | 404 |
| `VALIDATION_ERROR` | ErrValidation | Password fails strength rules (handler binding) | 400 |

### Propagation Strategy

Follows existing pattern:
- **Handler**: Gin `binding` tags validate request shape → `ErrValidation`. Calls service → maps service errors via `apperrors.RespondError`.
- **Service**: Checks business rules (self-delete, user existence) → returns sentinel `AppError`. Calls repo → maps repo errors via `pkgerrors.MapNotFound`.
- **Repo**: Returns `ErrNotFound` for missing records. GORM errors are wrapped.

## Cross-Layer Data Map

| Field Name | Storage Layer | Backend Model | API/DTO | Frontend Type | Validation Rule |
|------------|---------------|---------------|---------|---------------|-----------------|
| newPassword | — (transient) | — | `json:"newPassword"` binding:"required,min=8,max=64" | `newPassword: string` | required, 8–64 chars, letters+digits |
| userId (path param) | `biz_key` BIGINT | `BaseModel.BizKey` int64 | `:userId` path param | `userId: string` | parsed via `ParseBizKeyParam` |
| DeletedFlag | `deleted_flag` INT DEFAULT 0 | `BaseModel.DeletedFlag` int | — (json:"-") | — | internal only |
| DeletedTime | `deleted_time` DATETIME | `BaseModel.DeletedTime` time.Time | — (json:"-") | — | internal only |

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| Service (Go) | Unit | testing + testify | ResetPassword biz logic, SoftDeleteUser self-delete guard, user-not-found | 85% |
| Handler (Go) | Unit | httptest + Gin test context | Request binding, response codes, error mapping | 85% |
| Frontend API | Unit | vitest + vi.mock | API call patterns, request payloads | 80% |
| Frontend Page | Unit | @testing-library/react + vitest | Dialog open/close, form validation, mutation calls | 80% |

### Key Test Scenarios

**Backend service tests:**
- ResetPassword: happy path (password updated, bcrypt hash changed)
- ResetPassword: user not found → ErrUserNotFound
- SoftDeleteUser: happy path (DeletedFlag set to 1)
- SoftDeleteUser: self-delete → ErrCannotDeleteSelf
- SoftDeleteUser: user not found → ErrUserNotFound
- Login: deleted user rejected → ErrUserDeleted
- Auth middleware: deleted user's JWT rejected → 401

**Backend handler tests:**
- PUT /admin/users/:userId/password — 200, 400 (bad input), 404, 422
- DELETE /admin/users/:userId — 200, 404, 422 (self-delete)

**Frontend tests:**
- Reset password dialog: opens, validates, submits, closes on success
- Delete confirmation dialog: opens, submits, removes row on success
- Copy button: calls clipboard API, shows "已复制" feedback

### Overall Coverage Target

80%

## Security Considerations

### Threat Model

| Threat | Risk | Mitigation |
|--------|------|------------|
| Non-admin calls reset/delete API | High | Permission middleware (`user:update` code), super_admin always passes, others need explicit permission |
| Admin deletes own account | Medium | Service-level check: `callerID == user.ID` → ErrCannotDeleteSelf |
| Password transmitted in plaintext | Medium | HTTPS only; password never stored, immediately hashed with bcrypt |
| Deleted user reuses JWT | Medium | Auth middleware checks DeletedFlag on every request (already loads user from DB) |
| CSRF on delete | Low | Bearer token auth (not cookie-based), SPA architecture |

### Mitigations

- **Permission gating**: Both new routes use `deps.perm("user:update")` middleware, same as existing ToggleUserStatus.
- **Password strength**: Enforced at two levels — Gin binding tags (`min=8,max=64`) and frontend validation (letters+digits regex).
- **No password logging**: Password field is transient, never persisted in logs or response DTOs.
- **Soft delete only**: No hard delete. Historical data (progress, team membership) is preserved.

### Deferred: Rate Limiting on Password Reset

Rate limiting on `PUT /admin/users/:userId/password` is explicitly deferred. Rationale:

- Admin routes are protected by `deps.perm("user:update")` middleware plus super_admin role check — only 1-2 users typically hold this role.
- The existing auth rate limiter (applied to `/api/v1/auth/*` endpoints) is not reused here because admin operations are already tightly gated by permission, not by anonymous traffic.
- If the admin population grows, a per-admin rate limit (e.g., 10 resets/minute) can be added using the same `gin-contrib/limiter` pattern used on auth routes. This is a config-only change, no code redesign needed.

### Deferred: Audit Logging for Sensitive Operations

Audit logging for password reset and user deletion events is explicitly deferred. Rationale:

- The PRD marks audit logging as Out of Scope — no audit log infrastructure exists in the current system.
- Soft delete already records `DeletedTime` on the user record, providing a minimal trace of *when* a deletion occurred and *who* the deleted user was.
- When audit infrastructure is introduced project-wide (e.g., an `audit_log` table recording actor, action, target, timestamp), both `ResetPassword` and `SoftDeleteUser` should emit audit entries as adopters.

## PRD Coverage Map

| PRD Requirement / AC | Design Component | Interface / Model |
|----------------------|------------------|-------------------|
| Story 1: Reset password happy path | `AdminService.ResetPassword` + `AdminHandler.ResetPassword` | `ResetPasswordReq` DTO, `PUT /admin/users/:userId/password` |
| Story 1: Backend error / network timeout | Handler returns error, frontend keeps dialog open | Error response mapping in frontend mutation |
| Story 2: Password validation | Frontend form validation + Gin binding tags | `binding:"required,min=8,max=64"` + letters+digits regex |
| Story 2: Confirm-password mismatch | Frontend-only `confirmPassword` field, equality check before API call | Frontend form state (see Frontend Dialog States) |
| Story 3: Soft delete happy path | `AdminService.SoftDeleteUser` + `AdminHandler.DeleteUser` | `DELETE /admin/users/:userId` |
| Story 3: Stale state (user already deleted) | Service returns ErrUserNotFound, frontend shows message | `FindByBizKey` with `NotDeleted` scope |
| Story 4: Cannot delete self | Service checks `callerID == user.ID` | `ErrCannotDeleteSelf` |
| Story 5: Non-admin cannot see buttons | Frontend checks `isSuperAdmin` from auth store | Conditional rendering in UserManagementPage |
| Story 5: Direct API call returns 403 | Permission middleware `deps.perm("user:update")` | Router configuration |
| Story 6: Copy credentials | Frontend Clipboard API in create-user result dialog | `navigator.clipboard.writeText()` |
| 5.1: List filters deleted users | `ListFiltered` adds `NotDeleted` scope | GORM scope in `user_repo.go` |
| 5.5: Login rejects deleted users | `AuthService.Login` checks `DeletedFlag` | `ErrUserDeleted` |
| 5.5: JWT rejected for deleted users | `AuthMiddleware` checks `DeletedFlag` | 401 response |

## Open Questions

None — all decisions resolved during design.

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| Active JWT blacklist (token revocation) | Immediate invalidation | Requires Redis/store, complex token management | Passive check sufficient — middleware already loads user per request, zero extra latency |
| Auto-generate password on reset (like create) | Simpler frontend (no validation dialog) | Admin can't choose a known password to communicate verbally | PRD specifies admin manually enters password |
| Hard delete user | Clean data | Loses historical progress records, violates PRD requirement | PRD explicitly requires data preservation |
| New permission codes (`user:reset_password`, `user:delete`) | Fine-grained control | Over-engineering for single-role feature | `user:update` is sufficient; super_admin is the only operator |

### References
- `backend/internal/model/base.go` — BaseModel with DeletedFlag/DeletedTime
- `backend/internal/service/admin_service.go` — existing admin service patterns
- `backend/internal/repository/gorm/scopes.go` — NotDeleted scope
- `frontend/src/pages/UserManagementPage.tsx` — existing user management page

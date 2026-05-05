---
scope: backend
source: feature/improve-ui BIZ-002, BIZ-003
---

# User Management Rules

## BIZ-user-001: Auto-Generated Initial Password

When creating a user, the backend generates a random 12-character password (mixed case + digits). The password is returned in the API response via `initialPassword` field **exactly once**. It is never stored in retrievable form — only the bcrypt hash is persisted.

Frontend displays the password in a Dialog with a warning: "请妥善保管，关闭后无法再次查看". User must confirm before the dialog closes.

## BIZ-user-002: User Status Lifecycle

Users have `enabled`/`disabled` status:

- **Disabled users** cannot login — return `USER_DISABLED` error at login endpoint
- **SuperAdmin cannot disable themselves** — return `CANNOT_DISABLE_SELF` error
- **Disabling is preferred over deletion** for audit trails
- Status values: `"enabled"` or `"disabled"` only

## BIZ-user-003: Password Reset Rules

Super admin can reset any user's password via `PUT /admin/users/:userId/password`. The new password must be 8-64 characters and contain both letters and digits. Validation is enforced at two levels:

1. **Backend**: Gin binding tags `binding:"required,min=8,max=64"` plus service-level strength check (letters + digits regex).
2. **Frontend**: Client-side validation before API call (same rules: min 8 chars, must contain at least one letter and one digit).

The `confirmPassword` field is frontend-only and never sent to the backend.

**Why**: Two-level validation provides immediate UX feedback and defense-in-depth. The 8-64 range balances usability (not too short) with practical limits (bcrypt has a 72-byte limit).

**Source**: feature/user-management-reset-delete

## BIZ-user-004: Soft Delete for Users

Soft delete sets `DeletedFlag=1` and `DeletedTime=now` on the user record. The user is excluded from all list queries via a `NotDeleted` GORM scope. Soft-deleted users are blocked at two enforcement points:

1. **Login endpoint**: Returns `USER_DELETED` error (403) if `DeletedFlag != 0`.
2. **Auth middleware**: Checks `DeletedFlag` on every authenticated request. If deleted, returns 401 and the client-side JWT is effectively invalidated.

No database rows are physically removed. Historical data (progress entries, team memberships) is preserved.

**Why**: Hard delete would orphan historical data and break referential integrity. Passive JWT invalidation (checking DB on every request) avoids the complexity of token blacklists.

**Source**: feature/user-management-reset-delete

## BIZ-user-005: Self-Deletion Prevention

A user cannot delete their own account. The service layer checks `callerID == targetUser.ID` and returns `CANNOT_DELETE_SELF` (422) if they match. The frontend disables the delete button for the current user's own row.

**Why**: Prevents the admin from locking themselves out of the system. There is no self-service account recovery mechanism.

**Source**: feature/user-management-reset-delete

## BIZ-user-006: Auth Middleware Rejects Deleted Users on Every Request

The auth middleware (`AuthMiddleware`) loads the user from the database on every request. After loading, it checks `DeletedFlag`. If the flag is non-zero, the middleware aborts with 401, regardless of whether the JWT itself is valid.

**Why**: JWT tokens remain valid until expiry. A deleted user's token would otherwise continue to work until it naturally expires. The per-request DB check is the enforcement mechanism — no token blacklist needed.

**Source**: feature/user-management-reset-delete

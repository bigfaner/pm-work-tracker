---
created: 2026-04-27
related: design/tech-design.md
---

# API Handbook: 用户管理增强 — 重置密码 & 删除用户

## API Overview

Two new admin endpoints for user password reset and soft deletion, plus changes to existing auth behavior.

## Endpoints

### Reset Password

**Method**: `PUT`
**Path**: `/api/v1/admin/users/:userId/password`
**Auth**: Bearer token + `user:update` permission

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| newPassword | string | Yes | New password (8–64 chars, must contain letters and digits) |

```json
{
  "newPassword": "NewPass123"
}
```

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| bizKey | string | Target user's business key |
| username | string | Target user's username |
| displayName | string | Target user's display name |

```json
{
  "code": 0,
  "data": {
    "bizKey": "1234567890",
    "username": "alice",
    "displayName": "Alice Wang"
  }
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | VALIDATION_ERROR | newPassword missing, too short (<8), or too long (>64) |
| 404 | USER_NOT_FOUND | Target user not found or already deleted |
| 403 | FORBIDDEN | Caller lacks `user:update` permission |
| 401 | UNAUTHORIZED | Missing or invalid JWT |

---

### Delete User (Soft Delete)

**Method**: `DELETE`
**Path**: `/api/v1/admin/users/:userId`
**Auth**: Bearer token + `user:update` permission

#### Request

No request body. The `:userId` path parameter is the user's bizKey.

#### Response (200)

```json
{
  "code": 0,
  "data": null
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | USER_NOT_FOUND | Target user not found or already deleted |
| 422 | CANNOT_DELETE_SELF | Admin attempted to delete their own account |
| 403 | FORBIDDEN | Caller lacks `user:update` permission |
| 401 | UNAUTHORIZED | Missing or invalid JWT |

---

## Behavior Changes (Existing Endpoints)

### Login — Deleted User Rejection

**Affected Endpoint**: `POST /api/v1/auth/login`

When a soft-deleted user attempts to login, the response changes:

| Status | Code | Description |
|--------|------|-------------|
| 403 | USER_DELETED | Account has been deleted |

### JWT Validation — Deleted User Rejection

**Affected Middleware**: `AuthMiddleware` (applied to all authenticated routes)

When a deleted user's JWT is presented on any authenticated request:

| Status | Code | Description |
|--------|------|-------------|
| 401 | UNAUTHORIZED | User account no longer exists |

### List Users — Filter Deleted

**Affected Endpoint**: `GET /api/v1/admin/users`

Soft-deleted users are excluded from results. No API contract change — simply fewer rows returned.

## Data Contracts

> Both structs live in `backend/internal/dto/auth.go`.

### ResetPasswordReq

```go
type ResetPasswordReq struct {
    NewPassword string `json:"newPassword" binding:"required,min=8,max=64"`
}
```

### ResetPasswordResp

```go
type ResetPasswordResp struct {
    BizKey      string `json:"bizKey"`
    Username    string `json:"username"`
    DisplayName string `json:"displayName"`
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| CANNOT_DELETE_SELF | 422 | Super admin attempted to delete their own account |
| USER_DELETED | 403 | Login attempted by a soft-deleted user |
| USER_NOT_FOUND | 404 | Target user not found or already deleted |
| VALIDATION_ERROR | 400 | Request body failed binding validation |
| FORBIDDEN | 403 | Caller lacks required permission |
| UNAUTHORIZED | 401 | Missing, invalid, or revoked JWT |

---
scope: backend
source: feature/improve-ui TECH-003
---

# Security Conventions

## Password Handling

- Hashing: bcrypt (via `golang.org/x/crypto/bcrypt`)
- API response: `PasswordHash` field has `json:"-"` tag — never serialized
- Initial password: server-generated, 12 chars (mixed case + digits), returned once
- Storage: only bcrypt hash persisted, never plaintext or reversible encryption

## Authentication

- JWT (HS256, 24h expiry)
- Login rate limiting: 10 req/min/IP

## Authorization

- Middleware chain: Auth → TeamScope → RequireRole
- Team data isolation via `teamID` filtering in queries
- SuperAdmin self-protect: cannot disable own account

## Team Data Access

- Team list endpoint (`GET /v1/teams`) returns only teams where the requesting user is a member (via `pmw_team_members` join). SuperAdmin bypasses this filter and sees all teams.
- Destructive action buttons (delete, move) wrapped in `<PermissionGuard code="...">` on frontend as defensive UI. Backend `RequirePermission` middleware enforces actual authorization.

**Source**: feature/system-ux-optimization BIZ-011, TECH-010

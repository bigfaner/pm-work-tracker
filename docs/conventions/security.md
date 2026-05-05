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

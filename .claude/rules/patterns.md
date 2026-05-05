---
paths:
  - "backend/**/*.go"
---

# Backend Go Patterns

Detailed conventions in `docs/conventions/backend-helpers.md`, `docs/conventions/api-boundary.md`, `docs/conventions/repo-helpers.md`, `docs/conventions/error-codes.md`.

Key rules enforced by linters:

- Handler constructors: panic-on-nil validation, no method-level nil checks
- Error mapping: use `pkgerrors.MapNotFound(err, apperrors.ErrXxx)`
- Pagination: `dto.ApplyPaginationDefaults` in handlers only, repos receive offset/limit
- Repository: interface in `repository/`, GORM impl in `repository/gorm/`
- VO conversion: in handlers, not services

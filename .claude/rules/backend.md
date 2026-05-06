---
paths:
  - "backend/**/*.go"
---

# Backend Critical Rules

Detailed conventions: `docs/conventions/backend-helpers.md`, `docs/conventions/api-boundary.md`.

- Error mapping: `pkgerrors.MapNotFound(err, apperrors.ErrXxx)` — never compare `gorm.ErrRecordNotFound` directly
- Pagination: `dto.ApplyPaginationDefaults` in handlers only, repos receive offset/limit
- BizKey boundary: service layer uses `int64 bizKey`, repo layer uses `uint id` — never `uint(bizKey)` cast
- Soft-delete: all queries must include `.Where("deleted_flag = 0")` via `NotDeleted()` scope
- Handler constructors: panic-on-nil validation, no method-level nil checks

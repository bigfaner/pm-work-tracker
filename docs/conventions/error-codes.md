---
scope: global
source: feature/improve-ui TECH-001, TECH-002
verified: "2026-05-04"
---

# Error Code Registry

Source of truth: `backend/internal/pkg/errors/errors.go`

## Pattern

All API error codes follow `UPPER_SNAKE_CASE`, mapped to specific HTTP status codes. Error propagation: `Service → AppError → Handler (RespondError) → JSON {code, msg}`.

## Registered Error Codes

| Code | HTTP | Description |
|------|------|-------------|
| UNAUTHORIZED | 401 | Authentication required |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Generic resource not found |
| TEAM_NOT_FOUND | 404 | Team not found |
| ITEM_NOT_FOUND | 404 | Item not found |
| USER_NOT_FOUND | 404 | User not found |
| DECISION_LOG_NOT_FOUND | 404 | Decision log not found |
| VALIDATION_ERROR | 400 | Request field validation failed |
| TEAM_CODE_DUPLICATE | 400 | Team code already taken |
| INTERNAL_ERROR | 500 | Internal server error |
| ALREADY_EXISTS | 409 | Resource already exists |
| DUPLICATE_BIZ_KEY | 409 | biz_key uniqueness violation |
| NOT_TEAM_MEMBER | 403 | Not a member of this team |
| USER_DISABLED | 403 | Account disabled (login) |
| USER_DELETED | 403 | User account has been deleted |
| INVALID_STATUS | 422 | Invalid status transition |
| INVALID_FIELD | 422 | Invalid field name for update |
| ARCHIVE_NOT_ALLOWED | 422 | Only completed/closed items can be archived |
| PROGRESS_REGRESSION | 422 | Completion cannot be lower than previous record |
| ALREADY_MEMBER | 422 | User is already a team member |
| CANNOT_REMOVE_SELF | 422 | PM cannot remove themselves |
| CANNOT_MODIFY_SELF | 422 | Cannot modify your own permissions |
| ITEM_ALREADY_PROCESSED | 422 | Item already processed |
| NO_DATA | 422 | No data to export |
| USER_EXISTS | 422 | Username already exists |
| CANNOT_DISABLE_SELF | 422 | Cannot disable your own account |
| CANNOT_DELETE_SELF | 422 | Cannot delete your own account |
| FUTURE_WEEK_NOT_ALLOWED | 422 | Cannot create progress for future weeks |
| TERMINAL_MAIN_ITEM | 422 | Cannot modify completed/closed main item |
| SUB_ITEMS_NOT_TERMINAL | 422 | All sub-items must be completed/closed first |
| CANNOT_ASSIGN_PM_ROLE | 422 | Use transfer PM to change team PM |

## Adding New Error Codes

1. Define sentinel in `backend/internal/pkg/errors/errors.go`:
   ```go
   var ErrNewThing = &AppError{Code: "NEW_THING_ERROR", Message: "description", Status: 422}
   ```
2. Map in service via `MapNotFound(err, apperrors.ErrNewThing)` or direct return
3. Document in this table

## Frontend Error Display Strategy

| Error type | Display method |
|-----------|---------------|
| Form validation (422 specific fields) | Inline error below field |
| Login errors (USER_DISABLED, USER_DELETED) | Login page inline error |
| Business errors (422/400 other) | Toast notification |
| Auth errors (401) | Clear token, redirect to login |
| Permission errors (403) | Toast notification |
| Not found (404) | Toast notification |
| Server errors (500) | Toast notification |

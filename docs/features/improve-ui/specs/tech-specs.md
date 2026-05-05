---
feature: "improve-ui"
generated: "2026-05-04"
status: draft
---

# Technical Specifications: Improve UI

## Error Handling

### TECH-001: Error Code Registry Pattern

**Requirement**: All API error codes follow the format `UPPER_SNAKE_CASE`, mapped to specific HTTP status codes. New error codes are registered in `backend/internal/pkg/errors/` as sentinel `AppError` values.

**Scope**: [CROSS]

**Source**: design/tech-design.md Error Handling

Registered error codes:

| Code | HTTP Status | Description |
|------|-------------|-------------|
| USER_EXISTS | 422 | Account already exists |
| USER_NOT_FOUND | 404 | User not found |
| USER_DISABLED | 403 | Account disabled (login) |
| INVALID_STATUS | 422 | Invalid user status value |
| CANNOT_DISABLE_SELF | 422 | SuperAdmin self-disable |
| TEAM_NOT_FOUND | 404 | Team not found |
| FUTURE_WEEK_NOT_ALLOWED | 422 | Future week rejected |

### TECH-002: Frontend Error Display Strategy

**Requirement**: Frontend error display follows a code-based strategy:

| Error type | Display method |
|-----------|---------------|
| Form validation (422 specific fields) | Inline error below field |
| Login errors (USER_DISABLED) | Login page inline error |
| Business errors (422/400 other) | Toast notification |
| Auth errors (401) | Clear token, redirect to login |
| Permission errors (403) | Toast notification |
| Not found (404) | Toast notification |
| Server errors (500) | Toast notification |

**Scope**: [CROSS]

**Source**: design/tech-design.md "前端错误处理"

## Security

### TECH-003: Password Handling

**Requirement**: Passwords are hashed with bcrypt before storage. The `password_hash` field has `json:"-"` tag to prevent API leakage. Initial passwords are generated server-side (12 chars, mixed case + digits).

**Scope**: [CROSS]

**Source**: design/tech-design.md Security Considerations

## Frontend Architecture

### TECH-004: shadcn/ui Component Pattern

**Requirement**: UI primitives in `src/components/ui/` follow the shadcn/ui pattern: Radix UI primitives + Tailwind CSS + `forwardRef` + `cn()` className merge. Component source code is copied into the project (not imported as a library).

**Scope**: [CROSS]

**Source**: design/tech-design.md "前端组件架构"

Component categories:
- `ui/` — Generic primitives (button, input, select, dialog, badge, etc.)
- `layout/` — Layout components (AppLayout, Sidebar)
- `shared/` — Domain-aware compositions (StatusBadge, PriorityBadge, etc.)

### TECH-005: Frontend Dependency Stack

**Requirement**: Frontend uses this dependency stack (do not add alternatives):

| Purpose | Library |
|---------|---------|
| CSS | Tailwind CSS v4 |
| UI Primitives | Radix UI |
| Class merging | clsx + tailwind-merge (via `cn()`) |
| Variant management | class-variance-authority |
| Icons | lucide-react |
| Routing | react-router-dom |
| Data fetching | @tanstack/react-query |
| State management | zustand |
| HTTP client | axios |
| Date handling | dayjs |
| Charts/Gantt | frappe-gantt (style override with Tailwind) |

**Scope**: [CROSS]

**Source**: design/tech-design.md Dependencies

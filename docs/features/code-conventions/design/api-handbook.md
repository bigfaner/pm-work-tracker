---
created: 2026-04-20
related: design/tech-design.md
---

# API Handbook: Code Conventions (Internal)

> This feature does not change any API surface. This handbook documents the **internal code interfaces** (helpers, components) being created.

## Overview

No external API changes. The "interfaces" here are internal Go functions and React components.

## Internal Interfaces

### errors.MapNotFound

**Signature**:
```go
func MapNotFound(err error, domainErr *AppError) error
```

**Behavior**:

| Input | Output |
|-------|--------|
| `gorm.ErrRecordNotFound` | `domainErr` |
| `errors.ErrNotFound` | `domainErr` |
| Any other error | Input unchanged |

#### Usage

```go
// In service methods:
item, err := s.repo.FindByID(ctx, id)
if err != nil {
    return errors.MapNotFound(err, apperrors.ErrItemNotFound)
}
```

### dto.ApplyPaginationDefaults

**Signature**:
```go
func ApplyPaginationDefaults(page, pageSize int) (offset, page, pageSize int)
```

**Behavior**:

| Input (page, pageSize) | Output (offset, page, pageSize) |
|------------------------|--------------------------------|
| (0, 0) | (0, 1, 20) |
| (-1, -5) | (0, 1, 20) |
| (3, 10) | (20, 3, 10) |
| (1, 50) | (0, 1, 50) |

### dates.ParseDate

**Signature**:
```go
func ParseDate(s string) (time.Time, error)
```

**Behavior**:

| Input | Output |
|-------|--------|
| `"2024-01-15"` | `time.Date(2024, 1, 15, 0, 0, 0, 0, UTC)`, nil |
| `"invalid"` | zero time, error |
| `""` | zero time, error |

### Textarea Component

**Props**: Extends `React.TextareaHTMLAttributes<HTMLTextAreaElement>` (same pattern as Input).

**Default classes**: `flex w-full rounded-md border border-border-dark bg-white px-3 py-2 text-[13px] text-primary shadow-sm placeholder:text-tertiary focus:outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-200 min-h-[72px] resize-y`

### PrioritySelectItems Component

**Props**: None.

**Behavior**: Renders exactly 3 `<SelectItem>` options: P1, P2, P3. Caller provides `<Select>`, `<SelectTrigger>`, and `<SelectContent>`.

#### Usage

```tsx
<Select value={priority} onValueChange={setPriority}>
  <SelectTrigger><SelectValue placeholder="Priority" /></SelectTrigger>
  <SelectContent>
    <PrioritySelectItems />
  </SelectContent>
</Select>
```

## Error Codes

No new error codes. All existing error codes remain unchanged.

## Data Contracts

No data contract changes. JSON serialization format unchanged (already camelCase).

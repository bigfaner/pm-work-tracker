---
scope: backend
source: feature/code-conventions TECH-001–004
---

# Backend Helper Contracts

Shared helper functions that replace inline duplication across the backend.

## TECH-helpers-001: MapNotFound

**Package**: `internal/pkg/errors`

**Signature**:
```go
func MapNotFound(err error, domainErr *AppError) error
```

**Behavior**: Maps `gorm.ErrRecordNotFound` and `ErrNotFound` to the provided domain error. All other errors pass through unchanged.

**Usage**:
```go
item, err := s.repo.FindByID(ctx, id)
if err != nil {
    return nil, pkgerrors.MapNotFound(err, apperrors.ErrItemNotFound)
}
```

## TECH-helpers-002: ApplyPaginationDefaults

**Package**: `internal/dto`

**Signature**:
```go
func ApplyPaginationDefaults(page, pageSize int) (offset, page, pageSize int)
```

**Behavior**:

| Input (page, pageSize) | Output (offset, page, pageSize) |
|------------------------|--------------------------------|
| (0, 0)                 | (0, 1, 20)                     |
| (-1, -5)               | (0, 1, 20)                     |
| (3, 10)                | (20, 3, 10)                    |

**Constraint**: Apply in handler layer only. Repos receive pre-computed offset/limit.

## TECH-helpers-003: ParseDate

**Package**: `internal/pkg/dates`

**Signature**:
```go
const DateFormat = "2006-01-02"

func ParseDate(s string) (time.Time, error)
```

**Behavior**: Parses YYYY-MM-DD format. Returns zero time + error for invalid input.

## TECH-helpers-004: Handler Constructor Panic

**Pattern**: All handler constructors validate dependencies with panic-on-nil.

**Format**: `panic("<snake_case_handler>: <interfaceName> must not be nil")`

**Example**:
```go
func NewMainItemHandler(svc MainItemService) *MainItemHandler {
    if svc == nil {
        panic("main_item_handler: mainItemService must not be nil")
    }
    return &MainItemHandler{svc: svc}
}
```

**Rationale**: Catches miswired dependencies at startup, zero runtime cost per request.

---
feature: "code-conventions"
generated: "2026-05-04"
status: draft
---

# Technical Specifications: Code Conventions

## Backend Shared Helpers

### TECH-001: MapNotFound Error Mapping

**Requirement**: A generic `MapNotFound(err error, domainErr *AppError) error` function in `pkg/errors/` maps `gorm.ErrRecordNotFound` and `ErrNotFound` to domain-specific errors. All other errors pass through unchanged.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section 1, design/api-handbook.md

Usage:
```go
item, err := s.repo.FindByID(ctx, id)
if err != nil {
    return nil, pkgerrors.MapNotFound(err, apperrors.ErrItemNotFound)
}
```

Replaces all per-domain `mapXxxNotFound` functions.

### TECH-002: ApplyPaginationDefaults

**Requirement**: `ApplyPaginationDefaults(page, pageSize int) (offset, page, pageSize int)` normalizes pagination params. Defaults: page=1, pageSize=20. Applied in handler layer only; repos receive pre-computed offset/limit.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section 1, design/api-handbook.md

| Input (page, pageSize) | Output (offset, page, pageSize) |
|------------------------|--------------------------------|
| (0, 0)                 | (0, 1, 20)                     |
| (-1, -5)               | (0, 1, 20)                     |
| (3, 10)                | (20, 3, 10)                    |

### TECH-003: ParseDate Helper

**Requirement**: `ParseDate(s string) (time.Time, error)` in `pkg/dates/` parses YYYY-MM-DD format. Replaces all inline `time.Parse("2006-01-02", ...)` calls.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section 1, design/api-handbook.md

## Constructor Validation

### TECH-004: Handler Constructor Panic Pattern

**Requirement**: All handler constructors validate dependencies with panic-on-nil. Panic format: `<snake_case_handler>: <interfaceName> must not be nil`. No method-level nil checks.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section 2

```go
func NewMainItemHandler(svc MainItemService) *MainItemHandler {
    if svc == nil {
        panic("main_item_handler: mainItemService must not be nil")
    }
    return &MainItemHandler{svc: svc}
}
```

## Frontend Component Patterns

### TECH-005: Textarea Component

**Requirement**: `ui/textarea.tsx` uses `forwardRef + cn()` pattern mirroring `input.tsx`. Fixed base class string with `min-h-[72px] resize-y`.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section 3

### TECH-006: PrioritySelectItems Component

**Requirement**: `shared/PrioritySelectItems` renders exactly 3 SelectItem options (P1, P2, P3). Caller provides Select/Trigger/Content wrapper. Component has no props.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section 3

### TECH-007: Color Token Mapping

**Requirement**: All UI colors use CSS custom property theme tokens. Hardcoded Tailwind color classes (emerald-*, red-*, amber-*, slate-*) are prohibited.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section 4

| Hardcoded        | Token              |
|------------------|--------------------|
| `emerald-*`      | `success` family   |
| `red-*`          | `error` family     |
| `amber-*`        | `warning` family   |
| `slate-*`        | `secondary/tertiary` |

## Lint Configuration

### TECH-008: Lint Standards

**Requirement**: golangci-lint enables `tagliatelle` (json: camel) and `dupl` (threshold: 80). ESLint uses `no-restricted-syntax` to block hardcoded Tailwind colors.

**Scope**: [CROSS]

**Source**: design/tech-design.md Section 6

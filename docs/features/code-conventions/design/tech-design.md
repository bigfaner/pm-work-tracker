---
created: 2026-04-20
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: Code Conventions & Code Cleanup

## Overview

This feature has two goals:
1. **Codify conventions** into machine-readable artifacts (`.claude/rules/`, lint configs, architecture docs) so AI sessions and developers get consistent guidance.
2. **Clean up existing code** by eliminating duplicate patterns and unifying UI token usage.

**Scope corrections**:
1. PRD assumed snake_case JSON tags existed in models. Audit confirms **all model/dto/vo JSON tags are already camelCase**. We keep lint rules (tagliatelle) for regression prevention, but skip manual JSON tag cleanup. Similarly, the frontend has no snake_case bridge code to remove.
2. PRD Phase 2 calls for "unified repo CRUD pattern" as code changes. Audit shows 7 repos already follow a consistent pattern with minor variation. Demoted to documentation-only (`.claude/rules/patterns.md`), no code refactoring needed.

## Architecture

### Layer Placement

This feature touches **all layers** but does not add new ones:

```
.claude/rules/          ← NEW: AI convention files
docs/                   ← NEW: ARCHITECTURE.md, DECISIONS.md
backend/
  internal/pkg/         ← MODIFY: add helpers (errors, dates)
  internal/dto/         ← MODIFY: add pagination helper
  internal/handler/     ← MODIFY: constructor validation, remove nil checks
  internal/service/     ← MODIFY: use shared helpers
  .golangci.yml         ← MODIFY: add tagliatelle, dupl
frontend/
  src/components/ui/    ← NEW: textarea.tsx
  src/components/shared/← NEW: PrioritySelect.tsx
  eslint.config.js      ← NEW: color/textarea restrictions
  src/index.css         ← REFERENCE: existing theme tokens
```

### Component Diagram

```
Phase 1: Documentation & Lint
┌─────────────────┐     ┌──────────────────┐
│ .claude/rules/  │     │ Lint Configs     │
│ naming.md       │     │ golangci.yml     │
│ patterns.md     │     │ eslint.config.js │
│ frontend.md     │     └──────────────────┘
│ testing.md      │
└─────────────────┘
┌─────────────────┐     ┌──────────────────┐
│ docs/           │     │ Validation       │
│ ARCHITECTURE.md │     │ tagliatelle ✓    │
│ DECISIONS.md    │     │ ESLint ✓         │
└─────────────────┘     └──────────────────┘

Phase 2: Backend Cleanup
┌──────────┐    ┌──────────────┐    ┌─────────────┐
│ Handler  │───>│ Constructor  │    │ pkg/errors  │
│ (remove  │    │ panic on nil │    │ MapNotFound │
│  31 nil  │    └──────────────┘    │ generic fn  │
│  checks) │                        └─────────────┘
└──────────┘
┌──────────┐    ┌──────────────┐    ┌─────────────┐
│ Service  │───>│ dto/         │    │ pkg/dates   │
│ (use     │    │ ApplyDefaults│    │ ParseDate   │
│  helpers)│    │ pagination   │    │ helper      │
└──────────┘    └──────────────┘    └─────────────┘

Phase 3: Frontend Cleanup
┌──────────────────┐    ┌───────────────────┐
│ ui/textarea.tsx  │    │ shared/           │
│ (replace 14 raw  │    │ PrioritySelect.tsx│
│  <textarea>)     │    │(replace 21 copies)│
└──────────────────┘    └───────────────────┘
┌──────────────────┐
│ Color migration  │
│ 58 hardcoded →   │
│ theme tokens     │
└──────────────────┘
```

### Dependencies

**New Go dependencies**: None (helpers use existing stdlib + gorm)
**New npm dependencies**: None (ESLint already in devDeps, rules are built-in)
**New tool dependencies**: None (golangci-lint v2 already installed)

## Interfaces

### 1. Backend Shared Helpers

**pkg/errors — MapNotFound** (`internal/pkg/errors/`):

```go
// MapNotFound wraps any not-found error as a domain-specific AppError.
// Returns the original error unchanged if it's not a not-found error.
func MapNotFound(err error, domainErr *AppError) error {
    if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, ErrNotFound) {
        return domainErr
    }
    return err
}
```

Replaces: `mapItemNotFound`, `mapSubItemNotFound`, `mapPoolItemNotFound`, `mapMainItemNotFound`, `mapProgressNotFound` (5 functions).

**dto — ApplyPaginationDefaults** (`internal/dto/`):

```go
// ApplyPaginationDefaults sets missing pagination fields to defaults.
func ApplyPaginationDefaults(page, pageSize int) (offset, page, pageSize int) {
    if page <= 0 { page = 1 }
    if pageSize <= 0 { pageSize = 20 }
    offset = (page - 1) * pageSize
    return
}
```

Replaces: 6 inline pagination blocks (3 in handlers, 3 in repos). Apply once in handler layer only; repos accept pre-computed offset/limit.

**pkg/dates — ParseDate** (new file `internal/pkg/dates/dates.go`):

```go
const DateFormat = "2006-01-02"

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(s string) (time.Time, error) {
    return time.Parse(DateFormat, s)
}
```

Replaces: 11 inline `time.Parse("2006-01-02", ...)` calls.

### 2. Constructor Validation Pattern

Replace per-method nil-service checks with constructor-time validation:

```go
// Before (current — 31 occurrences):
func (h *MainItemHandler) List(c *gin.Context) {
    if h.svc == nil {
        c.JSON(http.StatusNotImplemented, gin.H{...})
        return
    }
    // ...
}

// After:
func NewMainItemHandler(svc MainItemService) *MainItemHandler {
    if svc == nil {
        panic("main_item_handler: service must not be nil")
    }
    return &MainItemHandler{svc: svc}
}
```

All 6 handler constructors get this pattern. The 31 method-level nil checks are removed.

**统一 panic 格式**：`panic("<handler_snake_case>: <service_interface_name> must not be nil")`，例如：
- `panic("main_item_handler: mainItemService must not be nil")`
- `panic("team_handler: teamService must not be nil")`

### 3. Frontend Components

**Textarea** (`src/components/ui/textarea.tsx`):

Mirror existing `input.tsx` pattern. Uses the same class string currently copy-pasted 14 times:

```tsx
// Pattern follows input.tsx: forwardRef + cn() merge
const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, ...props }, ref) => (
    <textarea
      className={cn(baseClasses, className)}
      ref={ref}
      {...props}
    />
  )
)
```

Base classes: `flex w-full rounded-md border border-border-dark bg-white px-3 py-2 text-[13px] text-primary shadow-sm placeholder:text-tertiary focus:outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-200 min-h-[72px] resize-y`

**PrioritySelect** (`src/components/shared/PrioritySelect.tsx`):

Provides only the 3 `<SelectItem>` options — caller supplies `<Select>`, `<SelectTrigger>`, and `<SelectContent>`:

```tsx
function PrioritySelectItems() {
  return (
    <>
      <SelectItem value="P1">P1</SelectItem>
      <SelectItem value="P2">P2</SelectItem>
      <SelectItem value="P3">P3</SelectItem>
    </>
  )
}
```

Usage:
```tsx
<Select value={priority} onValueChange={setPriority}>
  <SelectTrigger>...</SelectTrigger>
  <SelectContent>
    <PrioritySelectItems />
  </SelectContent>
</Select>
```

Replaces 21 inline `<SelectItem value="P1|P2|P3">` groups across 3 pages.

### 4. Color Token Mapping

Mapping from hardcoded Tailwind colors to existing theme tokens:

| Hardcoded | Token | Files affected |
|-----------|-------|----------------|
| `emerald-500/600/700` | `success` / `success-text` / `success-bg`（浅蓝色） | badge, button, toast, ProgressBar, WeeklyView |
| `red-400/500/600/800` | `error` / `error-text` / `error-bg` | badge, button, toast, dropdown-menu, pages |
| `amber-300/500/600/800` | `warning` / `warning-text` / `warning-bg` | badge, button, toast, ProgressBar, pages |
| `slate-*` | `secondary` / `tertiary` | various |

Strategy: Replace in `ui/` components first (badge, button, toast), then shared components, then pages.

**Theme token value change**: `--color-success` 系列从绿色改为浅蓝色：

| Token | 当前值（绿色） | 目标值（浅蓝色） |
|-------|---------------|-----------------|
| `--color-success` | `#059669` | `#3b82f6`（primary-500 蓝） |
| `--color-success-bg` | `#ecfdf5` | `#eff6ff`（primary-50） |
| `--color-success-text` | `#047857` | `#1d4ed8`（primary-700） |

### 5. `.claude/rules/` File Structure

Each file ≤ 200 lines, with positive and negative examples:

| File | Path Qualifier | Content |
|------|---------------|---------|
| `naming.md` | none (global) | JSON tag camelCase, Go exported names, TS interface naming |
| `patterns.md` | `backend/**/*.go` | Handler/Service/Repo templates, error handling, constructor pattern |
| `frontend.md` | `frontend/src/**/*.{ts,tsx}` | Component extraction rules, theme token usage, API module pattern |
| `testing.md` | none (global) | TDD flow, test naming, test helpers, mock patterns |

### 6. Lint Configuration

**Backend** — additions to `backend/.golangci.yml`:

```yaml
linters:
  enable:
    - tagliatelle    # enforce camelCase JSON tags
    - dupl           # detect duplicate code (threshold: 80)

linters-settings:
  tagliatelle:
    rules:
      json: camel
  dupl:
    threshold: 80
```

**Frontend** — new `frontend/eslint.config.js`:

```js
export default [
  js.configs.recommended,
  {
    files: ['src/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-syntax': [
        'error',
        // Block hardcoded Tailwind color classes
        { selector: 'Literal[value=/\\b(emerald|red|amber|slate)-\\d/]',
          message: 'Use theme tokens (success-*, error-*, warning-*, secondary, tertiary) instead of hardcoded colors.' },
      ],
    },
  },
];
```

## Data Models

No data model changes. No database migrations.

**Key clarification**: JSON tag format is a serialization concern, not a storage concern. GORM maps `json:"teamId"` to DB column `team_id` automatically via its naming strategy. No DB column changes needed.

## Error Handling

No new error types. The `MapNotFound` helper simplifies existing error mapping. Existing `apperrors.ErrXxx` sentinel values already carry correct HTTP status codes (e.g., `ErrItemNotFound` → 404), which are handled by the error middleware (`apperrors.RespondError`). `MapNotFound` only changes the translation from generic → domain-specific error; the HTTP response chain remains unchanged.

```go
// Before (in each service file):
func mapItemNotFound(err error) error { ... }  // 5 copies

// After (in service methods):
err = errors.MapNotFound(err, apperrors.ErrItemNotFound)
```

## Testing Strategy

### Tooling

| Layer | Framework | Libraries |
|-------|-----------|-----------|
| Go unit/integration | `go test` | `github.com/stretchr/testify` (assert + require) |
| Go lint | `golangci-lint v2` | tagliatelle, dupl |
| Frontend unit | `vitest` | `@testing-library/react`, `@testing-library/user-event` |
| Frontend E2E | `@playwright/test` | — |
| Frontend lint | `eslint` | `no-restricted-syntax` |

### Unit Tests

| Component | Test approach |
|-----------|--------------|
| `pkg/errors.MapNotFound` | Test with gorm.ErrRecordNotFound, ErrNotFound, and non-not-found errors |
| `dto.ApplyPaginationDefaults` | Test boundary cases (0, negative, normal values) |
| `pkg/dates.ParseDate` | Test valid date, invalid format, empty string |
| `Textarea` component | Render + className verification (mirrors input.test.tsx pattern) |
| `PrioritySelect` component | Render + option count verification |

### Integration Tests

- Run full backend test suite after each phase of cleanup
- Run full frontend test suite after each phase of cleanup
- No new integration tests needed — changes are structural refactors

### Coverage Target

- Maintain existing coverage (no regression)
- New helper functions should have ≥90% coverage

### Lint Validation Gates

| Phase | Gate |
|-------|------|
| Phase 1 | golangci-lint tagliatelle passes on all model files; ESLint color rule flags 50+ violations (existing hardcoded colors) |
| Phase 2 | All backend tests pass; `dupl` threshold violations reduced by 50%+; 0 nil-service checks in handlers |
| Phase 3 | All frontend tests pass; 0 hardcoded color classes; 14 raw `<textarea>` → `<Textarea>`; 21 PrioritySelect copies → 1 component |

## Security Considerations

No security impact. Changes are:
- Documentation and lint rules (no runtime effect)
- Internal code structure refactoring (no API surface change)
- UI component extraction (no behavior change)

## Implementation Phases

### Phase 1: Documentation & Lint (8 artifacts)

1. Write `docs/ARCHITECTURE.md` — layer descriptions, responsibilities, templates
2. Write `docs/DECISIONS.md` — key decisions (camelCase, VO layer, DI pattern, etc.)
3. Write `.claude/rules/naming.md` — naming conventions with examples
4. Write `.claude/rules/patterns.md` — backend layer patterns with templates
5. Write `.claude/rules/frontend.md` — frontend component patterns
6. Write `.claude/rules/testing.md` — testing conventions
7. Update `backend/.golangci.yml` — add tagliatelle + dupl
8. Create `frontend/eslint.config.js` — add color restriction rules

### Phase 2: Backend Cleanup (4 changes)

1. Add constructor validation to all 6 handler constructors, remove 31 nil checks
2. Add `errors.MapNotFound`, replace 5 domain-specific functions + 8 inline checks
3. Add `dto.ApplyPaginationDefaults`, remove 6 inline blocks (handlers only)
4. Add `pkg/dates.ParseDate`, replace 11 inline calls

### Phase 3: Frontend Cleanup (4 changes)

1. Create `ui/textarea.tsx`, replace 14 raw `<textarea>` elements
2. Create `shared/PrioritySelect.tsx`, replace 21 inline option groups
3. Migrate 58 hardcoded color instances to theme tokens
4. Clean up raw `<button>` elements where Button component applies

## Open Questions

None. All resolved:
- [x] `dupl` threshold → **80** (reduces test boilerplate noise)
- [x] ESLint color scope → **emerald/red/amber/slate**
- [x] PrioritySelect → **options only** (caller provides Select/Trigger/Content)

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| Central `internal/pkg/helpers/` package | One place for all helpers | Violates domain separation; becomes a junk drawer | Scatter by domain keeps helpers close to their context |
| Middleware for nil-service check | Catch at request time | Doesn't prevent startup with miswired deps; adds latency | Constructor panic catches bugs at startup, zero runtime cost |
| Code generation for CRUD patterns | Eliminate boilerplate | Over-engineering for 7 repos; hides logic | Documenting the pattern is sufficient; actual code is clear |
| CSS-in-JS for theme tokens | Dynamic theming | Added dependency; Tailwind v4 has native CSS var support | Existing Tailwind v4 `@theme` tokens already work |

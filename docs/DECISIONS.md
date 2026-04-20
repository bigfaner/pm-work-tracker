# Technical Decisions

Key architectural and coding decisions for PM Work Tracker, with context and rationale.

## 1. camelCase JSON Tags

**Context**: Go convention uses snake_case for struct tags, but the frontend is TypeScript/JavaScript where camelCase is standard.

**Choice**: All JSON tags use camelCase (`json:"teamId"`, `json:"createdAt"`).

**Rationale**: The frontend is the primary consumer of the API. Using camelCase eliminates manual key mapping in the API layer and reduces the chance of mismatches. GORM's naming strategy automatically maps `teamId` to the `team_id` database column, so there is zero impact on the storage layer.

**Consequences**: The `tagliatelle` linter enforces this convention going forward. Database column names remain snake_case (via GORM).

## 2. VO Layer for API Response Shaping

**Context**: GORM models contain internal fields (`DeletedAt`, relation foreign keys) and raw `time.Time` values that are not suitable for direct API responses.

**Choice**: A dedicated `vo` (View Object) package converts models to frontend-friendly structs.

**Rationale**: Without a VO layer, handlers would either expose internal model details or inline formatting logic. The VO layer gives a clean separation: models represent storage, VOs represent the API contract. VOs handle time formatting (`time.Time` -> `"2006-01-02"` string), field omission, and nested struct flattening.

**Consequences**: Each new model that needs API exposure requires a corresponding VO struct and constructor function. The conversion is explicit and happens in handlers, not services.

## 3. Constructor DI with Panic on Nil

**Context**: Handlers need service dependencies. Without validation, a nil service causes cryptic panics at request time.

**Choice**: Handler constructors validate all dependencies and `panic` immediately if any is nil. Method-level nil-service checks are removed entirely.

**Rationale**: Panicking at startup (during wiring) is preferable to returning a 501 error at request time. A nil dependency is a programming error, not a runtime condition. The panic makes the bug immediately visible during development or integration testing.

**Consequences**: All 6 handler constructors must validate their dependencies. The 31 method-level nil checks are removed, reducing boilerplate. Tests must provide non-nil dependencies.

## 4. Shared Helpers vs Local Copies

**Context**: Several utility patterns (not-found error mapping, pagination defaults, date parsing) were duplicated across multiple service and handler files.

**Choice**: Extract to shared helpers: `pkg/errors.MapNotFound`, `dto.ApplyPaginationDefaults`, `pkg/dates.ParseDate`. Each helper lives in the package closest to its domain.

**Rationale**: Scattering helpers by domain (errors in `pkg/errors`, pagination in `dto`, dates in `pkg/dates`) keeps them discoverable and avoids a catch-all `helpers` package. Shared helpers reduce duplication from 5+ copies to 1, making future changes single-point.

**Consequences**: New code should use these helpers instead of writing inline equivalents. The `dupl` linter (threshold 80) catches future duplication.

## 5. Tailwind v4 Theme Tokens vs Hardcoded Colors

**Context**: The codebase had hardcoded Tailwind color classes (`emerald-500`, `red-600`, `amber-300`) scattered across components and pages, making UI consistency difficult.

**Choice**: Use CSS custom property theme tokens (`text-success`, `text-error`, `text-warning`, `text-secondary`) defined in `frontend/src/index.css` via Tailwind v4's `@theme` directive.

**Rationale**: Theme tokens provide a single source of truth for the color palette. Changing a color requires updating one CSS variable instead of hunting through 50+ files. Tokens also convey semantic meaning (`text-success` vs `text-emerald-500`).

**Consequences**: ESLint `no-restricted-syntax` rule blocks hardcoded color classes. New components must use theme tokens. The mapping: `emerald-*` -> `success-*`, `red-*` -> `error-*`, `amber-*` -> `warning-*`, `slate-*` -> `secondary/tertiary`.

## 6. Interface-Driven Repositories

**Context**: Services need to access data without depending on GORM directly, and tests need to mock data access.

**Choice**: Repository interfaces are defined in `backend/internal/repository/` (e.g., `MainItemRepo`, `TeamRepo`), with GORM implementations in `repository/gorm/`.

**Rationale**: Interfaces decouple services from the database library. Tests can create mock repos. If we ever switch from GORM to another ORM or raw SQL, only the `gorm/` implementations change.

**Consequences**: Each new entity requires both an interface file and a GORM implementation file. The interface should live at `repository/xxx_repo.go` and the implementation at `repository/gorm/xxx_repo.go`.

## 7. Zustand for State Management

**Context**: React applications need client-side state for auth tokens, selected team, and UI state.

**Choice**: Zustand stores over React Context or Redux.

**Rationale**: Zustand has minimal boilerplate, no providers, and built-in selector optimization. For this app's state complexity (auth + team selection), Zustand is sufficient without the overhead of Redux.

**Consequences**: Stores are in `frontend/src/store/`. Currently `auth.ts` and `team.ts`. New global state should follow the same pattern: one file per store, flat structure, no nested reducers.

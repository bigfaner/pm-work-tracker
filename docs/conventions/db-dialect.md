---
scope: backend, database
source: feature/db-dialect-compat
---

# Database Dialect Abstraction Conventions

## TECH-dialect-001: Dialect Interface

A `Dialect` interface in `pkg/dbutil` abstracts SQLite/MySQL SQL syntax differences. Repositories and migrations call Dialect methods instead of writing raw SQL.

```go
type Dialect interface {
    CastInt(expr ColumnExpr) string  // SQLite: CAST(expr AS INTEGER), MySQL: CAST(expr AS SIGNED)
    Substr(str ColumnExpr, start int) string  // SQLite: SUBSTR(str, start), MySQL: SUBSTRING(str, start)
    Now() string  // SQLite: datetime('now'), MySQL: CURRENT_TIMESTAMP
}
```

Created via `NewDialect(db *gorm.DB) Dialect` which reads `db.Dialector.Name()` to select the implementation.

**Why**: The project supports both SQLite (dev/test) and MySQL (production). Hardcoded SQLite syntax in repository queries causes MySQL runtime errors. The Dialect interface centralizes all SQL dialect branching.

**Example**:
```go
// Before (broken on MySQL):
Select("MAX(CAST(SUBSTR(code, ?) AS INTEGER))", offset)

// After (dialect-safe):
Select("MAX(" + d.CastInt(d.Substr(dbutil.ColCode, offset)) + ")")
```

## TECH-dialect-002: ColumnExpr Type for SQL Injection Prevention

`CastInt` and `Substr` accept `ColumnExpr` (not `string`) as the expression parameter. `ColumnExpr` is a named type with pre-defined constants (e.g., `ColCode`). `NewColumnExpr()` is exported for test use only.

Production code must use the pre-defined constants. The type constraint prevents user-controlled strings from being injected into SQL expressions.

**Why**: `CastInt` and `Substr` embed their string arguments directly into SQL (not parameterized). The `ColumnExpr` type acts as a compile-time guard — a plain `string` cannot be implicitly passed. Code reviewers can spot any non-constant `ColumnExpr` usage as a visual signal.

**Example**:
```go
// Pre-defined constants — only valid inputs for production code
const (
    ColCode ColumnExpr = "code"
)

// Correct: using constant
d.Substr(dbutil.ColCode, 7)

// Wrong: plain string (won't compile)
d.Substr("code", 7)  // cannot use "code" (type string) as ColumnExpr
```

## TECH-dialect-003: Lint Check for Hardcoded SQLite Syntax

A lint check in `scripts/lint-staged.sh` scans `backend/internal/repository/gorm/*.go` (excluding test files) for hardcoded SQLite keywords:

| Pattern | Matches | Use Instead |
|---------|---------|-------------|
| `SUBSTR(` | SQLite SUBSTR function | `dialect.Substr()` |
| `CAST(...AS INTEGER)` | SQLite CAST type | `dialect.CastInt()` |
| `datetime('now')` | SQLite time function | `dialect.Now()` |
| `pragma_table_info` | SQLite PRAGMA query | `migration.HasColumn()` |

The check blocks commits that introduce new hardcoded SQLite syntax in the repository layer.

**Why**: Prevents regression — developers writing raw SQL in repos tend to use SQLite syntax (since SQLite is the local dev database). The lint catches this before it reaches MySQL production.

## TECH-dialect-004: NewDialect Fail-Fast on Unsupported Dialect

`NewDialect(db)` panics if `db` is nil or if `db.Dialector.Name()` is not `"sqlite"` or `"mysql"`. The panic uses a typed error `UnsupportedDialectError{Name: name}` so tests can assert the specific error type via `recover()`.

**Why**: Fails at startup rather than producing confusing SQL syntax errors at runtime. An unsupported database driver is a configuration error that should be caught immediately.

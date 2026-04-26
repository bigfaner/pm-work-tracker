package dbutil

import (
	"fmt"

	"gorm.io/gorm"
)

// ColumnExpr is a string that represents a SQL column name or expression.
// Production callers must use pre-defined constants (e.g., ColCode).
// NewColumnExpr is exported for test use only — non-test callers should
// never need it.
type ColumnExpr string

// Pre-defined column expressions — the only valid inputs to Dialect methods.
const (
	ColCode ColumnExpr = "code"
)

// NewColumnExpr creates a ColumnExpr. Exported for use in tests only.
// Production code must use the constants above.
func NewColumnExpr(s string) ColumnExpr { return ColumnExpr(s) }

// Dialect abstracts SQL syntax differences between database engines.
type Dialect interface {
	// CastInt returns a CAST expression that produces an integer result.
	// SQLite: CAST(expr AS INTEGER)
	// MySQL:  CAST(expr AS SIGNED)
	CastInt(expr ColumnExpr) string

	// Substr returns a substring extraction expression.
	// SQLite: SUBSTR(str, start)
	// MySQL:  SUBSTRING(str, start)
	// start is 1-indexed (consistent across both databases).
	Substr(str ColumnExpr, start int) string

	// Now returns a datetime expression for the current timestamp.
	// SQLite: datetime('now')
	// MySQL:  CURRENT_TIMESTAMP
	Now() string
}

type sqliteDialect struct{}

func (sqliteDialect) CastInt(expr ColumnExpr) string {
	return fmt.Sprintf("CAST(%s AS INTEGER)", expr)
}

func (sqliteDialect) Substr(str ColumnExpr, start int) string {
	return fmt.Sprintf("SUBSTR(%s, %d)", str, start)
}

func (sqliteDialect) Now() string {
	return "datetime('now')"
}

type mysqlDialect struct{}

func (mysqlDialect) CastInt(expr ColumnExpr) string {
	return fmt.Sprintf("CAST(%s AS SIGNED)", expr)
}

func (mysqlDialect) Substr(str ColumnExpr, start int) string {
	return fmt.Sprintf("SUBSTRING(%s, %d)", str, start)
}

func (mysqlDialect) Now() string {
	return "CURRENT_TIMESTAMP"
}

// UnsupportedDialectError is returned when an unrecognized dialect name is passed to NewDialect.
type UnsupportedDialectError struct {
	Name string
}

func (e UnsupportedDialectError) Error() string {
	return fmt.Sprintf("unsupported dialect: %s, only 'sqlite' and 'mysql' are supported", e.Name)
}

// NewDialect creates a Dialect based on the GORM Dialector name.
// Panics if db is nil (consistent with project panic-on-nil pattern).
// Panics with UnsupportedDialectError for unrecognized dialect names.
func NewDialect(db *gorm.DB) Dialect {
	if db == nil {
		panic("dbutil: db must not be nil")
	}

	switch db.Dialector.Name() {
	case "sqlite":
		return sqliteDialect{}
	case "mysql":
		return mysqlDialect{}
	default:
		panic(UnsupportedDialectError{Name: db.Dialector.Name()})
	}
}

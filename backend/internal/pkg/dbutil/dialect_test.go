package dbutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// mockDialector implements gorm.Dialector for testing.
type mockDialector struct {
	name string
}

func (d mockDialector) Name() string { return d.name }

func (d mockDialector) Initialize(_ *gorm.DB) error { return nil }

func (d mockDialector) Migrator(_ *gorm.DB) gorm.Migrator { return nil }

func (d mockDialector) DataTypeOf(_ *schema.Field) string { return "" }

func (d mockDialector) DefaultValueOf(_ *schema.Field) clause.Expression { return nil }

func (d mockDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
}

func (d mockDialector) QuoteTo(clause.Writer, string) {}

func (d mockDialector) Explain(sql string, vars ...interface{}) string { return sql }

func openDBWithDialector(name string) *gorm.DB {
	db, err := gorm.Open(mockDialector{name: name}, &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func TestNewColumnExpr(t *testing.T) {
	got := NewColumnExpr("my_col")
	assert.Equal(t, ColumnExpr("my_col"), got)
}

func TestNewDialect_SQLite(t *testing.T) {
	db := openDBWithDialector("sqlite")
	d := NewDialect(db)
	require.NotNil(t, d)

	assert.Equal(t, "CAST(item_code AS INTEGER)", d.CastInt(ColCode))
	assert.Equal(t, "SUBSTR(item_code, 7)", d.Substr(ColCode, 7))
	assert.Equal(t, "datetime('now')", d.Now())
}

func TestNewDialect_MySQL(t *testing.T) {
	db := openDBWithDialector("mysql")
	d := NewDialect(db)
	require.NotNil(t, d)

	assert.Equal(t, "CAST(item_code AS SIGNED)", d.CastInt(ColCode))
	assert.Equal(t, "SUBSTRING(item_code, 7)", d.Substr(ColCode, 7))
	assert.Equal(t, "CURRENT_TIMESTAMP", d.Now())
}

func TestNewDialect_UnsupportedDialect(t *testing.T) {
	defer func() {
		r := recover()
		require.NotNil(t, r)
		err, ok := r.(UnsupportedDialectError)
		require.True(t, ok, "expected UnsupportedDialectError, got %T: %v", r, r)
		assert.Equal(t, "postgres", err.Name)
		assert.Contains(t, err.Error(), "postgres")
	}()

	db := openDBWithDialector("postgres")
	NewDialect(db)
}

func TestNewDialect_NilDB(t *testing.T) {
	defer func() {
		r := recover()
		require.NotNil(t, r)
	}()

	NewDialect(nil)
}

func TestUnsupportedDialectError_Message(t *testing.T) {
	err := UnsupportedDialectError{Name: "oracle"}
	expected := fmt.Sprintf("unsupported dialect: %s, only 'sqlite' and 'mysql' are supported", "oracle")
	assert.Equal(t, expected, err.Error())
}

func TestDialect_ComposedExpression(t *testing.T) {
	tests := []struct {
		name     string
		dialect  Dialect
		expected string
	}{
		{
			name:     "sqlite MAX(CAST(SUBSTR(...)))",
			dialect:  NewDialect(openDBWithDialector("sqlite")),
			expected: "MAX(CAST(SUBSTR(item_code, 4) AS INTEGER))",
		},
		{
			name:     "mysql MAX(CAST(SUBSTRING(...)))",
			dialect:  NewDialect(openDBWithDialector("mysql")),
			expected: "MAX(CAST(SUBSTRING(item_code, 4) AS SIGNED))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := tt.dialect.Substr(ColCode, 4)
			cast := tt.dialect.CastInt(NewColumnExpr(sub))
			expr := "MAX(" + cast + ")"
			assert.Equal(t, tt.expected, expr)
		})
	}
}

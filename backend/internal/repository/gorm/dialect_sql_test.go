package gorm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"pm-work-tracker/backend/internal/pkg/dbutil"
)

// mockDialector implements gorm.Dialector for testing dialect SQL generation.
type mockDialector struct {
	name string
}

func (d mockDialector) Name() string                         { return d.name }
func (d mockDialector) Initialize(_ *gorm.DB) error          { return nil }
func (d mockDialector) Migrator(_ *gorm.DB) gorm.Migrator    { return nil }
func (d mockDialector) DataTypeOf(_ *schema.Field) string    { return "" }
func (d mockDialector) DefaultValueOf(_ *schema.Field) clause.Expression { return nil }
func (d mockDialector) BindVarTo(clause.Writer, *gorm.Statement, interface{}) {}
func (d mockDialector) QuoteTo(clause.Writer, string)        {}
func (d mockDialector) Explain(sql string, vars ...interface{}) string { return sql }

func openDBWithDialector(name string) *gorm.DB {
	db, err := gorm.Open(mockDialector{name: name}, &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func TestDialectSelectClause_SQLite(t *testing.T) {
	d := dbutil.NewDialect(openDBWithDialector("sqlite"))
	require.NotNil(t, d)

	startPos := len("FEAT") + 1 // = 5
	expr := "MAX(" + d.CastInt(dbutil.NewColumnExpr(d.Substr(dbutil.ColCode, startPos))) + ")"
	assert.Equal(t, "MAX(CAST(SUBSTR(item_code, 5) AS INTEGER))", expr)
}

func TestDialectSelectClause_MySQL(t *testing.T) {
	d := dbutil.NewDialect(openDBWithDialector("mysql"))
	require.NotNil(t, d)

	startPos := len("FEAT") + 1 // = 5
	expr := "MAX(" + d.CastInt(dbutil.NewColumnExpr(d.Substr(dbutil.ColCode, startPos))) + ")"
	assert.Equal(t, "MAX(CAST(SUBSTRING(item_code, 5) AS SIGNED))", expr)
}

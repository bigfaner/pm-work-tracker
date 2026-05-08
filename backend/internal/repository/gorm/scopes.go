package gorm

import (
	"fmt"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/pkg/dbutil"
)

// NotDeleted filters out soft-deleted records (deleted_flag = 0).
// Apply to all Find/First/List/Count queries on tables with soft-delete support.
func NotDeleted(db *gormlib.DB) *gormlib.DB {
	return db.Where("deleted_flag = 0")
}

// NotDeletedTable returns a scope that filters soft-deleted records with a table-qualified column.
// Use in multi-table queries (joins) where "deleted_flag" is ambiguous.
func NotDeletedTable(table string) func(db *gormlib.DB) *gormlib.DB {
	return func(db *gormlib.DB) *gormlib.DB {
		return db.Where(table + ".deleted_flag = 0")
	}
}

// nextSeqInTx checks whether child records have a higher code sequence than the current counter,
// and if so, advances the counter. Returns the sequence to use for formatting the next code.
func nextSeqInTx(tx *gormlib.DB, dialect dbutil.Dialect, parentTable, seqCol, parentKeyCol string, parentKeyVal int64, seq uint, code string, filterCol string, filterVal int64, childModel interface{}) (uint, error) {
	// If child records were inserted directly with a higher seq, skip past them.
	var maxSeq *int
	subExpr := dialect.Substr(dbutil.ColCode, len(code)+2)
	castExpr := dialect.CastInt(dbutil.NewColumnExpr(subExpr))
	if err := tx.Model(childModel).
		Where(filterCol+" = ?", filterVal).
		Select("MAX(" + castExpr + ")").
		Scan(&maxSeq).Error; err != nil {
		return 0, err
	}
	if maxSeq != nil && uint(*maxSeq) >= seq {
		seq = uint(*maxSeq) + 1
		updateSQL := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", parentTable, seqCol, parentKeyCol)
		if err := tx.Exec(updateSQL, seq, parentKeyVal).Error; err != nil {
			return 0, err
		}
	}
	return seq, nil
}

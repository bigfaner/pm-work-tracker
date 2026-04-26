package gorm

import "gorm.io/gorm"

// NotDeleted filters out soft-deleted records (deleted_flag = 0).
// Apply to all Find/First/List/Count queries on tables with soft-delete support.
func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_flag = 0")
}

// NotDeletedTable returns a scope that filters soft-deleted records with a table-qualified column.
// Use in multi-table queries (joins) where "deleted_flag" is ambiguous.
func NotDeletedTable(table string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(table + ".deleted_flag = 0")
	}
}

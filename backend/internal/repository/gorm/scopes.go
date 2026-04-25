package gorm

import "gorm.io/gorm"

// NotDeleted filters out soft-deleted records (deleted_flag = 0).
// Apply to all Find/First/List/Count queries on tables with soft-delete support.
func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_flag = 0")
}

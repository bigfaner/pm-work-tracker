package gorm

import gormlib "gorm.io/gorm"

// applyItemFilter applies shared filter fields (status, priority, assigneeID, isKeyItem)
// to a GORM query. Used by both MainItem and SubItem repositories.
func applyItemFilter(query *gormlib.DB, status, priority string, assigneeID *uint, isKeyItem *bool) *gormlib.DB {
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if assigneeID != nil {
		query = query.Where("assignee_id = ?", *assigneeID)
	}
	if isKeyItem != nil {
		query = query.Where("is_key_item = ?", *isKeyItem)
	}
	return query
}

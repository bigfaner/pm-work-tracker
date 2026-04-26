package gorm

import gormlib "gorm.io/gorm"

// applyItemFilter applies shared filter fields (status, priority, assigneeKey, isKeyItem)
// to a GORM query. Used by both MainItem and SubItem repositories.
func applyItemFilter(query *gormlib.DB, status, priority string, assigneeKey *string, isKeyItem *bool) *gormlib.DB {
	if status != "" {
		query = query.Where("item_status = ?", status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if assigneeKey != nil && *assigneeKey != "" {
		query = query.Where("assignee_key = ?", *assigneeKey)
	}
	if isKeyItem != nil {
		query = query.Where("is_key_item = ?", *isKeyItem)
	}
	return query
}

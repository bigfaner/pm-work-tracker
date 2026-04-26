package gorm

import (
	"strconv"

	gormlib "gorm.io/gorm"
)

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
		ak, err := strconv.ParseInt(*assigneeKey, 10, 64)
		if err != nil {
			// Invalid assigneeKey: return empty result, never all items
			query = query.Where("1 = 0")
		} else {
			query = query.Where("assignee_key = ?", ak)
		}
	}
	if isKeyItem != nil {
		query = query.Where("is_key_item = ?", *isKeyItem)
	}
	return query
}

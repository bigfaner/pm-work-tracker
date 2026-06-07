// Package gorm provides GORM-based repository implementations.
package gorm

import (
	"strconv"

	gormlib "gorm.io/gorm"
)

// applyItemFilter applies shared filter fields (statuses, priority, assigneeKey, isKeyItem)
// to a GORM query. Used by both MainItem and SubItem repositories.
func applyItemFilter(query *gormlib.DB, statuses []string, priority string, assigneeKey *string, isKeyItem *bool) *gormlib.DB {
	if len(statuses) > 0 {
		query = query.Where("item_status IN ?", statuses)
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

// applyMilestoneKeyFilter applies a milestone_key filter to a GORM query.
// A non-nil, non-empty value filters by the parsed bizKey.
// The special value "unassigned" filters for NULL milestone_key.
func applyMilestoneKeyFilter(query *gormlib.DB, milestoneKey *string) *gormlib.DB {
	if milestoneKey == nil || *milestoneKey == "" {
		return query
	}
	if *milestoneKey == "unassigned" {
		return query.Where("milestone_key IS NULL")
	}
	mk, err := strconv.ParseInt(*milestoneKey, 10, 64)
	if err != nil {
		// Invalid milestoneKey: return empty result
		return query.Where("1 = 0")
	}
	return query.Where("milestone_key = ?", mk)
}

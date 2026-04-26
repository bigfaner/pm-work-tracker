package repo

import (
	"context"
	"errors"
	"fmt"

	gormlib "gorm.io/gorm"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/model"
)

// acceptedFields maps model table names to their updatable field whitelist.
var acceptedFields = map[string]map[string]bool{
	model.MainItem{}.TableName(): {"title": true, "priority": true, "item_status": true, "assignee_key": true, "completion": true, "actual_end_date": true, "archived_at": true, "item_desc": true, "is_key_item": true, "plan_start_date": true, "expected_end_date": true},
	model.SubItem{}.TableName():  {"title": true, "item_status": true, "priority": true, "assignee_key": true, "completion": true, "actual_end_date": true, "item_desc": true, "plan_start_date": true, "expected_end_date": true},
	model.User{}.TableName():     {"display_name": true},
	model.ItemPool{}.TableName(): {"pool_status": true, "assigned_main_key": true, "assigned_sub_key": true, "assignee_key": true, "reject_reason": true, "reviewed_at": true, "reviewer_key": true},
}

// identifiable constrains T to model types that have an ID field.
type identifiable interface {
	model.MainItem | model.SubItem | model.User | model.Team | model.Role |
		model.ProgressRecord | model.StatusHistory | model.ItemPool | model.TeamMember
}

// FindByID retrieves a single record by primary key.
// Returns apperrors.ErrNotFound if the record does not exist.
func FindByID[T any](db *gormlib.DB, ctx context.Context, id uint) (*T, error) {
	var item T
	err := db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		if errors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

// FindByIDs retrieves multiple records by their primary keys.
// Returns a map keyed by ID. Missing IDs are absent from the map (no error).
// Empty input returns an empty map with no query.
func FindByIDs[T identifiable](db *gormlib.DB, ctx context.Context, ids []uint) (map[uint]*T, error) {
	result := make(map[uint]*T)
	if len(ids) == 0 {
		return result, nil
	}
	var items []*T
	if err := db.WithContext(ctx).Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, err
	}
	for _, item := range items {
		result[getID(item)] = item
	}
	return result, nil
}

// getID extracts the primary key from a model pointer via type switch.
func getID[T identifiable](v *T) uint {
	switch item := any(v).(type) {
	case *model.MainItem:
		return item.ID
	case *model.SubItem:
		return item.ID
	case *model.User:
		return item.ID
	case *model.Team:
		return item.ID
	case *model.Role:
		return item.ID
	case *model.ProgressRecord:
		return item.ID
	case *model.StatusHistory:
		return item.ID
	case *model.ItemPool:
		return item.ID
	case *model.TeamMember:
		return item.ID
	default:
		return 0
	}
}

// UpdateFields performs a partial update on the given model instance.
// Fields are validated against a whitelist per entity type.
// Returns apperrors.ErrInvalidField for unknown field keys.
// Returns apperrors.ErrNotFound if the team_id check fails (zero rows affected).
func UpdateFields[T any](db *gormlib.DB, ctx context.Context, item *T, teamKey int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}

	stmt := &gormlib.Statement{DB: db}
	stmt.Parse(item)
	tableName := stmt.Table
	allowed, ok := acceptedFields[tableName]
	if !ok {
		return fmt.Errorf("%w: unsupported entity %q", apperrors.ErrInvalidField, tableName)
	}
	for key := range fields {
		if !allowed[key] {
			return fmt.Errorf("%w: %q", apperrors.ErrInvalidField, key)
		}
	}

	query := db.WithContext(ctx).Model(item)
	if teamKey != 0 {
		query = query.Where("team_key = ?", teamKey)
	}
	result := query.Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

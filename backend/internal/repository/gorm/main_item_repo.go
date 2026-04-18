package gorm

import (
	"context"
	stderrors "errors"
	"fmt"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

type mainItemRepo struct {
	db *gormlib.DB
}

// NewGormMainItemRepo creates a GORM-backed MainItemRepo.
func NewGormMainItemRepo(db *gormlib.DB) repository.MainItemRepo {
	return &mainItemRepo{db: db}
}

func (r *mainItemRepo) Create(ctx context.Context, item *model.MainItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *mainItemRepo) FindByID(ctx context.Context, id uint) (*model.MainItem, error) {
	var item model.MainItem
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *mainItemRepo) Update(ctx context.Context, item *model.MainItem, fields map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(item).Where("team_id = ?", item.TeamID).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *mainItemRepo) List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	query := r.db.WithContext(ctx).Where("team_id = ?", teamID)

	// Filter out archived by default; include when filter.Archived is true
	if !filter.Archived {
		query = query.Where("archived_at IS NULL")
	}

	query = applyMainItemFilter(query, filter)

	var total int64
	if err := query.Model(&model.MainItem{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	if page.Page <= 0 {
		page.Page = 1
	}
	if page.PageSize <= 0 {
		page.PageSize = 20
	}
	offset := (page.Page - 1) * page.PageSize

	var items []model.MainItem
	if err := query.Order("id DESC").Offset(offset).Limit(page.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	return &dto.PageResult[model.MainItem]{
		Items: items,
		Total: total,
		Page:  page.Page,
		Size:  page.PageSize,
	}, nil
}

func (r *mainItemRepo) NextCode(ctx context.Context, teamID uint) (string, error) {
	var maxCode *string
	err := r.db.WithContext(ctx).
		Model(&model.MainItem{}).
		Where("team_id = ?", teamID).
		Select("MAX(code)").
		Scan(&maxCode).Error
	if err != nil {
		return "", err
	}

	if maxCode == nil || *maxCode == "" {
		return "MI-0001", nil
	}

	var seq int
	_, err = fmt.Sscanf(*maxCode, "MI-%04d", &seq)
	if err != nil {
		return "", fmt.Errorf("failed to parse code %q: %w", *maxCode, err)
	}
	return fmt.Sprintf("MI-%04d", seq+1), nil
}

func (r *mainItemRepo) CountByTeam(ctx context.Context, teamID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.MainItem{}).Where("team_id = ?", teamID).Count(&count).Error
	return count, err
}

func (r *mainItemRepo) ListNonArchivedByTeam(ctx context.Context, teamID uint) ([]model.MainItem, error) {
	var items []model.MainItem
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND archived_at IS NULL", teamID).
		Find(&items).Error
	return items, err
}

func applyMainItemFilter(query *gormlib.DB, filter dto.MainItemFilter) *gormlib.DB {
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}
	if filter.AssigneeID != nil {
		query = query.Where("assignee_id = ?", *filter.AssigneeID)
	}
	if filter.IsKeyItem != nil {
		query = query.Where("is_key_item = ?", *filter.IsKeyItem)
	}
	return query
}

package gorm

import (
	"context"
	stderrors "errors"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

type subItemRepo struct {
	db *gormlib.DB
}

// NewGormSubItemRepo creates a GORM-backed SubItemRepo.
func NewGormSubItemRepo(db *gormlib.DB) repository.SubItemRepo {
	return &subItemRepo{db: db}
}

func (r *subItemRepo) Create(ctx context.Context, item *model.SubItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *subItemRepo) FindByID(ctx context.Context, id uint) (*model.SubItem, error) {
	var item model.SubItem
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *subItemRepo) Update(ctx context.Context, item *model.SubItem, fields map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(item).Where("team_id = ?", item.TeamID).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *subItemRepo) List(ctx context.Context, teamID uint, mainItemID uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	query := r.db.WithContext(ctx).Where("team_id = ?", teamID)

	if mainItemID > 0 {
		query = query.Where("main_item_id = ?", mainItemID)
	}

	query = applySubItemFilter(query, filter)

	var total int64
	if err := query.Model(&model.SubItem{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if page.Page <= 0 {
		page.Page = 1
	}
	if page.PageSize <= 0 {
		page.PageSize = 20
	}
	offset := (page.Page - 1) * page.PageSize

	var items []model.SubItem
	if err := query.Order("id DESC").Offset(offset).Limit(page.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	return &dto.PageResult[model.SubItem]{
		Items: items,
		Total: total,
		Page:  page.Page,
		Size:  page.PageSize,
	}, nil
}

func (r *subItemRepo) ListByMainItem(ctx context.Context, mainItemID uint) ([]*model.SubItem, error) {
	var items []*model.SubItem
	err := r.db.WithContext(ctx).
		Where("main_item_id = ?", mainItemID).
		Find(&items).Error
	return items, err
}

func (r *subItemRepo) ListByTeam(ctx context.Context, teamID uint) ([]model.SubItem, error) {
	var items []model.SubItem
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Find(&items).Error
	return items, err
}

func applySubItemFilter(query *gormlib.DB, filter dto.SubItemFilter) *gormlib.DB {
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

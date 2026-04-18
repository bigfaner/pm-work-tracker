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

type itemPoolRepo struct {
	db *gormlib.DB
}

// NewGormItemPoolRepo creates a GORM-backed ItemPoolRepo.
func NewGormItemPoolRepo(db *gormlib.DB) repository.ItemPoolRepo {
	return &itemPoolRepo{db: db}
}

func (r *itemPoolRepo) Create(ctx context.Context, item *model.ItemPool) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *itemPoolRepo) FindByID(ctx context.Context, id uint) (*model.ItemPool, error) {
	var item model.ItemPool
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *itemPoolRepo) Update(ctx context.Context, item *model.ItemPool, fields map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(item).Where("team_id = ?", item.TeamID).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *itemPoolRepo) List(ctx context.Context, teamID uint, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error) {
	query := r.db.WithContext(ctx).Where("team_id = ?", teamID)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Model(&model.ItemPool{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if page.Page <= 0 {
		page.Page = 1
	}
	if page.PageSize <= 0 {
		page.PageSize = 20
	}
	offset := (page.Page - 1) * page.PageSize

	var items []model.ItemPool
	if err := query.Order("id DESC").Offset(offset).Limit(page.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	return &dto.PageResult[model.ItemPool]{
		Items: items,
		Total: total,
		Page:  page.Page,
		Size:  page.PageSize,
	}, nil
}

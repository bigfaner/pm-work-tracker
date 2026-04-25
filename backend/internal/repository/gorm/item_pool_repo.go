package gorm

import (
	"context"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/repo"
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
	return repo.FindByID[model.ItemPool](r.db, ctx, id)
}

func (r *itemPoolRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.ItemPool, error) {
	var item model.ItemPool
	err := r.db.WithContext(ctx).Where("biz_key = ?", bizKey).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *itemPoolRepo) Update(ctx context.Context, item *model.ItemPool, fields map[string]interface{}) error {
	return repo.UpdateFields[model.ItemPool](r.db, ctx, item, item.TeamID, fields)
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

	offset, p, ps := dto.ApplyPaginationDefaults(page.Page, page.PageSize)
	page.Page = p
	page.PageSize = ps

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

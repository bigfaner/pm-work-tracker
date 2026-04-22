package gorm

import (
	"context"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/repository"
)

type statusHistoryRepo struct {
	db *gormlib.DB
}

// NewGormStatusHistoryRepo creates a GORM-backed StatusHistoryRepo.
func NewGormStatusHistoryRepo(db *gormlib.DB) repository.StatusHistoryRepo {
	if db == nil {
		panic("status_history_repo: db must not be nil")
	}
	return &statusHistoryRepo{db: db}
}

func (r *statusHistoryRepo) Create(ctx context.Context, record *model.StatusHistory) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *statusHistoryRepo) FindByID(ctx context.Context, id uint) (*model.StatusHistory, error) {
	return repo.FindByID[model.StatusHistory](r.db, ctx, id)
}

func (r *statusHistoryRepo) ListByItem(ctx context.Context, itemType string, itemID uint, page dto.Pagination) (*dto.PageResult[model.StatusHistory], error) {
	query := r.db.WithContext(ctx).Where("item_type = ? AND item_id = ?", itemType, itemID)

	var total int64
	if err := query.Model(&model.StatusHistory{}).Count(&total).Error; err != nil {
		return nil, err
	}

	offset, p, ps := dto.ApplyPaginationDefaults(page.Page, page.PageSize)
	page.Page = p
	page.PageSize = ps

	var records []model.StatusHistory
	if err := query.Order("id DESC").Offset(offset).Limit(page.PageSize).Find(&records).Error; err != nil {
		return nil, err
	}

	return &dto.PageResult[model.StatusHistory]{
		Items: records,
		Total: total,
		Page:  page.Page,
		Size:  page.PageSize,
	}, nil
}

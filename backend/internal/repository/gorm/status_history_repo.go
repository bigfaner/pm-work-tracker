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
	var record model.StatusHistory
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *statusHistoryRepo) ListByItem(ctx context.Context, itemType string, itemID uint, page dto.Pagination) (*dto.PageResult[model.StatusHistory], error) {
	query := r.db.WithContext(ctx).Where("item_type = ? AND item_id = ?", itemType, itemID)

	var total int64
	if err := query.Model(&model.StatusHistory{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if page.Page <= 0 {
		page.Page = 1
	}
	if page.PageSize <= 0 {
		page.PageSize = 20
	}
	offset := (page.Page - 1) * page.PageSize

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

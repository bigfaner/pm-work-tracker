package gorm

import (
	"context"
	stderrors "errors"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/repository"
)

type decisionLogRepo struct {
	db *gormlib.DB
}

// NewGormDecisionLogRepo creates a GORM-backed DecisionLogRepo.
func NewGormDecisionLogRepo(db *gormlib.DB) repository.DecisionLogRepo {
	return &decisionLogRepo{db: db}
}

func (r *decisionLogRepo) Create(ctx context.Context, log *model.DecisionLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *decisionLogRepo) FindByID(ctx context.Context, id uint) (*model.DecisionLog, error) {
	return repo.FindByID[model.DecisionLog](r.db, ctx, id)
}

func (r *decisionLogRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.DecisionLog, error) {
	var log model.DecisionLog
	err := r.db.WithContext(ctx).Where("biz_key = ?", bizKey).First(&log).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &log, nil
}

func (r *decisionLogRepo) ListByItem(ctx context.Context, mainItemID uint, userID uint, offset, limit int) ([]model.DecisionLog, int64, error) {
	query := r.db.WithContext(ctx).
		Where("main_item_key = ? AND (log_status = 'published' OR created_by = ?)", mainItemID, userID)

	var total int64
	if err := query.Model(&model.DecisionLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []model.DecisionLog
	if err := query.Order("create_time DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *decisionLogRepo) Update(ctx context.Context, log *model.DecisionLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

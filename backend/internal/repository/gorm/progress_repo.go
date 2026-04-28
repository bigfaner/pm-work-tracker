package gorm

import (
	"context"
	stderrors "errors"
	"time"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/repository"
)

type progressRepo struct {
	db *gormlib.DB
}

// NewGormProgressRepo creates a GORM-backed ProgressRepo.
func NewGormProgressRepo(db *gormlib.DB) repository.ProgressRepo {
	return &progressRepo{db: db}
}

func (r *progressRepo) Create(ctx context.Context, record *model.ProgressRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *progressRepo) FindByID(ctx context.Context, id uint) (*model.ProgressRecord, error) {
	return repo.FindByID[model.ProgressRecord](r.db, ctx, id)
}

func (r *progressRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.ProgressRecord, error) {
	var record model.ProgressRecord
	err := r.db.WithContext(ctx).Where("biz_key = ?", bizKey).First(&record).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *progressRepo) ListBySubItem(ctx context.Context, teamBizKey int64, subItemBizKey int64) ([]model.ProgressRecord, error) {
	var records []model.ProgressRecord
	err := r.db.WithContext(ctx).
		Where("team_key = ? AND sub_item_key = ?", teamBizKey, subItemBizKey).
		Order("create_time ASC").
		Find(&records).Error
	return records, err
}

func (r *progressRepo) LatestBySubItem(ctx context.Context, subItemBizKey int64) (*model.ProgressRecord, error) {
	var record model.ProgressRecord
	err := r.db.WithContext(ctx).
		Where("sub_item_key = ?", subItemBizKey).
		Order("create_time DESC").
		First(&record).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *progressRepo) UpdateCompletion(ctx context.Context, recordID uint, completion float64) error {
	result := r.db.WithContext(ctx).
		Model(&model.ProgressRecord{}).
		Where("id = ?", recordID).
		Updates(map[string]interface{}{
			"completion_pct": completion,
			"is_pm_correct": 1,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *progressRepo) ListByTeamInRange(ctx context.Context, teamBizKey int64, start, end time.Time) ([]model.ProgressRecord, error) {
	var records []model.ProgressRecord
	err := r.db.WithContext(ctx).
		Where("team_key = ? AND create_time >= ? AND create_time < ?", teamBizKey, start, end).
		Order("create_time ASC").
		Find(&records).Error
	return records, err
}

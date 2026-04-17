package gorm

import (
	"context"
	stderrors "errors"
	"time"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/errors"
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
	var record model.ProgressRecord
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *progressRepo) ListBySubItem(ctx context.Context, teamID uint, subItemID uint) ([]model.ProgressRecord, error) {
	var records []model.ProgressRecord
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND sub_item_id = ?", teamID, subItemID).
		Order("created_at ASC").
		Find(&records).Error
	return records, err
}

func (r *progressRepo) LatestBySubItem(ctx context.Context, subItemID uint) (*model.ProgressRecord, error) {
	var record model.ProgressRecord
	err := r.db.WithContext(ctx).
		Where("sub_item_id = ?", subItemID).
		Order("created_at DESC").
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
			"completion":    completion,
			"is_pm_correct": true,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *progressRepo) ListByTeamInRange(ctx context.Context, teamID uint, start, end time.Time) ([]model.ProgressRecord, error) {
	var records []model.ProgressRecord
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND created_at >= ? AND created_at < ?", teamID, start, end).
		Order("created_at ASC").
		Find(&records).Error
	return records, err
}

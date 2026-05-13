package gorm

import (
	"context"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/repository"
)

type milestoneRepo struct {
	db *gormlib.DB
}

// NewGormMilestoneRepo creates a GORM-backed MilestoneRepo.
func NewGormMilestoneRepo(db *gormlib.DB) repository.MilestoneRepo {
	return &milestoneRepo{db: db}
}

func (r *milestoneRepo) Create(ctx context.Context, m *model.Milestone) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *milestoneRepo) FindByID(ctx context.Context, id uint) (*model.Milestone, error) {
	return repo.FindByID[model.Milestone](r.db, ctx, id)
}

func (r *milestoneRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error) {
	var m model.Milestone
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("biz_key = ?", bizKey).First(&m).Error
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrMilestoneNotFound)
	}
	return &m, nil
}

//nolint:dupl // same FindByBizKeys shape as mainItemRepo
func (r *milestoneRepo) FindByBizKeys(ctx context.Context, bizKeys []int64) (map[int64]*model.Milestone, error) {
	result := make(map[int64]*model.Milestone)
	if len(bizKeys) == 0 {
		return result, nil
	}
	var items []*model.Milestone
	if err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("biz_key IN ?", bizKeys).Find(&items).Error; err != nil {
		return nil, err
	}
	for _, item := range items {
		result[item.BizKey] = item
	}
	return result, nil
}

func (r *milestoneRepo) Update(ctx context.Context, m *model.Milestone, fields map[string]interface{}) error {
	return repo.UpdateFields[model.Milestone](r.db, ctx, m, m.TeamKey, fields)
}

func (r *milestoneRepo) ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error) {
	var items []model.Milestone
	err := r.db.WithContext(ctx).Scopes(NotDeleted).
		Where("milestone_map_key = ?", milestoneMapBizKey).
		Order("expected_end_date ASC").
		Find(&items).Error
	return items, err
}

func (r *milestoneRepo) ListByTeam(ctx context.Context, teamBizKey int64, excludeCancelled bool) ([]model.Milestone, error) {
	query := r.db.WithContext(ctx).Scopes(NotDeleted).
		Where("team_key = ?", teamBizKey)
	if excludeCancelled {
		query = query.Where("milestone_status != ?", "cancelled")
	}
	var items []model.Milestone
	err := query.Order("expected_end_date ASC").Find(&items).Error
	return items, err
}

func (r *milestoneRepo) SoftDelete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Scopes(NotDeleted).
		Model(&model.Milestone{}).
		Where("id = ?", id).
		Update("deleted_flag", 1)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *milestoneRepo) DeleteByMap(ctx context.Context, milestoneMapBizKey int64) error {
	result := r.db.WithContext(ctx).
		Model(&model.Milestone{}).
		Where("milestone_map_key = ? AND deleted_flag = 0", milestoneMapBizKey).
		Update("deleted_flag", 1)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

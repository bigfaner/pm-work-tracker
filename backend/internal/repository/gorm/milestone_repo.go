package gorm

import (
	"context"
	"time"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
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
		return nil, err
	}
	return &m, nil
}

func (r *milestoneRepo) FindBatchByBizKeys(ctx context.Context, bizKeys []int64) (map[int64]*model.Milestone, error) {
	result := make(map[int64]*model.Milestone)
	if len(bizKeys) == 0 {
		return result, nil
	}
	var items []model.Milestone
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("biz_key IN ?", bizKeys).Find(&items).Error
	if err != nil {
		return nil, err
	}
	for i := range items {
		result[items[i].BizKey] = &items[i]
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

func (r *milestoneRepo) ListByTeam(ctx context.Context, teamBizKey int64, filter dto.MilestoneTeamFilter) ([]model.Milestone, error) {
	query := r.db.WithContext(ctx).Scopes(NotDeleted).Where("team_key = ?", teamBizKey)

	if filter.Name != nil && *filter.Name != "" {
		query = query.Where("milestone_name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("milestone_status = ?", *filter.Status)
	}
	//nolint:misspell // "cancelled" is a domain status value per PRD/API contract.
	if filter.ExcludeCancelled != nil && *filter.ExcludeCancelled {
		query = query.Where("milestone_status != ?", "cancelled")
	}

	var items []model.Milestone
	err := query.Order("expected_end_date ASC").Find(&items).Error
	return items, err
}

func (r *milestoneRepo) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Milestone{}).
		Where("id = ? AND deleted_flag = 0", id).
		Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error
}

func (r *milestoneRepo) SoftDeleteByMap(ctx context.Context, milestoneMapBizKey int64) error {
	return r.db.WithContext(ctx).Model(&model.Milestone{}).
		Where("milestone_map_key = ? AND deleted_flag = 0", milestoneMapBizKey).
		Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error
}

func (r *milestoneRepo) ExistsByNameAndMap(ctx context.Context, milestoneMapBizKey int64, name string, excludeID *uint) (bool, error) {
	query := r.db.WithContext(ctx).Model(&model.Milestone{}).
		Scopes(NotDeleted).
		Where("milestone_map_key = ? AND milestone_name = ?", milestoneMapBizKey, name)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

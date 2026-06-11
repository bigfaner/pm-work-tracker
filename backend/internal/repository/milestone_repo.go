package repository

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
)

// MilestoneRepo defines persistence operations for Milestone entities.
type MilestoneRepo interface {
	Create(ctx context.Context, m *model.Milestone) error
	FindByID(ctx context.Context, id uint) (*model.Milestone, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error)
	FindBatchByBizKeys(ctx context.Context, bizKeys []int64) (map[int64]*model.Milestone, error)
	Update(ctx context.Context, m *model.Milestone, fields map[string]interface{}) error
	ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error)
	ListByTeam(ctx context.Context, teamBizKey int64, filter dto.MilestoneTeamFilter) ([]model.Milestone, error)
	SoftDelete(ctx context.Context, id uint) error
	SoftDeleteByMap(ctx context.Context, milestoneMapBizKey int64) error
	ExistsByNameAndMap(ctx context.Context, milestoneMapBizKey int64, name string, excludeID *uint) (bool, error)
}

package repository

import (
	"context"

	"pm-work-tracker/backend/internal/model"
)

// MilestoneRepo defines persistence operations for Milestone entities.
type MilestoneRepo interface {
	Create(ctx context.Context, m *model.Milestone) error
	FindByID(ctx context.Context, id uint) (*model.Milestone, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.Milestone, error)
	FindByBizKeys(ctx context.Context, bizKeys []int64) (map[int64]*model.Milestone, error)
	Update(ctx context.Context, m *model.Milestone, fields map[string]interface{}) error
	ListByMap(ctx context.Context, milestoneMapBizKey int64) ([]model.Milestone, error)
	ListByTeam(ctx context.Context, teamBizKey int64, excludeCancelled bool) ([]model.Milestone, error)
	SoftDelete(ctx context.Context, id uint) error
	DeleteByMap(ctx context.Context, milestoneMapBizKey int64) error
}

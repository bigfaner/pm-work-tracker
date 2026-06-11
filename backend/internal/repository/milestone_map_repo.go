package repository

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
)

// MilestoneMapRepo defines persistence operations for MilestoneMap entities.
type MilestoneMapRepo interface {
	Create(ctx context.Context, m *model.MilestoneMap) error
	FindByID(ctx context.Context, id uint) (*model.MilestoneMap, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.MilestoneMap, error)
	Update(ctx context.Context, m *model.MilestoneMap, fields map[string]interface{}) error
	List(ctx context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error)
	SoftDelete(ctx context.Context, id uint) error
}

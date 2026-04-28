package repository

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
)

// MainItemRepo defines persistence operations for MainItem entities.
type MainItemRepo interface {
	Create(ctx context.Context, item *model.MainItem) error
	FindByID(ctx context.Context, id uint) (*model.MainItem, error)
	FindByIDs(ctx context.Context, ids []uint) (map[uint]*model.MainItem, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error)
	FindByBizKeys(ctx context.Context, bizKeys []int64) (map[int64]*model.MainItem, error)
	Update(ctx context.Context, item *model.MainItem, fields map[string]interface{}) error
	List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error)
	ListByTeamAndStatus(ctx context.Context, teamID uint, status string) ([]model.MainItem, error)
	NextCode(ctx context.Context, teamBizKey int64) (string, error)
	CountByTeam(ctx context.Context, teamID uint) (int64, error)
	ListNonArchivedByTeam(ctx context.Context, teamID uint) ([]model.MainItem, error)
}

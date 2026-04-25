package repository

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
)

// ItemPoolRepo defines persistence operations for ItemPool entities.
type ItemPoolRepo interface {
	Create(ctx context.Context, item *model.ItemPool) error
	FindByID(ctx context.Context, id uint) (*model.ItemPool, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.ItemPool, error)
	Update(ctx context.Context, item *model.ItemPool, fields map[string]interface{}) error
	List(ctx context.Context, teamID uint, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error)
}

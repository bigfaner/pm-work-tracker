package repository

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
)

// SubItemRepo defines persistence operations for SubItem entities.
type SubItemRepo interface {
	Create(ctx context.Context, item *model.SubItem) error
	FindByID(ctx context.Context, id uint) (*model.SubItem, error)
	Update(ctx context.Context, item *model.SubItem, fields map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, teamID uint, mainItemID uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error)
	ListByMainItem(ctx context.Context, mainItemID uint) ([]*model.SubItem, error)
	ListByTeam(ctx context.Context, teamID uint) ([]model.SubItem, error)
	NextSubCode(ctx context.Context, mainItemID uint) (string, error)
}

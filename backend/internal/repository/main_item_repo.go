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
	Update(ctx context.Context, item *model.MainItem, fields map[string]interface{}) error
	List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error)
	NextCode(ctx context.Context, teamID uint) (string, error)
	CountByTeam(ctx context.Context, teamID uint) (int64, error)
}

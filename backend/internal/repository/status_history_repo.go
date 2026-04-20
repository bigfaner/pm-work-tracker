package repository

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
)

// StatusHistoryRepo defines persistence operations for StatusHistory entities.
type StatusHistoryRepo interface {
	Create(ctx context.Context, record *model.StatusHistory) error
	FindByID(ctx context.Context, id uint) (*model.StatusHistory, error)
	ListByItem(ctx context.Context, itemType string, itemID uint, page dto.Pagination) (*dto.PageResult[model.StatusHistory], error)
}

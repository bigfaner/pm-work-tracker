package repository

import (
	"context"

	"pm-work-tracker/backend/internal/model"
)

// DecisionLogRepo defines persistence operations for DecisionLog entities.
// Records are append-only after publishing: Update only allowed for drafts.
type DecisionLogRepo interface {
	Create(ctx context.Context, log *model.DecisionLog) error
	FindByID(ctx context.Context, id uint) (*model.DecisionLog, error)
	FindByBizKey(ctx context.Context, bizKey int64) (*model.DecisionLog, error)
	ListByItem(ctx context.Context, mainItemKey int64, userID uint, offset, limit int) ([]model.DecisionLog, int64, error)
	Update(ctx context.Context, log *model.DecisionLog) error
}

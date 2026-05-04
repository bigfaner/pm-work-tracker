package service

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
)

// DecisionLogService defines business operations for decision logs.
type DecisionLogService interface {
	Create(ctx context.Context, mainItemID uint, userID uint, req dto.DecisionLogCreateReq) (*model.DecisionLog, error)
	Update(ctx context.Context, bizKey int64, userID uint, req dto.DecisionLogUpdateReq) (*model.DecisionLog, error)
	Publish(ctx context.Context, bizKey int64, userID uint) (*model.DecisionLog, error)
	List(ctx context.Context, mainItemID uint, userID uint, page dto.Pagination) (*dto.PageResult[model.DecisionLog], error)
}

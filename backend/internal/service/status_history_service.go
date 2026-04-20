package service

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/repository"
)

// StatusHistoryService defines operations for recording and querying status change history.
type StatusHistoryService interface {
	Record(ctx context.Context, record *model.StatusHistory) error
	ListByItem(ctx context.Context, itemType string, itemID uint, page dto.Pagination) (*dto.PageResult[model.StatusHistory], error)
}

type statusHistoryService struct {
	repo repository.StatusHistoryRepo
}

// NewStatusHistoryService creates a new StatusHistoryService.
func NewStatusHistoryService(repo repository.StatusHistoryRepo) StatusHistoryService {
	if repo == nil {
		panic("status_history_service: repo must not be nil")
	}
	return &statusHistoryService{repo: repo}
}

func (s *statusHistoryService) Record(ctx context.Context, record *model.StatusHistory) error {
	return s.repo.Create(ctx, record)
}

func (s *statusHistoryService) ListByItem(ctx context.Context, itemType string, itemID uint, page dto.Pagination) (*dto.PageResult[model.StatusHistory], error) {
	return s.repo.ListByItem(ctx, itemType, itemID, page)
}

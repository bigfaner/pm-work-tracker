package service

import (
	"context"
	"encoding/json"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/repository"
)

// DecisionLogService defines business operations for decision logs.
type DecisionLogService interface {
	Create(ctx context.Context, mainItemKey int64, userID uint, req dto.DecisionLogCreateReq) (*model.DecisionLog, error)
	Update(ctx context.Context, bizKey int64, userID uint, req dto.DecisionLogUpdateReq) (*model.DecisionLog, error)
	Publish(ctx context.Context, bizKey int64, userID uint) (*model.DecisionLog, error)
	List(ctx context.Context, mainItemKey int64, userID uint, page dto.Pagination) (*dto.PageResult[model.DecisionLog], error)
}

type decisionLogService struct {
	repo         repository.DecisionLogRepo
	mainItemRepo repository.MainItemRepo
}

// NewDecisionLogService creates a new DecisionLogService.
func NewDecisionLogService(repo repository.DecisionLogRepo, mainItemRepo repository.MainItemRepo) DecisionLogService {
	return &decisionLogService{repo: repo, mainItemRepo: mainItemRepo}
}

func (s *decisionLogService) Create(ctx context.Context, mainItemKey int64, userID uint, req dto.DecisionLogCreateReq) (*model.DecisionLog, error) {
	// Validate main item exists and get TeamKey
	mainItem, err := s.mainItemRepo.FindByBizKey(ctx, mainItemKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}

	// Serialize tags to JSON string
	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return nil, apperrors.ErrValidation
	}

	log := &model.DecisionLog{
		BizKey:      snowflake.Generate(),
		MainItemKey: mainItemKey,
		TeamKey:     mainItem.TeamKey,
		Category:    req.Category,
		Tags:        string(tagsJSON),
		Content:     req.Content,
		LogStatus:   req.LogStatus,
		CreatedBy:   int64(userID),
	}

	if err := s.repo.Create(ctx, log); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *decisionLogService) Update(ctx context.Context, bizKey int64, userID uint, req dto.DecisionLogUpdateReq) (*model.DecisionLog, error) {
	log, err := s.repo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrDecisionLogNotFound)
	}

	// Enforce draft-only editing
	if log.LogStatus != "draft" {
		return nil, apperrors.ErrForbidden
	}

	// Enforce owner-only editing
	if log.CreatedBy != int64(userID) {
		return nil, apperrors.ErrForbidden
	}

	// Serialize tags
	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return nil, apperrors.ErrValidation
	}

	log.Category = req.Category
	log.Tags = string(tagsJSON)
	log.Content = req.Content

	if err := s.repo.Update(ctx, log); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *decisionLogService) Publish(ctx context.Context, bizKey int64, userID uint) (*model.DecisionLog, error) {
	log, err := s.repo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrDecisionLogNotFound)
	}

	// Enforce draft-only
	if log.LogStatus != "draft" {
		return nil, apperrors.ErrForbidden
	}

	// Enforce owner-only
	if log.CreatedBy != int64(userID) {
		return nil, apperrors.ErrForbidden
	}

	log.LogStatus = "published"

	if err := s.repo.Update(ctx, log); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *decisionLogService) List(ctx context.Context, mainItemKey int64, userID uint, page dto.Pagination) (*dto.PageResult[model.DecisionLog], error) {
	offset, pageNum, pageSize := dto.ApplyPaginationDefaults(page.Page, page.PageSize)

	logs, total, err := s.repo.ListByItem(ctx, mainItemKey, userID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &dto.PageResult[model.DecisionLog]{
		Items: logs,
		Total: total,
		Page:  pageNum,
		Size:  pageSize,
	}, nil
}

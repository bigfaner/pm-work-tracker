package service

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"

	stderrors "errors"
)

// ItemPoolService defines business operations for ItemPool.
type ItemPoolService interface {
	Submit(ctx context.Context, teamID, submitterID uint, req dto.SubmitItemPoolReq) (*model.ItemPool, error)
	Assign(ctx context.Context, teamID, pmID, poolItemID uint, req dto.AssignItemPoolReq) error
	Reject(ctx context.Context, teamID, pmID, poolItemID uint, reason string) error
	List(ctx context.Context, teamID uint, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error)
	Get(ctx context.Context, teamID, poolItemID uint) (*model.ItemPool, error)
}

type dbTransactor interface {
	Transaction(fc func(tx *gorm.DB) error) error
}

type itemPoolService struct {
	poolRepo   repository.ItemPoolRepo
	subRepo    repository.SubItemRepo
	mainRepo   repository.MainItemRepo
	db         dbTransactor
}

// NewItemPoolService creates a new ItemPoolService.
func NewItemPoolService(poolRepo repository.ItemPoolRepo, subRepo repository.SubItemRepo, mainRepo repository.MainItemRepo, db dbTransactor) ItemPoolService {
	return &itemPoolService{
		poolRepo: poolRepo,
		subRepo:  subRepo,
		mainRepo: mainRepo,
		db:       db,
	}
}

func (s *itemPoolService) Submit(ctx context.Context, teamID, submitterID uint, req dto.SubmitItemPoolReq) (*model.ItemPool, error) {
	item := &model.ItemPool{
		TeamID:         teamID,
		Title:          req.Title,
		Background:     req.Background,
		ExpectedOutput: req.ExpectedOutput,
		SubmitterID:    submitterID,
		Status:         "待分配",
	}
	if err := s.poolRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *itemPoolService) Assign(ctx context.Context, teamID, pmID, poolItemID uint, req dto.AssignItemPoolReq) error {
	poolItem, err := s.poolRepo.FindByID(ctx, poolItemID)
	if err != nil {
		return mapPoolItemNotFound(err)
	}
	if poolItem.TeamID != teamID {
		return apperrors.ErrForbidden
	}
	if poolItem.Status != "待分配" {
		return apperrors.ErrItemAlreadyProcessed
	}

	// Validate main item exists
	mainItem, err := s.mainRepo.FindByID(ctx, req.MainItemID)
	if err != nil {
		return mapMainItemNotFound(err)
	}
	if mainItem.TeamID != teamID {
		return apperrors.ErrItemNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// Create SubItem under the MainItem
		subItem := &model.SubItem{
			TeamID:     teamID,
			MainItemID: req.MainItemID,
			Title:      poolItem.Title,
			Description: poolItem.Background,
			Priority:   "P2",
			AssigneeID: &req.AssigneeID,
			Status:     "待开始",
			Weight:     1.0,
		}
		if err := s.subRepo.Create(ctx, subItem); err != nil {
			return err
		}

		// Update pool item
		fields := map[string]interface{}{
			"status":           "已分配",
			"assigned_main_id": req.MainItemID,
			"assigned_sub_id":  subItem.ID,
			"assignee_id":      req.AssigneeID,
			"reviewer_id":      pmID,
			"reviewed_at":      now,
		}
		return s.poolRepo.Update(ctx, poolItem, fields)
	})
}

func (s *itemPoolService) Reject(ctx context.Context, teamID, pmID, poolItemID uint, reason string) error {
	poolItem, err := s.poolRepo.FindByID(ctx, poolItemID)
	if err != nil {
		return mapPoolItemNotFound(err)
	}
	if poolItem.TeamID != teamID {
		return apperrors.ErrForbidden
	}
	if poolItem.Status != "待分配" {
		return apperrors.ErrItemAlreadyProcessed
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":        "已拒绝",
		"reject_reason": reason,
		"reviewer_id":   pmID,
		"reviewed_at":   now,
	}
	return s.poolRepo.Update(ctx, poolItem, fields)
}

func (s *itemPoolService) List(ctx context.Context, teamID uint, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error) {
	return s.poolRepo.List(ctx, teamID, filter, page)
}

func (s *itemPoolService) Get(ctx context.Context, teamID, poolItemID uint) (*model.ItemPool, error) {
	item, err := s.poolRepo.FindByID(ctx, poolItemID)
	if err != nil {
		return nil, mapPoolItemNotFound(err)
	}
	if item.TeamID != teamID {
		return nil, apperrors.ErrForbidden
	}
	return item, nil
}

func mapPoolItemNotFound(err error) error {
	if err == gorm.ErrRecordNotFound || stderrors.Is(err, apperrors.ErrNotFound) {
		return apperrors.ErrItemNotFound
	}
	return err
}

func mapMainItemNotFound(err error) error {
	if err == gorm.ErrRecordNotFound || stderrors.Is(err, apperrors.ErrNotFound) {
		return apperrors.ErrItemNotFound
	}
	return err
}

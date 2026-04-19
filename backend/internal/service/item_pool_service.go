package service

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/dates"
	"pm-work-tracker/backend/internal/repository"
)

// ItemPoolService defines business operations for ItemPool.
type ItemPoolService interface {
	Submit(ctx context.Context, teamID, submitterID uint, req dto.SubmitItemPoolReq) (*model.ItemPool, error)
	Assign(ctx context.Context, teamID, pmID, poolItemID uint, req dto.AssignItemPoolReq) error
	ConvertToMain(ctx context.Context, teamID, pmID, poolItemID uint, req dto.ConvertToMainItemReq) (*model.MainItem, error)
	Reject(ctx context.Context, teamID, pmID, poolItemID uint, reason string) error
	List(ctx context.Context, teamID uint, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error)
	Get(ctx context.Context, teamID, poolItemID uint) (*model.ItemPool, error)
}

type dbTransactor interface {
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
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
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
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
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if mainItem.TeamID != teamID {
		return apperrors.ErrItemNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// Create SubItem under the MainItem
		subItem := &model.SubItem{
			TeamID:      teamID,
			MainItemID:  req.MainItemID,
			Title:       poolItem.Title,
			Description: poolItem.Background,
			Priority:    defaultPriority(req.Priority),
			AssigneeID:  req.AssigneeID,
			Status:      "待开始",
			Weight:      1.0,
		}
		if req.StartDate != nil {
			if t, e := dates.ParseDate( *req.StartDate); e == nil {
				subItem.StartDate = &t
			}
		}
		if req.ExpectedEndDate != nil {
			if t, e := dates.ParseDate( *req.ExpectedEndDate); e == nil {
				subItem.ExpectedEndDate = &t
			}
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

func (s *itemPoolService) ConvertToMain(ctx context.Context, teamID, pmID, poolItemID uint, req dto.ConvertToMainItemReq) (*model.MainItem, error) {
	poolItem, err := s.poolRepo.FindByID(ctx, poolItemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if poolItem.TeamID != teamID {
		return nil, apperrors.ErrForbidden
	}
	if poolItem.Status != "待分配" {
		return nil, apperrors.ErrItemAlreadyProcessed
	}

	var created *model.MainItem
	err = s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		code, err := s.mainRepo.NextCode(ctx, teamID)
		if err != nil {
			return err
		}

		mainItem := &model.MainItem{
			TeamID:      teamID,
			Code:        code,
			Title:       poolItem.Title,
			Priority:    req.Priority,
			ProposerID:  pmID,
			AssigneeID:  req.AssigneeID,
			IsKeyItem:   false,
			Status:      "待开始",
		}

		if req.StartDate != nil {
			if t, e := dates.ParseDate( *req.StartDate); e == nil {
				mainItem.StartDate = &t
			}
		}
		if req.ExpectedEndDate != nil {
			if t, e := dates.ParseDate( *req.ExpectedEndDate); e == nil {
				mainItem.ExpectedEndDate = &t
			}
		}

		if err := s.mainRepo.Create(ctx, mainItem); err != nil {
			return err
		}

		fields := map[string]interface{}{
			"status":           "已分配",
			"assigned_main_id": mainItem.ID,
			"assignee_id":      req.AssigneeID,
			"reviewer_id":      pmID,
			"reviewed_at":      now,
		}
		if err := s.poolRepo.Update(ctx, poolItem, fields); err != nil {
			return err
		}

		created = mainItem
		return nil
	})
	return created, err
}

func (s *itemPoolService) Reject(ctx context.Context, teamID, pmID, poolItemID uint, reason string) error {
	poolItem, err := s.poolRepo.FindByID(ctx, poolItemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
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
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamID != teamID {
		return nil, apperrors.ErrForbidden
	}
	return item, nil
}

func defaultPriority(p string) string {
	if p == "" {
		return "P2"
	}
	return p
}


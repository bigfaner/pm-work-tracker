package service

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/pkg/dates"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/repository"
)

// ItemPoolService defines business operations for ItemPool.
type ItemPoolService interface {
	Submit(ctx context.Context, teamBizKey int64, submitterBizKey int64, req dto.SubmitItemPoolReq) (*model.ItemPool, error)
	Assign(ctx context.Context, teamBizKey int64, pmBizKey int64, poolItemID uint, req dto.AssignItemPoolReq) error
	ConvertToMain(ctx context.Context, teamBizKey int64, pmBizKey int64, poolItemID uint, req dto.ConvertToMainItemReq) (*model.MainItem, error)
	Reject(ctx context.Context, teamBizKey int64, pmBizKey int64, poolItemID uint, reason string) error
	List(ctx context.Context, teamBizKey int64, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error)
	Get(ctx context.Context, teamBizKey int64, poolItemID uint) (*model.ItemPool, error)
	Update(ctx context.Context, teamBizKey int64, poolItemID uint, req dto.UpdateItemPoolReq) (*model.ItemPool, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.ItemPool, error)
}

type itemPoolService struct {
	poolRepo   repository.ItemPoolRepo
	subRepo    repository.SubItemRepo
	mainRepo   repository.MainItemRepo
	db         repo.DBTransactor
}

// NewItemPoolService creates a new ItemPoolService.
func NewItemPoolService(poolRepo repository.ItemPoolRepo, subRepo repository.SubItemRepo, mainRepo repository.MainItemRepo, db repo.DBTransactor) ItemPoolService {
	return &itemPoolService{
		poolRepo: poolRepo,
		subRepo:  subRepo,
		mainRepo: mainRepo,
		db:       db,
	}
}

func (s *itemPoolService) Submit(ctx context.Context, teamBizKey int64, submitterBizKey int64, req dto.SubmitItemPoolReq) (*model.ItemPool, error) {
	item := &model.ItemPool{
		BaseModel:      model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:        teamBizKey,
		Title:          req.Title,
		Background:     req.Background,
		ExpectedOutput: req.ExpectedOutput,
		SubmitterKey:   submitterBizKey,
		PoolStatus:     "pending",
	}
	if err := s.poolRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *itemPoolService) Assign(ctx context.Context, teamBizKey int64, pmBizKey int64, poolItemID uint, req dto.AssignItemPoolReq) error {
	poolItem, err := s.poolRepo.FindByID(ctx, poolItemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if poolItem.TeamKey != teamBizKey {
		return apperrors.ErrForbidden
	}
	if poolItem.PoolStatus != "pending" {
		return apperrors.ErrItemAlreadyProcessed
	}

	// Validate main item exists
	mainBizKey, _ := pkg.ParseID(req.MainItemKey)
	mainItem, err := s.mainRepo.FindByBizKey(ctx, mainBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if mainItem.TeamKey != teamBizKey {
		return apperrors.ErrItemNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// Create SubItem under the MainItem
		subItem := &model.SubItem{
			BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
			TeamKey:     teamBizKey,
			MainItemKey: mainItem.BizKey,
			Title:       poolItem.Title,
			ItemDesc:    poolItem.Background,
			Priority:    defaultPriority(req.Priority),
			AssigneeKey: func() *int64 { if req.AssigneeKey != nil { v, _ := pkg.ParseID(*req.AssigneeKey); return &v }; return nil }(),
			ItemStatus:  "pending",
			Weight:      1.0,
		}
		if req.StartDate != nil {
			if t, e := dates.ParseDate(*req.StartDate); e == nil {
				subItem.PlanStartDate = &t
			}
		}
		if req.ExpectedEndDate != nil {
			if t, e := dates.ParseDate(*req.ExpectedEndDate); e == nil {
				subItem.ExpectedEndDate = &t
			}
		}
		if err := s.subRepo.Create(ctx, subItem); err != nil {
			return err
		}

		// Update pool item
		fields := map[string]interface{}{
			"pool_status":      "assigned",
			"assigned_main_key": mainItem.BizKey,
			"assigned_sub_key":  subItem.BizKey,
			"assignee_key":      func() *int64 { if req.AssigneeKey != nil { v, _ := pkg.ParseID(*req.AssigneeKey); return &v }; return nil }(),
			"reviewer_key":      pmBizKey,
			"reviewed_at":       now,
		}
		return s.poolRepo.Update(ctx, poolItem, fields)
	})
}

func (s *itemPoolService) ConvertToMain(ctx context.Context, teamBizKey int64, pmBizKey int64, poolItemID uint, req dto.ConvertToMainItemReq) (*model.MainItem, error) {
	poolItem, err := s.poolRepo.FindByID(ctx, poolItemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if poolItem.TeamKey != teamBizKey {
		return nil, apperrors.ErrForbidden
	}
	if poolItem.PoolStatus != "pending" {
		return nil, apperrors.ErrItemAlreadyProcessed
	}

	var created *model.MainItem
	err = s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		code, err := s.mainRepo.NextCode(ctx, teamBizKey)
		if err != nil {
			return err
		}

		mainItem := &model.MainItem{
			BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
			TeamKey:     teamBizKey,
			Code:        code,
			Title:       poolItem.Title,
			Priority:    req.Priority,
			ProposerKey: pmBizKey,
			AssigneeKey: func() *int64 { if req.AssigneeKey != nil { v, _ := pkg.ParseID(*req.AssigneeKey); return &v }; return nil }(),
			IsKeyItem:   false,
			ItemStatus:  "pending",
		}

		if req.StartDate != nil {
			if t, e := dates.ParseDate(*req.StartDate); e == nil {
				mainItem.PlanStartDate = &t
			}
		}
		if req.ExpectedEndDate != nil {
			if t, e := dates.ParseDate(*req.ExpectedEndDate); e == nil {
				mainItem.ExpectedEndDate = &t
			}
		}

		if err := s.mainRepo.Create(ctx, mainItem); err != nil {
			return err
		}

		fields := map[string]interface{}{
			"pool_status":       "assigned",
			"assigned_main_key": mainItem.BizKey,
			"assignee_key":      func() *int64 { if req.AssigneeKey != nil { v, _ := pkg.ParseID(*req.AssigneeKey); return &v }; return nil }(),
			"reviewer_key":      pmBizKey,
			"reviewed_at":       now,
		}
		if err := s.poolRepo.Update(ctx, poolItem, fields); err != nil {
			return err
		}

		created = mainItem
		return nil
	})
	return created, err
}

func (s *itemPoolService) Reject(ctx context.Context, teamBizKey int64, pmBizKey int64, poolItemID uint, reason string) error {
	poolItem, err := s.poolRepo.FindByID(ctx, poolItemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if poolItem.TeamKey != teamBizKey {
		return apperrors.ErrForbidden
	}
	if poolItem.PoolStatus != "pending" {
		return apperrors.ErrItemAlreadyProcessed
	}

	now := time.Now()
	fields := map[string]interface{}{
		"pool_status":   "rejected",
		"reject_reason": reason,
		"reviewer_key":  pmBizKey,
		"reviewed_at":   now,
	}
	return s.poolRepo.Update(ctx, poolItem, fields)
}

func (s *itemPoolService) List(ctx context.Context, teamBizKey int64, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error) {
	return s.poolRepo.List(ctx, teamBizKey, filter, page)
}

func (s *itemPoolService) Update(ctx context.Context, teamBizKey int64, poolItemID uint, req dto.UpdateItemPoolReq) (*model.ItemPool, error) {
	item, err := s.poolRepo.FindByID(ctx, poolItemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamKey != teamBizKey {
		return nil, apperrors.ErrForbidden
	}
	if item.PoolStatus != "pending" {
		return nil, apperrors.ErrItemAlreadyProcessed
	}

	fields := map[string]interface{}{}
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Background != nil {
		fields["background"] = *req.Background
	}
	if req.ExpectedOutput != nil {
		fields["expected_output"] = *req.ExpectedOutput
	}
	if len(fields) == 0 {
		return item, nil
	}
	if err := s.poolRepo.Update(ctx, item, fields); err != nil {
		return nil, err
	}
	return s.poolRepo.FindByID(ctx, poolItemID)
}

func (s *itemPoolService) Get(ctx context.Context, teamBizKey int64, poolItemID uint) (*model.ItemPool, error) {
	item, err := s.poolRepo.FindByID(ctx, poolItemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamKey != teamBizKey {
		return nil, apperrors.ErrForbidden
	}
	return item, nil
}

func (s *itemPoolService) GetByBizKey(ctx context.Context, bizKey int64) (*model.ItemPool, error) {
	item, err := s.poolRepo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	return item, nil
}

func defaultPriority(p string) string {
	if p == "" {
		return "P2"
	}
	return p
}


package service

import (
	"context"
	"time"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/dates"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/pkg/status"
	"pm-work-tracker/backend/internal/repository"
)

// SubItemChangeResult holds the result of a sub-item status change.
type SubItemChangeResult struct {
	SubItem       *model.SubItem
	LinkageResult *LinkageResult
}

// SubItemService defines business operations for SubItem.
type SubItemService interface {
	Create(ctx context.Context, teamID, callerID uint, req dto.SubItemCreateReq) (*model.SubItem, error)
	Update(ctx context.Context, teamID, itemID uint, req dto.SubItemUpdateReq) error
	ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) (*SubItemChangeResult, error)
	Delete(ctx context.Context, teamID, callerID, itemID uint) error
	Get(ctx context.Context, teamID, itemID uint) (*model.SubItem, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.SubItem, error)
	List(ctx context.Context, teamID uint, mainItemID *uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error)
	Assign(ctx context.Context, teamID, pmID, itemID, assigneeID uint) error
	AvailableTransitions(ctx context.Context, teamID, subID uint) ([]string, error)
}

type subItemService struct {
	subItemRepo      repository.SubItemRepo
	mainItemSvc      MainItemService
	statusHistorySvc StatusHistoryService
}

// NewSubItemService creates a new SubItemService.
func NewSubItemService(subItemRepo repository.SubItemRepo, mainItemSvc MainItemService, statusHistorySvc StatusHistoryService) SubItemService {
	return &subItemService{subItemRepo: subItemRepo, mainItemSvc: mainItemSvc, statusHistorySvc: statusHistorySvc}
}

func (s *subItemService) Create(ctx context.Context, teamID, callerID uint, req dto.SubItemCreateReq) (*model.SubItem, error) {
	mainBizKey, _ := pkg.ParseID(req.MainItemKey)
	mainItem, err := s.mainItemSvc.GetByBizKey(ctx, mainBizKey)
	if err != nil {
		return nil, err
	}
	if status.IsMainTerminal(mainItem.ItemStatus) {
		return nil, apperrors.ErrTerminalMainItem
	}

	code, err := s.subItemRepo.NextSubCode(ctx, mainItem.ID)
	if err != nil {
		return nil, err
	}

	item := &model.SubItem{
		BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:     int64(teamID),
		MainItemKey: int64(mainItem.ID),
		Code:        code,
		Title:       req.Title,
		ItemDesc:    req.Description,
		Priority:    req.Priority,
		AssigneeKey: func() *int64 { if req.AssigneeKey != "" { v, _ := pkg.ParseID(req.AssigneeKey); return &v }; return nil }(),
		ItemStatus:  "pending",
		Weight:      1.0,
	}

	if req.StartDate != nil {
		if t, err := dates.ParseDate(*req.StartDate); err == nil {
			item.PlanStartDate = &t
		}
	}
	if req.ExpectedEndDate != nil {
		if t, err := dates.ParseDate(*req.ExpectedEndDate); err == nil {
			item.ExpectedEndDate = &t
		}
	}

	if err := s.subItemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	// Trigger linkage evaluation after creating a new sub-item
	_, _ = s.mainItemSvc.EvaluateLinkage(ctx, mainItem.ID, callerID)

	return item, nil
}

func (s *subItemService) Update(ctx context.Context, teamID, itemID uint, req dto.SubItemUpdateReq) error {
	item, err := s.subItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamKey != int64(teamID) {
		return apperrors.ErrForbidden
	}

	fields := map[string]interface{}{}
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Description != nil {
		fields["item_desc"] = *req.Description
	}
	if req.Priority != nil {
		fields["priority"] = *req.Priority
	}
	if req.AssigneeKey != nil {
		assigneeKey, err := pkg.ParseIDPtr(req.AssigneeKey)
		if err != nil {
			return apperrors.ErrValidation
		}
		fields["assignee_key"] = assigneeKey
	}
	if req.StartDate != nil {
		fields["plan_start_date"] = *req.StartDate
	}
	if req.ExpectedEndDate != nil {
		fields["expected_end_date"] = *req.ExpectedEndDate
	}

	if len(fields) == 0 {
		return nil
	}

	return s.subItemRepo.Update(ctx, item, fields)
}

func (s *subItemService) ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) (*SubItemChangeResult, error) {
	item, err := s.subItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamKey != int64(teamID) {
		return nil, apperrors.ErrForbidden
	}

	if !status.IsValidTransition(status.SubItemTransitions, item.ItemStatus, newStatus) {
		return nil, apperrors.ErrInvalidStatus
	}

	fields := map[string]interface{}{
		"item_status": newStatus,
	}

	// Terminal side effects: force completion=100 and set actual_end_date
	if newStatus == "completed" || newStatus == "closed" {
		fields["completion_pct"] = float64(100)
		now := time.Now()
		fields["actual_end_date"] = &now
	}

	// Capture old status before update (repo may mutate the item)
	oldStatus := item.ItemStatus

	if err := s.subItemRepo.Update(ctx, item, fields); err != nil {
		return nil, err
	}

	// After terminal transition, recalculate parent MainItem completion
	if newStatus == "completed" || newStatus == "closed" {
		if err := s.mainItemSvc.RecalcCompletion(ctx, uint(item.MainItemKey)); err != nil {
			return nil, err
		}
	}

	// Record to status history
	if err := RecordStatusChange(s.statusHistorySvc, ctx, "sub_item", int64(itemID), oldStatus, newStatus, callerID, 0, ""); err != nil {
		return nil, err
	}

	// Evaluate linkage after status change
	linkageResult, _ := s.mainItemSvc.EvaluateLinkage(ctx, uint(item.MainItemKey), callerID)

	// Fetch updated item
	updated, err := s.subItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	return &SubItemChangeResult{SubItem: updated, LinkageResult: linkageResult}, nil
}

func (s *subItemService) Delete(ctx context.Context, teamID, callerID, itemID uint) error {
	item, err := s.subItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamKey != int64(teamID) {
		return apperrors.ErrForbidden
	}

	if err := s.subItemRepo.SoftDelete(ctx, itemID); err != nil {
		return err
	}

	// Trigger linkage evaluation after deleting a sub-item
	_, _ = s.mainItemSvc.EvaluateLinkage(ctx, uint(item.MainItemKey), callerID)

	return nil
}

func (s *subItemService) AvailableTransitions(ctx context.Context, teamID, subID uint) ([]string, error) {
	item, err := s.subItemRepo.FindByID(ctx, subID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamKey != int64(teamID) {
		return nil, apperrors.ErrForbidden
	}
	return status.GetAvailableTransitions(status.SubItemTransitions, item.ItemStatus), nil
}

func (s *subItemService) Get(ctx context.Context, teamID, itemID uint) (*model.SubItem, error) {
	item, err := s.subItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	return item, nil
}

func (s *subItemService) GetByBizKey(ctx context.Context, bizKey int64) (*model.SubItem, error) {
	item, err := s.subItemRepo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	return item, nil
}

func (s *subItemService) List(ctx context.Context, teamID uint, mainItemID *uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	mid := uint(0)
	if mainItemID != nil {
		mid = *mainItemID
	}
	return s.subItemRepo.List(ctx, teamID, mid, filter, page)
}

func (s *subItemService) Assign(ctx context.Context, teamID, pmID, itemID, assigneeID uint) error {
	item, err := s.subItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamKey != int64(teamID) {
		return apperrors.ErrForbidden
	}

	return s.subItemRepo.Update(ctx, item, map[string]interface{}{
		"assignee_key": assigneeID,
	})
}

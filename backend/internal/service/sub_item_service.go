package service

import (
	"context"
	"time"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/dates"
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
	mainItem, err := s.mainItemSvc.Get(ctx, req.MainItemID)
	if err != nil {
		return nil, err
	}
	if status.IsMainTerminal(mainItem.Status) {
		return nil, apperrors.ErrTerminalMainItem
	}

	code, err := s.subItemRepo.NextSubCode(ctx, req.MainItemID)
	if err != nil {
		return nil, err
	}

	item := &model.SubItem{
		TeamID:      teamID,
		MainItemID:  req.MainItemID,
		Code:        code,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		AssigneeID:  &req.AssigneeID,
		Status:      "pending",
		Weight:      1.0,
	}

	if req.StartDate != nil {
		if t, err := dates.ParseDate(*req.StartDate); err == nil {
			item.StartDate = &t
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
	_, _ = s.mainItemSvc.EvaluateLinkage(ctx, item.MainItemID, callerID)

	return item, nil
}

func (s *subItemService) Update(ctx context.Context, teamID, itemID uint, req dto.SubItemUpdateReq) error {
	item, err := s.subItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamID != teamID {
		return apperrors.ErrForbidden
	}

	fields := map[string]interface{}{}
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Priority != nil {
		fields["priority"] = *req.Priority
	}
	if req.AssigneeID != nil {
		fields["assignee_id"] = *req.AssigneeID
	}
	if req.StartDate != nil {
		fields["start_date"] = *req.StartDate
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
	if item.TeamID != teamID {
		return nil, apperrors.ErrForbidden
	}

	if !status.IsValidTransition(status.SubItemTransitions, item.Status, newStatus) {
		return nil, apperrors.ErrInvalidStatus
	}

	fields := map[string]interface{}{
		"status": newStatus,
	}

	// Terminal side effects: force completion=100 and set actual_end_date
	if newStatus == "completed" || newStatus == "closed" {
		fields["completion"] = float64(100)
		now := time.Now()
		fields["actual_end_date"] = &now
	}

	// Capture old status before update (repo may mutate the item)
	oldStatus := item.Status

	if err := s.subItemRepo.Update(ctx, item, fields); err != nil {
		return nil, err
	}

	// After terminal transition, recalculate parent MainItem completion
	if newStatus == "completed" || newStatus == "closed" {
		if err := s.mainItemSvc.RecalcCompletion(ctx, item.MainItemID); err != nil {
			return nil, err
		}
	}

	// Record to status history
	if s.statusHistorySvc != nil {
		_ = s.statusHistorySvc.Record(ctx, &model.StatusHistory{
			ItemType:   "sub_item",
			ItemID:     itemID,
			FromStatus: oldStatus,
			ToStatus:   newStatus,
			ChangedBy:  callerID,
			IsAuto:     false,
		})
	}

	// Evaluate linkage after status change
	linkageResult, _ := s.mainItemSvc.EvaluateLinkage(ctx, item.MainItemID, callerID)

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
	if item.TeamID != teamID {
		return apperrors.ErrForbidden
	}

	if err := s.subItemRepo.Delete(ctx, itemID); err != nil {
		return err
	}

	// Trigger linkage evaluation after deleting a sub-item
	_, _ = s.mainItemSvc.EvaluateLinkage(ctx, item.MainItemID, callerID)

	return nil
}

func (s *subItemService) AvailableTransitions(ctx context.Context, teamID, subID uint) ([]string, error) {
	item, err := s.subItemRepo.FindByID(ctx, subID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamID != teamID {
		return nil, apperrors.ErrForbidden
	}
	return status.GetAvailableTransitions(status.SubItemTransitions, item.Status), nil
}

func (s *subItemService) Get(ctx context.Context, teamID, itemID uint) (*model.SubItem, error) {
	item, err := s.subItemRepo.FindByID(ctx, itemID)
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
	if item.TeamID != teamID {
		return apperrors.ErrForbidden
	}

	return s.subItemRepo.Update(ctx, item, map[string]interface{}{
		"assignee_id": assigneeID,
	})
}

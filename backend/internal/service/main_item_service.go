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

// MainItemService defines business operations for MainItem.
type MainItemService interface {
	Create(ctx context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error)
	Update(ctx context.Context, teamID, itemID uint, req dto.MainItemUpdateReq) error
	Archive(ctx context.Context, teamID, itemID uint) error
	List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error)
	Get(ctx context.Context, itemID uint) (*model.MainItem, error)
	RecalcCompletion(ctx context.Context, mainItemID uint) error
	ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) (*model.MainItem, error)
	AvailableTransitions(ctx context.Context, teamID, callerID, itemID uint) ([]string, error)
}

type mainItemService struct {
	mainItemRepo      repository.MainItemRepo
	subItemRepo       repository.SubItemRepo
	statusHistorySvc  StatusHistoryService
}

// NewMainItemService creates a new MainItemService.
func NewMainItemService(mainItemRepo repository.MainItemRepo, subItemRepo repository.SubItemRepo, statusHistorySvc StatusHistoryService) MainItemService {
	return &mainItemService{mainItemRepo: mainItemRepo, subItemRepo: subItemRepo, statusHistorySvc: statusHistorySvc}
}

func (s *mainItemService) Create(ctx context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error) {
	code, err := s.mainItemRepo.NextCode(ctx, teamID)
	if err != nil {
		return nil, err
	}

	item := &model.MainItem{
		TeamID:      teamID,
		Code:        code,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		ProposerID:  pmID,
		AssigneeID:  &req.AssigneeID,
		IsKeyItem:   req.IsKeyItem,
		Status:      "pending",
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

	if err := s.mainItemRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *mainItemService) Update(ctx context.Context, teamID, itemID uint, req dto.MainItemUpdateReq) error {
	item, err := s.mainItemRepo.FindByID(ctx, itemID)
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
	if req.Priority != nil {
		fields["priority"] = *req.Priority
	}
	if req.AssigneeID != nil {
		fields["assignee_id"] = *req.AssigneeID
	}
	if req.IsKeyItem != nil {
		fields["is_key_item"] = *req.IsKeyItem
	}
	if req.StartDate != nil {
		fields["start_date"] = *req.StartDate
	}
	if req.ExpectedEndDate != nil {
		fields["expected_end_date"] = *req.ExpectedEndDate
	}
	if req.ActualEndDate != nil {
		fields["actual_end_date"] = *req.ActualEndDate
	}

	if len(fields) == 0 {
		return nil
	}

	return s.mainItemRepo.Update(ctx, item, fields)
}

func (s *mainItemService) Archive(ctx context.Context, teamID, itemID uint) error {
	item, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}

	if item.Status != "completed" && item.Status != "closed" {
		return apperrors.ErrArchiveNotAllowed
	}

	now := time.Now()
	return s.mainItemRepo.Update(ctx, item, map[string]interface{}{
		"archived_at": &now,
	})
}

func (s *mainItemService) List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return s.mainItemRepo.List(ctx, teamID, filter, page)
}

func (s *mainItemService) Get(ctx context.Context, itemID uint) (*model.MainItem, error) {
	item, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	return item, nil
}

func (s *mainItemService) RecalcCompletion(ctx context.Context, mainItemID uint) error {
	item, err := s.mainItemRepo.FindByID(ctx, mainItemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}

	subItems, err := s.subItemRepo.ListByMainItem(ctx, mainItemID)
	if err != nil {
		return err
	}

	completion := calcWeightedCompletion(subItems)
	return s.mainItemRepo.Update(ctx, item, map[string]interface{}{
		"completion": completion,
	})
}

func (s *mainItemService) ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) (*model.MainItem, error) {
	item, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamID != teamID {
		return nil, apperrors.ErrForbidden
	}

	// Self-transition check
	if newStatus == item.Status {
		return nil, apperrors.ErrInvalidStatus
	}

	// Validate transition
	if !status.IsValidTransition(status.MainItemTransitions, item.Status, newStatus) {
		return nil, apperrors.ErrInvalidStatus
	}

	// PM-only check: reviewing -> completed/progressing requires caller == proposer
	if item.Status == "reviewing" && (newStatus == "completed" || newStatus == "progressing") {
		if callerID != item.ProposerID {
			return nil, apperrors.ErrForbidden
		}
	}

	fields := map[string]interface{}{
		"status": newStatus,
	}

	// Terminal side effects
	if status.IsMainTerminal(newStatus) {
		fields["completion"] = float64(100)
		now := time.Now()
		fields["actual_end_date"] = &now
	}

	// Capture old status before update (repo may mutate the item)
	oldStatus := item.Status

	if err := s.mainItemRepo.Update(ctx, item, fields); err != nil {
		return nil, err
	}

	// Record to status history
	if s.statusHistorySvc != nil {
		_ = s.statusHistorySvc.Record(ctx, &model.StatusHistory{
			ItemType:   "main_item",
			ItemID:     itemID,
			FromStatus: oldStatus,
			ToStatus:   newStatus,
			ChangedBy:  callerID,
			IsAuto:     false,
		})
	}

	// Fetch updated item
	updated, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *mainItemService) AvailableTransitions(ctx context.Context, teamID, callerID, itemID uint) ([]string, error) {
	item, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamID != teamID {
		return nil, apperrors.ErrForbidden
	}

	transitions := status.GetAvailableTransitions(status.MainItemTransitions, item.Status)

	// PM-only filter: non-PM callers don't see completed/progressing when reviewing
	if item.Status == "reviewing" && callerID != item.ProposerID {
		filtered := make([]string, 0, len(transitions))
		for _, t := range transitions {
			if t != "completed" && t != "progressing" {
				filtered = append(filtered, t)
			}
		}
		return filtered, nil
	}

	return transitions, nil
}

// calcWeightedCompletion computes weighted average of SubItem completion values.
// If no sub-items, returns 0. If all weights are zero, falls back to simple average.
func calcWeightedCompletion(items []*model.SubItem) float64 {
	if len(items) == 0 {
		return 0
	}

	var totalWeight, weightedSum float64
	for _, si := range items {
		totalWeight += si.Weight
		weightedSum += si.Completion * si.Weight
	}

	if totalWeight == 0 {
		var sum float64
		for _, si := range items {
			sum += si.Completion
		}
		return sum / float64(len(items))
	}

	return weightedSum / totalWeight
}


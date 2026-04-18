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

// SubItemService defines business operations for SubItem.
type SubItemService interface {
	Create(ctx context.Context, teamID, callerID uint, req dto.SubItemCreateReq) (*model.SubItem, error)
	Update(ctx context.Context, teamID, itemID uint, req dto.SubItemUpdateReq) error
	ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) error
	Get(ctx context.Context, teamID, itemID uint) (*model.SubItem, error)
	List(ctx context.Context, teamID uint, mainItemID *uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error)
	Assign(ctx context.Context, teamID, pmID, itemID, assigneeID uint) error
}

type subItemService struct {
	subItemRepo  repository.SubItemRepo
	mainItemSvc  MainItemService
}

// NewSubItemService creates a new SubItemService.
func NewSubItemService(subItemRepo repository.SubItemRepo, mainItemSvc MainItemService) SubItemService {
	return &subItemService{subItemRepo: subItemRepo, mainItemSvc: mainItemSvc}
}

// allowedTransitions defines the valid status transitions.
var allowedTransitions = map[string]map[string]bool{
	"待开始": {"进行中": true, "已关闭": true},
	"进行中": {"阻塞中": true, "挂起": true, "待验收": true, "已延期": true, "已关闭": true},
	"阻塞中": {"进行中": true},
	"挂起":   {"进行中": true, "已关闭": true},
	"已延期": {"进行中": true},
	"待验收": {"已完成": true, "进行中": true},
	// 已完成 and 已关闭 are terminal — no outgoing transitions
}

func (s *subItemService) Create(ctx context.Context, teamID, callerID uint, req dto.SubItemCreateReq) (*model.SubItem, error) {
	item := &model.SubItem{
		TeamID:     teamID,
		MainItemID: req.MainItemID,
		Title:      req.Title,
		Description: req.Description,
		Priority:   req.Priority,
		AssigneeID: req.AssigneeID,
		Status:     "待开始",
		Weight:     1.0,
	}

	if err := s.subItemRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *subItemService) Update(ctx context.Context, teamID, itemID uint, req dto.SubItemUpdateReq) error {
	item, err := s.subItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return mapSubItemNotFound(err)
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

func (s *subItemService) ChangeStatus(ctx context.Context, teamID, callerID, itemID uint, newStatus string) error {
	item, err := s.subItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return mapSubItemNotFound(err)
	}
	if item.TeamID != teamID {
		return apperrors.ErrForbidden
	}

	if !isValidTransition(item.Status, newStatus) {
		return apperrors.ErrInvalidStatus
	}

	fields := map[string]interface{}{
		"status": newStatus,
	}

	// Delay-count logic when transitioning to 已延期
	if newStatus == "已延期" {
		newDelayCount := item.DelayCount + 1
		fields["delay_count"] = newDelayCount

		if newDelayCount >= 2 {
			fields["is_key_item"] = true
			fields["priority"] = "P1"
		}
	}

	// Set ActualEndDate when transitioning to 已完成
	if newStatus == "已完成" {
		now := time.Now()
		fields["actual_end_date"] = &now
	}

	if err := s.subItemRepo.Update(ctx, item, fields); err != nil {
		return err
	}

	// After completing, recalculate parent MainItem completion
	if newStatus == "已完成" {
		if err := s.mainItemSvc.RecalcCompletion(ctx, item.MainItemID); err != nil {
			return err
		}
	}

	return nil
}

func (s *subItemService) Get(ctx context.Context, teamID, itemID uint) (*model.SubItem, error) {
	item, err := s.subItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, mapSubItemNotFound(err)
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
		return mapSubItemNotFound(err)
	}
	if item.TeamID != teamID {
		return apperrors.ErrForbidden
	}

	return s.subItemRepo.Update(ctx, item, map[string]interface{}{
		"assignee_id": assigneeID,
	})
}

// isValidTransition checks if a status transition is allowed.
func isValidTransition(from, to string) bool {
	transitions, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	return transitions[to]
}

// mapSubItemNotFound translates not-found errors from the repo layer into ErrItemNotFound.
func mapSubItemNotFound(err error) error {
	if err == gorm.ErrRecordNotFound || stderrors.Is(err, apperrors.ErrNotFound) {
		return apperrors.ErrItemNotFound
	}
	return err
}

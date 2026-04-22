package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/dates"
	"pm-work-tracker/backend/internal/pkg/status"
	"pm-work-tracker/backend/internal/repository"
)

// LinkageResult holds the outcome of a linkage evaluation.
type LinkageResult struct {
	Triggered    bool   // whether linkage was attempted (had sub-items)
	Success      bool   // whether the transition succeeded
	TargetStatus string // the intended target status
	Remark       string // failure reason if not success
}

// Warning returns a human-readable warning string if linkage was triggered but failed.
func (r *LinkageResult) Warning() string {
	if r != nil && r.Triggered && !r.Success {
		return fmt.Sprintf("主事项状态联动失败：%s", r.Remark)
	}
	return ""
}

// linkageMuMap provides per-MainItem mutexes for linkage evaluation.
// Bounded to maxLinkageMuMapSize entries with LRU eviction.
const maxLinkageMuMapSize = 1000

var (
	linkageMuMap    = make(map[uint]*sync.Mutex)
	linkageAccess   = make(map[uint]uint64) // mainItemID -> access sequence number
	linkageSeq      uint64                  // monotonically increasing access counter
	linkageMapMu    sync.Mutex              // protects linkageMuMap, linkageAccess, and linkageSeq
)

// getLinkageMutex returns (or creates) a mutex for the given MainItem.
// When the map exceeds maxLinkageMuMapSize entries, the least recently used entry is evicted.
func getLinkageMutex(mainItemID uint) *sync.Mutex {
	linkageMapMu.Lock()
	defer linkageMapMu.Unlock()

	if mu, ok := linkageMuMap[mainItemID]; ok {
		linkageSeq++
		linkageAccess[mainItemID] = linkageSeq
		return mu
	}

	// Evict LRU entry if at capacity
	if len(linkageMuMap) >= maxLinkageMuMapSize {
		var oldestID uint
		var oldestSeq uint64
		first := true
		for id, seq := range linkageAccess {
			if first || seq < oldestSeq {
				oldestID = id
				oldestSeq = seq
				first = false
			}
		}
		delete(linkageMuMap, oldestID)
		delete(linkageAccess, oldestID)
	}

	mu := &sync.Mutex{}
	linkageSeq++
	linkageMuMap[mainItemID] = mu
	linkageAccess[mainItemID] = linkageSeq
	return mu
}

// resetLinkageMuMap resets the global linkage map. Used only in tests.
func resetLinkageMuMap() {
	linkageMapMu.Lock()
	defer linkageMapMu.Unlock()
	linkageMuMap = make(map[uint]*sync.Mutex)
	linkageAccess = make(map[uint]uint64)
	linkageSeq = 0
}

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
	EvaluateLinkage(ctx context.Context, mainItemID uint, changedBy uint) (*LinkageResult, error)
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
	if status.IsMainTerminal(item.Status) {
		return apperrors.ErrTerminalMainItem
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

	// Guard: cannot transition to terminal if any sub-item is non-terminal
	if status.IsMainTerminal(newStatus) {
		subs, err := s.subItemRepo.ListByMainItem(ctx, itemID)
		if err != nil {
			return nil, err
		}
		for _, sub := range subs {
			if !status.IsSubTerminal(sub.Status) {
				return nil, apperrors.ErrSubItemsNotTerminal
			}
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

// EvaluateLinkage evaluates the main-sub item linkage rules and updates MainItem status.
// It acquires a per-MainItem mutex to prevent race conditions.
func (s *mainItemService) EvaluateLinkage(ctx context.Context, mainItemID uint, changedBy uint) (*LinkageResult, error) {
	mu := getLinkageMutex(mainItemID)
	mu.Lock()
	defer mu.Unlock()

	mainItem, err := s.mainItemRepo.FindByID(ctx, mainItemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}

	subItems, err := s.subItemRepo.ListByMainItem(ctx, mainItemID)
	if err != nil {
		return nil, err
	}

	// No sub-items: no linkage triggered
	if len(subItems) == 0 {
		return nil, nil
	}

	// Evaluate 5-level priority rules
	targetStatus := evaluateLinkageTarget(subItems, mainItem.Status)
	if targetStatus == "" || targetStatus == mainItem.Status {
		return nil, nil
	}

	// Check if transition is valid
	if !status.IsValidTransition(status.MainItemTransitions, mainItem.Status, targetStatus) {
		// Linkage failed: record intent in status history
		remark := fmt.Sprintf("%s→%s 不允许", mainItem.Status, targetStatus)
		if s.statusHistorySvc != nil {
			_ = s.statusHistorySvc.Record(ctx, &model.StatusHistory{
				ItemType:   "main_item",
				ItemID:     mainItemID,
				FromStatus: mainItem.Status,
				ToStatus:   targetStatus,
				ChangedBy:  changedBy,
				IsAuto:     true,
				Remark:     remark,
			})
		}
		return &LinkageResult{
			Triggered:    true,
			Success:      false,
			TargetStatus: targetStatus,
			Remark:       remark,
		}, nil
	}

	// Apply transition
	fields := map[string]interface{}{
		"status": targetStatus,
	}

	// Terminal side effects
	if status.IsMainTerminal(targetStatus) {
		fields["completion"] = float64(100)
		now := time.Now()
		fields["actual_end_date"] = &now
	}

	oldStatus := mainItem.Status

	if err := s.mainItemRepo.Update(ctx, mainItem, fields); err != nil {
		return nil, err
	}

	// Record to status history (is_auto=true)
	if s.statusHistorySvc != nil {
		_ = s.statusHistorySvc.Record(ctx, &model.StatusHistory{
			ItemType:   "main_item",
			ItemID:     mainItemID,
			FromStatus: oldStatus,
			ToStatus:   targetStatus,
			ChangedBy:  changedBy,
			IsAuto:     true,
		})
	}

	return &LinkageResult{
		Triggered:    true,
		Success:      true,
		TargetStatus: targetStatus,
	}, nil
}

// evaluateLinkageTarget determines the target status based on 5-level priority rules.
// Returns empty string if no linkage rule matches.
func evaluateLinkageTarget(subItems []*model.SubItem, currentMainStatus string) string {
	allTerminal := true     // completed or closed
	allClosed := true       // closed only
	allPausingOrClosed := true
	hasCompleted := false
	hasBlocking := false
	hasProgressing := false

	for _, si := range subItems {
		s := si.Status
		isCompleted := s == "completed"
		isClosed := s == "closed"
		isTerminal := isCompleted || isClosed
		isPausing := s == "pausing"

		if !isTerminal {
			allTerminal = false
		}
		if !isClosed {
			allClosed = false
		}
		if !(isPausing || isClosed) {
			allPausingOrClosed = false
		}
		if isCompleted {
			hasCompleted = true
		}
		if s == "blocking" {
			hasBlocking = true
		}
		if s == "progressing" {
			hasProgressing = true
		}
	}

	// Priority 1: all completed/closed + at least one completed → reviewing
	if allTerminal && hasCompleted {
		return "reviewing"
	}

	// Priority 2: all closed → closed
	if allClosed {
		return "closed"
	}

	// Priority 3: all pausing (or pausing + closed) → pausing
	if allPausingOrClosed && !allClosed {
		return "pausing"
	}

	// Priority 4: any blocking (not all terminal) → blocking (only from pending/progressing)
	if hasBlocking && !allTerminal {
		if currentMainStatus == "pending" || currentMainStatus == "progressing" {
			return "blocking"
		}
	}

	// Priority 5: any progressing → progressing (only from pending)
	if hasProgressing && currentMainStatus == "pending" {
		return "progressing"
	}

	// AC-9: reviewing + non-terminal sub-items → revert to progressing
	if currentMainStatus == "reviewing" && !allTerminal {
		return "progressing"
	}

	return ""
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


package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/pkg/dates"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/snowflake"
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
	linkageMuMap  = make(map[int64]*sync.Mutex)
	linkageAccess = make(map[int64]uint64) // mainItemBizKey -> access sequence number
	linkageSeq    uint64                   // monotonically increasing access counter
	linkageMapMu  sync.Mutex               // protects linkageMuMap, linkageAccess, and linkageSeq
)

// getLinkageMutex returns (or creates) a mutex for the given MainItem.
// When the map exceeds maxLinkageMuMapSize entries, the least recently used entry is evicted.
func getLinkageMutex(mainItemBizKey int64) *sync.Mutex {
	linkageMapMu.Lock()
	defer linkageMapMu.Unlock()

	if mu, ok := linkageMuMap[mainItemBizKey]; ok {
		linkageSeq++
		linkageAccess[mainItemBizKey] = linkageSeq
		return mu
	}

	// Evict LRU entry if at capacity
	if len(linkageMuMap) >= maxLinkageMuMapSize {
		var oldestKey int64
		var oldestSeq uint64
		first := true
		for key, seq := range linkageAccess {
			if first || seq < oldestSeq {
				oldestKey = key
				oldestSeq = seq
				first = false
			}
		}
		delete(linkageMuMap, oldestKey)
		delete(linkageAccess, oldestKey)
	}

	mu := &sync.Mutex{}
	linkageSeq++
	linkageMuMap[mainItemBizKey] = mu
	linkageAccess[mainItemBizKey] = linkageSeq
	return mu
}

// resetLinkageMuMap resets the global linkage map. Used only in tests.
func resetLinkageMuMap() {
	linkageMapMu.Lock()
	defer linkageMapMu.Unlock()
	linkageMuMap = make(map[int64]*sync.Mutex)
	linkageAccess = make(map[int64]uint64)
	linkageSeq = 0
}

// MainItemService defines business operations for MainItem.
type MainItemService interface {
	Create(ctx context.Context, teamBizKey int64, pmBizKey int64, req dto.MainItemCreateReq) (*model.MainItem, error)
	Update(ctx context.Context, teamBizKey int64, itemID uint, req dto.MainItemUpdateReq) error
	Archive(ctx context.Context, teamBizKey int64, itemID uint) error
	Delete(ctx context.Context, teamBizKey int64, itemBizKey int64, operatorBizKey int64) error
	List(ctx context.Context, teamBizKey int64, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], map[int64]*dto.MainItemMatchInfo, error)
	Get(ctx context.Context, itemID uint) (*model.MainItem, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error)
	RecalcCompletion(ctx context.Context, mainItemBizKey int64) error
	ChangeStatus(ctx context.Context, teamBizKey int64, callerBizKey int64, itemID uint, newStatus string) (*model.MainItem, error)
	AvailableTransitions(ctx context.Context, teamBizKey int64, callerBizKey int64, itemID uint) ([]string, error)
	EvaluateLinkage(ctx context.Context, mainItemBizKey int64, changedByBizKey int64) (*LinkageResult, error)
}

type mainItemService struct {
	mainItemRepo     repository.MainItemRepo
	subItemRepo      repository.SubItemRepo
	statusHistorySvc StatusHistoryService
	milestoneRepo    repository.MilestoneRepo
	milestoneMapRepo repository.MilestoneMapRepo
}

// NewMainItemService creates a new MainItemService.
func NewMainItemService(mainItemRepo repository.MainItemRepo, subItemRepo repository.SubItemRepo, statusHistorySvc StatusHistoryService) MainItemService {
	return &mainItemService{mainItemRepo: mainItemRepo, subItemRepo: subItemRepo, statusHistorySvc: statusHistorySvc}
}

// WithMilestoneRepos attaches milestone repos to the service.
// Used for functional-option style DI when milestone features are enabled.
func WithMilestoneRepos(svc MainItemService, msRepo repository.MilestoneRepo, mapRepo repository.MilestoneMapRepo) MainItemService {
	s := svc.(*mainItemService)
	s.milestoneRepo = msRepo
	s.milestoneMapRepo = mapRepo
	return s
}

func (s *mainItemService) Create(ctx context.Context, teamBizKey, pmBizKey int64, req dto.MainItemCreateReq) (*model.MainItem, error) {
	code, err := s.mainItemRepo.NextCode(ctx, teamBizKey)
	if err != nil {
		return nil, err
	}

	item := &model.MainItem{
		BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:     teamBizKey,
		Code:        code,
		Title:       req.Title,
		ItemDesc:    req.Description,
		Priority:    req.Priority,
		ProposerKey: pmBizKey,
		AssigneeKey: func() *int64 {
			if req.AssigneeKey != "" {
				v, _ := pkg.ParseID(req.AssigneeKey)
				return &v
			}
			return nil
		}(),
		IsKeyItem:  req.IsKeyItem,
		ItemStatus: "pending",
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

	if err := s.mainItemRepo.Create(ctx, item); err != nil {
		if apperrors.IsMySQLDuplicateError(err) {
			return nil, apperrors.ErrDuplicateBizKey
		}
		return nil, err
	}
	return item, nil
}

func (s *mainItemService) Update(ctx context.Context, teamBizKey int64, itemID uint, req dto.MainItemUpdateReq) error {
	item, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamKey != teamBizKey {
		return apperrors.ErrForbidden
	}

	fields := map[string]interface{}{}

	// Milestone binding validation runs before the generic terminal check
	// so that BR-3 returns the specific ErrTerminalItemCannotMove error.
	if req.MilestoneKey != nil {
		if err := s.validateMilestoneBinding(ctx, item, *req.MilestoneKey, fields); err != nil {
			return err
		}
	}

	// Generic terminal check for non-milestone fields
	if status.IsMainTerminal(item.ItemStatus) {
		return apperrors.ErrTerminalMainItem
	}
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
	if req.IsKeyItem != nil {
		fields["is_key_item"] = *req.IsKeyItem
	}
	if req.StartDate != nil {
		fields["plan_start_date"] = *req.StartDate
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

func (s *mainItemService) Archive(ctx context.Context, _ int64, itemID uint) error {
	item, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}

	if item.ItemStatus != "completed" && item.ItemStatus != "closed" {
		return apperrors.ErrArchiveNotAllowed
	}

	now := time.Now()
	return s.mainItemRepo.Update(ctx, item, map[string]interface{}{
		"archived_at": &now,
	})
}

func (s *mainItemService) Delete(ctx context.Context, _, itemBizKey, operatorBizKey int64) error {
	item, err := s.mainItemRepo.FindByBizKey(ctx, itemBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}

	// Fetch sub-items for cascade delete
	subItems, err := s.subItemRepo.ListByMainItem(ctx, item.BizKey)
	if err != nil {
		return err
	}

	// Build sub-item ID list and status history records
	subItemIDs := make([]uint, 0, len(subItems))
	histories := make([]model.StatusHistory, 0, 1+len(subItems))

	// Main item audit record
	histories = append(histories, model.StatusHistory{
		ItemType:   "main_item",
		ItemKey:    item.BizKey,
		FromStatus: item.ItemStatus,
		ToStatus:   "deleted",
		ChangedBy:  operatorBizKey,
		IsAuto:     0,
	})

	// Sub-item audit records
	for _, sub := range subItems {
		subItemIDs = append(subItemIDs, sub.ID)
		histories = append(histories, model.StatusHistory{
			ItemType:   "sub_item",
			ItemKey:    sub.BizKey,
			FromStatus: sub.ItemStatus,
			ToStatus:   "deleted",
			ChangedBy:  operatorBizKey,
			IsAuto:     0,
		})
	}

	// Cascade soft-delete in single transaction (Hard Rule)
	return s.mainItemRepo.CascadeSoftDelete(ctx, item.ID, subItemIDs, histories)
}

func (s *mainItemService) List(ctx context.Context, teamBizKey int64, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], map[int64]*dto.MainItemMatchInfo, error) {
	// When assigneeKey is set, perform penetration filter in memory.
	if filter.AssigneeKey != nil && *filter.AssigneeKey != "" {
		return s.listWithPenetration(ctx, teamBizKey, filter, page)
	}

	// No assigneeKey: use standard repo-level filtering (no penetration needed).
	result, err := s.mainItemRepo.List(ctx, teamBizKey, filter, page)
	if err != nil {
		return nil, nil, err
	}

	// Terminal sort: sink terminal items to bottom while preserving relative order
	sortTerminalItems(result.Items)

	// Status-only filter: all items are "direct" matches
	var matchInfo map[int64]*dto.MainItemMatchInfo
	if len(filter.Statuses) > 0 {
		matchInfo = make(map[int64]*dto.MainItemMatchInfo, len(result.Items))
		for _, item := range result.Items {
			matchInfo[item.BizKey] = &dto.MainItemMatchInfo{MatchType: "direct"}
		}
	}

	return result, matchInfo, nil
}

func (s *mainItemService) listWithPenetration(ctx context.Context, teamBizKey int64, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], map[int64]*dto.MainItemMatchInfo, error) {
	assigneeBizKey, err := pkg.ParseID(*filter.AssigneeKey)
	if err != nil {
		// Invalid assigneeKey: return empty result
		return &dto.PageResult[model.MainItem]{
			Items: []model.MainItem{},
			Total: 0,
			Page:  page.Page,
			Size:  page.PageSize,
		}, nil, nil
	}

	// Fetch all non-archived main items for the team (no assignee filter at DB level)
	noAssigneeFilter := filter
	noAssigneeFilter.AssigneeKey = nil
	allItems, err := s.mainItemRepo.List(ctx, teamBizKey, noAssigneeFilter, dto.Pagination{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, nil, err
	}

	// Fetch all sub-items for the team
	subItems, err := s.subItemRepo.ListByTeam(ctx, teamBizKey)
	if err != nil {
		return nil, nil, err
	}

	// Index sub-items by main item BizKey
	subsByMain := make(map[int64][]model.SubItem)
	for _, si := range subItems {
		subsByMain[si.MainItemKey] = append(subsByMain[si.MainItemKey], si)
	}

	// Filter main items with penetration logic
	var filtered []model.MainItem
	matchInfo := make(map[int64]*dto.MainItemMatchInfo)
	statusSet := make(map[string]struct{}, len(filter.Statuses))
	for _, s := range filter.Statuses {
		statusSet[s] = struct{}{}
	}

	for _, mi := range allItems.Items {
		// Check status match (if statuses selected)
		statusMatch := len(statusSet) == 0
		if !statusMatch {
			_, statusMatch = statusSet[mi.ItemStatus]
		}

		// Check direct assignee match
		directMatch := mi.AssigneeKey != nil && *mi.AssigneeKey == assigneeBizKey

		// Check indirect match via sub-items
		var matchedSubIds []string
		for _, si := range subsByMain[mi.BizKey] {
			if si.AssigneeKey != nil && *si.AssigneeKey == assigneeBizKey {
				matchedSubIds = append(matchedSubIds, pkg.FormatID(si.BizKey))
			}
		}
		indirectMatch := len(matchedSubIds) > 0

		// AND logic when both filters active
		if len(statusSet) > 0 {
			// Status must match AND (direct OR indirect)
			if !statusMatch {
				continue
			}
			if !directMatch && !indirectMatch {
				continue
			}
		} else if !directMatch && !indirectMatch {
			// Only assignee filter: either direct or indirect
			continue
		}

		filtered = append(filtered, mi)
		switch {
		case directMatch && len(matchedSubIds) == 0:
			matchInfo[mi.BizKey] = &dto.MainItemMatchInfo{MatchType: "direct"}
		case directMatch:
			// Direct match but also has matching sub-items — report as direct with matched subs
			matchInfo[mi.BizKey] = &dto.MainItemMatchInfo{
				MatchType:         "direct",
				MatchedSubItemIds: matchedSubIds,
			}
		default:
			// Only indirect match
			matchInfo[mi.BizKey] = &dto.MainItemMatchInfo{
				MatchType:         "indirect",
				MatchedSubItemIds: matchedSubIds,
			}
		}
	}

	// Terminal sort: sink terminal items to bottom while preserving relative order
	sortTerminalItems(filtered)

	// Paginate filtered results
	total := int64(len(filtered))
	_, resultPage, pageSize := dto.ApplyPaginationDefaults(page.Page, page.PageSize)
	offset := (resultPage - 1) * pageSize

	if offset >= len(filtered) {
		return &dto.PageResult[model.MainItem]{
			Items: []model.MainItem{},
			Total: total,
			Page:  resultPage,
			Size:  pageSize,
		}, matchInfo, nil
	}

	end := offset + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return &dto.PageResult[model.MainItem]{
		Items: filtered[offset:end],
		Total: total,
		Page:  resultPage,
		Size:  pageSize,
	}, matchInfo, nil
}

func (s *mainItemService) Get(ctx context.Context, itemID uint) (*model.MainItem, error) {
	item, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	return item, nil
}

func (s *mainItemService) GetByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error) {
	item, err := s.mainItemRepo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	return item, nil
}

func (s *mainItemService) RecalcCompletion(ctx context.Context, mainItemBizKey int64) error {
	item, err := s.mainItemRepo.FindByBizKey(ctx, mainItemBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}

	subItems, err := s.subItemRepo.ListByMainItem(ctx, mainItemBizKey)
	if err != nil {
		return err
	}

	completion := calcWeightedCompletion(subItems)
	return s.mainItemRepo.Update(ctx, item, map[string]interface{}{
		"completion_pct": completion,
	})
}

func (s *mainItemService) ChangeStatus(ctx context.Context, teamBizKey, callerBizKey int64, itemID uint, newStatus string) (*model.MainItem, error) {
	item, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamKey != teamBizKey {
		return nil, apperrors.ErrForbidden
	}

	// Self-transition check
	if newStatus == item.ItemStatus {
		return nil, apperrors.ErrInvalidStatus
	}

	// Validate transition
	if !status.IsValidTransition(status.MainItemTransitions, item.ItemStatus, newStatus) {
		return nil, apperrors.ErrInvalidStatus
	}

	// PM-only check: reviewing -> completed/progressing requires caller == proposer
	if item.ItemStatus == "reviewing" && (newStatus == "completed" || newStatus == "progressing") {
		if callerBizKey != item.ProposerKey {
			return nil, apperrors.ErrForbidden
		}
	}

	// Guard: cannot transition to terminal if any sub-item is non-terminal
	if status.IsMainTerminal(newStatus) {
		subs, err := s.subItemRepo.ListByMainItem(ctx, item.BizKey) //nolint:govet // intentional shadow: inner-block err assignment
		if err != nil {
			return nil, err
		}
		for _, sub := range subs {
			if !status.IsSubTerminal(sub.ItemStatus) {
				return nil, apperrors.ErrSubItemsNotTerminal
			}
		}
	}

	fields := map[string]interface{}{
		"item_status": newStatus,
	}

	// Terminal side effects
	if status.IsMainTerminal(newStatus) {
		fields["completion_pct"] = float64(100)
		now := time.Now()
		fields["actual_end_date"] = &now
	}

	// Capture old status before update (repo may mutate the item)
	oldStatus := item.ItemStatus

	if err := s.mainItemRepo.Update(ctx, item, fields); err != nil { //nolint:govet // intentional shadow: inner-block err assignment
		return nil, err
	}

	// Record to status history
	if err := RecordStatusChange(s.statusHistorySvc, ctx, "main_item", item.BizKey, oldStatus, newStatus, callerBizKey, 0, ""); err != nil { //nolint:govet // intentional shadow: inner-block err assignment
		return nil, err
	}

	// Fetch updated item
	updated, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *mainItemService) AvailableTransitions(ctx context.Context, teamBizKey, callerBizKey int64, itemID uint) ([]string, error) {
	item, err := s.mainItemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	if item.TeamKey != teamBizKey {
		return nil, apperrors.ErrForbidden
	}

	transitions := status.GetAvailableTransitions(status.MainItemTransitions, item.ItemStatus)

	// PM-only filter: non-PM callers don't see completed/progressing when reviewing
	if item.ItemStatus == "reviewing" && callerBizKey != item.ProposerKey {
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
func (s *mainItemService) EvaluateLinkage(ctx context.Context, mainItemBizKey, changedByBizKey int64) (*LinkageResult, error) {
	mu := getLinkageMutex(mainItemBizKey)
	mu.Lock()
	defer mu.Unlock()

	mainItem, err := s.mainItemRepo.FindByBizKey(ctx, mainItemBizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}

	subItems, err := s.subItemRepo.ListByMainItem(ctx, mainItemBizKey)
	if err != nil {
		return nil, err
	}

	// No sub-items: no linkage triggered
	if len(subItems) == 0 {
		return nil, nil
	}

	// Evaluate 5-level priority rules
	targetStatus := evaluateLinkageTarget(subItems, mainItem.ItemStatus)
	if targetStatus == "" || targetStatus == mainItem.ItemStatus {
		return nil, nil
	}

	// Check if transition is valid
	if !status.IsValidTransition(status.MainItemTransitions, mainItem.ItemStatus, targetStatus) {
		// Linkage failed: record intent in status history
		remark := fmt.Sprintf("%s→%s 不允许", mainItem.ItemStatus, targetStatus)
		_ = RecordStatusChange(s.statusHistorySvc, ctx, "main_item", mainItemBizKey, mainItem.ItemStatus, targetStatus, changedByBizKey, 1, remark)
		return &LinkageResult{
			Triggered:    true,
			Success:      false,
			TargetStatus: targetStatus,
			Remark:       remark,
		}, nil
	}

	// Apply transition
	fields := map[string]interface{}{
		"item_status": targetStatus,
	}

	// Terminal side effects
	if status.IsMainTerminal(targetStatus) {
		fields["completion_pct"] = float64(100)
		now := time.Now()
		fields["actual_end_date"] = &now
	}

	oldStatus := mainItem.ItemStatus

	if err := s.mainItemRepo.Update(ctx, mainItem, fields); err != nil {
		return nil, err
	}

	// Record to status history (is_auto=true)
	_ = RecordStatusChange(s.statusHistorySvc, ctx, "main_item", mainItemBizKey, oldStatus, targetStatus, changedByBizKey, 1, "")

	return &LinkageResult{
		Triggered:    true,
		Success:      true,
		TargetStatus: targetStatus,
	}, nil
}

// evaluateLinkageTarget determines the target status based on 5-level priority rules.
// Returns empty string if no linkage rule matches.
func evaluateLinkageTarget(subItems []*model.SubItem, currentMainStatus string) string {
	allTerminal := true // completed or closed
	allClosed := true   // closed only
	allPausingOrClosed := true
	hasCompleted := false
	hasBlocking := false
	hasProgressing := false

	for _, si := range subItems {
		s := si.ItemStatus
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
		if !isPausing && !isClosed {
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

// sortTerminalItems sinks terminal main items to the bottom of the list
// while preserving the relative order of terminal and non-terminal groups.
func sortTerminalItems(items []model.MainItem) {
	sort.SliceStable(items, func(i, j int) bool {
		iTerminal := status.IsMainTerminal(items[i].ItemStatus)
		jTerminal := status.IsMainTerminal(items[j].ItemStatus)
		if iTerminal != jTerminal {
			return !iTerminal // terminal sinks to bottom
		}
		return false // preserve original relative order within each group
	})
}

// validateMilestoneBinding enforces BR-3 and BR-5 when MilestoneKey is modified.
// It sets the "milestone_key" field in the fields map on success.
func (s *mainItemService) validateMilestoneBinding(ctx context.Context, item *model.MainItem, milestoneKeyStr string, fields map[string]interface{}) error {
	// Empty string means unbind — no validation needed, just clear the key
	if milestoneKeyStr == "" {
		fields["milestone_key"] = nil
		return nil
	}

	// BR-3 check 1: terminal MI cannot change milestone_key
	if status.IsMainTerminal(item.ItemStatus) {
		return apperrors.ErrTerminalItemCannotMove
	}

	// Parse the target milestone bizKey
	msBizKey, err := pkg.ParseID(milestoneKeyStr)
	if err != nil {
		return apperrors.ErrValidation
	}

	// Look up the target milestone
	ms, err := s.milestoneRepo.FindByBizKey(ctx, msBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}

	// BR-3 check 2: target milestone in terminal state cannot receive MI
	if def, ok := status.GetMilestoneStatus(ms.MilestoneStatus); ok && def.Terminal {
		return apperrors.ErrTerminalMilestoneCannotReceive
	}

	// BR-5: parent milestone map in terminal state prevents milestone_key changes
	if s.milestoneMapRepo != nil {
		mmap, err := s.milestoneMapRepo.FindByBizKey(ctx, ms.MilestoneMapKey)
		if err != nil {
			return apperrors.MapNotFound(err, apperrors.ErrNotFound)
		}
		if def, ok := status.GetMilestoneMapStatus(mmap.MapStatus); ok && def.Terminal {
			return apperrors.ErrMapIsTerminal
		}
	}

	fields["milestone_key"] = msBizKey
	return nil
}

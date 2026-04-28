package service

import (
	"context"
	"time"

	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/pkg/status"
	"pm-work-tracker/backend/internal/repository"
)

// ProgressService defines business operations for progress records.
type ProgressService interface {
	Append(ctx context.Context, teamID, authorID, subItemID uint, completion float64, achievement, blocker, lesson string, isPM bool) (*model.ProgressRecord, error)
	CorrectCompletion(ctx context.Context, teamID, recordID uint, completion float64) error
	List(ctx context.Context, teamID, subItemID uint) ([]model.ProgressRecord, error)
	GetByBizKey(ctx context.Context, bizKey int64) (*model.ProgressRecord, error)
}

type progressService struct {
	progressRepo      repository.ProgressRepo
	subItemRepo       repository.SubItemRepo
	mainItemSvc       MainItemService
	statusHistorySvc  StatusHistoryService
}

// NewProgressService creates a new ProgressService.
func NewProgressService(progressRepo repository.ProgressRepo, subItemRepo repository.SubItemRepo, mainItemSvc MainItemService, statusHistorySvc StatusHistoryService) ProgressService {
	return &progressService{progressRepo: progressRepo, subItemRepo: subItemRepo, mainItemSvc: mainItemSvc, statusHistorySvc: statusHistorySvc}
}

func (s *progressService) Append(ctx context.Context, teamID, authorID, subItemID uint, completion float64, achievement, blocker, lesson string, isPM bool) (*model.ProgressRecord, error) {
	subItem, err := s.subItemRepo.FindByID(ctx, subItemID)
	if err != nil {
		return nil, err
	}

	latest, err := s.progressRepo.LatestBySubItem(ctx, subItem.BizKey)
	if err != nil {
		return nil, err
	}

	// Regression check: skip for PM
	if !isPM {
		if latest != nil && completion < latest.Completion {
			return nil, apperrors.ErrProgressRegression
		}
	}

	isFirstProgress := latest == nil

	record := &model.ProgressRecord{
		BizKey:      snowflake.Generate(),
		SubItemKey:  subItem.BizKey,
		TeamKey:     int64(teamID),
		AuthorKey:   int64(authorID),
		Completion:  completion,
		Achievement: achievement,
		Blocker:     blocker,
		Lesson:      lesson,
		CreateTime:  time.Now(),
	}

	if err := s.progressRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	// Auto-status-transition: determine target status
	currentStatus := subItem.ItemStatus
	targetStatus := currentStatus

	// Rule 1: first progress on pending sub-item -> progressing
	if isFirstProgress && currentStatus == "pending" {
		targetStatus = "progressing"
	}

	// Rule 2: 100% completion -> completed (if transition is valid from current or intermediate target)
	if completion == 100 {
		if status.IsValidTransition(status.SubItemTransitions, targetStatus, "completed") {
			targetStatus = "completed"
		}
	}

	// Build update fields
	fields := map[string]interface{}{
		"completion_pct": completion,
	}

	if targetStatus != currentStatus {
		fields["item_status"] = targetStatus
		if targetStatus == "completed" {
			now := time.Now()
			fields["completion_pct"] = float64(100)
			fields["actual_end_date"] = &now
		}

		// Record auto-transition to status history
		if err := RecordStatusChange(s.statusHistorySvc, ctx, "sub_item", int64(subItemID), currentStatus, targetStatus, authorID, 1, ""); err != nil {
			return nil, err
		}
	}

	if err := s.subItemRepo.Update(ctx, subItem, fields); err != nil {
		return nil, err
	}

	// Trigger MainItem completion rollup
	if err := s.mainItemSvc.RecalcCompletion(ctx, uint(subItem.MainItemKey)); err != nil {
		return nil, err
	}

	return record, nil
}

func (s *progressService) GetByBizKey(ctx context.Context, bizKey int64) (*model.ProgressRecord, error) {
	record, err := s.progressRepo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}
	return record, nil
}

func (s *progressService) CorrectCompletion(ctx context.Context, teamID, recordID uint, completion float64) error {
	record, err := s.progressRepo.FindByID(ctx, recordID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrItemNotFound)
	}

	if err := s.progressRepo.UpdateCompletion(ctx, recordID, completion); err != nil {
		return err
	}

	// Re-sync SubItem.Completion to the latest record's completion
	latest, err := s.progressRepo.LatestBySubItem(ctx, record.SubItemKey)
	if err != nil {
		return err
	}

	subItem, err := s.subItemRepo.FindByBizKey(ctx, record.SubItemKey)
	if err != nil {
		return err
	}

	// Use latest record's completion (which may or may not be the corrected one)
	var syncCompletion float64
	if latest != nil {
		syncCompletion = latest.Completion
		// If the corrected record IS the latest, use the new completion value
		if latest.ID == recordID {
			syncCompletion = completion
		}
	}

	if err := s.subItemRepo.Update(ctx, subItem, map[string]interface{}{
		"completion_pct": syncCompletion,
	}); err != nil {
		return err
	}

	// Trigger MainItem completion rollup
	return s.mainItemSvc.RecalcCompletion(ctx, uint(subItem.MainItemKey))
}

func (s *progressService) List(ctx context.Context, teamID, subItemID uint) ([]model.ProgressRecord, error) {
	subItem, err := s.subItemRepo.FindByID(ctx, subItemID)
	if err != nil {
		return nil, err
	}
	return s.progressRepo.ListBySubItem(ctx, teamID, subItem.BizKey)
}


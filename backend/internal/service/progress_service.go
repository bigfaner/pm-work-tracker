package service

import (
	"context"
	"time"

	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

// ProgressService defines business operations for progress records.
type ProgressService interface {
	Append(ctx context.Context, teamID, authorID, subItemID uint, completion float64, achievement, blocker, lesson string, isPM bool) (*model.ProgressRecord, error)
	CorrectCompletion(ctx context.Context, teamID, recordID uint, completion float64) error
	List(ctx context.Context, teamID, subItemID uint) ([]model.ProgressRecord, error)
}

type progressService struct {
	progressRepo  repository.ProgressRepo
	subItemRepo   repository.SubItemRepo
	mainItemSvc   MainItemService
}

// NewProgressService creates a new ProgressService.
func NewProgressService(progressRepo repository.ProgressRepo, subItemRepo repository.SubItemRepo, mainItemSvc MainItemService) ProgressService {
	return &progressService{progressRepo: progressRepo, subItemRepo: subItemRepo, mainItemSvc: mainItemSvc}
}

func (s *progressService) Append(ctx context.Context, teamID, authorID, subItemID uint, completion float64, achievement, blocker, lesson string, isPM bool) (*model.ProgressRecord, error) {
	subItem, err := s.subItemRepo.FindByID(ctx, subItemID)
	if err != nil {
		return nil, err
	}

	// Regression check: skip for PM
	if !isPM {
		latest, err := s.progressRepo.LatestBySubItem(ctx, subItemID)
		if err != nil {
			return nil, err
		}
		if latest != nil && completion < latest.Completion {
			return nil, apperrors.ErrProgressRegression
		}
	}

	record := &model.ProgressRecord{
		SubItemID:   subItemID,
		TeamID:      teamID,
		AuthorID:    authorID,
		Completion:  completion,
		Achievement: achievement,
		Blocker:     blocker,
		Lesson:      lesson,
		CreatedAt:   time.Now(),
	}

	if err := s.progressRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	// Update SubItem.Completion
	if err := s.subItemRepo.Update(ctx, subItem, map[string]interface{}{
		"completion": completion,
	}); err != nil {
		return nil, err
	}

	// Trigger MainItem completion rollup
	if err := s.mainItemSvc.RecalcCompletion(ctx, subItem.MainItemID); err != nil {
		return nil, err
	}

	return record, nil
}

func (s *progressService) CorrectCompletion(ctx context.Context, teamID, recordID uint, completion float64) error {
	record, err := s.progressRepo.FindByID(ctx, recordID)
	if err != nil {
		return mapProgressNotFound(err)
	}

	if err := s.progressRepo.UpdateCompletion(ctx, recordID, completion); err != nil {
		return err
	}

	// Re-sync SubItem.Completion to the latest record's completion
	latest, err := s.progressRepo.LatestBySubItem(ctx, record.SubItemID)
	if err != nil {
		return err
	}

	subItem, err := s.subItemRepo.FindByID(ctx, record.SubItemID)
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
		"completion": syncCompletion,
	}); err != nil {
		return err
	}

	// Trigger MainItem completion rollup
	return s.mainItemSvc.RecalcCompletion(ctx, subItem.MainItemID)
}

func (s *progressService) List(ctx context.Context, teamID, subItemID uint) ([]model.ProgressRecord, error) {
	return s.progressRepo.ListBySubItem(ctx, teamID, subItemID)
}

// mapProgressNotFound translates not-found errors from the repo layer into ErrItemNotFound.
func mapProgressNotFound(err error) error {
	if err == apperrors.ErrNotFound {
		return apperrors.ErrItemNotFound
	}
	return err
}

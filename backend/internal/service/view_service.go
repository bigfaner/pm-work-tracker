package service

import (
	"context"
	"time"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

// ViewService defines read-only view operations.
type ViewService interface {
	WeeklyView(ctx context.Context, teamID uint, weekStart time.Time) (*dto.WeeklyViewResult, error)
}

type viewService struct {
	mainItemRepo repository.MainItemRepo
	subItemRepo  repository.SubItemRepo
	progressRepo repository.ProgressRepo
}

// NewViewService creates a new ViewService.
func NewViewService(mainItemRepo repository.MainItemRepo, subItemRepo repository.SubItemRepo, progressRepo repository.ProgressRepo) ViewService {
	return &viewService{
		mainItemRepo: mainItemRepo,
		subItemRepo:  subItemRepo,
		progressRepo: progressRepo,
	}
}

func (s *viewService) WeeklyView(ctx context.Context, teamID uint, weekStart time.Time) (*dto.WeeklyViewResult, error) {
	// Validate weekStart is a Monday
	if weekStart.Weekday() != time.Monday {
		return nil, apperrors.ErrValidation
	}

	weekEnd := weekStart.AddDate(0, 0, 6)

	// Bulk fetch all non-archived main items for the team
	mainItems, err := s.mainItemRepo.ListNonArchivedByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Bulk fetch all sub-items for the team
	subItems, err := s.subItemRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Bulk fetch all progress records for the team in the week range
	// Use start of day for weekStart and start of next day for weekEnd+1
	rangeStart := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	rangeEnd := time.Date(weekEnd.Year(), weekEnd.Month(), weekEnd.Day(), 0, 0, 0, 0, weekEnd.Location()).AddDate(0, 0, 1)
	progressRecords, err := s.progressRepo.ListByTeamInRange(ctx, teamID, rangeStart, rangeEnd)
	if err != nil {
		return nil, err
	}

	// Index sub-items by main item ID
	subItemsByMain := make(map[uint][]model.SubItem)
	for _, si := range subItems {
		subItemsByMain[si.MainItemID] = append(subItemsByMain[si.MainItemID], si)
	}

	// Index progress records by sub item ID
	progressBySub := make(map[uint][]model.ProgressRecord)
	for _, pr := range progressRecords {
		progressBySub[pr.SubItemID] = append(progressBySub[pr.SubItemID], pr)
	}

	// Build groups
	var groups []dto.WeeklyGroupDTO
	for _, mi := range mainItems {
		subs, ok := subItemsByMain[mi.ID]
		if !ok || len(subs) == 0 {
			continue // Main items with no sub-items are omitted
		}

		group := dto.WeeklyGroupDTO{
			MainItem: dto.MainItemSummaryDTO{
				ID:         mi.ID,
				Title:      mi.Title,
				Completion: mi.Completion,
			},
		}

		for _, si := range subs {
			weekProgress := progressBySub[si.ID]

			if isNewlyCompleted(si, weekStart, weekEnd) {
				group.NewlyCompleted = append(group.NewlyCompleted, toSubItemWeekDTO(si, weekProgress))
			} else if len(weekProgress) > 0 {
				group.HasProgress = append(group.HasProgress, toSubItemWeekDTO(si, weekProgress))
			} else {
				group.NoChangeFromLastWeek = append(group.NoChangeFromLastWeek, dto.SubItemSummaryDTO{
					ID:         si.ID,
					Title:      si.Title,
					Status:     si.Status,
					Completion: si.Completion,
				})
			}
		}

		groups = append(groups, group)
	}

	return &dto.WeeklyViewResult{
		WeekStart: weekStart.Format("2006-01-02"),
		WeekEnd:   weekEnd.Format("2006-01-02"),
		Groups:    groups,
	}, nil
}

// isNewlyCompleted checks if a sub-item's ActualEndDate falls within the week range.
func isNewlyCompleted(si model.SubItem, weekStart, weekEnd time.Time) bool {
	if si.Status != "已完成" {
		return false
	}
	if si.ActualEndDate == nil {
		return false
	}
	end := *si.ActualEndDate
	return !end.Before(weekStart) && !end.After(weekEnd)
}

// toSubItemWeekDTO converts a SubItem + its week progress records into a SubItemWeekDTO.
func toSubItemWeekDTO(si model.SubItem, records []model.ProgressRecord) dto.SubItemWeekDTO {
	progressDTOs := make([]dto.ProgressRecordDTO, 0, len(records))
	for _, pr := range records {
		progressDTOs = append(progressDTOs, dto.ProgressRecordDTO{
			ID:          pr.ID,
			Completion:  pr.Completion,
			Achievement: pr.Achievement,
			Blocker:     pr.Blocker,
			CreatedAt:   pr.CreatedAt.Format(time.RFC3339),
		})
	}

	return dto.SubItemWeekDTO{
		ID:               si.ID,
		Title:            si.Title,
		Status:           si.Status,
		Completion:       si.Completion,
		MainItemID:       si.MainItemID,
		ProgressThisWeek: progressDTOs,
	}
}

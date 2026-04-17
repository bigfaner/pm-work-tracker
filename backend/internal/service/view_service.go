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
	GanttView(ctx context.Context, teamID uint, filter dto.GanttFilter) (*dto.GanttResult, error)
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

func (s *viewService) GanttView(ctx context.Context, teamID uint, filter dto.GanttFilter) (*dto.GanttResult, error) {
	// Fetch all non-archived main items for the team
	mainItems, err := s.mainItemRepo.ListNonArchivedByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Apply status filter if provided
	if filter.Status != "" {
		filtered := make([]model.MainItem, 0, len(mainItems))
		for _, mi := range mainItems {
			if mi.Status == filter.Status {
				filtered = append(filtered, mi)
			}
		}
		mainItems = filtered
	}

	// Fetch all sub-items for the team (single query, avoid N+1)
	subItems, err := s.subItemRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Index sub-items by main item ID
	subItemsByMain := make(map[uint][]model.SubItem)
	for _, si := range subItems {
		subItemsByMain[si.MainItemID] = append(subItemsByMain[si.MainItemID], si)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	items := make([]dto.GanttMainItemDTO, 0, len(mainItems))
	for _, mi := range mainItems {
		subs := subItemsByMain[mi.ID]
		subDTOs := make([]dto.GanttSubItemDTO, 0, len(subs))
		for _, si := range subs {
			subDTOs = append(subDTOs, dto.GanttSubItemDTO{
				ID:              si.ID,
				Title:           si.Title,
				StartDate:       formatDate(si.StartDate),
				ExpectedEndDate: formatDate(si.ExpectedEndDate),
				Completion:      si.Completion,
				Status:          si.Status,
			})
		}

		items = append(items, dto.GanttMainItemDTO{
			ID:              mi.ID,
			Title:           mi.Title,
			Priority:        mi.Priority,
			StartDate:       formatDate(mi.StartDate),
			ExpectedEndDate: formatDate(mi.ExpectedEndDate),
			Completion:      mi.Completion,
			Status:          mi.Status,
			IsOverdue:       computeIsOverdue(mi.ExpectedEndDate, mi.Status, today),
			SubItems:        subDTOs,
		})
	}

	return &dto.GanttResult{Items: items}, nil
}

// formatDate returns an ISO8601 date string or empty string for nil.
func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// computeIsOverdue returns true if expectedEndDate is before today and status is not terminal.
func computeIsOverdue(expectedEndDate *time.Time, status string, today time.Time) bool {
	if expectedEndDate == nil {
		return false
	}
	if status == "已完成" || status == "已关闭" {
		return false
	}
	return expectedEndDate.Before(today)
}

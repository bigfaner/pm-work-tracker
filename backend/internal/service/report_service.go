package service

import (
	"context"
	"time"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/report"
	"pm-work-tracker/backend/internal/repository"
)

// ReportService defines report operations.
type ReportService interface {
	Preview(ctx context.Context, teamID uint, weekStart time.Time) (*dto.ReportPreview, error)
	ExportMarkdown(ctx context.Context, teamID uint, weekStart time.Time) ([]byte, error)
}

type reportService struct {
	mainItemRepo repository.MainItemRepo
	subItemRepo  repository.SubItemRepo
	progressRepo repository.ProgressRepo
}

// NewReportService creates a new ReportService.
func NewReportService(mainItemRepo repository.MainItemRepo, subItemRepo repository.SubItemRepo, progressRepo repository.ProgressRepo) ReportService {
	return &reportService{
		mainItemRepo: mainItemRepo,
		subItemRepo:  subItemRepo,
		progressRepo: progressRepo,
	}
}

func (s *reportService) Preview(ctx context.Context, teamID uint, weekStart time.Time) (*dto.ReportPreview, error) {
	weekEnd := weekStart.AddDate(0, 0, 6)

	// Fetch all non-archived main items
	mainItems, err := s.mainItemRepo.ListNonArchivedByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Fetch all sub-items for the team
	subItems, err := s.subItemRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Fetch progress records for the week range
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

	// Build sections
	var sections []dto.ReportSectionDTO
	for _, mi := range mainItems {
		subs, ok := subItemsByMain[mi.ID]
		if !ok {
			continue
		}

		var reportSubs []dto.ReportSubItemDTO
		for _, si := range subs {
			weekProgress := progressBySub[si.ID]
			if len(weekProgress) == 0 {
				continue // Skip sub-items with no progress this week
			}

			achievements := filterNonEmpty(weekProgress, func(pr model.ProgressRecord) string {
				return pr.Achievement
			})
			blockers := filterNonEmpty(weekProgress, func(pr model.ProgressRecord) string {
				return pr.Blocker
			})

			reportSubs = append(reportSubs, dto.ReportSubItemDTO{
				ID:           si.ID,
				Title:        si.Title,
				Completion:   si.Completion,
				Achievements: achievements,
				Blockers:     blockers,
			})
		}

		if len(reportSubs) == 0 {
			continue // Skip main items with no sub-item progress
		}

		sections = append(sections, dto.ReportSectionDTO{
			MainItem: dto.MainItemSummaryDTO{
				ID:         mi.ID,
				Title:      mi.Title,
				Completion: mi.Completion,
				IsKeyItem:  mi.IsKeyItem,
			},
			SubItems: reportSubs,
		})
	}

	if len(sections) == 0 {
		return nil, apperrors.ErrNoData
	}

	return &dto.ReportPreview{
		WeekStart: weekStart.Format("2006-01-02"),
		WeekEnd:   weekEnd.Format("2006-01-02"),
		Sections:  sections,
	}, nil
}

func (s *reportService) ExportMarkdown(ctx context.Context, teamID uint, weekStart time.Time) ([]byte, error) {
	preview, err := s.Preview(ctx, teamID, weekStart)
	if err != nil {
		return nil, err
	}

	return report.RenderMarkdown(preview, weekStart), nil
}

// filterNonEmpty extracts non-empty strings from records using the given extractor.
func filterNonEmpty(records []model.ProgressRecord, extract func(model.ProgressRecord) string) []string {
	var result []string
	for _, pr := range records {
		val := extract(pr)
		if val != "" {
			result = append(result, val)
		}
	}
	return result
}

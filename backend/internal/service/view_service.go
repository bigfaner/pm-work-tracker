package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
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
	TableView(ctx context.Context, teamID uint, filter dto.TableFilter, page dto.Pagination) (*dto.PageResult[dto.TableRow], error)
	TableExportCSV(ctx context.Context, teamID uint, filter dto.TableFilter) ([]byte, error)
}

type viewService struct {
	mainItemRepo repository.MainItemRepo
	subItemRepo  repository.SubItemRepo
	progressRepo repository.ProgressRepo
	userRepo     repository.UserRepo
}

// NewViewService creates a new ViewService.
func NewViewService(mainItemRepo repository.MainItemRepo, subItemRepo repository.SubItemRepo, progressRepo repository.ProgressRepo) ViewService {
	return &viewService{
		mainItemRepo: mainItemRepo,
		subItemRepo:  subItemRepo,
		progressRepo: progressRepo,
	}
}

// NewViewServiceWithUserRepo creates a new ViewService with user repo for table view.
func NewViewServiceWithUserRepo(mainItemRepo repository.MainItemRepo, subItemRepo repository.SubItemRepo, progressRepo repository.ProgressRepo, userRepo repository.UserRepo) ViewService {
	return &viewService{
		mainItemRepo: mainItemRepo,
		subItemRepo:  subItemRepo,
		progressRepo: progressRepo,
		userRepo:     userRepo,
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

func (s *viewService) TableView(ctx context.Context, teamID uint, filter dto.TableFilter, page dto.Pagination) (*dto.PageResult[dto.TableRow], error) {
	var rows []dto.TableRow

	// Fetch main items (non-archived only)
	if filter.Type == "" || filter.Type == "main" {
		mainItems, err := s.mainItemRepo.ListNonArchivedByTeam(ctx, teamID)
		if err != nil {
			return nil, err
		}
		for _, mi := range mainItems {
			if !matchesFilterMain(mi, filter) {
				continue
			}
			rows = append(rows, mainItemToRow(mi))
		}
	}

	// Fetch sub-items
	if filter.Type == "" || filter.Type == "sub" {
		subItems, err := s.subItemRepo.ListByTeam(ctx, teamID)
		if err != nil {
			return nil, err
		}
		for _, si := range subItems {
			if !matchesFilterSub(si, filter) {
				continue
			}
			rows = append(rows, subItemToRow(si))
		}
	}

	// Resolve assignee names
	resolveAssigneeNames(ctx, rows, s.userRepo)

	// Sort
	sortTableRows(rows, filter)

	// Paginate
	total := int64(len(rows))
	start := (page.Page - 1) * page.PageSize
	if start >= len(rows) {
		return &dto.PageResult[dto.TableRow]{
			Items: []dto.TableRow{},
			Total: total,
			Page:  page.Page,
			Size:  page.PageSize,
		}, nil
	}
	end := start + page.PageSize
	if end > len(rows) {
		end = len(rows)
	}

	return &dto.PageResult[dto.TableRow]{
		Items: rows[start:end],
		Total: total,
		Page:  page.Page,
		Size:  page.PageSize,
	}, nil
}

func (s *viewService) TableExportCSV(ctx context.Context, teamID uint, filter dto.TableFilter) ([]byte, error) {
	// Fetch all matching rows (no pagination)
	var rows []dto.TableRow

	if filter.Type == "" || filter.Type == "main" {
		mainItems, err := s.mainItemRepo.ListNonArchivedByTeam(ctx, teamID)
		if err != nil {
			return nil, err
		}
		for _, mi := range mainItems {
			if !matchesFilterMain(mi, filter) {
				continue
			}
			rows = append(rows, mainItemToRow(mi))
		}
	}

	if filter.Type == "" || filter.Type == "sub" {
		subItems, err := s.subItemRepo.ListByTeam(ctx, teamID)
		if err != nil {
			return nil, err
		}
		for _, si := range subItems {
			if !matchesFilterSub(si, filter) {
				continue
			}
			rows = append(rows, subItemToRow(si))
		}
	}

	if len(rows) == 0 {
		return nil, apperrors.ErrNoData
	}

	resolveAssigneeNames(ctx, rows, s.userRepo)
	sortTableRows(rows, filter)

	// Write CSV
	var buf bytes.Buffer
	// UTF-8 BOM for Excel compatibility
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(&buf)
	// Header
	header := []string{"编号", "标题", "类型", "优先级", "负责人", "状态", "完成度", "预期完成时间", "实际完成时间"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, row := range rows {
		record := []string{
			row.Code,
			row.Title,
			row.Type,
			row.Priority,
			row.AssigneeName,
			row.Status,
			fmt.Sprintf("%.0f", row.Completion),
			derefString(row.ExpectedEndDate),
			derefString(row.ActualEndDate),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func mainItemToRow(mi model.MainItem) dto.TableRow {
	return dto.TableRow{
		ID:              mi.ID,
		Type:            "main",
		Code:            mi.Code,
		Title:           mi.Title,
		Priority:        mi.Priority,
		AssigneeID:      mi.AssigneeID,
		Status:          mi.Status,
		Completion:      mi.Completion,
		ExpectedEndDate: formatDatePtr(mi.ExpectedEndDate),
		ActualEndDate:   formatDatePtr(mi.ActualEndDate),
	}
}

func subItemToRow(si model.SubItem) dto.TableRow {
	return dto.TableRow{
		ID:              si.ID,
		Type:            "sub",
		Code:            fmt.Sprintf("SI-%04d", si.ID),
		Title:           si.Title,
		Priority:        si.Priority,
		AssigneeID:      si.AssigneeID,
		Status:          si.Status,
		Completion:      si.Completion,
		ExpectedEndDate: formatDatePtr(si.ExpectedEndDate),
		ActualEndDate:   formatDatePtr(si.ActualEndDate),
	}
}

func formatDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func matchesFilterMain(mi model.MainItem, filter dto.TableFilter) bool {
	if len(filter.Priority) > 0 && !contains(filter.Priority, mi.Priority) {
		return false
	}
	if len(filter.Status) > 0 && !contains(filter.Status, mi.Status) {
		return false
	}
	if filter.AssigneeID != nil {
		if mi.AssigneeID == nil || *mi.AssigneeID != *filter.AssigneeID {
			return false
		}
	}
	return true
}

func matchesFilterSub(si model.SubItem, filter dto.TableFilter) bool {
	if len(filter.Priority) > 0 && !contains(filter.Priority, si.Priority) {
		return false
	}
	if len(filter.Status) > 0 && !contains(filter.Status, si.Status) {
		return false
	}
	if filter.AssigneeID != nil {
		if si.AssigneeID == nil || *si.AssigneeID != *filter.AssigneeID {
			return false
		}
	}
	return true
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// priorityOrder maps priority strings to sortable integers.
var priorityOrder = map[string]int{"P1": 1, "P2": 2, "P3": 3}

func sortTableRows(rows []dto.TableRow, filter dto.TableFilter) {
	sort.SliceStable(rows, func(i, j int) bool {
		switch filter.SortBy {
		case "completion":
			return compareWithOrder(rows[i].Completion, rows[j].Completion, filter.SortOrder)
		case "status":
			return compareWithOrder(rows[i].Status, rows[j].Status, filter.SortOrder)
		case "expectedEndDate":
			return compareDatePtrWithOrder(rows[i].ExpectedEndDate, rows[j].ExpectedEndDate, filter.SortOrder)
		case "actualEndDate":
			return compareDatePtrWithOrder(rows[i].ActualEndDate, rows[j].ActualEndDate, filter.SortOrder)
		case "priority":
			pi, oki := priorityOrder[rows[i].Priority]
			pj, okj := priorityOrder[rows[j].Priority]
			if !oki {
				pi = 99
			}
			if !okj {
				pj = 99
			}
			return compareWithOrder(pi, pj, filter.SortOrder)
		default:
			// Default sort: priority DESC, then expectedEndDate ASC
			pi, oki := priorityOrder[rows[i].Priority]
			pj, okj := priorityOrder[rows[j].Priority]
			if !oki {
				pi = 99
			}
			if !okj {
				pj = 99
			}
			if pi != pj {
				return pi < pj // lower number = higher priority = first (DESC)
			}
			// Secondary sort: expectedEndDate ASC
			return compareDatePtr(rows[i].ExpectedEndDate, rows[j].ExpectedEndDate)
		}
	})
}

func compareWithOrder[T int | float64 | string](a, b T, order string) bool {
	if order == "desc" {
		return a > b
	}
	return a < b
}

func compareDatePtrWithOrder(a, b *string, order string) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return order != "desc" // nil goes last in asc, first in desc
	}
	if b == nil {
		return order == "desc"
	}
	if order == "desc" {
		return *a > *b
	}
	return *a < *b
}

func compareDatePtr(a, b *string) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return false // nil goes last
	}
	if b == nil {
		return true
	}
	return *a < *b
}

func resolveAssigneeNames(ctx context.Context, rows []dto.TableRow, userRepo repository.UserRepo) {
	if userRepo == nil {
		return
	}
	// Collect unique assignee IDs
	assigneeIDs := make(map[uint]struct{})
	for _, row := range rows {
		if row.AssigneeID != nil {
			assigneeIDs[*row.AssigneeID] = struct{}{}
		}
	}

	// Resolve names
	names := make(map[uint]string)
	for id := range assigneeIDs {
		user, err := userRepo.FindByID(ctx, id)
		if err == nil && user != nil {
			names[id] = user.DisplayName
		}
	}

	// Fill names into rows
	for i := range rows {
		if rows[i].AssigneeID != nil {
			rows[i].AssigneeName = names[*rows[i].AssigneeID]
		}
	}
}

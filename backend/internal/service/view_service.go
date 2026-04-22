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
	"pm-work-tracker/backend/internal/pkg/dates"
	"pm-work-tracker/backend/internal/repository"
)

// ViewService defines read-only view operations.
type ViewService interface {
	WeeklyComparison(ctx context.Context, teamID uint, weekStart time.Time) (*dto.WeeklyViewResponse, error)
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

func (s *viewService) WeeklyComparison(ctx context.Context, teamID uint, weekStart time.Time) (*dto.WeeklyViewResponse, error) {
	if weekStart.Weekday() != time.Monday {
		return nil, apperrors.ErrValidation
	}

	// Future week not allowed
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if weekStart.After(today) {
		return nil, apperrors.ErrFutureWeekNotAllowed
	}

	weekEnd := weekStart.AddDate(0, 0, 6)
	lastWeekStart := weekStart.AddDate(0, 0, -7)
	lastWeekEnd := lastWeekStart.AddDate(0, 0, 6)

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

	// Fetch progress records for both weeks
	thisWeekRangeStart := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	thisWeekRangeEnd := time.Date(weekEnd.Year(), weekEnd.Month(), weekEnd.Day(), 0, 0, 0, 0, weekEnd.Location()).AddDate(0, 0, 1)

	lastWeekRangeStart := time.Date(lastWeekStart.Year(), lastWeekStart.Month(), lastWeekStart.Day(), 0, 0, 0, 0, lastWeekStart.Location())
	lastWeekRangeEnd := time.Date(lastWeekEnd.Year(), lastWeekEnd.Month(), lastWeekEnd.Day(), 0, 0, 0, 0, lastWeekEnd.Location()).AddDate(0, 0, 1)

	// Fetch progress for both weeks in a single extended range
	combinedStart := lastWeekRangeStart
	combinedEnd := thisWeekRangeEnd
	allProgress, err := s.progressRepo.ListByTeamInRange(ctx, teamID, combinedStart, combinedEnd)
	if err != nil {
		return nil, err
	}

	// Split progress records into last week and this week
	lastWeekProgress := make(map[uint][]model.ProgressRecord)
	thisWeekProgress := make(map[uint][]model.ProgressRecord)
	for _, pr := range allProgress {
		if !pr.CreatedAt.Before(lastWeekRangeStart) && pr.CreatedAt.Before(lastWeekRangeEnd) {
			lastWeekProgress[pr.SubItemID] = append(lastWeekProgress[pr.SubItemID], pr)
		}
		if !pr.CreatedAt.Before(thisWeekRangeStart) && pr.CreatedAt.Before(thisWeekRangeEnd) {
			thisWeekProgress[pr.SubItemID] = append(thisWeekProgress[pr.SubItemID], pr)
		}
	}

	// Index sub-items by main item ID
	subItemsByMain := make(map[uint][]model.SubItem)
	for _, si := range subItems {
		subItemsByMain[si.MainItemID] = append(subItemsByMain[si.MainItemID], si)
	}

	// Determine which sub-items were active in each week.
	// A sub-item is "active" in a week if:
	//   1. It was created before or during that week (CreatedAt < weekEnd+1day)
	//   2. It was not completed before that week (ActualEndDate == nil || ActualEndDate >= weekStart)
	//   3. OR it had progress records in that week
	lastWeekActive := make(map[uint]struct{})
	thisWeekActive := make(map[uint]struct{})

	for _, si := range subItems {
		if _, ok := lastWeekProgress[si.ID]; ok {
			lastWeekActive[si.ID] = struct{}{}
		}
		if _, ok := thisWeekProgress[si.ID]; ok {
			thisWeekActive[si.ID] = struct{}{}
		}
		// Check if sub-item existed and was not completed before each week
		if isActiveInWeek(si, lastWeekStart, lastWeekEnd) {
			lastWeekActive[si.ID] = struct{}{}
		}
		if isActiveInWeek(si, weekStart, weekEnd) {
			thisWeekActive[si.ID] = struct{}{}
		}
	}

	// Get latest progress description for this week per sub-item
	latestProgressDesc := make(map[uint]string)
	for subID, records := range thisWeekProgress {
		if len(records) > 0 {
			latest := records[len(records)-1]
			desc := latest.Achievement
			if latest.Blocker != "" {
				if desc != "" {
					desc += "; "
				}
				desc += latest.Blocker
			}
			latestProgressDesc[subID] = desc
		}
	}

	// Resolve assignee names for sub-items (batch)
	assigneeNames := make(map[uint]string)
	if s.userRepo != nil {
		seen := make(map[uint]struct{})
		var ids []uint
		for _, si := range subItems {
			if si.AssigneeID != nil {
				if _, ok := seen[*si.AssigneeID]; !ok {
					seen[*si.AssigneeID] = struct{}{}
					ids = append(ids, *si.AssigneeID)
				}
			}
		}
		if len(ids) > 0 {
			users, err := s.userRepo.FindByIDs(ctx, ids)
			if err == nil {
				for id, u := range users {
					assigneeNames[id] = u.DisplayName
				}
			}
		}
	}

	// Build groups
	var groups []dto.WeeklyComparisonGroup
	var totalActive, totalNewlyCompleted, totalInProgress, totalBlocked int

	for _, mi := range mainItems {
		subs, ok := subItemsByMain[mi.ID]
		if !ok || len(subs) == 0 {
			continue
		}

		group := dto.WeeklyComparisonGroup{
			MainItem: dto.WeeklyMainItemSummary{
				ID:              mi.ID,
				Code:            mi.Code,
				Title:           mi.Title,
				Priority:        mi.Priority,
				Status:          mi.Status,
				StartDate:       formatDate(mi.StartDate),
				ExpectedEndDate: formatDate(mi.ExpectedEndDate),
				ActualEndDate:   dates.FormatTimePtr(mi.ActualEndDate),
				Completion:      mi.Completion,
				SubItemCount:    len(subs),
			},
		}

		// Get last week completion per sub-item (latest record in last week)
		lastWeekCompletion := make(map[uint]float64)
		for subID, records := range lastWeekProgress {
			if len(records) > 0 {
				lastWeekCompletion[subID] = records[len(records)-1].Completion
			}
		}

		for _, si := range subs {
			assigneeName := ""
			if si.AssigneeID != nil {
				assigneeName = assigneeNames[*si.AssigneeID]
			}

			inThisWeek := false
			if _, ok := thisWeekActive[si.ID]; ok {
				inThisWeek = true
			}
			if _, ok := thisWeekProgress[si.ID]; ok {
				inThisWeek = true
			}

			// Check if just completed this week
			isJustCompleted := isNewlyCompleted(si, weekStart, weekEnd)

			if isJustCompleted {
				totalNewlyCompleted++
			}

			snapshot := dto.SubItemSnapshot{
				ID:                  si.ID,
				Title:               si.Title,
				Priority:            si.Priority,
				Status:              si.Status,
				AssigneeName:        assigneeName,
				StartDate:           formatDate(si.StartDate),
				ExpectedEndDate:     formatDate(si.ExpectedEndDate),
				ActualEndDate:       dates.FormatTimePtr(si.ActualEndDate),
				Completion:          si.Completion,
				ProgressDescription: latestProgressDesc[si.ID],
			}

			// Populate individual progress records for this week
			if records, ok := thisWeekProgress[si.ID]; ok && len(records) > 0 {
				snapshot.ProgressRecords = make([]dto.ProgressRecordDTO, 0, len(records))
				for _, pr := range records {
					snapshot.ProgressRecords = append(snapshot.ProgressRecords, dto.ProgressRecordDTO{
						ID:          pr.ID,
						Completion:  pr.Completion,
						Achievement: pr.Achievement,
						Blocker:     pr.Blocker,
						CreatedAt:   pr.CreatedAt.Format(time.RFC3339),
					})
				}
			}

			if isJustCompleted {
				snapshot.JustCompleted = true
			}

			// Compute delta
			if lastComp, existed := lastWeekCompletion[si.ID]; existed {
				snapshot.Delta = si.Completion - lastComp
			} else {
				// Was not active last week
				snapshot.IsNew = true
				snapshot.Delta = 0
			}

			// Add to lastWeek if it was active last week
			if _, wasActiveLastWeek := lastWeekActive[si.ID]; wasActiveLastWeek {
				if _, hadProgress := lastWeekProgress[si.ID]; hadProgress {
					lastComp := lastWeekCompletion[si.ID]
					lastSnapshot := dto.SubItemSnapshot{
						ID:              si.ID,
						Title:           si.Title,
						Priority:        si.Priority,
						Status:          si.Status,
						AssigneeName:    assigneeName,
						StartDate:       formatDate(si.StartDate),
						ExpectedEndDate: formatDate(si.ExpectedEndDate),
						ActualEndDate:   dates.FormatTimePtr(si.ActualEndDate),
						Completion:      lastComp,
					}
					group.LastWeek = append(group.LastWeek, lastSnapshot)
				}
			}

			// Completed items with no change this week -> completedNoChange
			// Only include if the sub-item existed during this week AND was completed at or before this week.
			if si.Status == "completed" && !isJustCompleted {
				completedBeforeOrDuringWeek := si.ActualEndDate == nil || !si.ActualEndDate.After(weekEnd)
				existedDuringWeek := si.CreatedAt.Before(weekEnd.AddDate(0, 0, 1))
				if completedBeforeOrDuringWeek && existedDuringWeek {
					group.CompletedNoChange = append(group.CompletedNoChange, snapshot)
				}
				continue
			}

			// Add to thisWeek if active this week
			if inThisWeek || isJustCompleted {
				group.ThisWeek = append(group.ThisWeek, snapshot)
				totalActive++

				if si.Status == "progressing" {
					totalInProgress++
				}
				if si.Status == "blocking" {
					totalBlocked++
				}
			}
		}

		// Skip main items with no active sub-items in this week's view
		if len(group.ThisWeek) == 0 && len(group.LastWeek) == 0 && len(group.CompletedNoChange) == 0 {
			continue
		}

		groups = append(groups, group)
	}

	// Sort groups by main item priority ascending
	sort.SliceStable(groups, func(i, j int) bool {
		pi, oki := priorityOrder[groups[i].MainItem.Priority]
		pj, okj := priorityOrder[groups[j].MainItem.Priority]
		if !oki {
			pi = 99
		}
		if !okj {
			pj = 99
		}
		return pi < pj
	})

	return &dto.WeeklyViewResponse{
		WeekStart: weekStart.Format("2006-01-02"),
		WeekEnd:   weekEnd.Format("2006-01-02"),
		Stats: dto.WeeklyStats{
			ActiveSubItems: totalActive,
			NewlyCompleted: totalNewlyCompleted,
			InProgress:     totalInProgress,
			Blocked:        totalBlocked,
		},
		Groups: groups,
	}, nil
}

// isNewlyCompleted checks if a sub-item's ActualEndDate falls within the week range.
func isNewlyCompleted(si model.SubItem, weekStart, weekEnd time.Time) bool {
	if si.Status != "completed" {
		return false
	}
	if si.ActualEndDate == nil {
		return false
	}
	end := *si.ActualEndDate
	return !end.Before(weekStart) && !end.After(weekEnd)
}

// isActiveInWeek checks if a sub-item was active during a given week.
// Active means: created before the week ends, and not completed before the week starts.
func isActiveInWeek(si model.SubItem, weekStart, weekEnd time.Time) bool {
	weekEndNextDay := weekEnd.AddDate(0, 0, 1)
	if si.CreatedAt.After(weekEndNextDay) || si.CreatedAt.Equal(weekEndNextDay) {
		return false
	}
	if si.ActualEndDate != nil && si.ActualEndDate.Before(weekStart) {
		return false
	}
	return true
}

func (s *viewService) GanttView(ctx context.Context, teamID uint, filter dto.GanttFilter) (*dto.GanttResult, error) {
	// Fetch main items: use SQL pushdown for status filter, otherwise fetch all non-archived
	var mainItems []model.MainItem
	var err error
	if filter.Status != "" {
		mainItems, err = s.mainItemRepo.ListByTeamAndStatus(ctx, teamID, filter.Status)
	} else {
		mainItems, err = s.mainItemRepo.ListNonArchivedByTeam(ctx, teamID)
	}
	if err != nil {
		return nil, err
	}

	// Sort main items: priority ASC (P1 first), then created_at ASC
	sort.SliceStable(mainItems, func(i, j int) bool {
		pi, oki := priorityOrder[mainItems[i].Priority]
		pj, okj := priorityOrder[mainItems[j].Priority]
		if !oki {
			pi = 99
		}
		if !okj {
			pj = 99
		}
		if pi != pj {
			return pi < pj
		}
		return mainItems[i].CreatedAt.Before(mainItems[j].CreatedAt)
	})

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
	if status == "completed" || status == "closed" {
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
	header := []string{"编号", "标题", "类型", "优先级", "负责人", "状态", "完成度", "预期完成时间", "结束时间"}
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
		ExpectedEndDate: dates.FormatTimePtr(mi.ExpectedEndDate),
		ActualEndDate:   dates.FormatTimePtr(mi.ActualEndDate),
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
		ExpectedEndDate: dates.FormatTimePtr(si.ExpectedEndDate),
		ActualEndDate:   dates.FormatTimePtr(si.ActualEndDate),
	}
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
	if len(assigneeIDs) == 0 {
		return
	}

	// Batch resolve names with a single FindByIDs call
	ids := make([]uint, 0, len(assigneeIDs))
	for id := range assigneeIDs {
		ids = append(ids, id)
	}
	users, err := userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return
	}
	names := make(map[uint]string, len(users))
	for id, u := range users {
		names[id] = u.DisplayName
	}

	// Fill names into rows
	for i := range rows {
		if rows[i].AssigneeID != nil {
			rows[i].AssigneeName = names[*rows[i].AssigneeID]
		}
	}
}

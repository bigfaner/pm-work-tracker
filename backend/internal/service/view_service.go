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
	if err := validateWeekStart(weekStart); err != nil {
		return nil, err
	}

	weekEnd := weekStart.AddDate(0, 0, 6)
	lastWeekStart := weekStart.AddDate(0, 0, -7)

	mainItems, subItems, allProgress, err := s.fetchWeeklyData(ctx, teamID, lastWeekStart, weekEnd)
	if err != nil {
		return nil, err
	}

	subItemsByMain := indexSubItemsByMain(subItems)
	lastWeekProgress, thisWeekProgress := splitProgressByWeek(allProgress, lastWeekStart, weekStart, weekEnd)
	lastWeekActive, thisWeekActive := computeActiveSubItems(subItems, lastWeekProgress, thisWeekProgress, lastWeekStart, weekStart, weekEnd)
	lastWeekCompletion := computeLastWeekCompletion(lastWeekProgress)
	latestProgressDesc := computeLatestProgressDesc(thisWeekProgress)
	assigneeNames := resolveSubItemAssigneeNames(ctx, subItems, s.userRepo)

	groups, stats := buildWeeklyGroups(mainItems, subItemsByMain, lastWeekActive, thisWeekActive,
		lastWeekProgress, thisWeekProgress, lastWeekCompletion, latestProgressDesc, assigneeNames, weekStart, weekEnd)

	sortGroupsByPriority(groups)

	return &dto.WeeklyViewResponse{
		WeekStart: weekStart.Format("2006-01-02"),
		WeekEnd:   weekEnd.Format("2006-01-02"),
		Stats:     stats,
		Groups:    groups,
	}, nil
}

// validateWeekStart checks that weekStart is a Monday and not in the future.
func validateWeekStart(weekStart time.Time) error {
	if weekStart.Weekday() != time.Monday {
		return apperrors.ErrValidation
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if weekStart.After(today) {
		return apperrors.ErrFutureWeekNotAllowed
	}
	return nil
}

// fetchWeeklyData fetches main items, sub-items, and progress records for the two-week window.
func (s *viewService) fetchWeeklyData(ctx context.Context, teamID uint, lastWeekStart, weekEnd time.Time) (
	[]model.MainItem, []model.SubItem, []model.ProgressRecord, error,
) {
	mainItems, err := s.mainItemRepo.ListNonArchivedByTeam(ctx, teamID)
	if err != nil {
		return nil, nil, nil, err
	}
	subItems, err := s.subItemRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, nil, nil, err
	}
	combinedStart := startOfDay(lastWeekStart)
	combinedEnd := startOfDay(weekEnd).AddDate(0, 0, 1)
	allProgress, err := s.progressRepo.ListByTeamInRange(ctx, teamID, combinedStart, combinedEnd)
	if err != nil {
		return nil, nil, nil, err
	}
	return mainItems, subItems, allProgress, nil
}

// splitProgressByWeek partitions progress records into last-week and this-week buckets by sub-item ID.
func splitProgressByWeek(allProgress []model.ProgressRecord, lastWeekStart, thisWeekStart, weekEnd time.Time) (
	map[uint][]model.ProgressRecord, map[uint][]model.ProgressRecord,
) {
	lastWeekEnd := lastWeekStart.AddDate(0, 0, 7) // exclusive upper bound
	thisWeekEnd := thisWeekStart.AddDate(0, 0, 7) // exclusive upper bound

	lastWeek := make(map[uint][]model.ProgressRecord)
	thisWeek := make(map[uint][]model.ProgressRecord)
	for _, pr := range allProgress {
		if !pr.CreatedAt.Before(lastWeekStart) && pr.CreatedAt.Before(lastWeekEnd) {
			lastWeek[pr.SubItemID] = append(lastWeek[pr.SubItemID], pr)
		}
		if !pr.CreatedAt.Before(thisWeekStart) && pr.CreatedAt.Before(thisWeekEnd) {
			thisWeek[pr.SubItemID] = append(thisWeek[pr.SubItemID], pr)
		}
	}
	return lastWeek, thisWeek
}

// computeActiveSubItems determines which sub-items were active in each week.
func computeActiveSubItems(subItems []model.SubItem, lastWeekProgress, thisWeekProgress map[uint][]model.ProgressRecord, lastWeekStart, weekStart, weekEnd time.Time) (
	map[uint]struct{}, map[uint]struct{},
) {
	lastWeekEnd := lastWeekStart.AddDate(0, 0, 6)
	lastWeekActive := make(map[uint]struct{})
	thisWeekActive := make(map[uint]struct{})

	for _, si := range subItems {
		if _, ok := lastWeekProgress[si.ID]; ok {
			lastWeekActive[si.ID] = struct{}{}
		}
		if _, ok := thisWeekProgress[si.ID]; ok {
			thisWeekActive[si.ID] = struct{}{}
		}
		if isActiveInWeek(si, lastWeekStart, lastWeekEnd) {
			lastWeekActive[si.ID] = struct{}{}
		}
		if isActiveInWeek(si, weekStart, weekEnd) {
			thisWeekActive[si.ID] = struct{}{}
		}
	}
	return lastWeekActive, thisWeekActive
}

// computeLastWeekCompletion extracts the latest completion value per sub-item from last week's progress.
func computeLastWeekCompletion(lastWeekProgress map[uint][]model.ProgressRecord) map[uint]float64 {
	result := make(map[uint]float64)
	for subID, records := range lastWeekProgress {
		if len(records) > 0 {
			result[subID] = records[len(records)-1].Completion
		}
	}
	return result
}

// computeLatestProgressDesc builds the progress description from the latest this-week record per sub-item.
func computeLatestProgressDesc(thisWeekProgress map[uint][]model.ProgressRecord) map[uint]string {
	result := make(map[uint]string)
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
			result[subID] = desc
		}
	}
	return result
}

// resolveSubItemAssigneeNames batch-resolves assignee display names for sub-items.
func resolveSubItemAssigneeNames(ctx context.Context, subItems []model.SubItem, userRepo repository.UserRepo) map[uint]string {
	names := make(map[uint]string)
	if userRepo == nil {
		return names
	}
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
	if len(ids) == 0 {
		return names
	}
	users, err := userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return names
	}
	for id, u := range users {
		names[id] = u.DisplayName
	}
	return names
}

// indexSubItemsByMain groups sub-items by their main item ID.
func indexSubItemsByMain(subItems []model.SubItem) map[uint][]model.SubItem {
	result := make(map[uint][]model.SubItem)
	for _, si := range subItems {
		result[si.MainItemID] = append(result[si.MainItemID], si)
	}
	return result
}

// buildSubItemSnapshot creates a SubItemSnapshot for a sub-item with optional progress and delta info.
func buildSubItemSnapshot(si model.SubItem, assigneeName, progressDesc string, thisWeekProgress map[uint][]model.ProgressRecord, lastWeekCompletion map[uint]float64, weekStart, weekEnd time.Time) dto.SubItemSnapshot {
	snapshot := dto.SubItemSnapshot{
		ID:                  si.ID,
		Code:                si.Code,
		Title:               si.Title,
		Priority:            si.Priority,
		Status:              si.Status,
		AssigneeName:        assigneeName,
		StartDate:           formatDate(si.StartDate),
		ExpectedEndDate:     formatDate(si.ExpectedEndDate),
		ActualEndDate:       dates.FormatTimePtr(si.ActualEndDate),
		Completion:          si.Completion,
		ProgressDescription: progressDesc,
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

	if isNewlyCompleted(si, weekStart, weekEnd) {
		snapshot.JustCompleted = true
	}

	// Compute delta
	if lastComp, existed := lastWeekCompletion[si.ID]; existed {
		snapshot.Delta = si.Completion - lastComp
	} else {
		snapshot.IsNew = true
		snapshot.Delta = 0
	}

	return snapshot
}

// buildWeeklyGroups assembles the comparison groups and computes aggregate stats.
func buildWeeklyGroups(
	mainItems []model.MainItem,
	subItemsByMain map[uint][]model.SubItem,
	lastWeekActive, thisWeekActive map[uint]struct{},
	lastWeekProgress, thisWeekProgress map[uint][]model.ProgressRecord,
	lastWeekCompletion map[uint]float64,
	latestProgressDesc, assigneeNames map[uint]string,
	weekStart, weekEnd time.Time,
) ([]dto.WeeklyComparisonGroup, dto.WeeklyStats) {
	var groups []dto.WeeklyComparisonGroup
	var stats dto.WeeklyStats

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

		for _, si := range subs {
			assigneeName := ""
			if si.AssigneeID != nil {
				assigneeName = assigneeNames[*si.AssigneeID]
			}

			inThisWeek := isInThisWeek(si.ID, thisWeekActive, thisWeekProgress)
			justCompleted := isNewlyCompleted(si, weekStart, weekEnd)
			if justCompleted {
				stats.NewlyCompleted++
			}

			snapshot := buildSubItemSnapshot(si, assigneeName, latestProgressDesc[si.ID], thisWeekProgress, lastWeekCompletion, weekStart, weekEnd)

			appendLastWeekSnapshot(&group, si, assigneeName, lastWeekActive, lastWeekProgress, lastWeekCompletion)

			if isCompletedNoChange(si, justCompleted, weekEnd) {
				group.CompletedNoChange = append(group.CompletedNoChange, snapshot)
				continue
			}

			if inThisWeek || justCompleted {
				group.ThisWeek = append(group.ThisWeek, snapshot)
				stats.ActiveSubItems++
				if si.Status == "progressing" {
					stats.InProgress++
				}
				if si.Status == "blocking" {
					stats.Blocked++
				}
				if si.Status == "pending" {
					stats.Pending++
				}
				if si.Status == "pausing" {
					stats.Pausing++
				}
				if si.ExpectedEndDate != nil && si.ExpectedEndDate.Before(weekEnd) && si.Status != "completed" && si.Status != "closed" {
					stats.Overdue++
				}
			}
		}

		if len(group.ThisWeek) == 0 && len(group.LastWeek) == 0 && len(group.CompletedNoChange) == 0 {
			continue
		}
		groups = append(groups, group)
	}
	return groups, stats
}

// isInThisWeek returns true if the sub-item was active or had progress this week.
func isInThisWeek(subID uint, thisWeekActive map[uint]struct{}, thisWeekProgress map[uint][]model.ProgressRecord) bool {
	if _, ok := thisWeekActive[subID]; ok {
		return true
	}
	if _, ok := thisWeekProgress[subID]; ok {
		return true
	}
	return false
}

// appendLastWeekSnapshot adds a last-week snapshot to the group if the sub-item was active and had progress last week.
func appendLastWeekSnapshot(group *dto.WeeklyComparisonGroup, si model.SubItem, assigneeName string, lastWeekActive map[uint]struct{}, lastWeekProgress map[uint][]model.ProgressRecord, lastWeekCompletion map[uint]float64) {
	if _, wasActive := lastWeekActive[si.ID]; !wasActive {
		return
	}
	if _, hadProgress := lastWeekProgress[si.ID]; !hadProgress {
		return
	}
	lastSnapshot := dto.SubItemSnapshot{
		ID:              si.ID,
		Code:            si.Code,
		Title:           si.Title,
		Priority:        si.Priority,
		Status:          si.Status,
		AssigneeName:    assigneeName,
		StartDate:       formatDate(si.StartDate),
		ExpectedEndDate: formatDate(si.ExpectedEndDate),
		ActualEndDate:   dates.FormatTimePtr(si.ActualEndDate),
		Completion:      lastWeekCompletion[si.ID],
	}
	group.LastWeek = append(group.LastWeek, lastSnapshot)
}

// isCompletedNoChange returns true if a sub-item was completed before this week and not just completed.
func isCompletedNoChange(si model.SubItem, justCompleted bool, weekEnd time.Time) bool {
	if si.Status != "completed" || justCompleted {
		return false
	}
	completedBeforeOrDuringWeek := si.ActualEndDate == nil || !si.ActualEndDate.After(weekEnd)
	existedDuringWeek := si.CreatedAt.Before(weekEnd.AddDate(0, 0, 1))
	return completedBeforeOrDuringWeek && existedDuringWeek
}

// sortGroupsByPriority sorts groups by main item priority ascending (P1 first).
func sortGroupsByPriority(groups []dto.WeeklyComparisonGroup) {
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
}

// startOfDay truncates a time to midnight in its location.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
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
		Code:            si.Code,
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

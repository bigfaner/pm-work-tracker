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
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/pkg/dates"
	"pm-work-tracker/backend/internal/repository"
)

// ViewService defines read-only view operations.
type ViewService interface {
	WeeklyComparison(ctx context.Context, teamBizKey int64, weekStart time.Time) (*dto.WeeklyViewResponse, error)
	GanttView(ctx context.Context, teamBizKey int64, filter dto.GanttFilter) (*dto.GanttResult, error)
	TableView(ctx context.Context, teamBizKey int64, filter dto.TableFilter, page dto.Pagination) (*dto.PageResult[dto.TableRow], error)
	TableExportCSV(ctx context.Context, teamBizKey int64, filter dto.TableFilter) ([]byte, error)
}

type viewService struct {
	mainItemRepo repository.MainItemRepo
	subItemRepo  repository.SubItemRepo
	progressRepo repository.ProgressRepo
	userRepo     repository.UserRepo
}

// NewViewService creates a new ViewService. userRepo is optional; when provided,
// it enables assignee name resolution in table view and weekly comparison.
func NewViewService(mainItemRepo repository.MainItemRepo, subItemRepo repository.SubItemRepo, progressRepo repository.ProgressRepo, userRepo ...repository.UserRepo) ViewService {
	var ur repository.UserRepo
	if len(userRepo) > 0 {
		ur = userRepo[0]
	}
	return &viewService{
		mainItemRepo: mainItemRepo,
		subItemRepo:  subItemRepo,
		progressRepo: progressRepo,
		userRepo:     ur,
	}
}

func (s *viewService) WeeklyComparison(ctx context.Context, teamBizKey int64, weekStart time.Time) (*dto.WeeklyViewResponse, error) {
	if err := validateWeekStart(weekStart); err != nil {
		return nil, err
	}

	weekEnd := weekStart.AddDate(0, 0, 6)
	lastWeekStart := weekStart.AddDate(0, 0, -7)

	mainItems, subItems, allProgress, err := s.fetchWeeklyData(ctx, teamBizKey, lastWeekStart, weekEnd)
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
func (s *viewService) fetchWeeklyData(ctx context.Context, teamBizKey int64, lastWeekStart, weekEnd time.Time) (
	[]model.MainItem, []model.SubItem, []model.ProgressRecord, error,
) {
	mainItems, err := s.mainItemRepo.ListNonArchivedByTeam(ctx, teamBizKey)
	if err != nil {
		return nil, nil, nil, err
	}
	subItems, err := s.subItemRepo.ListByTeam(ctx, teamBizKey)
	if err != nil {
		return nil, nil, nil, err
	}
	combinedStart := startOfDay(lastWeekStart)
	combinedEnd := startOfDay(weekEnd).AddDate(0, 0, 1)
	allProgress, err := s.progressRepo.ListByTeamInRange(ctx, teamBizKey, combinedStart, combinedEnd)
	if err != nil {
		return nil, nil, nil, err
	}
	return mainItems, subItems, allProgress, nil
}

// splitProgressByWeek partitions progress records into last-week and this-week buckets by sub-item biz_key.
func splitProgressByWeek(allProgress []model.ProgressRecord, lastWeekStart, thisWeekStart, weekEnd time.Time) (
	map[int64][]model.ProgressRecord, map[int64][]model.ProgressRecord,
) {
	lastWeekEnd := lastWeekStart.AddDate(0, 0, 7) // exclusive upper bound
	thisWeekEnd := thisWeekStart.AddDate(0, 0, 7) // exclusive upper bound

	lastWeek := make(map[int64][]model.ProgressRecord)
	thisWeek := make(map[int64][]model.ProgressRecord)
	for _, pr := range allProgress {
		if !pr.CreateTime.Before(lastWeekStart) && pr.CreateTime.Before(lastWeekEnd) {
			lastWeek[pr.SubItemKey] = append(lastWeek[pr.SubItemKey], pr)
		}
		if !pr.CreateTime.Before(thisWeekStart) && pr.CreateTime.Before(thisWeekEnd) {
			thisWeek[pr.SubItemKey] = append(thisWeek[pr.SubItemKey], pr)
		}
	}
	return lastWeek, thisWeek
}

// computeActiveSubItems determines which sub-items were active in each week.
func computeActiveSubItems(subItems []model.SubItem, lastWeekProgress, thisWeekProgress map[int64][]model.ProgressRecord, lastWeekStart, weekStart, weekEnd time.Time) (
	map[int64]struct{}, map[int64]struct{},
) {
	lastWeekEnd := lastWeekStart.AddDate(0, 0, 6)
	lastWeekActive := make(map[int64]struct{})
	thisWeekActive := make(map[int64]struct{})

	for _, si := range subItems {
		if _, ok := lastWeekProgress[si.BizKey]; ok {
			lastWeekActive[si.BizKey] = struct{}{}
		}
		if _, ok := thisWeekProgress[si.BizKey]; ok {
			thisWeekActive[si.BizKey] = struct{}{}
		}
		if isActiveInWeek(si, lastWeekStart, lastWeekEnd) {
			lastWeekActive[si.BizKey] = struct{}{}
		}
		if isActiveInWeek(si, weekStart, weekEnd) {
			thisWeekActive[si.BizKey] = struct{}{}
		}
	}
	return lastWeekActive, thisWeekActive
}

// computeLastWeekCompletion extracts the latest completion value per sub-item from last week's progress.
func computeLastWeekCompletion(lastWeekProgress map[int64][]model.ProgressRecord) map[int64]float64 {
	result := make(map[int64]float64)
	for subBizKey, records := range lastWeekProgress {
		if len(records) > 0 {
			result[subBizKey] = records[len(records)-1].Completion
		}
	}
	return result
}

// computeLatestProgressDesc builds the progress description from the latest this-week record per sub-item.
func computeLatestProgressDesc(thisWeekProgress map[int64][]model.ProgressRecord) map[int64]string {
	result := make(map[int64]string)
	for subBizKey, records := range thisWeekProgress {
		if len(records) > 0 {
			latest := records[len(records)-1]
			desc := latest.Achievement
			if latest.Blocker != "" {
				if desc != "" {
					desc += "; "
				}
				desc += latest.Blocker
			}
			result[subBizKey] = desc
		}
	}
	return result
}

// resolveSubItemAssigneeNames batch-resolves assignee display names for sub-items.
func resolveSubItemAssigneeNames(ctx context.Context, subItems []model.SubItem, userRepo repository.UserRepo) map[int64]string {
	names := make(map[int64]string)
	if userRepo == nil {
		return names
	}
	seen := make(map[int64]struct{})
	var bizKeys []int64
	for _, si := range subItems {
		if si.AssigneeKey != nil {
			if _, ok := seen[*si.AssigneeKey]; !ok {
				seen[*si.AssigneeKey] = struct{}{}
				bizKeys = append(bizKeys, *si.AssigneeKey)
			}
		}
	}
	if len(bizKeys) == 0 {
		return names
	}
	users, err := userRepo.FindByBizKeys(ctx, bizKeys)
	if err != nil {
		return names
	}
	for bizKey, u := range users {
		names[bizKey] = u.DisplayName
	}
	return names
}

// indexSubItemsByMain groups sub-items by their main item BizKey.
func indexSubItemsByMain(subItems []model.SubItem) map[int64][]model.SubItem {
	result := make(map[int64][]model.SubItem)
	for _, si := range subItems {
		result[si.MainItemKey] = append(result[si.MainItemKey], si)
	}
	return result
}

// buildSubItemSnapshot creates a SubItemSnapshot for a sub-item with optional progress and delta info.
func buildSubItemSnapshot(si model.SubItem, assigneeName, progressDesc string, thisWeekProgress map[int64][]model.ProgressRecord, lastWeekCompletion map[int64]float64, weekStart, weekEnd time.Time) dto.SubItemSnapshot {
	snapshot := dto.SubItemSnapshot{
		BizKey:              pkg.FormatID(si.BizKey),
		Code:                si.Code,
		Title:               si.Title,
		Priority:            si.Priority,
		Status:              si.ItemStatus,
		AssigneeName:        assigneeName,
		StartDate:           formatDate(si.PlanStartDate),
		ExpectedEndDate:     formatDate(si.ExpectedEndDate),
		ActualEndDate:       dates.FormatTimePtr(si.ActualEndDate),
		Completion:          si.Completion,
		ProgressDescription: progressDesc,
	}

	// Populate individual progress records for this week
	if records, ok := thisWeekProgress[si.BizKey]; ok && len(records) > 0 {
		snapshot.ProgressRecords = make([]dto.ProgressRecordDTO, 0, len(records))
		for _, pr := range records {
			snapshot.ProgressRecords = append(snapshot.ProgressRecords, dto.ProgressRecordDTO{
				ID:          pr.ID,
				Completion:  pr.Completion,
				Achievement: pr.Achievement,
				Blocker:     pr.Blocker,
				CreatedAt:   pr.CreateTime.Format(time.RFC3339),
			})
		}
	}

	if isNewlyCompleted(si, weekStart, weekEnd) {
		snapshot.JustCompleted = true
	}

	// Compute delta
	if lastComp, existed := lastWeekCompletion[si.BizKey]; existed {
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
	subItemsByMain map[int64][]model.SubItem,
	lastWeekActive, thisWeekActive map[int64]struct{},
	lastWeekProgress, thisWeekProgress map[int64][]model.ProgressRecord,
	lastWeekCompletion map[int64]float64,
	latestProgressDesc map[int64]string,
	assigneeNames map[int64]string,
	weekStart, weekEnd time.Time,
) ([]dto.WeeklyComparisonGroup, dto.WeeklyStats) {
	var groups []dto.WeeklyComparisonGroup
	var stats dto.WeeklyStats

	for _, mi := range mainItems {
		subs, ok := subItemsByMain[mi.BizKey]
		if !ok || len(subs) == 0 {
			continue
		}

		group := dto.WeeklyComparisonGroup{
			MainItem: dto.WeeklyMainItemSummary{
				BizKey:          pkg.FormatID(mi.BizKey),
				Code:            mi.Code,
				Title:           mi.Title,
				Priority:        mi.Priority,
				Status:          mi.ItemStatus,
				StartDate:       formatDate(mi.PlanStartDate),
				ExpectedEndDate: formatDate(mi.ExpectedEndDate),
				ActualEndDate:   dates.FormatTimePtr(mi.ActualEndDate),
				Completion:      mi.Completion,
				SubItemCount:    len(subs),
			},
		}

		for _, si := range subs {
			assigneeName := ""
			if si.AssigneeKey != nil {
				assigneeName = assigneeNames[*si.AssigneeKey]
			}

			inThisWeek := isInThisWeek(si.BizKey, thisWeekActive, thisWeekProgress)
			justCompleted := isNewlyCompleted(si, weekStart, weekEnd)
			if justCompleted {
				stats.NewlyCompleted++
			}

			snapshot := buildSubItemSnapshot(si, assigneeName, latestProgressDesc[si.BizKey], thisWeekProgress, lastWeekCompletion, weekStart, weekEnd)

			appendLastWeekSnapshot(&group, si, assigneeName, lastWeekActive, lastWeekProgress, lastWeekCompletion)

			if isCompletedNoChange(si, justCompleted, weekEnd) {
				group.CompletedNoChange = append(group.CompletedNoChange, snapshot)
				continue
			}

			if inThisWeek || justCompleted {
				group.ThisWeek = append(group.ThisWeek, snapshot)
				stats.ActiveSubItems++
				if si.ItemStatus == "progressing" {
					stats.InProgress++
				}
				if si.ItemStatus == "blocking" {
					stats.Blocked++
				}
				if si.ItemStatus == "pending" {
					stats.Pending++
				}
				if si.ItemStatus == "pausing" {
					stats.Pausing++
				}
				if si.ExpectedEndDate != nil && si.ExpectedEndDate.Before(weekEnd) && si.ItemStatus != "completed" && si.ItemStatus != "closed" {
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
func isInThisWeek(subBizKey int64, thisWeekActive map[int64]struct{}, thisWeekProgress map[int64][]model.ProgressRecord) bool {
	if _, ok := thisWeekActive[subBizKey]; ok {
		return true
	}
	if _, ok := thisWeekProgress[subBizKey]; ok {
		return true
	}
	return false
}

// appendLastWeekSnapshot adds a last-week snapshot to the group if the sub-item was active and had progress last week.
func appendLastWeekSnapshot(group *dto.WeeklyComparisonGroup, si model.SubItem, assigneeName string, lastWeekActive map[int64]struct{}, lastWeekProgress map[int64][]model.ProgressRecord, lastWeekCompletion map[int64]float64) {
	if _, wasActive := lastWeekActive[si.BizKey]; !wasActive {
		return
	}
	if _, hadProgress := lastWeekProgress[si.BizKey]; !hadProgress {
		return
	}
	lastSnapshot := dto.SubItemSnapshot{
		BizKey:              pkg.FormatID(si.BizKey),
		Code:            si.Code,
		Title:           si.Title,
		Priority:        si.Priority,
		Status:          si.ItemStatus,
		AssigneeName:    assigneeName,
		StartDate:       formatDate(si.PlanStartDate),
		ExpectedEndDate: formatDate(si.ExpectedEndDate),
		ActualEndDate:   dates.FormatTimePtr(si.ActualEndDate),
		Completion:      lastWeekCompletion[si.BizKey],
	}
	group.LastWeek = append(group.LastWeek, lastSnapshot)
}

// isCompletedNoChange returns true if a sub-item was completed before this week and not just completed.
func isCompletedNoChange(si model.SubItem, justCompleted bool, weekEnd time.Time) bool {
	if si.ItemStatus != "completed" || justCompleted {
		return false
	}
	completedBeforeOrDuringWeek := si.ActualEndDate == nil || !si.ActualEndDate.After(weekEnd)
	existedDuringWeek := si.CreateTime.Before(weekEnd.AddDate(0, 0, 1))
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
	if si.ItemStatus != "completed" {
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
	if si.CreateTime.After(weekEndNextDay) || si.CreateTime.Equal(weekEndNextDay) {
		return false
	}
	if si.ActualEndDate != nil && si.ActualEndDate.Before(weekStart) {
		return false
	}
	return true
}

func (s *viewService) GanttView(ctx context.Context, teamBizKey int64, filter dto.GanttFilter) (*dto.GanttResult, error) {
	// Fetch main items: use SQL pushdown for status filter, otherwise fetch all non-archived
	var mainItems []model.MainItem
	var err error
	if filter.Status != "" {
		mainItems, err = s.mainItemRepo.ListByTeamAndStatus(ctx, teamBizKey, filter.Status)
	} else {
		mainItems, err = s.mainItemRepo.ListNonArchivedByTeam(ctx, teamBizKey)
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
		return mainItems[i].CreateTime.Before(mainItems[j].CreateTime)
	})

	// Fetch all sub-items for the team (single query, avoid N+1)
	subItems, err := s.subItemRepo.ListByTeam(ctx, teamBizKey)
	if err != nil {
		return nil, err
	}

	// Index sub-items by main item BizKey
	subItemsByMain := make(map[int64][]model.SubItem)
	for _, si := range subItems {
		subItemsByMain[si.MainItemKey] = append(subItemsByMain[si.MainItemKey], si)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	items := make([]dto.GanttMainItemDTO, 0, len(mainItems))
	for _, mi := range mainItems {
		subs := subItemsByMain[mi.BizKey]
		subDTOs := make([]dto.GanttSubItemDTO, 0, len(subs))
		for _, si := range subs {
			subDTOs = append(subDTOs, dto.GanttSubItemDTO{
				BizKey:              pkg.FormatID(si.BizKey),
				Title:           si.Title,
				StartDate:       formatDate(si.PlanStartDate),
				ExpectedEndDate: formatDate(si.ExpectedEndDate),
				Completion:      si.Completion,
				Status:          si.ItemStatus,
			})
		}

		items = append(items, dto.GanttMainItemDTO{
			BizKey:          pkg.FormatID(mi.BizKey),
			Title:           mi.Title,
			Priority:        mi.Priority,
			StartDate:       formatDate(mi.PlanStartDate),
			ExpectedEndDate: formatDate(mi.ExpectedEndDate),
			Completion:      mi.Completion,
			Status:          mi.ItemStatus,
			IsOverdue:       computeIsOverdue(mi.ExpectedEndDate, mi.ItemStatus, today),
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

func (s *viewService) TableView(ctx context.Context, teamBizKey int64, filter dto.TableFilter, page dto.Pagination) (*dto.PageResult[dto.TableRow], error) {
	var rows []dto.TableRow

	// Fetch main items (non-archived only)
	if filter.Type == "" || filter.Type == "main" {
		mainItems, err := s.mainItemRepo.ListNonArchivedByTeam(ctx, teamBizKey)
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
		subItems, err := s.subItemRepo.ListByTeam(ctx, teamBizKey)
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

func (s *viewService) TableExportCSV(ctx context.Context, teamBizKey int64, filter dto.TableFilter) ([]byte, error) {
	// Fetch all matching rows (no pagination)
	var rows []dto.TableRow

	if filter.Type == "" || filter.Type == "main" {
		mainItems, err := s.mainItemRepo.ListNonArchivedByTeam(ctx, teamBizKey)
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
		subItems, err := s.subItemRepo.ListByTeam(ctx, teamBizKey)
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
		BizKey:          pkg.FormatID(mi.BizKey),
		Type:            "main",
		Code:            mi.Code,
		Title:           mi.Title,
		Priority:        mi.Priority,
		AssigneeID:      pkg.FormatIDPtr(mi.AssigneeKey),
		Status:          mi.ItemStatus,
		Completion:      mi.Completion,
		ExpectedEndDate: dates.FormatTimePtr(mi.ExpectedEndDate),
		ActualEndDate:   dates.FormatTimePtr(mi.ActualEndDate),
	}
}

func subItemToRow(si model.SubItem) dto.TableRow {
	return dto.TableRow{
		BizKey:              pkg.FormatID(si.BizKey),
		Type:            "sub",
		Code:            si.Code,
		Title:           si.Title,
		Priority:        si.Priority,
		AssigneeID:      pkg.FormatIDPtr(si.AssigneeKey),
		Status:          si.ItemStatus,
		Completion:      si.Completion,
		ExpectedEndDate: dates.FormatTimePtr(si.ExpectedEndDate),
		ActualEndDate:   dates.FormatTimePtr(si.ActualEndDate),
	}
}

func matchesFilterMain(mi model.MainItem, filter dto.TableFilter) bool {
	if len(filter.Priority) > 0 && !contains(filter.Priority, mi.Priority) {
		return false
	}
	if len(filter.Status) > 0 && !contains(filter.Status, mi.ItemStatus) {
		return false
	}
	if filter.AssigneeKey != nil {
		if mi.AssigneeKey == nil || pkg.FormatID(*mi.AssigneeKey) != *filter.AssigneeKey {
			return false
		}
	}
	return true
}

func matchesFilterSub(si model.SubItem, filter dto.TableFilter) bool {
	if len(filter.Priority) > 0 && !contains(filter.Priority, si.Priority) {
		return false
	}
	if len(filter.Status) > 0 && !contains(filter.Status, si.ItemStatus) {
		return false
	}
	if filter.AssigneeKey != nil {
		if si.AssigneeKey == nil || pkg.FormatID(*si.AssigneeKey) != *filter.AssigneeKey {
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
	// Collect unique assignee BizKeys
	assigneeBizKeys := make(map[int64]struct{})
	for _, row := range rows {
		if row.AssigneeID != nil {
			if id, err := pkg.ParseID(*row.AssigneeID); err == nil {
				assigneeBizKeys[id] = struct{}{}
			}
		}
	}
	if len(assigneeBizKeys) == 0 {
		return
	}

	// Batch resolve names with a single FindByBizKeys call
	bizKeys := make([]int64, 0, len(assigneeBizKeys))
	for k := range assigneeBizKeys {
		bizKeys = append(bizKeys, k)
	}
	users, err := userRepo.FindByBizKeys(ctx, bizKeys)
	if err != nil {
		return
	}
	names := make(map[int64]string, len(users))
	for bizKey, u := range users {
		names[bizKey] = u.DisplayName
	}

	// Fill names into rows
	for i := range rows {
		if rows[i].AssigneeID != nil {
			if id, err := pkg.ParseID(*rows[i].AssigneeID); err == nil {
				rows[i].AssigneeName = names[id]
			}
		}
	}
}

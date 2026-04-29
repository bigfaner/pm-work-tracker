package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/snowflake"
)

// thisWeekMonday returns the Monday of the current week (for weekStart parameter).
func thisWeekMonday() time.Time {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1-weekday)
}

// seedViewItemData creates MainItems and SubItems with specified statuses directly in the DB.
// Returns the created sub-item IDs.
func seedViewItemData(t *testing.T, db *gorm.DB, teamBizKey int64, userID uint, weekStart time.Time) {
	t.Helper()

	now := time.Now()
	loc := now.Location()

	// Item 1: completed sub-item (completed this week)
	mi1 := &model.MainItem{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:      teamBizKey,
		Code:         "TAMA-W001",
		Title:        "Weekly Completed Item",
		Priority:     "P1",
		ProposerKey:  getUserBizKey(t, db, userID),
		ItemStatus:   "progressing",
		PlanStartDate:   ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, loc)),
		ExpectedEndDate: ptrTime(time.Date(2026, 6, 30, 0, 0, 0, 0, loc)),
	}
	require.NoError(t, db.Create(mi1).Error)
	endDate := weekStart.AddDate(0, 0, 2) // Wednesday of this week
	sub1 := &model.SubItem{
		BaseModel:      model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:        teamBizKey,
		MainItemKey:    mi1.BizKey,
		Title:          "Completed Sub",
		Priority:       "P2",
		ItemStatus:     "completed",
		Completion:     100,
		Weight:         1.0,
		PlanStartDate:  ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, loc)),
		ExpectedEndDate: ptrTime(time.Date(2026, 6, 30, 0, 0, 0, 0, loc)),
		ActualEndDate:  &endDate,
	}
	require.NoError(t, db.Create(sub1).Error)

	// Item 2: progressing sub-item
	mi2 := &model.MainItem{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:      teamBizKey,
		Code:         "TAMA-W002",
		Title:        "Weekly Progressing Item 1",
		Priority:     "P2",
		ProposerKey:  getUserBizKey(t, db, userID),
		ItemStatus:   "progressing",
		Completion:   30,
		PlanStartDate:   ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, loc)),
		ExpectedEndDate: ptrTime(time.Date(2026, 6, 30, 0, 0, 0, 0, loc)),
	}
	require.NoError(t, db.Create(mi2).Error)
	sub2 := &model.SubItem{
		BaseModel:      model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:        teamBizKey,
		MainItemKey:    mi2.BizKey,
		Title:          "Progressing Sub 1",
		Priority:       "P2",
		ItemStatus:     "progressing",
		Completion:     30,
		Weight:         1.0,
		PlanStartDate:  ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, loc)),
		ExpectedEndDate: ptrTime(time.Date(2026, 6, 30, 0, 0, 0, 0, loc)),
	}
	require.NoError(t, db.Create(sub2).Error)

	// Item 3: another progressing sub-item
	mi3 := &model.MainItem{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:      teamBizKey,
		Code:         "TAMA-W003",
		Title:        "Weekly Progressing Item 2",
		Priority:     "P3",
		ProposerKey:  getUserBizKey(t, db, userID),
		ItemStatus:   "progressing",
		Completion:   50,
		PlanStartDate:   ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, loc)),
		ExpectedEndDate: ptrTime(time.Date(2026, 6, 30, 0, 0, 0, 0, loc)),
	}
	require.NoError(t, db.Create(mi3).Error)
	sub3 := &model.SubItem{
		BaseModel:      model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:        teamBizKey,
		MainItemKey:    mi3.BizKey,
		Title:          "Progressing Sub 2",
		Priority:       "P3",
		ItemStatus:     "progressing",
		Completion:     50,
		Weight:         1.0,
		PlanStartDate:  ptrTime(time.Date(2026, 1, 1, 0, 0, 0, 0, loc)),
		ExpectedEndDate: ptrTime(time.Date(2026, 6, 30, 0, 0, 0, 0, loc)),
	}
	require.NoError(t, db.Create(sub3).Error)

	// Add progress records for this week to make sub-items appear as "active"
	pr1 := &model.ProgressRecord{
		BizKey:      snowflake.Generate(),
		SubItemKey:  sub1.BizKey,
		TeamKey:     teamBizKey,
		AuthorKey:   getUserBizKey(t, db, userID),
		Completion:  100,
		Achievement: "Finished the task",
		CreateTime:  weekStart.Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(pr1).Error)

	pr2 := &model.ProgressRecord{
		BizKey:      snowflake.Generate(),
		SubItemKey:  sub2.BizKey,
		TeamKey:     teamBizKey,
		AuthorKey:   getUserBizKey(t, db, userID),
		Completion:  30,
		Achievement: "Working on it",
		CreateTime:  weekStart.Add(48 * time.Hour),
	}
	require.NoError(t, db.Create(pr2).Error)

	pr3 := &model.ProgressRecord{
		BizKey:      snowflake.Generate(),
		SubItemKey:  sub3.BizKey,
		TeamKey:     teamBizKey,
		AuthorKey:   getUserBizKey(t, db, userID),
		Completion:  50,
		Achievement: "Halfway done",
		CreateTime:  weekStart.Add(48 * time.Hour),
	}
	require.NoError(t, db.Create(pr3).Error)
}

// ptrTime returns a pointer to the given time.
func ptrTime(t time.Time) *time.Time {
	return &t
}

// ========== Weekly View Tests ==========

func TestViews_Weekly_ThreeItems_ReturnsCorrectStats(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	weekStart := thisWeekMonday()

	seedViewItemData(t, db, data.teamABizKey, data.userAID, weekStart)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/weekly?weekStart=%s", data.teamABizKey, weekStart.Format("2006-01-02")),
		"", pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	respData := resp["data"].(map[string]interface{})
	stats := respData["stats"].(map[string]interface{})

	// 2 progressing sub-items (Progressing Sub 1, Progressing Sub 2)
	assert.Equal(t, float64(2), stats["inProgress"])
	// No overdue items (expectedEndDate is 2026-06-30, well in the future)
	assert.Equal(t, float64(0), stats["overdue"])
}

func TestViews_Weekly_EmptyTeam_ReturnsZeroStats(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	weekStart := thisWeekMonday()

	// Seed an empty team C
	teamC := &model.Team{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamName:  "Team C Empty",
		PmKey:     getUserBizKey(t, db, data.userAID),
		Code:      "TCEM",
	}
	require.NoError(t, db.Create(teamC).Error)

	pmRoleBizKey := findRoleBizKeyByName(t, db, "pm")
	pmRoleBizKeyInt, _ := strconv.ParseInt(pmRoleBizKey, 10, 64)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey:  teamC.BizKey,
		UserKey:  getUserBizKey(t, db, data.userAID),
		RoleKey:  &pmRoleBizKeyInt,
		JoinedAt: time.Now(),
	}).Error)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/weekly?weekStart=%s", teamC.BizKey, weekStart.Format("2006-01-02")),
		"", pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	respData := resp["data"].(map[string]interface{})
	stats := respData["stats"].(map[string]interface{})

	assert.Equal(t, float64(0), stats["activeSubItems"])
	assert.Equal(t, float64(0), stats["inProgress"])
	assert.Equal(t, float64(0), stats["overdue"])
	assert.Equal(t, float64(0), stats["newlyCompleted"])

	// Groups may be nil or empty
	groupsRaw := respData["groups"]
	if groupsRaw == nil {
		// nil is acceptable for empty groups
	} else {
		groups := groupsRaw.([]interface{})
		assert.Empty(t, groups)
	}
}

// ========== Gantt View Tests ==========

func TestViews_Gantt_ItemsHaveDatesAndSubItems(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	weekStart := thisWeekMonday()

	seedViewItemData(t, db, data.teamABizKey, data.userAID, weekStart)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/gantt", data.teamABizKey), "", pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	respData := resp["data"].(map[string]interface{})
	items := respData["items"].([]interface{})
	require.Len(t, items, 3)

	// Check first item has dates and nested sub-items
	item := items[0].(map[string]interface{})
	assert.NotEmpty(t, item["startDate"])
	assert.NotEmpty(t, item["expectedEndDate"])

	subs := item["subItems"].([]interface{})
	require.Len(t, subs, 1)

	sub := subs[0].(map[string]interface{})
	assert.NotEmpty(t, sub["title"])
	assert.NotEmpty(t, sub["startDate"])
	assert.NotEmpty(t, sub["expectedEndDate"])
}

func TestViews_Gantt_EmptyTeam_ReturnsEmptyArray(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	teamC := &model.Team{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamName:  "Team Gantt Empty",
		PmKey:     getUserBizKey(t, db, data.userAID),
		Code:      "TGEM",
	}
	require.NoError(t, db.Create(teamC).Error)
	pmRoleBizKey := findRoleBizKeyByName(t, db, "pm")
	pmRoleBizKeyInt, _ := strconv.ParseInt(pmRoleBizKey, 10, 64)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey:  teamC.BizKey,
		UserKey:  getUserBizKey(t, db, data.userAID),
		RoleKey:  &pmRoleBizKeyInt,
		JoinedAt: time.Now(),
	}).Error)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/gantt", teamC.BizKey), "", pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	respData := resp["data"].(map[string]interface{})
	items := respData["items"].([]interface{})
	assert.Empty(t, items)
}

// ========== Table View Tests ==========

func TestViews_Table_StatusFilter_ReturnsFilteredItems(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	weekStart := thisWeekMonday()

	seedViewItemData(t, db, data.teamABizKey, data.userAID, weekStart)

	// Filter for completed status
	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/table?status=completed", data.teamABizKey), "", pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	respData := resp["data"].(map[string]interface{})
	items := respData["items"].([]interface{})
	total := respData["total"].(float64)

	// Only completed items should be returned (at least the completed sub-item)
	assert.GreaterOrEqual(t, int(total), 1)
	for _, raw := range items {
		item := raw.(map[string]interface{})
		assert.Equal(t, "completed", item["status"])
	}
}

func TestViews_Table_Pagination_ReturnsCorrectTotal(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	// Create 3 main items via API
	for i := 0; i < 3; i++ {
		createTestMainItem(t, r, pmToken, data.teamABizKey, fmt.Sprintf("Page Item %d", i))
	}

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/table?page=1&pageSize=2&type=main", data.teamABizKey), "", pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	respData := resp["data"].(map[string]interface{})
	items := respData["items"].([]interface{})
	total := respData["total"].(float64)

	assert.Len(t, items, 2)      // page size 2
	assert.Equal(t, float64(3), total) // total 3 items
	assert.Equal(t, float64(1), respData["page"])
	assert.Equal(t, float64(2), respData["size"])
}

func TestViews_Table_EmptyTeam_ReturnsEmptyItems(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	teamC := &model.Team{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamName:  "Team Table Empty",
		PmKey:     getUserBizKey(t, db, data.userAID),
		Code:      "TTEM",
	}
	require.NoError(t, db.Create(teamC).Error)
	pmRoleBizKey := findRoleBizKeyByName(t, db, "pm")
	pmRoleBizKeyInt, _ := strconv.ParseInt(pmRoleBizKey, 10, 64)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey:  teamC.BizKey,
		UserKey:  getUserBizKey(t, db, data.userAID),
		RoleKey:  &pmRoleBizKeyInt,
		JoinedAt: time.Now(),
	}).Error)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/table", teamC.BizKey), "", pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	respData := resp["data"].(map[string]interface{})
	items := respData["items"].([]interface{})
	total := respData["total"].(float64)

	assert.Empty(t, items)
	assert.Equal(t, float64(0), total)
}

// ========== CSV Export Tests ==========

func TestViews_TableExport_HasBOMAndHeader(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	// Create at least one main item (table export includes main items)
	createTestMainItem(t, r, pmToken, data.teamABizKey, "Export Item")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/table/export", data.teamABizKey), "", pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	body := w.Body.Bytes()
	// Verify BOM prefix (0xEF 0xBB 0xBF)
	assert.True(t, len(body) >= 3 && body[0] == 0xEF && body[1] == 0xBB && body[2] == 0xBF,
		"response should start with UTF-8 BOM")

	// Strip BOM and check header row
	content := string(body[3:])
	lines := strings.Split(content, "\n")
	require.True(t, len(lines) >= 1, "CSV should have at least a header row")
	assert.Contains(t, lines[0], "编号")
	assert.Contains(t, lines[0], "标题")
}

func TestViews_TableExport_EmptyTeam_Returns422(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	teamC := &model.Team{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamName:  "Team Export Empty",
		PmKey:     getUserBizKey(t, db, data.userAID),
		Code:      "TEXM",
	}
	require.NoError(t, db.Create(teamC).Error)
	pmRoleBizKey := findRoleBizKeyByName(t, db, "pm")
	pmRoleBizKeyInt, _ := strconv.ParseInt(pmRoleBizKey, 10, 64)
	require.NoError(t, db.Create(&model.TeamMember{
		TeamKey:  teamC.BizKey,
		UserKey:  getUserBizKey(t, db, data.userAID),
		RoleKey:  &pmRoleBizKeyInt,
		JoinedAt: time.Now(),
	}).Error)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/table/export", teamC.BizKey), "", pmToken)
	// ErrNoData returns 422
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ========== Report Preview Tests ==========

func TestReports_WeeklyPreview_ContainsSections(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	weekStart := thisWeekMonday()

	// Seed report data for this week using the shared helper
	seedReportData(t, db, data.teamABizKey, data.userAID, weekStart)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/reports/weekly/preview?weekStart=%s", data.teamABizKey, weekStart.Format("2006-01-02")),
		"", pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, weekStart.Format("2006-01-02"), respData["weekStart"])

	sections := respData["sections"].([]interface{})
	require.GreaterOrEqual(t, len(sections), 1)

	section := sections[0].(map[string]interface{})
	mainItem := section["mainItem"].(map[string]interface{})
	assert.NotEmpty(t, mainItem["title"])

	subs := section["subItems"].([]interface{})
	require.GreaterOrEqual(t, len(subs), 1)

	sub := subs[0].(map[string]interface{})
	assert.NotEmpty(t, sub["title"])
	achievements := sub["achievements"].([]interface{})
	assert.GreaterOrEqual(t, len(achievements), 1)
}

func TestReports_WeeklyPreview_NoActivity_Returns422(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	// Use a past week where there's no activity
	weekStart := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC) // Monday 2025-01-06

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/reports/weekly/preview?weekStart=%s", data.teamABizKey, weekStart.Format("2006-01-02")),
		"", pmToken)
	// ErrNoData returns 422
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ========== Report Export Tests ==========

func TestReports_WeeklyExport_MarkdownContent(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")
	weekStart := thisWeekMonday()

	seedReportData(t, db, data.teamABizKey, data.userAID, weekStart)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/reports/weekly/export?weekStart=%s", data.teamABizKey, weekStart.Format("2006-01-02")),
		"", pmToken)
	require.Equal(t, http.StatusOK, w.Code)

	// Verify Content-Type is text/markdown
	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "text/markdown", contentType)

	body := w.Body.String()
	// Markdown should contain the report header
	assert.Contains(t, body, "周报")
	// Should contain the main item and sub-item titles from seedReportData
	assert.Contains(t, body, "Report Test Main Item")
	assert.Contains(t, body, "Report Test Sub Item")
	// Should contain completion and achievements
	assert.Contains(t, body, "完成度")
	assert.Contains(t, body, "成果")
	assert.Contains(t, body, "Completed half the work")
}

func TestReports_WeeklyExport_NoData_Returns422(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	weekStart := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/reports/weekly/export?weekStart=%s", data.teamABizKey, weekStart.Format("2006-01-02")),
		"", pmToken)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestReports_WeeklyExport_FutureWeek_Returns422(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	pmToken := loginAs(t, r, "userA", "passwordA")

	// Use a future Monday
	futureMonday := time.Date(2030, 1, 7, 0, 0, 0, 0, time.UTC)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/reports/weekly/export?weekStart=%s", data.teamABizKey, futureMonday.Format("2006-01-02")),
		"", pmToken)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ========== Permission Tests ==========

func TestViews_Weekly_NonMember_Returns403(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	userBToken := loginAs(t, r, "userB", "passwordB")
	weekStart := thisWeekMonday()

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/weekly?weekStart=%s", data.teamABizKey, weekStart.Format("2006-01-02")),
		"", userBToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestViews_Gantt_MemberWithoutPermission_Returns403(t *testing.T) {
	r, data, _ := setupLifecycleTest(t)
	// memberA has view:weekly and view:table, but NOT view:gantt
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/views/gantt", data.teamABizKey), "", memberToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestReports_WeeklyPreview_MemberWithPermission_Returns200(t *testing.T) {
	r, data, db := setupLifecycleTest(t)
	// memberA has report:export permission
	memberToken := loginAs(t, r, "memberA", "passwordMemberA")
	weekStart := thisWeekMonday()

	seedReportData(t, db, data.teamABizKey, data.userAID, weekStart)

	w := makeRequest(t, r, http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/reports/weekly/preview?weekStart=%s", data.teamABizKey, weekStart.Format("2006-01-02")),
		"", memberToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ========== Helper: get sub-item bizKey by title ==========

// getSubBizKey finds a sub-item's bizKey string by title directly from the database.
func getSubBizKey(t *testing.T, db *gorm.DB, title string) string {
	t.Helper()
	var sub model.SubItem
	require.NoError(t, db.Where("title = ?", title).First(&sub).Error)
	return fmt.Sprintf("%d", sub.BizKey)
}

// mustParseInt64 parses a string to int64, panicking on failure.
func mustParseInt64(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("mustParseInt64(%q): %v", s, err))
	}
	return v
}

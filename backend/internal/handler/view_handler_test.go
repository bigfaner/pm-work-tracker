package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Mock ViewService for handler tests
// ---------------------------------------------------------------------------

type mockViewService struct {
	comparisonResult struct {
		result *dto.WeeklyViewResponse
		err    error
	}
	ganttResult struct {
		result *dto.GanttResult
		err    error
	}
	tableResult struct {
		result *dto.PageResult[dto.TableRow]
		err    error
	}
	csvResult struct {
		bytes []byte
		err   error
	}

	// capture calls
	lastTeamID    uint
	lastWeekStart time.Time

	comparisonCalled bool

	ganttCalled bool
	lastFilter  dto.GanttFilter

	tableCalled bool
	lastTable   dto.TableFilter
	lastPage    dto.Pagination

	csvCalled bool
	lastCSV   dto.TableFilter
}

func (m *mockViewService) WeeklyComparison(_ context.Context, teamID uint, weekStart time.Time) (*dto.WeeklyViewResponse, error) {
	m.comparisonCalled = true
	m.lastTeamID = teamID
	m.lastWeekStart = weekStart
	return m.comparisonResult.result, m.comparisonResult.err
}

func (m *mockViewService) GanttView(_ context.Context, teamID uint, filter dto.GanttFilter) (*dto.GanttResult, error) {
	m.ganttCalled = true
	m.lastTeamID = teamID
	m.lastFilter = filter
	return m.ganttResult.result, m.ganttResult.err
}

func (m *mockViewService) TableView(_ context.Context, teamID uint, filter dto.TableFilter, page dto.Pagination) (*dto.PageResult[dto.TableRow], error) {
	m.tableCalled = true
	m.lastTeamID = teamID
	m.lastTable = filter
	m.lastPage = page
	return m.tableResult.result, m.tableResult.err
}

func (m *mockViewService) TableExportCSV(_ context.Context, teamID uint, filter dto.TableFilter) ([]byte, error) {
	m.csvCalled = true
	m.lastTeamID = teamID
	m.lastCSV = filter
	return m.csvResult.bytes, m.csvResult.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------


func depsWithViewSvc(t *testing.T, svc *mockViewService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{ RoleKey: func() *int64 { v := int64(1); return &v }()}}
	deps.View = NewViewHandler(svc)
	return deps
}

// monday returns a Monday date string for testing.
func monday() string {
	// 2026-04-13 is a Monday
	return "2026-04-13"
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/views/weekly
// ---------------------------------------------------------------------------

func TestWeeklyView_Success(t *testing.T) {
	svc := &mockViewService{}
	svc.comparisonResult.result = &dto.WeeklyViewResponse{
		WeekStart: "2026-04-13",
		WeekEnd:   "2026-04-19",
		Stats:     dto.WeeklyStats{},
		Groups:    []dto.WeeklyComparisonGroup{},
	}

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/weekly?weekStart="+monday(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "2026-04-13", data["weekStart"])
	assert.Equal(t, "2026-04-19", data["weekEnd"])

	assert.True(t, svc.comparisonCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
}

func TestWeeklyView_MissingWeekStart(t *testing.T) {
	svc := &mockViewService{}
	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/weekly", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.comparisonCalled)
}

func TestWeeklyView_InvalidDateFormat(t *testing.T) {
	svc := &mockViewService{}
	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/weekly?weekStart=not-a-date", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.comparisonCalled)
}

func TestWeeklyView_NotAMonday(t *testing.T) {
	svc := &mockViewService{}
	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	// 2026-04-14 is a Tuesday
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/weekly?weekStart=2026-04-14", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", resp["code"])
	assert.False(t, svc.comparisonCalled)
}

func TestWeeklyView_ServiceError(t *testing.T) {
	svc := &mockViewService{}
	svc.comparisonResult.err = errors.New("db error")

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/weekly?weekStart="+monday(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWeeklyView_FutureWeekNotAllowed(t *testing.T) {
	svc := &mockViewService{}
	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	// Use a future Monday: 2099-01-05
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/weekly?weekStart=2099-01-05", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "FUTURE_WEEK_NOT_ALLOWED", resp["code"])
	assert.False(t, svc.comparisonCalled)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/views/gantt
// ---------------------------------------------------------------------------

func TestGanttView_Success(t *testing.T) {
	svc := &mockViewService{}
	svc.ganttResult.result = &dto.GanttResult{
		Items: []dto.GanttMainItemDTO{},
	}

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/gantt", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
	assert.True(t, svc.ganttCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
}

func TestGanttView_WithStatusFilter(t *testing.T) {
	svc := &mockViewService{}
	svc.ganttResult.result = &dto.GanttResult{}

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/gantt?status=progressing", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.ganttCalled)
	assert.Equal(t, "progressing", svc.lastFilter.Status)
}

func TestGanttView_ServiceError(t *testing.T) {
	svc := &mockViewService{}
	svc.ganttResult.err = errors.New("db error")

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/gantt", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/views/table
// ---------------------------------------------------------------------------

func TestTableView_Success(t *testing.T) {
	svc := &mockViewService{}
	svc.tableResult.result = &dto.PageResult[dto.TableRow]{
		Items: []dto.TableRow{},
		Total: 0,
		Page:  1,
		Size:  50,
	}

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
	assert.True(t, svc.tableCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
	// Default page size is 50
	assert.Equal(t, 50, svc.lastPage.PageSize)
	assert.Equal(t, 1, svc.lastPage.Page)
}

func TestTableView_WithFilters(t *testing.T) {
	svc := &mockViewService{}
	svc.tableResult.result = &dto.PageResult[dto.TableRow]{}

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table?type=main&priority=P1&priority=P2&status=progressing&page=2&pageSize=10&sortBy=completion&sortOrder=desc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.tableCalled)
	assert.Equal(t, "main", svc.lastTable.Type)
	assert.Equal(t, 10, svc.lastPage.PageSize)
	assert.Equal(t, 2, svc.lastPage.Page)
}

func TestTableView_DefaultPageSize50(t *testing.T) {
	svc := &mockViewService{}
	svc.tableResult.result = &dto.PageResult[dto.TableRow]{}

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 50, svc.lastPage.PageSize)
}

func TestTableView_InvalidPageClamped(t *testing.T) {
	svc := &mockViewService{}
	svc.tableResult.result = &dto.PageResult[dto.TableRow]{}

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	// page=0 should be clamped to 1
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table?page=0", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, svc.lastPage.Page)
}

func TestTableView_ServiceError(t *testing.T) {
	svc := &mockViewService{}
	svc.tableResult.err = errors.New("db error")

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/views/table/export
// ---------------------------------------------------------------------------

func TestExportTable_Success(t *testing.T) {
	svc := &mockViewService{}
	svc.csvResult.bytes = []byte("code,title,type\r\nTEST-00001,Test,main\r\n")

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table/export", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, `attachment; filename="items-export.csv"`, w.Header().Get("Content-Disposition"))
	assert.True(t, svc.csvCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
}

func TestExportTable_WithFilters(t *testing.T) {
	svc := &mockViewService{}
	svc.csvResult.bytes = []byte("code,title,type\r\n")

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table/export?type=sub&status=已完成", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.csvCalled)
	assert.Equal(t, "sub", svc.lastCSV.Type)
}

func TestExportTable_NoData(t *testing.T) {
	svc := &mockViewService{}
	svc.csvResult.err = apperrors.ErrNoData

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table/export", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "NO_DATA", resp["code"])
}

func TestExportTable_ServiceError(t *testing.T) {
	svc := &mockViewService{}
	svc.csvResult.err = errors.New("db error")

	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table/export", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Auth + Team Membership enforcement
// ---------------------------------------------------------------------------

func TestWeeklyView_RequiresAuth(t *testing.T) {
	svc := &mockViewService{}
	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/weekly?weekStart="+monday(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, svc.comparisonCalled)
}

func TestGanttView_RequiresAuth(t *testing.T) {
	svc := &mockViewService{}
	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/gantt", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, svc.ganttCalled)
}

func TestTableView_RequiresAuth(t *testing.T) {
	svc := &mockViewService{}
	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, svc.tableCalled)
}

func TestExportTable_RequiresAuth(t *testing.T) {
	svc := &mockViewService{}
	deps := depsWithViewSvc(t, svc)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/table/export", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, svc.csvCalled)
}

func TestWeeklyView_RequiresTeamMembership(t *testing.T) {
	svc := &mockViewService{}
	deps, db := testDeps(t)
	// Ensure user ID=99 exists so AuthMiddleware doesn't return 401
	db.Create(&model.User{BaseModel: model.BaseModel{ID: 99}, Username: "testuser99", DisplayName: "Test User 99"})
	// Use a mock team repo that returns no membership (error)
	deps.View = NewViewHandler(svc)
	// default TeamRepo from testDeps has no members, so FindMember returns error
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 99, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/views/weekly?weekStart="+monday(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.comparisonCalled)
}

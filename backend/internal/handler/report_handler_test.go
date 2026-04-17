package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
// Mock ReportService for handler tests
// ---------------------------------------------------------------------------

type mockReportService struct {
	previewResult struct {
		result *dto.ReportPreview
		err    error
	}
	exportResult struct {
		bytes []byte
		err   error
	}

	// capture calls
	previewCalled bool
	lastTeamID    uint
	lastWeekStart time.Time

	exportCalled bool
}

func (m *mockReportService) Preview(_ context.Context, teamID uint, weekStart time.Time) (*dto.ReportPreview, error) {
	m.previewCalled = true
	m.lastTeamID = teamID
	m.lastWeekStart = weekStart
	return m.previewResult.result, m.previewResult.err
}

func (m *mockReportService) ExportMarkdown(_ context.Context, teamID uint, weekStart time.Time) ([]byte, error) {
	m.exportCalled = true
	m.lastTeamID = teamID
	m.lastWeekStart = weekStart
	return m.exportResult.bytes, m.exportResult.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func depsWithReportSvc(t *testing.T, svc *mockReportService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm"}}
	deps.Report = NewReportHandlerWithDeps(svc)
	return deps
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/reports/weekly/preview
// ---------------------------------------------------------------------------

func TestWeeklyPreview_Success(t *testing.T) {
	svc := &mockReportService{}
	svc.previewResult.result = &dto.ReportPreview{
		WeekStart: "2026-04-13",
		WeekEnd:   "2026-04-19",
		Sections: []dto.ReportSectionDTO{
			{
				MainItem: dto.MainItemSummaryDTO{ID: 1, Title: "Backend", Completion: 60},
				SubItems: []dto.ReportSubItemDTO{
					{ID: 10, Title: "Auth", Completion: 80, Achievements: []string{"done"}, Blockers: []string{}},
				},
			},
		},
	}

	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/preview?weekStart="+monday(), nil)
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

	assert.True(t, svc.previewCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
}

func TestWeeklyPreview_MissingWeekStart(t *testing.T) {
	svc := &mockReportService{}
	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/preview", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.previewCalled)
}

func TestWeeklyPreview_InvalidDateFormat(t *testing.T) {
	svc := &mockReportService{}
	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/preview?weekStart=not-a-date", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.previewCalled)
}

func TestWeeklyPreview_NotAMonday(t *testing.T) {
	svc := &mockReportService{}
	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	// 2026-04-14 is a Tuesday
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/preview?weekStart=2026-04-14", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", resp["code"])
	assert.False(t, svc.previewCalled)
}

func TestWeeklyPreview_NoData(t *testing.T) {
	svc := &mockReportService{}
	svc.previewResult.err = apperrors.ErrNoData

	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/preview?weekStart="+monday(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "NO_DATA", resp["code"])
}

func TestWeeklyPreview_ServiceError(t *testing.T) {
	svc := &mockReportService{}
	svc.previewResult.err = errors.New("db error")

	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/preview?weekStart="+monday(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWeeklyPreview_RequiresAuth(t *testing.T) {
	svc := &mockReportService{}
	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/preview?weekStart="+monday(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, svc.previewCalled)
}

func TestWeeklyPreview_RequiresTeamMembership(t *testing.T) {
	svc := &mockReportService{}
	deps, _ := testDeps(t)
	deps.Report = NewReportHandlerWithDeps(svc)
	// default TeamRepo from testDeps has no members
	r := SetupRouter(deps)

	token := signTestToken(t, 99, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/preview?weekStart="+monday(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.previewCalled)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/reports/weekly/export
// ---------------------------------------------------------------------------

func TestWeeklyExport_Success(t *testing.T) {
	svc := &mockReportService{}
	svc.exportResult.bytes = []byte("# Weekly Report\n## Backend\n- Auth: 80%\n")

	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/export?weekStart="+monday(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/markdown", w.Header().Get("Content-Type"))

	weekStart, _ := time.Parse("2006-01-02", monday())
	isoYear, isoWeek := weekStart.ISOWeek()
	expectedFilename := fmt.Sprintf(`attachment; filename="weekly-report-%d-W%02d.md"`, isoYear, isoWeek)
	assert.Equal(t, expectedFilename, w.Header().Get("Content-Disposition"))

	assert.True(t, svc.exportCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
	assert.Contains(t, w.Body.String(), "# Weekly Report")
}

func TestWeeklyExport_MissingWeekStart(t *testing.T) {
	svc := &mockReportService{}
	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/export", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.exportCalled)
}

func TestWeeklyExport_InvalidDateFormat(t *testing.T) {
	svc := &mockReportService{}
	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/export?weekStart=invalid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.exportCalled)
}

func TestWeeklyExport_NotAMonday(t *testing.T) {
	svc := &mockReportService{}
	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/export?weekStart=2026-04-14", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.exportCalled)
}

func TestWeeklyExport_NoData(t *testing.T) {
	svc := &mockReportService{}
	svc.exportResult.err = apperrors.ErrNoData

	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/export?weekStart="+monday(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "NO_DATA", resp["code"])
}

func TestWeeklyExport_ServiceError(t *testing.T) {
	svc := &mockReportService{}
	svc.exportResult.err = errors.New("db error")

	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/export?weekStart="+monday(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWeeklyExport_RequiresAuth(t *testing.T) {
	svc := &mockReportService{}
	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/export?weekStart="+monday(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, svc.exportCalled)
}

func TestWeeklyExport_FilenameFormat(t *testing.T) {
	svc := &mockReportService{}
	svc.exportResult.bytes = []byte("report")

	deps := depsWithReportSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/reports/weekly/export?weekStart=2026-04-13", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `attachment; filename="weekly-report-2026-W16.md"`, w.Header().Get("Content-Disposition"))
}

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/service"
)

// buildPermTestRouter creates a minimal Gin router for permission boundary testing.
// permCodes is injected into context before RequirePermission runs.
// handlerFn is a stub that returns 200 OK when reached.
func buildPermTestRouter(code string, permCodes []string, handlerFn gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/test",
		func(c *gin.Context) {
			c.Set("teamID", uint(1))
			c.Set("permCodes", permCodes)
			c.Next()
		},
		middleware.RequirePermission(code, nil),
		handlerFn,
	)
	return r
}

// ---------------------------------------------------------------------------
// Mock service types for permission matrix tests (zero-value stubs)
// ---------------------------------------------------------------------------

type mockMainItemSvc struct{}

func (m *mockMainItemSvc) Create(_ context.Context, _ int64, _ uint, _ dto.MainItemCreateReq) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemSvc) Update(_ context.Context, _ int64, _ uint, _ dto.MainItemUpdateReq) error {
	return nil
}
func (m *mockMainItemSvc) Archive(_ context.Context, _ int64, _ uint) error { return nil }
func (m *mockMainItemSvc) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemSvc) Get(_ context.Context, _ uint) (*model.MainItem, error) { return nil, nil }
func (m *mockMainItemSvc) GetByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemSvc) RecalcCompletion(_ context.Context, _ uint) error { return nil }
func (m *mockMainItemSvc) ChangeStatus(_ context.Context, _ int64, _, _ uint, _ string) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemSvc) AvailableTransitions(_ context.Context, _ int64, _, _ uint) ([]string, error) {
	return nil, nil
}
func (m *mockMainItemSvc) EvaluateLinkage(_ context.Context, _, _ uint) (*service.LinkageResult, error) {
	return nil, nil
}

var _ service.MainItemService = (*mockMainItemSvc)(nil)

type mockTeamSvc struct{}

func (m *mockTeamSvc) CreateTeam(_ context.Context, _ int64, _ dto.CreateTeamReq) (*model.Team, error) {
	return nil, nil
}
func (m *mockTeamSvc) GetTeam(_ context.Context, _ int64) (*model.Team, error) { return nil, nil }
func (m *mockTeamSvc) GetTeamDetail(_ context.Context, _ int64) (*dto.TeamDetailResp, error) {
	return nil, nil
}
func (m *mockTeamSvc) ListTeams(_ context.Context, _ uint, _ bool, _ string, _, _ int) ([]*dto.TeamListResp, int64, error) {
	return nil, 0, nil
}
func (m *mockTeamSvc) UpdateTeam(_ context.Context, _ int64, _ int64, _ dto.UpdateTeamReq) (*model.Team, error) {
	return nil, nil
}
func (m *mockTeamSvc) InviteMember(_ context.Context, _ int64, _ int64, _ dto.InviteMemberReq) error {
	return nil
}
func (m *mockTeamSvc) RemoveMember(_ context.Context, _ int64, _ int64, _ int64) error      { return nil }
func (m *mockTeamSvc) TransferPM(_ context.Context, _ int64, _ int64, _ int64) error        { return nil }
func (m *mockTeamSvc) DisbandTeam(_ context.Context, _ int64, _ int64, _ string) error     { return nil }
func (m *mockTeamSvc) UpdateMemberRole(_ context.Context, _, _ int64, _, _ int64) error    { return nil }
func (m *mockTeamSvc) ListMembers(_ context.Context, _ int64) ([]*dto.TeamMemberDTO, error) {
	return nil, nil
}
func (m *mockTeamSvc) SearchAvailableUsers(_ context.Context, _ int64, _ string) ([]*dto.UserSearchDTO, error) {
	return nil, nil
}

var _ service.TeamService = (*mockTeamSvc)(nil)

type mockProgressSvc struct{}

func (m *mockProgressSvc) Append(_ context.Context, _ int64, _, _ uint, _ float64, _, _, _ string, _ bool) (*model.ProgressRecord, error) {
	return nil, nil
}
func (m *mockProgressSvc) CorrectCompletion(_ context.Context, _ int64, _ uint, _ float64) error {
	return nil
}
func (m *mockProgressSvc) List(_ context.Context, _ int64, _ uint) ([]model.ProgressRecord, error) {
	return nil, nil
}
func (m *mockProgressSvc) GetByBizKey(_ context.Context, _ int64) (*model.ProgressRecord, error) {
	return nil, nil
}

var _ service.ProgressService = (*mockProgressSvc)(nil)

type mockItemPoolSvc struct{}

func (m *mockItemPoolSvc) Submit(_ context.Context, _ int64, _ uint, _ dto.SubmitItemPoolReq) (*model.ItemPool, error) {
	return nil, nil
}
func (m *mockItemPoolSvc) Assign(_ context.Context, _ int64, _, _ uint, _ dto.AssignItemPoolReq) error {
	return nil
}
func (m *mockItemPoolSvc) ConvertToMain(_ context.Context, _ int64, _, _ uint, _ dto.ConvertToMainItemReq) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockItemPoolSvc) Reject(_ context.Context, _ int64, _, _ uint, _ string) error { return nil }
func (m *mockItemPoolSvc) List(_ context.Context, _ int64, _ dto.ItemPoolFilter, _ dto.Pagination) (*dto.PageResult[model.ItemPool], error) {
	return nil, nil
}
func (m *mockItemPoolSvc) Get(_ context.Context, _ int64, _ uint) (*model.ItemPool, error) {
	return nil, nil
}
func (m *mockItemPoolSvc) Update(_ context.Context, _ int64, _ uint, _ dto.UpdateItemPoolReq) (*model.ItemPool, error) {
	return nil, nil
}
func (m *mockItemPoolSvc) GetByBizKey(_ context.Context, _ int64) (*model.ItemPool, error) {
	return nil, nil
}

var _ service.ItemPoolService = (*mockItemPoolSvc)(nil)

type mockViewSvc struct{}

func (m *mockViewSvc) WeeklyComparison(_ context.Context, _ int64, _ time.Time) (*dto.WeeklyViewResponse, error) {
	return nil, nil
}
func (m *mockViewSvc) GanttView(_ context.Context, _ int64, _ dto.GanttFilter) (*dto.GanttResult, error) {
	return nil, nil
}
func (m *mockViewSvc) TableView(_ context.Context, _ int64, _ dto.TableFilter, _ dto.Pagination) (*dto.PageResult[dto.TableRow], error) {
	return nil, nil
}
func (m *mockViewSvc) TableExportCSV(_ context.Context, _ int64, _ dto.TableFilter) ([]byte, error) {
	return nil, nil
}

var _ service.ViewService = (*mockViewSvc)(nil)

type mockReportSvc struct{}

func (m *mockReportSvc) Preview(_ context.Context, _ int64, _ time.Time) (*dto.ReportPreview, error) {
	return nil, nil
}
func (m *mockReportSvc) ExportMarkdown(_ context.Context, _ int64, _ time.Time) ([]byte, error) {
	return nil, nil
}

var _ service.ReportService = (*mockReportSvc)(nil)

// ---------------------------------------------------------------------------
// Tests for buildPermTestRouter infrastructure
// ---------------------------------------------------------------------------

// buildPermTestRouterFull creates a minimal router for any HTTP method/path.
// The terminal handler always returns 200 — only the permission gate is under test.
func buildPermTestRouterFull(method, path, code string, permCodes []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Handle(method, path,
		func(c *gin.Context) {
			c.Set("teamID", uint(1))
			c.Set("permCodes", permCodes)
			c.Next()
		},
		middleware.RequirePermission(code, nil),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	return r
}

// ---------------------------------------------------------------------------
// Infrastructure smoke tests
// ---------------------------------------------------------------------------

func TestBuildPermTestRouter_HasPermission(t *testing.T) {
	r := buildPermTestRouter("main_item:create", []string{"main_item:create"}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBuildPermTestRouter_NoPermission(t *testing.T) {
	r := buildPermTestRouter("main_item:create", []string{}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---------------------------------------------------------------------------
// Permission matrix tests — 12 endpoints × 2 cases each
// ---------------------------------------------------------------------------

func runPermCases(t *testing.T, method, path, code string) {
	t.Helper()
	cases := []struct {
		name      string
		permCodes []string
		wantCode  int
	}{
		{"has_permission", []string{code}, http.StatusOK},
		{"no_permission", []string{}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := buildPermTestRouterFull(method, path, code, tc.permCodes)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, path, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tc.wantCode, w.Code)
		})
	}
}

func TestPermMatrix_MainItemCreate(t *testing.T) {
	runPermCases(t, http.MethodPost, "/test", "main_item:create")
}

func TestPermMatrix_MainItemArchive(t *testing.T) {
	runPermCases(t, http.MethodPost, "/test/1/archive", "main_item:archive")
}

func TestPermMatrix_MainItemChangeStatus(t *testing.T) {
	runPermCases(t, http.MethodPut, "/test/1/status", "main_item:change_status")
}

func TestPermMatrix_TeamInvite(t *testing.T) {
	runPermCases(t, http.MethodPost, "/test", "team:invite")
}

func TestPermMatrix_TeamRemove(t *testing.T) {
	runPermCases(t, http.MethodDelete, "/test/1", "team:remove")
}

func TestPermMatrix_TeamTransfer(t *testing.T) {
	runPermCases(t, http.MethodPut, "/test", "team:transfer")
}

func TestPermMatrix_ProgressCreate(t *testing.T) {
	runPermCases(t, http.MethodPost, "/test", "progress:create")
}

func TestPermMatrix_ProgressUpdate(t *testing.T) {
	runPermCases(t, http.MethodPatch, "/test/1/completion", "progress:update")
}

func TestPermMatrix_ItemPoolSubmit(t *testing.T) {
	runPermCases(t, http.MethodPost, "/test", "item_pool:submit")
}

func TestPermMatrix_ItemPoolReview(t *testing.T) {
	runPermCases(t, http.MethodPost, "/test/1/assign", "item_pool:review")
}

func TestPermMatrix_WeeklyView(t *testing.T) {
	runPermCases(t, http.MethodGet, "/test", "view:weekly")
}

func TestPermMatrix_ReportExport(t *testing.T) {
	runPermCases(t, http.MethodGet, "/test", "report:export")
}

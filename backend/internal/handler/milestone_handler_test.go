package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
)

// ---------------------------------------------------------------------------
// Mock MilestoneService for handler tests
// ---------------------------------------------------------------------------

type mockMilestoneService struct {
	createResult struct {
		item *model.Milestone
		err  error
	}
	getResult struct {
		item *model.Milestone
		err  error
	}
	getByBizKeyResult struct {
		item *model.Milestone
		err  error
	}
	updateResult struct {
		err error
	}
	listByMapResult struct {
		items []model.Milestone
		err   error
	}
	listByTeamResult struct {
		items []model.Milestone
		err   error
	}
	deleteResult struct {
		err error
	}
	changeStatusResult struct {
		item *model.Milestone
		err  error
	}
	availableTransitionsResult struct {
		transitions []string
		err         error
	}

	// capture calls
	createCalled               bool
	lastTeamBizKey             int64
	lastMapBizKey              int64
	lastCreateReq              dto.MilestoneCreateReq
	getCalled                  bool
	lastMilestoneID            uint
	getByBizKeyCalled          bool
	lastBizKey                 int64
	updateCalled               bool
	lastUpdateID               uint
	lastUpdateReq              dto.MilestoneUpdateReq
	listByMapCalled            bool
	lastListByMapKey           int64
	listByTeamCalled           bool
	lastTeamFilter             dto.MilestoneTeamFilter
	deleteCalled               bool
	lastDeleteID               uint
	changeStatusCalled         bool
	lastNewStatus              string
	availableTransitionsCalled bool
}

func (m *mockMilestoneService) Create(_ context.Context, teamBizKey, mapBizKey int64, req dto.MilestoneCreateReq) (*model.Milestone, error) {
	m.createCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastMapBizKey = mapBizKey
	m.lastCreateReq = req
	return m.createResult.item, m.createResult.err
}

func (m *mockMilestoneService) Update(_ context.Context, teamBizKey int64, milestoneID uint, req dto.MilestoneUpdateReq) error {
	m.updateCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastUpdateID = milestoneID
	m.lastUpdateReq = req
	return m.updateResult.err
}

func (m *mockMilestoneService) Get(_ context.Context, milestoneID uint) (*model.Milestone, error) {
	m.getCalled = true
	m.lastMilestoneID = milestoneID
	return m.getResult.item, m.getResult.err
}

func (m *mockMilestoneService) GetByBizKey(_ context.Context, bizKey int64) (*model.Milestone, error) {
	m.getByBizKeyCalled = true
	m.lastBizKey = bizKey
	return m.getByBizKeyResult.item, m.getByBizKeyResult.err
}

func (m *mockMilestoneService) ListByMap(_ context.Context, mapBizKey int64) ([]model.Milestone, error) {
	m.listByMapCalled = true
	m.lastListByMapKey = mapBizKey
	return m.listByMapResult.items, m.listByMapResult.err
}

func (m *mockMilestoneService) ListByTeam(_ context.Context, teamBizKey int64, filter dto.MilestoneTeamFilter) ([]model.Milestone, error) {
	m.listByTeamCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastTeamFilter = filter
	return m.listByTeamResult.items, m.listByTeamResult.err
}

func (m *mockMilestoneService) Delete(_ context.Context, teamBizKey int64, milestoneID uint) error {
	m.deleteCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastDeleteID = milestoneID
	return m.deleteResult.err
}

func (m *mockMilestoneService) ChangeStatus(_ context.Context, teamBizKey int64, milestoneID uint, newStatus string) (*model.Milestone, error) {
	m.changeStatusCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastMilestoneID = milestoneID
	m.lastNewStatus = newStatus
	return m.changeStatusResult.item, m.changeStatusResult.err
}

func (m *mockMilestoneService) AvailableTransitions(_ context.Context, teamBizKey int64, milestoneID uint) ([]string, error) {
	m.availableTransitionsCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastMilestoneID = milestoneID
	return m.availableTransitionsResult.transitions, m.availableTransitionsResult.err
}

var _ service.MilestoneService = (*mockMilestoneService)(nil)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testMilestone creates a sample Milestone model for tests.
func testMilestone() *model.Milestone {
	return &model.Milestone{
		BaseModel:       model.BaseModel{ID: 1, BizKey: 200},
		TeamKey:         10,
		MilestoneMapKey: 100,
		MilestoneName:   "Test Milestone",
		MilestoneDesc:   "A test milestone",
		MilestoneStatus: "not_started",
	}
}

// mockMainItemRepoForMilestoneHandler is a test double for MainItemRepo.
type mockMainItemRepoForMilestoneHandler struct {
	items []model.MainItem
}

func (m *mockMainItemRepoForMilestoneHandler) FindByMilestoneKey(_ context.Context, _ int64) ([]model.MainItem, error) {
	return m.items, nil
}

// Satisfy repository.MainItemRepo interface with stubs.
func (m *mockMainItemRepoForMilestoneHandler) Create(_ context.Context, _ *model.MainItem) error {
	return nil
}
func (m *mockMainItemRepoForMilestoneHandler) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMilestoneHandler) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMilestoneHandler) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMilestoneHandler) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (m *mockMainItemRepoForMilestoneHandler) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemRepoForMilestoneHandler) NextCode(_ context.Context, _ int64) (string, error) {
	return "", nil
}
func (m *mockMainItemRepoForMilestoneHandler) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockMainItemRepoForMilestoneHandler) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMilestoneHandler) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMilestoneHandler) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMilestoneHandler) SoftDelete(_ context.Context, _ uint) error { return nil }
func (m *mockMainItemRepoForMilestoneHandler) CascadeSoftDelete(_ context.Context, _ uint, _ []uint, _ []model.StatusHistory) error {
	return nil
}
func (m *mockMainItemRepoForMilestoneHandler) CountByMilestoneKey(_ context.Context, _ int64) (int64, error) {
	return int64(len(m.items)), nil
}
func (m *mockMainItemRepoForMilestoneHandler) ClearMilestoneKeyByMilestone(_ context.Context, _ int64) error {
	return nil
}
func (m *mockMainItemRepoForMilestoneHandler) ClearMilestoneKeyByMap(_ context.Context, _ []int64) error {
	return nil
}

var _ repository.MainItemRepo = (*mockMainItemRepoForMilestoneHandler)(nil)

// depsWithMilestoneSvc wires a mock MilestoneService into test deps.
func depsWithMilestoneSvc(t *testing.T, svc *mockMilestoneService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{RoleKey: func() *int64 { v := int64(1); return &v }()}}

	mainItemRepo := &mockMainItemRepoForMilestoneHandler{}
	deps.Milestone = NewMilestoneHandler(svc, mainItemRepo)
	return deps
}

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/milestone-maps/:mapId/milestones (Create)
// ---------------------------------------------------------------------------

func TestMilestoneCreate_Success(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	svc.createResult.item = ms

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"milestoneName":"Test Milestone","milestoneDesc":"desc","expectedEndDate":"2026-07-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/milestone-maps/100/milestones", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test Milestone", data["milestoneName"])
	assert.Equal(t, "not_started", data["milestoneStatus"])

	assert.True(t, svc.createCalled)
	assert.Equal(t, int64(10), svc.lastTeamBizKey)
	assert.Equal(t, int64(100), svc.lastMapBizKey)
	assert.Equal(t, "Test Milestone", svc.lastCreateReq.MilestoneName)
}

func TestMilestoneCreate_MissingName(t *testing.T) {
	svc := &mockMilestoneService{}
	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"milestoneDesc":"no name"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/milestone-maps/100/milestones", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.createCalled)
}

func TestMilestoneCreate_BR5TerminalMap(t *testing.T) {
	svc := &mockMilestoneService{}
	svc.createResult.err = apperrors.ErrMapIsTerminal

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"milestoneName":"Test","expectedEndDate":"2026-07-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/milestone-maps/100/milestones", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMilestoneCreate_ServiceError(t *testing.T) {
	svc := &mockMilestoneService{}
	svc.createResult.err = errors.New("unexpected")

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"milestoneName":"Test","expectedEndDate":"2026-07-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/milestone-maps/100/milestones", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/milestone-maps/:mapId/milestones (ListByMap)
// ---------------------------------------------------------------------------

func TestMilestoneListByMap_Success(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	svc.listByMapResult.items = []model.Milestone{*ms}

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestone-maps/100/milestones", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	items, ok := data["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 1)

	assert.True(t, svc.listByMapCalled)
	assert.Equal(t, int64(100), svc.lastListByMapKey)
}

func TestMilestoneListByMap_EmptyResult(t *testing.T) {
	svc := &mockMilestoneService{}
	svc.listByMapResult.items = []model.Milestone{}

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestone-maps/100/milestones", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/milestones (ListByTeam)
// ---------------------------------------------------------------------------

func TestMilestoneListByTeam_Success(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	svc.listByTeamResult.items = []model.Milestone{*ms}

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestones", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	items, ok := data["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 1)

	assert.True(t, svc.listByTeamCalled)
	assert.Equal(t, int64(10), svc.lastTeamBizKey)
}

func TestMilestoneListByTeam_WithFilters(t *testing.T) {
	svc := &mockMilestoneService{}
	svc.listByTeamResult.items = []model.Milestone{}

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestones?status=not_started&excludeCancelled=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.listByTeamCalled)
	assert.NotNil(t, svc.lastTeamFilter.Status)
	assert.Equal(t, "not_started", *svc.lastTeamFilter.Status)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/milestones/:milestoneId (Get)
// ---------------------------------------------------------------------------

func TestMilestoneGet_Success(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	svc.getByBizKeyResult.item = ms

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestones/200", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test Milestone", data["milestoneName"])
	assert.Equal(t, "not_started", data["milestoneStatus"])

	assert.True(t, svc.getByBizKeyCalled)
	assert.Equal(t, int64(200), svc.lastBizKey)
}

func TestMilestoneGet_NotFound(t *testing.T) {
	svc := &mockMilestoneService{}
	svc.getByBizKeyResult.err = apperrors.ErrNotFound

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestones/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMilestoneGet_InvalidMilestoneId(t *testing.T) {
	svc := &mockMilestoneService{}
	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestones/invalid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.getByBizKeyCalled)
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/milestones/:milestoneId (Update)
// ---------------------------------------------------------------------------

func TestMilestoneUpdate_Success(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	svc.getByBizKeyResult.item = ms

	updated := *ms
	updated.MilestoneName = "Updated Milestone"
	svc.getResult.item = &updated

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"milestoneName":"Updated Milestone"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestones/200", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Updated Milestone", data["milestoneName"])

	assert.True(t, svc.getByBizKeyCalled)
	assert.True(t, svc.updateCalled)
	assert.Equal(t, uint(1), svc.lastUpdateID)
}

func TestMilestoneUpdate_NotFound(t *testing.T) {
	svc := &mockMilestoneService{}
	svc.getByBizKeyResult.err = apperrors.ErrNotFound

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"milestoneName":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestones/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMilestoneUpdate_BR5TerminalMap(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	svc.getByBizKeyResult.item = ms
	svc.updateResult.err = apperrors.ErrMapIsTerminal

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"milestoneName":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestones/200", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DELETE /teams/:teamId/milestones/:milestoneId (Delete)
// ---------------------------------------------------------------------------

func TestMilestoneDelete_Success(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	ms.MilestoneStatus = "not_started"
	svc.getByBizKeyResult.item = ms

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/10/milestones/200", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "deleted", data["message"])

	assert.True(t, svc.deleteCalled)
	assert.Equal(t, uint(1), svc.lastDeleteID)
}

func TestMilestoneDelete_BR4Violation(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	ms.MilestoneStatus = "in_progress"
	svc.getByBizKeyResult.item = ms
	svc.deleteResult.err = apperrors.ErrMilestoneCannotDelete

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/10/milestones/200", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMilestoneDelete_NotFound(t *testing.T) {
	svc := &mockMilestoneService{}
	svc.getByBizKeyResult.err = apperrors.ErrNotFound

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/10/milestones/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/milestones/:milestoneId/status (ChangeStatus)
// ---------------------------------------------------------------------------

func TestMilestoneChangeStatus_Success(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	svc.getByBizKeyResult.item = ms

	updated := *ms
	updated.MilestoneStatus = "in_progress"
	svc.changeStatusResult.item = &updated

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"in_progress"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestones/200/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "in_progress", data["milestoneStatus"])

	assert.True(t, svc.changeStatusCalled)
	assert.Equal(t, "in_progress", svc.lastNewStatus)
}

func TestMilestoneChangeStatus_InvalidTransition(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	ms.MilestoneStatus = "not_started"
	svc.getByBizKeyResult.item = ms
	svc.changeStatusResult.err = apperrors.ErrInvalidStatus

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"completed"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestones/200/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestMilestoneChangeStatus_BR1NonTerminalItems(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	svc.getByBizKeyResult.item = ms
	svc.changeStatusResult.err = apperrors.ErrMilestoneHasNonTerminalItems

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"completed"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestones/200/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMilestoneChangeStatus_BR5TerminalMap(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	svc.getByBizKeyResult.item = ms
	svc.changeStatusResult.err = apperrors.ErrMapIsTerminal

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"in_progress"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestones/200/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMilestoneChangeStatus_MissingStatus(t *testing.T) {
	svc := &mockMilestoneService{}
	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestones/200/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.changeStatusCalled)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/milestones/:milestoneId/available-transitions
// ---------------------------------------------------------------------------

func TestMilestoneAvailableTransitions_Success(t *testing.T) {
	svc := &mockMilestoneService{}
	ms := testMilestone()
	svc.getByBizKeyResult.item = ms
	svc.availableTransitionsResult.transitions = []string{"in_progress", "cancelled"} //nolint:misspell // domain status value

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestones/200/available-transitions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	transitions, ok := data["transitions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, transitions, 2)
	assert.Equal(t, "in_progress", transitions[0])
	assert.Equal(t, "cancelled", transitions[1]) //nolint:misspell // domain status value

	assert.True(t, svc.availableTransitionsCalled)
}

func TestMilestoneAvailableTransitions_NotFound(t *testing.T) {
	svc := &mockMilestoneService{}
	svc.getByBizKeyResult.err = apperrors.ErrNotFound

	deps := depsWithMilestoneSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestones/999/available-transitions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: VO enrichment
// ---------------------------------------------------------------------------

func TestBuildMilestoneVO_ComputedFields(t *testing.T) {
	ms := testMilestone()

	mainItemRepo := &mockMainItemRepoForMilestoneHandler{
		items: []model.MainItem{
			{Completion: 40},
			{Completion: 80},
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	v := buildMilestoneVO(ms, mainItemRepo, c)

	assert.Equal(t, 60.0, v.Completion) // (40+80)/2
	assert.Equal(t, 2, v.RelatedMICount)
	assert.Equal(t, "Test Milestone", v.MilestoneName)
	assert.Equal(t, "not_started", v.MilestoneStatus)
}

func TestBuildMilestoneVO_NoItems(t *testing.T) {
	ms := testMilestone()

	mainItemRepo := &mockMainItemRepoForMilestoneHandler{
		items: []model.MainItem{},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	v := buildMilestoneVO(ms, mainItemRepo, c)

	assert.Equal(t, 0.0, v.Completion)
	assert.Equal(t, 0, v.RelatedMICount)
}

func TestBuildMilestoneVOs_EmptySlice(t *testing.T) {
	mainItemRepo := &mockMainItemRepoForMilestoneHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	result := buildMilestoneVOs([]model.Milestone{}, mainItemRepo, c)

	assert.Equal(t, []vo.MilestoneVO{}, result)
}

// Suppress unused import for fmt (used in test helper compilation)
var _ = fmt.Sprintf

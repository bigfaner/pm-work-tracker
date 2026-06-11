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
)

// ---------------------------------------------------------------------------
// Mock MilestoneMapService for handler tests
// ---------------------------------------------------------------------------

type mockMilestoneMapService struct {
	createResult struct {
		item *model.MilestoneMap
		err  error
	}
	listResult struct {
		page *dto.PageResult[model.MilestoneMap]
		err  error
	}
	getResult struct {
		item *model.MilestoneMap
		err  error
	}
	getByBizKeyResult struct {
		item *model.MilestoneMap
		err  error
	}
	updateResult struct {
		err error
	}
	deleteResult struct {
		err error
	}
	changeStatusResult struct {
		item *model.MilestoneMap
		err  error
	}
	availableTransitionsResult struct {
		transitions []string
		err         error
	}

	// capture calls
	createCalled   bool
	lastTeamBizKey int64
	lastCreatorKey int64
	lastCreateReq  dto.MilestoneMapCreateReq

	listCalled bool
	lastFilter dto.MilestoneMapFilter
	lastPage   dto.Pagination

	getCalled bool
	lastMapID uint

	getByBizKeyCalled bool
	lastBizKey        int64

	updateCalled  bool
	lastUpdateID  uint
	lastUpdateReq dto.MilestoneMapUpdateReq

	deleteCalled bool
	lastDeleteID uint

	changeStatusCalled bool
	lastNewStatus      string

	availableTransitionsCalled bool
}

func (m *mockMilestoneMapService) Create(_ context.Context, teamBizKey int64, creatorBizKey int64, req dto.MilestoneMapCreateReq) (*model.MilestoneMap, error) {
	m.createCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastCreatorKey = creatorBizKey
	m.lastCreateReq = req
	return m.createResult.item, m.createResult.err
}

func (m *mockMilestoneMapService) Update(_ context.Context, teamBizKey int64, mapID uint, req dto.MilestoneMapUpdateReq) error {
	m.updateCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastUpdateID = mapID
	m.lastUpdateReq = req
	return m.updateResult.err
}

func (m *mockMilestoneMapService) Get(_ context.Context, mapID uint) (*model.MilestoneMap, error) {
	m.getCalled = true
	m.lastMapID = mapID
	return m.getResult.item, m.getResult.err
}

func (m *mockMilestoneMapService) GetByBizKey(_ context.Context, bizKey int64) (*model.MilestoneMap, error) {
	m.getByBizKeyCalled = true
	m.lastBizKey = bizKey
	return m.getByBizKeyResult.item, m.getByBizKeyResult.err
}

func (m *mockMilestoneMapService) List(_ context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error) {
	m.listCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastFilter = filter
	m.lastPage = page
	return m.listResult.page, m.listResult.err
}

func (m *mockMilestoneMapService) Delete(_ context.Context, teamBizKey int64, mapID uint) error {
	m.deleteCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastDeleteID = mapID
	return m.deleteResult.err
}

func (m *mockMilestoneMapService) ChangeStatus(_ context.Context, teamBizKey int64, mapID uint, newStatus string) (*model.MilestoneMap, error) {
	m.changeStatusCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastMapID = mapID
	m.lastNewStatus = newStatus
	return m.changeStatusResult.item, m.changeStatusResult.err
}

func (m *mockMilestoneMapService) AvailableTransitions(_ context.Context, teamBizKey int64, mapID uint) ([]string, error) {
	m.availableTransitionsCalled = true
	m.lastTeamBizKey = teamBizKey
	m.lastMapID = mapID
	return m.availableTransitionsResult.transitions, m.availableTransitionsResult.err
}

var _ service.MilestoneMapService = (*mockMilestoneMapService)(nil)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testMilestoneMap creates a sample MilestoneMap model for tests.
func testMilestoneMap() *model.MilestoneMap {
	return &model.MilestoneMap{
		BaseModel:   model.BaseModel{ID: 1, BizKey: 100},
		TeamKey:     10,
		CreatorKey:  5,
		AssigneeKey: 6,
		MapName:     "Test Map",
		MapDesc:     "A test milestone map",
		MapStatus:   "planning",
	}
}

// mockUserRepoForMilestoneMap is a test double for UserRepo that returns predictable names.
type mockUserRepoForMilestoneMap struct {
	users map[int64]*model.User
}

func (m *mockUserRepoForMilestoneMap) FindByBizKey(_ context.Context, bizKey int64) (*model.User, error) {
	if u, ok := m.users[bizKey]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("not found")
}

// Satisfy repository.UserRepo interface with stubs.
func (m *mockUserRepoForMilestoneMap) FindByID(_ context.Context, _ uint) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepoForMilestoneMap) FindByUsername(_ context.Context, _ string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepoForMilestoneMap) List(_ context.Context) ([]*model.User, error) { return nil, nil }
func (m *mockUserRepoForMilestoneMap) Update(_ context.Context, _ *model.User) error { return nil }
func (m *mockUserRepoForMilestoneMap) Create(_ context.Context, _ *model.User) error { return nil }
func (m *mockUserRepoForMilestoneMap) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.User, error) {
	return nil, nil
}
func (m *mockUserRepoForMilestoneMap) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.User, error) {
	return nil, nil
}
func (m *mockUserRepoForMilestoneMap) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepoForMilestoneMap) SearchAvailable(_ context.Context, _ int64, _ string, _ int) ([]*model.User, error) {
	return nil, nil
}
func (m *mockUserRepoForMilestoneMap) SoftDelete(_ context.Context, _ *model.User) error { return nil }

var _ repository.UserRepo = (*mockUserRepoForMilestoneMap)(nil)

// mockMilestoneRepoForMapHandler is a test double for MilestoneRepo.
type mockMilestoneRepoForMapHandler struct {
	milestones []model.Milestone
}

func (m *mockMilestoneRepoForMapHandler) ListByMap(_ context.Context, _ int64) ([]model.Milestone, error) {
	return m.milestones, nil
}

func (m *mockMilestoneRepoForMapHandler) Create(_ context.Context, _ *model.Milestone) error {
	return nil
}
func (m *mockMilestoneRepoForMapHandler) FindByID(_ context.Context, _ uint) (*model.Milestone, error) {
	return nil, nil
}
func (m *mockMilestoneRepoForMapHandler) FindByBizKey(_ context.Context, _ int64) (*model.Milestone, error) {
	return nil, nil
}
func (m *mockMilestoneRepoForMapHandler) FindBatchByBizKeys(_ context.Context, _ []int64) (map[int64]*model.Milestone, error) {
	return nil, nil
}
func (m *mockMilestoneRepoForMapHandler) Update(_ context.Context, _ *model.Milestone, _ map[string]interface{}) error {
	return nil
}
func (m *mockMilestoneRepoForMapHandler) ListByTeam(_ context.Context, _ int64, _ dto.MilestoneTeamFilter) ([]model.Milestone, error) {
	return nil, nil
}
func (m *mockMilestoneRepoForMapHandler) SoftDelete(_ context.Context, _ uint) error { return nil }
func (m *mockMilestoneRepoForMapHandler) SoftDeleteByMap(_ context.Context, _ int64) error {
	return nil
}
func (m *mockMilestoneRepoForMapHandler) ExistsByNameAndMap(_ context.Context, _ int64, _ string, _ *uint) (bool, error) {
	return false, nil
}

var _ repository.MilestoneRepo = (*mockMilestoneRepoForMapHandler)(nil)

// mockMainItemRepoForMapHandler is a test double for MainItemRepo.
type mockMainItemRepoForMapHandler struct {
	items []model.MainItem
}

func (m *mockMainItemRepoForMapHandler) FindByMilestoneKey(_ context.Context, _ int64) ([]model.MainItem, error) {
	return m.items, nil
}

// Satisfy repository.MainItemRepo interface with stubs.
func (m *mockMainItemRepoForMapHandler) Create(_ context.Context, _ *model.MainItem) error {
	return nil
}
func (m *mockMainItemRepoForMapHandler) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMapHandler) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMapHandler) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMapHandler) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (m *mockMainItemRepoForMapHandler) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemRepoForMapHandler) NextCode(_ context.Context, _ int64) (string, error) {
	return "", nil
}
func (m *mockMainItemRepoForMapHandler) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockMainItemRepoForMapHandler) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMapHandler) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMapHandler) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForMapHandler) SoftDelete(_ context.Context, _ uint) error { return nil }
func (m *mockMainItemRepoForMapHandler) CascadeSoftDelete(_ context.Context, _ uint, _ []uint, _ []model.StatusHistory) error {
	return nil
}
func (m *mockMainItemRepoForMapHandler) CountByMilestoneKey(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockMainItemRepoForMapHandler) ClearMilestoneKeyByMilestone(_ context.Context, _ int64) error {
	return nil
}
func (m *mockMainItemRepoForMapHandler) ClearMilestoneKeyByMap(_ context.Context, _ []int64) error {
	return nil
}

var _ repository.MainItemRepo = (*mockMainItemRepoForMapHandler)(nil)

// depsWithMilestoneMapSvc wires a mock MilestoneMapService into test deps.
func depsWithMilestoneMapSvc(t *testing.T, svc *mockMilestoneMapService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{RoleKey: func() *int64 { v := int64(1); return &v }()}}

	userRepo := &mockUserRepoForMilestoneMap{
		users: map[int64]*model.User{
			5: {DisplayName: "Creator User"},
			6: {DisplayName: "Assignee User"},
		},
	}
	milestoneRepo := &mockMilestoneRepoForMapHandler{}
	mainItemRepo := &mockMainItemRepoForMapHandler{}

	deps.MilestoneMap = NewMilestoneMapHandler(svc, userRepo, milestoneRepo, mainItemRepo)
	return deps
}

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/milestone-maps (Create)
// ---------------------------------------------------------------------------

func TestMilestoneMapCreate_Success(t *testing.T) {
	svc := &mockMilestoneMapService{}
	m := testMilestoneMap()
	svc.createResult.item = m
	// GetByBizKey is called for enrichment in the response — not needed for Create

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"mapName":"Test Map","mapDesc":"desc","assigneeBizKey":"6"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/milestone-maps", strings.NewReader(body))
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
	assert.Equal(t, "Test Map", data["mapName"])
	assert.Equal(t, "planning", data["mapStatus"])

	assert.True(t, svc.createCalled)
	assert.Equal(t, int64(10), svc.lastTeamBizKey)
	assert.Equal(t, int64(5), svc.lastCreatorKey)
	assert.Equal(t, "Test Map", svc.lastCreateReq.MapName)
}

func TestMilestoneMapCreate_MissingName(t *testing.T) {
	svc := &mockMilestoneMapService{}
	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"mapDesc":"no name"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/milestone-maps", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.createCalled)
}

func TestMilestoneMapCreate_ServiceError(t *testing.T) {
	svc := &mockMilestoneMapService{}
	svc.createResult.err = errors.New("unexpected")

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"mapName":"Test","assigneeBizKey":"6"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/milestone-maps", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/milestone-maps (List)
// ---------------------------------------------------------------------------

func TestMilestoneMapList_Success(t *testing.T) {
	svc := &mockMilestoneMapService{}
	m := testMilestoneMap()
	svc.listResult.page = &dto.PageResult[model.MilestoneMap]{
		Items: []model.MilestoneMap{*m},
		Total: 1,
		Page:  1,
		Size:  20,
	}

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestone-maps?page=1&pageSize=20", nil)
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

	assert.True(t, svc.listCalled)
	assert.Equal(t, int64(10), svc.lastTeamBizKey)
}

func TestMilestoneMapList_EmptyResult(t *testing.T) {
	svc := &mockMilestoneMapService{}
	svc.listResult.page = &dto.PageResult[model.MilestoneMap]{
		Items: []model.MilestoneMap{},
		Total: 0,
		Page:  1,
		Size:  20,
	}

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestone-maps", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/milestone-maps/:mapId (Get)
// ---------------------------------------------------------------------------

func TestMilestoneMapGet_Success(t *testing.T) {
	svc := &mockMilestoneMapService{}
	m := testMilestoneMap()
	svc.getByBizKeyResult.item = m

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestone-maps/100", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test Map", data["mapName"])
	assert.Equal(t, "Creator User", data["creatorName"])
	assert.Equal(t, "Assignee User", data["assigneeName"])

	assert.True(t, svc.getByBizKeyCalled)
	assert.Equal(t, int64(100), svc.lastBizKey)
}

func TestMilestoneMapGet_NotFound(t *testing.T) {
	svc := &mockMilestoneMapService{}
	svc.getByBizKeyResult.err = apperrors.ErrNotFound

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestone-maps/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMilestoneMapGet_InvalidMapId(t *testing.T) {
	svc := &mockMilestoneMapService{}
	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestone-maps/invalid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.getByBizKeyCalled)
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/milestone-maps/:mapId (Update)
// ---------------------------------------------------------------------------

func TestMilestoneMapUpdate_Success(t *testing.T) {
	svc := &mockMilestoneMapService{}
	m := testMilestoneMap()
	svc.getByBizKeyResult.item = m

	updated := *m
	updated.MapName = "Updated Map"
	svc.getResult.item = &updated

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"mapName":"Updated Map"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestone-maps/100", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Updated Map", data["mapName"])

	assert.True(t, svc.getByBizKeyCalled)
	assert.True(t, svc.updateCalled)
	assert.Equal(t, uint(1), svc.lastUpdateID)
}

func TestMilestoneMapUpdate_NotFound(t *testing.T) {
	svc := &mockMilestoneMapService{}
	svc.getByBizKeyResult.err = apperrors.ErrNotFound

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"mapName":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestone-maps/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DELETE /teams/:teamId/milestone-maps/:mapId (Delete)
// ---------------------------------------------------------------------------

func TestMilestoneMapDelete_Success(t *testing.T) {
	svc := &mockMilestoneMapService{}
	m := testMilestoneMap()
	m.MapStatus = "planning"
	svc.getByBizKeyResult.item = m

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/10/milestone-maps/100", nil)
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

func TestMilestoneMapDelete_BR4Violation(t *testing.T) {
	svc := &mockMilestoneMapService{}
	m := testMilestoneMap()
	m.MapStatus = "executing"
	svc.getByBizKeyResult.item = m
	svc.deleteResult.err = apperrors.ErrMapCannotDelete

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/10/milestone-maps/100", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/milestone-maps/:mapId/status (ChangeStatus)
// ---------------------------------------------------------------------------

func TestMilestoneMapChangeStatus_Success(t *testing.T) {
	svc := &mockMilestoneMapService{}
	m := testMilestoneMap()
	svc.getByBizKeyResult.item = m

	updated := *m
	updated.MapStatus = "reviewed"
	svc.changeStatusResult.item = &updated

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"reviewed"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestone-maps/100/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "reviewed", data["mapStatus"])

	assert.True(t, svc.changeStatusCalled)
	assert.Equal(t, "reviewed", svc.lastNewStatus)
}

func TestMilestoneMapChangeStatus_InvalidTransition(t *testing.T) {
	svc := &mockMilestoneMapService{}
	m := testMilestoneMap()
	m.MapStatus = "planning"
	svc.getByBizKeyResult.item = m
	svc.changeStatusResult.err = apperrors.ErrInvalidStatus

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"completed"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestone-maps/100/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestMilestoneMapChangeStatus_MissingStatus(t *testing.T) {
	svc := &mockMilestoneMapService{}
	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/milestone-maps/100/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.changeStatusCalled)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/milestone-maps/:mapId/available-transitions
// ---------------------------------------------------------------------------

func TestMilestoneMapAvailableTransitions_Success(t *testing.T) {
	svc := &mockMilestoneMapService{}
	m := testMilestoneMap()
	svc.getByBizKeyResult.item = m
	svc.availableTransitionsResult.transitions = []string{"reviewed", "cancelled"} //nolint:misspell // domain status value

	deps := depsWithMilestoneMapSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/milestone-maps/100/available-transitions", nil)
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
	assert.Equal(t, "reviewed", transitions[0])
	assert.Equal(t, "cancelled", transitions[1]) //nolint:misspell // domain status value

	assert.True(t, svc.availableTransitionsCalled)
}

// ---------------------------------------------------------------------------
// Tests: VO enrichment
// ---------------------------------------------------------------------------

func TestBuildMilestoneMapVO_UserEnrichment(t *testing.T) {
	m := testMilestoneMap()

	userRepo := &mockUserRepoForMilestoneMap{
		users: map[int64]*model.User{
			5: {DisplayName: "Alice"},
			6: {DisplayName: "Bob"},
		},
	}
	milestoneRepo := &mockMilestoneRepoForMapHandler{}
	mainItemRepo := &mockMainItemRepoForMapHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	v := buildMilestoneMapVO(m, userRepo, milestoneRepo, mainItemRepo, c)

	assert.Equal(t, "Alice", v.CreatorName)
	assert.Equal(t, "Bob", v.AssigneeName)
	assert.Equal(t, "Test Map", v.MapName)
	assert.Equal(t, "planning", v.MapStatus)
}

func TestBuildMilestoneMapVO_ComputedFields(t *testing.T) {
	m := testMilestoneMap()

	userRepo := &mockUserRepoForMilestoneMap{users: map[int64]*model.User{
		5: {DisplayName: "Creator"}, 6: {DisplayName: "Assignee"},
	}}
	milestoneRepo := &mockMilestoneRepoForMapHandler{
		milestones: []model.Milestone{
			{BaseModel: model.BaseModel{BizKey: 200}, MilestoneStatus: "not_started"},
		},
	}
	mainItemRepo := &mockMainItemRepoForMapHandler{
		items: []model.MainItem{
			{Completion: 50},
			{Completion: 70},
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	v := buildMilestoneMapVO(m, userRepo, milestoneRepo, mainItemRepo, c)

	assert.Equal(t, 1, v.MilestoneCount)
	assert.Equal(t, 2, v.ItemCount)
	assert.Equal(t, 60.0, v.OverallProgress) // (50+70)/2
}

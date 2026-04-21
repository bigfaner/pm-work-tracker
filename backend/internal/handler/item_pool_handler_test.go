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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

// ---------------------------------------------------------------------------
// Mock ItemPoolService for handler tests
// ---------------------------------------------------------------------------

type mockItemPoolService struct {
	submitResult struct {
		item *model.ItemPool
		err  error
	}
	listResult struct {
		page *dto.PageResult[model.ItemPool]
		err  error
	}
	getResult struct {
		item *model.ItemPool
		err  error
	}
	assignResult struct {
		err error
	}
	rejectResult struct {
		err error
	}

	// capture calls
	submitCalled    bool
	lastTeamID     uint
	lastSubmitterID uint
	lastSubmitReq  dto.SubmitItemPoolReq

	listCalled bool
	lastFilter dto.ItemPoolFilter
	lastPage   dto.Pagination

	getCalled   bool
	lastPoolID  uint

	assignCalled bool
	lastAssignID uint
	lastAssignReq dto.AssignItemPoolReq

	rejectCalled bool
	lastRejectID uint
	lastReason   string
}

func (m *mockItemPoolService) Submit(_ context.Context, teamID, submitterID uint, req dto.SubmitItemPoolReq) (*model.ItemPool, error) {
	m.submitCalled = true
	m.lastTeamID = teamID
	m.lastSubmitterID = submitterID
	m.lastSubmitReq = req
	return m.submitResult.item, m.submitResult.err
}

func (m *mockItemPoolService) List(_ context.Context, teamID uint, filter dto.ItemPoolFilter, page dto.Pagination) (*dto.PageResult[model.ItemPool], error) {
	m.listCalled = true
	m.lastTeamID = teamID
	m.lastFilter = filter
	m.lastPage = page
	return m.listResult.page, m.listResult.err
}

func (m *mockItemPoolService) Get(_ context.Context, teamID, poolID uint) (*model.ItemPool, error) {
	m.getCalled = true
	m.lastTeamID = teamID
	m.lastPoolID = poolID
	return m.getResult.item, m.getResult.err
}

func (m *mockItemPoolService) Assign(_ context.Context, teamID, pmID, poolID uint, req dto.AssignItemPoolReq) error {
	m.assignCalled = true
	m.lastTeamID = teamID
	m.lastPoolID = poolID
	m.lastAssignReq = req
	return m.assignResult.err
}

func (m *mockItemPoolService) ConvertToMain(_ context.Context, teamID, pmID, poolItemID uint, req dto.ConvertToMainItemReq) (*model.MainItem, error) {
	return nil, nil
}

func (m *mockItemPoolService) Reject(_ context.Context, teamID, pmID, poolID uint, reason string) error {
	m.rejectCalled = true
	m.lastTeamID = teamID
	m.lastPoolID = poolID
	m.lastReason = reason
	return m.rejectResult.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func ptrUint(v uint) *uint { return &v }

// depsWithItemPoolSvc wires a mock ItemPoolService into test deps.
func depsWithItemPoolSvc(t *testing.T, svc *mockItemPoolService, userRepo repository.UserRepo) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm", RoleID: ptrUint(1)}}
	deps.ItemPool = NewItemPoolHandler(svc, userRepo, &mockMainItemRepoForPool{})
	return deps
}

// depsWithItemPoolMemberRole wires a member (non-PM) role for testing RBAC.
func depsWithItemPoolMemberRole(t *testing.T, svc *mockItemPoolService, userRepo repository.UserRepo) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "member", RoleID: ptrUint(2)}}
	deps.ItemPool = NewItemPoolHandler(svc, userRepo, &mockMainItemRepoForPool{})
	return deps
}

// helper to create an ItemPool model for tests.
func testItemPool(id uint, teamID uint) *model.ItemPool {
	return &model.ItemPool{
		TeamID:     teamID,
		Title:      "Test Pool Item",
		Status:     "pending",
		SubmitterID: 5,
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/item-pool (Submit)
// ---------------------------------------------------------------------------

func TestSubmitItemPool_Success(t *testing.T) {
	svc := &mockItemPoolService{}
	item := testItemPool(1, 10)
	item.ID = 1
	svc.submitResult.item = item

	userRepo := &mockUserRepoForHandler{user: &model.User{DisplayName: "Test User"}}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Test Pool Item","background":"some bg","expectedOutput":"some output"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool", strings.NewReader(body))
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
	assert.Equal(t, "Test Pool Item", data["title"])
	assert.True(t, svc.submitCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
	assert.Equal(t, uint(5), svc.lastSubmitterID)
}

func TestSubmitItemPool_MemberCanSubmit(t *testing.T) {
	// Any team member can submit, not just PM
	svc := &mockItemPoolService{}
	item := testItemPool(1, 10)
	item.ID = 1
	svc.submitResult.item = item

	userRepo := &mockUserRepoForHandler{user: &model.User{DisplayName: "Member"}}

	deps := depsWithItemPoolMemberRole(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Member Pool Item"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, svc.submitCalled)
}

func TestSubmitItemPool_MissingTitle(t *testing.T) {
	svc := &mockItemPoolService{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"background":"no title"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.submitCalled)
}

func TestSubmitItemPool_ServiceError(t *testing.T) {
	svc := &mockItemPoolService{}
	svc.submitResult.err = errors.New("unexpected")

	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Test"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/item-pool (List)
// ---------------------------------------------------------------------------

func TestListItemPool_Success(t *testing.T) {
	svc := &mockItemPoolService{}
	svc.listResult.page = &dto.PageResult[model.ItemPool]{
		Items: []model.ItemPool{*testItemPool(1, 10), *testItemPool(2, 10)},
		Total: 2,
		Page:  1,
		Size:  20,
	}

	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/item-pool?page=1&pageSize=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(2), data["total"])
	assert.True(t, svc.listCalled)
}

func TestListItemPool_WithStatusFilter(t *testing.T) {
	svc := &mockItemPoolService{}
	svc.listResult.page = &dto.PageResult[model.ItemPool]{}

	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/item-pool?status=pending&page=1&pageSize=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.listCalled)
	assert.Equal(t, "pending", svc.lastFilter.Status)
	assert.Equal(t, 10, svc.lastPage.PageSize)
}

func TestListItemPool_ServiceError(t *testing.T) {
	svc := &mockItemPoolService{}
	svc.listResult.err = errors.New("db error")

	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/item-pool", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/item-pool/:poolId (Get)
// ---------------------------------------------------------------------------

func TestGetItemPool_Success(t *testing.T) {
	svc := &mockItemPoolService{}
	item := testItemPool(5, 10)
	item.ID = 5
	item.Background = "some background"
	item.ExpectedOutput = "some output"
	svc.getResult.item = item

	userRepo := &mockUserRepoForHandler{user: &model.User{DisplayName: "Submitter"}}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/item-pool/5", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test Pool Item", data["title"])
	assert.True(t, svc.getCalled)
	assert.Equal(t, uint(5), svc.lastPoolID)
}

func TestGetItemPool_InvalidPoolID(t *testing.T) {
	svc := &mockItemPoolService{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/item-pool/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.getCalled)
}

func TestGetItemPool_NotFound(t *testing.T) {
	svc := &mockItemPoolService{}
	svc.getResult.err = apperrors.ErrItemNotFound

	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/item-pool/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ITEM_NOT_FOUND", resp["code"])
}

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/item-pool/:poolId/assign (Assign)
// ---------------------------------------------------------------------------

func TestAssignItemPool_Success(t *testing.T) {
	svc := &mockItemPoolService{}

	// Assign handler calls Get after Assign to retrieve the updated item
	assignedSubID := uint(10)
	assignedMainID := uint(1)
	assigneeID := uint(3)
	updatedItem := &model.ItemPool{
		TeamID:         10,
		Title:          "Pool item",
		Status:         "assigned",
		AssignedMainID: &assignedMainID,
		AssignedSubID:  &assignedSubID,
		AssigneeID:     &assigneeID,
		SubmitterID:    5,
	}
	updatedItem.ID = 5
	svc.getResult.item = updatedItem

	userRepo := &mockUserRepoForHandler{user: &model.User{DisplayName: "Submitter"}}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"mainItemId":1,"assigneeId":3,"startDate":"2026-01-01","expectedEndDate":"2026-02-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/5/assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(10), data["subItemId"])

	assert.True(t, svc.assignCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
	assert.Equal(t, uint(1), svc.lastAssignReq.MainItemID)
	assert.Equal(t, ptrUint(3), svc.lastAssignReq.AssigneeID)
}

func TestAssignItemPool_RequiresPM(t *testing.T) {
	svc := &mockItemPoolService{}

	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolMemberRole(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"mainItemId":1,"assigneeId":3,"priority":"P2","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/5/assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.assignCalled)
}

func TestAssignItemPool_InvalidPoolID(t *testing.T) {
	svc := &mockItemPoolService{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"mainItemId":1,"assigneeId":3,"priority":"P2","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/abc/assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.assignCalled)
}

func TestAssignItemPool_MissingRequiredFields(t *testing.T) {
	svc := &mockItemPoolService{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/5/assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.assignCalled)
}

func TestAssignItemPool_AlreadyProcessed(t *testing.T) {
	svc := &mockItemPoolService{}
	svc.assignResult.err = apperrors.ErrItemAlreadyProcessed

	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"mainItemId":1,"assigneeId":3,"startDate":"2026-01-01","expectedEndDate":"2026-02-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/5/assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ITEM_ALREADY_PROCESSED", resp["code"])
}

func TestAssignItemPool_SuperAdminBypass(t *testing.T) {
	svc := &mockItemPoolService{}

	// Assign handler calls Get after Assign
	assignedSubID := uint(10)
	assignedMainID := uint(1)
	assigneeID := uint(3)
	updatedItem := &model.ItemPool{
		TeamID:         10,
		Status:         "assigned",
		AssignedMainID: &assignedMainID,
		AssignedSubID:  &assignedSubID,
		AssigneeID:     &assigneeID,
		SubmitterID:    5,
	}
	updatedItem.ID = 5
	svc.getResult.item = updatedItem

	userRepo := &mockUserRepoForHandler{user: &model.User{DisplayName: "Admin"}}

	// Use member-role team, but superadmin token
	deps := depsWithItemPoolMemberRole(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "admin")
	body := `{"mainItemId":1,"assigneeId":3,"priority":"P2","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/5/assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.assignCalled)
}

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/item-pool/:poolId/reject (Reject)
// ---------------------------------------------------------------------------

func TestRejectItemPool_Success(t *testing.T) {
	svc := &mockItemPoolService{}

	// Reject handler calls Get after Reject to retrieve the updated item
	rejectedItem := &model.ItemPool{
		TeamID:       10,
		Title:        "Pool item",
		Status:       "rejected",
		RejectReason: "Not enough priority",
		SubmitterID:  5,
	}
	rejectedItem.ID = 5
	svc.getResult.item = rejectedItem

	userRepo := &mockUserRepoForHandler{user: &model.User{DisplayName: "Submitter"}}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"reason":"Not enough priority"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/5/reject", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.rejectCalled)
	assert.Equal(t, "Not enough priority", svc.lastReason)
}

func TestRejectItemPool_RequiresPM(t *testing.T) {
	svc := &mockItemPoolService{}

	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolMemberRole(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"reason":"some reason"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/5/reject", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.rejectCalled)
}

func TestRejectItemPool_InvalidPoolID(t *testing.T) {
	svc := &mockItemPoolService{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"reason":"some reason"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/abc/reject", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.rejectCalled)
}

func TestRejectItemPool_ReasonRequired(t *testing.T) {
	svc := &mockItemPoolService{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/5/reject", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.rejectCalled)
}

func TestRejectItemPool_AlreadyProcessed(t *testing.T) {
	svc := &mockItemPoolService{}
	svc.rejectResult.err = apperrors.ErrItemAlreadyProcessed

	userRepo := &mockUserRepoForHandler{}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"reason":"some reason"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/5/reject", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ITEM_ALREADY_PROCESSED", resp["code"])
}

// ---------------------------------------------------------------------------
// Tests: Response shape matches Data Contract
// ---------------------------------------------------------------------------

func TestGetItemPool_ResponseShapeMatchesDataContract(t *testing.T) {
	svc := &mockItemPoolService{}
	now := time.Now()
	assignedMainID := uint(1)
	assignedSubID := uint(10)
	assigneeID := uint(3)
	reviewerID := uint(2)

	item := &model.ItemPool{
		TeamID:         10,
		Title:          "优化首页加载速度",
		Background:     "用户反馈首页加载超过 3 秒",
		ExpectedOutput: "首页 LCP < 1.5 秒",
		SubmitterID:    5,
		Status:         "assigned",
		AssignedMainID: &assignedMainID,
		AssignedSubID:  &assignedSubID,
		AssigneeID:     &assigneeID,
		RejectReason:   "",
		ReviewedAt:     &now,
		ReviewerID:     &reviewerID,
	}
	item.ID = 50

	svc.getResult.item = item

	userRepo := &mockUserRepoForHandler{user: &model.User{DisplayName: "王五"}}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/item-pool/50", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})

	// Verify all fields from Data Contract are present
	assert.Equal(t, float64(50), data["id"])
	assert.Equal(t, "优化首页加载速度", data["title"])
	assert.Equal(t, "用户反馈首页加载超过 3 秒", data["background"])
	assert.Equal(t, "首页 LCP < 1.5 秒", data["expectedOutput"])
	assert.Equal(t, float64(5), data["submitterId"])
	assert.Equal(t, "王五", data["submitterName"])
	assert.Equal(t, "assigned", data["status"])
	assert.Equal(t, float64(1), data["assignedMainId"])
	assert.Equal(t, float64(10), data["assignedSubId"])
	assert.Equal(t, float64(3), data["assigneeId"])
}

func TestSubmitItemPool_ResponseShapeMatchesDataContract(t *testing.T) {
	svc := &mockItemPoolService{}
	item := &model.ItemPool{
		TeamID:         10,
		Title:          "优化首页加载速度",
		Background:     "用户反馈首页加载超过 3 秒",
		ExpectedOutput: "首页 LCP < 1.5 秒",
		SubmitterID:    5,
		Status:         "pending",
	}
	item.ID = 50

	svc.submitResult.item = item

	userRepo := &mockUserRepoForHandler{user: &model.User{DisplayName: "王五"}}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"优化首页加载速度","background":"用户反馈首页加载超过 3 秒","expectedOutput":"首页 LCP < 1.5 秒"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "pending", data["status"])
	assert.Equal(t, "王五", data["submitterName"])
	assert.Equal(t, float64(5), data["submitterId"])
}

// ---------------------------------------------------------------------------
// Tests: Assign returns subItemId
// ---------------------------------------------------------------------------

func TestAssignItemPool_ReturnsSubItemId(t *testing.T) {
	svc := &mockItemPoolService{}

	// After assign, Get should return the updated item with assignedSubId
	assignedSubID := uint(123)
	assignedMainID := uint(1)
	assigneeID := uint(3)

	// The handler will call Get after Assign to retrieve the updated item
	updatedItem := &model.ItemPool{
		TeamID:         10,
		Title:          "Pool item",
		Status:         "assigned",
		AssignedMainID: &assignedMainID,
		AssignedSubID:  &assignedSubID,
		AssigneeID:     &assigneeID,
		SubmitterID:    5,
	}
	updatedItem.ID = 5

	// Set up: Assign succeeds, then Get returns the updated item
	svc.getResult.item = updatedItem

	userRepo := &mockUserRepoForHandler{user: &model.User{DisplayName: "Submitter"}}

	deps := depsWithItemPoolSvc(t, svc, userRepo)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := fmt.Sprintf(`{"mainItemId":%d,"assigneeId":%d,"priority":"P2","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`, assignedMainID, assigneeID)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/item-pool/5/assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(123), data["subItemId"])
}

// mockMainItemRepoForPool is a minimal mock for repository.MainItemRepo used by item pool tests.
type mockMainItemRepoForPool struct{}

func (m *mockMainItemRepoForPool) Create(_ context.Context, _ *model.MainItem) error { return nil }
func (m *mockMainItemRepoForPool) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForPool) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (m *mockMainItemRepoForPool) List(_ context.Context, _ uint, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemRepoForPool) NextCode(_ context.Context, _ uint) (string, error) {
	return "", nil
}
func (m *mockMainItemRepoForPool) CountByTeam(_ context.Context, _ uint) (int64, error) {
	return 0, nil
}
func (m *mockMainItemRepoForPool) ListNonArchivedByTeam(_ context.Context, _ uint) ([]model.MainItem, error) {
	return nil, nil
}

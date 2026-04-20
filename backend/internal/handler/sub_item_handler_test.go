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
)

// ---------------------------------------------------------------------------
// Mock SubItemService for handler tests
// ---------------------------------------------------------------------------

type mockSubItemService struct {
	createResult struct {
		item *model.SubItem
		err  error
	}
	getResult struct {
		item *model.SubItem
		err  error
	}
	listResult struct {
		page *dto.PageResult[model.SubItem]
		err  error
	}
	updateResult struct {
		err error
	}
	changeStatusResult struct {
		err error
	}
	assignResult struct {
		err error
	}

	// capture calls
	createCalled      bool
	lastTeamID        uint
	lastCallerID      uint
	lastCreateReq     dto.SubItemCreateReq

	getCalled   bool
	lastItemID  uint

	listCalled    bool
	lastMainID    *uint
	lastFilter    dto.SubItemFilter
	lastPage      dto.Pagination

	updateCalled  bool
	lastUpdateID  uint
	lastUpdateReq dto.SubItemUpdateReq

	changeStatusCalled bool
	lastNewStatus      string

	assignCalled  bool
	lastPmID      uint
	lastAssigneeID uint
}

func (m *mockSubItemService) Create(_ context.Context, teamID, callerID uint, req dto.SubItemCreateReq) (*model.SubItem, error) {
	m.createCalled = true
	m.lastTeamID = teamID
	m.lastCallerID = callerID
	m.lastCreateReq = req
	return m.createResult.item, m.createResult.err
}

func (m *mockSubItemService) Update(_ context.Context, teamID, itemID uint, req dto.SubItemUpdateReq) error {
	m.updateCalled = true
	m.lastTeamID = teamID
	m.lastUpdateID = itemID
	m.lastUpdateReq = req
	return m.updateResult.err
}

func (m *mockSubItemService) ChangeStatus(_ context.Context, teamID, callerID, itemID uint, newStatus string) error {
	m.changeStatusCalled = true
	m.lastTeamID = teamID
	m.lastCallerID = callerID
	m.lastItemID = itemID
	m.lastNewStatus = newStatus
	return m.changeStatusResult.err
}

func (m *mockSubItemService) Get(_ context.Context, teamID, itemID uint) (*model.SubItem, error) {
	m.getCalled = true
	m.lastTeamID = teamID
	m.lastItemID = itemID
	return m.getResult.item, m.getResult.err
}

func (m *mockSubItemService) List(_ context.Context, teamID uint, mainItemID *uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	m.listCalled = true
	m.lastTeamID = teamID
	m.lastMainID = mainItemID
	m.lastFilter = filter
	m.lastPage = page
	return m.listResult.page, m.listResult.err
}

func (m *mockSubItemService) Assign(_ context.Context, teamID, pmID, itemID, assigneeID uint) error {
	m.assignCalled = true
	m.lastTeamID = teamID
	m.lastPmID = pmID
	m.lastItemID = itemID
	m.lastAssigneeID = assigneeID
	return m.assignResult.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// depsWithSubItemSvc wires a mock SubItemService with PM role.

func depsWithSubItemSvc(t *testing.T, svc *mockSubItemService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm", RoleID: ptrUint(1)}}
	deps.SubItem = NewSubItemHandler(svc)
	return deps
}

// depsWithSubItemSvcMemberRole wires with member (non-PM) role.

func depsWithSubItemSvcMemberRole(t *testing.T, svc *mockSubItemService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "member", RoleID: ptrUint(2)}}
	deps.SubItem = NewSubItemHandler(svc)
	return deps
}

// testSubItem creates a SubItem model for tests.
func testSubItem(id uint, teamID uint) *model.SubItem {
	return &model.SubItem{
		TeamID:     teamID,
		MainItemID: 1,
		Title:      "Test SubItem",
		Priority:   "P2",
		Status:     "待开始",
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/main-items/:itemId/sub-items (Create)
// ---------------------------------------------------------------------------

func TestCreateSubItem_Success(t *testing.T) {
	svc := &mockSubItemService{}
	item := testSubItem(1, 10)
	item.ID = 1
	svc.createResult.item = item

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"mainItemId":1,"title":"Test SubItem","priority":"P2","assigneeId":1,"startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/1/sub-items", strings.NewReader(body))
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
	assert.Equal(t, "Test SubItem", data["title"])
	assert.True(t, svc.createCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
	assert.Equal(t, uint(5), svc.lastCallerID)
}

func TestCreateSubItem_MemberCanCreate(t *testing.T) {
	// Any team member can create a sub-item (not PM-only)
	svc := &mockSubItemService{}
	item := testSubItem(1, 10)
	item.ID = 1
	svc.createResult.item = item

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"mainItemId":1,"title":"Test SubItem","priority":"P2","assigneeId":1,"startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/1/sub-items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, svc.createCalled)
}

func TestCreateSubItem_InvalidBody(t *testing.T) {
	svc := &mockSubItemService{}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/1/sub-items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.createCalled)
}

func TestCreateSubItem_ServiceError(t *testing.T) {
	svc := &mockSubItemService{}
	svc.createResult.err = errors.New("unexpected")

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"mainItemId":1,"title":"Test","priority":"P2","assigneeId":1,"startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/1/sub-items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/main-items/:itemId/sub-items (List)
// ---------------------------------------------------------------------------

func TestListSubItems_Success(t *testing.T) {
	svc := &mockSubItemService{}
	svc.listResult.page = &dto.PageResult[model.SubItem]{
		Items: []model.SubItem{*testSubItem(1, 10), *testSubItem(2, 10)},
		Total: 2,
		Page:  1,
		Size:  20,
	}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/1/sub-items?page=1&pageSize=20", nil)
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

func TestListSubItems_WithFilters(t *testing.T) {
	svc := &mockSubItemService{}
	svc.listResult.page = &dto.PageResult[model.SubItem]{}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/1/sub-items?priority=P2&status=进行中&page=2&pageSize=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.listCalled)
	assert.Equal(t, "P2", svc.lastFilter.Priority)
	assert.Equal(t, "进行中", svc.lastFilter.Status)
	assert.Equal(t, 2, svc.lastPage.Page)
	assert.Equal(t, 10, svc.lastPage.PageSize)
}

func TestListSubItems_ServiceError(t *testing.T) {
	svc := &mockSubItemService{}
	svc.listResult.err = errors.New("db error")

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/1/sub-items", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/sub-items/:subId (Get)
// ---------------------------------------------------------------------------

func TestGetSubItem_Success(t *testing.T) {
	svc := &mockSubItemService{}
	item := testSubItem(1, 10)
	item.ID = 1
	item.Status = "进行中"
	item.Completion = 60
	svc.getResult.item = item

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test SubItem", data["title"])
	assert.True(t, svc.getCalled)
	assert.Equal(t, uint(1), svc.lastItemID)
}

func TestGetSubItem_InvalidID(t *testing.T) {
	svc := &mockSubItemService{}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.getCalled)
}

func TestGetSubItem_NotFound(t *testing.T) {
	svc := &mockSubItemService{}
	svc.getResult.err = apperrors.ErrItemNotFound

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ITEM_NOT_FOUND", resp["code"])
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/sub-items/:subId (Update)
// ---------------------------------------------------------------------------

func TestUpdateSubItem_Success_AsPM(t *testing.T) {
	svc := &mockSubItemService{}
	item := testSubItem(1, 10)
	item.ID = 1
	item.Title = "Updated Title"
	svc.getResult.item = item

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Updated Title","priority":"P1"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.updateCalled)
	assert.Equal(t, uint(1), svc.lastUpdateID)
	require.NotNil(t, svc.lastUpdateReq.Title)
	assert.Equal(t, "Updated Title", *svc.lastUpdateReq.Title)
}

func TestUpdateSubItem_Success_AsAssignee(t *testing.T) {
	// Member who is the assignee should be able to update
	svc := &mockSubItemService{}
	assigneeID := uint(5)
	item := testSubItem(1, 10)
	item.ID = 1
	item.AssigneeID = &assigneeID
	item.Title = "Updated"
	svc.getResult.item = item

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser") // userID=5, same as assigneeID
	body := `{"title":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.updateCalled)
}

func TestUpdateSubItem_Forbidden_NonPMNonAssignee(t *testing.T) {
	// Member who is NOT the assignee and NOT PM should be forbidden
	svc := &mockSubItemService{}
	assigneeID := uint(99) // different user
	item := testSubItem(1, 10)
	item.ID = 1
	item.AssigneeID = &assigneeID
	svc.getResult.item = item

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser") // userID=5, NOT the assignee (99)
	body := `{"title":"Hacked"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.updateCalled)
}

func TestUpdateSubItem_InvalidID(t *testing.T) {
	svc := &mockSubItemService{}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/abc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.updateCalled)
}

func TestUpdateSubItem_NotFound(t *testing.T) {
	svc := &mockSubItemService{}
	svc.updateResult.err = apperrors.ErrItemNotFound

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/sub-items/:subId/status (ChangeStatus)
// ---------------------------------------------------------------------------

func TestChangeStatus_Success(t *testing.T) {
	svc := &mockSubItemService{}
	item := testSubItem(1, 10)
	item.ID = 1
	item.Status = "进行中"
	svc.getResult.item = item

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"进行中"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.changeStatusCalled)
	assert.Equal(t, "进行中", svc.lastNewStatus)
}

func TestChangeStatus_InvalidStatus_422(t *testing.T) {
	svc := &mockSubItemService{}
	svc.changeStatusResult.err = apperrors.ErrInvalidStatus

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"已完成"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_STATUS", resp["code"])
}

func TestChangeStatus_InvalidID(t *testing.T) {
	svc := &mockSubItemService{}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"进行中"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/abc/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.changeStatusCalled)
}

func TestChangeStatus_InvalidBody(t *testing.T) {
	svc := &mockSubItemService{}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.changeStatusCalled)
}

func TestChangeStatus_AsAssignee(t *testing.T) {
	// Assignee should be able to change status
	svc := &mockSubItemService{}
	assigneeID := uint(5)
	item := testSubItem(1, 10)
	item.ID = 1
	item.AssigneeID = &assigneeID
	item.Status = "进行中"
	svc.getResult.item = item

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"进行中"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.changeStatusCalled)
}

func TestChangeStatus_Forbidden_NonPMNonAssignee(t *testing.T) {
	svc := &mockSubItemService{}
	assigneeID := uint(99)
	item := testSubItem(1, 10)
	item.ID = 1
	item.AssigneeID = &assigneeID
	svc.getResult.item = item

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"进行中"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.changeStatusCalled)
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/sub-items/:subId/assignee (Assign)
// ---------------------------------------------------------------------------

func TestAssignSubItem_Success(t *testing.T) {
	svc := &mockSubItemService{}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"assigneeId":3}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/assignee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.assignCalled)
	assert.Equal(t, uint(5), svc.lastPmID)
	assert.Equal(t, uint(1), svc.lastItemID)
	assert.Equal(t, uint(3), svc.lastAssigneeID)
}

func TestAssignSubItem_RequiresPM(t *testing.T) {
	svc := &mockSubItemService{}

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"assigneeId":3}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/assignee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.assignCalled)
}

func TestAssignSubItem_InvalidID(t *testing.T) {
	svc := &mockSubItemService{}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"assigneeId":3}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/abc/assignee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.assignCalled)
}

func TestAssignSubItem_InvalidBody(t *testing.T) {
	svc := &mockSubItemService{}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/assignee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.assignCalled)
}

func TestAssignSubItem_ItemNotFound(t *testing.T) {
	svc := &mockSubItemService{}
	svc.assignResult.err = apperrors.ErrItemNotFound

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := `{"assigneeId":3}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/999/assignee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAssignSubItem_SuperAdminBypass(t *testing.T) {
	svc := &mockSubItemService{}

	// Use member-role team, but superadmin token
	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "admin")
	body := `{"assigneeId":3}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/assignee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.assignCalled)
}

// ---------------------------------------------------------------------------
// Tests: Response shape matches Data Contract
// ---------------------------------------------------------------------------

func TestGetSubItem_ResponseShapeMatchesDataContract(t *testing.T) {
	svc := &mockSubItemService{}
	now := time.Now()
	assigneeID := uint(3)
	item := &model.SubItem{
		TeamID:          10,
		MainItemID:      1,
		Title:           "实现支付接口",
		Description:     "对接微信支付",
		Priority:        "P2",
		AssigneeID:      &assigneeID,
		StartDate:       &now,
		ExpectedEndDate: &now,
		Status:          "进行中",
		Completion:      60.0,
		IsKeyItem:       false,
		DelayCount:      0,
		Weight:          1.0,
	}
	item.ID = 1

	svc.getResult.item = item

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})

	// Verify all expected fields
	assert.Equal(t, "实现支付接口", data["title"])
	assert.Equal(t, "P2", data["priority"])
	assert.Equal(t, float64(3), data["assigneeId"])
	assert.Equal(t, "进行中", data["status"])
	assert.Equal(t, 60.0, data["completion"])
	assert.Equal(t, false, data["isKeyItem"])
	assert.Equal(t, float64(0), data["delayCount"])
	assert.Equal(t, float64(1), data["mainItemId"])
}

func TestCreateSubItem_ResponseShapeMatchesDataContract(t *testing.T) {
	svc := &mockSubItemService{}
	assigneeID := uint(3)
	item := &model.SubItem{
		TeamID:     10,
		MainItemID: 1,
		Title:      "New SubItem",
		Priority:   "P2",
		AssigneeID: &assigneeID,
		Status:     "待开始",
		Weight:     1.0,
	}
	item.ID = 5
	svc.createResult.item = item

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps)

	token := signTestToken(t, 5, "testuser")
	body := fmt.Sprintf(`{"mainItemId":1,"title":"New SubItem","priority":"P2","assigneeId":%d,"startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`, assigneeID)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/1/sub-items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(5), data["id"])
	assert.Equal(t, "New SubItem", data["title"])
	assert.Equal(t, "P2", data["priority"])
	assert.Equal(t, "待开始", data["status"])
}

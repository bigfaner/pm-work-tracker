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
	"pm-work-tracker/backend/internal/service"
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
		result *service.SubItemChangeResult
		err    error
	}
	assignResult struct {
		err error
	}
	availableTransitionsResult struct {
		transitions []string
		err         error
	}

	// capture calls
	createCalled  bool
	lastTeamID    uint
	lastCallerID  uint
	lastCreateReq dto.SubItemCreateReq

	getCalled  bool
	lastItemID uint

	listCalled bool
	lastMainID *int64
	lastFilter dto.SubItemFilter
	lastPage   dto.Pagination

	updateCalled  bool
	lastUpdateID  uint
	lastUpdateReq dto.SubItemUpdateReq

	changeStatusCalled bool
	lastNewStatus      string

	assignCalled   bool
	lastPmID       uint
	lastAssigneeID uint

	availableTransitionsCalled bool
}

func (m *mockSubItemService) Create(_ context.Context, teamBizKey int64, callerBizKey int64, req dto.SubItemCreateReq) (*model.SubItem, error) {
	m.createCalled = true
	m.lastTeamID = uint(teamBizKey)
	m.lastCallerID = uint(callerBizKey)
	m.lastCreateReq = req
	return m.createResult.item, m.createResult.err
}

func (m *mockSubItemService) Update(_ context.Context, teamBizKey int64, itemID uint, req dto.SubItemUpdateReq) error {
	m.updateCalled = true
	m.lastTeamID = uint(teamBizKey)
	m.lastUpdateID = itemID
	m.lastUpdateReq = req
	return m.updateResult.err
}

func (m *mockSubItemService) ChangeStatus(_ context.Context, teamBizKey int64, callerBizKey int64, itemID uint, newStatus string) (*service.SubItemChangeResult, error) {
	m.changeStatusCalled = true
	m.lastTeamID = uint(teamBizKey)
	m.lastCallerID = uint(callerBizKey)
	m.lastItemID = itemID
	m.lastNewStatus = newStatus
	return m.changeStatusResult.result, m.changeStatusResult.err
}

func (m *mockSubItemService) Get(_ context.Context, teamBizKey int64, itemID uint) (*model.SubItem, error) {
	m.getCalled = true
	m.lastTeamID = uint(teamBizKey)
	m.lastItemID = itemID
	return m.getResult.item, m.getResult.err
}

func (m *mockSubItemService) List(_ context.Context, teamBizKey int64, mainItemID *int64, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	m.listCalled = true
	m.lastTeamID = uint(teamBizKey)
	m.lastMainID = mainItemID
	m.lastFilter = filter
	m.lastPage = page
	return m.listResult.page, m.listResult.err
}

func (m *mockSubItemService) Assign(_ context.Context, teamBizKey int64, pmBizKey int64, itemID uint, assigneeBizKey int64) error {
	m.assignCalled = true
	m.lastTeamID = uint(teamBizKey)
	m.lastPmID = uint(pmBizKey)
	m.lastItemID = itemID
	m.lastAssigneeID = uint(assigneeBizKey)
	return m.assignResult.err
}

func (m *mockSubItemService) AvailableTransitions(_ context.Context, teamBizKey int64, subID uint) ([]string, error) {
	m.availableTransitionsCalled = true
	m.lastTeamID = uint(teamBizKey)
	m.lastItemID = subID
	return m.availableTransitionsResult.transitions, m.availableTransitionsResult.err
}
func (m *mockSubItemService) Delete(_ context.Context, _ int64, _ int64, _ uint) error {
	return nil
}
func (m *mockSubItemService) GetByBizKey(_ context.Context, bizKey int64) (*model.SubItem, error) {
	return &model.SubItem{BaseModel: model.BaseModel{ID: uint(bizKey)}}, nil
}

// mockMainItemSvcForSubItem resolves bizKey for SubItemHandler List tests.
type mockMainItemSvcForSubItem struct{}

func (m *mockMainItemSvcForSubItem) Create(_ context.Context, _ int64, _ int64, _ dto.MainItemCreateReq) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemSvcForSubItem) Update(_ context.Context, _ int64, _ uint, _ dto.MainItemUpdateReq) error {
	return nil
}
func (m *mockMainItemSvcForSubItem) Archive(_ context.Context, _ int64, _ uint) error { return nil }
func (m *mockMainItemSvcForSubItem) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemSvcForSubItem) Get(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemSvcForSubItem) GetByBizKey(_ context.Context, bizKey int64) (*model.MainItem, error) {
	return &model.MainItem{BaseModel: model.BaseModel{ID: uint(bizKey)}}, nil
}
func (m *mockMainItemSvcForSubItem) RecalcCompletion(_ context.Context, _ int64) error { return nil }
func (m *mockMainItemSvcForSubItem) ChangeStatus(_ context.Context, _ int64, _ int64, _ uint, _ string) (*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemSvcForSubItem) AvailableTransitions(_ context.Context, _ int64, _ int64, _ uint) ([]string, error) {
	return nil, nil
}
func (m *mockMainItemSvcForSubItem) EvaluateLinkage(_ context.Context, _ int64, _ int64) (*service.LinkageResult, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// depsWithSubItemSvc wires a mock SubItemService with PM role.

func depsWithSubItemSvc(t *testing.T, svc *mockSubItemService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{RoleKey: func() *int64 { v := int64(1); return &v }()}}
	deps.SubItem = NewSubItemHandler(svc, &mockMainItemSvcForSubItem{})
	return deps
}

// depsWithSubItemSvcMemberRole wires with member (non-PM) role.

func depsWithSubItemSvcMemberRole(t *testing.T, svc *mockSubItemService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{RoleKey: func() *int64 { v := int64(2); return &v }()}}
	deps.SubItem = NewSubItemHandler(svc, &mockMainItemSvcForSubItem{})
	return deps
}

// testSubItem creates a SubItem model for tests.
func testSubItem(id uint, teamID uint) *model.SubItem {
	return &model.SubItem{
		TeamKey:     int64(teamID),
		MainItemKey: int64(1),
		Title:       "Test SubItem",
		Priority:    "P2",
		ItemStatus:  "pending",
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
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"mainItemKey":"1","title":"Test SubItem","priority":"P2","assigneeKey":"1","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
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
	svc := &mockSubItemService{}
	item := testSubItem(1, 10)
	item.ID = 1
	svc.createResult.item = item

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"mainItemKey":"1","title":"Test SubItem","priority":"P2","assigneeKey":"1","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
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
	r := SetupRouter(deps, nil)

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
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"mainItemKey":"1","title":"Test","priority":"P2","assigneeKey":"1","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
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
	r := SetupRouter(deps, nil)

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
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/1/sub-items?priority=P2&status=progressing&page=2&pageSize=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.listCalled)
	assert.Equal(t, "P2", svc.lastFilter.Priority)
	assert.Equal(t, "progressing", svc.lastFilter.Status)
	assert.Equal(t, 2, svc.lastPage.Page)
	assert.Equal(t, 10, svc.lastPage.PageSize)
}

func TestListSubItems_ServiceError(t *testing.T) {
	svc := &mockSubItemService{}
	svc.listResult.err = errors.New("db error")

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps, nil)

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
	item.BizKey = 1
	item.ItemStatus = "progressing"
	item.Completion = 60
	svc.getResult.item = item

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps, nil)

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
	r := SetupRouter(deps, nil)

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
	r := SetupRouter(deps, nil)

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
	r := SetupRouter(deps, nil)

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
	svc := &mockSubItemService{}
	assigneeID := uint(5)
	item := testSubItem(1, 10)
	item.ID = 1
	item.AssigneeKey = func() *int64 { v := int64(assigneeID); return &v }()
	item.Title = "Updated"
	svc.getResult.item = item

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps, nil)

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
	svc := &mockSubItemService{}
	assigneeID := uint(99) // different user
	item := testSubItem(1, 10)
	item.ID = 1
	item.AssigneeKey = func() *int64 { v := int64(assigneeID); return &v }()
	svc.getResult.item = item

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps, nil)

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
	r := SetupRouter(deps, nil)

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
	r := SetupRouter(deps, nil)

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
	item.BizKey = 1
	item.ItemStatus = "progressing"
	svc.changeStatusResult.result = &service.SubItemChangeResult{SubItem: item}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"progressing"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.changeStatusCalled)
	assert.Equal(t, "progressing", svc.lastNewStatus)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	subItem, ok := data["subItem"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "progressing", subItem["itemStatus"])
}

func TestChangeStatus_InvalidStatus_422(t *testing.T) {
	svc := &mockSubItemService{}
	svc.changeStatusResult.err = apperrors.ErrInvalidStatus

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"completed"}`
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
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"progressing"}`
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
	r := SetupRouter(deps, nil)

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
	svc := &mockSubItemService{}
	assigneeID := uint(5)
	item := testSubItem(1, 10)
	item.ID = 1
	item.AssigneeKey = func() *int64 { v := int64(assigneeID); return &v }()
	item.ItemStatus = "progressing"
	svc.getResult.item = item // needed for assignee check
	svc.changeStatusResult.result = &service.SubItemChangeResult{SubItem: item}

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"progressing"}`
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
	item.AssigneeKey = func() *int64 { v := int64(assigneeID); return &v }()
	svc.getResult.item = item

	deps := depsWithSubItemSvcMemberRole(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"progressing"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/sub-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.changeStatusCalled)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/sub-items/:subId/available-transitions
// ---------------------------------------------------------------------------

func TestAvailableTransitions_Success(t *testing.T) {
	svc := &mockSubItemService{}
	svc.availableTransitionsResult.transitions = []string{"progressing", "closed"}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/1/available-transitions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.availableTransitionsCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
	assert.Equal(t, uint(1), svc.lastItemID)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	transitions, ok := data["transitions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, transitions, 2)
}

func TestAvailableTransitions_NotFound(t *testing.T) {
	svc := &mockSubItemService{}
	svc.availableTransitionsResult.err = apperrors.ErrItemNotFound

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/sub-items/999/available-transitions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/sub-items/:subId/assignee (Assign)
// ---------------------------------------------------------------------------

func TestAssignSubItem_Success(t *testing.T) {
	svc := &mockSubItemService{}

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"assigneeKey":"3"}`
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
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"assigneeKey":"3"}`
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
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"assigneeKey":"3"}`
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
	r := SetupRouter(deps, nil)

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
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"assigneeKey":"3"}`
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
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 1, "admin")
	body := `{"assigneeKey":"3"}`
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
		TeamKey:         10,
		MainItemKey:     int64(1),
		Title:           "实现支付接口",
		ItemDesc:        "对接微信支付",
		Priority:        "P2",
		AssigneeKey:     func() *int64 { v := int64(assigneeID); return &v }(),
		PlanStartDate:   &now,
		ExpectedEndDate: &now,
		ItemStatus:      "progressing",
		Completion:      60.0,
		IsKeyItem:       false,
		Weight:          1.0,
	}
	item.ID = 1
	item.BizKey = 1

	svc.getResult.item = item

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps, nil)

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
	assert.Equal(t, "3", data["assigneeKey"])
	assert.Equal(t, "progressing", data["itemStatus"])
	assert.Equal(t, 60.0, data["completion"])
	assert.Equal(t, false, data["isKeyItem"])
	assert.Equal(t, "1", data["mainItemKey"])
}

func TestCreateSubItem_ResponseShapeMatchesDataContract(t *testing.T) {
	svc := &mockSubItemService{}
	assigneeID := uint(3)
	item := &model.SubItem{
		TeamKey:     10,
		MainItemKey: int64(1),
		Title:       "New SubItem",
		Priority:    "P2",
		AssigneeKey: func() *int64 { v := int64(assigneeID); return &v }(),
		ItemStatus:  "pending",
		Weight:      1.0,
	}
	item.ID = 5
	item.BizKey = 5
	svc.createResult.item = item

	deps := depsWithSubItemSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := fmt.Sprintf(`{"mainItemKey":"1","title":"New SubItem","priority":"P2","assigneeKey":"%d","startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`, assigneeID)
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
	assert.Equal(t, "5", data["bizKey"])
	assert.Equal(t, "New SubItem", data["title"])
	assert.Equal(t, "P2", data["priority"])
	assert.Equal(t, "pending", data["itemStatus"])
}

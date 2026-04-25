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
	"pm-work-tracker/backend/internal/service"
)

// ---------------------------------------------------------------------------
// Mock MainItemService for handler tests
// ---------------------------------------------------------------------------

type mockMainItemService struct {
	createResult struct {
		item *model.MainItem
		err  error
	}
	getResult struct {
		item *model.MainItem
		err  error
	}
	listResult struct {
		page *dto.PageResult[model.MainItem]
		err  error
	}
	updateResult struct {
		err error
	}
	archiveResult struct {
		err error
	}
	changeStatusResult struct {
		item *model.MainItem
		err  error
	}
	availableTransitionsResult struct {
		transitions []string
		err         error
	}

	// capture calls
	createCalled  bool
	lastTeamID    uint
	lastPmID      uint
	lastCreateReq dto.MainItemCreateReq

	getCalled  bool
	lastItemID uint

	listCalled   bool
	lastFilter   dto.MainItemFilter
	lastPage     dto.Pagination

	updateCalled bool
	lastUpdateID uint
	lastUpdateReq dto.MainItemUpdateReq

	archiveCalled bool
	archiveItemID uint

	changeStatusCalled bool
	lastCallerID      uint
	lastNewStatus     string

	availableTransitionsCalled bool
}

func (m *mockMainItemService) Create(_ context.Context, teamID, pmID uint, req dto.MainItemCreateReq) (*model.MainItem, error) {
	m.createCalled = true
	m.lastTeamID = teamID
	m.lastPmID = pmID
	m.lastCreateReq = req
	return m.createResult.item, m.createResult.err
}

func (m *mockMainItemService) Update(_ context.Context, teamID, itemID uint, req dto.MainItemUpdateReq) error {
	m.updateCalled = true
	m.lastTeamID = teamID
	m.lastUpdateID = itemID
	m.lastUpdateReq = req
	return m.updateResult.err
}

func (m *mockMainItemService) Archive(_ context.Context, teamID, itemID uint) error {
	m.archiveCalled = true
	m.lastTeamID = teamID
	m.archiveItemID = itemID
	return m.archiveResult.err
}

func (m *mockMainItemService) List(_ context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	m.listCalled = true
	m.lastTeamID = teamID
	m.lastFilter = filter
	m.lastPage = page
	return m.listResult.page, m.listResult.err
}

func (m *mockMainItemService) Get(_ context.Context, itemID uint) (*model.MainItem, error) {
	m.getCalled = true
	m.lastItemID = itemID
	return m.getResult.item, m.getResult.err
}

func (m *mockMainItemService) RecalcCompletion(_ context.Context, _ uint) error {
	return nil
}

func (m *mockMainItemService) ChangeStatus(_ context.Context, teamID, callerID, itemID uint, newStatus string) (*model.MainItem, error) {
	m.changeStatusCalled = true
	m.lastTeamID = teamID
	m.lastCallerID = callerID
	m.lastItemID = itemID
	m.lastNewStatus = newStatus
	return m.changeStatusResult.item, m.changeStatusResult.err
}

func (m *mockMainItemService) AvailableTransitions(_ context.Context, teamID, callerID, itemID uint) ([]string, error) {
	m.availableTransitionsCalled = true
	m.lastTeamID = teamID
	m.lastCallerID = callerID
	m.lastItemID = itemID
	return m.availableTransitionsResult.transitions, m.availableTransitionsResult.err
}

func (m *mockMainItemService) EvaluateLinkage(_ context.Context, _ uint, _ uint) (*service.LinkageResult, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Mock SubItemRepo for Get endpoint (subItems summary)
// ---------------------------------------------------------------------------

type mockSubItemRepoForHandler struct {
	items []*model.SubItem
	err   error
}

func (m *mockSubItemRepoForHandler) Create(_ context.Context, _ *model.SubItem) error {
	return nil
}
func (m *mockSubItemRepoForHandler) FindByID(_ context.Context, _ uint) (*model.SubItem, error) {
	return nil, nil
}
func (m *mockSubItemRepoForHandler) Update(_ context.Context, _ *model.SubItem, _ map[string]interface{}) error {
	return nil
}
func (m *mockSubItemRepoForHandler) List(_ context.Context, _ uint, _ uint, _ dto.SubItemFilter, _ dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, nil
}
func (m *mockSubItemRepoForHandler) ListByMainItem(_ context.Context, _ uint) ([]*model.SubItem, error) {
	return m.items, m.err
}
func (m *mockSubItemRepoForHandler) ListByTeam(_ context.Context, _ uint) ([]model.SubItem, error) {
	return nil, nil
}
func (m *mockSubItemRepoForHandler) SoftDelete(_ context.Context, _ uint) error {
	return nil
}
func (m *mockSubItemRepoForHandler) FindByBizKey(_ context.Context, _ int64) (*model.SubItem, error) {
	return nil, nil
}
func (m *mockSubItemRepoForHandler) NextSubCode(_ context.Context, _ uint) (string, error) {
	return "", nil
}
// Helpers
// ---------------------------------------------------------------------------

// depsWithMainItemSvc wires a mock MainItemService into test deps.

func depsWithMainItemSvc(t *testing.T, svc *mockMainItemService, userRepo repository.UserRepo, subItemRepo *mockSubItemRepoForHandler) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "pm", RoleID: ptrUint(1)}}
	deps.MainItem = NewMainItemHandler(svc, userRepo, subItemRepo)
	return deps
}

// depsWithMemberRole wires a member (non-PM) role for testing RBAC.
func depsWithMemberRoleMainItem(t *testing.T, svc *mockMainItemService, userRepo repository.UserRepo, subItemRepo *mockSubItemRepoForHandler) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{Role: "member", RoleID: ptrUint(2)}}
	deps.MainItem = NewMainItemHandler(svc, userRepo, subItemRepo)
	return deps
}

// helper to create a MainItem model for tests.
func testMainItem(id uint, teamID uint) *model.MainItem {
	return &model.MainItem{
		TeamID:   teamID,
		Code:     fmt.Sprintf("TEST-%05d", id),
		Title:    "Test Item",
		Priority: "P1",
		Status:   "pending",
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/main-items (Create)
// ---------------------------------------------------------------------------

func TestCreateMainItem_Success(t *testing.T) {
	svc := &mockMainItemService{}
	item := testMainItem(1, 10)
	item.ID = 1
	item.ProposerID = 5
	svc.createResult.item = item

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Test Item","priority":"P1","assigneeId":1,"startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items", strings.NewReader(body))
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
	assert.Equal(t, "Test Item", data["title"])
	assert.True(t, svc.createCalled)
	assert.Equal(t, uint(10), svc.lastTeamID)
	assert.Equal(t, uint(5), svc.lastPmID)
}

func TestCreateMainItem_WithOptionalFields(t *testing.T) {
	svc := &mockMainItemService{}
	item := testMainItem(1, 10)
	item.ID = 1
	svc.createResult.item = item

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Test Item","priority":"P1","assigneeId":3,"startDate":"2026-04-01","expectedEndDate":"2026-04-30","isKeyItem":true}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, svc.createCalled)
	assert.Equal(t, uint(3), svc.lastCreateReq.AssigneeID)
	assert.True(t, svc.lastCreateReq.IsKeyItem)
}

func TestCreateMainItem_RequiresPM(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMemberRoleMainItem(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Test Item","priority":"P1","assigneeId":1,"startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.createCalled)
}

func TestCreateMainItem_InvalidBody(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.createCalled)
}

func TestCreateMainItem_ServiceError(t *testing.T) {
	svc := &mockMainItemService{}
	svc.createResult.err = errors.New("unexpected")

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Test","priority":"P1","assigneeId":1,"startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/main-items (List)
// ---------------------------------------------------------------------------

func TestListMainItems_Success(t *testing.T) {
	svc := &mockMainItemService{}
	svc.listResult.page = &dto.PageResult[model.MainItem]{
		Items: []model.MainItem{*testMainItem(1, 10), *testMainItem(2, 10)},
		Total: 2,
		Page:  1,
		Size:  20,
	}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items?page=1&pageSize=20", nil)
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

func TestListMainItems_WithFilters(t *testing.T) {
	svc := &mockMainItemService{}
	svc.listResult.page = &dto.PageResult[model.MainItem]{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items?priority=P1&status=progressing&page=2&pageSize=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.listCalled)
	assert.Equal(t, "P1", svc.lastFilter.Priority)
	assert.Equal(t, "progressing", svc.lastFilter.Status)
	assert.Equal(t, 2, svc.lastPage.Page)
	assert.Equal(t, 10, svc.lastPage.PageSize)
}

func TestListMainItems_ServiceError(t *testing.T) {
	svc := &mockMainItemService{}
	svc.listResult.err = errors.New("db error")

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/main-items/:itemId (Get)
// ---------------------------------------------------------------------------

func TestGetMainItem_Success(t *testing.T) {
	svc := &mockMainItemService{}
	item := testMainItem(1, 10)
	item.ID = 1
	item.Status = "progressing"
	item.Completion = 45.5
	svc.getResult.item = item

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{
		items: []*model.SubItem{
			{Title: "Sub 1", Status: "progressing", Completion: 60},
		},
	}
	subItemRepo.items[0].ID = 10
	subItemRepo.items[0].MainItemID = 1

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test Item", data["title"])
	assert.True(t, svc.getCalled)
	assert.Equal(t, uint(1), svc.lastItemID)

	// Check subItems array is present
	subItems, ok := data["subItems"].([]interface{})
	require.True(t, ok)
	assert.Len(t, subItems, 1)
}

func TestGetMainItem_InvalidItemID(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.getCalled)
}

func TestGetMainItem_NotFound(t *testing.T) {
	svc := &mockMainItemService{}
	svc.getResult.err = apperrors.ErrItemNotFound

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ITEM_NOT_FOUND", resp["code"])
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/main-items/:itemId (Update)
// ---------------------------------------------------------------------------

func TestUpdateMainItem_Success(t *testing.T) {
	svc := &mockMainItemService{}
	updatedItem := testMainItem(1, 10)
	updatedItem.ID = 1
	updatedItem.Title = "Updated Title"
	updatedItem.Priority = "P2"
	svc.getResult.item = updatedItem

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Updated Title","priority":"P2"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.updateCalled)
	assert.Equal(t, uint(1), svc.lastUpdateID)
	require.NotNil(t, svc.lastUpdateReq.Title)
	assert.Equal(t, "Updated Title", *svc.lastUpdateReq.Title)
}

func TestUpdateMainItem_RequiresPM(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMemberRoleMainItem(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.updateCalled)
}

func TestUpdateMainItem_InvalidItemID(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/abc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.updateCalled)
}

func TestUpdateMainItem_ItemNotFound(t *testing.T) {
	svc := &mockMainItemService{}
	svc.updateResult.err = apperrors.ErrItemNotFound

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"title":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/main-items/:itemId/archive (Archive)
// ---------------------------------------------------------------------------

func TestArchiveMainItem_Success(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/1/archive", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.archiveCalled)
	assert.Equal(t, uint(1), svc.archiveItemID)
}

func TestArchiveMainItem_RequiresPM(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMemberRoleMainItem(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/1/archive", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.archiveCalled)
}

func TestArchiveMainItem_ArchiveNotAllowed(t *testing.T) {
	svc := &mockMainItemService{}
	svc.archiveResult.err = apperrors.ErrArchiveNotAllowed

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/1/archive", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ARCHIVE_NOT_ALLOWED", resp["code"])
}

func TestArchiveMainItem_InvalidItemID(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/abc/archive", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.archiveCalled)
}

func TestArchiveMainItem_ItemNotFound(t *testing.T) {
	svc := &mockMainItemService{}
	svc.archiveResult.err = apperrors.ErrItemNotFound

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/999/archive", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: SuperAdmin bypasses PM role check
// ---------------------------------------------------------------------------

func TestCreateMainItem_SuperAdminBypass(t *testing.T) {
	svc := &mockMainItemService{}
	item := testMainItem(1, 10)
	item.ID = 1
	svc.createResult.item = item

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	// Use member-role team, but superadmin token
	deps := depsWithMemberRoleMainItem(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 1, "admin")
	body := `{"title":"Test","priority":"P1","assigneeId":1,"startDate":"2024-01-01","expectedEndDate":"2024-03-01"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// SuperAdmin bypasses team role check — but handler checks callerTeamRole
	// The handler should check for pm role OR superadmin user role
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, svc.createCalled)
}

// ---------------------------------------------------------------------------
// Tests: Response shape matches Data Contract
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/main-items/:itemId/status (ChangeStatus)
// ---------------------------------------------------------------------------

func TestChangeStatusMainItem_Success(t *testing.T) {
	svc := &mockMainItemService{}
	updatedItem := testMainItem(1, 10)
	updatedItem.ID = 1
	updatedItem.Status = "progressing"
	svc.changeStatusResult.item = updatedItem

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"progressing"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.changeStatusCalled)
	assert.Equal(t, uint(1), svc.lastItemID)
	assert.Equal(t, uint(5), svc.lastCallerID)
	assert.Equal(t, "progressing", svc.lastNewStatus)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

func TestChangeStatusMainItem_InvalidBody(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.changeStatusCalled)
}

func TestChangeStatusMainItem_InvalidTransition(t *testing.T) {
	svc := &mockMainItemService{}
	svc.changeStatusResult.err = apperrors.ErrInvalidStatus

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"completed"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestChangeStatusMainItem_RequiresPermission(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	// Member role doesn't have main_item:change_status
	deps := depsWithMemberRoleMainItem(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"progressing"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.changeStatusCalled)
}

func TestChangeStatusMainItem_InvalidItemID(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"status":"progressing"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/abc/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.changeStatusCalled)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/main-items/:itemId/available-transitions
// ---------------------------------------------------------------------------

func TestAvailableTransitionsMainItem_Success(t *testing.T) {
	svc := &mockMainItemService{}
	svc.availableTransitionsResult.transitions = []string{"progressing", "closed"}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/1/available-transitions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.availableTransitionsCalled)
	assert.Equal(t, uint(1), svc.lastItemID)
	assert.Equal(t, uint(5), svc.lastCallerID)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	transitions, ok := data["transitions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, transitions, 2)
}

func TestAvailableTransitionsMainItem_ItemNotFound(t *testing.T) {
	svc := &mockMainItemService{}
	svc.availableTransitionsResult.err = apperrors.ErrItemNotFound

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/999/available-transitions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAvailableTransitionsMainItem_InvalidItemID(t *testing.T) {
	svc := &mockMainItemService{}

	userRepo := &mockUserRepoForHandler{}
	subItemRepo := &mockSubItemRepoForHandler{}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/abc/available-transitions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.availableTransitionsCalled)
}

// ---------------------------------------------------------------------------
// Tests: Response shape matches Data Contract
// ---------------------------------------------------------------------------

func TestGetMainItem_ResponseShapeMatchesDataContract(t *testing.T) {
	svc := &mockMainItemService{}
	now := time.Now()
	assigneeID := uint(3)
	item := &model.MainItem{
		TeamID:          10,
		Code:            "TEST-00001",
		Title:           "接入新支付渠道",
		Priority:        "P1",
		ProposerID:      2,
		AssigneeID:      &assigneeID,
		StartDate:       &now,
		ExpectedEndDate: &now,
		Status:          "progressing",
		Completion:      45.5,
		IsKeyItem:       false,
	}
	item.ID = 1

	svc.getResult.item = item

	proposer := &model.User{}
	proposer.ID = 2
	proposer.DisplayName = "张三"
	assignee := &model.User{}
	assignee.ID = 3
	assignee.DisplayName = "李四"

	userRepo := &mockUserRepoForHandler{user: proposer}
	subItemRepo := &mockSubItemRepoForHandler{items: []*model.SubItem{}}

	deps := depsWithMainItemSvc(t, svc, userRepo, subItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})

	// Verify all fields from Data Contract are present
	assert.Equal(t, "TEST-00001", data["code"])
	assert.Equal(t, "接入新支付渠道", data["title"])
	assert.Equal(t, "P1", data["priority"])
	assert.Equal(t, float64(2), data["proposerId"])
	assert.Equal(t, float64(3), data["assigneeId"])
	assert.Equal(t, "progressing", data["status"])
	assert.Equal(t, 45.5, data["completion"])
	assert.Equal(t, false, data["isKeyItem"])
	assert.NotNil(t, data["subItems"])
}

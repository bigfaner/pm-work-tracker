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
// Mock DecisionLogService for handler tests
// ---------------------------------------------------------------------------

type mockDecisionLogService struct {
	createResult struct {
		log *model.DecisionLog
		err error
	}
	updateResult struct {
		log *model.DecisionLog
		err error
	}
	publishResult struct {
		log *model.DecisionLog
		err error
	}
	listResult struct {
		page *dto.PageResult[model.DecisionLog]
		err  error
	}

	// capture calls
	createCalled   bool
	lastMainItemID uint
	lastUserID     uint
	lastCreateReq  dto.DecisionLogCreateReq

	updateCalled  bool
	lastLogBizKey int64
	lastUpdateReq dto.DecisionLogUpdateReq

	publishCalled bool

	listCalled bool
	lastPage   dto.Pagination
}

func (m *mockDecisionLogService) Create(_ context.Context, mainItemID uint, userID uint, req dto.DecisionLogCreateReq) (*model.DecisionLog, error) {
	m.createCalled = true
	m.lastMainItemID = mainItemID
	m.lastUserID = userID
	m.lastCreateReq = req
	return m.createResult.log, m.createResult.err
}
func (m *mockDecisionLogService) Update(_ context.Context, bizKey int64, userID uint, req dto.DecisionLogUpdateReq) (*model.DecisionLog, error) {
	m.updateCalled = true
	m.lastLogBizKey = bizKey
	m.lastUserID = userID
	m.lastUpdateReq = req
	return m.updateResult.log, m.updateResult.err
}
func (m *mockDecisionLogService) Publish(_ context.Context, bizKey int64, userID uint) (*model.DecisionLog, error) {
	m.publishCalled = true
	m.lastLogBizKey = bizKey
	m.lastUserID = userID
	return m.publishResult.log, m.publishResult.err
}
func (m *mockDecisionLogService) List(_ context.Context, mainItemID uint, userID uint, page dto.Pagination) (*dto.PageResult[model.DecisionLog], error) {
	m.listCalled = true
	m.lastMainItemID = mainItemID
	m.lastUserID = userID
	m.lastPage = page
	return m.listResult.page, m.listResult.err
}

// ---------------------------------------------------------------------------
// Mock MainItemRepo for decision log handler tests
// ---------------------------------------------------------------------------

type mockMainItemRepoForDecisionLog struct {
	item *model.MainItem
	err  error
}

func (m *mockMainItemRepoForDecisionLog) Create(_ context.Context, _ *model.MainItem) error {
	return nil
}
func (m *mockMainItemRepoForDecisionLog) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	return m.item, m.err
}
func (m *mockMainItemRepoForDecisionLog) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) FindByBizKey(_ context.Context, bizKey int64) (*model.MainItem, error) {
	if m.item != nil || m.err != nil {
		return m.item, m.err
	}
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (m *mockMainItemRepoForDecisionLog) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}
func (m *mockMainItemRepoForDecisionLog) NextCode(_ context.Context, _ int64) (string, error) {
	return "", nil
}
func (m *mockMainItemRepoForDecisionLog) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockMainItemRepoForDecisionLog) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}

// Shared mock instances for constructor tests
var (
	mockDecisionLogSvc = &mockDecisionLogService{}
	mockUserRepo       = &mockUserRepoForHandler{}
	mockMainItemRepo   = &mockMainItemRepoForDecisionLog{}
)

func TestNewDecisionLogHandler_NilDependencies(t *testing.T) {
	t.Run("panics on nil service", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDecisionLogHandler(nil, mockUserRepo, mockMainItemRepo)
		})
	})
	t.Run("panics on nil userRepo", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDecisionLogHandler(mockDecisionLogSvc, nil, mockMainItemRepo)
		})
	})
	t.Run("panics on nil mainItemRepo", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDecisionLogHandler(mockDecisionLogSvc, mockUserRepo, nil)
		})
	})
	t.Run("succeeds with all dependencies", func(t *testing.T) {
		h := NewDecisionLogHandler(mockDecisionLogSvc, mockUserRepo, mockMainItemRepo)
		assert.NotNil(t, h)
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testDecisionLog creates a DecisionLog model for tests.
func testDecisionLog(id uint, mainItemKey int64) *model.DecisionLog {
	return &model.DecisionLog{
		ID:          id,
		BizKey:      int64(id*100 + 1),
		MainItemKey: mainItemKey,
		TeamKey:     10,
		Category:    "technical",
		Tags:        `["缓存策略"]`,
		Content:     "Decided to use Redis caching",
		LogStatus:   "draft",
		CreatedBy:   5,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
}

// depsWithDecisionLogSvc wires a mock DecisionLogService into test deps.
func depsWithDecisionLogSvc(t *testing.T, svc *mockDecisionLogService, userRepo repository.UserRepo, mainItemRepo *mockMainItemRepoForDecisionLog) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{RoleKey: func() *int64 { v := int64(1); return &v }()}}
	deps.DecisionLog = NewDecisionLogHandler(svc, userRepo, mainItemRepo)
	return deps
}

// depsWithMemberRoleDecisionLog wires a member (non-PM) role for testing RBAC.
func depsWithMemberRoleDecisionLog(t *testing.T, svc *mockDecisionLogService, userRepo repository.UserRepo, mainItemRepo *mockMainItemRepoForDecisionLog) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{member: &model.TeamMember{RoleKey: func() *int64 { v := int64(2); return &v }()}}
	deps.DecisionLog = NewDecisionLogHandler(svc, userRepo, mainItemRepo)
	return deps
}

// ---------------------------------------------------------------------------
// Tests: POST /teams/:teamId/main-items/:itemId/decision-logs (Create)
// ---------------------------------------------------------------------------

func TestDecisionLogCreate_Success(t *testing.T) {
	svc := &mockDecisionLogService{}
	log := testDecisionLog(1, 100)
	svc.createResult.log = log

	mainItemRepo := &mockMainItemRepoForDecisionLog{
		item: &model.MainItem{BaseModel: model.BaseModel{ID: 1}, TeamKey: 10},
	}

	userRepo := &mockUserRepoForHandler{
		user: &model.User{DisplayName: "Test User"},
	}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"category":"technical","tags":["缓存策略"],"content":"Decided to use Redis caching","logStatus":"draft"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/100/decision-logs", strings.NewReader(body))
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
	assert.Equal(t, "technical", data["category"])
	assert.Equal(t, "draft", data["logStatus"])
	assert.True(t, svc.createCalled)
	assert.Equal(t, uint(1), svc.lastMainItemID)
	assert.Equal(t, uint(5), svc.lastUserID)
}

func TestDecisionLogCreate_RequiresPermission(t *testing.T) {
	svc := &mockDecisionLogService{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{}

	// Member role doesn't have main_item:update
	deps := depsWithMemberRoleDecisionLog(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"category":"technical","content":"Test","logStatus":"draft"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/100/decision-logs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.createCalled)
}

func TestDecisionLogCreate_InvalidBody(t *testing.T) {
	svc := &mockDecisionLogService{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{
		item: &model.MainItem{BaseModel: model.BaseModel{ID: 1}},
	}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/100/decision-logs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.createCalled)
}

func TestDecisionLogCreate_MainItemNotFound(t *testing.T) {
	svc := &mockDecisionLogService{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{
		err: apperrors.ErrItemNotFound,
	}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"category":"technical","content":"Test","logStatus":"draft"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/999/decision-logs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.False(t, svc.createCalled)
}

func TestDecisionLogCreate_ServiceError(t *testing.T) {
	svc := &mockDecisionLogService{}
	svc.createResult.err = errors.New("unexpected")

	mainItemRepo := &mockMainItemRepoForDecisionLog{
		item: &model.MainItem{BaseModel: model.BaseModel{ID: 1}},
	}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"category":"technical","content":"Test","logStatus":"draft"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/100/decision-logs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PUT /teams/:teamId/main-items/:itemId/decision-logs/:logId (Update)
// ---------------------------------------------------------------------------

func TestDecisionLogUpdate_Success(t *testing.T) {
	svc := &mockDecisionLogService{}
	updatedLog := testDecisionLog(1, 100)
	updatedLog.Content = "Updated content"
	updatedLog.Category = "risk"
	svc.updateResult.log = updatedLog

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{
		user: &model.User{DisplayName: "Test User"},
	}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"category":"risk","content":"Updated content"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/100/decision-logs/101", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.updateCalled)
	assert.Equal(t, int64(101), svc.lastLogBizKey)
	assert.Equal(t, "risk", svc.lastUpdateReq.Category)
	assert.Equal(t, "Updated content", svc.lastUpdateReq.Content)
}

func TestDecisionLogUpdate_RequiresPermission(t *testing.T) {
	svc := &mockDecisionLogService{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithMemberRoleDecisionLog(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"category":"risk","content":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/100/decision-logs/101", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.updateCalled)
}

func TestDecisionLogUpdate_InvalidLogId(t *testing.T) {
	svc := &mockDecisionLogService{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"category":"risk","content":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/100/decision-logs/abc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.updateCalled)
}

func TestDecisionLogUpdate_InvalidBody(t *testing.T) {
	svc := &mockDecisionLogService{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/100/decision-logs/101", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.updateCalled)
}

func TestDecisionLogUpdate_Forbidden(t *testing.T) {
	svc := &mockDecisionLogService{}
	svc.updateResult.err = apperrors.ErrForbidden

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"category":"risk","content":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/100/decision-logs/101", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDecisionLogUpdate_NotFound(t *testing.T) {
	svc := &mockDecisionLogService{}
	svc.updateResult.err = apperrors.ErrDecisionLogNotFound

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"category":"risk","content":"Updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/10/main-items/100/decision-logs/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PATCH /teams/:teamId/main-items/:itemId/decision-logs/:logId/publish (Publish)
// ---------------------------------------------------------------------------

func TestDecisionLogPublish_Success(t *testing.T) {
	svc := &mockDecisionLogService{}
	publishedLog := testDecisionLog(1, 100)
	publishedLog.LogStatus = "published"
	svc.publishResult.log = publishedLog

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{
		user: &model.User{DisplayName: "Test User"},
	}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/main-items/100/decision-logs/101/publish", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "published", data["logStatus"])

	assert.True(t, svc.publishCalled)
	assert.Equal(t, int64(101), svc.lastLogBizKey)
	assert.Equal(t, uint(5), svc.lastUserID)
}

func TestDecisionLogPublish_RequiresPermission(t *testing.T) {
	svc := &mockDecisionLogService{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithMemberRoleDecisionLog(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/main-items/100/decision-logs/101/publish", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.publishCalled)
}

func TestDecisionLogPublish_InvalidLogId(t *testing.T) {
	svc := &mockDecisionLogService{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/main-items/100/decision-logs/abc/publish", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.publishCalled)
}

func TestDecisionLogPublish_Forbidden(t *testing.T) {
	svc := &mockDecisionLogService{}
	svc.publishResult.err = apperrors.ErrForbidden

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/main-items/100/decision-logs/101/publish", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDecisionLogPublish_NotFound(t *testing.T) {
	svc := &mockDecisionLogService{}
	svc.publishResult.err = apperrors.ErrDecisionLogNotFound

	mainItemRepo := &mockMainItemRepoForDecisionLog{}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/10/main-items/100/decision-logs/999/publish", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /teams/:teamId/main-items/:itemId/decision-logs (List)
// ---------------------------------------------------------------------------

func TestDecisionLogList_Success(t *testing.T) {
	svc := &mockDecisionLogService{}
	svc.listResult.page = &dto.PageResult[model.DecisionLog]{
		Items: []model.DecisionLog{*testDecisionLog(1, 100), *testDecisionLog(2, 100)},
		Total: 2,
		Page:  1,
		Size:  20,
	}

	mainItemRepo := &mockMainItemRepoForDecisionLog{
		item: &model.MainItem{BaseModel: model.BaseModel{ID: 1}},
	}
	userRepo := &mockUserRepoForHandler{
		user: &model.User{DisplayName: fmt.Sprintf("User %d", 5)},
	}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/100/decision-logs?page=1&pageSize=20", nil)
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
	assert.Equal(t, uint(1), svc.lastMainItemID)
}

func TestDecisionLogList_RequiresTeamMembership(t *testing.T) {
	// List requires team membership (enforced by TeamScopeMiddleware)
	// Not a permission-gated route, so PM role is not needed
	svc := &mockDecisionLogService{}
	svc.listResult.page = &dto.PageResult[model.DecisionLog]{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{
		item: &model.MainItem{BaseModel: model.BaseModel{ID: 1}},
	}
	userRepo := &mockUserRepoForHandler{}

	// Use member role — should still work since list doesn't require main_item:update
	deps := depsWithMemberRoleDecisionLog(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/100/decision-logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// Should succeed — list only requires team membership
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.listCalled)
}

func TestDecisionLogList_PaginationDefaults(t *testing.T) {
	svc := &mockDecisionLogService{}
	svc.listResult.page = &dto.PageResult[model.DecisionLog]{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{
		item: &model.MainItem{BaseModel: model.BaseModel{ID: 1}},
	}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/100/decision-logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Service receives the raw pagination — defaults applied inside service
	assert.True(t, svc.listCalled)
}

func TestDecisionLogList_MainItemNotFound(t *testing.T) {
	svc := &mockDecisionLogService{}

	mainItemRepo := &mockMainItemRepoForDecisionLog{
		err: apperrors.ErrItemNotFound,
	}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/999/decision-logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.False(t, svc.listCalled)
}

func TestDecisionLogList_ServiceError(t *testing.T) {
	svc := &mockDecisionLogService{}
	svc.listResult.err = errors.New("db error")

	mainItemRepo := &mockMainItemRepoForDecisionLog{
		item: &model.MainItem{BaseModel: model.BaseModel{ID: 1}},
	}
	userRepo := &mockUserRepoForHandler{}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/10/main-items/100/decision-logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: Response shape matches Data Contract
// ---------------------------------------------------------------------------

func TestDecisionLogCreate_ResponseShapeMatchesDataContract(t *testing.T) {
	svc := &mockDecisionLogService{}
	now := time.Now()
	log := &model.DecisionLog{
		ID:          1,
		BizKey:      1893456789012345678,
		MainItemKey: 1893456789012345001,
		TeamKey:     10,
		Category:    "technical",
		Tags:        `["缓存策略","性能优化"]`,
		Content:     "决定采用 Redis 缓存热点数据",
		LogStatus:   "published",
		CreatedBy:   5,
		CreateTime:  now,
		UpdateTime:  now,
	}
	svc.createResult.log = log

	mainItemRepo := &mockMainItemRepoForDecisionLog{
		item: &model.MainItem{BaseModel: model.BaseModel{ID: 1}},
	}
	userRepo := &mockUserRepoForHandler{
		user: &model.User{DisplayName: "张三"},
	}

	deps := depsWithDecisionLogSvc(t, svc, userRepo, mainItemRepo)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	body := `{"category":"technical","tags":["缓存策略","性能优化"],"content":"决定采用 Redis 缓存热点数据","logStatus":"published"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/10/main-items/100/decision-logs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})

	// Verify all VO fields are present
	assert.Equal(t, "1893456789012345678", data["bizKey"])
	assert.Equal(t, "1893456789012345001", data["mainItemKey"])
	assert.Equal(t, "technical", data["category"])
	assert.Equal(t, "决定采用 Redis 缓存热点数据", data["content"])
	assert.Equal(t, "published", data["logStatus"])
	assert.Equal(t, "5", data["createdBy"])
	assert.Equal(t, "张三", data["creatorName"])
	assert.NotNil(t, data["createTime"])
	assert.NotNil(t, data["updateTime"])

	// Tags should be parsed to []string
	tags, ok := data["tags"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tags, 2)
}

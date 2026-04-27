package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Mock AdminService for handler tests
// ---------------------------------------------------------------------------

type mockAdminService struct {
		listUsersResult struct {
		users []*model.User
		err   error
	}
	listAllTeamsResult struct {
		teams []*dto.AdminTeamDTO
		err   error
	}

	// New method results
	listUsersFilteredResult struct {
		items []*dto.AdminUserDTO
		total int
		err   error
	}
	getUserResult struct {
		user *dto.AdminUserDTO
		err  error
	}
	createUserResult struct {
		user *dto.AdminUserDTO
		err  error
	}
	updateUserResult struct {
		user *dto.AdminUserDTO
		err  error
	}
	toggleUserStatusResult struct {
		user *dto.AdminUserDTO
		err  error
	}

	// capture calls
	listUsersCalled    bool
	listAllTeamsCalled bool

	// new captures
	lastListUsersSearch   string
	lastListUsersPage     int
	lastListUsersPageSize int
	lastGetUserID         int64
	lastCreateUserReq     *dto.CreateUserReq
	lastUpdateUserID      int64
	lastUpdateUserReq     *dto.UpdateUserReq
	lastToggleCallerID    uint
	lastToggleTargetID    int64
	lastToggleStatus      string
}

func (m *mockAdminService) ListUsers(_ context.Context, search string, page, pageSize int) ([]*dto.AdminUserDTO, int, error) {
	m.listUsersCalled = true
	m.lastListUsersSearch = search
	m.lastListUsersPage = page
	m.lastListUsersPageSize = pageSize
	return m.listUsersFilteredResult.items, m.listUsersFilteredResult.total, m.listUsersFilteredResult.err
}

func (m *mockAdminService) GetUser(_ context.Context, userID int64) (*dto.AdminUserDTO, error) {
	m.lastGetUserID = userID
	return m.getUserResult.user, m.getUserResult.err
}

func (m *mockAdminService) CreateUser(_ context.Context, req *dto.CreateUserReq) (*dto.AdminUserDTO, error) {
	m.lastCreateUserReq = req
	return m.createUserResult.user, m.createUserResult.err
}

func (m *mockAdminService) UpdateUser(_ context.Context, userID int64, req *dto.UpdateUserReq) (*dto.AdminUserDTO, error) {
	m.lastUpdateUserID = userID
	m.lastUpdateUserReq = req
	return m.updateUserResult.user, m.updateUserResult.err
}

func (m *mockAdminService) ToggleUserStatus(_ context.Context, callerID uint, targetUserID int64, status string) (*dto.AdminUserDTO, error) {
	m.lastToggleCallerID = callerID
	m.lastToggleTargetID = targetUserID
	m.lastToggleStatus = status
	return m.toggleUserStatusResult.user, m.toggleUserStatusResult.err
}

func (m *mockAdminService) ListAllTeams(_ context.Context) ([]*dto.AdminTeamDTO, error) {
	m.listAllTeamsCalled = true
	return m.listAllTeamsResult.teams, m.listAllTeamsResult.err
}

func (m *mockAdminService) ResetPassword(_ context.Context, _ int64, _ string) (*dto.ResetPasswordResp, error) {
	return nil, nil
}

func (m *mockAdminService) SoftDeleteUser(_ context.Context, _ uint, _ int64) error {
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// depsWithAdminSvc creates test deps with a mock AdminService wired in.
func depsWithAdminSvc(t *testing.T, svc *mockAdminService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.Admin = NewAdminHandler(svc)
	return deps
}

// signSuperAdminToken creates a JWT token for the superadmin user.
func signSuperAdminToken(t *testing.T, userID uint) string {
	t.Helper()
	return signTestToken(t, userID, "admin")
}

// ---------------------------------------------------------------------------
// Tests: GET /v1/admin/users (ListUsers)
// ---------------------------------------------------------------------------

func TestAdminListUsers_Success(t *testing.T) {
	svc := &mockAdminService{}
	svc.listUsersFilteredResult.items = []*dto.AdminUserDTO{
		{BizKey: "1", Username: "alice", DisplayName: "Alice", IsSuperAdmin: true, Status: "enabled", Teams: []dto.TeamSummary{}},
		{BizKey: "2", Username: "bob", DisplayName: "Bob", Status: "enabled", Teams: []dto.TeamSummary{}},
	}
	svc.listUsersFilteredResult.total = 2

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "data should be an object")

	items, ok := data["items"].([]interface{})
	require.True(t, ok, "items should be an array")
	assert.Len(t, items, 2)

	user0 := items[0].(map[string]interface{})
	assert.Equal(t, "1", user0["bizKey"])
	assert.Equal(t, "alice", user0["username"])
	assert.Equal(t, "Alice", user0["displayName"])
	assert.Equal(t, true, user0["isSuperAdmin"])

	assert.Equal(t, float64(2), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(20), data["pageSize"])
}

func TestAdminListUsers_WithSearch(t *testing.T) {
	svc := &mockAdminService{}
	svc.listUsersFilteredResult.items = []*dto.AdminUserDTO{
		{BizKey: "1", Username: "alice", DisplayName: "Alice"},
	}
	svc.listUsersFilteredResult.total = 1

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?search=alice", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "alice", svc.lastListUsersSearch)
}

func TestAdminListUsers_DefaultPageSize(t *testing.T) {
	svc := &mockAdminService{}
	svc.listUsersFilteredResult.items = []*dto.AdminUserDTO{}
	svc.listUsersFilteredResult.total = 0

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 20, svc.lastListUsersPageSize)
	assert.Equal(t, 1, svc.lastListUsersPage)
}

func TestAdminListUsers_ServiceError(t *testing.T) {
	svc := &mockAdminService{}
	svc.listUsersFilteredResult.err = errors.New("db error")

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdminListUsers_CustomPagination(t *testing.T) {
	svc := &mockAdminService{}
	svc.listUsersFilteredResult.items = []*dto.AdminUserDTO{}
	svc.listUsersFilteredResult.total = 0

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?page=3&pageSize=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 3, svc.lastListUsersPage)
	assert.Equal(t, 10, svc.lastListUsersPageSize)
}

// ---------------------------------------------------------------------------
// Tests: POST /v1/admin/users (CreateUser)
// ---------------------------------------------------------------------------

func TestAdminCreateUser_Success(t *testing.T) {
	svc := &mockAdminService{}
	svc.createUserResult.user = &dto.AdminUserDTO{
		BizKey: "3",
		Username:        "newuser",
		DisplayName:     "New User",
		Email:           "new@test.com",
		Status: "enabled",
		Teams:           []dto.TeamSummary{{BizKey: "10", Name: "Team A", }},
		InitialPassword: "Abc123XYZdef",
	}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"username":"newuser","displayName":"New User","email":"new@test.com","teamKey":"10"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
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
	assert.Equal(t, "newuser", data["username"])
	assert.Equal(t, "Abc123XYZdef", data["initialPassword"])
	assert.NotNil(t, svc.lastCreateUserReq)
	assert.Equal(t, "newuser", svc.lastCreateUserReq.Username)
	require.NotNil(t, svc.lastCreateUserReq.TeamKey)
	assert.Equal(t, "10", *svc.lastCreateUserReq.TeamKey)
}

func TestAdminCreateUser_DuplicateUsername(t *testing.T) {
	svc := &mockAdminService{}
	svc.createUserResult.err = apperrors.ErrUserExists

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"username":"existing","displayName":"Existing"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "USER_EXISTS", resp["code"])
}

func TestAdminCreateUser_ValidationFail(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminCreateUser_TeamNotFound(t *testing.T) {
	svc := &mockAdminService{}
	svc.createUserResult.err = apperrors.ErrTeamNotFound

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"username":"newuser","displayName":"New User","teamKey":"999"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "TEAM_NOT_FOUND", resp["code"])
}

func TestAdminCreateUser_InternalError(t *testing.T) {
	svc := &mockAdminService{}
	svc.createUserResult.err = apperrors.ErrInternal

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"username":"newuser","displayName":"New User"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /v1/admin/users/:userId (GetUser)
// ---------------------------------------------------------------------------

func TestAdminGetUser_Success(t *testing.T) {
	svc := &mockAdminService{}
	svc.getUserResult.user = &dto.AdminUserDTO{
		BizKey: "5",
		Username:     "bob",
		DisplayName:  "Bob",
		Email:        "bob@test.com",
		Status: "enabled",
		IsSuperAdmin: false,
		Teams:        []dto.TeamSummary{{BizKey: "1", Name: "Team A", }},
	}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/5", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(5), svc.lastGetUserID)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "bob", data["username"])
	assert.Equal(t, "bob@test.com", data["email"])
}

func TestAdminGetUser_NotFound(t *testing.T) {
	svc := &mockAdminService{}
	svc.getUserResult.err = apperrors.ErrUserNotFound

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "USER_NOT_FOUND", resp["code"])
}

func TestAdminGetUser_InvalidId(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminGetUser_ServiceError(t *testing.T) {
	svc := &mockAdminService{}
	svc.getUserResult.err = apperrors.ErrInternal

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/5", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PUT /v1/admin/users/:userId (UpdateUser)
// ---------------------------------------------------------------------------

func TestAdminUpdateUser_Success(t *testing.T) {
	svc := &mockAdminService{}
	svc.updateUserResult.user = &dto.AdminUserDTO{
		BizKey: "5",
		Username:    "bob",
		DisplayName: "Robert",
		Email:       "robert@test.com",
		Status: "enabled",
		Teams:       []dto.TeamSummary{{BizKey: "2", Name: "Team B", }},
	}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"displayName":"Robert","email":"robert@test.com"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(5), svc.lastUpdateUserID)
	require.NotNil(t, svc.lastUpdateUserReq.DisplayName)
	assert.Equal(t, "Robert", *svc.lastUpdateUserReq.DisplayName)
}

func TestAdminUpdateUser_NotFound(t *testing.T) {
	svc := &mockAdminService{}
	svc.updateUserResult.err = apperrors.ErrUserNotFound

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"displayName":"Robert"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminUpdateUser_ValidationFail(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"displayName":""}` // fails min=1
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminUpdateUser_InvalidId(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"displayName":"Robert"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/abc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminUpdateUser_ServiceError(t *testing.T) {
	svc := &mockAdminService{}
	svc.updateUserResult.err = apperrors.ErrInternal

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"displayName":"Robert"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PUT /v1/admin/users/:userId/status (ToggleUserStatus)
// ---------------------------------------------------------------------------

func TestAdminListTeams_OutOfBoundsPage(t *testing.T) {
	svc := &mockAdminService{}
	svc.listAllTeamsResult.teams = []*dto.AdminTeamDTO{
		{BizKey: "1", Name: "T1"},
	}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/teams?page=99&pageSize=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	assert.Empty(t, items)
	assert.Equal(t, float64(1), data["total"])
}

func TestAdminToggleUserStatus_MissingBody(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5/status", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminToggleUserStatus_DisableSuccess(t *testing.T) {
	svc := &mockAdminService{}
	svc.toggleUserStatusResult.user = &dto.AdminUserDTO{
		BizKey: "5",
		Username: "bob",
		Status: "disabled",
		Teams:    []dto.TeamSummary{},
	}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"status":"disabled"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(1), svc.lastToggleCallerID)
	assert.Equal(t, int64(5), svc.lastToggleTargetID)
	assert.Equal(t, "disabled", svc.lastToggleStatus)
}

func TestAdminToggleUserStatus_CannotDisableSelf(t *testing.T) {
	svc := &mockAdminService{}
	svc.toggleUserStatusResult.err = apperrors.ErrCannotDisableSelf

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"status":"disabled"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "CANNOT_DISABLE_SELF", resp["code"])
}

func TestAdminToggleUserStatus_InvalidStatus(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"status":"suspended"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminToggleUserStatus_UserNotFound(t *testing.T) {
	svc := &mockAdminService{}
	svc.toggleUserStatusResult.err = apperrors.ErrUserNotFound

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"status":"disabled"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/999/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminToggleUserStatus_EnableSuccess(t *testing.T) {
	svc := &mockAdminService{}
	svc.toggleUserStatusResult.user = &dto.AdminUserDTO{
		BizKey: "5", Username: "bob", Status: "enabled", Teams: []dto.TeamSummary{},
	}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"status":"enabled"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "enabled", svc.lastToggleStatus)
}

func TestAdminToggleUserStatus_InvalidId(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"status":"disabled"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/abc/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminToggleUserStatus_ServiceError(t *testing.T) {
	svc := &mockAdminService{}
	svc.toggleUserStatusResult.err = apperrors.ErrInternal

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	body := `{"status":"disabled"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /v1/admin/teams (ListTeams)
// ---------------------------------------------------------------------------

func TestAdminListTeams_Success(t *testing.T) {
	svc := &mockAdminService{}
	svc.listAllTeamsResult.teams = []*dto.AdminTeamDTO{
		{BizKey: "1", Name: "Alpha", PMDisplayName: "Alice", MemberCount: 3, MainItemCount: 5, CreatedAt: "2026-01-01T00:00:00Z"},
		{BizKey: "2", Name: "Beta", PMDisplayName: "Bob", MemberCount: 2, MainItemCount: 10, CreatedAt: "2026-02-01T00:00:00Z"},
	}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/teams", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "data should be an object")

	items, ok := data["items"].([]interface{})
	require.True(t, ok, "items should be an array")
	assert.Len(t, items, 2)

	team0 := items[0].(map[string]interface{})
	assert.Equal(t, "1", team0["bizKey"])
	assert.Equal(t, "Alpha", team0["name"])
	assert.Equal(t, "Alice", team0["pmDisplayName"])
	assert.Equal(t, float64(3), team0["memberCount"])
	assert.Equal(t, float64(5), team0["mainItemCount"])

	assert.Equal(t, float64(2), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(50), data["pageSize"])
}

func TestAdminListTeams_Pagination(t *testing.T) {
	teams := []*dto.AdminTeamDTO{
		{BizKey: "1", Name: "T1", PMDisplayName: "PM1", MemberCount: 1, MainItemCount: 1},
		{BizKey: "2", Name: "T2", PMDisplayName: "PM2", MemberCount: 2, MainItemCount: 2},
		{BizKey: "3", Name: "T3", PMDisplayName: "PM3", MemberCount: 3, MainItemCount: 3},
	}

	svc := &mockAdminService{}
	svc.listAllTeamsResult.teams = teams

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/teams?page=2&pageSize=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	assert.Len(t, items, 1)
	assert.Equal(t, float64(3), data["total"])
	assert.Equal(t, float64(2), data["page"])
	assert.Equal(t, float64(1), data["pageSize"])

	item0 := items[0].(map[string]interface{})
	assert.Equal(t, "T2", item0["name"])
}

func TestAdminListTeams_Empty(t *testing.T) {
	svc := &mockAdminService{}
	svc.listAllTeamsResult.teams = []*dto.AdminTeamDTO{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/teams", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	assert.Empty(t, items)
	assert.Equal(t, float64(0), data["total"])
}

func TestAdminListTeams_ServiceError(t *testing.T) {
	svc := &mockAdminService{}
	svc.listAllTeamsResult.err = errors.New("db error")

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps, nil)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/teams", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

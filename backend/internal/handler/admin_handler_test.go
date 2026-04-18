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
	"gorm.io/gorm"

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
	setCanCreateTeamErr error
	listAllTeamsResult  struct {
		teams []*dto.AdminTeamDTO
		err   error
	}

	// capture calls
	listUsersCalled        bool
	setCanCreateTeamCalled bool
	lastSuperAdminID       uint
	lastTargetUserID       uint
	lastCanCreate          bool
	listAllTeamsCalled     bool
}

func (m *mockAdminService) ListUsers(_ context.Context) ([]*model.User, error) {
	m.listUsersCalled = true
	return m.listUsersResult.users, m.listUsersResult.err
}

func (m *mockAdminService) SetCanCreateTeam(_ context.Context, superAdminID, targetUserID uint, canCreate bool) error {
	m.setCanCreateTeamCalled = true
	m.lastSuperAdminID = superAdminID
	m.lastTargetUserID = targetUserID
	m.lastCanCreate = canCreate
	return m.setCanCreateTeamErr
}

func (m *mockAdminService) ListAllTeams(_ context.Context) ([]*dto.AdminTeamDTO, error) {
	m.listAllTeamsCalled = true
	return m.listAllTeamsResult.teams, m.listAllTeamsResult.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// depsWithAdminSvc creates test deps with a mock AdminService wired in.
func depsWithAdminSvc(t *testing.T, svc *mockAdminService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.Admin = NewAdminHandlerWithDeps(svc)
	return deps
}

// signSuperAdminToken creates a JWT token with superadmin role.
func signSuperAdminToken(t *testing.T, userID uint) string {
	t.Helper()
	return signTestToken(t, userID, "superadmin")
}

// ---------------------------------------------------------------------------
// Tests: GET /api/v1/admin/users (ListUsers)
// ---------------------------------------------------------------------------

func TestAdminListUsers_Success(t *testing.T) {
	svc := &mockAdminService{}
	svc.listUsersResult.users = []*model.User{
		{Model: gorm.Model{ID: 1}, Username: "alice", DisplayName: "Alice", CanCreateTeam: true, IsSuperAdmin: true},
		{Model: gorm.Model{ID: 2}, Username: "bob", DisplayName: "Bob", CanCreateTeam: false, IsSuperAdmin: false},
	}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

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
	assert.Equal(t, float64(1), user0["id"])
	assert.Equal(t, "alice", user0["username"])
	assert.Equal(t, "Alice", user0["displayName"])
	assert.Equal(t, true, user0["canCreateTeam"])
	assert.Equal(t, true, user0["isSuperAdmin"])

	assert.Equal(t, float64(2), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(50), data["pageSize"])
}

func TestAdminListUsers_Pagination(t *testing.T) {
	// Create 5 users, request page 2 with pageSize=2
	users := []*model.User{
		{Model: gorm.Model{ID: 1}, Username: "u1", DisplayName: "User 1"},
		{Model: gorm.Model{ID: 2}, Username: "u2", DisplayName: "User 2"},
		{Model: gorm.Model{ID: 3}, Username: "u3", DisplayName: "User 3"},
		{Model: gorm.Model{ID: 4}, Username: "u4", DisplayName: "User 4"},
		{Model: gorm.Model{ID: 5}, Username: "u5", DisplayName: "User 5"},
	}

	svc := &mockAdminService{}
	svc.listUsersResult.users = users

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?page=2&pageSize=2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)

	items := data["items"].([]interface{})
	assert.Len(t, items, 2)
	assert.Equal(t, float64(5), data["total"])
	assert.Equal(t, float64(2), data["page"])
	assert.Equal(t, float64(2), data["pageSize"])

	// Should be users 3 and 4 (index 2 and 3)
	item0 := items[0].(map[string]interface{})
	assert.Equal(t, "u3", item0["username"])
	item1 := items[1].(map[string]interface{})
	assert.Equal(t, "u4", item1["username"])
}

func TestAdminListUsers_DefaultPagination(t *testing.T) {
	svc := &mockAdminService{}
	svc.listUsersResult.users = []*model.User{
		{Model: gorm.Model{ID: 1}, Username: "alice", DisplayName: "Alice"},
	}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(50), data["pageSize"])
}

func TestAdminListUsers_Empty(t *testing.T) {
	svc := &mockAdminService{}
	svc.listUsersResult.users = []*model.User{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
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

func TestAdminListUsers_ServiceError(t *testing.T) {
	svc := &mockAdminService{}
	svc.listUsersResult.err = errors.New("db error")

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: PUT /api/v1/admin/users/:userId/can-create-team
// ---------------------------------------------------------------------------

func TestAdminUpdateCanCreateTeam_Success(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	body := `{"canCreateTeam":true}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5/can-create-team", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.setCanCreateTeamCalled)
	assert.Equal(t, uint(1), svc.lastSuperAdminID)
	assert.Equal(t, uint(5), svc.lastTargetUserID)
	assert.True(t, svc.lastCanCreate)
}

func TestAdminUpdateCanCreateTeam_CannotModifySelf(t *testing.T) {
	svc := &mockAdminService{}
	svc.setCanCreateTeamErr = apperrors.ErrCannotModifySelf

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	body := `{"canCreateTeam":true}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/1/can-create-team", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "CANNOT_MODIFY_SELF", resp["code"])
}

func TestAdminUpdateCanCreateTeam_InvalidUserId(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	body := `{"canCreateTeam":true}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/abc/can-create-team", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", resp["code"])
}

func TestAdminUpdateCanCreateTeam_InvalidBody(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5/can-create-team", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminUpdateCanCreateTeam_Revoke(t *testing.T) {
	svc := &mockAdminService{}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	body := `{"canCreateTeam":false}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/5/can-create-team", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, svc.lastCanCreate)
}

func TestAdminUpdateCanCreateTeam_TargetNotFound(t *testing.T) {
	svc := &mockAdminService{}
	svc.setCanCreateTeamErr = apperrors.ErrNotFound

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	body := `{"canCreateTeam":true}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/999/can-create-team", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GET /api/v1/admin/teams (ListTeams)
// ---------------------------------------------------------------------------

func TestAdminListTeams_Success(t *testing.T) {
	svc := &mockAdminService{}
	svc.listAllTeamsResult.teams = []*dto.AdminTeamDTO{
		{ID: 1, Name: "Alpha", PMDisplayName: "Alice", MemberCount: 3, MainItemCount: 5, CreatedAt: "2026-01-01T00:00:00Z"},
		{ID: 2, Name: "Beta", PMDisplayName: "Bob", MemberCount: 2, MainItemCount: 10, CreatedAt: "2026-02-01T00:00:00Z"},
	}

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

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
	assert.Equal(t, float64(1), team0["id"])
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
		{ID: 1, Name: "T1", PMDisplayName: "PM1", MemberCount: 1, MainItemCount: 1},
		{ID: 2, Name: "T2", PMDisplayName: "PM2", MemberCount: 2, MainItemCount: 2},
		{ID: 3, Name: "T3", PMDisplayName: "PM3", MemberCount: 3, MainItemCount: 3},
	}

	svc := &mockAdminService{}
	svc.listAllTeamsResult.teams = teams

	deps := depsWithAdminSvc(t, svc)
	r := SetupRouter(deps)

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
	r := SetupRouter(deps)

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
	r := SetupRouter(deps)

	token := signSuperAdminToken(t, 1)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/teams", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

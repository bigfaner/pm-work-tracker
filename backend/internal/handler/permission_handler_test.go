package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/pkg/permissions"
	"pm-work-tracker/backend/internal/service"
)

// ---------------------------------------------------------------------------
// Mock RoleService for permission handler tests
// ---------------------------------------------------------------------------

type mockRoleServiceForPermission struct {
	listPermissionCodesFn func(ctx context.Context) []permissions.ResourcePermissions
	getUserPermissionsFn  func(ctx context.Context, userID uint) (*service.UserPermissions, error)

	// capture calls
	listPermissionCodesCalled bool
	getUserPermissionsCalled  bool
	lastUserID                uint
}

func (m *mockRoleServiceForPermission) ListRoles(_ context.Context) ([]service.RoleListItem, error) {
	return nil, nil
}
func (m *mockRoleServiceForPermission) GetRole(_ context.Context, _ int64) (*service.RoleDetail, error) {
	return nil, nil
}
func (m *mockRoleServiceForPermission) CreateRole(_ context.Context, _ dto.CreateRoleReq) (*service.RoleListItem, error) {
	return nil, nil
}
func (m *mockRoleServiceForPermission) UpdateRole(_ context.Context, _ int64, _ dto.UpdateRoleReq) (*service.RoleDetail, error) {
	return nil, nil
}
func (m *mockRoleServiceForPermission) DeleteRole(_ context.Context, _ int64) error {
	return nil
}
func (m *mockRoleServiceForPermission) ListPermissionCodes(ctx context.Context) []permissions.ResourcePermissions {
	m.listPermissionCodesCalled = true
	if m.listPermissionCodesFn != nil {
		return m.listPermissionCodesFn(ctx)
	}
	return nil
}
func (m *mockRoleServiceForPermission) GetUserPermissions(ctx context.Context, userID uint) (*service.UserPermissions, error) {
	m.getUserPermissionsCalled = true
	m.lastUserID = userID
	if m.getUserPermissionsFn != nil {
		return m.getUserPermissionsFn(ctx, userID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Tests: GET /api/v1/admin/permissions (ListPermissionCodes)
// ---------------------------------------------------------------------------

func TestPermissionHandler_ListPermissionCodes_Success(t *testing.T) {
	mockSvc := &mockRoleServiceForPermission{
		listPermissionCodesFn: func(_ context.Context) []permissions.ResourcePermissions {
			return []permissions.ResourcePermissions{
				{
					Resource: "team",
					Permissions: []permissions.Permission{
						{Code: "team:create", Description: "Create team"},
						{Code: "team:read", Description: "Read team"},
					},
				},
				{
					Resource: "main_item",
					Permissions: []permissions.Permission{
						{Code: "main_item:create", Description: "Create main item"},
					},
				},
			}
		},
	}

	deps := depsWithPermissionHandler(t, mockSvc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 1, "admin")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, mockSvc.listPermissionCodesCalled)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)

	first := data[0].(map[string]interface{})
	assert.Equal(t, "team", first["resource"])
	perms := first["permissions"].([]interface{})
	assert.Len(t, perms, 2)
}

func TestPermissionHandler_ListPermissionCodes_EmptyResult(t *testing.T) {
	mockSvc := &mockRoleServiceForPermission{
		listPermissionCodesFn: func(_ context.Context) []permissions.ResourcePermissions {
			return []permissions.ResourcePermissions{}
		},
	}

	deps := depsWithPermissionHandler(t, mockSvc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 1, "admin")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp["data"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, data)
}

// ---------------------------------------------------------------------------
// Tests: GET /api/v1/me/permissions (GetUserPermissions)
// ---------------------------------------------------------------------------

func TestPermissionHandler_GetUserPermissions_Success(t *testing.T) {
	mockSvc := &mockRoleServiceForPermission{
		getUserPermissionsFn: func(_ context.Context, userID uint) (*service.UserPermissions, error) {
			return &service.UserPermissions{
				IsSuperAdmin:    false,
				TeamPermissions: map[int64][]string{10: {"main_item:create", "main_item:read"}},
			}, nil
		},
	}

	deps := depsWithPermissionHandler(t, mockSvc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, mockSvc.getUserPermissionsCalled)
	assert.Equal(t, uint(5), mockSvc.lastUserID)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, data["isSuperAdmin"])

	teamPerms, ok := data["teamPermissions"].(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, teamPerms, 1)
}

func TestPermissionHandler_GetUserPermissions_SuperAdmin(t *testing.T) {
	mockSvc := &mockRoleServiceForPermission{
		getUserPermissionsFn: func(_ context.Context, userID uint) (*service.UserPermissions, error) {
			return &service.UserPermissions{
				IsSuperAdmin:    true,
				TeamPermissions: map[int64][]string{},
			}, nil
		},
	}

	deps := depsWithPermissionHandler(t, mockSvc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 1, "admin")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, true, data["isSuperAdmin"])
}

func TestPermissionHandler_GetUserPermissions_ServiceError(t *testing.T) {
	mockSvc := &mockRoleServiceForPermission{
		getUserPermissionsFn: func(_ context.Context, _ uint) (*service.UserPermissions, error) {
			return nil, errors.New("database error")
		},
	}

	deps := depsWithPermissionHandler(t, mockSvc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 5, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPermissionHandler_GetUserPermissions_RequiresAuth(t *testing.T) {
	mockSvc := &mockRoleServiceForPermission{}

	deps := depsWithPermissionHandler(t, mockSvc)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/permissions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, mockSvc.getUserPermissionsCalled)
}

// ---------------------------------------------------------------------------
// Helper: depsWithPermissionHandler wires a mock RoleService into test deps.
// ---------------------------------------------------------------------------

func depsWithPermissionHandler(t *testing.T, svc *mockRoleServiceForPermission) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.Permission = NewPermissionHandler(svc)
	return deps
}

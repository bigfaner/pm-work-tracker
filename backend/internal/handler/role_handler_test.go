package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/model"
	appjwt "pm-work-tracker/backend/internal/pkg/jwt"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/service"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// rbacTestEnv sets up a full integration environment with DB, repos, service, and router
// for role and permission handler tests.
func rbacTestEnv(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	snowflake.Init(1)

	require.NoError(t, db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.RolePermission{},
		&model.Team{},
		&model.TeamMember{},
	))

	// Seed a superadmin user (ID=1) and a regular user (ID=2)
	require.NoError(t, db.Create(&model.User{
		Username: "rbacsuper", DisplayName: "RBAC Super", IsSuperAdmin: true,
	}).Error)
	require.NoError(t, db.Create(&model.User{
		Username: "rbacregular", DisplayName: "RBAC Regular", IsSuperAdmin: false,
	}).Error)

	// Seed preset roles
	require.NoError(t, db.Create(&model.Role{BaseModel: model.BaseModel{BizKey: 1}, Name: "superadmin", Description: "superadmin", IsPreset: true}).Error)
	require.NoError(t, db.Create(&model.Role{BaseModel: model.BaseModel{BizKey: 2}, Name: "pm", Description: "PM", IsPreset: true}).Error)
	require.NoError(t, db.Create(&model.Role{BaseModel: model.BaseModel{BizKey: 3}, Name: "member", Description: "Member", IsPreset: true}).Error)

	userRepo := gormrepo.NewGormUserRepo(db)
	roleRepo := gormrepo.NewGormRoleRepo(db)
	teamRepo := gormrepo.NewGormTeamRepo(db)
	roleSvc := service.NewRoleService(roleRepo, userRepo)

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "test-secret-that-is-at-least-32-bytes!!",
		},
		CORS: config.CORSConfig{
			Origins: []string{"http://localhost:3000"},
		},
		Server: config.ServerConfig{
			GinMode:  "test",
			BasePath: "/api",
		},
	}

	deps := &Dependencies{
		Config:     cfg,
		UserRepo:   userRepo,
		RoleRepo:   roleRepo,
		TeamRepo:   teamRepo,
		Auth:       NewAuthHandler(&rbacStubAuthService{}),
		Team:       NewTeamHandler(&StubTeamSvc{}, &StubRouterRepoUser{}),
		MainItem:   NewMainItemHandler(&StubMainItemSvc{}, &StubRouterRepoUser{}, &StubRouterRepoSubItem{}),
		SubItem:    NewSubItemHandler(&StubSubItemSvc{}, &StubMainItemSvc{}),
		Progress:   NewProgressHandler(&StubProgressSvc{}, &StubRouterRepoUser{}, &StubSubItemSvc{}),
		ItemPool:   NewItemPoolHandler(&StubItemPoolSvc{}, &StubRouterRepoUser{}, &StubRouterRepoMainItem{}),
		View:       NewViewHandler(&StubViewSvc{}),
		Report:     NewReportHandler(&StubReportSvc{}),
		Admin:      NewAdminHandler(&StubAdminSvc{}),
		Role:       NewRoleHandler(roleSvc),
		Permission: NewPermissionHandler(roleSvc),
	}
	r := SetupRouter(deps, nil)
	return r, db
}

// rbacStubAuthService is a test stub for AuthService.
type rbacStubAuthService struct{}

func (s *rbacStubAuthService) Login(_ context.Context, _, _ string) (string, *model.User, error) {
	return "", nil, fmt.Errorf("not implemented")
}
func (s *rbacStubAuthService) Logout(_ context.Context, _ string) error { return nil }
func (s *rbacStubAuthService) ParseToken(_ context.Context, token string) (*appjwt.Claims, error) {
	return appjwt.Verify(token, "test-secret-that-is-at-least-32-bytes!!")
}

// Ensure rbacStubAuthService has the right context import
var _ context.Context

// rbacSignSuperToken creates a JWT for the superadmin user (ID=1).
func rbacSignSuperToken(t *testing.T) string {
	t.Helper()
	token, err := appjwt.Sign(1, "superadmin", "test-secret-that-is-at-least-32-bytes!!")
	require.NoError(t, err)
	return token
}

// rbacSignRegularToken creates a JWT for the regular user (ID=2).
func rbacSignRegularToken(t *testing.T) string {
	t.Helper()
	token, err := appjwt.Sign(2, "member", "test-secret-that-is-at-least-32-bytes!!")
	require.NoError(t, err)
	return token
}

// ---- ListRoles Tests ----

func TestListRoles_RequiresSuperAdmin(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignRegularToken(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestListRoles_RequiresAuth(t *testing.T) {
	r, _ := rbacTestEnv(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListRoles_Success(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(0), body["code"])

	data := body["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	assert.GreaterOrEqual(t, len(items), 3) // at least 3 preset roles
}

// ---- CreateRole Tests ----

func TestCreateRole_Success(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	body := `{"name":"viewer","description":"read-only viewer","permissionCodes":["team:read","main_item:read"]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "viewer", data["roleName"])
	assert.Equal(t, false, data["isPreset"])
}

func TestCreateRole_ValidationError(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	// Missing name
	body := `{"description":"no name","permissionCodes":["team:read"]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRole_EmptyPermissionCodes(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	body := `{"name":"norole","description":"empty perms","permissionCodes":[]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRole_DuplicateName(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	body := `{"name":"pm","description":"duplicate","permissionCodes":["team:read"]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ERR_ROLE_NAME_EXISTS", resp["code"])
}

func TestCreateRole_InvalidPermissionCode(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	body := `{"name":"badrole","description":"bad perm","permissionCodes":["nonexistent:action"]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ERR_INVALID_PERMISSION_CODE", resp["code"])
}

// ---- GetRole Tests ----

func TestGetRole_Success(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles/2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "pm", data["roleName"])
	assert.Equal(t, true, data["isPreset"])
}

func TestGetRole_NotFound(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- UpdateRole Tests ----

func TestUpdateRole_Success(t *testing.T) {
	r, db := rbacTestEnv(t)

	// Create a custom role
	require.NoError(t, db.Create(&model.Role{BaseModel: model.BaseModel{BizKey: 100}, Name: "custom1", Description: "custom role", IsPreset: false}).Error)
	var customRole model.Role
	require.NoError(t, db.Where("name = ?", "custom1").First(&customRole).Error)

	token := rbacSignSuperToken(t)
	body := `{"description":"updated desc","permissionCodes":["team:read"]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%d", customRole.BizKey), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "updated desc", data["roleDesc"])
}

func TestUpdateRole_PresetSuperAdmin_Immutable(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	body := `{"description":"try to modify superadmin"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/1", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ERR_PRESET_ROLE_IMMUTABLE", resp["code"])
}

func TestUpdateRole_PresetPM_CannotChangeName(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	body := `{"name":"newname"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/2", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateRole_NotFound(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	body := `{"description":"does not exist"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/999", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- DeleteRole Tests ----

func TestDeleteRole_Success(t *testing.T) {
	r, db := rbacTestEnv(t)

	require.NoError(t, db.Create(&model.Role{BaseModel: model.BaseModel{BizKey: 101}, Name: "deleteme", Description: "will be deleted", IsPreset: false}).Error)
	var customRole model.Role
	require.NoError(t, db.Where("name = ?", "deleteme").First(&customRole).Error)

	token := rbacSignSuperToken(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/admin/roles/%d", customRole.BizKey), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteRole_PresetImmutable(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/roles/2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ERR_PRESET_ROLE_IMMUTABLE", resp["code"])
}

// ---- ListPermissionCodes Tests ----

func TestListPermissionCodes_Success(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 1)

	first := data[0].(map[string]interface{})
	assert.Contains(t, first, "resource")
	assert.Contains(t, first, "permissions")
}

func TestListPermissionCodes_RequiresSuperAdmin(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignRegularToken(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---- GetUserPermissions Tests ----

func TestGetUserPermissions_RequiresAuth(t *testing.T) {
	r, _ := rbacTestEnv(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/permissions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUserPermissions_Success_SuperAdmin(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, true, data["isSuperAdmin"])
}

func TestGetUserPermissions_Success_RegularUser(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignRegularToken(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/permissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, false, data["isSuperAdmin"])
	perms := data["teamPermissions"].(map[string]interface{})
	assert.Empty(t, perms)
}

// ---- Route Registration Tests ----

func TestRoleRoutes_AllRegistered(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignSuperToken(t)

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/admin/roles"},
		{"POST", "/api/v1/admin/roles"},
		{"GET", "/api/v1/admin/roles/1"},
		{"PUT", "/api/v1/admin/roles/1"},
		{"DELETE", "/api/v1/admin/roles/1"},
		{"GET", "/api/v1/admin/permissions"},
		{"GET", "/api/v1/me/permissions"},
	}

	for _, route := range routes {
		w := httptest.NewRecorder()
		var body *strings.Reader
		if route.method == "POST" || route.method == "PUT" {
			body = strings.NewReader(`{"name":"test","permissionCodes":["team:read"]}`)
		} else {
			body = strings.NewReader("")
		}
		req := httptest.NewRequest(route.method, route.path, body)
		req.Header.Set("Authorization", "Bearer "+token)
		if route.method == "POST" || route.method == "PUT" {
			req.Header.Set("Content-Type", "application/json")
		}
		r.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code,
			"route %s %s should be registered", route.method, route.path)
	}
}

func TestRoleRoutes_RegularUserForbidden(t *testing.T) {
	r, _ := rbacTestEnv(t)

	token := rbacSignRegularToken(t)

	adminRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/admin/roles"},
		{"POST", "/api/v1/admin/roles"},
		{"GET", "/api/v1/admin/roles/1"},
		{"PUT", "/api/v1/admin/roles/1"},
		{"DELETE", "/api/v1/admin/roles/1"},
		{"GET", "/api/v1/admin/permissions"},
	}

	for _, route := range adminRoutes {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(route.method, route.path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code,
			"route %s %s should be forbidden for regular user", route.method, route.path)
	}
}

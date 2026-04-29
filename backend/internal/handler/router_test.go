package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	appjwt "pm-work-tracker/backend/internal/pkg/jwt"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// testDeps creates a Dependencies struct wired for testing.
func testDeps(t testing.TB) (*Dependencies, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	teamRepo := gormrepo.NewGormTeamRepo(db)
	userRepo := gormrepo.NewGormUserRepo(db)
	roleRepo := gormrepo.NewGormRoleRepo(db)

	// Auto-migrate required tables so AuthMiddleware can query users
	err = db.AutoMigrate(&model.User{})
	require.NoError(t, err)

	// Auto-migrate RBAC tables so RequirePermission middleware can query roles
	err = db.AutoMigrate(&model.Team{}, &model.Role{}, &model.RolePermission{}, &model.TeamMember{})
	require.NoError(t, err)

	// Seed a PM role (ID=1) with a broad set of permissions for testing.
	// Most handler tests use mockTeamRepo which returns RoleKey=&1 (biz_key),
	// and TeamScopeMiddleware will call FindByBizKey(1) to resolve permissions.
	pmRole := model.Role{Name: "pm", Description: "Project Manager", IsPreset: true}
	require.NoError(t, db.Create(&pmRole).Error)
	pmRole.BizKey = int64(pmRole.ID)
	require.NoError(t, db.Save(&pmRole).Error)
	allPermCodes := []string{
		"team:create", "team:read", "team:update", "team:delete", "team:invite",
		"team:remove", "team:transfer", "main_item:create", "main_item:read",
		"main_item:update", "main_item:archive", "main_item:change_status",
		"sub_item:create", "sub_item:read", "sub_item:update", "sub_item:assign", "sub_item:change_status",
		"progress:create", "progress:read", "progress:update",
		"item_pool:submit", "item_pool:review",
		"view:weekly", "view:gantt", "view:table",
		"report:export",
		"user:list", "user:read", "user:assign_role",
		"role:read", "role:create", "role:update", "role:delete",
	}
	for _, code := range allPermCodes {
		require.NoError(t, db.Create(&model.RolePermission{RoleKey: pmRole.BizKey, PermissionCode: code}).Error)
	}

	// Seed a member role with standard member permissions.
	memberRole := model.Role{Name: "member", Description: "Team Member", IsPreset: true}
	require.NoError(t, db.Create(&memberRole).Error)
	memberRole.BizKey = int64(memberRole.ID)
	require.NoError(t, db.Save(&memberRole).Error)
	memberPermCodes := []string{
		"main_item:read", "sub_item:create", "sub_item:read", "sub_item:update",
		"sub_item:change_status", "progress:create", "progress:read",
		"item_pool:submit", "view:weekly", "view:table", "report:export",
	}
	for _, code := range memberPermCodes {
		require.NoError(t, db.Create(&model.RolePermission{RoleKey: memberRole.BizKey, PermissionCode: code}).Error)
	}

	// Seed test users so AuthMiddleware can load them.
	// ID=1: superadmin user
	// ID=2: regular member (used by most handler tests)
	// ID=5: regular member (used by item pool and other handler tests)
	require.NoError(t, db.Create(&model.User{
		BaseModel:    model.BaseModel{BizKey: 1},
		Username:     "admin",
		DisplayName:  "Admin User",
		IsSuperAdmin: true,
	}).Error)
	require.NoError(t, db.Create(&model.User{
		Username:     "testuser1",
		BaseModel:    model.BaseModel{BizKey: 2},
		DisplayName:  "Test User 1",
		IsSuperAdmin: false,
	}).Error)
	// Ensure we have user ID=5 by creating users until we reach it
	for i := 3; i <= 5; i++ {
		var count int64
		db.Model(&model.User{}).Count(&count)
		if count >= 5 {
			break
		}
		db.Create(&model.User{
			BaseModel:    model.BaseModel{BizKey: int64(i)},
			Username:     fmt.Sprintf("testuser%d", i),
			DisplayName:  fmt.Sprintf("Test User %d", i),
			IsSuperAdmin: false,
		})
	}

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

	return &Dependencies{
		Config:   cfg,
		TeamRepo: teamRepo,
		UserRepo: userRepo,
		RoleRepo: roleRepo,
		Auth:     NewAuthHandler(&stubAuthService{}),
		Team:     NewTeamHandler(&StubTeamSvc{}, &StubRouterRepoUser{}),
		MainItem: NewMainItemHandler(&StubMainItemSvc{}, &StubRouterRepoUser{}, &StubRouterRepoSubItem{}),
		SubItem:  NewSubItemHandler(&StubSubItemSvc{}, &StubMainItemSvc{}),
		Progress: NewProgressHandler(&StubProgressSvc{}, &StubRouterRepoUser{}, &StubSubItemSvc{}),
		ItemPool: NewItemPoolHandler(&StubItemPoolSvc{}, &StubRouterRepoUser{}, &StubRouterRepoMainItem{}),
		View:     NewViewHandler(&StubViewSvc{}),
		Report:   NewReportHandler(&StubReportSvc{}),
		Admin:      NewAdminHandler(&StubAdminSvc{}),
		Role:       NewRoleHandler(&StubRoleSvc{}),
		Permission: NewPermissionHandler(&StubRoleSvc{}),
	}, db
}

// signTestToken creates a valid JWT for testing.
// If userID is 1, the token is for the superadmin user; otherwise it's a regular user.
func signTestToken(t testing.TB, userID uint, username string) string {
	t.Helper()
	token, err := appjwt.Sign(userID, username, "test-secret-that-is-at-least-32-bytes!!")
	require.NoError(t, err)
	return token
}

func TestHealthCheck(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
}

func TestHealthCheck_NoAuthRequired(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	// Request without Authorization header should still succeed
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthRoutes_LoginReturns400OnEmptyBody(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthRoutes_LogoutRequiresAuth(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	// Without auth
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// With auth — logout is implemented, returns 200
	token := signTestToken(t, 2, "testuser1")
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeamRoutes_RequireAuthAndTeamScope(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	// Without auth — should get 401
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTeamRoutes_WithAuthReturns500(t *testing.T) {
	deps, _ := testDeps(t)
	// Inject a mock that always succeeds
	deps.TeamRepo = &mockTeamRepo{}
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 2, "testuser1")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTeamListCreateRoutes_RequireAuth(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	// POST /api/v1/teams without auth
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// GET /api/v1/teams without auth
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTeamListCreateRoutes_WithSuperAdminAuth(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 1, "admin")

	// POST /api/v1/teams with superadmin auth — stub service returns error -> 500
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", strings.NewReader(`{"name":"test","code":"TEST"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// GET /api/v1/teams with superadmin auth — stub service returns error -> 500
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdminRoutes_RequireSuperAdmin(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	// Regular member should get 403
	token := signTestToken(t, 2, "testuser1")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// SuperAdmin — stub service returns error -> 500
	superToken := signTestToken(t, 1, "admin")
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+superToken)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdminRoutes_RequireAuth(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCORS_HeadersSet(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	r.ServeHTTP(w, req)

	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_RejectUnknownOrigin(t *testing.T) {
	deps, _ := testDeps(t)
	deps.Config.CORS.Origins = []string{"http://localhost:3000"}
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://evil.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	r.ServeHTTP(w, req)

	// Origin should not be echoed back for non-allowed origins
	assert.NotEqual(t, "http://evil.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_WildcardWhenNoOriginsConfigured(t *testing.T) {
	deps, _ := testDeps(t)
	deps.Config.CORS.Origins = nil
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	r.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestRateLimit_Login(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	// Send more than 10 requests to login — the 11th+ should be rate limited
	got429 := false
	for i := 0; i < 15; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			got429 = true
			break
		}
	}
	assert.True(t, got429, "expected at least one 429 response after many login attempts")
}

func TestRateLimit_OnlyAppliesToLogin(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	// Other endpoints should not be rate limited
	for i := 0; i < 15; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestUnknownRoute_404(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTeamScopedRoutes_AllRegistered(t *testing.T) {
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{}
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 1, "admin")

	routes := []struct {
		method string
		path   string
	}{
		// Main items
		{"POST", "/api/v1/teams/1/main-items"},
		{"GET", "/api/v1/teams/1/main-items"},
		{"GET", "/api/v1/teams/1/main-items/1"},
		{"PUT", "/api/v1/teams/1/main-items/1"},
		{"PUT", "/api/v1/teams/1/main-items/1/status"},
		{"GET", "/api/v1/teams/1/main-items/1/available-transitions"},
		{"POST", "/api/v1/teams/1/main-items/1/archive"},
		// Sub items
		{"POST", "/api/v1/teams/1/main-items/1/sub-items"},
		{"GET", "/api/v1/teams/1/main-items/1/sub-items"},
		{"GET", "/api/v1/teams/1/sub-items/1"},
		{"PUT", "/api/v1/teams/1/sub-items/1"},
		{"PUT", "/api/v1/teams/1/sub-items/1/status"},
		{"PUT", "/api/v1/teams/1/sub-items/1/assignee"},
		// Progress
		{"POST", "/api/v1/teams/1/sub-items/1/progress"},
		{"GET", "/api/v1/teams/1/sub-items/1/progress"},
		{"PATCH", "/api/v1/teams/1/progress/1/completion"},
		// Item pool
		{"POST", "/api/v1/teams/1/item-pool"},
		{"GET", "/api/v1/teams/1/item-pool"},
		{"GET", "/api/v1/teams/1/item-pool/1"},
		{"POST", "/api/v1/teams/1/item-pool/1/assign"},
		{"POST", "/api/v1/teams/1/item-pool/1/reject"},
		// Views
		{"GET", "/api/v1/teams/1/views/weekly"},
		{"GET", "/api/v1/teams/1/views/gantt"},
		{"GET", "/api/v1/teams/1/views/table"},
		{"GET", "/api/v1/teams/1/views/table/export"},
		// Reports
		{"GET", "/api/v1/teams/1/reports/weekly/preview"},
		{"GET", "/api/v1/teams/1/reports/weekly/export"},
		// Team info
		{"GET", "/api/v1/teams/1"},
		{"PUT", "/api/v1/teams/1"},
		{"DELETE", "/api/v1/teams/1"},
		// Members
		{"GET", "/api/v1/teams/1/members"},
		{"GET", "/api/v1/teams/1/search-users"},
		{"POST", "/api/v1/teams/1/members"},
		{"DELETE", "/api/v1/teams/1/members/2"},
		{"PUT", "/api/v1/teams/1/pm"},
	}

	for _, route := range routes {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(route.method, route.path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		// Should not be 404 (route not found) — should be either 501 or another error
		assert.NotEqual(t, http.StatusNotFound, w.Code,
			"route %s %s should be registered", route.method, route.path)
	}
}

func TestAdminRoutes_AllRegistered(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)

	superToken := signTestToken(t, 1, "admin")

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/admin/users"},
		{"GET", "/api/v1/admin/teams"},
	}

	for _, route := range routes {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(route.method, route.path, nil)
		req.Header.Set("Authorization", "Bearer "+superToken)
		r.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code,
			"route %s %s should be registered", route.method, route.path)
	}
}

// TestRequirePermission_AdminRoutes verifies that admin routes gated by new
// permission codes return 403 when the authenticated user lacks the required
// permission. Uses a regular member user (ID=2) which has no admin-level
// permissions in the seed data.

func TestRequirePermission_RoleRead(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)
	token := signTestToken(t, 2, "testuser1")

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/admin/roles"},
		{"GET", "/api/v1/admin/roles/1"},
		{"GET", "/api/v1/admin/permissions"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(route.method, route.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "%s %s", route.method, route.path)
			assert.Contains(t, w.Body.String(), "ERR_FORBIDDEN", "%s %s", route.method, route.path)
		})
	}
}

func TestRequirePermission_RoleCreate(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)
	token := signTestToken(t, 2, "testuser1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_FORBIDDEN")
}

func TestRequirePermission_RoleUpdate(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)
	token := signTestToken(t, 2, "testuser1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_FORBIDDEN")
}

func TestRequirePermission_RoleDelete(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)
	token := signTestToken(t, 2, "testuser1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/roles/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_FORBIDDEN")
}

func TestRequirePermission_UserList(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)
	token := signTestToken(t, 2, "testuser1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_FORBIDDEN")
}

func TestRequirePermission_UserRead(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)
	token := signTestToken(t, 2, "testuser1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_FORBIDDEN")
}

func TestRequirePermission_UserAssignRole(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps, nil)
	token := signTestToken(t, 2, "testuser1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_FORBIDDEN")
}

// mockTeamRepo is a test double that satisfies repository.TeamRepo.
type mockTeamRepo struct {
	member *model.TeamMember
}

func (m *mockTeamRepo) Create(_ context.Context, _ *model.Team) error { return nil }
func (m *mockTeamRepo) FindByID(_ context.Context, _ uint) (*model.Team, error) {
	return &model.Team{}, nil
}
func (m *mockTeamRepo) List(_ context.Context) ([]*model.Team, error) {
	return nil, nil
}

func (m *mockTeamRepo) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.Team, int64, error) {
	return nil, 0, nil
}
func (m *mockTeamRepo) Update(_ context.Context, _ *model.Team) error { return nil }
func (m *mockTeamRepo) SoftDelete(_ context.Context, _ uint) error    { return nil }
func (m *mockTeamRepo) FindByBizKey(_ context.Context, bizKey int64) (*model.Team, error) {
	return &model.Team{BaseModel: model.BaseModel{ID: uint(bizKey)}}, nil
}
func (m *mockTeamRepo) AddMember(_ context.Context, _ *model.TeamMember) error {
	return nil
}
func (m *mockTeamRepo) RemoveMember(_ context.Context, _, _ int64) error { return nil }
func (m *mockTeamRepo) FindMember(_ context.Context, _, _ int64) (*model.TeamMember, error) {
	if m.member != nil {
		return m.member, nil
	}
	// Default: member with RoleID=1 (PM role seeded in testDeps)
	pmRoleID := uint(1)
	return &model.TeamMember{ RoleKey: func() *int64 { v := int64(pmRoleID); return &v }()}, nil
}
func (m *mockTeamRepo) ListMembers(_ context.Context, _ int64) ([]*dto.TeamMemberDTO, error) {
	return nil, nil
}
func (m *mockTeamRepo) CountMembers(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockTeamRepo) UpdateMember(_ context.Context, _ *model.TeamMember) error {
	return nil
}
func (m *mockTeamRepo) ListAllTeams(_ context.Context) ([]*dto.AdminTeamDTO, error) {
	return nil, nil
}

func (m *mockTeamRepo) FindPMMembers(_ context.Context, _ []int64) (map[int64]string, error) {
	return map[int64]string{}, nil
}

func (m *mockTeamRepo) FindTeamsByUserIDs(_ context.Context, _ []uint) (map[uint][]dto.TeamSummary, error) {
	return map[uint][]dto.TeamSummary{}, nil
}

func (m *mockTeamRepo) FindTeamsByUserBizKeys(_ context.Context, _ []int64) (map[int64][]dto.TeamSummary, error) {
	return map[int64][]dto.TeamSummary{}, nil
}

// stubAuthService is a minimal stub for service.AuthService used by testDeps.
// It returns ErrUnauthorized on Login (no real auth logic needed for router tests).
type stubAuthService struct{}

func (s *stubAuthService) Login(_ context.Context, _, _ string) (string, *model.User, error) {
	return "", nil, apperrors.ErrUnauthorized
}
func (s *stubAuthService) Logout(_ context.Context, _ string) error { return nil }
func (s *stubAuthService) ParseToken(_ context.Context, token string) (*appjwt.Claims, error) {
	return appjwt.Verify(token, "test-secret-that-is-at-least-32-bytes!!")
}

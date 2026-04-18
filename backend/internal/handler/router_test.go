package handler

import (
	"context"
	"encoding/json"
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
func testDeps(t *testing.T) (*Dependencies, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	teamRepo := gormrepo.NewGormTeamRepo(db)

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "test-secret-that-is-at-least-32-bytes!!",
		},
		CORS: config.CORSConfig{
			Origins: []string{"http://localhost:3000"},
		},
		Server: config.ServerConfig{
			GinMode: "test",
		},
	}

	return &Dependencies{
		Config:   cfg,
		TeamRepo: teamRepo,
		Auth:     NewAuthHandler(&stubAuthService{}),
		Team:     NewTeamHandler(),
		MainItem: NewMainItemHandler(),
		SubItem:  NewSubItemHandler(),
		Progress: NewProgressHandler(),
		ItemPool: NewItemPoolHandler(),
		View:     NewViewHandler(),
		Report:   NewReportHandler(),
		Admin:    NewAdminHandler(),
	}, db
}

// signTestToken creates a valid JWT for testing.
func signTestToken(t *testing.T, userID uint, role string) string {
	t.Helper()
	token, err := appjwt.Sign(userID, role, "test-secret-that-is-at-least-32-bytes!!")
	require.NoError(t, err)
	return token
}

func TestHealthCheck(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

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
	r := SetupRouter(deps)

	// Request without Authorization header should still succeed
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthRoutes_LoginReturns400OnEmptyBody(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthRoutes_LogoutRequiresAuth(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

	// Without auth
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// With auth — logout is implemented, returns 200
	token := signTestToken(t, 1, "member")
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeamRoutes_RequireAuthAndTeamScope(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

	// Without auth — should get 401
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTeamRoutes_WithAuthReturns501(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

	// Create a user and team member so TeamScopeMiddleware passes
	db, _ := testDeps(t) // we need a second db, but let's just use the deps' teamRepo
	_ = db
	// Instead, inject a mock that always succeeds
	deps.TeamRepo = &mockTeamRepo{}
	r = SetupRouter(deps)

	token := signTestToken(t, 1, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestTeamListCreateRoutes_RequireAuth(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

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

func TestTeamListCreateRoutes_WithAuth(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")

	// POST /api/v1/teams with auth
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotImplemented, w.Code)

	// GET /api/v1/teams with auth
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestAdminRoutes_RequireSuperAdmin(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

	// Regular member should get 403
	token := signTestToken(t, 1, "member")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// SuperAdmin should get 501 (not implemented)
	superToken := signTestToken(t, 1, "superadmin")
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+superToken)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestAdminRoutes_RequireAuth(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCORS_HeadersSet(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

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
	r := SetupRouter(deps)

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
	r := SetupRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	r.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestRateLimit_Login(t *testing.T) {
	deps, _ := testDeps(t)
	r := SetupRouter(deps)

	// Send more than 10 requests to login — the 11th+ should be rate limited
	// Note: the rate limiter allows bursts up to `limit`, so we need to exceed that.
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
	r := SetupRouter(deps)

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
	r := SetupRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTeamScopedRoutes_AllRegistered(t *testing.T) {
	deps, _ := testDeps(t)
	deps.TeamRepo = &mockTeamRepo{}
	r := SetupRouter(deps)

	token := signTestToken(t, 1, "member")

	routes := []struct {
		method string
		path   string
	}{
		// Main items
		{"POST", "/api/v1/teams/1/main-items"},
		{"GET", "/api/v1/teams/1/main-items"},
		{"GET", "/api/v1/teams/1/main-items/1"},
		{"PUT", "/api/v1/teams/1/main-items/1"},
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
	r := SetupRouter(deps)

	superToken := signTestToken(t, 1, "superadmin")

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/admin/users"},
		{"PUT", "/api/v1/admin/users/1/can-create-team"},
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
func (m *mockTeamRepo) Update(_ context.Context, _ *model.Team) error { return nil }
func (m *mockTeamRepo) Delete(_ context.Context, _ uint) error        { return nil }
func (m *mockTeamRepo) AddMember(_ context.Context, _ *model.TeamMember) error {
	return nil
}
func (m *mockTeamRepo) RemoveMember(_ context.Context, _, _ uint) error { return nil }
func (m *mockTeamRepo) FindMember(_ context.Context, _, _ uint) (*model.TeamMember, error) {
	if m.member != nil {
		return m.member, nil
	}
	return &model.TeamMember{Role: "member"}, nil
}
func (m *mockTeamRepo) ListMembers(_ context.Context, _ uint) ([]*dto.TeamMemberDTO, error) {
	return nil, nil
}
func (m *mockTeamRepo) UpdateMember(_ context.Context, _ *model.TeamMember) error {
	return nil
}
func (m *mockTeamRepo) ListAllTeams(_ context.Context) ([]*dto.AdminTeamDTO, error) {
	return nil, nil
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

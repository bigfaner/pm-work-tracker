package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appjwt "pm-work-tracker/backend/internal/pkg/jwt"
)

// ========== Auth Flow Tests ==========

func TestAuthFlow_LoginWithCorrectCredentials_Returns200(t *testing.T) {
	r, _ := setupTestRouter(t)

	body := `{"username":"userA","password":"passwordA"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, data["token"])

	user, ok := data["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "userA", user["username"])
}

func TestAuthFlow_LoginWithWrongPassword_Returns401(t *testing.T) {
	r, _ := setupTestRouter(t)

	body := `{"username":"userA","password":"wrongpassword"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "UNAUTHORIZED", resp["code"])
}

func TestAuthFlow_TokenOnProtectedRoute_Returns200(t *testing.T) {
	r, data := setupTestRouter(t)

	// Login to get a real token
	token := loginAs(t, r, "userA", "passwordA")

	// Use the token on a protected route: GET /api/v1/teams/:teamAId/main-items
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// Should be 200 (empty list) not 401
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthFlow_ExpiredToken_Returns401(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Sign a token that is already expired
	claims := &appjwt.Claims{
		UserID:   999,
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := signTokenWithClaims(t, claims)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1/main-items", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthFlow_NoToken_Returns401(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/1/main-items", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ========== Team Isolation Tests ==========

func TestTeamIsolation_UserACannotAccessTeamB_Returns403(t *testing.T) {
	r, data := setupTestRouter(t)

	token := loginAs(t, r, "userA", "passwordA")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamBBizKey), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_TEAM_MEMBER", resp["code"])
}

func TestTeamIsolation_UserACanAccessTeamA_Returns200(t *testing.T) {
	r, data := setupTestRouter(t)

	token := loginAs(t, r, "userA", "passwordA")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeamIsolation_UserBCannotAccessTeamA_Returns403(t *testing.T) {
	r, data := setupTestRouter(t)

	token := loginAs(t, r, "userB", "passwordB")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_TEAM_MEMBER", resp["code"])
}

// ========== SuperAdmin Bypass Tests ==========

func TestSuperAdmin_CanAccessTeamA(t *testing.T) {
	r, data := setupTestRouter(t)

	token := loginAs(t, r, "superadmin", "adminPass")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamABizKey), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSuperAdmin_CanAccessTeamB(t *testing.T) {
	r, data := setupTestRouter(t)

	token := loginAs(t, r, "superadmin", "adminPass")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/teams/%d/main-items", data.teamBBizKey), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSuperAdmin_CanAccessAdminRoutes(t *testing.T) {
	r, _ := setupTestRouter(t)

	token := loginAs(t, r, "superadmin", "adminPass")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegularUser_CannotAccessAdminRoutes(t *testing.T) {
	r, _ := setupTestRouter(t)

	token := loginAs(t, r, "userA", "passwordA")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Logout Test ==========

func TestLogout_Returns200(t *testing.T) {
	r, _ := setupTestRouter(t)

	token := loginAs(t, r, "userA", "passwordA")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["code"])
}

func TestLogout_WithoutAuth_Returns401(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

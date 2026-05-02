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

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	appjwt "pm-work-tracker/backend/internal/pkg/jwt"
)

// mockAuthService implements service.AuthService for testing.
type mockAuthService struct {
	loginResult struct {
		token string
		user  *model.User
		err   error
	}
	logoutErr error
	parseErr  error
}

func (m *mockAuthService) Login(_ context.Context, _, _ string) (string, *model.User, error) {
	return m.loginResult.token, m.loginResult.user, m.loginResult.err
}

func (m *mockAuthService) Logout(_ context.Context, _ string) error {
	return m.logoutErr
}

func (m *mockAuthService) ParseToken(_ context.Context, token string) (*appjwt.Claims, error) {
	return appjwt.Verify(token, "test-secret-that-is-at-least-32-bytes!!")
}

// --- Tests ---

func TestLogin_Success(t *testing.T) {
	authSvc := &mockAuthService{}
	authSvc.loginResult.token = "jwt-token-123"
	authSvc.loginResult.user = &model.User{
		Username:     "alice",
		DisplayName:  "Alice",
		IsSuperAdmin: true,
	}
	// GORM Model has ID field
	authSvc.loginResult.user.ID = 1
	authSvc.loginResult.user.BizKey = 1

	deps := depsWithAuthSvc(t, authSvc)
	r := SetupRouter(deps, nil)

	body := `{"username":"alice","password":"secret123"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Top-level envelope
	assert.Equal(t, float64(0), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "data should be an object")
	assert.Equal(t, "jwt-token-123", data["token"])

	user, ok := data["user"].(map[string]interface{})
	require.True(t, ok, "user should be an object")

	assert.Equal(t, "1", user["bizKey"])
	assert.Equal(t, "alice", user["username"])
	assert.Equal(t, "Alice", user["displayName"])
	assert.Equal(t, true, user["isSuperAdmin"])
}

func TestLogin_MissingUsername_Returns400(t *testing.T) {
	authSvc := &mockAuthService{}
	deps := depsWithAuthSvc(t, authSvc)
	r := SetupRouter(deps, nil)

	body := `{"password":"secret123"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", resp["code"])
}

func TestLogin_MissingPassword_Returns400(t *testing.T) {
	authSvc := &mockAuthService{}
	deps := depsWithAuthSvc(t, authSvc)
	r := SetupRouter(deps, nil)

	body := `{"username":"alice"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", resp["code"])
}

func TestLogin_EmptyBody_Returns400(t *testing.T) {
	authSvc := &mockAuthService{}
	deps := depsWithAuthSvc(t, authSvc)
	r := SetupRouter(deps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_WrongCredentials_Returns401(t *testing.T) {
	authSvc := &mockAuthService{}
	authSvc.loginResult.err = apperrors.ErrUnauthorized

	deps := depsWithAuthSvc(t, authSvc)
	r := SetupRouter(deps, nil)

	body := `{"username":"alice","password":"wrong"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", resp["code"])
}

func TestLogin_PasswordNeverInResponse(t *testing.T) {
	authSvc := &mockAuthService{}
	authSvc.loginResult.token = "tok"
	authSvc.loginResult.user = &model.User{
		Username:    "bob",
		DisplayName: "Bob",
	}
	authSvc.loginResult.user.ID = 5

	deps := depsWithAuthSvc(t, authSvc)
	r := SetupRouter(deps, nil)

	body := `{"username":"bob","password":"secret"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify no password appears in the response body
	respStr := w.Body.String()
	assert.NotContains(t, respStr, "password")
	assert.NotContains(t, respStr, "passwordHash")
}

func TestLogout_Success(t *testing.T) {
	authSvc := &mockAuthService{}
	deps := depsWithAuthSvc(t, authSvc)
	r := SetupRouter(deps, nil)

	token := signTestToken(t, 1, "testuser")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
	assert.Nil(t, resp["data"])
}

// depsWithAuthSvc creates test deps with a mock AuthService wired in.
func depsWithAuthSvc(t *testing.T, authSvc *mockAuthService) *Dependencies {
	t.Helper()
	deps, _ := testDeps(t)
	deps.Auth = NewAuthHandler(authSvc)
	return deps
}

// Ensure dto types compile (compile-time check).
var _ dto.LoginReq
var _ dto.LoginResp

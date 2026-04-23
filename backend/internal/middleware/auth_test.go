package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pm-work-tracker/backend/internal/model"
	jwtpkg "pm-work-tracker/backend/internal/pkg/jwt"
	"pm-work-tracker/backend/internal/repository"
)

const testSecret = "test-secret-that-is-at-least-32b"

func init() {
	gin.SetMode(gin.TestMode)
}

// mockUserRepo is a testify mock for UserRepo.
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) List(ctx context.Context) ([]*model.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) SearchAvailable(_ context.Context, _ uint, _ string, _ int) ([]*model.User, error) {
	return nil, nil
}

var _ repository.UserRepo = (*mockUserRepo)(nil)

// setupAuthRouter creates a test router with AuthMiddleware and a dummy handler
// that captures the userID and isSuperAdmin from context.
func setupAuthRouter(jwtSecret string, userRepo repository.UserRepo) (*gin.Engine, *capturedAuthContext) {
	r := gin.New()
	cc := &capturedAuthContext{}
	r.Use(AuthMiddleware(jwtSecret, userRepo))
	r.GET("/test", func(c *gin.Context) {
		cc.userID = GetUserID(c)
		cc.isSuperAdmin = IsSuperAdmin(c)
		c.Status(http.StatusOK)
	})
	return r, cc
}

type capturedAuthContext struct {
	userID       uint
	isSuperAdmin bool
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	repo := new(mockUserRepo)
	r, _ := setupAuthRouter(testSecret, repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	repo := new(mockUserRepo)
	r, _ := setupAuthRouter(testSecret, repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	repo := new(mockUserRepo)
	r, _ := setupAuthRouter(testSecret, repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	repo := new(mockUserRepo)
	r, _ := setupAuthRouter(testSecret, repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	tokenStr, err := jwtpkg.Sign(1, "testuser", "different-secret-that-is-also-32")
	assert.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken_SetsContext(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByID", mock.Anything, uint(42)).Return(&model.User{
		Username:     "testuser",
		IsSuperAdmin: false,
	}, nil)

	r, cc := setupAuthRouter(testSecret, repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	tokenStr, err := jwtpkg.Sign(42, "testuser", testSecret)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(42), cc.userID)
	assert.False(t, cc.isSuperAdmin)
}

func TestAuthMiddleware_SuperAdmin_SetsIsSuperAdmin(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByID", mock.Anything, uint(1)).Return(&model.User{
		Username:     "admin",
		IsSuperAdmin: true,
	}, nil)

	r, cc := setupAuthRouter(testSecret, repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	tokenStr, err := jwtpkg.Sign(1, "admin", testSecret)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(1), cc.userID)
	assert.True(t, cc.isSuperAdmin)
}

func TestAuthMiddleware_UserNotFound_Returns401(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByID", mock.Anything, uint(99)).Return(nil, fmt.Errorf("not found"))

	r, _ := setupAuthRouter(testSecret, repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	tokenStr, err := jwtpkg.Sign(99, "testuser", testSecret)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUserID_NoValue(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, uint(0), GetUserID(c))
}

func TestIsSuperAdmin_NoValue(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.False(t, IsSuperAdmin(c))
}

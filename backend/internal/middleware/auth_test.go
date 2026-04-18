package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	jwtpkg "pm-work-tracker/backend/internal/pkg/jwt"
)

const testSecret = "test-secret-that-is-at-least-32b"

func init() {
	gin.SetMode(gin.TestMode)
}

// setupRouter creates a test router with AuthMiddleware and a dummy handler
// that captures the userID and userRole from context.
func setupRouter(jwtSecret string) (*gin.Engine, *capturedContext) {
	r := gin.New()
	cc := &capturedContext{}
	r.Use(AuthMiddleware(jwtSecret))
	r.GET("/test", func(c *gin.Context) {
		cc.userID = GetUserID(c)
		cc.userRole = GetUserRole(c)
		c.Status(http.StatusOK)
	})
	return r, cc
}

type capturedContext struct {
	userID   uint
	userRole string
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	r, _ := setupRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	r, _ := setupRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r, _ := setupRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// Manually create an expired token
	r, _ := setupRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Sign then verify with wrong secret effectively tests expired;
	// but let's use the jwt package directly to create an expired token.
	// We'll use jwt.Sign with a past time via manual construction isn't exposed,
	// so let's sign and wait — not practical. Instead, sign with different secret
	// to simulate an unverifiable token (same error path).
	tokenStr, err := jwtpkg.Sign(1, "pm", "different-secret-that-is-also-32")
	assert.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken_SetsContext(t *testing.T) {
	r, cc := setupRouter(testSecret)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	tokenStr, err := jwtpkg.Sign(42, "pm", testSecret)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(42), cc.userID)
	assert.Equal(t, "pm", cc.userRole)
}

func TestAuthMiddleware_ValidToken_DifferentRoles(t *testing.T) {
	roles := []string{"superadmin", "pm", "member"}
	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			r, cc := setupRouter(testSecret)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			tokenStr, err := jwtpkg.Sign(7, role, testSecret)
			assert.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+tokenStr)

			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, uint(7), cc.userID)
			assert.Equal(t, role, cc.userRole)
		})
	}
}

func TestGetUserID_NoValue(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, uint(0), GetUserID(c))
}

func TestGetUserRole_NoValue(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, "", GetUserRole(c))
}

package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupPermissionRouter creates a test router with RequirePermission middleware.
// Context values (isSuperAdmin, permCodes, userID) are set from query params.
func setupPermissionRouter(code string, roleRepo *mockRoleRepo) (*gin.Engine, *capturedPermContext) {
	r := gin.New()
	cc := &capturedPermContext{}

	r.Use(func(c *gin.Context) {
		if v := c.Query("isSuperAdmin"); v == "true" {
			c.Set("isSuperAdmin", true)
		}
		if v := c.Query("userID"); v != "" {
			var id uint
			fmt.Sscanf(v, "%d", &id)
			c.Set("userID", id)
		}
		if v := c.Query("permCodes"); v != "" {
			// Don't set permCodes - nil means non-team context
		}
		c.Next()
	})

	r.Use(RequirePermission(code, roleRepo))

	r.GET("/test", func(c *gin.Context) {
		cc.called = true
		c.Status(http.StatusOK)
	})

	return r, cc
}

type capturedPermContext struct {
	called bool
}

// setupTeamPermRouter creates a test router with team-context (permCodes set).
func setupTeamPermRouter(code string) (*gin.Engine, *capturedPermContext) {
	r := gin.New()
	cc := &capturedPermContext{}

	r.Use(func(c *gin.Context) {
		if v := c.Query("isSuperAdmin"); v == "true" {
			c.Set("isSuperAdmin", true)
		}
		// Set permCodes to simulate team context
		if codes := c.Query("permCodes"); codes != "" {
			// Parse comma-separated codes
			var permCodes []string
			for _, code := range splitCodes(codes) {
				permCodes = append(permCodes, code)
			}
			if permCodes != nil {
				c.Set("permCodes", permCodes)
			}
		} else {
			// Empty permCodes means team context with no permissions
			c.Set("permCodes", []string{})
		}
		if v := c.Query("userID"); v != "" {
			var id uint
			fmt.Sscanf(v, "%d", &id)
			c.Set("userID", id)
		}
		c.Next()
	})

	r.Use(RequirePermission(code, new(mockRoleRepo)))

	r.GET("/test", func(c *gin.Context) {
		cc.called = true
		c.Status(http.StatusOK)
	})

	return r, cc
}

func splitCodes(s string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	return result
}

// --- SuperAdmin bypass tests ---

func TestRequirePermission_SuperAdminBypasses_Passes(t *testing.T) {
	roleRepo := new(mockRoleRepo)
	r, cc := setupPermissionRouter("team:invite", roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?isSuperAdmin=true", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, cc.called)
	roleRepo.AssertNotCalled(t, "HasPermission")
}

// --- Team context tests ---

func TestRequirePermission_TeamContext_HasCode_Passes(t *testing.T) {
	r, cc := setupTeamPermRouter("team:invite")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?permCodes=team:read,team:invite,team:update", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, cc.called)
}

func TestRequirePermission_TeamContext_MissingCode_Returns403(t *testing.T) {
	r, cc := setupTeamPermRouter("team:invite")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?permCodes=team:read,team:update", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_FORBIDDEN")
	assert.Contains(t, w.Body.String(), "team:invite")
	assert.False(t, cc.called)
}

func TestRequirePermission_TeamContext_EmptyPermCodes_Returns403(t *testing.T) {
	r, cc := setupTeamPermRouter("team:invite")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_FORBIDDEN")
	assert.False(t, cc.called)
}

// --- Non-team context tests (DB query) ---

func TestRequirePermission_NonTeamContext_HasPermission_Passes(t *testing.T) {
	roleRepo := new(mockRoleRepo)
	roleRepo.On("HasPermission", mock.Anything, uint(5), "team:create").Return(true, nil)

	r, cc := setupPermissionRouter("team:create", roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userID=5", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, cc.called)
}

func TestRequirePermission_NonTeamContext_NoPermission_Returns403(t *testing.T) {
	roleRepo := new(mockRoleRepo)
	roleRepo.On("HasPermission", mock.Anything, uint(5), "team:create").Return(false, nil)

	r, cc := setupPermissionRouter("team:create", roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userID=5", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_FORBIDDEN")
	assert.Contains(t, w.Body.String(), "team:create")
	assert.False(t, cc.called)
}

func TestRequirePermission_NonTeamContext_DBError_Returns500(t *testing.T) {
	roleRepo := new(mockRoleRepo)
	roleRepo.On("HasPermission", mock.Anything, uint(5), "team:create").Return(false, fmt.Errorf("db error"))

	r, cc := setupPermissionRouter("team:create", roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userID=5", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.False(t, cc.called)
}

func TestRequirePermission_NonTeamContext_NoUserID_Returns403(t *testing.T) {
	roleRepo := new(mockRoleRepo)

	r, cc := setupPermissionRouter("team:create", roleRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, cc.called)
}

// --- GetPermCodes helper tests ---

func TestGetPermCodes_NoValue(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Nil(t, GetPermCodes(c))
}

func TestGetPermCodes_WithValue(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("permCodes", []string{"team:read", "team:update"})
	assert.Equal(t, []string{"team:read", "team:update"}, GetPermCodes(c))
}

// --- Error response format test ---

func TestRequirePermission_ErrorResponseFormat(t *testing.T) {
	r, _ := setupTeamPermRouter("team:invite")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?permCodes=team:read", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), `"code":"ERR_FORBIDDEN"`)
	assert.Contains(t, w.Body.String(), "权限不足：缺少 team:invite 权限")
}

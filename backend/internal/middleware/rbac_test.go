package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- RequireRole tests ---

func setupRBACRouter(requiredRoles ...string) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		role := c.Query("userRole")
		if role != "" {
			c.Set("userRole", role)
		}
		c.Next()
	})
	r.Use(RequireRole(requiredRoles...))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestRequireRole_MemberCallingPMRoute_Returns403(t *testing.T) {
	r := setupRBACRouter("pm", "superadmin")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userRole=member", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "FORBIDDEN")
}

func TestRequireRole_PMCallingPMRoute_Passes(t *testing.T) {
	r := setupRBACRouter("pm", "superadmin")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userRole=pm", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_SuperAdminNotInPMOnlyList_Returns403(t *testing.T) {
	// RequireRole is a strict membership check — no hierarchy bypass.
	// SuperAdmin is only special in RequireTeamRole.
	r := setupRBACRouter("pm")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userRole=superadmin", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_SuperAdminOnAdminRoute_Passes(t *testing.T) {
	// Admin routes use RequireRole("superadmin") — the actual use case.
	r := setupRBACRouter("superadmin")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userRole=superadmin", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_NoRoleInContext_Returns403(t *testing.T) {
	r := setupRBACRouter("pm")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_SingleRoleMatch(t *testing.T) {
	r := setupRBACRouter("superadmin")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userRole=superadmin", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- RequireTeamRole tests ---

func setupTeamRBACRouter(requiredRoles ...string) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		userRole := c.Query("userRole")
		teamRole := c.Query("callerTeamRole")
		if userRole != "" {
			c.Set("userRole", userRole)
		}
		if teamRole != "" {
			c.Set("callerTeamRole", teamRole)
		}
		c.Next()
	})
	r.Use(RequireTeamRole(requiredRoles...))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestRequireTeamRole_MemberCallingPMRoute_Returns403(t *testing.T) {
	r := setupTeamRBACRouter("pm")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userRole=member&callerTeamRole=member", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "FORBIDDEN")
}

func TestRequireTeamRole_PMCallingPMRoute_Passes(t *testing.T) {
	r := setupTeamRBACRouter("pm")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userRole=pm&callerTeamRole=pm", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireTeamRole_SuperAdminBypassesTeamRole_Passes(t *testing.T) {
	// SuperAdmin passes RequireTeamRole regardless of callerTeamRole
	r := setupTeamRBACRouter("pm")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userRole=superadmin&callerTeamRole=member", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireTeamRole_SuperAdminWithNoTeamRole_Passes(t *testing.T) {
	// SuperAdmin passes even without callerTeamRole in context
	r := setupTeamRBACRouter("pm")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userRole=superadmin", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireTeamRole_NoContext_Returns403(t *testing.T) {
	r := setupTeamRBACRouter("pm")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireTeamRole_MemberRoleNotInAllowedList(t *testing.T) {
	r := setupTeamRBACRouter("pm", "superadmin")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?userRole=member&callerTeamRole=member", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

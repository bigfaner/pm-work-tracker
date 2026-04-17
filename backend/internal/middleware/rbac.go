package middleware

import (
	"github.com/gin-gonic/gin"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// RequireRole returns a Gin middleware that checks the authenticated user's
// global role (set by AuthMiddleware) against the allowed roles.
// If the user's role is not in the list, it aborts with 403.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := toSet(roles)
	return func(c *gin.Context) {
		if !allowed[GetUserRole(c)] {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrForbidden)
			return
		}
		c.Next()
	}
}

// RequireTeamRole returns a Gin middleware that checks the caller's team-scoped
// role (set by TeamScopeMiddleware) against the allowed roles.
// SuperAdmin users (userRole == "superadmin") always pass, regardless of team role.
func RequireTeamRole(roles ...string) gin.HandlerFunc {
	allowed := toSet(roles)
	return func(c *gin.Context) {
		// SuperAdmin bypasses team role check
		if GetUserRole(c) == "superadmin" {
			c.Next()
			return
		}
		if !allowed[GetCallerTeamRole(c)] {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrForbidden)
			return
		}
		c.Next()
	}
}

// toSet converts a slice of strings to a lookup map.
func toSet(keys []string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}

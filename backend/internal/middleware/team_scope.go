package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

// TeamScopeMiddleware verifies team membership for all /api/v1/teams/:teamId/* routes.
// It must run after AuthMiddleware in the middleware chain.
//
// It extracts teamId from the URL, checks membership (or SuperAdmin bypass),
// and injects teamID + callerTeamRole into the Gin context.
func TeamScopeMiddleware(teamRepo repository.TeamRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Parse teamId from URL param
		teamIDStr := c.Param("teamId")
		teamID, err := strconv.ParseUint(teamIDStr, 10, 64)
		if err != nil {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrValidation)
			return
		}
		teamIDUint := uint(teamID)

		// 2. Get caller identity from context (set by AuthMiddleware)
		userRole := GetUserRole(c)

		// 3. SuperAdmin bypasses membership check
		if userRole == "superadmin" {
			c.Set("teamID", teamIDUint)
			c.Set("callerTeamRole", "superadmin")
			c.Next()
			return
		}

		// 4. Look up TeamMember record
		userID := GetUserID(c)
		member, err := teamRepo.FindMember(c.Request.Context(), teamIDUint, userID)
		if err != nil {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrNotTeamMember)
			return
		}

		// 5. Inject teamID and callerTeamRole into context
		c.Set("teamID", teamIDUint)
		c.Set("callerTeamRole", member.Role)
		c.Next()
	}
}

// GetTeamID extracts the scoped team ID from the Gin context.
func GetTeamID(c *gin.Context) uint {
	if v, ok := c.Get("teamID"); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// GetCallerTeamRole extracts the caller's role within the scoped team from the Gin context.
func GetCallerTeamRole(c *gin.Context) string {
	if v, ok := c.Get("callerTeamRole"); ok {
		if role, ok := v.(string); ok {
			return role
		}
	}
	return ""
}

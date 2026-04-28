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
// and injects teamID, callerTeamRole, and permCodes into the Gin context.
func TeamScopeMiddleware(teamRepo repository.TeamRepo, roleRepo repository.RoleRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Parse teamId from URL param as bizKey (snowflake int64)
		teamIDStr := c.Param("teamId")
		teamBizKey, err := strconv.ParseInt(teamIDStr, 10, 64)
		if err != nil || teamBizKey <= 0 {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrValidation)
			return
		}

		// Resolve bizKey → internal auto-increment ID
		team, err := teamRepo.FindByBizKey(c.Request.Context(), teamBizKey)
		if err != nil {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrTeamNotFound)
			return
		}
		// 2. SuperAdmin bypasses membership check
		if IsSuperAdmin(c) {
			c.Set("teamBizKey", teamBizKey)
			c.Set("callerTeamRole", "superadmin")
			c.Set("permCodes", []string{})
			c.Next()
			return
		}

		// 3. Look up TeamMember record
		userID := GetUserID(c)
		member, err := teamRepo.FindMember(c.Request.Context(), team.ID, userID)
		if err != nil {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrNotTeamMember)
			return
		}

		// 4. Load permission codes from role
		// member.RoleKey is the role's biz_key; resolve to auto-increment id first.
		var permCodes []string
		if member.RoleKey != nil {
			role, err := roleRepo.FindByBizKey(c.Request.Context(), *member.RoleKey)
			if err != nil {
				c.Abort()
				apperrors.RespondError(c, apperrors.ErrInternal)
				return
			}
			codes, err := roleRepo.ListPermissions(c.Request.Context(), role.BizKey)
			if err != nil {
				c.Abort()
				apperrors.RespondError(c, apperrors.ErrInternal)
				return
			}
			permCodes = codes
		}

		// 5. Inject teamBizKey, callerTeamRole, and permCodes into context
		c.Set("teamBizKey", teamBizKey)
		c.Set("callerTeamRole", "member")
		c.Set("permCodes", permCodes)
		c.Next()
	}
}

// GetTeamBizKey extracts the scoped team biz key from the Gin context.
func GetTeamBizKey(c *gin.Context) int64 {
	if v, ok := c.Get("teamBizKey"); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

// GetTeamID returns the team biz key cast to uint, for handlers that pass teamID
// to services whose interfaces still use uint. The cast is lossless for positive
// snowflake IDs, and service-layer comparisons of the form int64(teamID) == item.TeamKey
// remain correct because item.TeamKey stores the biz key.
func GetTeamID(c *gin.Context) uint {
	return uint(GetTeamBizKey(c))
}

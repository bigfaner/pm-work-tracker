package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

// RequirePermission returns a Gin middleware that checks if the authenticated user
// has the specified permission code.
//
// Check order:
//  1. SuperAdmin -> always pass
//  2. Team context (permCodes set) -> code in permCodes -> pass / 403
//  3. Non-team context -> query RoleRepo.HasPermission(userID, code) -> pass / 403
func RequirePermission(code string, roleRepo repository.RoleRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. SuperAdmin always passes
		if IsSuperAdmin(c) {
			c.Next()
			return
		}

		// 2. Team context: check permCodes
		permCodes := GetPermCodes(c)
		if permCodes != nil {
			if containsString(permCodes, code) {
				c.Next()
				return
			}
			c.Abort()
			apperrors.RespondError(c, &apperrors.AppError{
				Code:    "ERR_FORBIDDEN",
				Status:  403,
				Message: fmt.Sprintf("权限不足：缺少 %s 权限", code),
			})
			return
		}

		// 3. Non-team context: query DB
		userBizKey := GetUserBizKey(c)
		if userBizKey == 0 {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrForbidden)
			return
		}

		ok, err := roleRepo.HasPermission(c.Request.Context(), userBizKey, code)
		if err != nil {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrInternal)
			return
		}
		if ok {
			c.Next()
			return
		}

		c.Abort()
		apperrors.RespondError(c, &apperrors.AppError{
			Code:    "ERR_FORBIDDEN",
			Status:  403,
			Message: fmt.Sprintf("权限不足：缺少 %s 权限", code),
		})
	}
}

// GetPermCodes extracts the permission codes from the Gin context (set by TeamScopeMiddleware).
func GetPermCodes(c *gin.Context) []string {
	if v, ok := c.Get("permCodes"); ok {
		if codes, ok := v.([]string); ok {
			return codes
		}
	}
	return nil
}

// containsString checks if a string exists in a slice.
func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

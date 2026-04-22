package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	jwtpkg "pm-work-tracker/backend/internal/pkg/jwt"
	"pm-work-tracker/backend/internal/repository"
)

// AuthMiddleware validates the JWT on every protected route and injects
// the authenticated user's ID, role, and isSuperAdmin flag into the Gin context.
// After JWT extraction, it loads the User from DB to set isSuperAdmin.
func AuthMiddleware(jwtSecret string, userRepo repository.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrUnauthorized)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrUnauthorized)
			return
		}

		claims, err := jwtpkg.Verify(parts[1], jwtSecret)
		if err != nil {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrUnauthorized)
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		// Load User from DB to set isSuperAdmin
		user, err := userRepo.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.Abort()
			apperrors.RespondError(c, apperrors.ErrUnauthorized)
			return
		}
		c.Set("isSuperAdmin", user.IsSuperAdmin)

		c.Next()
	}
}

// GetUserID extracts the authenticated user's ID from the Gin context.
func GetUserID(c *gin.Context) uint {
	if v, ok := c.Get("userID"); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// IsSuperAdmin returns whether the authenticated user is a super admin.
func IsSuperAdmin(c *gin.Context) bool {
	if v, ok := c.Get("isSuperAdmin"); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

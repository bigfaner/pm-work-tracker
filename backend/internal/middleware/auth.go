package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	jwtpkg "pm-work-tracker/backend/internal/pkg/jwt"
)

// AuthMiddleware validates the JWT on every protected route and injects
// the authenticated user's ID and role into the Gin context.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
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
		c.Set("userRole", claims.Role)
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

// GetUserRole extracts the authenticated user's role from the Gin context.
func GetUserRole(c *gin.Context) string {
	if v, ok := c.Get("userRole"); ok {
		if role, ok := v.(string); ok {
			return role
		}
	}
	return ""
}

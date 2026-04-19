package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/service"
)

// PermissionHandler handles permission query endpoints.
type PermissionHandler struct {
	roleSvc service.RoleService
}

// NewPermissionHandler creates a PermissionHandler stub (nil service → 501 responses).
func NewPermissionHandler() *PermissionHandler {
	return &PermissionHandler{}
}

// NewPermissionHandlerWithDeps creates a PermissionHandler wired to a real RoleService.
func NewPermissionHandlerWithDeps(roleSvc service.RoleService) *PermissionHandler {
	return &PermissionHandler{roleSvc: roleSvc}
}

// ListPermissionCodes handles GET /api/v1/admin/permissions
// Returns the static permission registry grouped by resource — no DB query needed.
func (h *PermissionHandler) ListPermissionCodes(c *gin.Context) {
	if h.roleSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	groups := h.roleSvc.ListPermissionCodes(c.Request.Context())
	apperrors.RespondOK(c, groups)
}

// GetUserPermissions handles GET /api/v1/me/permissions
// Returns the current user's permission map (is_superadmin + team_permissions).
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	if h.roleSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	userID := middleware.GetUserID(c)
	perms, err := h.roleSvc.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, perms)
}

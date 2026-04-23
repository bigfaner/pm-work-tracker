package handler

import (
	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/service"
)

// PermissionHandler handles permission query endpoints.
type PermissionHandler struct {
	roleSvc service.RoleService
}

// NewPermissionHandler creates a new PermissionHandler with service dependency.
func NewPermissionHandler(roleSvc service.RoleService) *PermissionHandler {
	if roleSvc == nil {
		panic("permission_handler: roleService must not be nil")
	}
	return &PermissionHandler{roleSvc: roleSvc}
}

// ListPermissionCodes handles GET /api/v1/admin/permissions
// Returns the static permission registry grouped by resource — no DB query needed.
func (h *PermissionHandler) ListPermissionCodes(c *gin.Context) {
	groups := h.roleSvc.ListPermissionCodes(c.Request.Context())
	apperrors.RespondOK(c, groups)
}

// GetUserPermissions handles GET /api/v1/me/permissions
// Returns the current user's permission map (is_superadmin + team_permissions).
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	perms, err := h.roleSvc.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, perms)
}

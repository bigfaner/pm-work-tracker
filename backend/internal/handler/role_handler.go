package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/service"
)

// RoleHandler handles role CRUD endpoints under /api/v1/admin/roles.
type RoleHandler struct {
	roleSvc service.RoleService
}

// NewRoleHandler creates a RoleHandler stub (nil service → 501 responses).
func NewRoleHandler() *RoleHandler {
	return &RoleHandler{}
}

// NewRoleHandlerWithDeps creates a RoleHandler wired to a real RoleService.
func NewRoleHandlerWithDeps(roleSvc service.RoleService) *RoleHandler {
	return &RoleHandler{roleSvc: roleSvc}
}

// ListRoles handles GET /api/v1/admin/roles
func (h *RoleHandler) ListRoles(c *gin.Context) {
	if h.roleSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	items, err := h.roleSvc.ListRoles(c.Request.Context())
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	page, pageSize := parsePagination(c, 20)
	apperrors.RespondOK(c, gin.H{
		"items":    items,
		"total":    len(items),
		"page":     page,
		"pageSize": pageSize,
	})
}

// CreateRoleReq is the request body for POST /api/v1/admin/roles.
type CreateRoleReq struct {
	Name            string   `json:"name" binding:"required"`
	Description     string   `json:"description"`
	PermissionCodes []string `json:"permission_codes" binding:"required,min=1"`
}

// CreateRole handles POST /api/v1/admin/roles
func (h *RoleHandler) CreateRole(c *gin.Context) {
	if h.roleSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	var req CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	role, err := h.roleSvc.CreateRole(c.Request.Context(), service.CreateRoleReq{
		Name:            req.Name,
		Description:     req.Description,
		PermissionCodes: req.PermissionCodes,
	})
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": role})
}

// GetRole handles GET /api/v1/admin/roles/:id
func (h *RoleHandler) GetRole(c *gin.Context) {
	if h.roleSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	roleID, err := parseRoleID(c)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	detail, err := h.roleSvc.GetRole(c.Request.Context(), roleID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, detail)
}

// UpdateRoleReq is the request body for PUT /api/v1/admin/roles/:id.
type UpdateRoleReq struct {
	Name            *string  `json:"name,omitempty"`
	Description     *string  `json:"description,omitempty"`
	PermissionCodes []string `json:"permission_codes,omitempty"`
}

// UpdateRole handles PUT /api/v1/admin/roles/:id
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	if h.roleSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	roleID, err := parseRoleID(c)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	var req UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	detail, err := h.roleSvc.UpdateRole(c.Request.Context(), roleID, service.UpdateRoleReq{
		Name:            req.Name,
		Description:     req.Description,
		PermissionCodes: req.PermissionCodes,
	})
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, detail)
}

// DeleteRole handles DELETE /api/v1/admin/roles/:id
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	if h.roleSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	roleID, err := parseRoleID(c)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	if err := h.roleSvc.DeleteRole(c.Request.Context(), roleID); err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// parseRoleID extracts the role ID from the URL path parameter.
func parseRoleID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

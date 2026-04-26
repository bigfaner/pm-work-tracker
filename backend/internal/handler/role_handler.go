package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	pkgHandler "pm-work-tracker/backend/internal/pkg/handler"
	"pm-work-tracker/backend/internal/service"
)

// RoleHandler handles role CRUD endpoints under /api/v1/admin/roles.
type RoleHandler struct {
	roleSvc service.RoleService
}

// NewRoleHandler creates a RoleHandler wired to a RoleService.
func NewRoleHandler(roleSvc service.RoleService) *RoleHandler {
	if roleSvc == nil {
		panic("role_handler: roleService must not be nil")
	}
	return &RoleHandler{roleSvc: roleSvc}
}

// ListRoles handles GET /api/v1/admin/roles
func (h *RoleHandler) ListRoles(c *gin.Context) {
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

// CreateRole handles POST /api/v1/admin/roles
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	role, err := h.roleSvc.CreateRole(c.Request.Context(), req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": role})
}

// GetRole handles GET /api/v1/admin/roles/:id
func (h *RoleHandler) GetRole(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "id")
	if !ok {
		return
	}

	detail, err := h.roleSvc.GetRole(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, detail)
}

// UpdateRole handles PUT /api/v1/admin/roles/:id
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	detail, err := h.roleSvc.UpdateRole(c.Request.Context(), bizKey, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, detail)
}

// DeleteRole handles DELETE /api/v1/admin/roles/:id
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "id")
	if !ok {
		return
	}

	if err := h.roleSvc.DeleteRole(c.Request.Context(), bizKey); err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}


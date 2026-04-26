package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/service"
)

// AdminHandler handles admin endpoints.
type AdminHandler struct {
	adminSvc service.AdminService
}

// NewAdminHandler creates a new AdminHandler with service dependency.
func NewAdminHandler(adminSvc service.AdminService) *AdminHandler {
	if adminSvc == nil {
		panic("admin_handler: adminService must not be nil")
	}
	return &AdminHandler{adminSvc: adminSvc}
}

// ListUsers handles GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, pageSize := parsePagination(c, 20)

	search := c.Query("search")

	items, total, err := h.adminSvc.ListUsers(c.Request.Context(), search, page, pageSize)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// CreateUser handles POST /api/v1/admin/users
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	user, err := h.adminSvc.CreateUser(c.Request.Context(), &req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": user})
}

// GetUser handles GET /api/v1/admin/users/:userId
func (h *AdminHandler) GetUser(c *gin.Context) {
	bizKey, ok := parseUserBizKey(c)
	if !ok {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	user, err := h.adminSvc.GetUser(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, user)
}

// UpdateUser handles PUT /api/v1/admin/users/:userId
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	bizKey, ok := parseUserBizKey(c)
	if !ok {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	user, err := h.adminSvc.UpdateUser(c.Request.Context(), bizKey, &req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, user)
}

// ToggleUserStatus handles PUT /api/v1/admin/users/:userId/status
func (h *AdminHandler) ToggleUserStatus(c *gin.Context) {
	bizKey, ok := parseUserBizKey(c)
	if !ok {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	var req dto.UpdateUserStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	callerID := middleware.GetUserID(c)
	user, err := h.adminSvc.ToggleUserStatus(c.Request.Context(), callerID, bizKey, req.Status)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, user)
}

// ListTeams handles GET /api/v1/admin/teams
func (h *AdminHandler) ListTeams(c *gin.Context) {
	page, pageSize := parsePagination(c, 50)

	teams, err := h.adminSvc.ListAllTeams(c.Request.Context())
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	items := paginateAdminTeams(teams, page, pageSize)
	apperrors.RespondOK(c, gin.H{
		"items":    items,
		"total":    len(teams),
		"page":     page,
		"pageSize": pageSize,
	})
}

// parseUserBizKey extracts the user bizKey from the URL path parameter.
func parseUserBizKey(c *gin.Context) (int64, bool) {
	idStr := c.Param("userId")
	bizKey, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return bizKey, true
}

// parsePagination extracts page and pageSize from query params with defaults.
func parsePagination(c *gin.Context, defaultPageSize int) (int, int) {
	page := 1
	pageSize := defaultPageSize

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}

	return page, pageSize
}

// paginateAdminTeams applies pagination to an admin team list.
func paginateAdminTeams(teams []*dto.AdminTeamDTO, page, pageSize int) []*dto.AdminTeamDTO {
	total := len(teams)
	start := (page - 1) * pageSize
	if start >= total {
		return []*dto.AdminTeamDTO{}
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return teams[start:end]
}

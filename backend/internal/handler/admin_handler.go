package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/service"
)

// AdminHandler handles admin endpoints.
type AdminHandler struct {
	adminSvc service.AdminService
}

// NewAdminHandler creates a new AdminHandler (stub, for router setup before service is ready).
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// NewAdminHandlerWithDeps creates a new AdminHandler with service dependency.
func NewAdminHandlerWithDeps(adminSvc service.AdminService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc}
}

// UpdateCanCreateTeamReq is the request DTO for toggling canCreateTeam.
type UpdateCanCreateTeamReq struct {
	CanCreateTeam *bool `json:"canCreateTeam" binding:"required"`
}

// ListUsers handles GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	if h.adminSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	page, pageSize := parsePagination(c, 50)

	users, err := h.adminSvc.ListUsers(c.Request.Context())
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	items := paginateUsers(users, page, pageSize)
	apperrors.RespondOK(c, gin.H{
		"items":    items,
		"total":    len(users),
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateCanCreateTeam handles PUT /api/v1/admin/users/:userId/can-create-team
func (h *AdminHandler) UpdateCanCreateTeam(c *gin.Context) {
	if h.adminSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	// Parse userId path param as uint
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	var req UpdateCanCreateTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	callerID := middleware.GetUserID(c)
	if err := h.adminSvc.SetCanCreateTeam(c.Request.Context(), callerID, uint(userID), *req.CanCreateTeam); err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// ListTeams handles GET /api/v1/admin/teams
func (h *AdminHandler) ListTeams(c *gin.Context) {
	if h.adminSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

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

// paginateUsers applies pagination to a user list and converts to DTOs.
func paginateUsers(users []*model.User, page, pageSize int) []dto.UserDTO {
	total := len(users)
	start := (page - 1) * pageSize
	if start >= total {
		return []dto.UserDTO{}
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	result := make([]dto.UserDTO, 0, end-start)
	for i := start; i < end; i++ {
		u := users[i]
		result = append(result, dto.UserDTO{
			ID:            u.ID,
			Username:      u.Username,
			DisplayName:   u.DisplayName,
			IsSuperAdmin:  u.IsSuperAdmin,
			CanCreateTeam: u.CanCreateTeam,
		})
	}
	return result
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

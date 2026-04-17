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

// SubItemHandler handles sub item endpoints.
type SubItemHandler struct {
	svc service.SubItemService
}

// NewSubItemHandler creates a new SubItemHandler (stub, for router setup before service is ready).
func NewSubItemHandler() *SubItemHandler {
	return &SubItemHandler{}
}

// NewSubItemHandlerWithDeps creates a new SubItemHandler with service dependency.
func NewSubItemHandlerWithDeps(svc service.SubItemService) *SubItemHandler {
	return &SubItemHandler{svc: svc}
}

// isPMOrSuperAdmin returns true if the caller has PM role in the team or is a superadmin.
func (h *SubItemHandler) isPMOrSuperAdmin(c *gin.Context) bool {
	if middleware.GetUserRole(c) == "superadmin" {
		return true
	}
	return middleware.GetCallerTeamRole(c) == "pm"
}

// parseSubID extracts and validates the subId path param as uint.
func parseSubID(c *gin.Context) (uint, bool) {
	idStr := c.Param("subId")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return 0, false
	}
	return uint(id), true
}

// parseItemIDAsUint extracts and validates the itemId path param as uint.
func parseItemIDAsUint(c *gin.Context) (uint, bool) {
	idStr := c.Param("itemId")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return 0, false
	}
	return uint(id), true
}

// Create handles POST /api/v1/teams/:teamId/main-items/:itemId/sub-items
func (h *SubItemHandler) Create(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)
	callerID := middleware.GetUserID(c)

	var req dto.SubItemCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	item, err := h.svc.Create(c.Request.Context(), teamID, callerID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": subItemToDTO(item)})
}

// List handles GET /api/v1/teams/:teamId/main-items/:itemId/sub-items
func (h *SubItemHandler) List(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)

	mainIDStr := c.Param("itemId")
	mainIDUint, err := strconv.ParseUint(mainIDStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}
	mainID := uint(mainIDUint)

	var filter dto.SubItemFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	var page dto.Pagination
	if err := c.ShouldBindQuery(&page); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}
	if page.Page <= 0 {
		page.Page = 1
	}
	if page.PageSize <= 0 {
		page.PageSize = 20
	}

	result, err := h.svc.List(c.Request.Context(), teamID, &mainID, filter, page)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, result)
}

// Get handles GET /api/v1/teams/:teamId/sub-items/:subId
func (h *SubItemHandler) Get(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	subID, ok := parseSubID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)

	item, err := h.svc.Get(c.Request.Context(), teamID, subID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, subItemToDTO(item))
}

// Update handles PUT /api/v1/teams/:teamId/sub-items/:subId
func (h *SubItemHandler) Update(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	subID, ok := parseSubID(c)
	if !ok {
		return
	}

	// RBAC: PM or current assignee can update
	if !h.isPMOrSuperAdmin(c) {
		// Check if caller is the assignee
		teamID := middleware.GetTeamID(c)
		item, err := h.svc.Get(c.Request.Context(), teamID, subID)
		if err != nil {
			apperrors.RespondError(c, err)
			return
		}
		callerID := middleware.GetUserID(c)
		if item.AssigneeID == nil || *item.AssigneeID != callerID {
			apperrors.RespondError(c, apperrors.ErrForbidden)
			return
		}
	}

	teamID := middleware.GetTeamID(c)

	var req dto.SubItemUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.svc.Update(c.Request.Context(), teamID, subID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch updated item for response
	updated, err := h.svc.Get(c.Request.Context(), teamID, subID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, subItemToDTO(updated))
}

// ChangeStatus handles PUT /api/v1/teams/:teamId/sub-items/:subId/status
func (h *SubItemHandler) ChangeStatus(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	subID, ok := parseSubID(c)
	if !ok {
		return
	}

	// RBAC: PM or current assignee can change status
	if !h.isPMOrSuperAdmin(c) {
		teamID := middleware.GetTeamID(c)
		item, err := h.svc.Get(c.Request.Context(), teamID, subID)
		if err != nil {
			apperrors.RespondError(c, err)
			return
		}
		callerID := middleware.GetUserID(c)
		if item.AssigneeID == nil || *item.AssigneeID != callerID {
			apperrors.RespondError(c, apperrors.ErrForbidden)
			return
		}
	}

	teamID := middleware.GetTeamID(c)
	callerID := middleware.GetUserID(c)

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.svc.ChangeStatus(c.Request.Context(), teamID, callerID, subID, req.Status)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// Assign handles PUT /api/v1/teams/:teamId/sub-items/:subId/assignee
func (h *SubItemHandler) Assign(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	// PM only
	if !h.isPMOrSuperAdmin(c) {
		apperrors.RespondError(c, apperrors.ErrForbidden)
		return
	}

	subID, ok := parseSubID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	var req struct {
		AssigneeID uint `json:"assignee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.svc.Assign(c.Request.Context(), teamID, pmID, subID, req.AssigneeID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// subItemToDTO converts a model.SubItem to a response map matching the Data Contract.
func subItemToDTO(item *model.SubItem) gin.H {
	return gin.H{
		"id":               item.ID,
		"mainItemId":       item.MainItemID,
		"title":            item.Title,
		"description":      item.Description,
		"priority":         item.Priority,
		"assigneeId":       item.AssigneeID,
		"startDate":        item.StartDate,
		"expectedEndDate":  item.ExpectedEndDate,
		"actualEndDate":    item.ActualEndDate,
		"status":           item.Status,
		"completion":       item.Completion,
		"isKeyItem":        item.IsKeyItem,
		"delayCount":       item.DelayCount,
		"weight":           item.Weight,
		"createdAt":        item.CreatedAt,
		"updatedAt":        item.UpdatedAt,
	}
}

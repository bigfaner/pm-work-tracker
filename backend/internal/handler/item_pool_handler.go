package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
)

// ItemPoolHandler handles item pool endpoints.
type ItemPoolHandler struct {
	svc      service.ItemPoolService
	userRepo repository.UserRepo
}

// NewItemPoolHandler creates a new ItemPoolHandler (stub, for router setup before service is ready).
func NewItemPoolHandler() *ItemPoolHandler {
	return &ItemPoolHandler{}
}

// NewItemPoolHandlerWithDeps creates a new ItemPoolHandler with service and repo dependencies.
func NewItemPoolHandlerWithDeps(svc service.ItemPoolService, userRepo repository.UserRepo) *ItemPoolHandler {
	return &ItemPoolHandler{svc: svc, userRepo: userRepo}
}

// Submit handles POST /api/v1/teams/:teamId/item-pool
func (h *ItemPoolHandler) Submit(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)
	userID := middleware.GetUserID(c)

	var req dto.SubmitItemPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	item, err := h.svc.Submit(c.Request.Context(), teamID, userID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": itemPoolToDTO(item, h.userRepo, c)})
}

// List handles GET /api/v1/teams/:teamId/item-pool
func (h *ItemPoolHandler) List(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)

	var filter dto.ItemPoolFilter
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

	result, err := h.svc.List(c.Request.Context(), teamID, filter, page)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, result)
}

// Get handles GET /api/v1/teams/:teamId/item-pool/:poolId
func (h *ItemPoolHandler) Get(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	poolID, ok := parsePoolID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)

	item, err := h.svc.Get(c.Request.Context(), teamID, poolID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, itemPoolToDTO(item, h.userRepo, c))
}

// Assign handles POST /api/v1/teams/:teamId/item-pool/:poolId/assign
func (h *ItemPoolHandler) Assign(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	poolID, ok := parsePoolID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	var req dto.AssignItemPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.svc.Assign(c.Request.Context(), teamID, pmID, poolID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch the updated item to get the assignedSubId
	updated, err := h.svc.Get(c.Request.Context(), teamID, poolID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"subItemId": updated.AssignedSubID}})
}

// Reject handles POST /api/v1/teams/:teamId/item-pool/:poolId/reject
func (h *ItemPoolHandler) Reject(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	poolID, ok := parsePoolID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	var req dto.RejectItemPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.svc.Reject(c.Request.Context(), teamID, pmID, poolID, req.Reason)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch the updated item for response
	updated, err := h.svc.Get(c.Request.Context(), teamID, poolID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, itemPoolToDTO(updated, h.userRepo, c))
}

// parsePoolID extracts and validates the poolId path param as uint.
func parsePoolID(c *gin.Context) (uint, bool) {
	idStr := c.Param("poolId")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return 0, false
	}
	return uint(id), true
}

// itemPoolToDTO converts a model.ItemPool to a response map matching the Data Contract.
// Includes submitterName via user lookup.
func itemPoolToDTO(item *model.ItemPool, userRepo repository.UserRepo, c *gin.Context) gin.H {
	m := gin.H{
		"id":              item.ID,
		"title":           item.Title,
		"background":      item.Background,
		"expectedOutput":  item.ExpectedOutput,
		"submitterId":     item.SubmitterID,
		"status":          item.Status,
		"assignedMainId":  item.AssignedMainID,
		"assignedSubId":   item.AssignedSubID,
		"assigneeId":      item.AssigneeID,
		"rejectReason":    item.RejectReason,
		"reviewedAt":      item.ReviewedAt,
		"createdAt":       item.CreatedAt,
	}

	// Look up submitter name
	if userRepo != nil && item.SubmitterID > 0 {
		if user, err := userRepo.FindByID(c.Request.Context(), item.SubmitterID); err == nil && user != nil {
			m["submitterName"] = user.DisplayName
		}
	}

	return m
}

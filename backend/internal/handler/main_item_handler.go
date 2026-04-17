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

// MainItemHandler handles main item endpoints.
type MainItemHandler struct {
	svc         service.MainItemService
	subItemRepo repository.SubItemRepo
	userRepo    repository.UserRepo
}

// NewMainItemHandler creates a new MainItemHandler (stub, for router setup before service is ready).
func NewMainItemHandler() *MainItemHandler {
	return &MainItemHandler{}
}

// NewMainItemHandlerWithDeps creates a new MainItemHandler with service and repo dependencies.
func NewMainItemHandlerWithDeps(svc service.MainItemService, userRepo repository.UserRepo, subItemRepo repository.SubItemRepo) *MainItemHandler {
	return &MainItemHandler{svc: svc, userRepo: userRepo, subItemRepo: subItemRepo}
}

// isPMOrSuperAdmin returns true if the caller has PM role in the team or is a superadmin.
func (h *MainItemHandler) isPMOrSuperAdmin(c *gin.Context) bool {
	if middleware.GetUserRole(c) == "superadmin" {
		return true
	}
	return middleware.GetCallerTeamRole(c) == "pm"
}

// parseItemID extracts and validates the itemId path param as uint.
func parseItemID(c *gin.Context) (uint, bool) {
	idStr := c.Param("itemId")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return 0, false
	}
	return uint(id), true
}

// Create handles POST /api/v1/teams/:teamId/main-items
func (h *MainItemHandler) Create(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	if !h.isPMOrSuperAdmin(c) {
		apperrors.RespondError(c, apperrors.ErrForbidden)
		return
	}

	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	var req dto.MainItemCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	item, err := h.svc.Create(c.Request.Context(), teamID, pmID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": mainItemToDTO(item)})
}

// List handles GET /api/v1/teams/:teamId/main-items
func (h *MainItemHandler) List(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)

	var filter dto.MainItemFilter
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

// Get handles GET /api/v1/teams/:teamId/main-items/:itemId
func (h *MainItemHandler) Get(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	itemID, ok := parseItemID(c)
	if !ok {
		return
	}

	item, err := h.svc.Get(c.Request.Context(), itemID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	resp := mainItemToDTO(item)

	// Fetch subItems summary
	subItems, _ := h.subItemRepo.ListByMainItem(c.Request.Context(), itemID)
	resp["subItems"] = subItemsToSummaryDTOs(subItems)

	apperrors.RespondOK(c, resp)
}

// Update handles PUT /api/v1/teams/:teamId/main-items/:itemId
func (h *MainItemHandler) Update(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	if !h.isPMOrSuperAdmin(c) {
		apperrors.RespondError(c, apperrors.ErrForbidden)
		return
	}

	itemID, ok := parseItemID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)

	var req dto.MainItemUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.svc.Update(c.Request.Context(), teamID, itemID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch updated item for response
	updated, err := h.svc.Get(c.Request.Context(), itemID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, mainItemToDTO(updated))
}

// Archive handles POST /api/v1/teams/:teamId/main-items/:itemId/archive
func (h *MainItemHandler) Archive(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	if !h.isPMOrSuperAdmin(c) {
		apperrors.RespondError(c, apperrors.ErrForbidden)
		return
	}

	itemID, ok := parseItemID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)

	err := h.svc.Archive(c.Request.Context(), teamID, itemID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// mainItemToDTO converts a model.MainItem to a response map matching the Data Contract.
func mainItemToDTO(item *model.MainItem) gin.H {
	m := gin.H{
		"id":               item.ID,
		"code":             item.Code,
		"title":            item.Title,
		"priority":         item.Priority,
		"proposerId":       item.ProposerID,
		"assigneeId":       item.AssigneeID,
		"status":           item.Status,
		"completion":       item.Completion,
		"isKeyItem":        item.IsKeyItem,
		"delayCount":       item.DelayCount,
		"archivedAt":       item.ArchivedAt,
		"startDate":        item.StartDate,
		"expectedEndDate":  item.ExpectedEndDate,
		"actualEndDate":    item.ActualEndDate,
		"createdAt":        item.CreatedAt,
		"updatedAt":        item.UpdatedAt,
	}
	return m
}

// subItemsToSummaryDTOs converts sub items to a summary array for the Get response.
func subItemsToSummaryDTOs(items []*model.SubItem) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, si := range items {
		result = append(result, gin.H{
			"id":              si.ID,
			"title":           si.Title,
			"status":          si.Status,
			"completion":      si.Completion,
			"assigneeId":      si.AssigneeID,
			"priority":        si.Priority,
			"startDate":       si.StartDate,
			"expectedEndDate": si.ExpectedEndDate,
		})
	}
	return result
}

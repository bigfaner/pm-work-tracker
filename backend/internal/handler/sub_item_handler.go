package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
)

// SubItemHandler handles sub item endpoints.
type SubItemHandler struct {
	svc service.SubItemService
}

// NewSubItemHandler creates a new SubItemHandler with service dependency.
func NewSubItemHandler(svc service.SubItemService) *SubItemHandler {
	if svc == nil {
		panic("sub_item_handler: subItemService must not be nil")
	}
	return &SubItemHandler{svc: svc}
}

// isPMOrSuperAdmin checks if the caller is a PM or superadmin using the permission codes in context.
// This is used for the assignee business rule pattern in sub_item handlers.
func isPMOrSuperAdmin(c *gin.Context) bool {
	if middleware.IsSuperAdmin(c) {
		return true
	}
	// Check if the caller has PM-level permissions by checking a PM-specific permission code.
	// The "team:invite" permission is typically PM-only.
	permCodes := middleware.GetPermCodes(c)
	if permCodes != nil {
		for _, code := range permCodes {
			if code == "team:invite" {
				return true
			}
		}
	}
	return false
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

// Create handles POST /api/v1/teams/:teamId/main-items/:itemId/sub-items
func (h *SubItemHandler) Create(c *gin.Context) {
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

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": vo.NewSubItemVO(item)})
}

// List handles GET /api/v1/teams/:teamId/main-items/:itemId/sub-items
func (h *SubItemHandler) List(c *gin.Context) {
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
	_, page.Page, page.PageSize = dto.ApplyPaginationDefaults(page.Page, page.PageSize)

	result, err := h.svc.List(c.Request.Context(), teamID, &mainID, filter, page)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	voItems := make([]vo.SubItemVO, 0, len(result.Items))
	for i := range result.Items {
		voItems = append(voItems, vo.NewSubItemVO(&result.Items[i]))
	}
	apperrors.RespondOK(c, gin.H{
		"items": voItems,
		"total": result.Total,
		"page":  result.Page,
		"size":  result.Size,
	})
}

// Get handles GET /api/v1/teams/:teamId/sub-items/:subId
func (h *SubItemHandler) Get(c *gin.Context) {
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

	apperrors.RespondOK(c, vo.NewSubItemVO(item))
}

// Update handles PUT /api/v1/teams/:teamId/sub-items/:subId
func (h *SubItemHandler) Update(c *gin.Context) {
	subID, ok := parseSubID(c)
	if !ok {
		return
	}

	// Assignee pattern: PM/SuperAdmin can update all, other members only their assigned items.
	// Permission check (sub_item:update) is done by RequirePermission middleware.
	if !isPMOrSuperAdmin(c) {
		// Check if caller is the assignee
		teamID := middleware.GetTeamID(c)
		item, err := h.svc.Get(c.Request.Context(), teamID, subID)
		if err != nil {
			apperrors.RespondError(c, err)
			return
		}
		callerID := middleware.GetUserID(c)
		if item.AssigneeKey == nil || uint(*item.AssigneeKey) != callerID {
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

	apperrors.RespondOK(c, vo.NewSubItemVO(updated))
}

// ChangeStatus handles PUT /api/v1/teams/:teamId/sub-items/:subId/status
func (h *SubItemHandler) ChangeStatus(c *gin.Context) {
	subID, ok := parseSubID(c)
	if !ok {
		return
	}

	// Assignee pattern: PM/SuperAdmin can change status for all, other members only their assigned items.
	// Permission check (sub_item:change_status) is done by RequirePermission middleware.
	if !isPMOrSuperAdmin(c) {
		teamID := middleware.GetTeamID(c)
		item, err := h.svc.Get(c.Request.Context(), teamID, subID)
		if err != nil {
			apperrors.RespondError(c, err)
			return
		}
		callerID := middleware.GetUserID(c)
		if item.AssigneeKey == nil || uint(*item.AssigneeKey) != callerID {
			apperrors.RespondError(c, apperrors.ErrForbidden)
			return
		}
	}

	teamID := middleware.GetTeamID(c)
	callerID := middleware.GetUserID(c)

	var req dto.ChangeStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	result, err := h.svc.ChangeStatus(c.Request.Context(), teamID, callerID, subID, req.Status)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	response := gin.H{"subItem": vo.NewSubItemVO(result.SubItem)}
	if result.LinkageResult != nil && result.LinkageResult.Warning() != "" {
		response["linkageWarning"] = result.LinkageResult.Warning()
	}
	apperrors.RespondOK(c, response)
}

// AvailableTransitions handles GET /api/v1/teams/:teamId/sub-items/:subId/available-transitions
func (h *SubItemHandler) AvailableTransitions(c *gin.Context) {
	subID, ok := parseSubID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)

	transitions, err := h.svc.AvailableTransitions(c.Request.Context(), teamID, subID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{"transitions": transitions})
}

// Assign handles PUT /api/v1/teams/:teamId/sub-items/:subId/assignee
func (h *SubItemHandler) Assign(c *gin.Context) {
	subID, ok := parseSubID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	var req dto.AssignSubItemReq
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

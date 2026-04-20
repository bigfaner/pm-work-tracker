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
	"pm-work-tracker/backend/internal/vo"
)

// ItemPoolHandler handles item pool endpoints.
type ItemPoolHandler struct {
	svc         service.ItemPoolService
	userRepo    repository.UserRepo
	mainItemRepo repository.MainItemRepo
}

// NewItemPoolHandler creates a new ItemPoolHandler with service and repo dependencies.
func NewItemPoolHandler(svc service.ItemPoolService, userRepo repository.UserRepo, mainItemRepo repository.MainItemRepo) *ItemPoolHandler {
	if svc == nil {
		panic("item_pool_handler: itemPoolService must not be nil")
	}
	if userRepo == nil {
		panic("item_pool_handler: userRepo must not be nil")
	}
	if mainItemRepo == nil {
		panic("item_pool_handler: mainItemRepo must not be nil")
	}
	return &ItemPoolHandler{svc: svc, userRepo: userRepo, mainItemRepo: mainItemRepo}
}

// Submit handles POST /api/v1/teams/:teamId/item-pool
func (h *ItemPoolHandler) Submit(c *gin.Context) {
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

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": itemPoolToVO(item, h.userRepo, h.mainItemRepo, c)})
}

// List handles GET /api/v1/teams/:teamId/item-pool
func (h *ItemPoolHandler) List(c *gin.Context) {
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
	_, page.Page, page.PageSize = dto.ApplyPaginationDefaults(page.Page, page.PageSize)

	result, err := h.svc.List(c.Request.Context(), teamID, filter, page)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	voItems := make([]vo.ItemPoolVO, 0, len(result.Items))
	for i := range result.Items {
		voItems = append(voItems, itemPoolToVO(&result.Items[i], h.userRepo, h.mainItemRepo, c))
	}
	apperrors.RespondOK(c, gin.H{
		"items": voItems,
		"total": result.Total,
		"page":  result.Page,
		"size":  result.Size,
	})
}

// Get handles GET /api/v1/teams/:teamId/item-pool/:poolId
func (h *ItemPoolHandler) Get(c *gin.Context) {
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

	apperrors.RespondOK(c, itemPoolToVO(item, h.userRepo, h.mainItemRepo, c))
}

// Assign handles POST /api/v1/teams/:teamId/item-pool/:poolId/assign
func (h *ItemPoolHandler) Assign(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"mainItemId": updated.AssignedMainID, "subItemId": updated.AssignedSubID}})
}

// ConvertToMain handles POST /api/v1/teams/:teamId/item-pool/:poolId/convert-to-main
func (h *ItemPoolHandler) ConvertToMain(c *gin.Context) {
	poolID, ok := parsePoolID(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	var req dto.ConvertToMainItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	mainItem, err := h.svc.ConvertToMain(c.Request.Context(), teamID, pmID, poolID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"mainItemId": mainItem.ID}})
}

// Reject handles POST /api/v1/teams/:teamId/item-pool/:poolId/reject
func (h *ItemPoolHandler) Reject(c *gin.Context) {
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

	apperrors.RespondOK(c, itemPoolToVO(updated, h.userRepo, h.mainItemRepo, c))
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

// itemPoolToVO converts a model.ItemPool to an ItemPoolVO.
func itemPoolToVO(item *model.ItemPool, userRepo repository.UserRepo, mainItemRepo repository.MainItemRepo, c *gin.Context) vo.ItemPoolVO {
	submitterName := ""
	if userRepo != nil && item.SubmitterID > 0 {
		if user, err := userRepo.FindByID(c.Request.Context(), item.SubmitterID); err == nil && user != nil {
			submitterName = user.DisplayName
		}
	}
	v := vo.NewItemPoolVO(item, submitterName)
	if mainItemRepo != nil && item.AssignedMainID != nil {
		if mi, err := mainItemRepo.FindByID(c.Request.Context(), *item.AssignedMainID); err == nil && mi != nil {
			v.AssignedMainCode = mi.Code
			v.AssignedMainTitle = mi.Title
		}
	}
	return v
}

package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/pkg"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	pkgHandler "pm-work-tracker/backend/internal/pkg/handler"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
)

// SubItemHandler handles sub item endpoints.
type SubItemHandler struct {
	svc         service.SubItemService
	mainItemSvc service.MainItemService
}

// NewSubItemHandler creates a new SubItemHandler with service dependency.
func NewSubItemHandler(svc service.SubItemService, mainItemSvc service.MainItemService) *SubItemHandler {
	if svc == nil {
		panic("sub_item_handler: subItemService must not be nil")
	}
	if mainItemSvc == nil {
		panic("sub_item_handler: mainItemService must not be nil")
	}
	return &SubItemHandler{svc: svc, mainItemSvc: mainItemSvc}
}

// hasPermCode checks whether the current context has a specific permission code.
func hasPermCode(c *gin.Context, code string) bool {
	permCodes := middleware.GetPermCodes(c)
	for _, pc := range permCodes {
		if pc == code {
			return true
		}
	}
	return false
}

// Create handles POST /api/v1/teams/:teamId/main-items/:itemId/sub-items
//
//nolint:dupl // similar request handling shape as MainItem.Create but different business logic
func (h *SubItemHandler) Create(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	callerBizKey := middleware.GetUserBizKey(c)

	var req dto.SubItemCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	item, err := h.svc.Create(c.Request.Context(), teamBizKey, callerBizKey, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": vo.NewSubItemVO(item)})
}

// List handles GET /api/v1/teams/:teamId/main-items/:itemId/sub-items
func (h *SubItemHandler) List(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)

	mainBizKey, ok := pkgHandler.ParseBizKeyParam(c, "itemId")
	if !ok {
		return
	}
	mainItem, err := h.mainItemSvc.GetByBizKey(c.Request.Context(), mainBizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}
	mainItemBizKey := mainItem.BizKey

	var filter dto.SubItemFilter
	if bindErr := c.ShouldBindQuery(&filter); bindErr != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	var page dto.Pagination
	if bindErr := c.ShouldBindQuery(&page); bindErr != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}
	_, page.Page, page.PageSize = dto.ApplyPaginationDefaults(page.Page, page.PageSize)

	result, err := h.svc.List(c.Request.Context(), teamBizKey, &mainItemBizKey, filter, page)
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
	subID, ok := pkgHandler.ResolveBizKey(c, "subId", func(ctx context.Context, bizKey int64) (uint, error) {
		item, err := h.svc.GetByBizKey(ctx, bizKey)
		if err != nil {
			return 0, err
		}
		return item.ID, nil
	})
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	item, err := h.svc.Get(c.Request.Context(), teamBizKey, subID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, vo.NewSubItemVO(item))
}

// Update handles PUT /api/v1/teams/:teamId/sub-items/:subId
func (h *SubItemHandler) Update(c *gin.Context) {
	subID, ok := pkgHandler.ResolveBizKey(c, "subId", func(ctx context.Context, bizKey int64) (uint, error) {
		item, err := h.svc.GetByBizKey(ctx, bizKey)
		if err != nil {
			return 0, err
		}
		return item.ID, nil
	})
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	var req dto.SubItemUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.svc.Update(c.Request.Context(), teamBizKey, subID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch updated item for response
	updated, err := h.svc.Get(c.Request.Context(), teamBizKey, subID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, vo.NewSubItemVO(updated))
}

// ChangeStatus handles PUT /api/v1/teams/:teamId/sub-items/:subId/status
func (h *SubItemHandler) ChangeStatus(c *gin.Context) {
	subID, ok := pkgHandler.ResolveBizKey(c, "subId", func(ctx context.Context, bizKey int64) (uint, error) {
		item, err := h.svc.GetByBizKey(ctx, bizKey)
		if err != nil {
			return 0, err
		}
		return item.ID, nil
	})
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)
	callerBizKey := middleware.GetUserBizKey(c)

	var req dto.ChangeStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	result, err := h.svc.ChangeStatus(c.Request.Context(), teamBizKey, callerBizKey, subID, req.Status)
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
	subID, ok := pkgHandler.ResolveBizKey(c, "subId", func(ctx context.Context, bizKey int64) (uint, error) {
		item, err := h.svc.GetByBizKey(ctx, bizKey)
		if err != nil {
			return 0, err
		}
		return item.ID, nil
	})
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	transitions, err := h.svc.AvailableTransitions(c.Request.Context(), teamBizKey, subID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{"transitions": transitions})
}

// Assign handles PUT /api/v1/teams/:teamId/sub-items/:subId/assignee
func (h *SubItemHandler) Assign(c *gin.Context) {
	subID, ok := pkgHandler.ResolveBizKey(c, "subId", func(ctx context.Context, bizKey int64) (uint, error) {
		item, err := h.svc.GetByBizKey(ctx, bizKey)
		if err != nil {
			return 0, err
		}
		return item.ID, nil
	})
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)
	pmBizKey := middleware.GetUserBizKey(c)

	var req dto.AssignSubItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.svc.Assign(c.Request.Context(), teamBizKey, pmBizKey, subID, func() int64 { v, _ := pkg.ParseID(req.AssigneeKey); return v }())
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// DeleteSubItem handles DELETE /api/v1/teams/:teamId/sub-items/:subId
func (h *SubItemHandler) DeleteSubItem(c *gin.Context) {
	subBizKey, ok := pkgHandler.ParseBizKeyParam(c, "subId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)
	callerBizKey := middleware.GetUserBizKey(c)

	item, err := h.svc.GetByBizKey(c.Request.Context(), subBizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	err = h.svc.Delete(c.Request.Context(), teamBizKey, callerBizKey, item.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{"message": "ok"})
}

// Move handles PUT /api/v1/teams/:teamId/sub-items/:subId/move
func (h *SubItemHandler) Move(c *gin.Context) {
	subBizKey, ok := pkgHandler.ParseBizKeyParam(c, "subId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)
	callerBizKey := middleware.GetUserBizKey(c)

	var req struct {
		TargetMainItemBizKey string `json:"targetMainItemBizKey" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	targetMainItemBizKey, err := pkg.ParseID(req.TargetMainItemBizKey)
	if err != nil || targetMainItemBizKey <= 0 {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	result, err := h.svc.Move(c.Request.Context(), teamBizKey, subBizKey, targetMainItemBizKey, callerBizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{
		"newSubCode":     result.NewSubCode,
		"mainItemBizKey": fmt.Sprintf("%d", result.MainItemBizKey),
	})
}

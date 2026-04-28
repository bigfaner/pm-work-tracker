package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	pkgHandler "pm-work-tracker/backend/internal/pkg/handler"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
)

// MainItemHandler handles main item endpoints.
type MainItemHandler struct {
	svc         service.MainItemService
	subItemRepo repository.SubItemRepo
	userRepo    repository.UserRepo
}

// NewMainItemHandler creates a new MainItemHandler with service and repo dependencies.
func NewMainItemHandler(svc service.MainItemService, userRepo repository.UserRepo, subItemRepo repository.SubItemRepo) *MainItemHandler {
	if svc == nil {
		panic("main_item_handler: mainItemService must not be nil")
	}
	if userRepo == nil {
		panic("main_item_handler: userRepo must not be nil")
	}
	if subItemRepo == nil {
		panic("main_item_handler: subItemRepo must not be nil")
	}
	return &MainItemHandler{svc: svc, userRepo: userRepo, subItemRepo: subItemRepo}
}

// Create handles POST /api/v1/teams/:teamId/main-items
func (h *MainItemHandler) Create(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	pmID := middleware.GetUserID(c)

	var req dto.MainItemCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	item, err := h.svc.Create(c.Request.Context(), teamBizKey, pmID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": vo.NewMainItemVO(item)})
}

// List handles GET /api/v1/teams/:teamId/main-items
func (h *MainItemHandler) List(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)

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
	_, page.Page, page.PageSize = dto.ApplyPaginationDefaults(page.Page, page.PageSize)

	result, err := h.svc.List(c.Request.Context(), teamBizKey, filter, page)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	voItems := make([]vo.MainItemVO, 0, len(result.Items))
	for i := range result.Items {
		voItems = append(voItems, vo.NewMainItemVO(&result.Items[i]))
	}
	apperrors.RespondOK(c, gin.H{
		"items": voItems,
		"total": result.Total,
		"page":  result.Page,
		"size":  result.Size,
	})
}

// Get handles GET /api/v1/teams/:teamId/main-items/:itemId
func (h *MainItemHandler) Get(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "itemId")
	if !ok {
		return
	}

	item, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	itemVO := vo.NewMainItemVO(item)

	// Fetch subItems summary
	subItems, _ := h.subItemRepo.ListByMainItem(c.Request.Context(), item.BizKey)

	apperrors.RespondOK(c, gin.H{
		"bizKey":          itemVO.BizKey,
		"teamKey":         itemVO.TeamKey,
		"code":            itemVO.Code,
		"title":           itemVO.Title,
		"itemDesc":        itemVO.ItemDesc,
		"priority":        itemVO.Priority,
		"proposerKey":     itemVO.ProposerKey,
		"assigneeKey":     itemVO.AssigneeKey,
		"planStartDate":   itemVO.PlanStartDate,
		"expectedEndDate": itemVO.ExpectedEndDate,
		"actualEndDate":   itemVO.ActualEndDate,
		"itemStatus":      itemVO.ItemStatus,
		"statusName":      itemVO.StatusName,
		"completion":      itemVO.Completion,
		"isKeyItem":       itemVO.IsKeyItem,
		"archivedAt":      itemVO.ArchivedAt,
		"createTime":      itemVO.CreateTime,
		"dbUpdateTime":    itemVO.DbUpdateTime,
		"subItems":        vo.NewSubItemSummaryVOs(subItems),
	})
}

// Update handles PUT /api/v1/teams/:teamId/main-items/:itemId
func (h *MainItemHandler) Update(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "itemId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	var req dto.MainItemUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	item, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	err = h.svc.Update(c.Request.Context(), teamBizKey, item.ID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch updated item for response
	updated, err := h.svc.Get(c.Request.Context(), item.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, vo.NewMainItemVO(updated))
}

// Archive handles POST /api/v1/teams/:teamId/main-items/:itemId/archive
func (h *MainItemHandler) Archive(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "itemId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	item, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	err = h.svc.Archive(c.Request.Context(), teamBizKey, item.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// ChangeStatus handles PUT /api/v1/teams/:teamId/main-items/:itemId/status
func (h *MainItemHandler) ChangeStatus(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "itemId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)
	callerID := middleware.GetUserID(c)

	var req dto.ChangeStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	record, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	item, err := h.svc.ChangeStatus(c.Request.Context(), teamBizKey, callerID, record.ID, req.Status)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, vo.NewMainItemVO(item))
}

// AvailableTransitions handles GET /api/v1/teams/:teamId/main-items/:itemId/available-transitions
func (h *MainItemHandler) AvailableTransitions(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "itemId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)
	callerID := middleware.GetUserID(c)

	record, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	transitions, err := h.svc.AvailableTransitions(c.Request.Context(), teamBizKey, callerID, record.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{"transitions": transitions})
}

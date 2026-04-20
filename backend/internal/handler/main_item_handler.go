package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
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

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": vo.NewMainItemVO(item)})
}

// List handles GET /api/v1/teams/:teamId/main-items
func (h *MainItemHandler) List(c *gin.Context) {
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
	_, page.Page, page.PageSize = dto.ApplyPaginationDefaults(page.Page, page.PageSize)

	result, err := h.svc.List(c.Request.Context(), teamID, filter, page)
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
	itemID, ok := parseItemID(c)
	if !ok {
		return
	}

	item, err := h.svc.Get(c.Request.Context(), itemID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	itemVO := vo.NewMainItemVO(item)

	// Fetch subItems summary
	subItems, _ := h.subItemRepo.ListByMainItem(c.Request.Context(), itemID)

	apperrors.RespondOK(c, gin.H{
		"id":              itemVO.ID,
		"teamId":          itemVO.TeamID,
		"code":            itemVO.Code,
		"title":           itemVO.Title,
		"priority":        itemVO.Priority,
		"proposerId":      itemVO.ProposerID,
		"assigneeId":      itemVO.AssigneeID,
		"startDate":       itemVO.StartDate,
		"expectedEndDate": itemVO.ExpectedEndDate,
		"actualEndDate":   itemVO.ActualEndDate,
		"status":          itemVO.Status,
		"completion":      itemVO.Completion,
		"isKeyItem":       itemVO.IsKeyItem,
		"archivedAt":      itemVO.ArchivedAt,
		"createdAt":       itemVO.CreatedAt,
		"updatedAt":       itemVO.UpdatedAt,
		"subItems":        vo.NewSubItemSummaryVOs(subItems),
	})
}

// Update handles PUT /api/v1/teams/:teamId/main-items/:itemId
func (h *MainItemHandler) Update(c *gin.Context) {
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

	apperrors.RespondOK(c, vo.NewMainItemVO(updated))
}

// Archive handles POST /api/v1/teams/:teamId/main-items/:itemId/archive
func (h *MainItemHandler) Archive(c *gin.Context) {
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

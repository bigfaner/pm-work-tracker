package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	pkgHandler "pm-work-tracker/backend/internal/pkg/handler"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
)

// MainItemHandler handles main item endpoints.
type MainItemHandler struct {
	svc          service.MainItemService
	subItemRepo  repository.SubItemRepo
	userRepo     repository.UserRepo
	milestoneSvc *milestoneEnricher
}

// milestoneEnricher wraps the milestoneRepo for enriching handler responses.
type milestoneEnricher struct {
	milestoneRepo repository.MilestoneRepo
}

// NewMainItemHandler creates a new MainItemHandler with service and repo dependencies.
func NewMainItemHandler(svc service.MainItemService, userRepo repository.UserRepo, subItemRepo repository.SubItemRepo, milestoneRepo repository.MilestoneRepo) *MainItemHandler {
	if svc == nil {
		panic("main_item_handler: mainItemService must not be nil")
	}
	if userRepo == nil {
		panic("main_item_handler: userRepo must not be nil")
	}
	if subItemRepo == nil {
		panic("main_item_handler: subItemRepo must not be nil")
	}
	return &MainItemHandler{svc: svc, userRepo: userRepo, subItemRepo: subItemRepo, milestoneSvc: &milestoneEnricher{milestoneRepo: milestoneRepo}}
}

// Create handles POST /api/v1/teams/:teamId/main-items
//
//nolint:dupl // similar request handling shape as SubItem.Create but different business logic
func (h *MainItemHandler) Create(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	pmBizKey := middleware.GetUserBizKey(c)

	var req dto.MainItemCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	item, err := h.svc.Create(c.Request.Context(), teamBizKey, pmBizKey, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	itemVO := vo.NewMainItemVO(item)
	itemVO.MilestoneName = h.milestoneSvc.resolveMilestoneName(c, item)

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": itemVO})
}

// List handles GET /api/v1/teams/:teamId/main-items
//
//nolint:dupl // similar request handling shape as MilestoneMap.List but different business logic
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

	// Batch milestone name enrichment
	milestoneNames := h.milestoneSvc.batchResolveMilestoneNamesFromSlice(c, result.Items)

	voItems := make([]vo.MainItemVO, 0, len(result.Items))
	for i := range result.Items {
		itemVO := vo.NewMainItemVO(&result.Items[i])
		if result.Items[i].MilestoneKey != nil {
			if name, ok := milestoneNames[*result.Items[i].MilestoneKey]; ok {
				itemVO.MilestoneName = name
			}
		}
		voItems = append(voItems, itemVO)
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
	itemVO.MilestoneName = h.milestoneSvc.resolveMilestoneName(c, item)

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
		"milestoneKey":    itemVO.MilestoneKey,
		"milestoneName":   itemVO.MilestoneName,
		"createTime":      itemVO.CreateTime,
		"dbUpdateTime":    itemVO.DbUpdateTime,
		"subItems":        vo.NewSubItemSummaryVOs(subItems),
	})
}

// Update handles PUT /api/v1/teams/:teamId/main-items/:itemId
//
//nolint:dupl // similar request handling shape as MilestoneMap.Update but different business logic
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

	itemVO := vo.NewMainItemVO(updated)
	itemVO.MilestoneName = h.milestoneSvc.resolveMilestoneName(c, updated)

	apperrors.RespondOK(c, itemVO)
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
	callerBizKey := middleware.GetUserBizKey(c)

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

	item, err := h.svc.ChangeStatus(c.Request.Context(), teamBizKey, callerBizKey, record.ID, req.Status)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	itemVO := vo.NewMainItemVO(item)
	itemVO.MilestoneName = h.milestoneSvc.resolveMilestoneName(c, item)

	apperrors.RespondOK(c, itemVO)
}

// AvailableTransitions handles GET /api/v1/teams/:teamId/main-items/:itemId/available-transitions
func (h *MainItemHandler) AvailableTransitions(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "itemId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)
	callerBizKey := middleware.GetUserBizKey(c)

	record, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	transitions, err := h.svc.AvailableTransitions(c.Request.Context(), teamBizKey, callerBizKey, record.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{"transitions": transitions})
}

// resolveMilestoneName resolves the milestone name for a single MainItem.
func (e *milestoneEnricher) resolveMilestoneName(c *gin.Context, item *model.MainItem) string {
	if item.MilestoneKey == nil {
		return ""
	}
	names := e.batchResolveMilestoneNamesFromPtrs(c, []*model.MainItem{item})
	return names[*item.MilestoneKey]
}

// batchResolveMilestoneNamesFromSlice enriches []model.MainItem (value slice) with milestone names.
func (e *milestoneEnricher) batchResolveMilestoneNamesFromSlice(c *gin.Context, items []model.MainItem) map[int64]string {
	// Collect non-nil milestoneKey values
	bizKeys := make([]int64, 0)
	for i := range items {
		if items[i].MilestoneKey != nil {
			bizKeys = append(bizKeys, *items[i].MilestoneKey)
		}
	}
	return e.resolveBizKeys(c, bizKeys)
}

// batchResolveMilestoneNamesFromPtrs enriches []*model.MainItem with milestone names.
func (e *milestoneEnricher) batchResolveMilestoneNamesFromPtrs(c *gin.Context, items []*model.MainItem) map[int64]string {
	bizKeys := make([]int64, 0)
	for _, item := range items {
		if item.MilestoneKey != nil {
			bizKeys = append(bizKeys, *item.MilestoneKey)
		}
	}
	return e.resolveBizKeys(c, bizKeys)
}

// resolveBizKeys performs batch lookup of milestone bizKeys and returns map[int64]string.
func (e *milestoneEnricher) resolveBizKeys(c *gin.Context, bizKeys []int64) map[int64]string {
	if len(bizKeys) == 0 {
		return nil
	}

	milestoneMap, err := e.milestoneRepo.FindByBizKeys(c.Request.Context(), bizKeys)
	if err != nil {
		result := make(map[int64]string, len(bizKeys))
		for _, bk := range bizKeys {
			result[bk] = "--"
		}
		return result
	}

	result := make(map[int64]string, len(bizKeys))
	for _, bk := range bizKeys {
		if m, ok := milestoneMap[bk]; ok {
			result[bk] = m.MilestoneName
		} else {
			result[bk] = "--"
		}
	}
	return result
}

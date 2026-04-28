package handler

import (
	"context"
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
	teamBizKey := middleware.GetTeamBizKey(c)
	userBizKey := middleware.GetUserBizKey(c)

	var req dto.SubmitItemPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	item, err := h.svc.Submit(c.Request.Context(), teamBizKey, userBizKey, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": buildItemPoolVOs([]model.ItemPool{*item}, h.userRepo, h.mainItemRepo, c)[0]})
}

// List handles GET /api/v1/teams/:teamId/item-pool
func (h *ItemPoolHandler) List(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)

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

	result, err := h.svc.List(c.Request.Context(), teamBizKey, filter, page)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	voItems := buildItemPoolVOs(result.Items, h.userRepo, h.mainItemRepo, c)
	apperrors.RespondOK(c, gin.H{
		"items": voItems,
		"total": result.Total,
		"page":  result.Page,
		"size":  result.Size,
	})
}

// Get handles GET /api/v1/teams/:teamId/item-pool/:poolId
func (h *ItemPoolHandler) Get(c *gin.Context) {
	poolID, ok := pkgHandler.ResolveBizKey(c, "poolId", func(ctx context.Context, bizKey int64) (uint, error) {
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

	item, err := h.svc.Get(c.Request.Context(), teamBizKey, poolID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildItemPoolVOs([]model.ItemPool{*item}, h.userRepo, h.mainItemRepo, c)[0])
}

// Assign handles POST /api/v1/teams/:teamId/item-pool/:poolId/assign
func (h *ItemPoolHandler) Assign(c *gin.Context) {
	poolID, ok := pkgHandler.ResolveBizKey(c, "poolId", func(ctx context.Context, bizKey int64) (uint, error) {
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

	var req dto.AssignItemPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.svc.Assign(c.Request.Context(), teamBizKey, pmBizKey, poolID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch the updated item to get the assignedSubId
	updated, err := h.svc.Get(c.Request.Context(), teamBizKey, poolID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"mainItemBizKey": updated.AssignedMainKey, "subItemBizKey": updated.AssignedSubKey}})
}

// ConvertToMain handles POST /api/v1/teams/:teamId/item-pool/:poolId/convert-to-main
func (h *ItemPoolHandler) ConvertToMain(c *gin.Context) {
	poolID, ok := pkgHandler.ResolveBizKey(c, "poolId", func(ctx context.Context, bizKey int64) (uint, error) {
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

	var req dto.ConvertToMainItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	mainItem, err := h.svc.ConvertToMain(c.Request.Context(), teamBizKey, pmBizKey, poolID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"mainItemBizKey": mainItem.BizKey}})
}

// Update handles PUT /api/v1/teams/:teamId/item-pool/:poolId
func (h *ItemPoolHandler) Update(c *gin.Context) {
	poolID, ok := pkgHandler.ResolveBizKey(c, "poolId", func(ctx context.Context, bizKey int64) (uint, error) {
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

	var req dto.UpdateItemPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	item, err := h.svc.Update(c.Request.Context(), teamBizKey, poolID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildItemPoolVOs([]model.ItemPool{*item}, h.userRepo, h.mainItemRepo, c)[0])
}

// Reject handles POST /api/v1/teams/:teamId/item-pool/:poolId/reject
func (h *ItemPoolHandler) Reject(c *gin.Context) {
	poolID, ok := pkgHandler.ResolveBizKey(c, "poolId", func(ctx context.Context, bizKey int64) (uint, error) {
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

	var req dto.RejectItemPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.svc.Reject(c.Request.Context(), teamBizKey, pmBizKey, poolID, req.Reason)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch the updated item for response
	updated, err := h.svc.Get(c.Request.Context(), teamBizKey, poolID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildItemPoolVOs([]model.ItemPool{*updated}, h.userRepo, h.mainItemRepo, c)[0])
}

// buildItemPoolVOs converts a slice of ItemPool to VOs using batch lookups (fixes N+1).
// Single-item callers pass a 1-element slice; the batch path has no N+1 overhead.
func buildItemPoolVOs(items []model.ItemPool, userRepo repository.UserRepo, mainItemRepo repository.MainItemRepo, c *gin.Context) []vo.ItemPoolVO {
	if len(items) == 0 {
		return []vo.ItemPoolVO{}
	}

	ctx := c.Request.Context()

	// Collect unique IDs
	submitterIDs := make(map[uint]struct{})
	mainItemBizKeys := make(map[int64]struct{})
	for i := range items {
		if items[i].SubmitterKey > 0 {
			submitterIDs[uint(items[i].SubmitterKey)] = struct{}{}
		}
		if items[i].AssignedMainKey != nil {
			mainItemBizKeys[*items[i].AssignedMainKey] = struct{}{}
		}
	}

	// Batch lookups
	userMap := make(map[uint]*model.User)
	if len(submitterIDs) > 0 {
		ids := mapKeysToSlice(submitterIDs)
		if m, err := userRepo.FindByIDs(ctx, ids); err == nil {
			userMap = m
		}
	}

	mainItemMap := make(map[int64]*model.MainItem)
	if len(mainItemBizKeys) > 0 {
		keys := int64MapKeysToSlice(mainItemBizKeys)
		if m, err := mainItemRepo.FindByBizKeys(ctx, keys); err == nil {
			mainItemMap = m
		}
	}

	// Build VOs from maps
	result := make([]vo.ItemPoolVO, 0, len(items))
	for i := range items {
		submitterName := ""
		if u, ok := userMap[uint(items[i].SubmitterKey)]; ok {
			submitterName = u.DisplayName
		}
		v := vo.NewItemPoolVO(&items[i], submitterName)
		if items[i].AssignedMainKey != nil {
			if mi, ok := mainItemMap[*items[i].AssignedMainKey]; ok {
				v.AssignedMainCode = mi.Code
				v.AssignedMainTitle = mi.Title
			}
		}
		result = append(result, v)
	}
	return result
}

// mapKeysToSlice extracts map keys to a slice for batch lookups.
func mapKeysToSlice(m map[uint]struct{}) []uint {
	ids := make([]uint, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	return ids
}

// int64MapKeysToSlice extracts int64 map keys to a slice for batch lookups.
func int64MapKeysToSlice(m map[int64]struct{}) []int64 {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

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

// MilestoneMapHandler handles milestone map endpoints.
type MilestoneMapHandler struct {
	svc           service.MilestoneMapService
	userRepo      repository.UserRepo
	milestoneRepo repository.MilestoneRepo
	mainItemRepo  repository.MainItemRepo
}

// NewMilestoneMapHandler creates a new MilestoneMapHandler with service and repo dependencies.
//
//nolint:dupl // constructor panic-on-nil pattern is identical across handlers by convention
func NewMilestoneMapHandler(svc service.MilestoneMapService, userRepo repository.UserRepo, milestoneRepo repository.MilestoneRepo, mainItemRepo repository.MainItemRepo) *MilestoneMapHandler {
	if svc == nil {
		panic("milestone_map_handler: milestoneMapService must not be nil")
	}
	if userRepo == nil {
		panic("milestone_map_handler: userRepo must not be nil")
	}
	if milestoneRepo == nil {
		panic("milestone_map_handler: milestoneRepo must not be nil")
	}
	if mainItemRepo == nil {
		panic("milestone_map_handler: mainItemRepo must not be nil")
	}
	return &MilestoneMapHandler{svc: svc, userRepo: userRepo, milestoneRepo: milestoneRepo, mainItemRepo: mainItemRepo}
}

// Create handles POST /api/v1/teams/:teamId/milestone-maps
func (h *MilestoneMapHandler) Create(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	userBizKey := middleware.GetUserBizKey(c)

	var req dto.MilestoneMapCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	m, err := h.svc.Create(c.Request.Context(), teamBizKey, userBizKey, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": buildMilestoneMapVO(m, h.userRepo, h.milestoneRepo, h.mainItemRepo, c)})
}

// List handles GET /api/v1/teams/:teamId/milestone-maps
func (h *MilestoneMapHandler) List(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)

	var filter dto.MilestoneMapFilter
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

	voItems := buildMilestoneMapVOs(result.Items, h.userRepo, h.milestoneRepo, h.mainItemRepo, c)
	apperrors.RespondOK(c, gin.H{
		"items": voItems,
		"total": result.Total,
		"page":  result.Page,
		"size":  result.Size,
	})
}

// Get handles GET /api/v1/teams/:teamId/milestone-maps/:mapId
func (h *MilestoneMapHandler) Get(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
	if !ok {
		return
	}

	m, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildMilestoneMapVO(m, h.userRepo, h.milestoneRepo, h.mainItemRepo, c))
}

// Update handles PUT /api/v1/teams/:teamId/milestone-maps/:mapId
func (h *MilestoneMapHandler) Update(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	var req dto.MilestoneMapUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	m, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	err = h.svc.Update(c.Request.Context(), teamBizKey, m.ID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch updated entity for response
	updated, err := h.svc.Get(c.Request.Context(), m.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildMilestoneMapVO(updated, h.userRepo, h.milestoneRepo, h.mainItemRepo, c))
}

// Delete handles DELETE /api/v1/teams/:teamId/milestone-maps/:mapId
//
//nolint:dupl // handler pattern is structurally identical to MilestoneHandler.Delete by convention
func (h *MilestoneMapHandler) Delete(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	m, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	err = h.svc.Delete(c.Request.Context(), teamBizKey, m.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{"message": "deleted"})
}

// ChangeStatus handles PUT /api/v1/teams/:teamId/milestone-maps/:mapId/status
func (h *MilestoneMapHandler) ChangeStatus(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	var req dto.ChangeStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	m, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	updated, err := h.svc.ChangeStatus(c.Request.Context(), teamBizKey, m.ID, req.Status)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildMilestoneMapVO(updated, h.userRepo, h.milestoneRepo, h.mainItemRepo, c))
}

// AvailableTransitions handles GET /api/v1/teams/:teamId/milestone-maps/:mapId/available-transitions
//
//nolint:dupl // handler pattern is structurally identical to MilestoneHandler.AvailableTransitions by convention
func (h *MilestoneMapHandler) AvailableTransitions(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	m, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	transitions, err := h.svc.AvailableTransitions(c.Request.Context(), teamBizKey, m.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{"transitions": transitions})
}

// buildMilestoneMapVOs converts a slice of MilestoneMap to VOs using batch lookups.
func buildMilestoneMapVOs(items []model.MilestoneMap, userRepo repository.UserRepo, milestoneRepo repository.MilestoneRepo, mainItemRepo repository.MainItemRepo, c *gin.Context) []vo.MilestoneMapVO {
	if len(items) == 0 {
		return []vo.MilestoneMapVO{}
	}

	result := make([]vo.MilestoneMapVO, 0, len(items))
	for i := range items {
		result = append(result, buildMilestoneMapVO(&items[i], userRepo, milestoneRepo, mainItemRepo, c))
	}
	return result
}

// buildMilestoneMapVO converts a single MilestoneMap to a VO with user name enrichment and computed fields.
func buildMilestoneMapVO(m *model.MilestoneMap, userRepo repository.UserRepo, milestoneRepo repository.MilestoneRepo, mainItemRepo repository.MainItemRepo, c *gin.Context) vo.MilestoneMapVO {
	ctx := c.Request.Context()
	v := vo.NewMilestoneMapVO(m)

	// Enrich creator and assignee names
	if m.CreatorKey > 0 {
		if u, err := userRepo.FindByBizKey(ctx, m.CreatorKey); err == nil {
			v.CreatorName = u.DisplayName
		}
	}
	if m.AssigneeKey > 0 {
		if u, err := userRepo.FindByBizKey(ctx, m.AssigneeKey); err == nil {
			v.AssigneeName = u.DisplayName
		}
	}

	// Enrich computed fields: milestone count, item count, overall progress
	milestones, err := milestoneRepo.ListByMap(ctx, m.BizKey)
	if err == nil {
		v.MilestoneCount = len(milestones)
		var totalCompletion float64
		var itemCount float64
		for _, ms := range milestones {
			items, err := mainItemRepo.FindByMilestoneKey(ctx, ms.BizKey)
			if err != nil || len(items) == 0 {
				continue
			}
			for _, item := range items {
				totalCompletion += item.Completion
				itemCount++
			}
		}
		v.ItemCount = int(itemCount)
		if itemCount > 0 {
			v.OverallProgress = totalCompletion / itemCount
		}
	}

	return v
}

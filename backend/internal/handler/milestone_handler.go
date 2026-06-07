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

// MilestoneHandler handles milestone endpoints.
type MilestoneHandler struct {
	svc          service.MilestoneService
	mainItemRepo repository.MainItemRepo
}

// NewMilestoneHandler creates a new MilestoneHandler with service and repo dependencies.
//
//nolint:dupl // constructor panic-on-nil pattern is identical across handlers by convention
func NewMilestoneHandler(svc service.MilestoneService, mainItemRepo repository.MainItemRepo) *MilestoneHandler {
	if svc == nil {
		panic("milestone_handler: milestoneService must not be nil")
	}
	if mainItemRepo == nil {
		panic("milestone_handler: mainItemRepo must not be nil")
	}
	return &MilestoneHandler{svc: svc, mainItemRepo: mainItemRepo}
}

// Create handles POST /api/v1/teams/:teamId/milestone-maps/:mapId/milestones
func (h *MilestoneHandler) Create(c *gin.Context) {
	mapBizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	var req dto.MilestoneCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	ms, err := h.svc.Create(c.Request.Context(), teamBizKey, mapBizKey, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": buildMilestoneVO(ms, h.mainItemRepo, c)})
}

// ListByMap handles GET /api/v1/teams/:teamId/milestone-maps/:mapId/milestones
func (h *MilestoneHandler) ListByMap(c *gin.Context) {
	mapBizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
	if !ok {
		return
	}

	milestones, err := h.svc.ListByMap(c.Request.Context(), mapBizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	voItems := buildMilestoneVOs(milestones, h.mainItemRepo, c)
	apperrors.RespondOK(c, gin.H{
		"items": voItems,
		"total": len(voItems),
	})
}

// ListByTeam handles GET /api/v1/teams/:teamId/milestones
func (h *MilestoneHandler) ListByTeam(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)

	var filter dto.MilestoneTeamFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	milestones, err := h.svc.ListByTeam(c.Request.Context(), teamBizKey, filter)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	voItems := buildMilestoneVOs(milestones, h.mainItemRepo, c)
	apperrors.RespondOK(c, gin.H{
		"items": voItems,
		"total": len(voItems),
	})
}

// Get handles GET /api/v1/teams/:teamId/milestones/:milestoneId
func (h *MilestoneHandler) Get(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "milestoneId")
	if !ok {
		return
	}

	ms, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildMilestoneVO(ms, h.mainItemRepo, c))
}

// Update handles PUT /api/v1/teams/:teamId/milestones/:milestoneId
func (h *MilestoneHandler) Update(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "milestoneId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	var req dto.MilestoneUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	ms, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	err = h.svc.Update(c.Request.Context(), teamBizKey, ms.ID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch updated entity for response
	updated, err := h.svc.Get(c.Request.Context(), ms.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildMilestoneVO(updated, h.mainItemRepo, c))
}

// Delete handles DELETE /api/v1/teams/:teamId/milestones/:milestoneId
//
//nolint:dupl // handler pattern is structurally identical to MilestoneMapHandler.Delete by convention
func (h *MilestoneHandler) Delete(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "milestoneId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	ms, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	err = h.svc.Delete(c.Request.Context(), teamBizKey, ms.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{"message": "deleted"})
}

// ChangeStatus handles PUT /api/v1/teams/:teamId/milestones/:milestoneId/status
func (h *MilestoneHandler) ChangeStatus(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "milestoneId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	var req dto.ChangeStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	ms, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	updated, err := h.svc.ChangeStatus(c.Request.Context(), teamBizKey, ms.ID, req.Status)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, buildMilestoneVO(updated, h.mainItemRepo, c))
}

// AvailableTransitions handles GET /api/v1/teams/:teamId/milestones/:milestoneId/available-transitions
//
//nolint:dupl // handler pattern is structurally identical to MilestoneMapHandler.AvailableTransitions by convention
func (h *MilestoneHandler) AvailableTransitions(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "milestoneId")
	if !ok {
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	ms, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	transitions, err := h.svc.AvailableTransitions(c.Request.Context(), teamBizKey, ms.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{"transitions": transitions})
}

// buildMilestoneVOs converts a slice of Milestone to VOs with computed fields.
func buildMilestoneVOs(items []model.Milestone, mainItemRepo repository.MainItemRepo, c *gin.Context) []vo.MilestoneVO {
	if len(items) == 0 {
		return []vo.MilestoneVO{}
	}

	result := make([]vo.MilestoneVO, 0, len(items))
	for i := range items {
		result = append(result, buildMilestoneVO(&items[i], mainItemRepo, c))
	}
	return result
}

// buildMilestoneVO converts a single Milestone to a VO with computed fields (completion and related MI count).
func buildMilestoneVO(m *model.Milestone, mainItemRepo repository.MainItemRepo, c *gin.Context) vo.MilestoneVO {
	ctx := c.Request.Context()
	v := vo.NewMilestoneVO(m)

	// Enrich computed fields: completion (average MI completion) and related MI count
	items, err := mainItemRepo.FindByMilestoneKey(ctx, m.BizKey)
	if err == nil && len(items) > 0 {
		var total float64
		for _, item := range items {
			total += item.Completion
		}
		v.Completion = total / float64(len(items))
		v.RelatedMICount = len(items)
	}

	return v
}

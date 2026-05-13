package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	pkgHandler "pm-work-tracker/backend/internal/pkg/handler"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
)

// MilestoneHandler handles milestone endpoints.
type MilestoneHandler struct {
	svc service.MilestoneService
}

// NewMilestoneHandler creates a new MilestoneHandler with service dependencies.
func NewMilestoneHandler(svc service.MilestoneService) *MilestoneHandler {
	if svc == nil {
		panic("milestone_handler: milestoneService must not be nil")
	}
	return &MilestoneHandler{svc: svc}
}

// Create handles POST /api/v1/teams/:teamId/milestone-maps/:mapId/milestones
func (h *MilestoneHandler) Create(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	mapBizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
	if !ok {
		return
	}

	var req dto.MilestoneCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	m, err := h.svc.Create(c.Request.Context(), teamBizKey, mapBizKey, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondCreated(c, vo.NewMilestoneVO(m))
}

// ListByMap handles GET /api/v1/teams/:teamId/milestone-maps/:mapId/milestones
func (h *MilestoneHandler) ListByMap(c *gin.Context) {
	mapBizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
	if !ok {
		return
	}

	items, err := h.svc.ListByMap(c.Request.Context(), mapBizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	voItems := make([]vo.MilestoneVO, 0, len(items))
	for i := range items {
		voItems = append(voItems, vo.NewMilestoneVO(&items[i]))
	}
	apperrors.RespondOK(c, gin.H{
		"items": voItems,
		"total": len(voItems),
	})
}

// ListByTeam handles GET /api/v1/teams/:teamId/milestones
func (h *MilestoneHandler) ListByTeam(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)

	excludeCancelled := true
	if v := c.Query("excludeCancelled"); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err == nil {
			excludeCancelled = parsed
		}
	}

	items, err := h.svc.ListByTeam(c.Request.Context(), teamBizKey, excludeCancelled)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	voItems := make([]vo.MilestoneVO, 0, len(items))
	for i := range items {
		voItems = append(voItems, vo.NewMilestoneVO(&items[i]))
	}
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

	m, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, vo.NewMilestoneVO(m))
}

// Update handles PUT /api/v1/teams/:teamId/milestones/:milestoneId
func (h *MilestoneHandler) Update(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "milestoneId")
	if !ok {
		return
	}

	var req dto.MilestoneUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	// Resolve bizKey to internal ID
	record, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	updated, err := h.svc.Update(c.Request.Context(), record.ID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, vo.NewMilestoneVO(updated))
}

// Delete handles DELETE /api/v1/teams/:teamId/milestones/:milestoneId
//
//nolint:dupl // similar to MilestoneMapHandler.Delete but different entity/service
func (h *MilestoneHandler) Delete(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "milestoneId")
	if !ok {
		return
	}

	// Resolve bizKey to internal ID
	record, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), record.ID); err != nil {
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

	var req dto.ChangeStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	// Resolve bizKey to internal ID
	record, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	updated, err := h.svc.ChangeStatus(c.Request.Context(), record.ID, req.Status)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, vo.NewMilestoneVO(updated))
}

// AvailableTransitions handles GET /api/v1/teams/:teamId/milestones/:milestoneId/available-transitions
func (h *MilestoneHandler) AvailableTransitions(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "milestoneId")
	if !ok {
		return
	}

	// Resolve bizKey to internal ID
	record, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	transitions, err := h.svc.AvailableTransitions(c.Request.Context(), record.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{"transitions": transitions})
}

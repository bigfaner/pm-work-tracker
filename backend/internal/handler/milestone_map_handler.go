package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	pkgHandler "pm-work-tracker/backend/internal/pkg/handler"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
)

// MilestoneMapHandler handles milestone map endpoints.
type MilestoneMapHandler struct {
	svc service.MilestoneMapService
}

// NewMilestoneMapHandler creates a new MilestoneMapHandler with service dependency.
func NewMilestoneMapHandler(svc service.MilestoneMapService) *MilestoneMapHandler {
	if svc == nil {
		panic("milestone_map_handler: milestoneMapService must not be nil")
	}
	return &MilestoneMapHandler{svc: svc}
}

// Create handles POST /api/v1/teams/:teamId/milestone-maps
func (h *MilestoneMapHandler) Create(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)

	var req dto.MilestoneMapCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	m, err := h.svc.Create(c.Request.Context(), teamBizKey, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": vo.NewMilestoneMapVO(m)})
}

// List handles GET /api/v1/teams/:teamId/milestone-maps
//
//nolint:dupl // similar request handling shape as MainItem.List but different business logic
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

	voItems := make([]vo.MilestoneMapVO, 0, len(result.Items))
	for i := range result.Items {
		voItems = append(voItems, vo.NewMilestoneMapVO(&result.Items[i]))
	}
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

	apperrors.RespondOK(c, vo.NewMilestoneMapVO(m))
}

// Update handles PUT /api/v1/teams/:teamId/milestone-maps/:mapId
//
//nolint:dupl // similar request handling shape as MainItem.Update but different business logic
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

	record, err := h.svc.GetByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	err = h.svc.Update(c.Request.Context(), teamBizKey, record.ID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	// Fetch updated record for response
	updated, err := h.svc.Get(c.Request.Context(), record.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, vo.NewMilestoneMapVO(updated))
}

// Delete handles DELETE /api/v1/teams/:teamId/milestone-maps/:mapId
func (h *MilestoneMapHandler) Delete(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
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

// ChangeStatus handles PUT /api/v1/teams/:teamId/milestone-maps/:mapId/status
func (h *MilestoneMapHandler) ChangeStatus(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
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

	apperrors.RespondOK(c, vo.NewMilestoneMapVO(updated))
}

// AvailableTransitions handles GET /api/v1/teams/:teamId/milestone-maps/:mapId/available-transitions
func (h *MilestoneMapHandler) AvailableTransitions(c *gin.Context) {
	bizKey, ok := pkgHandler.ParseBizKeyParam(c, "mapId")
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

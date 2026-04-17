package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/service"
)

// ViewHandler handles view endpoints (weekly, gantt, table).
type ViewHandler struct {
	viewSvc service.ViewService
}

// NewViewHandler creates a new ViewHandler (stub, for router setup before service is ready).
func NewViewHandler() *ViewHandler {
	return &ViewHandler{}
}

// NewViewHandlerWithDeps creates a new ViewHandler with service dependencies.
func NewViewHandlerWithDeps(viewSvc service.ViewService) *ViewHandler {
	return &ViewHandler{viewSvc: viewSvc}
}

// Weekly handles GET /api/v1/teams/:teamId/views/weekly
func (h *ViewHandler) Weekly(c *gin.Context) {
	if h.viewSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	weekStartStr := c.Query("weekStart")
	if weekStartStr == "" {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	weekStart, err := time.Parse("2006-01-02", weekStartStr)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	teamID := middleware.GetTeamID(c)

	result, err := h.viewSvc.WeeklyView(c.Request.Context(), teamID, weekStart)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, result)
}

// Gantt handles GET /api/v1/teams/:teamId/views/gantt
func (h *ViewHandler) Gantt(c *gin.Context) {
	if h.viewSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	var filter dto.GanttFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	teamID := middleware.GetTeamID(c)

	result, err := h.viewSvc.GanttView(c.Request.Context(), teamID, filter)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, result)
}

// Table handles GET /api/v1/teams/:teamId/views/table
func (h *ViewHandler) Table(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// ExportTable handles GET /api/v1/teams/:teamId/views/table/export
func (h *ViewHandler) ExportTable(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

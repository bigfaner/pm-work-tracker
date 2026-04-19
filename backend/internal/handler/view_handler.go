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
	// Normalize to local timezone so weekday and future-check match user's perspective
	loc := time.Now().Location()
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, loc)

	if weekStart.Weekday() != time.Monday {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	// Validate weekStart is not in the future
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	if weekStart.After(today) {
		apperrors.RespondError(c, apperrors.ErrFutureWeekNotAllowed)
		return
	}

	teamID := middleware.GetTeamID(c)

	result, err := h.viewSvc.WeeklyComparison(c.Request.Context(), teamID, weekStart)
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
	if h.viewSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	var filter dto.TableFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	page := dto.Pagination{Page: 1, PageSize: 50}
	if err := c.ShouldBindQuery(&page); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}
	if page.Page < 1 {
		page.Page = 1
	}
	if page.PageSize < 1 {
		page.PageSize = 50
	}

	teamID := middleware.GetTeamID(c)

	result, err := h.viewSvc.TableView(c.Request.Context(), teamID, filter, page)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, result)
}

// ExportTable handles GET /api/v1/teams/:teamId/views/table/export
func (h *ViewHandler) ExportTable(c *gin.Context) {
	if h.viewSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	var filter dto.TableFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	teamID := middleware.GetTeamID(c)

	csvBytes, err := h.viewSvc.TableExportCSV(c.Request.Context(), teamID, filter)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.Header("Content-Disposition", `attachment; filename="items-export.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}

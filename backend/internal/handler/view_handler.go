package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/dates"
	"pm-work-tracker/backend/internal/service"
)

// ViewHandler handles view endpoints (weekly, gantt, table).
type ViewHandler struct {
	viewSvc service.ViewService
}

// NewViewHandler creates a new ViewHandler with service dependency.
func NewViewHandler(viewSvc service.ViewService) *ViewHandler {
	if viewSvc == nil {
		panic("view_handler: viewService must not be nil")
	}
	return &ViewHandler{viewSvc: viewSvc}
}

// Weekly handles GET /api/v1/teams/:teamId/views/weekly
func (h *ViewHandler) Weekly(c *gin.Context) {
	weekStartStr := c.Query("weekStart")
	if weekStartStr == "" {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	weekStart, err := dates.ParseDate(weekStartStr)
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

	teamBizKey := middleware.GetTeamBizKey(c)

	result, err := h.viewSvc.WeeklyComparison(c.Request.Context(), teamBizKey, weekStart)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, result)
}

// Gantt handles GET /api/v1/teams/:teamId/views/gantt
func (h *ViewHandler) Gantt(c *gin.Context) {
	var filter dto.GanttFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	result, err := h.viewSvc.GanttView(c.Request.Context(), teamBizKey, filter)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, result)
}

// Table handles GET /api/v1/teams/:teamId/views/table
func (h *ViewHandler) Table(c *gin.Context) {
	var filter dto.TableFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	var page dto.Pagination
	_ = c.ShouldBindQuery(&page)
	_, page.Page, page.PageSize = dto.ApplyPaginationWithDefault(page.Page, page.PageSize, 50)

	teamBizKey := middleware.GetTeamBizKey(c)

	result, err := h.viewSvc.TableView(c.Request.Context(), teamBizKey, filter, page)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, result)
}

// ExportTable handles GET /api/v1/teams/:teamId/views/table/export
func (h *ViewHandler) ExportTable(c *gin.Context) {
	var filter dto.TableFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	teamBizKey := middleware.GetTeamBizKey(c)

	csvBytes, err := h.viewSvc.TableExportCSV(c.Request.Context(), teamBizKey, filter)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.Header("Content-Disposition", `attachment; filename="items-export.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}

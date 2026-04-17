package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/middleware"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/service"
)

// ReportHandler handles report endpoints.
type ReportHandler struct {
	svc service.ReportService
}

// NewReportHandler creates a new ReportHandler (stub, for router setup before service is ready).
func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

// NewReportHandlerWithDeps creates a new ReportHandler with service dependencies.
func NewReportHandlerWithDeps(svc service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// WeeklyPreview handles GET /api/v1/teams/:teamId/reports/weekly/preview
func (h *ReportHandler) WeeklyPreview(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	weekStart, ok := parseWeekStart(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)

	result, err := h.svc.Preview(c.Request.Context(), teamID, weekStart)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, result)
}

// WeeklyExport handles GET /api/v1/teams/:teamId/reports/weekly/export
func (h *ReportHandler) WeeklyExport(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	weekStart, ok := parseWeekStart(c)
	if !ok {
		return
	}

	teamID := middleware.GetTeamID(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	mdBytes, err := h.svc.ExportMarkdown(ctx, teamID, weekStart)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	isoYear, isoWeek := weekStart.ISOWeek()
	filename := fmt.Sprintf("weekly-report-%d-W%02d.md", isoYear, isoWeek)

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "text/markdown", mdBytes)
}

// parseWeekStart validates the weekStart query param as a Monday ISO8601 date.
// Returns the parsed time and true on success; responds with error and false on failure.
func parseWeekStart(c *gin.Context) (time.Time, bool) {
	weekStartStr := c.Query("weekStart")
	if weekStartStr == "" {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return time.Time{}, false
	}

	weekStart, err := time.Parse("2006-01-02", weekStartStr)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return time.Time{}, false
	}

	if weekStart.Weekday() != time.Monday {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return time.Time{}, false
	}

	return weekStart, true
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReportHandler handles report endpoints.
type ReportHandler struct {
	// Will be wired in a later task.
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

// WeeklyPreview handles GET /api/v1/teams/:teamId/reports/weekly/preview
func (h *ReportHandler) WeeklyPreview(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// WeeklyExport handles GET /api/v1/teams/:teamId/reports/weekly/export
func (h *ReportHandler) WeeklyExport(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

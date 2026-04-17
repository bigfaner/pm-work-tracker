package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ViewHandler handles view endpoints (weekly, gantt, table).
type ViewHandler struct {
	// Will be wired in a later task.
}

// NewViewHandler creates a new ViewHandler.
func NewViewHandler() *ViewHandler {
	return &ViewHandler{}
}

// Weekly handles GET /api/v1/teams/:teamId/views/weekly
func (h *ViewHandler) Weekly(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Gantt handles GET /api/v1/teams/:teamId/views/gantt
func (h *ViewHandler) Gantt(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Table handles GET /api/v1/teams/:teamId/views/table
func (h *ViewHandler) Table(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// ExportTable handles GET /api/v1/teams/:teamId/views/table/export
func (h *ViewHandler) ExportTable(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProgressHandler handles progress record endpoints.
type ProgressHandler struct {
	// Will be wired to service.ProgressService in a later task.
}

// NewProgressHandler creates a new ProgressHandler.
func NewProgressHandler() *ProgressHandler {
	return &ProgressHandler{}
}

// Append handles POST /api/v1/teams/:teamId/sub-items/:itemId/progress
func (h *ProgressHandler) Append(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// List handles GET /api/v1/teams/:teamId/sub-items/:itemId/progress
func (h *ProgressHandler) List(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// CorrectCompletion handles PATCH /api/v1/teams/:teamId/progress/:recordId/completion
func (h *ProgressHandler) CorrectCompletion(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SubItemHandler handles sub item endpoints.
type SubItemHandler struct {
	// Will be wired to service.SubItemService in a later task.
}

// NewSubItemHandler creates a new SubItemHandler.
func NewSubItemHandler() *SubItemHandler {
	return &SubItemHandler{}
}

// Create handles POST /api/v1/teams/:teamId/main-items/:mainId/sub-items
func (h *SubItemHandler) Create(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// List handles GET /api/v1/teams/:teamId/main-items/:mainId/sub-items
func (h *SubItemHandler) List(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Get handles GET /api/v1/teams/:teamId/sub-items/:itemId
func (h *SubItemHandler) Get(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Update handles PUT /api/v1/teams/:teamId/sub-items/:itemId
func (h *SubItemHandler) Update(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// ChangeStatus handles PUT /api/v1/teams/:teamId/sub-items/:itemId/status
func (h *SubItemHandler) ChangeStatus(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Assign handles PUT /api/v1/teams/:teamId/sub-items/:itemId/assignee
func (h *SubItemHandler) Assign(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ItemPoolHandler handles item pool endpoints.
type ItemPoolHandler struct {
	// Will be wired to service.ItemPoolService in a later task.
}

// NewItemPoolHandler creates a new ItemPoolHandler.
func NewItemPoolHandler() *ItemPoolHandler {
	return &ItemPoolHandler{}
}

// Submit handles POST /api/v1/teams/:teamId/item-pool
func (h *ItemPoolHandler) Submit(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// List handles GET /api/v1/teams/:teamId/item-pool
func (h *ItemPoolHandler) List(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Get handles GET /api/v1/teams/:teamId/item-pool/:poolId
func (h *ItemPoolHandler) Get(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Assign handles POST /api/v1/teams/:teamId/item-pool/:poolId/assign
func (h *ItemPoolHandler) Assign(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Reject handles POST /api/v1/teams/:teamId/item-pool/:poolId/reject
func (h *ItemPoolHandler) Reject(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

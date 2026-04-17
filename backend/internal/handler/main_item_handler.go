package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MainItemHandler handles main item endpoints.
type MainItemHandler struct {
	// Will be wired to service.MainItemService in a later task.
}

// NewMainItemHandler creates a new MainItemHandler.
func NewMainItemHandler() *MainItemHandler {
	return &MainItemHandler{}
}

// Create handles POST /api/v1/teams/:teamId/main-items
func (h *MainItemHandler) Create(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// List handles GET /api/v1/teams/:teamId/main-items
func (h *MainItemHandler) List(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Get handles GET /api/v1/teams/:teamId/main-items/:itemId
func (h *MainItemHandler) Get(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Update handles PUT /api/v1/teams/:teamId/main-items/:itemId
func (h *MainItemHandler) Update(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Archive handles POST /api/v1/teams/:teamId/main-items/:itemId/archive
func (h *MainItemHandler) Archive(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

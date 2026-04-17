package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminHandler handles admin endpoints.
type AdminHandler struct {
	// Will be wired in a later task.
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// ListUsers handles GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// UpdateCanCreateTeam handles PUT /api/v1/admin/users/:userId/can-create-team
func (h *AdminHandler) UpdateCanCreateTeam(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// ListTeams handles GET /api/v1/admin/teams
func (h *AdminHandler) ListTeams(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TeamHandler handles team endpoints.
type TeamHandler struct {
	// Will be wired to service.TeamService in a later task.
}

// NewTeamHandler creates a new TeamHandler.
func NewTeamHandler() *TeamHandler {
	return &TeamHandler{}
}

// Create handles POST /api/v1/teams
func (h *TeamHandler) Create(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// List handles GET /api/v1/teams
func (h *TeamHandler) List(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Get handles GET /api/v1/teams/:teamId
func (h *TeamHandler) Get(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Update handles PUT /api/v1/teams/:teamId
func (h *TeamHandler) Update(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// Disband handles DELETE /api/v1/teams/:teamId
func (h *TeamHandler) Disband(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// ListMembers handles GET /api/v1/teams/:teamId/members
func (h *TeamHandler) ListMembers(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// InviteMember handles POST /api/v1/teams/:teamId/members
func (h *TeamHandler) InviteMember(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// RemoveMember handles DELETE /api/v1/teams/:teamId/members/:userId
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

// TransferPM handles PUT /api/v1/teams/:teamId/pm
func (h *TeamHandler) TransferPM(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
}

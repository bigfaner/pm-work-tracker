package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
)

// TeamHandler handles team endpoints.
type TeamHandler struct {
	teamSvc  service.TeamService
	userRepo repository.UserRepo
}

// NewTeamHandler creates a new TeamHandler (stub, for router setup before service is ready).
func NewTeamHandler() *TeamHandler {
	return &TeamHandler{}
}

// NewTeamHandlerWithDeps creates a new TeamHandler with service and repo dependencies.
func NewTeamHandlerWithDeps(teamSvc service.TeamService, userRepo repository.UserRepo) *TeamHandler {
	return &TeamHandler{teamSvc: teamSvc, userRepo: userRepo}
}

// Create handles POST /api/v1/teams
func (h *TeamHandler) Create(c *gin.Context) {
	if h.teamSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)

	// SuperAdmin bypasses CanCreateTeam check
	if userRole != "superadmin" {
		user, err := h.userRepo.FindByID(c.Request.Context(), userID)
		if err != nil || !user.CanCreateTeam {
			apperrors.RespondError(c, apperrors.ErrForbidden)
			return
		}
	}

	var req dto.CreateTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	team, err := h.teamSvc.CreateTeam(c.Request.Context(), userID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": teamToDTO(team)})
}

// List handles GET /api/v1/teams
func (h *TeamHandler) List(c *gin.Context) {
	if h.teamSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	isSuperAdmin := userRole == "superadmin"

	teams, err := h.teamSvc.ListTeams(c.Request.Context(), userID, isSuperAdmin)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	dtos := make([]gin.H, len(teams))
	for i, t := range teams {
		dtos[i] = teamToDTO(t)
	}
	apperrors.RespondOK(c, dtos)
}

// Get handles GET /api/v1/teams/:teamId
func (h *TeamHandler) Get(c *gin.Context) {
	if h.teamSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)

	detail, err := h.teamSvc.GetTeamDetail(c.Request.Context(), teamID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, detail)
}

// Update handles PUT /api/v1/teams/:teamId
func (h *TeamHandler) Update(c *gin.Context) {
	if h.teamSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	var req dto.UpdateTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	team, err := h.teamSvc.UpdateTeam(c.Request.Context(), pmID, teamID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, teamToDTO(team))
}

// Disband handles DELETE /api/v1/teams/:teamId
func (h *TeamHandler) Disband(c *gin.Context) {
	if h.teamSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)
	callerID := middleware.GetUserID(c)

	var req dto.DisbandTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.teamSvc.DisbandTeam(c.Request.Context(), callerID, teamID, req.ConfirmName)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// ListMembers handles GET /api/v1/teams/:teamId/members
func (h *TeamHandler) ListMembers(c *gin.Context) {
	if h.teamSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)

	members, err := h.teamSvc.ListMembers(c.Request.Context(), teamID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, members)
}

// InviteMember handles POST /api/v1/teams/:teamId/members
func (h *TeamHandler) InviteMember(c *gin.Context) {
	if h.teamSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	var req dto.InviteMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.teamSvc.InviteMember(c.Request.Context(), pmID, teamID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// RemoveMember handles DELETE /api/v1/teams/:teamId/members/:userId
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	if h.teamSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	targetUserIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err = h.teamSvc.RemoveMember(c.Request.Context(), pmID, teamID, uint(targetUserID))
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// TransferPM handles PUT /api/v1/teams/:teamId/pm
func (h *TeamHandler) TransferPM(c *gin.Context) {
	if h.teamSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"code": "NOT_IMPLEMENTED", "message": "not implemented"})
		return
	}

	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	var req dto.TransferPMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	err := h.teamSvc.TransferPM(c.Request.Context(), pmID, teamID, req.NewPmUserID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// teamToDTO converts a model.Team to a response map.
func teamToDTO(team *model.Team) gin.H {
	return gin.H{
		"id":          team.ID,
		"name":        team.Name,
		"description": team.Description,
		"pmId":        team.PmID,
		"createdAt":   team.CreatedAt,
		"updatedAt":   team.UpdatedAt,
	}
}

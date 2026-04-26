package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
)

// TeamHandler handles team endpoints.
type TeamHandler struct {
	teamSvc  service.TeamService
	userRepo repository.UserRepo
}

// NewTeamHandler creates a new TeamHandler with service and repo dependencies.
func NewTeamHandler(teamSvc service.TeamService, userRepo repository.UserRepo) *TeamHandler {
	if teamSvc == nil {
		panic("team_handler: teamService must not be nil")
	}
	if userRepo == nil {
		panic("team_handler: userRepo must not be nil")
	}
	return &TeamHandler{teamSvc: teamSvc, userRepo: userRepo}
}

// Create handles POST /api/v1/teams
// Permission check (team:create) is handled by RequirePermission middleware.
func (h *TeamHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

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
	userID := middleware.GetUserID(c)
	isSuperAdmin := middleware.IsSuperAdmin(c)
	search := c.Query("search")
	page, pageSize := parsePagination(c, 20)

	teams, total, err := h.teamSvc.ListTeams(c.Request.Context(), userID, isSuperAdmin, search, page, pageSize)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, gin.H{
		"items":    teams,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Get handles GET /api/v1/teams/:teamId
func (h *TeamHandler) Get(c *gin.Context) {
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
	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	targetUserID, ok := h.resolveUserBizKey(c)
	if !ok {
		return
	}

	err := h.teamSvc.RemoveMember(c.Request.Context(), pmID, teamID, targetUserID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// UpdateMemberRole handles PUT /api/v1/teams/:teamId/members/:userId/role
func (h *TeamHandler) UpdateMemberRole(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	pmID := middleware.GetUserID(c)

	targetUserID, ok := h.resolveUserBizKey(c)
	if !ok {
		return
	}

	var req dto.UpdateMemberRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	roleKey, _ := pkg.ParseID(req.RoleKey)
	if err := h.teamSvc.UpdateMemberRole(c.Request.Context(), pmID, teamID, targetUserID, uint(roleKey)); err != nil {
	apperrors.RespondError(c, err)
	return
	}

	apperrors.RespondOK(c, nil)
}

// TransferPM handles PUT /api/v1/teams/:teamId/pm
func (h *TeamHandler) TransferPM(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	callerID := middleware.GetUserID(c)

	var req dto.TransferPMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	// SuperAdmin is not the team PM, so fetch the actual PM ID to pass the ownership check.
	pmID := callerID
	if middleware.IsSuperAdmin(c) {
		team, err := h.teamSvc.GetTeam(c.Request.Context(), teamID)
		if err != nil {
			apperrors.RespondError(c, err)
			return
		}
		pmID = uint(team.PmKey)
	}

	err := h.teamSvc.TransferPM(c.Request.Context(), pmID, teamID, func() uint { v, _ := pkg.ParseID(req.NewPmUserKey); return uint(v) }())
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// SearchUsers handles GET /api/v1/teams/:teamId/search-users
func (h *TeamHandler) SearchUsers(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	search := c.Query("search")

	users, err := h.teamSvc.SearchAvailableUsers(c.Request.Context(), teamID, search)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, users)
}

// teamToDTO converts a model.Team to a response map.
func teamToDTO(team *model.Team) gin.H {
	return gin.H{
		"bizKey":      pkg.FormatID(team.BizKey),
		"name":        team.TeamName,
		"description": team.TeamDesc,
		"code":        team.Code,
		"pmKey":       pkg.FormatID(team.PmKey),
		"createdAt":   team.CreateTime,
		"updatedAt":   team.DbUpdateTime,
	}
}

// resolveUserBizKey parses the userId path param as a bizKey and resolves it to an internal uint ID.
func (h *TeamHandler) resolveUserBizKey(c *gin.Context) (uint, bool) {
	idStr := c.Param("userId")
	bizKey, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return 0, false
	}
	user, err := h.userRepo.FindByBizKey(c.Request.Context(), bizKey)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrItemNotFound)
		return 0, false
	}
	return user.ID, true
}

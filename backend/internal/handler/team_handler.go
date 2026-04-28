package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/pkg"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	pkgHandler "pm-work-tracker/backend/internal/pkg/handler"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
	"pm-work-tracker/backend/internal/vo"
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

	apperrors.RespondCreated(c, vo.NewTeamVO(team))
}

// List handles GET /api/v1/teams
func (h *TeamHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	isSuperAdmin := middleware.IsSuperAdmin(c)
	search := c.Query("search")
	page, pageSize := parsePageParams(c)
	_, page, pageSize = dto.ApplyPaginationDefaults(page, pageSize)

	teams, total, err := h.teamSvc.ListTeams(c.Request.Context(), userID, isSuperAdmin, search, page, pageSize)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, dto.TeamListPage{
		Items:    teams,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Get handles GET /api/v1/teams/:teamId
func (h *TeamHandler) Get(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	teamID := uint(teamBizKey)

	detail, err := h.teamSvc.GetTeamDetail(c.Request.Context(), teamID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, detail)
}

// Update handles PUT /api/v1/teams/:teamId
func (h *TeamHandler) Update(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	teamID := uint(teamBizKey)
	pmID := middleware.GetUserID(c)

	var req dto.UpdateTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	if middleware.IsSuperAdmin(c) {
		team, err := h.teamSvc.GetTeam(c.Request.Context(), teamID)
		if err != nil {
			apperrors.RespondError(c, err)
			return
		}
		pmID = uint(team.PmKey)
	}

	team, err := h.teamSvc.UpdateTeam(c.Request.Context(), pmID, teamID, req)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, vo.NewTeamVO(team))
}

// Disband handles DELETE /api/v1/teams/:teamId
func (h *TeamHandler) Disband(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	teamID := uint(teamBizKey)
	callerID := middleware.GetUserID(c)

	var req dto.DisbandTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	if middleware.IsSuperAdmin(c) {
		team, err := h.teamSvc.GetTeam(c.Request.Context(), teamID)
		if err != nil {
			apperrors.RespondError(c, err)
			return
		}
		callerID = uint(team.PmKey)
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
	teamBizKey := middleware.GetTeamBizKey(c)
	teamID := uint(teamBizKey)

	members, err := h.teamSvc.ListMembers(c.Request.Context(), teamID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, members)
}

// InviteMember handles POST /api/v1/teams/:teamId/members
func (h *TeamHandler) InviteMember(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	teamID := uint(teamBizKey)
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
	teamBizKey := middleware.GetTeamBizKey(c)
	teamID := uint(teamBizKey)
	pmID := middleware.GetUserID(c)

	targetUserID, ok := pkgHandler.ResolveBizKey(c, "userId", func(ctx context.Context, bizKey int64) (uint, error) {
		user, err := h.userRepo.FindByBizKey(ctx, bizKey)
		if err != nil {
			return 0, apperrors.ErrItemNotFound
		}
		return user.ID, nil
	})
	if !ok {
		return
	}

	if middleware.IsSuperAdmin(c) {
		team, err := h.teamSvc.GetTeam(c.Request.Context(), teamID)
		if err != nil {
			apperrors.RespondError(c, err)
			return
		}
		pmID = uint(team.PmKey)
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
	teamBizKey := middleware.GetTeamBizKey(c)
	teamID := uint(teamBizKey)
	pmID := middleware.GetUserID(c)

	targetUserID, ok := pkgHandler.ResolveBizKey(c, "userId", func(ctx context.Context, bizKey int64) (uint, error) {
		user, err := h.userRepo.FindByBizKey(ctx, bizKey)
		if err != nil {
			return 0, apperrors.ErrItemNotFound
		}
		return user.ID, nil
	})
	if !ok {
		return
	}

	var req dto.UpdateMemberRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	if middleware.IsSuperAdmin(c) {
		team, err := h.teamSvc.GetTeam(c.Request.Context(), teamID)
		if err != nil {
			apperrors.RespondError(c, err)
			return
		}
		pmID = uint(team.PmKey)
	}

	roleKey, _ := pkg.ParseID(req.RoleKey)
	if err := h.teamSvc.UpdateMemberRole(c.Request.Context(), pmID, teamID, targetUserID, roleKey); err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// TransferPM handles PUT /api/v1/teams/:teamId/pm
func (h *TeamHandler) TransferPM(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	teamID := uint(teamBizKey)
	callerID := middleware.GetUserID(c)

	var req dto.TransferPMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	// Resolve newPmUserKey (bizKey) to internal user ID
	newPmBizKey, err := pkg.ParseID(req.NewPmUserKey)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}
	newPmUser, err := h.userRepo.FindByBizKey(c.Request.Context(), newPmBizKey)
	if err != nil {
		apperrors.RespondError(c, apperrors.ErrUserNotFound)
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

	err = h.teamSvc.TransferPM(c.Request.Context(), pmID, teamID, newPmUser.ID)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, nil)
}

// SearchUsers handles GET /api/v1/teams/:teamId/search-users
func (h *TeamHandler) SearchUsers(c *gin.Context) {
	teamBizKey := middleware.GetTeamBizKey(c)
	teamID := uint(teamBizKey)
	search := c.Query("search")

	users, err := h.teamSvc.SearchAvailableUsers(c.Request.Context(), teamID, search)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, users)
}

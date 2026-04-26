package service

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/repository"
)

// TransactionDB abstracts the gorm.DB.Transaction method for testability.
type TransactionDB interface {
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
}

type TeamService interface {
	CreateTeam(ctx context.Context, creatorID uint, req dto.CreateTeamReq) (*model.Team, error)
	GetTeam(ctx context.Context, teamID uint) (*model.Team, error)
	GetTeamDetail(ctx context.Context, teamID uint) (*dto.TeamDetailResp, error)
	ListTeams(ctx context.Context, callerID uint, isSuperAdmin bool, search string, page, pageSize int) ([]*dto.TeamListResp, int64, error)
	UpdateTeam(ctx context.Context, pmID, teamID uint, req dto.UpdateTeamReq) (*model.Team, error)
	InviteMember(ctx context.Context, pmID, teamID uint, req dto.InviteMemberReq) error
	RemoveMember(ctx context.Context, pmID, teamID, targetUserID uint) error
	TransferPM(ctx context.Context, currentPMID, teamID, newPMID uint) error
	DisbandTeam(ctx context.Context, callerID uint, teamID uint, confirmName string) error
	UpdateMemberRole(ctx context.Context, pmID, teamID, targetUserID, roleID uint) error
	ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error)
	SearchAvailableUsers(ctx context.Context, teamID uint, search string) ([]*dto.UserSearchDTO, error)
}

type teamService struct {
	teamRepo    repository.TeamRepo
	userRepo    repository.UserRepo
	mainItemRepo repository.MainItemRepo
	roleRepo    repository.RoleRepo
	db          TransactionDB
}

func NewTeamService(teamRepo repository.TeamRepo, userRepo repository.UserRepo, mainItemRepo repository.MainItemRepo, roleRepo repository.RoleRepo, db TransactionDB) TeamService {
	return &teamService{teamRepo: teamRepo, userRepo: userRepo, mainItemRepo: mainItemRepo, roleRepo: roleRepo, db: db}
}

func (s *teamService) CreateTeam(ctx context.Context, creatorID uint, req dto.CreateTeamReq) (*model.Team, error) {
	team := &model.Team{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamName:  req.Name,
		TeamDesc:  req.Description,
		Code:      req.Code,
		PmKey:     int64(creatorID),
	}
	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	member := &model.TeamMember{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:   int64(team.ID),
		UserKey:   int64(creatorID),
		JoinedAt:  time.Now(),
	}
	if s.roleRepo != nil {
		if pmRole, err := s.roleRepo.FindByName(ctx, "pm"); err == nil {
			roleKey := int64(pmRole.ID)
			member.RoleKey = &roleKey
		}
	}
	if err := s.teamRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return team, nil
}

func (s *teamService) GetTeam(ctx context.Context, teamID uint) (*model.Team, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	return team, nil
}

func (s *teamService) ListTeams(ctx context.Context, _ uint, _ bool, search string, page, pageSize int) ([]*dto.TeamListResp, int64, error) {
	offset, page, pageSize := dto.ApplyPaginationDefaults(page, pageSize)
	teams, total, err := s.teamRepo.ListFiltered(ctx, search, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	teamIDs := make([]uint, len(teams))
	for i, t := range teams {
		teamIDs[i] = t.ID
	}
	pmNames, _ := s.teamRepo.FindPMMembers(ctx, teamIDs)

	result := make([]*dto.TeamListResp, len(teams))
	for i, t := range teams {
		result[i] = &dto.TeamListResp{
			BizKey:        pkg.FormatID(t.BizKey),
			Name:          t.TeamName,
			Description:   t.TeamDesc,
			Code:          t.Code,
			PmKey:         pkg.FormatID(t.PmKey),
			PmDisplayName: pmNames[t.ID],
			CreatedAt:     t.CreateTime.Format(time.RFC3339),
			UpdatedAt:     t.DbUpdateTime.Format(time.RFC3339),
		}
	}
	return result, total, nil
}

func (s *teamService) GetTeamDetail(ctx context.Context, teamID uint) (*dto.TeamDetailResp, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}

	pm, err := s.userRepo.FindByID(ctx, uint(team.PmKey))
	if err != nil {
		return nil, err
	}

	// Use COUNT(*) for member count instead of loading all members
	memberCount, err := s.teamRepo.CountMembers(ctx, teamID)
	if err != nil {
		// Fallback: load all members and count
		members, listErr := s.teamRepo.ListMembers(ctx, teamID)
		if listErr != nil {
			return nil, listErr
		}
		memberCount = int64(len(members))
	}

	var mainItemCount int64
	if s.mainItemRepo != nil {
		mainItemCount, _ = s.mainItemRepo.CountByTeam(ctx, teamID)
	}

	return &dto.TeamDetailResp{
		BizKey:        pkg.FormatID(team.BizKey),
		Name:          team.TeamName,
		Description:   team.TeamDesc,
		Code:          team.Code,
		PmKey:         pkg.FormatID(team.PmKey),
		PmDisplayName: pm.DisplayName,
		MemberCount:   int(memberCount),
		MainItemCount: int(mainItemCount),
		CreatedAt:     team.CreateTime.Format(time.RFC3339),
		UpdatedAt:     team.DbUpdateTime.Format(time.RFC3339),
	}, nil
}

func (s *teamService) UpdateTeam(ctx context.Context, pmID, teamID uint, req dto.UpdateTeamReq) (*model.Team, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmKey != int64(pmID) {
		return nil, apperrors.ErrForbidden
	}

	team.TeamName = req.Name
	team.TeamDesc = req.Description
	if err := s.teamRepo.Update(ctx, team); err != nil {
		return nil, err
	}
	return team, nil
}

func (s *teamService) InviteMember(ctx context.Context, pmID, teamID uint, req dto.InviteMemberReq) error {
	if _, err := s.teamRepo.FindByID(ctx, teamID); err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if roleID, err := pkg.ParseID(req.RoleKey); err == nil && s.isPMRole(ctx, uint(roleID)) {
		return apperrors.ErrCannotAssignPMRole
	}

	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}

	existing, err := s.teamRepo.FindMember(ctx, teamID, user.ID)
	if err != nil && err != apperrors.ErrNotFound {
		return err
	}
	if existing != nil {
		return apperrors.ErrAlreadyMember
	}

	member := &model.TeamMember{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:   int64(teamID),
		UserKey:   int64(user.ID),
		RoleKey:   func() *int64 { v, _ := pkg.ParseID(req.RoleKey); return &v }(),
		JoinedAt:  time.Now(),
	}
	return s.teamRepo.AddMember(ctx, member)
}

func (s *teamService) RemoveMember(ctx context.Context, pmID, teamID, targetUserID uint) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmKey != int64(pmID) {
		return apperrors.ErrForbidden
	}
	if targetUserID == pmID {
		return apperrors.ErrCannotRemoveSelf
	}

	return s.teamRepo.RemoveMember(ctx, teamID, targetUserID)
}

func (s *teamService) TransferPM(ctx context.Context, currentPMID, teamID, newPMID uint) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmKey != int64(currentPMID) {
		return apperrors.ErrForbidden
	}

	// Verify new PM is a team member
	newPMMember, err := s.teamRepo.FindMember(ctx, teamID, newPMID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotTeamMember)
	}

	// Atomic transfer via transaction
	return s.db.Transaction(func(_ *gorm.DB) error {
		// Update team PM
		team.PmKey = int64(newPMID)
		if err := s.teamRepo.Update(ctx, team); err != nil {
			return err
		}

		// New PM gets "pm" role
		if s.roleRepo != nil {
			if pmRole, err := s.roleRepo.FindByName(ctx, "pm"); err == nil {
				roleKey := int64(pmRole.ID)
				newPMMember.RoleKey = &roleKey
			}
		}
		if err := s.teamRepo.UpdateMember(ctx, newPMMember); err != nil {
			return err
		}

		// Old PM gets "member" role
		oldPMMember, err := s.teamRepo.FindMember(ctx, teamID, currentPMID)
		if err != nil {
			return err
		}
		if s.roleRepo != nil {
			if memberRole, err := s.roleRepo.FindByName(ctx, "member"); err == nil {
				roleKey := int64(memberRole.ID)
				oldPMMember.RoleKey = &roleKey
			}
		}
		return s.teamRepo.UpdateMember(ctx, oldPMMember)
	})
}

func (s *teamService) DisbandTeam(ctx context.Context, callerID uint, teamID uint, confirmName string) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmKey != int64(callerID) {
		return apperrors.ErrForbidden
	}
	if team.TeamName != confirmName {
		return apperrors.ErrValidation
	}

	return s.teamRepo.SoftDelete(ctx, teamID)
}

func (s *teamService) UpdateMemberRole(ctx context.Context, pmID, teamID, targetUserID, roleID uint) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmKey != int64(pmID) {
		return apperrors.ErrForbidden
	}

	if s.isPMRole(ctx, roleID) {
		return apperrors.ErrCannotAssignPMRole
	}

	member, err := s.teamRepo.FindMember(ctx, teamID, targetUserID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotTeamMember)
	}

	roleKey := int64(roleID)
	member.RoleKey = &roleKey
	return s.teamRepo.UpdateMember(ctx, member)
}

// isPMRole returns true if the given roleID corresponds to the "pm" preset role.
func (s *teamService) isPMRole(ctx context.Context, roleID uint) bool {
	if s.roleRepo == nil {
		return false
	}
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return false
	}
	return role.Name == "pm"
}

func (s *teamService) ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error) {
	return s.teamRepo.ListMembers(ctx, teamID)
}

func (s *teamService) SearchAvailableUsers(ctx context.Context, teamID uint, search string) ([]*dto.UserSearchDTO, error) {
	users, err := s.userRepo.SearchAvailable(ctx, teamID, search, 20)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.UserSearchDTO, len(users))
	for i, u := range users {
		result[i] = &dto.UserSearchDTO{
			BizKey:      pkg.FormatID(u.BizKey),
			Username:    u.Username,
			DisplayName: u.DisplayName,
		}
	}

	return result, nil
}

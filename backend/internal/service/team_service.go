package service

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/repository"
)

type TeamService interface {
	CreateTeam(ctx context.Context, creatorBizKey int64, req dto.CreateTeamReq) (*model.Team, error)
	GetTeam(ctx context.Context, teamBizKey int64) (*model.Team, error)
	GetTeamDetail(ctx context.Context, teamBizKey int64) (*dto.TeamDetailResp, error)
	ListTeams(ctx context.Context, callerID uint, isSuperAdmin bool, search string, page, pageSize int) ([]*dto.TeamListResp, int64, error)
	UpdateTeam(ctx context.Context, pmBizKey int64, teamBizKey int64, req dto.UpdateTeamReq) (*model.Team, error)
	InviteMember(ctx context.Context, pmBizKey int64, teamBizKey int64, req dto.InviteMemberReq) error
	RemoveMember(ctx context.Context, pmBizKey int64, teamBizKey int64, targetUserBizKey int64) error
	TransferPM(ctx context.Context, currentPMBizKey int64, teamBizKey int64, newPMBizKey int64) error
	DisbandTeam(ctx context.Context, callerBizKey int64, teamBizKey int64, confirmName string) error
	UpdateMemberRole(ctx context.Context, pmBizKey int64, targetUserBizKey int64, teamBizKey int64, roleBizKey int64) error
	ListMembers(ctx context.Context, teamBizKey int64) ([]*dto.TeamMemberDTO, error)
	SearchAvailableUsers(ctx context.Context, teamBizKey int64, search string) ([]*dto.UserSearchDTO, error)
}

type teamService struct {
	teamRepo    repository.TeamRepo
	userRepo    repository.UserRepo
	mainItemRepo repository.MainItemRepo
	roleRepo    repository.RoleRepo
	db          repo.DBTransactor
}

func NewTeamService(teamRepo repository.TeamRepo, userRepo repository.UserRepo, mainItemRepo repository.MainItemRepo, roleRepo repository.RoleRepo, db repo.DBTransactor) TeamService {
	return &teamService{teamRepo: teamRepo, userRepo: userRepo, mainItemRepo: mainItemRepo, roleRepo: roleRepo, db: db}
}

func (s *teamService) CreateTeam(ctx context.Context, creatorBizKey int64, req dto.CreateTeamReq) (*model.Team, error) {
	team := &model.Team{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamName:  req.Name,
		TeamDesc:  req.Description,
		Code:      req.Code,
		PmKey:     creatorBizKey,
	}
	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	member := &model.TeamMember{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:   team.BizKey,
		UserKey:   creatorBizKey,
		JoinedAt:  time.Now(),
	}
	if s.roleRepo != nil {
		if pmRole, err := s.roleRepo.FindByName(ctx, "pm"); err == nil {
			roleKey := pmRole.BizKey
			member.RoleKey = &roleKey
		}
	}
	if err := s.teamRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return team, nil
}

func (s *teamService) GetTeam(ctx context.Context, teamBizKey int64) (*model.Team, error) {
	team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
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

	teamBizKeys := make([]int64, len(teams))
	for i, t := range teams {
		teamBizKeys[i] = t.BizKey
	}
	pmNames, _ := s.teamRepo.FindPMMembers(ctx, teamBizKeys)

	result := make([]*dto.TeamListResp, len(teams))
	for i, t := range teams {
		result[i] = &dto.TeamListResp{
			BizKey:        pkg.FormatID(t.BizKey),
			Name:          t.TeamName,
			Description:   t.TeamDesc,
			Code:          t.Code,
			PmKey:         pkg.FormatID(t.PmKey),
			PmDisplayName: pmNames[t.BizKey],
			CreatedAt:     t.CreateTime.Format(time.RFC3339),
			UpdatedAt:     t.DbUpdateTime.Format(time.RFC3339),
		}
	}
	return result, total, nil
}

func (s *teamService) GetTeamDetail(ctx context.Context, teamBizKey int64) (*dto.TeamDetailResp, error) {
	team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}

	pm, err := s.userRepo.FindByBizKey(ctx, team.PmKey)
	if err != nil {
		return nil, err
	}

	// Use COUNT(*) for member count instead of loading all members
	memberCount, err := s.teamRepo.CountMembers(ctx, team.BizKey)
	if err != nil {
		// Fallback: load all members and count
		members, listErr := s.teamRepo.ListMembers(ctx, team.BizKey)
		if listErr != nil {
			return nil, listErr
		}
		memberCount = int64(len(members))
	}

	var mainItemCount int64
	if s.mainItemRepo != nil {
		mainItemCount, _ = s.mainItemRepo.CountByTeam(ctx, team.BizKey)
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

func (s *teamService) UpdateTeam(ctx context.Context, pmBizKey int64, teamBizKey int64, req dto.UpdateTeamReq) (*model.Team, error) {
	team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmKey != pmBizKey {
		return nil, apperrors.ErrForbidden
	}

	team.TeamName = req.Name
	team.TeamDesc = req.Description
	if err := s.teamRepo.Update(ctx, team); err != nil {
		return nil, err
	}
	return team, nil
}

func (s *teamService) InviteMember(ctx context.Context, pmBizKey int64, teamBizKey int64, req dto.InviteMemberReq) error {
	team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if roleID, err := pkg.ParseID(req.RoleKey); err == nil && s.isPMRole(ctx, roleID) {
		return apperrors.ErrCannotAssignPMRole
	}

	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotFound)
	}

	existing, err := s.teamRepo.FindMember(ctx, team.BizKey, user.BizKey)
	if err != nil && err != apperrors.ErrNotFound {
		return err
	}
	if existing != nil {
		return apperrors.ErrAlreadyMember
	}

	member := &model.TeamMember{
		BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
		TeamKey:   team.BizKey,
		UserKey:   user.BizKey,
		RoleKey:   func() *int64 { v, _ := pkg.ParseID(req.RoleKey); return &v }(),
		JoinedAt:  time.Now(),
	}
	return s.teamRepo.AddMember(ctx, member)
}

func (s *teamService) RemoveMember(ctx context.Context, pmBizKey int64, teamBizKey int64, targetUserBizKey int64) error {
	team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmKey != pmBizKey {
		return apperrors.ErrForbidden
	}
	if targetUserBizKey == pmBizKey {
		return apperrors.ErrCannotRemoveSelf
	}

	return s.teamRepo.RemoveMember(ctx, team.BizKey, targetUserBizKey)
}

func (s *teamService) TransferPM(ctx context.Context, currentPMBizKey int64, teamBizKey int64, newPMBizKey int64) error {
	team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmKey != currentPMBizKey {
		return apperrors.ErrForbidden
	}

	// Verify new PM is a team member
	newPMMember, err := s.teamRepo.FindMember(ctx, team.BizKey, newPMBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotTeamMember)
	}

	// Atomic transfer via transaction
	return s.db.Transaction(func(_ *gorm.DB) error {
		// Update team PM
		team.PmKey = newPMBizKey
		if err := s.teamRepo.Update(ctx, team); err != nil {
			return err
		}

		// New PM gets "pm" role
		if s.roleRepo != nil {
			if pmRole, err := s.roleRepo.FindByName(ctx, "pm"); err == nil {
				roleKey := pmRole.BizKey
				newPMMember.RoleKey = &roleKey
			}
		}
		if err := s.teamRepo.UpdateMember(ctx, newPMMember); err != nil {
			return err
		}

		// Old PM gets "member" role
		oldPMMember, err := s.teamRepo.FindMember(ctx, team.BizKey, currentPMBizKey)
		if err != nil {
			return err
		}
		if s.roleRepo != nil {
			if memberRole, err := s.roleRepo.FindByName(ctx, "member"); err == nil {
				roleKey := memberRole.BizKey
				oldPMMember.RoleKey = &roleKey
			}
		}
		return s.teamRepo.UpdateMember(ctx, oldPMMember)
	})
}

func (s *teamService) DisbandTeam(ctx context.Context, callerBizKey int64, teamBizKey int64, confirmName string) error {
	team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmKey != callerBizKey {
		return apperrors.ErrForbidden
	}
	if team.TeamName != confirmName {
		return apperrors.ErrValidation
	}

	return s.teamRepo.SoftDelete(ctx, team.ID)
}

func (s *teamService) UpdateMemberRole(ctx context.Context, pmBizKey int64, targetUserBizKey int64, teamBizKey int64, roleBizKey int64) error {
	team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmKey != pmBizKey {
		return apperrors.ErrForbidden
	}

	if s.isPMRole(ctx, roleBizKey) {
		return apperrors.ErrCannotAssignPMRole
	}

	member, err := s.teamRepo.FindMember(ctx, team.BizKey, targetUserBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotTeamMember)
	}

	member.RoleKey = &roleBizKey
	return s.teamRepo.UpdateMember(ctx, member)
}

// isPMRole returns true if the given bizKey corresponds to the "pm" preset role.
func (s *teamService) isPMRole(ctx context.Context, bizKey int64) bool {
	if s.roleRepo == nil {
		return false
	}
	role, err := s.roleRepo.FindByBizKey(ctx, bizKey)
	if err != nil {
		return false
	}
	return role.Name == "pm"
}

func (s *teamService) ListMembers(ctx context.Context, teamBizKey int64) ([]*dto.TeamMemberDTO, error) {
	team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	return s.teamRepo.ListMembers(ctx, team.BizKey)
}

func (s *teamService) SearchAvailableUsers(ctx context.Context, teamBizKey int64, search string) ([]*dto.UserSearchDTO, error) {
	team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	users, err := s.userRepo.SearchAvailable(ctx, team.BizKey, search, 20)
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

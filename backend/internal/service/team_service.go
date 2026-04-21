package service

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
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
	ListTeams(ctx context.Context, callerID uint, isSuperAdmin bool) ([]*dto.TeamListResp, error)
	UpdateTeam(ctx context.Context, pmID, teamID uint, req dto.UpdateTeamReq) (*model.Team, error)
	InviteMember(ctx context.Context, pmID, teamID uint, req dto.InviteMemberReq) error
	RemoveMember(ctx context.Context, pmID, teamID, targetUserID uint) error
	TransferPM(ctx context.Context, currentPMID, teamID, newPMID uint) error
	DisbandTeam(ctx context.Context, callerID uint, teamID uint, confirmName string) error
	UpdateMemberRole(ctx context.Context, pmID, teamID, targetUserID uint, role string) error
	ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error)
	SearchAvailableUsers(ctx context.Context, teamID uint, search string) ([]*dto.UserSearchDTO, error)
}

type teamService struct {
	teamRepo    repository.TeamRepo
	userRepo    repository.UserRepo
	mainItemRepo repository.MainItemRepo
	db          TransactionDB
}

func NewTeamService(teamRepo repository.TeamRepo, userRepo repository.UserRepo, mainItemRepo repository.MainItemRepo, db TransactionDB) TeamService {
	return &teamService{teamRepo: teamRepo, userRepo: userRepo, mainItemRepo: mainItemRepo, db: db}
}

func (s *teamService) CreateTeam(ctx context.Context, creatorID uint, req dto.CreateTeamReq) (*model.Team, error) {
	team := &model.Team{
		Name:        req.Name,
		Description: req.Description,
		PmID:        creatorID,
	}
	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	member := &model.TeamMember{
		TeamID:   team.ID,
		UserID:   creatorID,
		Role:     "pm",
		JoinedAt: time.Now(),
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

func (s *teamService) ListTeams(ctx context.Context, _ uint, _ bool) ([]*dto.TeamListResp, error) {
	teams, err := s.teamRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	teamIDs := make([]uint, len(teams))
	for i, t := range teams {
		teamIDs[i] = t.ID
	}
	pmNames, _ := s.teamRepo.FindPMMembers(ctx, teamIDs)

	result := make([]*dto.TeamListResp, len(teams))
	for i, t := range teams {
		result[i] = &dto.TeamListResp{
			ID:            t.ID,
			Name:          t.Name,
			Description:   t.Description,
			PmID:          t.PmID,
			PmDisplayName: pmNames[t.ID],
			CreatedAt:     t.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     t.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

func (s *teamService) GetTeamDetail(ctx context.Context, teamID uint) (*dto.TeamDetailResp, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}

	pm, err := s.userRepo.FindByID(ctx, team.PmID)
	if err != nil {
		return nil, err
	}

	members, err := s.teamRepo.ListMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}

	var mainItemCount int64
	if s.mainItemRepo != nil {
		mainItemCount, _ = s.mainItemRepo.CountByTeam(ctx, teamID)
	}

	return &dto.TeamDetailResp{
		ID:            team.ID,
		Name:          team.Name,
		Description:   team.Description,
		PmID:          team.PmID,
		PmDisplayName: pm.DisplayName,
		MemberCount:   len(members),
		MainItemCount: int(mainItemCount),
		CreatedAt:     team.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     team.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *teamService) UpdateTeam(ctx context.Context, pmID, teamID uint, req dto.UpdateTeamReq) (*model.Team, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmID != pmID {
		return nil, apperrors.ErrForbidden
	}

	team.Name = req.Name
	team.Description = req.Description
	if err := s.teamRepo.Update(ctx, team); err != nil {
		return nil, err
	}
	return team, nil
}

func (s *teamService) InviteMember(ctx context.Context, pmID, teamID uint, req dto.InviteMemberReq) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	_ = team.PmID // permission is enforced by RequirePermission middleware

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
		TeamID:   teamID,
		UserID:   user.ID,
		RoleID:   &req.RoleID,
		JoinedAt: time.Now(),
	}
	return s.teamRepo.AddMember(ctx, member)
}

func (s *teamService) RemoveMember(ctx context.Context, pmID, teamID, targetUserID uint) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmID != pmID {
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
	if team.PmID != currentPMID {
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
		team.PmID = newPMID
		if err := s.teamRepo.Update(ctx, team); err != nil {
			return err
		}

		// New PM gets "pm" role
		newPMMember.Role = "pm"
		if err := s.teamRepo.UpdateMember(ctx, newPMMember); err != nil {
			return err
		}

		// Old PM gets "member" role
		oldPMMember, err := s.teamRepo.FindMember(ctx, teamID, currentPMID)
		if err != nil {
			return err
		}
		oldPMMember.Role = "member"
		return s.teamRepo.UpdateMember(ctx, oldPMMember)
	})
}

func (s *teamService) DisbandTeam(ctx context.Context, callerID uint, teamID uint, confirmName string) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmID != callerID {
		return apperrors.ErrForbidden
	}
	if team.Name != confirmName {
		return apperrors.ErrValidation
	}

	return s.teamRepo.Delete(ctx, teamID)
}

func (s *teamService) UpdateMemberRole(ctx context.Context, pmID, teamID, targetUserID uint, role string) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrTeamNotFound)
	}
	if team.PmID != pmID {
		return apperrors.ErrForbidden
	}

	member, err := s.teamRepo.FindMember(ctx, teamID, targetUserID)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrNotTeamMember)
	}

	member.Role = role
	return s.teamRepo.UpdateMember(ctx, member)
}

func (s *teamService) ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error) {
	return s.teamRepo.ListMembers(ctx, teamID)
}

func (s *teamService) SearchAvailableUsers(ctx context.Context, teamID uint, search string) ([]*dto.UserSearchDTO, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Get existing member user IDs
	members, err := s.teamRepo.ListMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}
	memberIDs := make(map[uint]bool, len(members))
	for _, m := range members {
		memberIDs[m.UserID] = true
	}

	// Filter by search and exclude existing members
	var result []*dto.UserSearchDTO
	for _, u := range users {
		if memberIDs[u.ID] {
			continue
		}
		if search != "" {
			if !containsIgnoreCase(u.Username, search) && !containsIgnoreCase(u.DisplayName, search) {
				continue
			}
		}
		result = append(result, &dto.UserSearchDTO{
			ID:          u.ID,
			Username:    u.Username,
			DisplayName: u.DisplayName,
		})
		// Limit results to 20
		if len(result) >= 20 {
			break
		}
	}

	if result == nil {
		result = []*dto.UserSearchDTO{}
	}
	return result, nil
}

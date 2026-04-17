package service

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

// TransactionDB abstracts the gorm.DB.Transaction method for testability.
type TransactionDB interface {
	Transaction(fc func(tx interface{}) error) error
}

type TeamService interface {
	CreateTeam(ctx context.Context, creatorID uint, req dto.CreateTeamReq) (*model.Team, error)
	GetTeam(ctx context.Context, teamID uint) (*model.Team, error)
	GetTeamDetail(ctx context.Context, teamID uint) (*dto.TeamDetailResp, error)
	ListTeams(ctx context.Context, callerID uint, isSuperAdmin bool) ([]*model.Team, error)
	UpdateTeam(ctx context.Context, pmID, teamID uint, req dto.UpdateTeamReq) (*model.Team, error)
	InviteMember(ctx context.Context, pmID, teamID uint, req dto.InviteMemberReq) error
	RemoveMember(ctx context.Context, pmID, teamID, targetUserID uint) error
	TransferPM(ctx context.Context, currentPMID, teamID, newPMID uint) error
	DisbandTeam(ctx context.Context, callerID uint, teamID uint, confirmName string) error
	UpdateMemberRole(ctx context.Context, pmID, teamID, targetUserID uint, role string) error
	ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error)
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
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, err
	}
	return team, nil
}

func (s *teamService) ListTeams(ctx context.Context, _ uint, _ bool) ([]*model.Team, error) {
	return s.teamRepo.List(ctx)
}

func (s *teamService) GetTeamDetail(ctx context.Context, teamID uint) (*dto.TeamDetailResp, error) {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, err
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
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, err
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
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrTeamNotFound
		}
		return err
	}
	if team.PmID != pmID {
		return apperrors.ErrForbidden
	}

	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrNotFound
		}
		return err
	}

	existing, err := s.teamRepo.FindMember(ctx, teamID, user.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if existing != nil {
		return apperrors.ErrAlreadyMember
	}

	member := &model.TeamMember{
		TeamID:   teamID,
		UserID:   user.ID,
		Role:     req.Role,
		JoinedAt: time.Now(),
	}
	return s.teamRepo.AddMember(ctx, member)
}

func (s *teamService) RemoveMember(ctx context.Context, pmID, teamID, targetUserID uint) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrTeamNotFound
		}
		return err
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
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrTeamNotFound
		}
		return err
	}
	if team.PmID != currentPMID {
		return apperrors.ErrForbidden
	}

	// Verify new PM is a team member
	newPMMember, err := s.teamRepo.FindMember(ctx, teamID, newPMID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrNotTeamMember
		}
		return err
	}

	// Atomic transfer via transaction
	return s.db.Transaction(func(_ interface{}) error {
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
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrTeamNotFound
		}
		return err
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
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrTeamNotFound
		}
		return err
	}
	if team.PmID != pmID {
		return apperrors.ErrForbidden
	}

	member, err := s.teamRepo.FindMember(ctx, teamID, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrNotTeamMember
		}
		return err
	}

	member.Role = role
	return s.teamRepo.UpdateMember(ctx, member)
}

func (s *teamService) ListMembers(ctx context.Context, teamID uint) ([]*dto.TeamMemberDTO, error) {
	return s.teamRepo.ListMembers(ctx, teamID)
}

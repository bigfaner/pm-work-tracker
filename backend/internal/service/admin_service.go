package service

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

// AdminService defines admin-only operations.
type AdminService interface {
	ListUsers(ctx context.Context) ([]*model.User, error)
	SetCanCreateTeam(ctx context.Context, superAdminID, targetUserID uint, canCreate bool) error
	ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error)
}

type adminService struct {
	userRepo repository.UserRepo
	teamRepo repository.TeamRepo
}

// NewAdminService creates a new AdminService.
func NewAdminService(userRepo repository.UserRepo, teamRepo repository.TeamRepo) AdminService {
	return &adminService{userRepo: userRepo, teamRepo: teamRepo}
}

func (s *adminService) ListUsers(ctx context.Context) ([]*model.User, error) {
	return s.userRepo.List(ctx)
}

func (s *adminService) SetCanCreateTeam(ctx context.Context, superAdminID, targetUserID uint, canCreate bool) error {
	if superAdminID == targetUserID {
		return apperrors.ErrCannotModifySelf
	}

	user, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	user.CanCreateTeam = canCreate
	return s.userRepo.Update(ctx, user)
}

func (s *adminService) ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error) {
	return s.teamRepo.ListAllTeams(ctx)
}

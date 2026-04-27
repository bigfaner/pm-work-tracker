package service

import (
	"context"
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/vo"
)

// AdminService defines admin-only operations.
type AdminService interface {
	ListUsers(ctx context.Context, search string, page, pageSize int) ([]*dto.AdminUserDTO, int, error)
	GetUser(ctx context.Context, userBizKey int64) (*dto.AdminUserDTO, error)
	CreateUser(ctx context.Context, req *dto.CreateUserReq) (*dto.AdminUserDTO, error)
	UpdateUser(ctx context.Context, userBizKey int64, req *dto.UpdateUserReq) (*dto.AdminUserDTO, error)
	ToggleUserStatus(ctx context.Context, callerID uint, targetBizKey int64, status string) (*dto.AdminUserDTO, error)
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

const passwordCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func generatePassword(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordCharset))))
		if err != nil {
			return "", err
		}
		b[i] = passwordCharset[n.Int64()]
	}
	return string(b), nil
}

func (s *adminService) ListUsers(ctx context.Context, search string, page, pageSize int) ([]*dto.AdminUserDTO, int, error) {
	offset := (page - 1) * pageSize

	users, total, err := s.userRepo.ListFiltered(ctx, search, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// Collect user IDs for team lookup
	userIDs := make([]uint, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*dto.AdminUserDTO, len(users))
	for i, u := range users {
		items[i] = modelToAdminUserDTO(u, teamsMap[u.ID])
	}

	return items, int(total), nil
}

func (s *adminService) GetUser(ctx context.Context, userBizKey int64) (*dto.AdminUserDTO, error) {
	user, err := s.userRepo.FindByBizKey(ctx, userBizKey)
	if err != nil {
		return nil, err
	}

	teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, []uint{user.ID})
	if err != nil {
		return nil, err
	}

	return modelToAdminUserDTO(user, teamsMap[user.ID]), nil
}

func (s *adminService) CreateUser(ctx context.Context, req *dto.CreateUserReq) (*dto.AdminUserDTO, error) {
	// Check for duplicate username
	existing, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil && existing != nil {
		return nil, apperrors.ErrUserExists
	}
	// If error is not ErrNotFound, it's a real error
	if err != nil && err != apperrors.ErrNotFound {
		return nil, err
	}

	// Generate random password
	password, err := generatePassword(12)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	user := &model.User{
		BaseModel:    model.BaseModel{BizKey: snowflake.Generate()},
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		Email:        req.Email,
		PasswordHash: string(hash),
		UserStatus:   "enabled",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// If teamKey provided, add to team
	var teams []dto.TeamSummary
	if req.TeamKey != nil && *req.TeamKey != "" {
		teamBizKey, err := pkg.ParseID(*req.TeamKey)
		if err != nil {
			return nil, apperrors.ErrTeamNotFound
		}
		team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
		if err != nil {
			return nil, apperrors.ErrTeamNotFound
		}
		member := &model.TeamMember{
			BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
			TeamKey:   int64(team.ID),
			UserKey:   int64(user.ID),
		}
		if err := s.teamRepo.AddMember(ctx, member); err != nil {
			return nil, err
		}
		teams = []dto.TeamSummary{{BizKey: pkg.FormatID(int64(team.ID)), TeamID: team.ID, Name: "", Role: "member"}}
		// Fetch team name
		teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, []uint{user.ID})
		if err == nil && len(teamsMap[user.ID]) > 0 {
			teams = teamsMap[user.ID]
		}
	}

	return &dto.AdminUserDTO{
		BizKey:          pkg.FormatID(user.BizKey),
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		Email:           user.Email,
		Status:          user.UserStatus,
		Teams:           teams,
		InitialPassword: password,
	}, nil
}

func (s *adminService) UpdateUser(ctx context.Context, userBizKey int64, req *dto.UpdateUserReq) (*dto.AdminUserDTO, error) {
	user, err := s.userRepo.FindByBizKey(ctx, userBizKey)
	if err != nil {
		return nil, err
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Email != nil {
		user.Email = *req.Email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Handle team assignment
	if req.TeamKey != nil {
		// Remove from all current teams, add to new one
		teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, []uint{user.ID})
		if err != nil {
			return nil, err
		}
		for _, t := range teamsMap[user.ID] {
			_ = s.teamRepo.RemoveMember(ctx, t.TeamID, user.ID)
		}
		if *req.TeamKey != "" {
			teamBizKey, err := pkg.ParseID(*req.TeamKey)
			if err != nil {
				return nil, apperrors.ErrTeamNotFound
			}
			team, err := s.teamRepo.FindByBizKey(ctx, teamBizKey)
			if err != nil {
				return nil, apperrors.ErrTeamNotFound
			}
			member := &model.TeamMember{
				BaseModel: model.BaseModel{BizKey: snowflake.Generate()},
				TeamKey:   int64(team.ID),
				UserKey:   int64(user.ID),
			}
			if err := s.teamRepo.AddMember(ctx, member); err != nil {
				return nil, err
			}
		}
	}

	// Reload teams
	teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, []uint{user.ID})
	if err != nil {
		return nil, err
	}

	return modelToAdminUserDTO(user, teamsMap[user.ID]), nil
}

func (s *adminService) ToggleUserStatus(ctx context.Context, callerID uint, targetBizKey int64, status string) (*dto.AdminUserDTO, error) {
	user, err := s.userRepo.FindByBizKey(ctx, targetBizKey)
	if err != nil {
		return nil, err
	}

	if status == "disabled" && callerID == user.ID {
		return nil, apperrors.ErrCannotDisableSelf
	}

	user.UserStatus = status
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, []uint{user.ID})
	if err != nil {
		return nil, err
	}

	return modelToAdminUserDTO(user, teamsMap[user.ID]), nil
}

func (s *adminService) ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error) {
	return s.teamRepo.ListAllTeams(ctx)
}

func modelToAdminUserDTO(u *model.User, teams []dto.TeamSummary) *dto.AdminUserDTO {
	if teams == nil {
		teams = []dto.TeamSummary{}
	}
	base := vo.NewUserVO(u)
	return &dto.AdminUserDTO{
		BizKey:       base.BizKey,
		Username:     base.Username,
		DisplayName:  base.DisplayName,
		Email:        base.Email,
		Status:       base.Status,
		IsSuperAdmin: base.IsSuperAdmin,
		Teams:        teams,
	}
}



package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
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
	ResetPassword(ctx context.Context, targetBizKey int64, newPassword string) (*dto.ResetPasswordResp, error)
	SoftDeleteUser(ctx context.Context, callerID uint, targetBizKey int64) error
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

	// Collect user BizKeys for team lookup
	userBizKeys := make([]int64, len(users))
	for i, u := range users {
		userBizKeys[i] = u.BizKey
	}

	teamsMap, err := s.teamRepo.FindTeamsByUserBizKeys(ctx, userBizKeys)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*dto.AdminUserDTO, len(users))
	for i, u := range users {
		items[i] = modelToAdminUserDTO(u, teamsMap[u.BizKey])
	}

	return items, int(total), nil
}

func (s *adminService) GetUser(ctx context.Context, userBizKey int64) (*dto.AdminUserDTO, error) {
	user, err := s.userRepo.FindByBizKey(ctx, userBizKey)
	if err != nil {
		return nil, err
	}

	teamsMap, err := s.teamRepo.FindTeamsByUserBizKeys(ctx, []int64{user.BizKey})
	if err != nil {
		return nil, err
	}

	return modelToAdminUserDTO(user, teamsMap[user.BizKey]), nil
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
			TeamKey:   team.BizKey,
			UserKey:   user.BizKey,
			JoinedAt:  time.Now(),
		}
		if err := s.teamRepo.AddMember(ctx, member); err != nil {
			return nil, err
		}
		teams = []dto.TeamSummary{{BizKey: pkg.FormatID(team.BizKey), Name: "", Role: "member"}}
		// Fetch team name
		teamsMap, err := s.teamRepo.FindTeamsByUserBizKeys(ctx, []int64{user.BizKey})
		if err == nil && len(teamsMap[user.BizKey]) > 0 {
			teams = teamsMap[user.BizKey]
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
		teamsMap, err := s.teamRepo.FindTeamsByUserBizKeys(ctx, []int64{user.BizKey})
		if err != nil {
			return nil, err
		}
		for _, t := range teamsMap[user.BizKey] {
			teamBizKey, _ := pkg.ParseID(t.BizKey)
			_ = s.teamRepo.RemoveMember(ctx, teamBizKey, user.BizKey)
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
				TeamKey:   team.BizKey,
				UserKey:   user.BizKey,
				JoinedAt:  time.Now(),
			}
			if err := s.teamRepo.AddMember(ctx, member); err != nil {
				return nil, err
			}
		}
	}

	// Reload teams
	teamsMap, err := s.teamRepo.FindTeamsByUserBizKeys(ctx, []int64{user.BizKey})
	if err != nil {
		return nil, err
	}

	return modelToAdminUserDTO(user, teamsMap[user.BizKey]), nil
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

	teamsMap, err := s.teamRepo.FindTeamsByUserBizKeys(ctx, []int64{user.BizKey})
	if err != nil {
		return nil, err
	}

	return modelToAdminUserDTO(user, teamsMap[user.BizKey]), nil
}

func (s *adminService) ResetPassword(ctx context.Context, targetBizKey int64, newPassword string) (*dto.ResetPasswordResp, error) {
	user, err := s.userRepo.FindByBizKey(ctx, targetBizKey)
	if err != nil {
		return nil, apperrors.MapNotFound(err, apperrors.ErrUserNotFound)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	user.PasswordHash = string(hash)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &dto.ResetPasswordResp{
		BizKey:      pkg.FormatID(user.BizKey),
		Username:    user.Username,
		DisplayName: user.DisplayName,
	}, nil
}

func (s *adminService) SoftDeleteUser(ctx context.Context, callerID uint, targetBizKey int64) error {
	user, err := s.userRepo.FindByBizKey(ctx, targetBizKey)
	if err != nil {
		return apperrors.MapNotFound(err, apperrors.ErrUserNotFound)
	}

	if callerID == user.ID {
		return apperrors.ErrCannotDeleteSelf
	}

	return s.userRepo.SoftDelete(ctx, user)
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

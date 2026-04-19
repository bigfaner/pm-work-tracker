package service

import (
	"context"
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/repository"
)

// AdminService defines admin-only operations.
type AdminService interface {
	ListUsers(ctx context.Context, search string, page, pageSize int) ([]*dto.AdminUserDTO, int, error)
	GetUser(ctx context.Context, userID uint) (*dto.AdminUserDTO, error)
	CreateUser(ctx context.Context, req *dto.CreateUserReq) (*dto.AdminUserDTO, error)
	UpdateUser(ctx context.Context, userID uint, req *dto.UpdateUserReq) (*dto.AdminUserDTO, error)
	ToggleUserStatus(ctx context.Context, callerID, targetUserID uint, status string) (*dto.AdminUserDTO, error)
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
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Filter
	filtered := make([]*model.User, 0, len(users))
	for _, u := range users {
		if search != "" {
			if !containsIgnoreCase(u.Username, search) && !containsIgnoreCase(u.DisplayName, search) {
				continue
			}
		}
		filtered = append(filtered, u)
	}

	total := len(filtered)

	// Paginate
	start := (page - 1) * pageSize
	if start >= total {
		return []*dto.AdminUserDTO{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	pageUsers := filtered[start:end]

	// Collect user IDs for team lookup
	userIDs := make([]uint, len(pageUsers))
	for i, u := range pageUsers {
		userIDs[i] = u.ID
	}

	teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*dto.AdminUserDTO, len(pageUsers))
	for i, u := range pageUsers {
		items[i] = modelToAdminUserDTO(u, teamsMap[u.ID])
	}

	return items, total, nil
}

func (s *adminService) GetUser(ctx context.Context, userID uint) (*dto.AdminUserDTO, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, []uint{userID})
	if err != nil {
		return nil, err
	}

	return modelToAdminUserDTO(user, teamsMap[userID]), nil
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
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		Email:        req.Email,
		PasswordHash: string(hash),
		Status:       "enabled",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// If teamId provided, add to team
	var teams []dto.TeamSummary
	if req.TeamID != nil {
		_, err := s.teamRepo.FindByID(ctx, *req.TeamID)
		if err != nil {
			return nil, apperrors.ErrTeamNotFound
		}
		member := &model.TeamMember{
			TeamID: *req.TeamID,
			UserID: user.ID,
			Role:   "member",
		}
		if err := s.teamRepo.AddMember(ctx, member); err != nil {
			return nil, err
		}
		teams = []dto.TeamSummary{{ID: *req.TeamID, Name: "", Role: "member"}}
		// Fetch team name
		teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, []uint{user.ID})
		if err == nil && len(teamsMap[user.ID]) > 0 {
			teams = teamsMap[user.ID]
		}
	}

	return &dto.AdminUserDTO{
		ID:              user.ID,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		Email:           user.Email,
		Status:          user.Status,
		Teams:           teams,
		InitialPassword: password,
	}, nil
}

func (s *adminService) UpdateUser(ctx context.Context, userID uint, req *dto.UpdateUserReq) (*dto.AdminUserDTO, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
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
	if req.TeamID != nil {
		// Remove from all current teams, add to new one
		teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, []uint{userID})
		if err != nil {
			return nil, err
		}
		for _, t := range teamsMap[userID] {
			_ = s.teamRepo.RemoveMember(ctx, t.ID, userID)
		}
		if *req.TeamID > 0 {
			_, err := s.teamRepo.FindByID(ctx, *req.TeamID)
			if err != nil {
				return nil, apperrors.ErrTeamNotFound
			}
			member := &model.TeamMember{
				TeamID: *req.TeamID,
				UserID: userID,
				Role:   "member",
			}
			if err := s.teamRepo.AddMember(ctx, member); err != nil {
				return nil, err
			}
		}
	}

	// Reload teams
	teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, []uint{userID})
	if err != nil {
		return nil, err
	}

	return modelToAdminUserDTO(user, teamsMap[userID]), nil
}

func (s *adminService) ToggleUserStatus(ctx context.Context, callerID, targetUserID uint, status string) (*dto.AdminUserDTO, error) {
	if status == "disabled" && callerID == targetUserID {
		return nil, apperrors.ErrCannotDisableSelf
	}

	user, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	user.Status = status
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	teamsMap, err := s.teamRepo.FindTeamsByUserIDs(ctx, []uint{targetUserID})
	if err != nil {
		return nil, err
	}

	return modelToAdminUserDTO(user, teamsMap[targetUserID]), nil
}

func (s *adminService) ListAllTeams(ctx context.Context) ([]*dto.AdminTeamDTO, error) {
	return s.teamRepo.ListAllTeams(ctx)
}

func modelToAdminUserDTO(u *model.User, teams []dto.TeamSummary) *dto.AdminUserDTO {
	if teams == nil {
		teams = []dto.TeamSummary{}
	}
	return &dto.AdminUserDTO{
		ID:           u.ID,
		Username:     u.Username,
		DisplayName:  u.DisplayName,
		Email:        u.Email,
		Status:       u.Status,
		IsSuperAdmin: u.IsSuperAdmin,
		Teams:        teams,
	}
}

func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return strContains(sLower, substrLower)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func strContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

package service

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	appjwt "pm-work-tracker/backend/internal/pkg/jwt"
	"pm-work-tracker/backend/internal/repository"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (token string, user *model.User, err error)
	Logout(ctx context.Context, token string) error
	ParseToken(ctx context.Context, token string) (*appjwt.Claims, error)
}

type authService struct {
	userRepo  repository.UserRepo
	jwtSecret string
}

func NewAuthService(userRepo repository.UserRepo, jwtSecret string) AuthService {
	return &authService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *authService) Login(ctx context.Context, username, password string) (string, *model.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", nil, apperrors.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, apperrors.ErrUnauthorized
	}

	role := "member"
	if user.IsSuperAdmin {
		role = "superadmin"
	}

	token, err := appjwt.Sign(user.ID, role, s.jwtSecret)
	if err != nil {
		return "", nil, apperrors.ErrInternal
	}

	return token, user, nil
}

func (s *authService) ParseToken(ctx context.Context, token string) (*appjwt.Claims, error) {
	return appjwt.Verify(token, s.jwtSecret)
}

func (s *authService) Logout(_ context.Context, _ string) error {
	return nil
}

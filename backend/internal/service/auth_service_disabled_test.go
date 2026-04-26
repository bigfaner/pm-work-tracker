package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

func TestLogin_DisabledUserCorrectPassword_ReturnsErrUserDisabled(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "disabled").
		Return(&model.User{
			BaseModel:    model.BaseModel{ID: 10},
			Username:     "disabled",
			PasswordHash: prehashedPassword123,
			UserStatus: "disabled",
		}, nil)

	svc := NewAuthService(repo, testJWTSecret)
	_, _, err := svc.Login(context.Background(), "disabled", "password123")
	assert.ErrorIs(t, err, apperrors.ErrUserDisabled)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Status)
	assert.Equal(t, "USER_DISABLED", appErr.Code)
	repo.AssertExpectations(t)
}

func TestLogin_DisabledUserWrongPassword_ReturnsErrUnauthorized(t *testing.T) {
	// Wrong password for a disabled user must still return the generic
	// UNAUTHORIZED error — never reveal account status on bad credentials.
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "disabled").
		Return(&model.User{
			BaseModel:    model.BaseModel{ID: 10},
			Username:     "disabled",
			PasswordHash: prehashedPassword123,
			UserStatus: "disabled",
		}, nil)

	svc := NewAuthService(repo, testJWTSecret)
	_, _, err := svc.Login(context.Background(), "disabled", "wrong-password")
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
	repo.AssertExpectations(t)
}

func TestLogin_EnabledUser_Succeeds(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "active").
		Return(&model.User{
			BaseModel:    model.BaseModel{ID: 11},
			Username:     "active",
			PasswordHash: prehashedPassword123,
			UserStatus: "enabled",
		}, nil)

	svc := NewAuthService(repo, testJWTSecret)
	token, user, err := svc.Login(context.Background(), "active", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, uint(11), user.ID)
	repo.AssertExpectations(t)
}

func TestLogin_DefaultStatus_Succeeds(t *testing.T) {
	// Users created before the Status field was added have Status="" (zero value)
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "legacy").
		Return(&model.User{
			BaseModel:    model.BaseModel{ID: 12},
			Username:     "legacy",
			PasswordHash: prehashedPassword123,
			UserStatus:   "",
		}, nil)

	svc := NewAuthService(repo, testJWTSecret)
	_, _, err := svc.Login(context.Background(), "legacy", "password123")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

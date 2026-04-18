package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

func TestLogin_DisabledUser_ReturnsErrUserDisabled(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "disabled").
		Return(&model.User{
			Model:        gorm.Model{ID: 10},
			Username:     "disabled",
			PasswordHash: prehashedPassword123,
			Status:       "disabled",
		}, nil)

	svc := NewAuthService(repo, testJWTSecret)
	_, _, err := svc.Login(context.Background(), "disabled", "password123")
	assert.ErrorIs(t, err, apperrors.ErrUserDisabled)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Status)
	repo.AssertExpectations(t)
}

func TestLogin_EnabledUser_Succeeds(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "active").
		Return(&model.User{
			Model:        gorm.Model{ID: 11},
			Username:     "active",
			PasswordHash: prehashedPassword123,
			Status:       "enabled",
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
			Model:        gorm.Model{ID: 12},
			Username:     "legacy",
			PasswordHash: prehashedPassword123,
			Status:       "",
		}, nil)

	svc := NewAuthService(repo, testJWTSecret)
	_, _, err := svc.Login(context.Background(), "legacy", "password123")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

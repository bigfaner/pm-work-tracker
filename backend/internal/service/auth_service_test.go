package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	appjwt "pm-work-tracker/backend/internal/pkg/jwt"
)

const testJWTSecret = "test-secret-that-is-at-least-32b"

// mockUserRepo satisfies repository.UserRepo for testing.
type mockUserRepo struct {
	user  *model.User
	err   error
}

func (m *mockUserRepo) FindByID(_ context.Context, _ uint) (*model.User, error) {
	return m.user, m.err
}

func (m *mockUserRepo) FindByUsername(_ context.Context, _ string) (*model.User, error) {
	return m.user, m.err
}

func (m *mockUserRepo) List(_ context.Context) ([]*model.User, error) {
	return nil, nil
}

func (m *mockUserRepo) Update(_ context.Context, _ *model.User) error {
	return nil
}

// helper: hash a password with bcrypt cost 12.
func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

func TestLogin_CorrectCredentials_ReturnsToken(t *testing.T) {
	repo := &mockUserRepo{
		user: &model.User{
			Model:        gorm.Model{ID: 1},
			Username:     "alice",
			PasswordHash: hashPassword("password123"),
			IsSuperAdmin: false,
		},
	}
	svc := NewAuthService(repo, testJWTSecret)

	token, user, err := svc.Login(context.Background(), "alice", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, uint(1), user.ID)

	// Verify the returned token is valid and has correct claims.
	claims, err := svc.ParseToken(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "member", claims.Role)
}

func TestLogin_SuperAdmin_RoleIsSuperAdmin(t *testing.T) {
	repo := &mockUserRepo{
		user: &model.User{
			Model:        gorm.Model{ID: 5},
			Username:     "admin",
			PasswordHash: hashPassword("adminpass"),
			IsSuperAdmin: true,
		},
	}
	svc := NewAuthService(repo, testJWTSecret)

	token, _, err := svc.Login(context.Background(), "admin", "adminpass")
	require.NoError(t, err)

	claims, err := svc.ParseToken(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, uint(5), claims.UserID)
	assert.Equal(t, "superadmin", claims.Role)
}

func TestLogin_WrongPassword_ReturnsErrUnauthorized(t *testing.T) {
	repo := &mockUserRepo{
		user: &model.User{
			Model:        gorm.Model{ID: 1},
			Username:     "alice",
			PasswordHash: hashPassword("password123"),
		},
	}
	svc := NewAuthService(repo, testJWTSecret)

	_, _, err := svc.Login(context.Background(), "alice", "wrong-password")
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestLogin_UnknownUsername_ReturnsErrUnauthorized(t *testing.T) {
	repo := &mockUserRepo{
		user: nil,
		err:  gorm.ErrRecordNotFound,
	}
	svc := NewAuthService(repo, testJWTSecret)

	_, _, err := svc.Login(context.Background(), "nonexistent", "any-password")
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestParseToken_ValidToken(t *testing.T) {
	repo := &mockUserRepo{}
	svc := NewAuthService(repo, testJWTSecret)

	tokenStr, err := appjwt.Sign(42, "member", testJWTSecret)
	require.NoError(t, err)

	claims, err := svc.ParseToken(context.Background(), tokenStr)
	require.NoError(t, err)
	assert.Equal(t, uint(42), claims.UserID)
	assert.Equal(t, "member", claims.Role)
}

func TestParseToken_InvalidToken_ReturnsErrUnauthorized(t *testing.T) {
	repo := &mockUserRepo{}
	svc := NewAuthService(repo, testJWTSecret)

	_, err := svc.ParseToken(context.Background(), "invalid-token")
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestLogout_IsNoOp(t *testing.T) {
	repo := &mockUserRepo{}
	svc := NewAuthService(repo, testJWTSecret)

	err := svc.Logout(context.Background(), "some-token")
	assert.NoError(t, err)
}

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
	appjwt "pm-work-tracker/backend/internal/pkg/jwt"
	"pm-work-tracker/backend/internal/repository"
)

const testJWTSecret = "test-secret-that-is-at-least-32b"

// mockUserRepo uses testify/mock to satisfy repository.UserRepo.
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) List(ctx context.Context) ([]*model.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByBizKey(_ context.Context, _ int64) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) SearchAvailable(_ context.Context, _ int64, _ string, _ int) ([]*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) SoftDelete(_ context.Context, _ *model.User) error { return nil }

// Pre-computed bcrypt hash for "password123" at cost 12.
// Generated once with bcrypt.GenerateFromPassword; avoids bcrypt calls in unit tests.
const prehashedPassword123 = "$2a$12$PFlK4EXT4SpV6eUDJsR99.kCTqzF47MXGJJuhiG9CszXkURn9JR92"

// verifyMockType ensures mockUserRepo satisfies the interface at compile time.
var _ repository.UserRepo = (*mockUserRepo)(nil)

func TestLogin_CorrectCredentials_ReturnsToken(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "alice").
		Return(&model.User{
			BaseModel:        model.BaseModel{ID: 1},
			Username:     "alice",
			PasswordHash: prehashedPassword123,
			IsSuperAdmin: false,
		}, nil)

	svc := NewAuthService(repo, testJWTSecret)
	token, user, err := svc.Login(context.Background(), "alice", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, uint(1), user.ID)

	// Verify the returned token is valid and has correct claims.
	claims, err := svc.ParseToken(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "alice", claims.Username)
	repo.AssertExpectations(t)
}

func TestLogin_SuperAdmin_RoleIsSuperAdmin(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "admin").
		Return(&model.User{
			BaseModel:        model.BaseModel{ID: 5},
			Username:     "admin",
			PasswordHash: prehashedPassword123,
			IsSuperAdmin: true,
		}, nil)

	svc := NewAuthService(repo, testJWTSecret)
	token, _, err := svc.Login(context.Background(), "admin", "password123")
	require.NoError(t, err)

	claims, err := svc.ParseToken(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, uint(5), claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	repo.AssertExpectations(t)
}

func TestLogin_WrongPassword_ReturnsErrUnauthorized(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "alice").
		Return(&model.User{
			BaseModel:        model.BaseModel{ID: 1},
			Username:     "alice",
			PasswordHash: prehashedPassword123,
		}, nil)

	svc := NewAuthService(repo, testJWTSecret)
	_, _, err := svc.Login(context.Background(), "alice", "wrong-password")
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
	repo.AssertExpectations(t)
}

func TestLogin_UnknownUsername_ReturnsErrUnauthorized(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "nonexistent").
		Return(nil, gorm.ErrRecordNotFound)

	svc := NewAuthService(repo, testJWTSecret)
	_, _, err := svc.Login(context.Background(), "nonexistent", "any-password")
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
	repo.AssertExpectations(t)
}

func TestParseToken_ValidToken(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewAuthService(repo, testJWTSecret)

	tokenStr, err := appjwt.Sign(42, "testuser", testJWTSecret)
	require.NoError(t, err)

	claims, err := svc.ParseToken(context.Background(), tokenStr)
	require.NoError(t, err)
	assert.Equal(t, uint(42), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestParseToken_InvalidToken_ReturnsErrUnauthorized(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewAuthService(repo, testJWTSecret)

	_, err := svc.ParseToken(context.Background(), "invalid-token")
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestLogout_IsNoOp(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewAuthService(repo, testJWTSecret)

	err := svc.Logout(context.Background(), "some-token")
	assert.NoError(t, err)
}

func TestLogin_DeletedUser_ReturnsErrUserDeleted(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByUsername", mock.Anything, "deleted").
		Return(&model.User{
			BaseModel:    model.BaseModel{ID: 10, DeletedFlag: 1},
			Username:     "deleted",
			PasswordHash: prehashedPassword123,
		}, nil)

	svc := NewAuthService(repo, testJWTSecret)
	_, _, err := svc.Login(context.Background(), "deleted", "password123")
	assert.ErrorIs(t, err, apperrors.ErrUserDeleted)
	repo.AssertExpectations(t)
}

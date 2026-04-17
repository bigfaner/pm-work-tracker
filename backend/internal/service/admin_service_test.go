package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Mock repos for AdminService tests
// ---------------------------------------------------------------------------

type mockAdminUserRepo struct {
	users []*model.User
	user  *model.User
	err   error
	// capture calls
	updatedUser *model.User
}

func (m *mockAdminUserRepo) FindByID(_ context.Context, _ uint) (*model.User, error) {
	return m.user, m.err
}

func (m *mockAdminUserRepo) FindByUsername(_ context.Context, _ string) (*model.User, error) {
	return m.user, m.err
}

func (m *mockAdminUserRepo) List(_ context.Context) ([]*model.User, error) {
	return m.users, m.err
}

func (m *mockAdminUserRepo) Update(_ context.Context, user *model.User) error {
	m.updatedUser = user
	return nil
}

// mockAdminTeamRepo implements repository.TeamRepo for admin service tests.
type mockAdminTeamRepo struct {
	teams      []*dto.AdminTeamDTO
	listAllErr error
}

func (m *mockAdminTeamRepo) Create(_ context.Context, _ *model.Team) error      { return nil }
func (m *mockAdminTeamRepo) FindByID(_ context.Context, _ uint) (*model.Team, error) {
	return nil, nil
}
func (m *mockAdminTeamRepo) List(_ context.Context) ([]*model.Team, error)       { return nil, nil }
func (m *mockAdminTeamRepo) Update(_ context.Context, _ *model.Team) error       { return nil }
func (m *mockAdminTeamRepo) Delete(_ context.Context, _ uint) error              { return nil }
func (m *mockAdminTeamRepo) AddMember(_ context.Context, _ *model.TeamMember) error {
	return nil
}
func (m *mockAdminTeamRepo) RemoveMember(_ context.Context, _, _ uint) error     { return nil }
func (m *mockAdminTeamRepo) FindMember(_ context.Context, _, _ uint) (*model.TeamMember, error) {
	return nil, nil
}
func (m *mockAdminTeamRepo) ListMembers(_ context.Context, _ uint) ([]*dto.TeamMemberDTO, error) {
	return nil, nil
}
func (m *mockAdminTeamRepo) UpdateMember(_ context.Context, _ *model.TeamMember) error {
	return nil
}
func (m *mockAdminTeamRepo) ListAllTeams(_ context.Context) ([]*dto.AdminTeamDTO, error) {
	return m.teams, m.listAllErr
}

// ---------------------------------------------------------------------------
// Tests: ListUsers
// ---------------------------------------------------------------------------

func TestAdminListUsers_Success(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		users: []*model.User{
			{Model: gorm.Model{ID: 1}, Username: "alice", DisplayName: "Alice", CanCreateTeam: true},
			{Model: gorm.Model{ID: 2}, Username: "bob", DisplayName: "Bob", CanCreateTeam: false},
		},
	}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	users, err := svc.ListUsers(context.Background())
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "alice", users[0].Username)
	assert.Equal(t, "bob", users[1].Username)
}

func TestAdminListUsers_Empty(t *testing.T) {
	userRepo := &mockAdminUserRepo{users: []*model.User{}}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	users, err := svc.ListUsers(context.Background())
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestAdminListUsers_RepoError(t *testing.T) {
	userRepo := &mockAdminUserRepo{err: errors.New("db error")}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	_, err := svc.ListUsers(context.Background())
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: SetCanCreateTeam
// ---------------------------------------------------------------------------

func TestSetCanCreateTeam_CannotModifySelf(t *testing.T) {
	userRepo := &mockAdminUserRepo{}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	err := svc.SetCanCreateTeam(context.Background(), 1, 1, true)
	assert.ErrorIs(t, err, apperrors.ErrCannotModifySelf)
}

func TestSetCanCreateTeam_TargetNotFound(t *testing.T) {
	userRepo := &mockAdminUserRepo{user: nil, err: errors.New("not found")}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	err := svc.SetCanCreateTeam(context.Background(), 1, 99, true)
	assert.Error(t, err)
}

func TestSetCanCreateTeam_Success(t *testing.T) {
	targetUser := &model.User{
		Model:         gorm.Model{ID: 5},
		Username:      "bob",
		DisplayName:   "Bob",
		CanCreateTeam: false,
	}
	userRepo := &mockAdminUserRepo{user: targetUser}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	err := svc.SetCanCreateTeam(context.Background(), 1, 5, true)
	require.NoError(t, err)
	assert.True(t, targetUser.CanCreateTeam)
	assert.Equal(t, userRepo.updatedUser, targetUser)
}

func TestSetCanCreateTeam_Revoke(t *testing.T) {
	targetUser := &model.User{
		Model:         gorm.Model{ID: 5},
		Username:      "bob",
		DisplayName:   "Bob",
		CanCreateTeam: true,
	}
	userRepo := &mockAdminUserRepo{user: targetUser}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	err := svc.SetCanCreateTeam(context.Background(), 1, 5, false)
	require.NoError(t, err)
	assert.False(t, targetUser.CanCreateTeam)
}

// ---------------------------------------------------------------------------
// Tests: ListAllTeams
// ---------------------------------------------------------------------------

func TestAdminListAllTeams_Success(t *testing.T) {
	teamRepo := &mockAdminTeamRepo{
		teams: []*dto.AdminTeamDTO{
			{ID: 1, Name: "Alpha", PMDisplayName: "Alice", MemberCount: 3, MainItemCount: 5, CreatedAt: "2026-01-01T00:00:00Z"},
			{ID: 2, Name: "Beta", PMDisplayName: "Bob", MemberCount: 2, MainItemCount: 10, CreatedAt: "2026-02-01T00:00:00Z"},
		},
	}
	svc := NewAdminService(&mockAdminUserRepo{}, teamRepo)

	teams, err := svc.ListAllTeams(context.Background())
	require.NoError(t, err)
	assert.Len(t, teams, 2)
	assert.Equal(t, "Alpha", teams[0].Name)
	assert.Equal(t, 3, teams[0].MemberCount)
	assert.Equal(t, 5, teams[0].MainItemCount)
}

func TestAdminListAllTeams_Empty(t *testing.T) {
	teamRepo := &mockAdminTeamRepo{teams: []*dto.AdminTeamDTO{}}
	svc := NewAdminService(&mockAdminUserRepo{}, teamRepo)

	teams, err := svc.ListAllTeams(context.Background())
	require.NoError(t, err)
	assert.Empty(t, teams)
}

func TestAdminListAllTeams_RepoError(t *testing.T) {
	teamRepo := &mockAdminTeamRepo{listAllErr: errors.New("db error")}
	svc := NewAdminService(&mockAdminUserRepo{}, teamRepo)

	_, err := svc.ListAllTeams(context.Background())
	assert.Error(t, err)
}

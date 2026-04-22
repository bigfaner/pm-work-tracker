package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Mock repos for AdminService tests
// ---------------------------------------------------------------------------

type mockAdminUserRepo struct {
	users   []*model.User
	user    *model.User
	err     error
	created *model.User
	updated *model.User
	findErr error // separate error for FindByUsername

	// ListFiltered captures calls and returns configurable results
	listFilteredFn func(ctx context.Context, search string, offset, limit int) ([]*model.User, int64, error)
	listFilteredCalled bool
	listFilteredSearch string
	listFilteredOffset int
	listFilteredLimit  int
}

func (m *mockAdminUserRepo) FindByID(_ context.Context, _ uint) (*model.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.user, m.err
}

func (m *mockAdminUserRepo) FindByUsername(_ context.Context, username string) (*model.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	// Look through users list for matching username
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockAdminUserRepo) List(_ context.Context) ([]*model.User, error) {
	return m.users, m.err
}

func (m *mockAdminUserRepo) Create(_ context.Context, user *model.User) error {
	m.created = user
	user.ID = 100 // simulate auto-increment
	return nil
}

func (m *mockAdminUserRepo) Update(_ context.Context, user *model.User) error {
	m.updated = user
	return nil
}

func (m *mockAdminUserRepo) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.User, error) {
	return nil, nil
}

func (m *mockAdminUserRepo) ListFiltered(ctx context.Context, search string, offset, limit int) ([]*model.User, int64, error) {
	m.listFilteredCalled = true
	m.listFilteredSearch = search
	m.listFilteredOffset = offset
	m.listFilteredLimit = limit
	if m.listFilteredFn != nil {
		return m.listFilteredFn(ctx, search, offset, limit)
	}
	return nil, 0, nil
}

func (m *mockAdminUserRepo) SearchAvailable(_ context.Context, _ uint, _ string, _ int) ([]*model.User, error) {
	return nil, nil
}

// mockAdminTeamRepo implements repository.TeamRepo for admin service tests.
type mockAdminTeamRepo struct {
	teams           []*dto.AdminTeamDTO
	listAllErr      error
	teamByID        *model.Team
	teamByIDErr     error
	addMemberErr    error
	removeMemberErr error
	teamsByUserIDs  map[uint][]dto.TeamSummary
	teamsByUIDErr   error
}

func (m *mockAdminTeamRepo) Create(_ context.Context, _ *model.Team) error      { return nil }
func (m *mockAdminTeamRepo) FindByID(_ context.Context, _ uint) (*model.Team, error) {
	return m.teamByID, m.teamByIDErr
}
func (m *mockAdminTeamRepo) List(_ context.Context) ([]*model.Team, error)       { return nil, nil }
func (m *mockAdminTeamRepo) Update(_ context.Context, _ *model.Team) error       { return nil }
func (m *mockAdminTeamRepo) Delete(_ context.Context, _ uint) error              { return nil }
func (m *mockAdminTeamRepo) AddMember(_ context.Context, member *model.TeamMember) error {
	return m.addMemberErr
}
func (m *mockAdminTeamRepo) RemoveMember(_ context.Context, _, _ uint) error {
	return m.removeMemberErr
}
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
func (m *mockAdminTeamRepo) FindTeamsByUserIDs(_ context.Context, _ []uint) (map[uint][]dto.TeamSummary, error) {
	if m.teamsByUIDErr != nil {
		return nil, m.teamsByUIDErr
	}
	if m.teamsByUserIDs != nil {
		return m.teamsByUserIDs, nil
	}
	return map[uint][]dto.TeamSummary{}, nil
}

func (m *mockAdminTeamRepo) FindPMMembers(_ context.Context, _ []uint) (map[uint]string, error) {
	return map[uint]string{}, nil
}

// ---------------------------------------------------------------------------
// Tests: ListUsers
// ---------------------------------------------------------------------------

func TestAdminListUsers_Success(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		listFilteredFn: func(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
			return []*model.User{
				{BaseModel: model.BaseModel{ID: 1}, Username: "alice", DisplayName: "Alice"},
				{BaseModel: model.BaseModel{ID: 2}, Username: "bob", DisplayName: "Bob"},
			}, 2, nil
		},
	}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	items, total, err := svc.ListUsers(context.Background(), "", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 2, int(total))
	assert.Len(t, items, 2)
	assert.Equal(t, "alice", items[0].Username)
	assert.Equal(t, "bob", items[1].Username)
	assert.True(t, userRepo.listFilteredCalled)
}

func TestAdminListUsers_Empty(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		listFilteredFn: func(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
			return []*model.User{}, 0, nil
		},
	}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	items, total, err := svc.ListUsers(context.Background(), "", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 0, int(total))
	assert.Empty(t, items)
}

func TestAdminListUsers_RepoError(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		listFilteredFn: func(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	_, _, err := svc.ListUsers(context.Background(), "", 1, 20)
	assert.Error(t, err)
}

func TestAdminListUsers_SearchFilter(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		listFilteredFn: func(_ context.Context, search string, _, _ int) ([]*model.User, int64, error) {
			if search == "alice" {
				return []*model.User{
					{BaseModel: model.BaseModel{ID: 1}, Username: "alice", DisplayName: "Alice Wonderland"},
					{BaseModel: model.BaseModel{ID: 3}, Username: "charlie", DisplayName: "Alice Cooper"},
				}, 2, nil
			}
			return []*model.User{}, 0, nil
		},
	}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	items, total, err := svc.ListUsers(context.Background(), "alice", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 2, int(total))
	assert.Len(t, items, 2)
	assert.Equal(t, "alice", items[0].Username)
	assert.Equal(t, "charlie", items[1].Username)
	assert.Equal(t, "alice", userRepo.listFilteredSearch)
}

func TestAdminListUsers_Pagination(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		listFilteredFn: func(_ context.Context, _ string, offset, limit int) ([]*model.User, int64, error) {
			all := []*model.User{
				{BaseModel: model.BaseModel{ID: 1}, Username: "u1"},
				{BaseModel: model.BaseModel{ID: 2}, Username: "u2"},
				{BaseModel: model.BaseModel{ID: 3}, Username: "u3"},
				{BaseModel: model.BaseModel{ID: 4}, Username: "u4"},
				{BaseModel: model.BaseModel{ID: 5}, Username: "u5"},
			}
			start := offset
			if start > len(all) {
				start = len(all)
			}
			end := offset + limit
			if end > len(all) {
				end = len(all)
			}
			return all[start:end], 5, nil
		},
	}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	items, total, err := svc.ListUsers(context.Background(), "", 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, int(total))
	assert.Len(t, items, 2)
	assert.Equal(t, "u3", items[0].Username)
	assert.Equal(t, "u4", items[1].Username)
	assert.Equal(t, 2, userRepo.listFilteredOffset)
	assert.Equal(t, 2, userRepo.listFilteredLimit)
}

func TestAdminListUsers_WithTeams(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		listFilteredFn: func(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
			return []*model.User{
				{BaseModel: model.BaseModel{ID: 1}, Username: "alice", DisplayName: "Alice"},
			}, 1, nil
		},
	}
	teamRepo := &mockAdminTeamRepo{
		teamsByUserIDs: map[uint][]dto.TeamSummary{
			1: {{ID: 10, Name: "Team A", Role: "member"}},
		},
	}
	svc := NewAdminService(userRepo, teamRepo)

	items, total, err := svc.ListUsers(context.Background(), "", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, int(total))
	require.Len(t, items, 1)
	require.Len(t, items[0].Teams, 1)
	assert.Equal(t, uint(10), items[0].Teams[0].ID)
	assert.Equal(t, "Team A", items[0].Teams[0].Name)
}

// ---------------------------------------------------------------------------
// Tests: GetUser
// ---------------------------------------------------------------------------

func TestAdminGetUser_Success(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 5}, Username: "bob", DisplayName: "Bob", Email: "bob@test.com", Status: "enabled"},
	}
	teamRepo := &mockAdminTeamRepo{
		teamsByUserIDs: map[uint][]dto.TeamSummary{
			5: {{ID: 1, Name: "Team A", Role: "member"}},
		},
	}
	svc := NewAdminService(userRepo, teamRepo)

	user, err := svc.GetUser(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, uint(5), user.ID)
	assert.Equal(t, "bob", user.Username)
	assert.Equal(t, "bob@test.com", user.Email)
	require.Len(t, user.Teams, 1)
	assert.Equal(t, "Team A", user.Teams[0].Name)
}

func TestAdminGetUser_NotFound(t *testing.T) {
	userRepo := &mockAdminUserRepo{findErr: apperrors.ErrNotFound}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	_, err := svc.GetUser(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Tests: CreateUser
// ---------------------------------------------------------------------------

func TestAdminCreateUser_Success(t *testing.T) {
	userRepo := &mockAdminUserRepo{}
	teamRepo := &mockAdminTeamRepo{
		teamByID: &model.Team{BaseModel: model.BaseModel{ID: 10}, Name: "Team A"},
		teamsByUserIDs: map[uint][]dto.TeamSummary{
			100: {{ID: 10, Name: "Team A", Role: "member"}},
		},
	}
	svc := NewAdminService(userRepo, teamRepo)

	teamID := uint(10)
	req := &dto.CreateUserReq{
		Username:    "newuser",
		DisplayName: "New User",
		Email:       "new@test.com",
		TeamID:      &teamID,
	}

	user, err := svc.CreateUser(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "newuser", user.Username)
	assert.Equal(t, "New User", user.DisplayName)
	assert.Equal(t, "new@test.com", user.Email)
	assert.Equal(t, "enabled", user.Status)
	assert.NotEmpty(t, user.InitialPassword)
	assert.Len(t, user.InitialPassword, 12)

	// Verify password was hashed
	assert.NotEmpty(t, userRepo.created.PasswordHash)
	assert.NotEqual(t, user.InitialPassword, userRepo.created.PasswordHash)
}

func TestAdminCreateUser_DuplicateUsername(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		users: []*model.User{
			{BaseModel: model.BaseModel{ID: 1}, Username: "existing"},
		},
	}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	req := &dto.CreateUserReq{
		Username:    "existing",
		DisplayName: "Existing User",
	}

	_, err := svc.CreateUser(context.Background(), req)
	assert.ErrorIs(t, err, apperrors.ErrUserExists)
}

func TestAdminCreateUser_TeamNotFound(t *testing.T) {
	userRepo := &mockAdminUserRepo{}
	teamRepo := &mockAdminTeamRepo{
		teamByIDErr: apperrors.ErrNotFound,
	}
	svc := NewAdminService(userRepo, teamRepo)

	teamID := uint(999)
	req := &dto.CreateUserReq{
		Username:    "newuser",
		DisplayName: "New User",
		TeamID:      &teamID,
	}

	_, err := svc.CreateUser(context.Background(), req)
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestAdminCreateUser_NoTeam(t *testing.T) {
	userRepo := &mockAdminUserRepo{}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	req := &dto.CreateUserReq{
		Username:    "newuser",
		DisplayName: "New User",
	}

	user, err := svc.CreateUser(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "newuser", user.Username)
	assert.Empty(t, user.Teams)
	assert.NotEmpty(t, user.InitialPassword)
}

// ---------------------------------------------------------------------------
// Tests: UpdateUser
// ---------------------------------------------------------------------------

func TestAdminUpdateUser_Success(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 5}, Username: "bob", DisplayName: "Bob", Email: "bob@test.com"},
	}
	teamRepo := &mockAdminTeamRepo{
		teamsByUserIDs: map[uint][]dto.TeamSummary{
			5: {{ID: 2, Name: "Team B", Role: "member"}},
		},
	}
	svc := NewAdminService(userRepo, teamRepo)

	req := &dto.UpdateUserReq{
		DisplayName: strPtr("Robert"),
		Email:       strPtr("robert@test.com"),
	}

	user, err := svc.UpdateUser(context.Background(), 5, req)
	require.NoError(t, err)
	assert.Equal(t, "Robert", user.DisplayName)
	assert.Equal(t, "robert@test.com", user.Email)
	assert.Equal(t, "Robert", userRepo.updated.DisplayName)
}

func TestAdminUpdateUser_NotFound(t *testing.T) {
	userRepo := &mockAdminUserRepo{findErr: apperrors.ErrNotFound}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	req := &dto.UpdateUserReq{DisplayName: strPtr("Robert")}

	_, err := svc.UpdateUser(context.Background(), 999, req)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Tests: ToggleUserStatus
// ---------------------------------------------------------------------------

func TestAdminToggleUserStatus_DisableSuccess(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 5}, Username: "bob", Status: "enabled"},
	}
	teamRepo := &mockAdminTeamRepo{
		teamsByUserIDs: map[uint][]dto.TeamSummary{},
	}
	svc := NewAdminService(userRepo, teamRepo)

	user, err := svc.ToggleUserStatus(context.Background(), 1, 5, "disabled")
	require.NoError(t, err)
	assert.Equal(t, "disabled", user.Status)
	assert.Equal(t, "disabled", userRepo.updated.Status)
}

func TestAdminToggleUserStatus_EnableSuccess(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 5}, Username: "bob", Status: "disabled"},
	}
	teamRepo := &mockAdminTeamRepo{
		teamsByUserIDs: map[uint][]dto.TeamSummary{},
	}
	svc := NewAdminService(userRepo, teamRepo)

	user, err := svc.ToggleUserStatus(context.Background(), 1, 5, "enabled")
	require.NoError(t, err)
	assert.Equal(t, "enabled", user.Status)
}

func TestAdminToggleUserStatus_CannotDisableSelf(t *testing.T) {
	userRepo := &mockAdminUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 1}, Username: "admin", Status: "enabled"},
	}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	_, err := svc.ToggleUserStatus(context.Background(), 1, 1, "disabled")
	assert.ErrorIs(t, err, apperrors.ErrCannotDisableSelf)
}

func TestAdminToggleUserStatus_UserNotFound(t *testing.T) {
	userRepo := &mockAdminUserRepo{findErr: apperrors.ErrNotFound}
	svc := NewAdminService(userRepo, &mockAdminTeamRepo{})

	_, err := svc.ToggleUserStatus(context.Background(), 1, 999, "disabled")
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
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

// ---------------------------------------------------------------------------
// Tests: generatePassword
// ---------------------------------------------------------------------------

func TestGeneratePassword_Length(t *testing.T) {
	pw, err := generatePassword(12)
	require.NoError(t, err)
	assert.Len(t, pw, 12)
}

func TestGeneratePassword_Charset(t *testing.T) {
	pw, err := generatePassword(100)
	require.NoError(t, err)
	for _, c := range pw {
		assert.True(t, (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	}
}

func TestGeneratePassword_Uniqueness(t *testing.T) {
	pw1, _ := generatePassword(12)
	pw2, _ := generatePassword(12)
	assert.NotEqual(t, pw1, pw2, "two generated passwords should differ")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func strPtr(s string) *string {
	return &s
}

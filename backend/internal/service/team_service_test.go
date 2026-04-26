package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// ---------------------------------------------------------------------------
// Mock repos
// ---------------------------------------------------------------------------

type mockTeamRepo struct {
	team    *model.Team
	teams   []*model.Team
	members []*dto.TeamMemberDTO
	member  *model.TeamMember
	// findMembersByUserID maps userID -> TeamMember for multi-lookup scenarios
	findMembersByUserID map[uint]*model.TeamMember
	// per-operation errors (nil = success)
	findByIDErr   error
	findMemberErr error
	addMemberErr  error
	updateErr     error
	deleteErr     error
	createErr     error
	listErr       error
	// CountMembers support
	countMembersVal    int64
	countMembersErr    error
	countMembersCalled bool
	listMembersCalled  bool
	// capture calls
	createdTeam    *model.Team
	createdMember  *model.TeamMember
	updatedTeam    *model.Team
	updatedMembers []*model.TeamMember
	deletedTeamID  uint
	removedMember  struct {
		teamID uint
		userID uint
	}
}

func (m *mockTeamRepo) Create(_ context.Context, team *model.Team) error {
	m.createdTeam = team
	if m.createErr != nil {
		return m.createErr
	}
	team.ID = 1
	return nil
}

func (m *mockTeamRepo) FindByID(_ context.Context, teamID uint) (*model.Team, error) {
	if m.team != nil {
		return m.team, nil
	}
	return nil, m.findByIDErr
}

func (m *mockTeamRepo) List(_ context.Context) ([]*model.Team, error) {
	return m.teams, m.listErr
}

func (m *mockTeamRepo) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.Team, int64, error) {
	return m.teams, int64(len(m.teams)), m.listErr
}

func (m *mockTeamRepo) Update(_ context.Context, team *model.Team) error {
	m.updatedTeam = team
	return m.updateErr
}

func (m *mockTeamRepo) SoftDelete(_ context.Context, teamID uint) error {
	m.deletedTeamID = teamID
	return m.deleteErr
}

func (m *mockTeamRepo) FindByBizKey(_ context.Context, _ int64) (*model.Team, error) {
	return nil, nil
}

func (m *mockTeamRepo) AddMember(_ context.Context, member *model.TeamMember) error {
	m.createdMember = member
	return m.addMemberErr
}

func (m *mockTeamRepo) RemoveMember(_ context.Context, teamID, userID uint) error {
	m.removedMember.teamID = teamID
	m.removedMember.userID = userID
	return nil
}

func (m *mockTeamRepo) FindMember(_ context.Context, teamID, userID uint) (*model.TeamMember, error) {
	if m.findMembersByUserID != nil {
		if member, ok := m.findMembersByUserID[userID]; ok {
			return member, nil
		}
		return nil, m.findMemberErr
	}
	if m.member != nil {
		return m.member, nil
	}
	return nil, m.findMemberErr
}

func (m *mockTeamRepo) ListMembers(_ context.Context, teamID uint) ([]*dto.TeamMemberDTO, error) {
	m.listMembersCalled = true
	var result []*dto.TeamMemberDTO
	for _, mem := range m.members {
		if mem.TeamKey == fmt.Sprintf("%d", teamID) {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (m *mockTeamRepo) CountMembers(_ context.Context, _ uint) (int64, error) {
	m.countMembersCalled = true
	if m.countMembersErr != nil {
		return 0, m.countMembersErr
	}
	return m.countMembersVal, nil
}

func (m *mockTeamRepo) UpdateMember(_ context.Context, member *model.TeamMember) error {
	m.updatedMembers = append(m.updatedMembers, member)
	return m.updateErr
}

func (m *mockTeamRepo) ListAllTeams(_ context.Context) ([]*dto.AdminTeamDTO, error) {
	return nil, nil
}

func (m *mockTeamRepo) FindTeamsByUserIDs(_ context.Context, _ []uint) (map[uint][]dto.TeamSummary, error) {
	return map[uint][]dto.TeamSummary{}, nil
}

func (m *mockTeamRepo) FindPMMembers(_ context.Context, _ []uint) (map[uint]string, error) {
	result := make(map[uint]string)
	for _, mem := range m.members {
		if mem.Role == "pm" {
			teamID, _ := strconv.ParseUint(mem.TeamKey, 10, 64)
			result[uint(teamID)] = mem.DisplayName
		}
	}
	return result, nil
}

// mockTeamUserRepo is a separate mock for UserRepo used in team_service tests.
type mockTeamUserRepo struct {
	user *model.User
	err  error

	// SearchAvailable captures calls and returns configurable results
	searchAvailableFn     func(ctx context.Context, teamID uint, search string, limit int) ([]*model.User, error)
	searchAvailableCalled bool
	searchAvailableTeamID uint
	searchAvailableSearch string
	searchAvailableLimit  int
}

func (m *mockTeamUserRepo) FindByID(_ context.Context, _ uint) (*model.User, error) {
	return m.user, m.err
}

func (m *mockTeamUserRepo) FindByUsername(_ context.Context, _ string) (*model.User, error) {
	return m.user, m.err
}

func (m *mockTeamUserRepo) List(_ context.Context) ([]*model.User, error) {
	return nil, nil
}

func (m *mockTeamUserRepo) Update(_ context.Context, _ *model.User) error {
	return nil
}

func (m *mockTeamUserRepo) Create(_ context.Context, _ *model.User) error {
	return nil
}
func (m *mockTeamUserRepo) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.User, error) {
	return nil, nil
}
func (m *mockTeamUserRepo) FindByBizKey(_ context.Context, _ int64) (*model.User, error) {
	return nil, nil
}
func (m *mockTeamUserRepo) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
	return nil, 0, nil
}
func (m *mockTeamUserRepo) SearchAvailable(ctx context.Context, teamID uint, search string, limit int) ([]*model.User, error) {
	m.searchAvailableCalled = true
	m.searchAvailableTeamID = teamID
	m.searchAvailableSearch = search
	m.searchAvailableLimit = limit
	if m.searchAvailableFn != nil {
		return m.searchAvailableFn(ctx, teamID, search, limit)
	}
	return nil, nil
}

// mockDB used for transaction-based tests (TransferPM).
type mockDB struct {
	called bool
	err    error
}

func (m *mockDB) Transaction(fn func(txDB *gorm.DB) error, _ ...*sql.TxOptions) error {
	m.called = true
	if m.err != nil {
		return m.err
	}
	return fn(nil)
}

// ---------------------------------------------------------------------------
// Tests: CreateTeam
// ---------------------------------------------------------------------------

func TestCreateTeam_Success(t *testing.T) {
	repo := &mockTeamRepo{}
	userRepo := &mockTeamUserRepo{}
	pmRole := &model.Role{Name: "pm"}
	pmRole.ID = 2
	roleRepo := &mockRoleRepo{roleByName: pmRole}
	svc := NewTeamService(repo, userRepo, &mockMainItemRepo{}, roleRepo, &mockDB{})

	team, err := svc.CreateTeam(context.Background(), 1, dto.CreateTeamReq{
		Name: "Alpha Team",
		Description: "A test team",
		Code:        "ALPHA",
	})
	require.NoError(t, err)
	assert.Equal(t, "Alpha Team", team.TeamName)
	assert.Equal(t, "A test team", team.TeamDesc)
	assert.Equal(t, uint(1), uint(team.PmKey))

	// Verify creator was added as member with pm role ID
	assert.NotNil(t, repo.createdMember)
	assert.Equal(t, uint(1), uint(repo.createdMember.UserKey))
	require.NotNil(t, repo.createdMember.RoleKey)
	assert.Equal(t, int64(2), *repo.createdMember.RoleKey)
}

func TestCreateTeam_CodeFieldPersisted(t *testing.T) {
	repo := &mockTeamRepo{}
	svc := NewTeamService(repo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	team, err := svc.CreateTeam(context.Background(), 1, dto.CreateTeamReq{
		Name: "Beta Team",
		Code: "BETA",
	})
	require.NoError(t, err)
	assert.Equal(t, "BETA", team.Code)
	assert.Equal(t, "BETA", repo.createdTeam.Code)
}

func TestCreateTeam_RepoError(t *testing.T) {
	repo := &mockTeamRepo{createErr: errors.New("db error")}
	userRepo := &mockTeamUserRepo{}
	svc := NewTeamService(repo, userRepo, &mockMainItemRepo{}, nil, &mockDB{})

	_, err := svc.CreateTeam(context.Background(), 1, dto.CreateTeamReq{Name: "Team"})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: GetTeam
// ---------------------------------------------------------------------------

func TestGetTeam_Success(t *testing.T) {
	repo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, TeamName: "Alpha", PmKey: 10},
	}
	svc := NewTeamService(repo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	team, err := svc.GetTeam(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), team.ID)
	assert.Equal(t, "Alpha", team.TeamName)
}

func TestGetTeam_NotFound(t *testing.T) {
	repo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(repo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	_, err := svc.GetTeam(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

// ---------------------------------------------------------------------------
// Tests: ListTeams
// ---------------------------------------------------------------------------

func TestListTeams_Success(t *testing.T) {
	repo := &mockTeamRepo{
		teams: []*model.Team{
			{BaseModel: model.BaseModel{ID: 1}, TeamName: "Team A", PmKey: 10},
			{BaseModel: model.BaseModel{ID: 2}, TeamName: "Team B", PmKey: 20},
		},
		members: []*dto.TeamMemberDTO{
			{TeamKey: "1", UserBizKey: "10", Role: "pm", DisplayName: "Alice"},
			{TeamKey: "2", UserBizKey: "20", Role: "pm", DisplayName: "Bob"},
		},
	}
	svc := NewTeamService(repo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	teams, total, err := svc.ListTeams(context.Background(), 1, false, "", 1, 20)
	require.NoError(t, err)
	require.Len(t, teams, 2)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, "Alice", teams[0].PmDisplayName)
	assert.Equal(t, "Bob", teams[1].PmDisplayName)
}

// ---------------------------------------------------------------------------
// Tests: InviteMember
// ---------------------------------------------------------------------------

func TestInviteMember_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team:          &model.Team{BaseModel: model.BaseModel{ID: 1}, TeamName: "Alpha", PmKey: 10},
		member:        nil, // invited user is not yet a member
		findMemberErr: apperrors.ErrNotFound, // FindMember returns not found
	}
	userRepo := &mockTeamUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 5}, Username: "bob"},
	}
	svc := NewTeamService(teamRepo, userRepo, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.InviteMember(context.Background(), 10, 1, dto.InviteMemberReq{Username: "bob"})
	require.NoError(t, err)

	// Verify member was added
	assert.NotNil(t, teamRepo.createdMember)
	assert.Equal(t, uint(5), uint(teamRepo.createdMember.UserKey))
	assert.Equal(t, uint(1), uint(teamRepo.createdMember.TeamKey))
}

func TestInviteMember_TeamNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.InviteMember(context.Background(), 10, 999, dto.InviteMemberReq{Username: "bob"})
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestInviteMember_UserNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
	}
	userRepo := &mockTeamUserRepo{err: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, userRepo, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.InviteMember(context.Background(), 10, 1, dto.InviteMemberReq{Username: "ghost"})
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestInviteMember_AlreadyMember(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
		member: &model.TeamMember{TeamKey: int64(1), UserKey: int64(5)},
	}
	userRepo := &mockTeamUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 5}, Username: "bob"},
	}
	svc := NewTeamService(teamRepo, userRepo, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.InviteMember(context.Background(), 10, 1, dto.InviteMemberReq{Username: "bob"})
	assert.ErrorIs(t, err, apperrors.ErrAlreadyMember)
}

func TestInviteMember_CannotAssignPMRole(t *testing.T) {
	pmRole := &model.Role{Name: "pm"}
	pmRole.ID = 2
	roleRepo := &mockRoleRepo{roleByID: pmRole}
	teamRepo := &mockTeamRepo{
		team:          &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
		findMemberErr: apperrors.ErrNotFound,
	}
	userRepo := &mockTeamUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 5}, Username: "bob"},
	}
	svc := NewTeamService(teamRepo, userRepo, &mockMainItemRepo{}, roleRepo, &mockDB{})

	err := svc.InviteMember(context.Background(), 10, 1, dto.InviteMemberReq{Username: "bob", RoleKey: "2"})
	assert.ErrorIs(t, err, apperrors.ErrCannotAssignPMRole)
}

// TestInviteMember_CallerNotPM is removed: PM check moved to RequirePermission middleware.

// ---------------------------------------------------------------------------
// Tests: RemoveMember
// ---------------------------------------------------------------------------

func TestRemoveMember_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.RemoveMember(context.Background(), 10, 1, 5)
	require.NoError(t, err)
	assert.Equal(t, uint(1), teamRepo.removedMember.teamID)
	assert.Equal(t, uint(5), teamRepo.removedMember.userID)
}

func TestRemoveMember_TeamNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.RemoveMember(context.Background(), 10, 999, 5)
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestRemoveMember_CannotRemoveSelf(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.RemoveMember(context.Background(), 10, 1, 10)
	assert.ErrorIs(t, err, apperrors.ErrCannotRemoveSelf)
}

func TestRemoveMember_CallerNotPM(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.RemoveMember(context.Background(), 99, 1, 5)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

// ---------------------------------------------------------------------------
// Tests: TransferPM
// ---------------------------------------------------------------------------

func TestTransferPM_Success(t *testing.T) {
	db := &mockDB{}
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, TeamName: "Alpha", PmKey: 10},
		findMembersByUserID: map[uint]*model.TeamMember{
			20: {BaseModel: model.BaseModel{ID: 2}, TeamKey: int64(1), UserKey: int64(20)},
			10: {BaseModel: model.BaseModel{ID: 1}, TeamKey: int64(1), UserKey: int64(10)},
		},
	}
	pmRole := &model.Role{Name: "pm"}
	pmRole.ID = 2
	memberRole := &model.Role{Name: "member"}
	memberRole.ID = 3
	roleRepo := &mockRoleRepo{roles: []model.Role{*pmRole, *memberRole}}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, roleRepo, db)

	err := svc.TransferPM(context.Background(), 10, 1, 20)
	require.NoError(t, err)
	assert.True(t, db.called)

	// Verify team PM was updated
	assert.NotNil(t, teamRepo.updatedTeam)
	assert.Equal(t, uint(20), uint(teamRepo.updatedTeam.PmKey))

	// Verify two member role updates: new PM -> pm roleID, old PM -> member roleID
	require.Len(t, teamRepo.updatedMembers, 2)
	require.NotNil(t, teamRepo.updatedMembers[0].RoleKey)
	assert.Equal(t, int64(2), *teamRepo.updatedMembers[0].RoleKey)
	assert.Equal(t, uint(20), uint(teamRepo.updatedMembers[0].UserKey))
	require.NotNil(t, teamRepo.updatedMembers[1].RoleKey)
	assert.Equal(t, int64(3), *teamRepo.updatedMembers[1].RoleKey)
	assert.Equal(t, uint(10), uint(teamRepo.updatedMembers[1].UserKey))
}

func TestTransferPM_TeamNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.TransferPM(context.Background(), 10, 999, 20)
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestTransferPM_CallerNotPM(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.TransferPM(context.Background(), 99, 1, 20)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestTransferPM_TargetNotMember(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team:          &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
		member:        nil,
		findMemberErr: gorm.ErrRecordNotFound,
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.TransferPM(context.Background(), 10, 1, 99)
	assert.ErrorIs(t, err, apperrors.ErrNotTeamMember)
}

// ---------------------------------------------------------------------------
// Tests: DisbandTeam
// ---------------------------------------------------------------------------

func TestDisbandTeam_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, TeamName: "Alpha Team", PmKey: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.DisbandTeam(context.Background(), 10, 1, "Alpha Team")
	require.NoError(t, err)
	assert.Equal(t, uint(1), teamRepo.deletedTeamID)
}

func TestDisbandTeam_TeamNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.DisbandTeam(context.Background(), 10, 999, "anything")
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestDisbandTeam_CallerNotPM(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, TeamName: "Alpha", PmKey: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.DisbandTeam(context.Background(), 99, 1, "Alpha")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestDisbandTeam_ConfirmNameMismatch(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, TeamName: "Alpha Team", PmKey: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.DisbandTeam(context.Background(), 10, 1, "Wrong Name")
	assert.ErrorIs(t, err, apperrors.ErrValidation)
}

// ---------------------------------------------------------------------------
// Tests: ListMembers
// ---------------------------------------------------------------------------

func TestListMembers_Success(t *testing.T) {
	now := time.Now()
	teamRepo := &mockTeamRepo{
		members: []*dto.TeamMemberDTO{
			{TeamKey: "1", UserBizKey: "10", DisplayName: "Alice", Username: "alice", Role: "pm", JoinedAt: now.Format(time.RFC3339)},
			{TeamKey: "1", UserBizKey: "5", DisplayName: "Bob", Username: "bob", JoinedAt: now.Format(time.RFC3339)},
		},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	members, err := svc.ListMembers(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Equal(t, "Alice", members[0].DisplayName)
	assert.Equal(t, "pm", members[0].Role)
	assert.Equal(t, "Bob", members[1].DisplayName)
}

func TestListMembers_Empty(t *testing.T) {
	teamRepo := &mockTeamRepo{members: []*dto.TeamMemberDTO{}}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	members, err := svc.ListMembers(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, members)
}

// ---------------------------------------------------------------------------
// Tests: UpdateMemberRole
// ---------------------------------------------------------------------------

func TestUpdateMemberRole_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team:   &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
		member: &model.TeamMember{BaseModel: model.BaseModel{ID: 1}, TeamKey: int64(1), UserKey: int64(5)},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.UpdateMemberRole(context.Background(), 10, 1, 5, 3)
	require.NoError(t, err)
	require.Len(t, teamRepo.updatedMembers, 1)
	require.NotNil(t, teamRepo.updatedMembers[0].RoleKey)
	assert.Equal(t, int64(3), *teamRepo.updatedMembers[0].RoleKey)
}

func TestUpdateMemberRole_TeamNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.UpdateMemberRole(context.Background(), 10, 999, 5, 3)
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestUpdateMemberRole_CallerNotPM(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.UpdateMemberRole(context.Background(), 99, 1, 5, 3)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestUpdateMemberRole_TargetNotMember(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team:          &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
		member:        nil,
		findMemberErr: gorm.ErrRecordNotFound,
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, nil, &mockDB{})

	err := svc.UpdateMemberRole(context.Background(), 10, 1, 99, 3)
	assert.ErrorIs(t, err, apperrors.ErrNotTeamMember)
}

func TestUpdateMemberRole_CannotAssignPMRole(t *testing.T) {
	pmRole := &model.Role{Name: "pm"}
	pmRole.ID = 2
	roleRepo := &mockRoleRepo{roleByID: pmRole}
	teamRepo := &mockTeamRepo{
		team:   &model.Team{BaseModel: model.BaseModel{ID: 1}, PmKey: 10},
		member: &model.TeamMember{BaseModel: model.BaseModel{ID: 1}, TeamKey: int64(1), UserKey: int64(5)},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, roleRepo, &mockDB{})

	err := svc.UpdateMemberRole(context.Background(), 10, 1, 5, 2)
	assert.ErrorIs(t, err, apperrors.ErrCannotAssignPMRole)
}

// ---------------------------------------------------------------------------
// Tests: SearchAvailableUsers
// ---------------------------------------------------------------------------

func TestSearchAvailableUsers_Success(t *testing.T) {
	userRepo := &mockTeamUserRepo{
		searchAvailableFn: func(_ context.Context, _ uint, _ string, _ int) ([]*model.User, error) {
			return []*model.User{
				{BaseModel: model.BaseModel{ID: 10, BizKey: 10}, Username: "alice", DisplayName: "Alice"},
				{BaseModel: model.BaseModel{ID: 20, BizKey: 20}, Username: "bob", DisplayName: "Bob"},
			}, nil
		},
	}
	svc := NewTeamService(&mockTeamRepo{}, userRepo, &mockMainItemRepo{}, nil, &mockDB{})

	result, err := svc.SearchAvailableUsers(context.Background(), 1, "ali")
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "10", result[0].BizKey)
	assert.Equal(t, "alice", result[0].Username)
	assert.True(t, userRepo.searchAvailableCalled)
	assert.Equal(t, uint(1), userRepo.searchAvailableTeamID)
	assert.Equal(t, "ali", userRepo.searchAvailableSearch)
	assert.Equal(t, 20, userRepo.searchAvailableLimit)
}

func TestSearchAvailableUsers_Empty(t *testing.T) {
	userRepo := &mockTeamUserRepo{
		searchAvailableFn: func(_ context.Context, _ uint, _ string, _ int) ([]*model.User, error) {
			return []*model.User{}, nil
		},
	}
	svc := NewTeamService(&mockTeamRepo{}, userRepo, &mockMainItemRepo{}, nil, &mockDB{})

	result, err := svc.SearchAvailableUsers(context.Background(), 1, "")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSearchAvailableUsers_RepoError(t *testing.T) {
	userRepo := &mockTeamUserRepo{
		searchAvailableFn: func(_ context.Context, _ uint, _ string, _ int) ([]*model.User, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewTeamService(&mockTeamRepo{}, userRepo, &mockMainItemRepo{}, nil, &mockDB{})

	_, err := svc.SearchAvailableUsers(context.Background(), 1, "")
	assert.Error(t, err)
}

func TestSearchAvailableUsers_NilResult(t *testing.T) {
	userRepo := &mockTeamUserRepo{
		searchAvailableFn: func(_ context.Context, _ uint, _ string, _ int) ([]*model.User, error) {
			return nil, nil
		},
	}
	svc := NewTeamService(&mockTeamRepo{}, userRepo, &mockMainItemRepo{}, nil, &mockDB{})

	result, err := svc.SearchAvailableUsers(context.Background(), 1, "")
	require.NoError(t, err)
	assert.NotNil(t, result)
	// nil slice from repo should be converted to empty slice
	assert.Empty(t, result)
}

// ---------------------------------------------------------------------------
// Tests: GetTeamDetail
// ---------------------------------------------------------------------------

func TestGetTeamDetail_UsesCountMembers(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{
			BaseModel: model.BaseModel{ID: 1},
			TeamName: "Alpha",
			PmKey:      10,
		},
		members: []*dto.TeamMemberDTO{
			{TeamKey: "1", UserBizKey: "10", DisplayName: "Alice", },
			{TeamKey: "1", UserBizKey: "20", DisplayName: "Bob", },
			{TeamKey: "1", UserBizKey: "30", DisplayName: "Charlie", },
		},
		countMembersVal: 3,
	}
	userRepo := &mockTeamUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 10}, DisplayName: "Alice PM"},
	}
	mainItemRepo := &mockMainItemRepo{}
	svc := NewTeamService(teamRepo, userRepo, mainItemRepo, nil, &mockDB{})

	detail, err := svc.GetTeamDetail(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 3, detail.MemberCount)
	assert.True(t, teamRepo.countMembersCalled, "should call CountMembers instead of ListMembers")
	assert.False(t, teamRepo.listMembersCalled, "should not call ListMembers for member count")
}

func TestGetTeamDetail_CountMembersFallback(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{
			BaseModel: model.BaseModel{ID: 1},
			TeamName: "Alpha",
			PmKey:      10,
		},
		countMembersErr: errors.New("not supported"),
		members: []*dto.TeamMemberDTO{
			{TeamKey: "1", UserBizKey: "10", DisplayName: "Alice", },
		},
	}
	userRepo := &mockTeamUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 10}, DisplayName: "Alice PM"},
	}
	mainItemRepo := &mockMainItemRepo{}
	svc := NewTeamService(teamRepo, userRepo, mainItemRepo, nil, &mockDB{})

	detail, err := svc.GetTeamDetail(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, detail.MemberCount)
	// Should fall back to ListMembers when CountMembers fails
	assert.True(t, teamRepo.countMembersCalled, "should try CountMembers first")
	assert.True(t, teamRepo.listMembersCalled, "should fall back to ListMembers on error")
}

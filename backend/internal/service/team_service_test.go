package service

import (
	"context"
	"database/sql"
	"errors"
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

func (m *mockTeamRepo) Update(_ context.Context, team *model.Team) error {
	m.updatedTeam = team
	return m.updateErr
}

func (m *mockTeamRepo) Delete(_ context.Context, teamID uint) error {
	m.deletedTeamID = teamID
	return m.deleteErr
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
	var result []*dto.TeamMemberDTO
	for _, mem := range m.members {
		if mem.TeamID == teamID {
			result = append(result, mem)
		}
	}
	return result, nil
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
			result[mem.TeamID] = mem.DisplayName
		}
	}
	return result, nil
}

// mockTeamUserRepo is a separate mock for UserRepo used in team_service tests.
type mockTeamUserRepo struct {
	user *model.User
	err  error
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
	svc := NewTeamService(repo, userRepo, &mockMainItemRepo{}, &mockDB{})

	team, err := svc.CreateTeam(context.Background(), 1, dto.CreateTeamReq{
		Name:        "Alpha Team",
		Description: "A test team",
		Code:        "ALPHA",
	})
	require.NoError(t, err)
	assert.Equal(t, "Alpha Team", team.Name)
	assert.Equal(t, "A test team", team.Description)
	assert.Equal(t, uint(1), team.PmID)

	// Verify creator was added as member with role "pm"
	assert.NotNil(t, repo.createdMember)
	assert.Equal(t, uint(1), repo.createdMember.UserID)
	assert.Equal(t, "pm", repo.createdMember.Role)
}

func TestCreateTeam_CodeFieldPersisted(t *testing.T) {
	repo := &mockTeamRepo{}
	svc := NewTeamService(repo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

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
	svc := NewTeamService(repo, userRepo, &mockMainItemRepo{}, &mockDB{})

	_, err := svc.CreateTeam(context.Background(), 1, dto.CreateTeamReq{Name: "Team"})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: GetTeam
// ---------------------------------------------------------------------------

func TestGetTeam_Success(t *testing.T) {
	repo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, Name: "Alpha", PmID: 10},
	}
	svc := NewTeamService(repo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	team, err := svc.GetTeam(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), team.ID)
	assert.Equal(t, "Alpha", team.Name)
}

func TestGetTeam_NotFound(t *testing.T) {
	repo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(repo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	_, err := svc.GetTeam(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

// ---------------------------------------------------------------------------
// Tests: ListTeams
// ---------------------------------------------------------------------------

func TestListTeams_Success(t *testing.T) {
	repo := &mockTeamRepo{
		teams: []*model.Team{
			{BaseModel: model.BaseModel{ID: 1}, Name: "Team A", PmID: 10},
			{BaseModel: model.BaseModel{ID: 2}, Name: "Team B", PmID: 20},
		},
		members: []*dto.TeamMemberDTO{
			{TeamID: 1, UserID: 10, Role: "pm", DisplayName: "Alice"},
			{TeamID: 2, UserID: 20, Role: "pm", DisplayName: "Bob"},
		},
	}
	svc := NewTeamService(repo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	teams, err := svc.ListTeams(context.Background(), 1, false)
	require.NoError(t, err)
	require.Len(t, teams, 2)
	assert.Equal(t, "Alice", teams[0].PmDisplayName)
	assert.Equal(t, "Bob", teams[1].PmDisplayName)
}

// ---------------------------------------------------------------------------
// Tests: InviteMember
// ---------------------------------------------------------------------------

func TestInviteMember_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team:          &model.Team{BaseModel: model.BaseModel{ID: 1}, Name: "Alpha", PmID: 10},
		member:        nil, // invited user is not yet a member
		findMemberErr: apperrors.ErrNotFound, // FindMember returns not found
	}
	userRepo := &mockTeamUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 5}, Username: "bob"},
	}
	svc := NewTeamService(teamRepo, userRepo, &mockMainItemRepo{}, &mockDB{})

	err := svc.InviteMember(context.Background(), 10, 1, dto.InviteMemberReq{Username: "bob"})
	require.NoError(t, err)

	// Verify member was added
	assert.NotNil(t, teamRepo.createdMember)
	assert.Equal(t, uint(5), teamRepo.createdMember.UserID)
	assert.Equal(t, uint(1), teamRepo.createdMember.TeamID)
}

func TestInviteMember_TeamNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.InviteMember(context.Background(), 10, 999, dto.InviteMemberReq{Username: "bob"})
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestInviteMember_UserNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmID: 10},
	}
	userRepo := &mockTeamUserRepo{err: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, userRepo, &mockMainItemRepo{}, &mockDB{})

	err := svc.InviteMember(context.Background(), 10, 1, dto.InviteMemberReq{Username: "ghost"})
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestInviteMember_AlreadyMember(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmID: 10},
		member: &model.TeamMember{TeamID: 1, UserID: 5, Role: "member"},
	}
	userRepo := &mockTeamUserRepo{
		user: &model.User{BaseModel: model.BaseModel{ID: 5}, Username: "bob"},
	}
	svc := NewTeamService(teamRepo, userRepo, &mockMainItemRepo{}, &mockDB{})

	err := svc.InviteMember(context.Background(), 10, 1, dto.InviteMemberReq{Username: "bob"})
	assert.ErrorIs(t, err, apperrors.ErrAlreadyMember)
}

// TestInviteMember_CallerNotPM is removed: PM check moved to RequirePermission middleware.

// ---------------------------------------------------------------------------
// Tests: RemoveMember
// ---------------------------------------------------------------------------

func TestRemoveMember_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmID: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.RemoveMember(context.Background(), 10, 1, 5)
	require.NoError(t, err)
	assert.Equal(t, uint(1), teamRepo.removedMember.teamID)
	assert.Equal(t, uint(5), teamRepo.removedMember.userID)
}

func TestRemoveMember_TeamNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.RemoveMember(context.Background(), 10, 999, 5)
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestRemoveMember_CannotRemoveSelf(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmID: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.RemoveMember(context.Background(), 10, 1, 10)
	assert.ErrorIs(t, err, apperrors.ErrCannotRemoveSelf)
}

func TestRemoveMember_CallerNotPM(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmID: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.RemoveMember(context.Background(), 99, 1, 5)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

// ---------------------------------------------------------------------------
// Tests: TransferPM
// ---------------------------------------------------------------------------

func TestTransferPM_Success(t *testing.T) {
	db := &mockDB{}
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, Name: "Alpha", PmID: 10},
		findMembersByUserID: map[uint]*model.TeamMember{
			20: {ID: 2, TeamID: 1, UserID: 20, Role: "member"},
			10: {ID: 1, TeamID: 1, UserID: 10, Role: "pm"},
		},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, db)

	err := svc.TransferPM(context.Background(), 10, 1, 20)
	require.NoError(t, err)
	assert.True(t, db.called)

	// Verify team PM was updated
	assert.NotNil(t, teamRepo.updatedTeam)
	assert.Equal(t, uint(20), teamRepo.updatedTeam.PmID)

	// Verify two member role updates: new PM -> "pm", old PM -> "member"
	require.Len(t, teamRepo.updatedMembers, 2)
	assert.Equal(t, "pm", teamRepo.updatedMembers[0].Role)
	assert.Equal(t, uint(20), teamRepo.updatedMembers[0].UserID)
	assert.Equal(t, "member", teamRepo.updatedMembers[1].Role)
	assert.Equal(t, uint(10), teamRepo.updatedMembers[1].UserID)
}

func TestTransferPM_TeamNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.TransferPM(context.Background(), 10, 999, 20)
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestTransferPM_CallerNotPM(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmID: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.TransferPM(context.Background(), 99, 1, 20)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestTransferPM_TargetNotMember(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team:          &model.Team{BaseModel: model.BaseModel{ID: 1}, PmID: 10},
		member:        nil,
		findMemberErr: gorm.ErrRecordNotFound,
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.TransferPM(context.Background(), 10, 1, 99)
	assert.ErrorIs(t, err, apperrors.ErrNotTeamMember)
}

// ---------------------------------------------------------------------------
// Tests: DisbandTeam
// ---------------------------------------------------------------------------

func TestDisbandTeam_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, Name: "Alpha Team", PmID: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.DisbandTeam(context.Background(), 10, 1, "Alpha Team")
	require.NoError(t, err)
	assert.Equal(t, uint(1), teamRepo.deletedTeamID)
}

func TestDisbandTeam_TeamNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.DisbandTeam(context.Background(), 10, 999, "anything")
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestDisbandTeam_CallerNotPM(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, Name: "Alpha", PmID: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.DisbandTeam(context.Background(), 99, 1, "Alpha")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestDisbandTeam_ConfirmNameMismatch(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, Name: "Alpha Team", PmID: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

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
			{TeamID: 1, UserID: 10, DisplayName: "Alice", Username: "alice", Role: "pm", JoinedAt: now.Format(time.RFC3339)},
			{TeamID: 1, UserID: 5, DisplayName: "Bob", Username: "bob", Role: "member", JoinedAt: now.Format(time.RFC3339)},
		},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	members, err := svc.ListMembers(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Equal(t, "Alice", members[0].DisplayName)
	assert.Equal(t, "pm", members[0].Role)
	assert.Equal(t, "Bob", members[1].DisplayName)
}

func TestListMembers_Empty(t *testing.T) {
	teamRepo := &mockTeamRepo{members: []*dto.TeamMemberDTO{}}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	members, err := svc.ListMembers(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, members)
}

// ---------------------------------------------------------------------------
// Tests: UpdateMemberRole
// ---------------------------------------------------------------------------

func TestUpdateMemberRole_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team:   &model.Team{BaseModel: model.BaseModel{ID: 1}, PmID: 10},
		member: &model.TeamMember{ID: 1, TeamID: 1, UserID: 5, Role: "member"},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.UpdateMemberRole(context.Background(), 10, 1, 5, "leader")
	require.NoError(t, err)
	require.Len(t, teamRepo.updatedMembers, 1)
	assert.Equal(t, "leader", teamRepo.updatedMembers[0].Role)
}

func TestUpdateMemberRole_TeamNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{findByIDErr: gorm.ErrRecordNotFound}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.UpdateMemberRole(context.Background(), 10, 999, 5, "leader")
	assert.ErrorIs(t, err, apperrors.ErrTeamNotFound)
}

func TestUpdateMemberRole_CallerNotPM(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team: &model.Team{BaseModel: model.BaseModel{ID: 1}, PmID: 10},
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.UpdateMemberRole(context.Background(), 99, 1, 5, "leader")
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestUpdateMemberRole_TargetNotMember(t *testing.T) {
	teamRepo := &mockTeamRepo{
		team:          &model.Team{BaseModel: model.BaseModel{ID: 1}, PmID: 10},
		member:        nil,
		findMemberErr: gorm.ErrRecordNotFound,
	}
	svc := NewTeamService(teamRepo, &mockTeamUserRepo{}, &mockMainItemRepo{}, &mockDB{})

	err := svc.UpdateMemberRole(context.Background(), 10, 1, 99, "leader")
	assert.ErrorIs(t, err, apperrors.ErrNotTeamMember)
}

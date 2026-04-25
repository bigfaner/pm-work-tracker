package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/permissions"
	"pm-work-tracker/backend/internal/repository"
)

// ---------------------------------------------------------------------------
// Mock repos for RoleService tests
// ---------------------------------------------------------------------------

type mockRoleRepo struct {
	roles       []model.Role
	roleByID    *model.Role
	roleByName  *model.Role
	perms       []string
	memberCount int64
	hasPerm     bool
	created     *model.Role
	updated     *model.Role
	deletedID   *uint
	setPermID   *uint
	setPermCodes []string

	// For GetUserPermissions
	userTeamPerms map[uint][]string

	// Error controls
	listErr      error
	findByIDErr  error
	findByNameErr error
	createErr    error
	updateErr    error
	deleteErr    error
	listPermErr  error
	setPermErr   error
	countErr     error
	hasPermErr   error
	userTeamPermErr error
}

func (m *mockRoleRepo) List(_ context.Context) ([]model.Role, error) {
	return m.roles, m.listErr
}

func (m *mockRoleRepo) FindByID(_ context.Context, id uint) (*model.Role, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if m.roleByID != nil {
		return m.roleByID, nil
	}
	// search in roles list
	for i := range m.roles {
		if m.roles[i].ID == id {
			return &m.roles[i], nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockRoleRepo) FindByName(_ context.Context, name string) (*model.Role, error) {
	if m.findByNameErr != nil {
		return nil, m.findByNameErr
	}
	if m.roleByName != nil {
		return m.roleByName, nil
	}
	for i := range m.roles {
		if m.roles[i].Name == name {
			return &m.roles[i], nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockRoleRepo) Create(_ context.Context, role *model.Role) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.created = role
	role.ID = uint(len(m.roles) + 100)
	return nil
}

func (m *mockRoleRepo) Update(_ context.Context, role *model.Role) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updated = role
	return nil
}

func (m *mockRoleRepo) Delete(_ context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedID = &id
	return nil
}

func (m *mockRoleRepo) ListPermissions(_ context.Context, roleID uint) ([]string, error) {
	return m.perms, m.listPermErr
}

func (m *mockRoleRepo) SetPermissions(_ context.Context, roleID uint, codes []string) error {
	if m.setPermErr != nil {
		return m.setPermErr
	}
	m.setPermID = &roleID
	m.setPermCodes = codes
	return nil
}

func (m *mockRoleRepo) CountMembersByRoleID(_ context.Context, roleID uint) (int64, error) {
	return m.memberCount, m.countErr
}

func (m *mockRoleRepo) HasPermission(_ context.Context, userID uint, code string) (bool, error) {
	return m.hasPerm, m.hasPermErr
}

func (m *mockRoleRepo) GetUserTeamPermissions(_ context.Context, userID uint) (map[uint][]string, error) {
	if m.userTeamPermErr != nil {
		return nil, m.userTeamPermErr
	}
	if m.userTeamPerms != nil {
		return m.userTeamPerms, nil
	}
	return map[uint][]string{}, nil
}

// mockRoleUserRepo implements repository.UserRepo for role service tests.
type mockRoleUserRepo struct {
	user    *model.User
	findErr error
}

func (m *mockRoleUserRepo) FindByID(_ context.Context, _ uint) (*model.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.user, nil
}

func (m *mockRoleUserRepo) FindByUsername(_ context.Context, _ string) (*model.User, error) {
	return nil, nil
}

func (m *mockRoleUserRepo) List(_ context.Context) ([]*model.User, error) {
	return nil, nil
}

func (m *mockRoleUserRepo) Create(_ context.Context, _ *model.User) error {
	return nil
}

func (m *mockRoleUserRepo) Update(_ context.Context, _ *model.User) error {
	return nil
}
func (m *mockRoleUserRepo) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.User, error) {
	return nil, nil
}
func (m *mockRoleUserRepo) FindByBizKey(_ context.Context, _ int64) (*model.User, error) {
	return nil, nil
}
func (m *mockRoleUserRepo) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
	return nil, 0, nil
}
func (m *mockRoleUserRepo) SearchAvailable(_ context.Context, _ uint, _ string, _ int) ([]*model.User, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helper: build the service under test
// ---------------------------------------------------------------------------

func newTestRoleService(repo repository.RoleRepo, userRepo repository.UserRepo) RoleService {
	return NewRoleService(repo, userRepo)
}

// ---------------------------------------------------------------------------
// Tests: ListRoles
// ---------------------------------------------------------------------------

func TestRoleService_ListRoles_ReturnsAll(t *testing.T) {
	r1 := model.Role{BaseModel: model.BaseModel{ID: 1}, Name: "superadmin", IsPreset: true}
	r2 := model.Role{BaseModel: model.BaseModel{ID: 2}, Name: "pm", IsPreset: true}
	repo := &mockRoleRepo{roles: []model.Role{r1, r2}, perms: []string{"team:read"}, memberCount: 3}
	svc := newTestRoleService(repo, nil)

	result, err := svc.ListRoles(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "superadmin", result[0].Name)
	assert.Equal(t, "pm", result[1].Name)
}

func TestRoleService_ListRoles_Empty(t *testing.T) {
	repo := &mockRoleRepo{roles: []model.Role{}}
	svc := newTestRoleService(repo, nil)

	result, err := svc.ListRoles(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRoleService_ListRoles_RepoError(t *testing.T) {
	repo := &mockRoleRepo{listErr: assert.AnError}
	svc := newTestRoleService(repo, nil)

	_, err := svc.ListRoles(context.Background())
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Tests: GetRole
// ---------------------------------------------------------------------------

func TestRoleService_GetRole_Found(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 2}, Name: "pm", Description: "PM role", IsPreset: true}
	repo := &mockRoleRepo{roleByID: role, perms: []string{"team:create", "team:read"}, memberCount: 5}
	svc := newTestRoleService(repo, nil)

	result, err := svc.GetRole(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, uint(2), result.ID)
	assert.Equal(t, "pm", result.Name)
	assert.Equal(t, int64(5), result.MemberCount)
}

func TestRoleService_GetRole_NotFound(t *testing.T) {
	repo := &mockRoleRepo{findByIDErr: apperrors.ErrNotFound}
	svc := newTestRoleService(repo, nil)

	_, err := svc.GetRole(context.Background(), 999)
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

// ---------------------------------------------------------------------------
// Tests: CreateRole
// ---------------------------------------------------------------------------

func TestRoleService_CreateRole_Success(t *testing.T) {
	repo := &mockRoleRepo{perms: []string{}, memberCount: 0}
	svc := newTestRoleService(repo, nil)

	result, err := svc.CreateRole(context.Background(), dto.CreateRoleReq{
		Name:            "viewer",
		Description: "read only",
		PermissionCodes: []string{"team:read", "main_item:read"},
	})
	require.NoError(t, err)
	assert.Equal(t, "viewer", result.Name)
	assert.Equal(t, "read only", result.Description)
	assert.False(t, result.IsPreset)
	// Verify SetPermissions was called
	assert.NotNil(t, repo.setPermID)
	assert.Equal(t, []string{"team:read", "main_item:read"}, repo.setPermCodes)
}

func TestRoleService_CreateRole_NameTooShort(t *testing.T) {
	svc := newTestRoleService(&mockRoleRepo{}, nil)

	_, err := svc.CreateRole(context.Background(), dto.CreateRoleReq{
		Name:            "a",
		PermissionCodes: []string{"team:read"},
	})
	assert.ErrorIs(t, err, ErrValidation)
}

func TestRoleService_CreateRole_NameTooLong(t *testing.T) {
	svc := newTestRoleService(&mockRoleRepo{}, nil)

	longName := ""
	for i := 0; i < 51; i++ {
		longName += "x"
	}

	_, err := svc.CreateRole(context.Background(), dto.CreateRoleReq{
		Name:            longName,
		PermissionCodes: []string{"team:read"},
	})
	assert.ErrorIs(t, err, ErrValidation)
}

func TestRoleService_CreateRole_DuplicateName(t *testing.T) {
	existing := model.Role{BaseModel: model.BaseModel{ID: 1}, Name: "pm"}
	repo := &mockRoleRepo{roles: []model.Role{existing}}
	svc := newTestRoleService(repo, nil)

	_, err := svc.CreateRole(context.Background(), dto.CreateRoleReq{
		Name:            "pm",
		PermissionCodes: []string{"team:read"},
	})
	assert.ErrorIs(t, err, ErrRoleNameExists)
}

func TestRoleService_CreateRole_NoPermissionCodes(t *testing.T) {
	svc := newTestRoleService(&mockRoleRepo{}, nil)

	_, err := svc.CreateRole(context.Background(), dto.CreateRoleReq{
		Name:            "viewer",
		PermissionCodes: []string{},
	})
	assert.ErrorIs(t, err, ErrValidation)
}

func TestRoleService_CreateRole_InvalidPermissionCode(t *testing.T) {
	svc := newTestRoleService(&mockRoleRepo{}, nil)

	_, err := svc.CreateRole(context.Background(), dto.CreateRoleReq{
		Name:            "viewer",
		PermissionCodes: []string{"team:read", "invalid:code"},
	})
	assert.ErrorIs(t, err, ErrInvalidPermissionCode)
}

func TestRoleService_CreateRole_BoundaryNameLength2(t *testing.T) {
	repo := &mockRoleRepo{}
	svc := newTestRoleService(repo, nil)

	result, err := svc.CreateRole(context.Background(), dto.CreateRoleReq{
		Name:            "ab",
		PermissionCodes: []string{"team:read"},
	})
	require.NoError(t, err)
	assert.Equal(t, "ab", result.Name)
}

func TestRoleService_CreateRole_BoundaryNameLength50(t *testing.T) {
	name50 := ""
	for i := 0; i < 50; i++ {
		name50 += "x"
	}
	repo := &mockRoleRepo{}
	svc := newTestRoleService(repo, nil)

	result, err := svc.CreateRole(context.Background(), dto.CreateRoleReq{
		Name:            name50,
		PermissionCodes: []string{"team:read"},
	})
	require.NoError(t, err)
	assert.Equal(t, 50, len(result.Name))
}

// ---------------------------------------------------------------------------
// Tests: UpdateRole
// ---------------------------------------------------------------------------

func TestRoleService_UpdateRole_Success(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 4}, Name: "viewer", Description: "old desc", IsPreset: false}
	repo := &mockRoleRepo{roleByID: role, perms: []string{"team:read", "main_item:read"}, memberCount: 0}
	svc := newTestRoleService(repo, nil)

	newName := "viewer2"
	result, err := svc.UpdateRole(context.Background(), 4, dto.UpdateRoleReq{
		Name:            &newName,
		Description:     strPtr("new desc"),
		PermissionCodes: []string{"team:read", "main_item:read", "report:export"},
	})
	require.NoError(t, err)
	assert.Equal(t, "viewer2", result.Name)
	assert.NotNil(t, repo.updated)
	assert.Equal(t, []string{"team:read", "main_item:read", "report:export"}, repo.setPermCodes)
}

func TestRoleService_UpdateRole_NotFound(t *testing.T) {
	repo := &mockRoleRepo{findByIDErr: apperrors.ErrNotFound}
	svc := newTestRoleService(repo, nil)

	name := "x"
	_, err := svc.UpdateRole(context.Background(), 999, dto.UpdateRoleReq{Name: &name})
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

func TestRoleService_UpdateRole_SuperadminImmutable(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 1}, Name: "superadmin", IsPreset: true}
	repo := &mockRoleRepo{roleByID: role}
	svc := newTestRoleService(repo, nil)

	desc := "new desc"
	_, err := svc.UpdateRole(context.Background(), 1, dto.UpdateRoleReq{Description: &desc})
	assert.ErrorIs(t, err, ErrPresetRoleImmutable)
}

func TestRoleService_UpdateRole_PresetRoleCanChangeDescAndPerms(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 2}, Name: "pm", Description: "old", IsPreset: true}
	repo := &mockRoleRepo{roleByID: role, perms: []string{"team:read"}, memberCount: 3}
	svc := newTestRoleService(repo, nil)

	result, err := svc.UpdateRole(context.Background(), 2, dto.UpdateRoleReq{
		Description:     strPtr("new desc"),
		PermissionCodes: []string{"team:create", "team:read"},
	})
	require.NoError(t, err)
	assert.Equal(t, "new desc", result.Description)
	assert.Equal(t, "pm", result.Name) // name unchanged
}

func TestRoleService_UpdateRole_PresetRoleCannotChangeName(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 2}, Name: "pm", IsPreset: true}
	repo := &mockRoleRepo{roleByID: role}
	svc := newTestRoleService(repo, nil)

	newName := "pm-v2"
	_, err := svc.UpdateRole(context.Background(), 2, dto.UpdateRoleReq{Name: &newName})
	assert.ErrorIs(t, err, ErrPresetRoleImmutable)
}

func TestRoleService_UpdateRole_DuplicateName(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 4}, Name: "viewer", IsPreset: false}
	existing := model.Role{BaseModel: model.BaseModel{ID: 5}, Name: "editor"}
	repo := &mockRoleRepo{roleByID: role, roles: []model.Role{*role, existing}}
	svc := newTestRoleService(repo, nil)

	newName := "editor"
	_, err := svc.UpdateRole(context.Background(), 4, dto.UpdateRoleReq{Name: &newName})
	assert.ErrorIs(t, err, ErrRoleNameExists)
}

func TestRoleService_UpdateRole_InvalidPermissionCode(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 4}, Name: "viewer", IsPreset: false}
	repo := &mockRoleRepo{roleByID: role}
	svc := newTestRoleService(repo, nil)

	_, err := svc.UpdateRole(context.Background(), 4, dto.UpdateRoleReq{
		PermissionCodes: []string{"invalid:code"},
	})
	assert.ErrorIs(t, err, ErrInvalidPermissionCode)
}

func TestRoleService_UpdateRole_EmptyPermissionCodes(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 4}, Name: "viewer", IsPreset: false}
	repo := &mockRoleRepo{roleByID: role}
	svc := newTestRoleService(repo, nil)

	_, err := svc.UpdateRole(context.Background(), 4, dto.UpdateRoleReq{
		PermissionCodes: []string{},
	})
	assert.ErrorIs(t, err, ErrValidation)
}

func TestRoleService_UpdateRole_NameValidation(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 4}, Name: "viewer", IsPreset: false}
	repo := &mockRoleRepo{roleByID: role}
	svc := newTestRoleService(repo, nil)

	shortName := "a"
	_, err := svc.UpdateRole(context.Background(), 4, dto.UpdateRoleReq{Name: &shortName})
	assert.ErrorIs(t, err, ErrValidation)
}

// ---------------------------------------------------------------------------
// Tests: DeleteRole
// ---------------------------------------------------------------------------

func TestRoleService_DeleteRole_Success(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 4}, Name: "viewer", IsPreset: false}
	repo := &mockRoleRepo{roleByID: role, memberCount: 0}
	svc := newTestRoleService(repo, nil)

	err := svc.DeleteRole(context.Background(), 4)
	require.NoError(t, err)
	assert.NotNil(t, repo.deletedID)
	assert.Equal(t, uint(4), *repo.deletedID)
}

func TestRoleService_DeleteRole_NotFound(t *testing.T) {
	repo := &mockRoleRepo{findByIDErr: apperrors.ErrNotFound}
	svc := newTestRoleService(repo, nil)

	err := svc.DeleteRole(context.Background(), 999)
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

func TestRoleService_DeleteRole_PresetRole(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 1}, Name: "superadmin", IsPreset: true}
	repo := &mockRoleRepo{roleByID: role, memberCount: 0}
	svc := newTestRoleService(repo, nil)

	err := svc.DeleteRole(context.Background(), 1)
	assert.ErrorIs(t, err, ErrPresetRoleImmutable)
}

func TestRoleService_DeleteRole_InUse(t *testing.T) {
	role := &model.Role{BaseModel: model.BaseModel{ID: 4}, Name: "viewer", IsPreset: false}
	repo := &mockRoleRepo{roleByID: role, memberCount: 5}
	svc := newTestRoleService(repo, nil)

	err := svc.DeleteRole(context.Background(), 4)
	assert.ErrorIs(t, err, ErrRoleInUse)
}

// ---------------------------------------------------------------------------
// Tests: ListPermissionCodes
// ---------------------------------------------------------------------------

func TestRoleService_ListPermissionCodes(t *testing.T) {
	svc := newTestRoleService(&mockRoleRepo{}, nil)

	result := svc.ListPermissionCodes(context.Background())
	assert.Equal(t, permissions.Registry, result)
}

// ---------------------------------------------------------------------------
// Tests: GetUserPermissions
// ---------------------------------------------------------------------------

func TestRoleService_GetUserPermissions_SuperAdmin(t *testing.T) {
	user := &model.User{BaseModel: model.BaseModel{ID: 1}, Username: "admin", IsSuperAdmin: true}
	repo := &mockRoleRepo{userTeamPerms: map[uint][]string{1: {"team:read"}}}
	userRepo := &mockRoleUserRepo{user: user}
	svc := newTestRoleService(repo, userRepo)

	result, err := svc.GetUserPermissions(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, result.IsSuperAdmin)
	assert.Equal(t, map[uint][]string{1: {"team:read"}}, result.TeamPermissions)
}

func TestRoleService_GetUserPermissions_NormalUser(t *testing.T) {
	user := &model.User{BaseModel: model.BaseModel{ID: 2}, Username: "user1", IsSuperAdmin: false}
	teamPerms := map[uint][]string{
		10: {"team:read", "team:update"},
		20: {"team:read", "main_item:create"},
	}
	repo := &mockRoleRepo{userTeamPerms: teamPerms}
	userRepo := &mockRoleUserRepo{user: user}
	svc := newTestRoleService(repo, userRepo)

	result, err := svc.GetUserPermissions(context.Background(), 2)
	require.NoError(t, err)
	assert.False(t, result.IsSuperAdmin)
	assert.Equal(t, teamPerms, result.TeamPermissions)
}

func TestRoleService_GetUserPermissions_UserNotFound(t *testing.T) {
	repo := &mockRoleRepo{}
	userRepo := &mockRoleUserRepo{findErr: apperrors.ErrNotFound}
	svc := newTestRoleService(repo, userRepo)

	_, err := svc.GetUserPermissions(context.Background(), 999)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Tests: ListRoles with permission_count and member_count
// ---------------------------------------------------------------------------

func TestRoleService_ListRoles_WithCounts(t *testing.T) {
	r1 := model.Role{BaseModel: model.BaseModel{ID: 1}, Name: "superadmin", IsPreset: true}
	r2 := model.Role{BaseModel: model.BaseModel{ID: 2}, Name: "pm", IsPreset: true}
	repo := &mockRoleRepo{
		roles:       []model.Role{r1, r2},
		perms:       []string{"team:read", "team:update"},
		memberCount: 7,
	}
	svc := newTestRoleService(repo, nil)

	result, err := svc.ListRoles(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 2)

	// superadmin gets TotalCodeCount(), pm gets len(perms)=2
	assert.Equal(t, permissions.TotalCodeCount(), result[0].PermissionCount)
	assert.Equal(t, int64(7), result[0].MemberCount)

	assert.Equal(t, 2, result[1].PermissionCount)
	assert.Equal(t, int64(7), result[1].MemberCount)
}

// strPtr is defined in admin_service_test.go

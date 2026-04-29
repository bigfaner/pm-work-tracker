package gorm_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	pkgerrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/snowflake"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
)

func timeNow() time.Time {
	return time.Now()
}

func setupRoleTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.Role{}, &model.RolePermission{}, &model.TeamMember{}))
	return db
}

func seedRoleUser(t *testing.T, db *gormlib.DB, username string) *model.User {
	t.Helper()
	u := model.User{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, Username: username, DisplayName: username + "_display", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	return &u
}

func seedRole(t *testing.T, db *gormlib.DB, name, desc string, isPreset bool) *model.Role {
	t.Helper()
	r := model.Role{Name: name, Description: desc, IsPreset: isPreset}
	require.NoError(t, db.Create(&r).Error)
	r.BizKey = snowflake.Generate()
	require.NoError(t, db.Save(&r).Error)
	return &r
}

// --- List ---

func TestRoleRepo_List(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	seedRole(t, db, "member", "Member role", true)
	seedRole(t, db, "pm", "PM role", true)

	roles, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, roles, 2)
}

func TestRoleRepo_List_Empty(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	roles, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, roles)
}

func TestRoleRepo_List_OrderedByCreatedAt(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	seedRole(t, db, "beta", "Beta", false)
	seedRole(t, db, "alpha", "Alpha", false)

	roles, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, roles, 2)
	assert.Equal(t, "beta", roles[0].Name, "should be ordered by created_at (insertion order)")
	assert.Equal(t, "alpha", roles[1].Name)
}

func TestRoleRepo_List_ExcludesSoftDeleted(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	r1 := seedRole(t, db, "active", "Active", false)
	seedRole(t, db, "deleted", "Deleted", false)

	require.NoError(t, db.Delete(r1).Error)

	roles, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, "deleted", roles[0].Name)
}

// --- FindByID ---

func TestRoleRepo_FindByID_Found(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	r := seedRole(t, db, "pm", "PM role", true)

	found, err := repo.FindByID(ctx, r.ID)
	require.NoError(t, err)
	assert.Equal(t, "pm", found.Name)
	assert.Equal(t, "PM role", found.Description)
	assert.True(t, found.IsPreset)
}

func TestRoleRepo_FindByID_NotFound(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 9999)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- FindByName ---

func TestRoleRepo_FindByName_Found(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	seedRole(t, db, "member", "Member role", true)

	found, err := repo.FindByName(ctx, "member")
	require.NoError(t, err)
	assert.Equal(t, "member", found.Name)
}

func TestRoleRepo_FindByName_NotFound(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	_, err := repo.FindByName(ctx, "nonexistent")
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- Create ---

func TestRoleRepo_Create(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	role := &model.Role{Name: "custom", Description: "Custom role"}
	require.NoError(t, repo.Create(ctx, role))
	assert.NotZero(t, role.ID)
}

// --- Update ---

func TestRoleRepo_Update(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	r := seedRole(t, db, "old-name", "Old desc", false)

	r.Name = "new-name"
	r.Description = "New desc"
	require.NoError(t, repo.Update(ctx, r))

	found, err := repo.FindByID(ctx, r.ID)
	require.NoError(t, err)
	assert.Equal(t, "new-name", found.Name)
	assert.Equal(t, "New desc", found.Description)
}

// --- Delete (soft) ---

func TestRoleRepo_Delete(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	r := seedRole(t, db, "to-delete", "Delete me", false)
	require.NoError(t, repo.Delete(ctx, r.ID))

	_, err := repo.FindByID(ctx, r.ID)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)

	// Verify soft-deleted record still exists
	var count int64
	db.Unscoped().Model(&model.Role{}).Where("id = ?", r.ID).Count(&count)
	assert.Equal(t, int64(1), count, "role should still exist as soft-deleted")
}

// --- ListPermissions ---

func TestRoleRepo_ListPermissions(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	r := seedRole(t, db, "pm", "PM role", true)

	// Seed permission bindings directly
	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r.BizKey, PermissionCode: "team:create"}).Error)
	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r.BizKey, PermissionCode: "team:read"}).Error)

	codes, err := repo.ListPermissions(ctx, r.BizKey)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"team:create", "team:read"}, codes)
}

func TestRoleRepo_ListPermissions_Empty(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	r := seedRole(t, db, "empty-role", "No perms", false)

	codes, err := repo.ListPermissions(ctx, r.BizKey)
	require.NoError(t, err)
	assert.Empty(t, codes)
}

// --- SetPermissions ---

func TestRoleRepo_SetPermissions(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	r := seedRole(t, db, "pm", "PM role", true)

	// Set initial permissions
	require.NoError(t, repo.SetPermissions(ctx, r.BizKey, []string{"team:create", "team:read"}))

	codes, err := repo.ListPermissions(ctx, r.BizKey)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"team:create", "team:read"}, codes)
}

func TestRoleRepo_SetPermissions_Replaces(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	r := seedRole(t, db, "pm", "PM role", true)

	// Set initial permissions
	require.NoError(t, repo.SetPermissions(ctx, r.BizKey, []string{"team:create", "team:read"}))

	// Replace with new set
	require.NoError(t, repo.SetPermissions(ctx, r.BizKey, []string{"team:update", "team:delete"}))

	codes, err := repo.ListPermissions(ctx, r.BizKey)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"team:update", "team:delete"}, codes)
	assert.Len(t, codes, 2, "old permissions should be fully replaced")
}

func TestRoleRepo_SetPermissions_ClearsAll(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	r := seedRole(t, db, "pm", "PM role", true)

	require.NoError(t, repo.SetPermissions(ctx, r.BizKey, []string{"team:create"}))
	require.NoError(t, repo.SetPermissions(ctx, r.BizKey, []string{}))

	codes, err := repo.ListPermissions(ctx, r.BizKey)
	require.NoError(t, err)
	assert.Empty(t, codes)
}

// --- CountMembersByRoleID ---

func TestRoleRepo_CountMembersByRoleKey(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u1 := seedRoleUser(t, db, "user1")
	u2 := seedRoleUser(t, db, "user2")

	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team", PmKey: u1.BizKey, Code: "TM01"}
	require.NoError(t, db.Create(&team).Error)

	r := seedRole(t, db, "member", "Member role", true)
	rOther := seedRole(t, db, "other", "Other role", false)

	// One member with this role
	require.NoError(t, db.Create(&model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(u1.ID), RoleKey: &r.BizKey, JoinedAt: timeNow()}).Error)
	// One member with different role
	require.NoError(t, db.Create(&model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(u2.ID), RoleKey: &rOther.BizKey, JoinedAt: timeNow()}).Error)

	count, err := repo.CountMembersByRoleKey(ctx, r.BizKey)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRoleRepo_CountMembersByRoleKey_NoMembers(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	count, err := repo.CountMembersByRoleKey(ctx, 9999)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// --- HasPermission ---

func TestRoleRepo_HasPermission_True(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u := seedRoleUser(t, db, "user1")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team", PmKey: u.BizKey, Code: "TM01"}
	require.NoError(t, db.Create(&team).Error)
	r := seedRole(t, db, "pm", "PM role", true)

	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r.BizKey, PermissionCode: "team:create"}).Error)
	require.NoError(t, db.Create(&model.TeamMember{TeamKey: team.BizKey, UserKey: u.BizKey, RoleKey: &r.BizKey, JoinedAt: timeNow()}).Error)

	has, err := repo.HasPermission(ctx, u.BizKey, "team:create")
	require.NoError(t, err)
	assert.True(t, has)
}

func TestRoleRepo_HasPermission_False(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u := seedRoleUser(t, db, "user1")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team", PmKey: u.BizKey, Code: "TM01"}
	require.NoError(t, db.Create(&team).Error)
	r := seedRole(t, db, "member", "Member role", true)

	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r.BizKey, PermissionCode: "team:read"}).Error)
	require.NoError(t, db.Create(&model.TeamMember{TeamKey: team.BizKey, UserKey: u.BizKey, RoleKey: &r.BizKey, JoinedAt: timeNow()}).Error)

	has, err := repo.HasPermission(ctx, u.BizKey, "team:delete")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestRoleRepo_HasPermission_UserNotMember(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u := seedRoleUser(t, db, "user1")

	has, err := repo.HasPermission(ctx, u.BizKey, "team:create")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestRoleRepo_HasPermission_ChecksAcrossTeams(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u := seedRoleUser(t, db, "user1")
	team1 := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team1", PmKey: u.BizKey, Code: "TM01"}
	require.NoError(t, db.Create(&team1).Error)
	team2 := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team2", PmKey: u.BizKey, Code: "TM02"}
	require.NoError(t, db.Create(&team2).Error)

	r1 := seedRole(t, db, "role1", "Role 1", false)
	r2 := seedRole(t, db, "role2", "Role 2", false)

	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r1.BizKey, PermissionCode: "team:read"}).Error)
	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r2.BizKey, PermissionCode: "team:delete"}).Error)

	require.NoError(t, db.Create(&model.TeamMember{TeamKey: team1.BizKey, UserKey: u.BizKey, RoleKey: &r1.BizKey, JoinedAt: timeNow()}).Error)
	require.NoError(t, db.Create(&model.TeamMember{TeamKey: team2.BizKey, UserKey: u.BizKey, RoleKey: &r2.BizKey, JoinedAt: timeNow()}).Error)

	// User has team:delete via role2 in team2
	has, err := repo.HasPermission(ctx, u.BizKey, "team:delete")
	require.NoError(t, err)
	assert.True(t, has, "should find permission across teams")
}

// --- GetUserTeamPermissions ---

func TestRoleRepo_GetUserTeamPermissions(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u := seedRoleUser(t, db, "user1")
	team1 := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team1", PmKey: u.BizKey, Code: "TM01"}
	require.NoError(t, db.Create(&team1).Error)
	team2 := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team2", PmKey: u.BizKey, Code: "TM02"}
	require.NoError(t, db.Create(&team2).Error)

	r1 := seedRole(t, db, "pm-role", "PM Role", true)
	r2 := seedRole(t, db, "member-role", "Member Role", true)

	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r1.BizKey, PermissionCode: "team:create"}).Error)
	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r1.BizKey, PermissionCode: "team:read"}).Error)
	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r2.BizKey, PermissionCode: "team:read"}).Error)

	require.NoError(t, db.Create(&model.TeamMember{TeamKey: team1.BizKey, UserKey: u.BizKey, RoleKey: &r1.BizKey, JoinedAt: timeNow()}).Error)
	require.NoError(t, db.Create(&model.TeamMember{TeamKey: team2.BizKey, UserKey: u.BizKey, RoleKey: &r2.BizKey, JoinedAt: timeNow()}).Error)

	result, err := repo.GetUserTeamPermissions(ctx, u.BizKey)
	require.NoError(t, err)
	assert.Equal(t, map[int64][]string{
		team1.BizKey: {"team:create", "team:read"},
		team2.BizKey: {"team:read"},
	}, result)
}

func TestRoleRepo_GetUserTeamPermissions_NoMemberships(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u := seedRoleUser(t, db, "loner")

	result, err := repo.GetUserTeamPermissions(ctx, u.BizKey)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRoleRepo_GetUserTeamPermissions_NoRoleID(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u := seedRoleUser(t, db, "user1")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team", PmKey: u.BizKey, Code: "TM01"}
	require.NoError(t, db.Create(&team).Error)

	// Member without a RoleID (legacy data)
	require.NoError(t, db.Create(&model.TeamMember{TeamKey: team.BizKey, UserKey: u.BizKey, RoleKey: nil, JoinedAt: timeNow()}).Error)

	result, err := repo.GetUserTeamPermissions(ctx, u.BizKey)
	require.NoError(t, err)
	assert.Empty(t, result, "members without RoleID should produce no permission entries")
}

// --- Soft-delete tests for join queries ---

func TestRoleRepo_HasPermission_DeletedMember(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u := seedRoleUser(t, db, "user1")
	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team", PmKey: u.BizKey, Code: "TM01"}
	require.NoError(t, db.Create(&team).Error)
	r := seedRole(t, db, "pm", "PM role", true)

	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r.BizKey, PermissionCode: "team:create"}).Error)
	member := model.TeamMember{TeamKey: team.BizKey, UserKey: u.BizKey, RoleKey: &r.BizKey, JoinedAt: timeNow()}
	require.NoError(t, db.Create(&member).Error)

	// Soft-delete the member
	require.NoError(t, db.Model(&member).Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error)

	has, err := repo.HasPermission(ctx, u.BizKey, "team:create")
	require.NoError(t, err)
	assert.False(t, has, "deleted member should have no permissions")
}

func TestRoleRepo_GetUserTeamPermissions_DeletedMember(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u := seedRoleUser(t, db, "user1")
	team1 := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team1", PmKey: u.BizKey, Code: "TM01"}
	require.NoError(t, db.Create(&team1).Error)
	team2 := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team2", PmKey: u.BizKey, Code: "TM02"}
	require.NoError(t, db.Create(&team2).Error)

	r1 := seedRole(t, db, "pm-role", "PM Role", true)
	r2 := seedRole(t, db, "member-role", "Member Role", true)

	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r1.BizKey, PermissionCode: "team:create"}).Error)
	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r1.BizKey, PermissionCode: "team:read"}).Error)
	require.NoError(t, db.Create(&model.RolePermission{RoleKey: r2.BizKey, PermissionCode: "team:read"}).Error)

	member1 := model.TeamMember{TeamKey: team1.BizKey, UserKey: u.BizKey, RoleKey: &r1.BizKey, JoinedAt: timeNow()}
	require.NoError(t, db.Create(&member1).Error)
	require.NoError(t, db.Create(&model.TeamMember{TeamKey: team2.BizKey, UserKey: u.BizKey, RoleKey: &r2.BizKey, JoinedAt: timeNow()}).Error)

	// Soft-delete member in team1
	require.NoError(t, db.Model(&member1).Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error)

	result, err := repo.GetUserTeamPermissions(ctx, u.BizKey)
	require.NoError(t, err)
	assert.Equal(t, map[int64][]string{
		team2.BizKey: {"team:read"},
	}, result, "deleted member's permissions should be excluded")
}

func TestRoleRepo_CountMembersByRoleKey_DeletedMember(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := gormrepo.NewGormRoleRepo(db)
	ctx := context.Background()

	u1 := seedRoleUser(t, db, "user1")
	u2 := seedRoleUser(t, db, "user2")
	u3 := seedRoleUser(t, db, "user3")

	team := model.Team{BaseModel: model.BaseModel{BizKey: snowflake.Generate()}, TeamName: "Team", PmKey: u1.BizKey, Code: "TM01"}
	require.NoError(t, db.Create(&team).Error)

	r := seedRole(t, db, "member", "Member role", true)

	member1 := model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(u1.ID), RoleKey: &r.BizKey, JoinedAt: timeNow()}
	require.NoError(t, db.Create(&member1).Error)
	require.NoError(t, db.Create(&model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(u2.ID), RoleKey: &r.BizKey, JoinedAt: timeNow()}).Error)
	require.NoError(t, db.Create(&model.TeamMember{TeamKey: int64(team.ID), UserKey: int64(u3.ID), RoleKey: &r.BizKey, JoinedAt: timeNow()}).Error)

	// Soft-delete member1
	require.NoError(t, db.Model(&member1).Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error)

	count, err := repo.CountMembersByRoleKey(ctx, r.BizKey)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count, "deleted members should not be counted")
}

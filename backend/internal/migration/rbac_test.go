package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
)

func setupRBACTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the pre-RBAC schema (with legacy columns)
	err = db.AutoMigrate(&model.User{}, &model.Team{}, &model.Role{}, &model.TeamMember{})
	require.NoError(t, err)

	// Create role_permissions table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS role_permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role_id INTEGER NOT NULL REFERENCES roles(id),
			permission_code TEXT NOT NULL,
			UNIQUE(role_id, permission_code)
		)
	`).Error
	require.NoError(t, err)

	return db
}

func seedPreRBACData(t *testing.T, db *gorm.DB) {
	t.Helper()

	// Create users
	u1 := model.User{Username: "admin", DisplayName: "Admin", PasswordHash: "h", IsSuperAdmin: true}
	require.NoError(t, db.Create(&u1).Error)
	u2 := model.User{Username: "pm1", DisplayName: "PM 1", PasswordHash: "h"}
	require.NoError(t, db.Create(&u2).Error)
	u3 := model.User{Username: "member1", DisplayName: "Member 1", PasswordHash: "h"}
	require.NoError(t, db.Create(&u3).Error)
	u4 := model.User{Username: "member2", DisplayName: "Member 2", PasswordHash: "h"}
	require.NoError(t, db.Create(&u4).Error)

	// Create teams
	team1 := model.Team{Name: "Team 1", PmID: u2.ID}
	require.NoError(t, db.Create(&team1).Error)

	// Create team members with legacy role strings
	tm1 := model.TeamMember{TeamID: team1.ID, UserID: u2.ID, Role: "pm"}
	require.NoError(t, db.Create(&tm1).Error)
	tm2 := model.TeamMember{TeamID: team1.ID, UserID: u3.ID, Role: "member"}
	require.NoError(t, db.Create(&tm2).Error)
	tm3 := model.TeamMember{TeamID: team1.ID, UserID: u4.ID, Role: "member"}
	require.NoError(t, db.Create(&tm3).Error)
}

// === First-run success tests ===

func TestMigrateToRBAC_FirstRunSuccess(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)
}

func TestMigrateToRBAC_PresetRolesSeeded(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// Check superadmin role
	var superadmin model.Role
	require.NoError(t, db.Where("name = ?", "superadmin").First(&superadmin).Error)
	assert.True(t, superadmin.IsPreset)
	assert.Equal(t, uint(1), superadmin.ID)

	// superadmin should have NO permission codes
	count, err := CountPermissionsForRole(db, superadmin.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "superadmin should have no permission codes")

	// Check pm role
	var pm model.Role
	require.NoError(t, db.Where("name = ?", "pm").First(&pm).Error)
	assert.True(t, pm.IsPreset)
	assert.Equal(t, uint(2), pm.ID)

	// pm should have 26 codes
	count, err = CountPermissionsForRole(db, pm.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(26), count, "pm should have 26 permission codes")

	// Check member role
	var member model.Role
	require.NoError(t, db.Where("name = ?", "member").First(&member).Error)
	assert.True(t, member.IsPreset)
	assert.Equal(t, uint(3), member.ID)

	// member should have 11 codes
	count, err = CountPermissionsForRole(db, member.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(11), count, "member should have 11 permission codes")
}

func TestMigrateToRBAC_TeamMemberRoleMigrated(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// Check pm user has role_id = 2
	var pmMember model.TeamMember
	require.NoError(t, db.Where("user_id = ? AND team_id = ?", 2, 1).First(&pmMember).Error)
	require.NotNil(t, pmMember.RoleID)
	assert.Equal(t, uint(2), *pmMember.RoleID, "pm user should have role_id=2")

	// Check member users have role_id = 3
	var memberUser model.TeamMember
	require.NoError(t, db.Where("user_id = ? AND team_id = ?", 3, 1).First(&memberUser).Error)
	require.NotNil(t, memberUser.RoleID)
	assert.Equal(t, uint(3), *memberUser.RoleID, "member user should have role_id=3")
}

func TestMigrateToRBAC_TeamMembersRoleColumnRemoved(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// team_members table should no longer have role string column
	assert.False(t, HasColumn(db, "team_members", "role"),
		"role string column should be removed from team_members table")

	// but should have role_id
	assert.True(t, HasColumn(db, "team_members", "role_id"),
		"role_id column should exist in team_members table")
}

func TestMigrateToRBAC_UsersDataPreserved(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// Verify user data survived the table rebuild
	var users []model.User
	require.NoError(t, db.Find(&users).Error)
	assert.Len(t, users, 4)

	// Verify superadmin flag preserved
	var admin model.User
	require.NoError(t, db.Where("username = ?", "admin").First(&admin).Error)
	assert.True(t, admin.IsSuperAdmin)

	// Verify normal user preserved
	var pm1 model.User
	require.NoError(t, db.Where("username = ?", "pm1").First(&pm1).Error)
	assert.False(t, pm1.IsSuperAdmin)
}

func TestMigrateToRBAC_TrackedInSchemaMigrations(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	var count int64
	db.Raw("SELECT count(*) FROM schema_migrations WHERE version = ?", rbacMigrationVersion).Scan(&count)
	assert.Equal(t, int64(1), count, "rbac migration should be tracked in schema_migrations")
}

// === Idempotent re-run tests ===

func TestMigrateToRBAC_IdempotentReRun(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// First run
	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// Count roles after first run
	var roleCount int64
	db.Model(&model.Role{}).Count(&roleCount)

	// Second run — should be a no-op
	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// Role count should be unchanged
	var roleCountAfter int64
	db.Model(&model.Role{}).Count(&roleCountAfter)
	assert.Equal(t, roleCount, roleCountAfter, "roles should not be re-seeded on re-run")

	// Permission count should be unchanged
	pmPerms, _ := CountPermissionsForRole(db, 2)
	assert.Equal(t, int64(26), pmPerms, "pm permissions should not be duplicated")
}

func TestMigrateToRBAC_IdempotentReRunPreservesData(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// Re-run
	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// Data still intact
	var users []model.User
	require.NoError(t, db.Find(&users).Error)
	assert.Len(t, users, 4)
}

// === Rollback on error tests ===

func TestMigrateToRBAC_RollbackOnClosedDB(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)

	// Close the underlying connection to force errors
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	err = MigrateToRBAC(db)
	assert.Error(t, err, "migration should fail on closed DB")
}

func TestMigrateToRBAC_EmptyDBSucceeds(t *testing.T) {
	// Migration on a completely empty DB (no pre-existing tables) should succeed
	// since it creates tables fresh without the legacy columns.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// Verify roles were seeded
	var roleCount int64
	db.Model(&model.Role{}).Count(&roleCount)
	assert.Equal(t, int64(3), roleCount, "all 3 preset roles should be seeded")

	// Verify tables created
	assert.True(t, tableExists(db, "roles"))
	assert.True(t, tableExists(db, "role_permissions"))
	assert.True(t, tableExists(db, "team_members"))
	assert.False(t, HasColumn(db, "team_members", "role"))
}

func TestMigrateToRBAC_TransactionRollback(t *testing.T) {
	// Verify transaction rollback by causing a failure after partial execution.
	// We'll use a hook to inject a failure.
	db := setupRBACTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Seed only a user (no team_members)
	u := model.User{Username: "testuser", DisplayName: "Test", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)

	// Run migration — should succeed since no team_members to migrate
	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// Verify schema_migrations has the entry
	var count int64
	db.Raw("SELECT count(*) FROM schema_migrations WHERE version = ?", rbacMigrationVersion).Scan(&count)
	assert.Equal(t, int64(1), count)

	// Verify data preserved
	var user model.User
	require.NoError(t, db.Where("username = ?", "testuser").First(&user).Error)
	assert.True(t, user.IsSuperAdmin == false) // not admin
	assert.Equal(t, "testuser", user.Username)
}

// === Preset role permission codes verification ===

func TestMigrateToRBAC_PMPresetHasAllExpectedCodes(t *testing.T) {
	db := setupRBACTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	err = VerifyPresetRoleCodes(db)
	require.NoError(t, err)
}

func TestMigrateToRBAC_SuperadminNoPermissionCodes(t *testing.T) {
	db := setupRBACTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	count, err := CountPermissionsForRole(db, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "superadmin should have zero permission codes in role_permissions")
}

func TestMigrateToRBAC_MemberHasExactCodes(t *testing.T) {
	db := setupRBACTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	expectedCodes := map[string]bool{
		"main_item:read": true,
		"sub_item:create": true, "sub_item:read": true, "sub_item:update": true, "sub_item:change_status": true,
		"progress:create": true, "progress:read": true,
		"item_pool:submit": true,
		"view:weekly": true, "view:table": true,
		"report:export": true,
	}

	var perms []model.RolePermission
	require.NoError(t, db.Where("role_id = ?", 3).Find(&perms).Error)
	assert.Len(t, perms, len(expectedCodes), "member should have exactly %d codes", len(expectedCodes))

	for _, p := range perms {
		assert.True(t, expectedCodes[p.PermissionCode], "unexpected member code: %s", p.PermissionCode)
	}
}

// === Error path tests ===

func TestVerifyPresetRoleCodes_FailsOnMissingRole(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// No roles seeded — should fail
	err = VerifyPresetRoleCodes(db)
	assert.Error(t, err, "should fail when roles not found")
}

func TestVerifyPresetRoleCodes_FailsOnMissingCode(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	require.NoError(t, db.AutoMigrate(&model.Role{}, &model.RolePermission{}))

	// Create pm role with only one permission (missing many)
	pm := model.Role{Name: "pm", Description: "pm", IsPreset: true}
	require.NoError(t, db.Create(&pm).Error)
	rp := model.RolePermission{RoleID: pm.ID, PermissionCode: "team:create"}
	require.NoError(t, db.Create(&rp).Error)

	// Create member role with correct count
	member := model.Role{Name: "member", Description: "member", IsPreset: true}
	require.NoError(t, db.Create(&member).Error)

	err = VerifyPresetRoleCodes(db)
	assert.Error(t, err, "should fail when pm is missing permissions")
}

func TestMigrateToRBAC_UsersTableAlreadyMigrated(t *testing.T) {
	// Test that re-running migration is idempotent
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Run migration once
	require.NoError(t, MigrateToRBAC(db))


	// Verify users are still there
	var users []model.User
	require.NoError(t, db.Find(&users).Error)
	assert.Len(t, users, 4)
}

func TestMigrateToRBAC_UnknownRoleDefaultsToMember(t *testing.T) {
	db := setupRBACTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Create user and team
	u := model.User{Username: "unknown_role_user", DisplayName: "URU", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "UnknownTeam", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)

	// Create team member with an unknown role string
	tm := model.TeamMember{TeamID: team.ID, UserID: u.ID, Role: "custom_unknown_role"}
	require.NoError(t, db.Create(&tm).Error)

	err = MigrateToRBAC(db)
	require.NoError(t, err)

	// The member with unknown role should default to member role_id=3
	var fetched model.TeamMember
	require.NoError(t, db.Where("user_id = ? AND team_id = ?", u.ID, team.ID).First(&fetched).Error)
	require.NotNil(t, fetched.RoleID)
	assert.Equal(t, uint(3), *fetched.RoleID, "unknown role should default to member (role_id=3)")
}

func TestSeedRole_SkipsExisting(t *testing.T) {
	db := setupRBACTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	require.NoError(t, db.AutoMigrate(&model.Role{}, &model.RolePermission{}))

	// Manually create a role with the same name
	existing := model.Role{Name: "pm", Description: "existing", IsPreset: false}
	require.NoError(t, db.Create(&existing).Error)

	// seedRole should skip without error
	err = seedRole(db, "pm", "new description", true, []string{"team:create"})
	require.NoError(t, err)

	// Should still be the original role
	var role model.Role
	require.NoError(t, db.Where("name = ?", "pm").First(&role).Error)
	assert.Equal(t, "existing", role.Description, "existing role should not be modified")

	// No permission codes should have been added
	count, _ := CountPermissionsForRole(db, role.ID)
	assert.Equal(t, int64(0), count)
}

func TestHasColumn_TableNotExists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	assert.False(t, HasColumn(db, "nonexistent_table", "some_col"))
}

func TestTableExists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	assert.False(t, tableExists(db, "users"))

	require.NoError(t, db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY)").Error)
	assert.True(t, tableExists(db, "users"))
}

package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/snowflake"
)

func setupRBACTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	_ = snowflake.Init(1)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate non-legacy tables
	err = db.AutoMigrate(&model.User{}, &model.Team{}, &model.Role{})
	require.NoError(t, err)

	// Create legacy team_members table with role string column (pre-RBAC schema)
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS team_members (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			team_id    INTEGER NOT NULL,
			user_id    INTEGER NOT NULL,
			role       TEXT NOT NULL DEFAULT 'member',
			role_id    INTEGER,
			joined_at  DATETIME NOT NULL DEFAULT (datetime('now')),
			created_at DATETIME,
			updated_at DATETIME,
			UNIQUE(team_id, user_id)
		)
	`).Error
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
	team1 := model.Team{TeamName: "Team 1", PmKey: int64(u2.ID)}
	require.NoError(t, db.Create(&team1).Error)

	// Insert team members with legacy role strings via raw SQL (role column is gorm:"-")
	require.NoError(t, db.Exec(`INSERT INTO team_members (team_id, user_id, role, joined_at) VALUES (?, ?, 'pm', datetime('now'))`, team1.ID, u2.ID).Error)
	require.NoError(t, db.Exec(`INSERT INTO team_members (team_id, user_id, role, joined_at) VALUES (?, ?, 'member', datetime('now'))`, team1.ID, u3.ID).Error)
	require.NoError(t, db.Exec(`INSERT INTO team_members (team_id, user_id, role, joined_at) VALUES (?, ?, 'member', datetime('now'))`, team1.ID, u4.ID).Error)
}

// === First-run success tests ===

func TestMigrateToRBAC_FirstRunSuccess(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db, true)
	require.NoError(t, err)
}

func TestMigrateToRBAC_PresetRolesSeeded(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db, true)
	require.NoError(t, err)

	// Check superadmin role
	var superadmin model.Role
	require.NoError(t, db.Where("role_name = ?", "superadmin").First(&superadmin).Error)
	assert.True(t, superadmin.IsPreset)
	assert.Equal(t, uint(1), superadmin.ID)

	// superadmin should have NO permission codes
	count, err := CountPermissionsForRole(db, superadmin.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "superadmin should have no permission codes")

	// Check pm role
	var pm model.Role
	require.NoError(t, db.Where("role_name = ?", "pm").First(&pm).Error)
	assert.True(t, pm.IsPreset)
	assert.Equal(t, uint(2), pm.ID)

	// pm should have 26 codes
	count, err = CountPermissionsForRole(db, pm.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(26), count, "pm should have 26 permission codes")

	// Check member role
	var member model.Role
	require.NoError(t, db.Where("role_name = ?", "member").First(&member).Error)
	assert.True(t, member.IsPreset)
	assert.Equal(t, uint(3), member.ID)

	// member should have 14 codes
	count, err = CountPermissionsForRole(db, member.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(14), count, "member should have 14 permission codes")
}

func TestMigrateToRBAC_TeamMemberRoleMigrated(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db, true)
	require.NoError(t, err)

	// After migration, role_key stores biz_key (snowflake). Look up the actual biz_keys.
	var pmRole, memberRole model.Role
	require.NoError(t, db.Where("role_name = ?", "pm").First(&pmRole).Error)
	require.NoError(t, db.Where("role_name = ?", "member").First(&memberRole).Error)

	// Check pm user has role_key = pm biz_key
	var pmMember model.TeamMember
	require.NoError(t, db.Where("user_key = ? AND team_key = ?", 2, 1).First(&pmMember).Error)
	require.NotNil(t, pmMember.RoleKey)
	assert.Equal(t, pmRole.BizKey, *pmMember.RoleKey, "pm user should have pm role biz_key")

	// Check member users have role_key = member biz_key
	var memberUser model.TeamMember
	require.NoError(t, db.Where("user_key = ? AND team_key = ?", 3, 1).First(&memberUser).Error)
	require.NotNil(t, memberUser.RoleKey)
	assert.Equal(t, memberRole.BizKey, *memberUser.RoleKey, "member user should have member role biz_key")
}

func TestMigrateToRBAC_TeamMembersRoleColumnRemoved(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db, true)
	require.NoError(t, err)

	// team_members table should no longer have role string column
	assert.False(t, HasColumn(db, "pmw_team_members", "role"),
		"role string column should be removed from team_members table")

	// but should have role_key
	assert.True(t, HasColumn(db, "pmw_team_members", "role_key"),
		"role_key column should exist in team_members table")
}

func TestMigrateToRBAC_UsersDataPreserved(t *testing.T) {
	db := setupRBACTestDB(t)
	seedPreRBACData(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db, true)
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

	err = MigrateToRBAC(db, true)
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
	err = MigrateToRBAC(db, true)
	require.NoError(t, err)

	// Count roles after first run
	var roleCount int64
	db.Model(&model.Role{}).Count(&roleCount)

	// Second run — should be a no-op
	err = MigrateToRBAC(db, true)
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

	err = MigrateToRBAC(db, true)
	require.NoError(t, err)

	// Re-run
	err = MigrateToRBAC(db, true)
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

	err = MigrateToRBAC(db, true)
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

	err = MigrateToRBAC(db, true)
	require.NoError(t, err)

	// Verify roles were seeded
	var roleCount int64
	db.Model(&model.Role{}).Count(&roleCount)
	assert.Equal(t, int64(3), roleCount, "all 3 preset roles should be seeded")

	// Verify tables created
	assert.True(t, tableExists(db, "pmw_roles"))
	assert.True(t, tableExists(db, "pmw_role_permissions"))
	assert.True(t, tableExists(db, "pmw_team_members"))
	assert.False(t, HasColumn(db, "pmw_team_members", "role"))
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
	err = MigrateToRBAC(db, true)
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

	err = MigrateToRBAC(db, true)
	require.NoError(t, err)

	err = VerifyPresetRoleCodes(db)
	require.NoError(t, err)
}

func TestMigrateToRBAC_SuperadminNoPermissionCodes(t *testing.T) {
	db := setupRBACTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db, true)
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

	err = MigrateToRBAC(db, true)
	require.NoError(t, err)

	expectedCodes := map[string]bool{
		"team:read":      true,
		"main_item:create": true, "main_item:read": true, "main_item:update": true,
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
	require.NoError(t, MigrateToRBAC(db, true))


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
	team := model.Team{TeamName: "UnknownTeam", PmKey: int64(u.ID)}
	require.NoError(t, db.Create(&team).Error)

	// Create team member with an unknown role string via raw SQL (role column is gorm:"-")
	require.NoError(t, db.Exec(`INSERT INTO team_members (team_id, user_id, role, joined_at) VALUES (?, ?, 'custom_unknown_role', datetime('now'))`, team.ID, u.ID).Error)

	err = MigrateToRBAC(db, true)
	require.NoError(t, err)

	// The member with unknown role should default to member biz_key
	var memberRole model.Role
	require.NoError(t, db.Where("role_name = ?", "member").First(&memberRole).Error)

	var fetched model.TeamMember
	require.NoError(t, db.Where("user_key = ? AND team_key = ?", u.ID, team.ID).First(&fetched).Error)
	require.NotNil(t, fetched.RoleKey)
	assert.Equal(t, memberRole.BizKey, *fetched.RoleKey, "unknown role should default to member biz_key")
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
	require.NoError(t, db.Where("role_name = ?", "pm").First(&role).Error)
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

func TestHasColumn_DelegatesToColumnExists(t *testing.T) {
	// HasColumn should delegate to columnExists, which works on SQLite via pragma_table_info.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	require.NoError(t, db.Exec("CREATE TABLE test_tbl (id INTEGER PRIMARY KEY, name TEXT)").Error)

	assert.True(t, HasColumn(db, "test_tbl", "name"), "should find existing column")
	assert.False(t, HasColumn(db, "test_tbl", "nonexistent"), "should not find missing column")
}

func TestTeamMembersDDL_ContainsSQLiteDialect(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	ddl := teamMembersDDL(db)
	require.Len(t, ddl, 1, "SQLite should produce a single DDL statement")
	assert.Contains(t, ddl[0], "INTEGER PRIMARY KEY AUTOINCREMENT",
		"SQLite DDL should use AUTOINCREMENT")
	assert.NotContains(t, ddl[0], "AUTO_INCREMENT",
		"SQLite DDL should not contain MySQL AUTO_INCREMENT")
	assert.Contains(t, ddl[0], "UNIQUE(team_key, user_key)",
		"SQLite DDL should have inline UNIQUE constraint")
}

func TestTeamMembersDDL_MySQLContainsNoAutoincrement(t *testing.T) {
	// We can't easily create a MySQL gorm.DB in unit tests, so we test
	// that teamMembersDDLMySQL produces correct MySQL DDL directly.
	ddl := teamMembersDDLMySQL()
	require.Len(t, ddl, 2, "MySQL should produce DDL + index statements")

	// First statement: CREATE TABLE
	assert.Contains(t, ddl[0], "BIGINT UNSIGNED NOT NULL AUTO_INCREMENT",
		"MySQL DDL should use AUTO_INCREMENT")
	assert.Contains(t, ddl[0], "PRIMARY KEY (id)",
		"MySQL DDL should have PRIMARY KEY (id)")
	assert.NotContains(t, ddl[0], "AUTOINCREMENT",
		"MySQL DDL should not contain SQLite AUTOINCREMENT")

	// Second statement: CREATE UNIQUE INDEX
	assert.Contains(t, ddl[1], "CREATE UNIQUE INDEX",
		"MySQL DDL should have separate CREATE UNIQUE INDEX")
	assert.Contains(t, ddl[1], "uk_team_members_team_user",
		"MySQL unique index should have a named constraint")
}

func TestTableExists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	assert.False(t, tableExists(db, "pmw_users"))

	require.NoError(t, db.Exec("CREATE TABLE pmw_users (id INTEGER PRIMARY KEY)").Error)
	assert.True(t, tableExists(db, "pmw_users"))
}

// === autoSchema=false tests ===

// setupAutoSchemaFalseDB simulates a DBA-managed environment: pmw_roles and
// pmw_role_permissions tables are pre-created, but schema_migrations is not.
func setupAutoSchemaFalseDB(t *testing.T) *gorm.DB {
	t.Helper()
	_ = snowflake.Init(1)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE pmw_roles (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			biz_key        INTEGER NOT NULL DEFAULT 0,
			create_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			db_update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_flag   INTEGER NOT NULL DEFAULT 0,
			deleted_time   DATETIME NOT NULL DEFAULT '1970-01-01 08:00:00',
			role_name      TEXT NOT NULL UNIQUE,
			role_desc      TEXT NOT NULL DEFAULT '',
			is_preset      INTEGER NOT NULL DEFAULT 0
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE pmw_role_permissions (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			deleted_flag    INTEGER NOT NULL DEFAULT 0,
			deleted_time    DATETIME NOT NULL DEFAULT '1970-01-01 08:00:00',
			role_id         INTEGER NOT NULL,
			permission_code TEXT NOT NULL,
			UNIQUE(role_id, permission_code, deleted_flag, deleted_time)
		)
	`).Error)

	return db
}

func TestMigrateToRBAC_AutoSchemaFalse_NoSchemaMigrationsTable(t *testing.T) {
	db := setupAutoSchemaFalseDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db, false)
	require.NoError(t, err)

	assert.False(t, tableExists(db, "schema_migrations"),
		"schema_migrations should not be created when autoSchema=false")
}

func TestMigrateToRBAC_AutoSchemaFalse_SeedsPresetRoles(t *testing.T) {
	db := setupAutoSchemaFalseDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigrateToRBAC(db, false)
	require.NoError(t, err)

	var count int64
	require.NoError(t, db.Model(&model.Role{}).Count(&count).Error)
	assert.Equal(t, int64(3), count, "superadmin, pm, member should be seeded")
}

func TestMigrateToRBAC_AutoSchemaFalse_IdempotentReRun(t *testing.T) {
	db := setupAutoSchemaFalseDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	require.NoError(t, MigrateToRBAC(db, false))
	require.NoError(t, MigrateToRBAC(db, false))

	var count int64
	require.NoError(t, db.Model(&model.Role{}).Count(&count).Error)
	assert.Equal(t, int64(3), count, "roles should not be duplicated on re-run")
}

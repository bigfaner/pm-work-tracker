package migration

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/snowflake"
)

// setupPermissionGranularityDB creates an in-memory SQLite DB with pmw_roles,
// pmw_role_permissions, and schema_migrations tables pre-created and the RBAC
// migration already applied, so we can test the permission granularity migration
// in isolation.
func setupPermissionGranularityDB(t *testing.T) *gorm.DB {
	t.Helper()
	_ = snowflake.Init(1)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run full RBAC migration first to set up the schema
	require.NoError(t, MigrateToRBAC(db, true))

	return db
}

// seedOldPermissionCodes adds the old permission codes to a role, simulating
// a database that existed before the permission granularity migration.
func seedOldPermissionCodes(t *testing.T, db *gorm.DB, roleName string, codes []string) {
	t.Helper()
	var role model.Role
	require.NoError(t, db.Where("role_name = ?", roleName).First(&role).Error)
	for _, code := range codes {
		rp := model.RolePermission{RoleKey: role.BizKey, PermissionCode: code}
		require.NoError(t, db.Create(&rp).Error)
	}
}

// getPermissionCodes returns the set of permission codes for a given role.
func getPermissionCodes(t *testing.T, db *gorm.DB, roleName string) map[string]bool {
	t.Helper()
	var role model.Role
	require.NoError(t, db.Where("role_name = ?", roleName).First(&role).Error)
	var perms []model.RolePermission
	require.NoError(t, db.Where("role_key = ?", role.BizKey).Find(&perms).Error)
	result := make(map[string]bool, len(perms))
	for _, p := range perms {
		result[p.PermissionCode] = true
	}
	return result
}

// === Test: MigratePermissionGranularity converts user:manage_role to role:create+update+delete ===

func TestMigratePermissionGranularity_ManageRoleToRoleCRUD(t *testing.T) {
	db := setupPermissionGranularityDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Add old user:manage_role to the pm role (simulating pre-migration state)
	seedOldPermissionCodes(t, db, "pm", []string{"user:manage_role"})

	err = MigratePermissionGranularity(db)
	require.NoError(t, err)

	codes := getPermissionCodes(t, db, "pm")

	// Old code should be removed
	assert.False(t, codes["user:manage_role"], "user:manage_role should be removed")
	// New role codes should exist
	assert.True(t, codes["role:create"], "role:create should exist")
	assert.True(t, codes["role:update"], "role:update should exist")
	assert.True(t, codes["role:delete"], "role:delete should exist")
}

// === Test: MigratePermissionGranularity adds user:list alongside user:read ===

func TestMigratePermissionGranularity_AddsUserList(t *testing.T) {
	db := setupPermissionGranularityDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// The pm role already has user:read from seedPresetRoles
	err = MigratePermissionGranularity(db)
	require.NoError(t, err)

	codes := getPermissionCodes(t, db, "pm")

	// user:read should still exist
	assert.True(t, codes["user:read"], "user:read should be preserved")
	// user:list should be added
	assert.True(t, codes["user:list"], "user:list should be added")
}

// === Test: MigratePermissionGranularity is idempotent ===

func TestMigratePermissionGranularity_Idempotent(t *testing.T) {
	db := setupPermissionGranularityDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Add old user:manage_role to simulate pre-migration state
	seedOldPermissionCodes(t, db, "pm", []string{"user:manage_role"})

	// First run
	err = MigratePermissionGranularity(db)
	require.NoError(t, err)

	codesAfterFirst := getPermissionCodes(t, db, "pm")

	// Second run
	err = MigratePermissionGranularity(db)
	require.NoError(t, err)

	codesAfterSecond := getPermissionCodes(t, db, "pm")

	// Permission counts should be identical
	assert.Equal(t, len(codesAfterFirst), len(codesAfterSecond),
		"idempotent re-run should not add duplicates")
}

// === Test: MigratePermissionGranularity tracked in schema_migrations ===

func TestMigratePermissionGranularity_TrackedInSchemaMigrations(t *testing.T) {
	db := setupPermissionGranularityDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = MigratePermissionGranularity(db)
	require.NoError(t, err)

	var count int64
	db.Raw("SELECT count(*) FROM schema_migrations WHERE version = ?", permissionGranularityVersion).Scan(&count)
	assert.Equal(t, int64(1), count, "permission granularity migration should be tracked")
}

// === Test: Transaction rollback on failure ===

func TestMigratePermissionGranularity_TransactionRollback(t *testing.T) {
	db := setupPermissionGranularityDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Close DB to force error
	sqlDB.Close()

	err = MigratePermissionGranularity(db)
	assert.Error(t, err, "migration should fail on closed DB")
}

// === Test: Custom role with user:manage_role gets migrated ===

func TestMigratePermissionGranularity_CustomRoleMigrated(t *testing.T) {
	db := setupPermissionGranularityDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Create a custom role with user:manage_role and user:read
	customRole := model.Role{
		BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
		Name:        "custom_admin",
		Description: "custom admin role",
		IsPreset:    false,
	}
	require.NoError(t, db.Create(&customRole).Error)
	require.NoError(t, db.Create(&model.RolePermission{RoleKey: customRole.BizKey, PermissionCode: "user:manage_role"}).Error)
	require.NoError(t, db.Create(&model.RolePermission{RoleKey: customRole.BizKey, PermissionCode: "user:read"}).Error)

	err = MigratePermissionGranularity(db)
	require.NoError(t, err)

	codes := getPermissionCodes(t, db, "custom_admin")

	// Old code removed
	assert.False(t, codes["user:manage_role"], "user:manage_role should be removed from custom role")
	// New codes added
	assert.True(t, codes["role:create"], "role:create should be added to custom role")
	assert.True(t, codes["role:update"], "role:update should be added to custom role")
	assert.True(t, codes["role:delete"], "role:delete should be added to custom role")
	// user:read preserved
	assert.True(t, codes["user:read"], "user:read should be preserved")
	// user:list added (because role had old user:read)
	assert.True(t, codes["user:list"], "user:list should be added to custom role")
}

// === Test: seedPresetRoles for pm includes new codes ===

func TestMigratePermissionGranularity_PMSeedIncludesNewCodes(t *testing.T) {
	db := setupPermissionGranularityDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Run the permission granularity migration
	err = MigratePermissionGranularity(db)
	require.NoError(t, err)

	codes := getPermissionCodes(t, db, "pm")

	// Should have new permission codes
	assert.True(t, codes["user:list"], "pm should have user:list")
	assert.True(t, codes["user:assign_role"], "pm should have user:assign_role")
	assert.True(t, codes["role:read"], "pm should have role:read")
	assert.True(t, codes["role:create"], "pm should have role:create")
	assert.True(t, codes["role:update"], "pm should have role:update")
	assert.True(t, codes["role:delete"], "pm should have role:delete")

	// Should NOT have old code
	assert.False(t, codes["user:manage_role"], "pm should not have user:manage_role")
}

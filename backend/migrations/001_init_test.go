package migrations_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func Test001Init_CreatesAllTables(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Read and execute the migration file
	content := readMigrationFile(t, "001_init.sql")

	// Execute the migration
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Verify all expected tables exist
	expectedTables := []string{
		"users",
		"teams",
		"team_members",
		"main_items",
		"sub_items",
		"progress_records",
		"item_pools",
	}

	for _, table := range expectedTables {
		var count int64
		err := db.Raw("SELECT count(*) FROM " + table).Scan(&count).Error
		assert.NoError(t, err, "table %s should exist", table)
	}
}

func Test001Init_UserTableHasCorrectSchema(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Test unique index on username
	db.Exec("INSERT INTO users (username, display_name, password_hash, is_super_admin, created_at, updated_at) VALUES ('user1', 'User One', 'hash1', 0, datetime('now'), datetime('now'))")
	err = db.Exec("INSERT INTO users (username, display_name, password_hash, is_super_admin, created_at, updated_at) VALUES ('user1', 'User Two', 'hash2', 0, datetime('now'), datetime('now'))").Error
	assert.Error(t, err, "duplicate username should be rejected")
}

func Test001Init_TeamMemberUniqueIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Insert team and users
	db.Exec("INSERT INTO users (username, display_name, password_hash, is_super_admin, created_at, updated_at) VALUES ('u1', 'U1', 'h', 0, datetime('now'), datetime('now'))")
	db.Exec("INSERT INTO teams (name, pm_id, created_at, updated_at) VALUES ('Team1', 1, datetime('now'), datetime('now'))")

	// First member OK
	err = db.Exec("INSERT INTO team_members (team_id, user_id, role, joined_at, created_at, updated_at) VALUES (1, 1, 'member', datetime('now'), datetime('now'), datetime('now'))").Error
	require.NoError(t, err)

	// Duplicate team_id+user_id should fail
	err = db.Exec("INSERT INTO team_members (team_id, user_id, role, joined_at, created_at, updated_at) VALUES (1, 1, 'pm', datetime('now'), datetime('now'), datetime('now'))").Error
	assert.Error(t, err, "duplicate (team_id, user_id) should be rejected")
}

func readMigrationFile(t *testing.T, filename string) []byte {
	t.Helper()
	paths := []string{
		filename,
		filepath.Join("migrations", filename),
		filepath.Join("..", "..", "migrations", filename),
	}
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			return data
		}
	}
	t.Fatalf("migration file %s not found", filename)
	return nil
}

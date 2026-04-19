package migrations_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func Test004_AddsEmailAndStatusColumns(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Apply 001 first to create the tables
	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Apply 004
	content = readMigrationFile(t, "004_user_email_status.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Insert a user and verify the new columns work
	err = db.Exec(`INSERT INTO users (username, display_name, password_hash, is_super_admin, email, status, created_at, updated_at)
		VALUES ('testuser', 'Test User', 'hash', 0, 'test@example.com', 'enabled', datetime('now'), datetime('now'))`).Error
	require.NoError(t, err)

	// Verify email and status columns exist and have correct values
	var email, status string
	err = db.Raw("SELECT email FROM users WHERE username = 'testuser'").Scan(&email).Error
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", email)

	err = db.Raw("SELECT status FROM users WHERE username = 'testuser'").Scan(&status).Error
	assert.NoError(t, err)
	assert.Equal(t, "enabled", status)
}

func Test004_DefaultStatusIsEnabled(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	content = readMigrationFile(t, "004_user_email_status.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Insert without specifying status - should default to 'enabled'
	err = db.Exec(`INSERT INTO users (username, display_name, password_hash, is_super_admin, email, created_at, updated_at)
		VALUES ('defaultuser', 'Default User', 'hash', 0, '', datetime('now'), datetime('now'))`).Error
	require.NoError(t, err)

	var status string
	err = db.Raw("SELECT status FROM users WHERE username = 'defaultuser'").Scan(&status).Error
	assert.NoError(t, err)
	assert.Equal(t, "enabled", status)
}

func Test004_ExistingUsersNotAffected(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Insert a user BEFORE migration
	err = db.Exec(`INSERT INTO users (username, display_name, password_hash, is_super_admin, created_at, updated_at)
		VALUES ('olduser', 'Old User', 'hash', 0, datetime('now'), datetime('now'))`).Error
	require.NoError(t, err)

	// Apply 004
	content = readMigrationFile(t, "004_user_email_status.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Old user should have default status 'enabled' and empty email
	var email, status string
	err = db.Raw("SELECT email FROM users WHERE username = 'olduser'").Scan(&email).Error
	assert.NoError(t, err)
	assert.Equal(t, "", email)

	err = db.Raw("SELECT status FROM users WHERE username = 'olduser'").Scan(&status).Error
	assert.NoError(t, err)
	assert.Equal(t, "enabled", status)
}

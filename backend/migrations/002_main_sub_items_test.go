package migrations_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func Test002_AddsSubItemsTeamPriorityIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Apply 001 first to create the tables
	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Apply 002
	content = readMigrationFile(t, "002_main_sub_items.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Verify the index exists by checking SQLite index list
	var count int64
	err = db.Raw("SELECT count(*) FROM sqlite_master WHERE type='index' AND name='idx_sub_items_team_priority'").Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "idx_sub_items_team_priority index should exist")
}

func Test002_Idempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	content = readMigrationFile(t, "002_main_sub_items.sql")
	// Running twice should not error (IF NOT EXISTS)
	err = db.Exec(string(content)).Error
	assert.NoError(t, err)
	err = db.Exec(string(content)).Error
	assert.NoError(t, err)
}

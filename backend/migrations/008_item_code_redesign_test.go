package migrations_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func Test008_AddsTeamsCodeColumnAndIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	content = readMigrationFile(t, "008_item_code_redesign.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	var count int64
	err = db.Raw("SELECT count(*) FROM sqlite_master WHERE type='index' AND name='idx_teams_code'").Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "idx_teams_code index should exist")
}

func Test008_AddsSubItemsCodeColumnAndIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	content = readMigrationFile(t, "008_item_code_redesign.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	var count int64
	err = db.Raw("SELECT count(*) FROM sqlite_master WHERE type='index' AND name='idx_sub_items_main_code'").Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "idx_sub_items_main_code index should exist")
}

func Test008_TeamsCodeColumnExists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	content = readMigrationFile(t, "008_item_code_redesign.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Insert a row with code to verify the column exists
	err = db.Exec("INSERT INTO teams (name, pm_id, code, created_at, updated_at) VALUES ('T1', 1, 'ALPHA', datetime('now'), datetime('now'))").Error
	assert.NoError(t, err, "teams.code column should exist and accept values")
}

func Test008_SubItemsCodeColumnExists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	content = readMigrationFile(t, "008_item_code_redesign.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Insert a row with code to verify the column exists
	err = db.Exec("INSERT INTO sub_items (team_id, main_item_id, title, priority, code, created_at, updated_at) VALUES (1, 1, 'sub', 'P1', 'ALPHA-00001-01', datetime('now'), datetime('now'))").Error
	assert.NoError(t, err, "sub_items.code column should exist and accept values")
}

package migrations_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func Test009_AddsSubItemsMainCodeIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	require.NoError(t, db.Exec(string(content)).Error)

	content = readMigrationFile(t, "008_item_code_redesign.sql")
	require.NoError(t, db.Exec(string(content)).Error)

	// Populate unique codes before adding the unique index.
	require.NoError(t, db.Exec("INSERT INTO sub_items (team_id, main_item_id, title, priority, code, created_at, updated_at) VALUES (1, 1, 's1', 'P1', 'ALPHA-00001-01', datetime('now'), datetime('now'))").Error)
	require.NoError(t, db.Exec("INSERT INTO sub_items (team_id, main_item_id, title, priority, code, created_at, updated_at) VALUES (1, 1, 's2', 'P1', 'ALPHA-00001-02', datetime('now'), datetime('now'))").Error)

	content = readMigrationFile(t, "009_sub_items_code_index.sql")
	require.NoError(t, db.Exec(string(content)).Error)

	var count int64
	err = db.Raw("SELECT count(*) FROM sqlite_master WHERE type='index' AND name='idx_sub_items_main_code'").Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "idx_sub_items_main_code index should exist after 009")
}

func Test009_UniqueIndexEnforced(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	require.NoError(t, db.Exec(string(content)).Error)

	content = readMigrationFile(t, "008_item_code_redesign.sql")
	require.NoError(t, db.Exec(string(content)).Error)

	require.NoError(t, db.Exec("INSERT INTO sub_items (team_id, main_item_id, title, priority, code, created_at, updated_at) VALUES (1, 1, 's1', 'P1', 'ALPHA-00001-01', datetime('now'), datetime('now'))").Error)

	content = readMigrationFile(t, "009_sub_items_code_index.sql")
	require.NoError(t, db.Exec(string(content)).Error)

	// Duplicate (main_item_id, code) should be rejected.
	err = db.Exec("INSERT INTO sub_items (team_id, main_item_id, title, priority, code, created_at, updated_at) VALUES (1, 1, 's2', 'P1', 'ALPHA-00001-01', datetime('now'), datetime('now'))").Error
	assert.Error(t, err, "duplicate (main_item_id, code) should violate unique index")
}

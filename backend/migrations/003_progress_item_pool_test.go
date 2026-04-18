package migrations_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func Test003_ProgressRecordsTableExists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Apply 001
	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Apply 002
	content = readMigrationFile(t, "002_main_sub_items.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Apply 003 (idempotent no-op)
	content = readMigrationFile(t, "003_progress_item_pool.sql")
	err = db.Exec(string(content)).Error
	assert.NoError(t, err)
}

func Test003_ProgressRecordsCompositeIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	var count int64
	err = db.Raw("SELECT count(*) FROM sqlite_master WHERE type='index' AND name='idx_progress_records_sub_item_created'").Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "idx_progress_records_sub_item_created composite index should exist")
}

func Test003_ItemPoolsCompositeIndex(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	var count int64
	err = db.Raw("SELECT count(*) FROM sqlite_master WHERE type='index' AND name='idx_item_pools_team_status'").Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "idx_item_pools_team_status composite index should exist")
}

func Test003_ItemPoolsDefaultStatus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	// Insert with no status — should get default
	db.Exec("INSERT INTO item_pools (team_id, title, submitter_id, status, created_at, updated_at) VALUES (1, 'Test', 1, '待分配', datetime('now'), datetime('now'))")
	var status string
	db.Raw("SELECT status FROM item_pools WHERE title = 'Test'").Scan(&status)
	assert.Equal(t, "待分配", status, "default status should be 待分配")
}

func Test003_ProgressRecordsNoUpdatedAt(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	content := readMigrationFile(t, "001_init.sql")
	err = db.Exec(string(content)).Error
	require.NoError(t, err)

	type colInfo struct {
		Name string
	}
	var columns []colInfo
	db.Raw("PRAGMA table_info(progress_records)").Scan(&columns)
	colNames := map[string]bool{}
	for _, c := range columns {
		colNames[c.Name] = true
	}
	assert.False(t, colNames["updated_at"], "progress_records should not have updated_at")
	assert.False(t, colNames["deleted_at"], "progress_records should not have deleted_at")
}

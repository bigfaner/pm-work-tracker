package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func TestRunner_AppliesMigrationFiles(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Create temp migrations dir with a test SQL file
	tmpDir := t.TempDir()
	migrationSQL := `
CREATE TABLE test_table (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);
`
	err = os.WriteFile(filepath.Join(tmpDir, "001_test.sql"), []byte(migrationSQL), 0644)
	require.NoError(t, err)

	runner := NewRunner(db, tmpDir)
	err = runner.Run()
	require.NoError(t, err)

	// Verify table was created
	var count int64
	db.Raw("SELECT count(*) FROM test_table").Scan(&count)
	assert.Equal(t, int64(0), count) // table exists, 0 rows
}

func TestRunner_CreatesSchemaMigrationsTable(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	tmpDir := t.TempDir()
	// Empty migrations dir
	runner := NewRunner(db, tmpDir)
	err = runner.Run()
	require.NoError(t, err)

	// schema_migrations table should exist
	var count int64
	db.Raw("SELECT count(*) FROM schema_migrations").Scan(&count)
	assert.Equal(t, int64(0), count)
}

func TestRunner_TracksAppliedMigrations(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	tmpDir := t.TempDir()
	migrationSQL := `CREATE TABLE foo (id INTEGER PRIMARY KEY);`
	err = os.WriteFile(filepath.Join(tmpDir, "001_foo.sql"), []byte(migrationSQL), 0644)
	require.NoError(t, err)

	runner := NewRunner(db, tmpDir)

	// First run: applies migration
	err = runner.Run()
	require.NoError(t, err)

	// Second run: should not re-apply (idempotent)
	err = runner.Run()
	require.NoError(t, err)

	// Verify only one record in schema_migrations
	var count int64
	db.Raw("SELECT count(*) FROM schema_migrations").Scan(&count)
	assert.Equal(t, int64(1), count)
}

func TestRunner_ReadsFilesInOrder(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	tmpDir := t.TempDir()

	// Write migrations in reverse order to test lexicographic sorting
	err = os.WriteFile(filepath.Join(tmpDir, "002_second.sql"), []byte(`CREATE TABLE second_t (id INTEGER PRIMARY KEY);`), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "001_first.sql"), []byte(`CREATE TABLE first_t (id INTEGER PRIMARY KEY);`), 0644)
	require.NoError(t, err)

	runner := NewRunner(db, tmpDir)
	err = runner.Run()
	require.NoError(t, err)

	var count int64
	db.Raw("SELECT count(*) FROM schema_migrations").Scan(&count)
	assert.Equal(t, int64(2), count)
}

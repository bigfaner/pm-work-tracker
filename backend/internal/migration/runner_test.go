package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func TestRunSchema_CreatesTables(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	tmpDir := t.TempDir()
	schemaSQL := `
CREATE TABLE test_table (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);
`
	schemaFile := filepath.Join(tmpDir, "schema.sql")
	err = os.WriteFile(schemaFile, []byte(schemaSQL), 0644)
	require.NoError(t, err)

	err = RunSchema(db, schemaFile)
	require.NoError(t, err)

	var count int64
	db.Raw("SELECT count(*) FROM test_table").Scan(&count)
	assert.Equal(t, int64(0), count)
}

func TestRunSchema_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	tmpDir := t.TempDir()
	schemaSQL := `
CREATE TABLE IF NOT EXISTS foo (id INTEGER PRIMARY KEY, name TEXT);
`
	schemaFile := filepath.Join(tmpDir, "schema.sql")
	err = os.WriteFile(schemaFile, []byte(schemaSQL), 0644)
	require.NoError(t, err)

	err = RunSchema(db, schemaFile)
	require.NoError(t, err)

	// Re-run: should succeed without error (IF NOT EXISTS)
	err = RunSchema(db, schemaFile)
	require.NoError(t, err)
}

func TestRunSchema_FileNotFound(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = RunSchema(db, "/nonexistent/schema.sql")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read schema")
}

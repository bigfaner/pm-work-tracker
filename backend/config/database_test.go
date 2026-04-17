package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDB_SQLiteInMemory(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_PATH", ":memory:")
	defer os.Unsetenv("DB_DRIVER")
	defer os.Unsetenv("DB_PATH")

	db, err := InitDB()
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify connection is alive
	sqlDB, err := db.DB()
	require.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())
	sqlDB.Close()
}

func TestInitDB_SQLiteDefaultPath(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_PATH", filepath.Join(tmpDir, "test.db"))
	defer os.Unsetenv("DB_DRIVER")
	defer os.Unsetenv("DB_PATH")

	db, err := InitDB()
	require.NoError(t, err)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()
}

func TestInitDB_InvalidDriver(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	defer os.Unsetenv("DB_DRIVER")

	db, err := InitDB()
	assert.Nil(t, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

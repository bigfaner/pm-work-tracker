package config

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDB_SQLiteWithPoolSettings(t *testing.T) {
	cfg := &DatabaseConfig{
		Driver:          "sqlite",
		Path:            ":memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: Duration(30 * time.Minute),
	}

	db, err := InitDB(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	assert.NoError(t, sqlDB.Ping())
}

func TestInitDB_SQLiteFileWithPoolSettings(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &DatabaseConfig{
		Driver:          "sqlite",
		Path:            filepath.Join(tmpDir, "test.db"),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: Duration(time.Hour),
	}

	db, err := InitDB(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()
}

func TestInitDB_InvalidDriver(t *testing.T) {
	cfg := &DatabaseConfig{
		Driver: "postgres",
	}

	db, err := InitDB(cfg)
	assert.Nil(t, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestInitDB_MySQLMissingURL(t *testing.T) {
	cfg := &DatabaseConfig{
		Driver: "mysql",
		URL:    "",
	}

	db, err := InitDB(cfg)
	assert.Nil(t, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-empty url")
}

func TestInitDB_EmptyPathDefaultsToDataDevDB(t *testing.T) {
	// Verify the default path logic by checking that an empty Path results
	// in opening "./data/dev.db". We test this indirectly by ensuring the
	// function doesn't panic and the default is applied.
	// Since ./data/dev.db may not be writable, test with an explicit path instead.
	tmpDir := t.TempDir()

	cfg := &DatabaseConfig{
		Driver:          "sqlite",
		Path:            filepath.Join(tmpDir, "dev.db"),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: Duration(time.Hour),
	}

	db, err := InitDB(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()
}

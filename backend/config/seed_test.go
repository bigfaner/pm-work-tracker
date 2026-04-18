package config

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}))
	return db
}

func TestSeedAdmin_SuccessfulCreation(t *testing.T) {
	db := setupTestDB(t)
	authCfg := &AuthConfig{
		InitialAdmin: InitialAdminConfig{
			Username: "admin",
			Password: "secret123",
		},
	}

	err := SeedAdmin(db, authCfg)
	assert.NoError(t, err)

	var user model.User
	require.NoError(t, db.Where("username = ?", "admin").First(&user).Error)
	assert.Equal(t, "admin", user.Username)
	assert.True(t, user.IsSuperAdmin)
	assert.NotEqual(t, "secret123", user.PasswordHash) // not plaintext

	// Verify password is bcrypt-hashed and can be verified
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("secret123"))
	assert.NoError(t, err)
}

func TestSeedAdmin_SkipWhenUsernameEmpty(t *testing.T) {
	db := setupTestDB(t)
	authCfg := &AuthConfig{
		InitialAdmin: InitialAdminConfig{
			Username: "",
			Password: "secret123",
		},
	}

	err := SeedAdmin(db, authCfg)
	assert.NoError(t, err)

	var count int64
	db.Model(&model.User{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestSeedAdmin_SkipWhenUserExists(t *testing.T) {
	db := setupTestDB(t)

	// Pre-create the admin user
	existing := model.User{
		Username:     "admin",
		DisplayName:  "admin",
		PasswordHash: "oldhash",
		IsSuperAdmin: true,
	}
	require.NoError(t, db.Create(&existing).Error)

	authCfg := &AuthConfig{
		InitialAdmin: InitialAdminConfig{
			Username: "admin",
			Password: "newsecret",
		},
	}

	err := SeedAdmin(db, authCfg)
	assert.NoError(t, err)

	// Verify original user was NOT overwritten
	var user model.User
	require.NoError(t, db.Where("username = ?", "admin").First(&user).Error)
	assert.Equal(t, "oldhash", user.PasswordHash)
}

func TestSeedAdmin_PasswordIsHashed(t *testing.T) {
	db := setupTestDB(t)
	authCfg := &AuthConfig{
		InitialAdmin: InitialAdminConfig{
			Username: "admin",
			Password: "mypassword",
		},
	}

	err := SeedAdmin(db, authCfg)
	require.NoError(t, err)

	var user model.User
	require.NoError(t, db.Where("username = ?", "admin").First(&user).Error)

	// Verify the hash is a valid bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("mypassword"))
	assert.NoError(t, err, "password should be bcrypt-hashed and verifiable")
}

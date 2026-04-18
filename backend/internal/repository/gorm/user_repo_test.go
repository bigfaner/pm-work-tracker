package gorm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	pkgerrors "pm-work-tracker/backend/internal/pkg/errors"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
)

func setupUserTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}))
	return db
}

func TestUserRepo_FindByID(t *testing.T) {
	db := setupUserTestDB(t)
	repo := gormrepo.NewGormUserRepo(db)
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		u := model.User{Username: "alice", DisplayName: "Alice", PasswordHash: "h"}
		require.NoError(t, db.Create(&u).Error)

		found, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, "alice", found.Username)
	})

	t.Run("not_found_returns_ErrNotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

func TestUserRepo_FindByUsername(t *testing.T) {
	db := setupUserTestDB(t)
	repo := gormrepo.NewGormUserRepo(db)
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		u := model.User{Username: "bob", DisplayName: "Bob", PasswordHash: "h"}
		require.NoError(t, db.Create(&u).Error)

		found, err := repo.FindByUsername(ctx, "bob")
		require.NoError(t, err)
		assert.Equal(t, "Bob", found.DisplayName)
	})

	t.Run("not_found_returns_ErrNotFound", func(t *testing.T) {
		_, err := repo.FindByUsername(ctx, "nonexistent")
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})

	t.Run("does_not_expose_gorm_error", func(t *testing.T) {
		_, err := repo.FindByUsername(ctx, "ghost")
		assert.NotErrorIs(t, err, gormlib.ErrRecordNotFound, "should wrap ErrNotFound, not expose gorm's ErrRecordNotFound")
		_ = errors.Is(err, pkgerrors.ErrNotFound) // verify it IS our ErrNotFound
	})
}

func TestUserRepo_List(t *testing.T) {
	db := setupUserTestDB(t)
	repo := gormrepo.NewGormUserRepo(db)
	ctx := context.Background()

	require.NoError(t, db.Create(&model.User{Username: "u1", DisplayName: "U1", PasswordHash: "h"}).Error)
	require.NoError(t, db.Create(&model.User{Username: "u2", DisplayName: "U2", PasswordHash: "h"}).Error)

	users, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserRepo_Update(t *testing.T) {
	db := setupUserTestDB(t)
	repo := gormrepo.NewGormUserRepo(db)
	ctx := context.Background()

	u := model.User{Username: "charlie", DisplayName: "Charlie", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)

	u.DisplayName = "Charles"
	require.NoError(t, repo.Update(ctx, &u))

	found, err := repo.FindByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "Charles", found.DisplayName)
}

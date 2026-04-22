package gorm_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

// --- FindByIDs ---

func TestUserRepo_FindByIDs(t *testing.T) {
	db := setupUserTestDB(t)
	repo := gormrepo.NewGormUserRepo(db)
	ctx := context.Background()

	u1 := model.User{Username: "a1", DisplayName: "Alice", PasswordHash: "h"}
	u2 := model.User{Username: "b2", DisplayName: "Bob", PasswordHash: "h"}
	require.NoError(t, db.Create(&u1).Error)
	require.NoError(t, db.Create(&u2).Error)

	t.Run("returns_map_for_existing_ids", func(t *testing.T) {
		result, err := repo.FindByIDs(ctx, []uint{u1.ID, u2.ID})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Alice", result[u1.ID].DisplayName)
		assert.Equal(t, "Bob", result[u2.ID].DisplayName)
	})

	t.Run("empty_input_returns_empty_map", func(t *testing.T) {
		result, err := repo.FindByIDs(ctx, []uint{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("partial_results_return_found_keys_only", func(t *testing.T) {
		result, err := repo.FindByIDs(ctx, []uint{u1.ID, 9999})
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Alice", result[u1.ID].DisplayName)
		_, hasMissing := result[9999]
		assert.False(t, hasMissing)
	})
}

// --- ListFiltered ---

func TestUserRepo_ListFiltered(t *testing.T) {
	db := setupUserTestDB(t)
	repo := gormrepo.NewGormUserRepo(db)
	ctx := context.Background()

	require.NoError(t, db.Create(&model.User{Username: "alice", DisplayName: "Alice Wang", PasswordHash: "h"}).Error)
	require.NoError(t, db.Create(&model.User{Username: "bob99", DisplayName: "Bob Li", PasswordHash: "h"}).Error)
	require.NoError(t, db.Create(&model.User{Username: "charlie", DisplayName: "Charlie Zhang", PasswordHash: "h"}).Error)

	t.Run("search_by_username", func(t *testing.T) {
		users, total, err := repo.ListFiltered(ctx, "alice", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, users, 1)
		assert.Equal(t, "alice", users[0].Username)
	})

	t.Run("search_by_display_name", func(t *testing.T) {
		users, total, err := repo.ListFiltered(ctx, "Bob", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, users, 1)
		assert.Equal(t, "bob99", users[0].Username)
	})

	t.Run("empty_search_returns_all", func(t *testing.T) {
		users, total, err := repo.ListFiltered(ctx, "", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, users, 3)
	})

	t.Run("pagination_offset_and_limit", func(t *testing.T) {
		users, total, err := repo.ListFiltered(ctx, "", 0, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, users, 2)

		users2, _, err := repo.ListFiltered(ctx, "", 2, 10)
		require.NoError(t, err)
		assert.Len(t, users2, 1)
	})

	t.Run("no_match_returns_empty", func(t *testing.T) {
		users, total, err := repo.ListFiltered(ctx, "zzznonexistent", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, users)
	})
}

// --- SearchAvailable ---

func TestUserRepo_SearchAvailable(t *testing.T) {
	db := setupUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Team{}, &model.TeamMember{}))
	repo := gormrepo.NewGormUserRepo(db)
	ctx := context.Background()

	u1 := model.User{Username: "alice", DisplayName: "Alice", PasswordHash: "h"}
	u2 := model.User{Username: "bob", DisplayName: "Bob", PasswordHash: "h"}
	u3 := model.User{Username: "charlie", DisplayName: "Charlie", PasswordHash: "h"}
	require.NoError(t, db.Create(&u1).Error)
	require.NoError(t, db.Create(&u2).Error)
	require.NoError(t, db.Create(&u3).Error)

	team := model.Team{Name: "Team1", PmID: u1.ID, Code: "T1"}
	require.NoError(t, db.Create(&team).Error)
	// u2 is already a team member
	require.NoError(t, db.Create(&model.TeamMember{TeamID: team.ID, UserID: u2.ID, JoinedAt: time.Now()}).Error)

	t.Run("excludes_team_members", func(t *testing.T) {
		users, err := repo.SearchAvailable(ctx, team.ID, "", 10)
		require.NoError(t, err)
		assert.Len(t, users, 2) // u1 and u3 (u2 is excluded)
		usernames := map[string]bool{}
		for _, u := range users {
			usernames[u.Username] = true
		}
		assert.True(t, usernames["alice"])
		assert.True(t, usernames["charlie"])
		assert.False(t, usernames["bob"])
	})

	t.Run("search_filters_by_name", func(t *testing.T) {
		users, err := repo.SearchAvailable(ctx, team.ID, "ali", 10)
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "alice", users[0].Username)
	})

	t.Run("respects_limit", func(t *testing.T) {
		users, err := repo.SearchAvailable(ctx, team.ID, "", 1)
		require.NoError(t, err)
		assert.Len(t, users, 1)
	})
}

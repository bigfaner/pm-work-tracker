package gorm_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	pkgerrors "pm-work-tracker/backend/internal/pkg/errors"
	gormrepo "pm-work-tracker/backend/internal/repository/gorm"
)

func setupItemPoolTestDB(t *testing.T) *gormlib.DB {
	t.Helper()
	db, err := gormlib.Open(sqlite.Open(":memory:"), &gormlib.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Team{}, &model.ItemPool{}))
	return db
}

func seedItemPoolData(t *testing.T, db *gormlib.DB) (*model.User, *model.Team) {
	t.Helper()
	u := model.User{Username: "ip_user", DisplayName: "IP User", PasswordHash: "h"}
	require.NoError(t, db.Create(&u).Error)
	team := model.Team{Name: "IP Team", PmID: u.ID}
	require.NoError(t, db.Create(&team).Error)
	return &u, &team
}

func createItemPool(t *testing.T, db *gormlib.DB, teamID, submitterID uint, title, status string) *model.ItemPool {
	t.Helper()
	item := model.ItemPool{
		TeamID:      teamID,
		Title:       title,
		SubmitterID: submitterID,
		Status:      status,
	}
	require.NoError(t, db.Create(&item).Error)
	return &item
}

// --- Create ---

func TestItemPoolRepo_Create(t *testing.T) {
	db := setupItemPoolTestDB(t)
	repo := gormrepo.NewGormItemPoolRepo(db)
	ctx := context.Background()

	u, team := seedItemPoolData(t, db)
	item := &model.ItemPool{
		TeamID:          team.ID,
		Title:           "New Suggestion",
		Background:      "Some context",
		ExpectedOutput:  "Expected result",
		SubmitterID:     u.ID,
		Status:          "待分配",
	}
	require.NoError(t, repo.Create(ctx, item))
	assert.NotZero(t, item.ID)
}

// --- FindByID ---

func TestItemPoolRepo_FindByID(t *testing.T) {
	db := setupItemPoolTestDB(t)
	repo := gormrepo.NewGormItemPoolRepo(db)
	ctx := context.Background()

	u, team := seedItemPoolData(t, db)
	item := createItemPool(t, db, team.ID, u.ID, "Find Me", "待分配")

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, item.ID)
		require.NoError(t, err)
		assert.Equal(t, "Find Me", found.Title)
		assert.Equal(t, team.ID, found.TeamID)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 9999)
		assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
	})
}

// --- Update ---

func TestItemPoolRepo_Update(t *testing.T) {
	db := setupItemPoolTestDB(t)
	repo := gormrepo.NewGormItemPoolRepo(db)
	ctx := context.Background()

	u, team := seedItemPoolData(t, db)
	item := createItemPool(t, db, team.ID, u.ID, "Assign Me", "待分配")

	mainID := u.ID
	subID := u.ID
	assigneeID := u.ID
	fields := map[string]interface{}{
		"status":           "已分配",
		"assigned_main_id": mainID,
		"assigned_sub_id":  subID,
		"assignee_id":      assigneeID,
	}
	require.NoError(t, repo.Update(ctx, item, fields))

	found, err := repo.FindByID(ctx, item.ID)
	require.NoError(t, err)
	assert.Equal(t, "已分配", found.Status)
	assert.Equal(t, mainID, *found.AssignedMainID)
}

func TestItemPoolRepo_Update_NotFound(t *testing.T) {
	db := setupItemPoolTestDB(t)
	repo := gormrepo.NewGormItemPoolRepo(db)
	ctx := context.Background()

	_, team := seedItemPoolData(t, db)
	fakeItem := &model.ItemPool{Model: gormlib.Model{ID: 9999}, TeamID: team.ID}
	fields := map[string]interface{}{"status": "已分配"}
	err := repo.Update(ctx, fakeItem, fields)
	assert.ErrorIs(t, err, pkgerrors.ErrNotFound)
}

// --- List ---

func TestItemPoolRepo_List(t *testing.T) {
	db := setupItemPoolTestDB(t)
	repo := gormrepo.NewGormItemPoolRepo(db)
	ctx := context.Background()

	u, team := seedItemPoolData(t, db)

	createItemPool(t, db, team.ID, u.ID, "Pool A", "待分配")
	createItemPool(t, db, team.ID, u.ID, "Pool B", "已分配")
	createItemPool(t, db, team.ID, u.ID, "Pool C", "已拒绝")

	// Another team - should not appear
	u2 := model.User{Username: "ip_other", DisplayName: "IP Other", PasswordHash: "h"}
	require.NoError(t, db.Create(&u2).Error)
	team2 := model.Team{Name: "IP Team2", PmID: u2.ID}
	require.NoError(t, db.Create(&team2).Error)
	createItemPool(t, db, team2.ID, u2.ID, "Other Team", "待分配")

	t.Run("all_for_team", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.ItemPoolFilter{}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("filter_by_status", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.ItemPoolFilter{Status: "已分配"}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Pool B", result.Items[0].Title)
	})

	t.Run("filter_by_status_rejected", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.ItemPoolFilter{Status: "已拒绝"}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Pool C", result.Items[0].Title)
	})

	t.Run("pagination", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.ItemPoolFilter{}, dto.Pagination{Page: 1, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 2, result.Size)
	})

	t.Run("pagination_page2", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.ItemPoolFilter{}, dto.Pagination{Page: 2, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 1)
	})

	t.Run("default_pagination", func(t *testing.T) {
		result, err := repo.List(ctx, team.ID, dto.ItemPoolFilter{}, dto.Pagination{})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.Size)
	})

	t.Run("team_isolation", func(t *testing.T) {
		result, err := repo.List(ctx, team2.ID, dto.ItemPoolFilter{}, dto.Pagination{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Other Team", result.Items[0].Title)
	})
}
